package service

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// adminIDsCacheTTL 控制管理员 ID 列表的内存缓存时长。
// 客服未读数被前台按 ~10s 轮询,缓存可避免每次都查 users 表(role 列无索引)。
// 管理员变更极少,分钟级陈旧完全可接受。
const adminIDsCacheTTL = 60 * time.Second

type ChatService struct {
	convRepo ChatConversationRepository
	msgRepo  ChatMessageRepository
	userRepo UserRepository

	adminIDsMu       sync.Mutex
	adminIDs         []int64
	adminIDsExpireAt time.Time
}

func NewChatService(convRepo ChatConversationRepository, msgRepo ChatMessageRepository, userRepo UserRepository) *ChatService {
	return &ChatService{
		convRepo: convRepo,
		msgRepo:  msgRepo,
		userRepo: userRepo,
	}
}

// GetOpenConversation 返回访客当前的开启会话(不创建);没有则返回 nil。
// 用于前台气泡在页面加载时"探查"是否有未读的管理员回复,而不副作用地新建空会话。
func (s *ChatService) GetOpenConversation(ctx context.Context, guestToken *string, userID *int64) (*ChatConversation, error) {
	if userID != nil {
		if conv, err := s.convRepo.GetOpenByUserID(ctx, *userID); err == nil && conv != nil {
			return conv, nil
		}
		return nil, nil
	}
	if guestToken != nil && *guestToken != "" {
		if conv, err := s.convRepo.GetOpenByGuestToken(ctx, *guestToken); err == nil && conv != nil {
			return conv, nil
		}
	}
	return nil, nil
}

func (s *ChatService) GetOrCreateConversation(ctx context.Context, guestToken *string, userID *int64, visitorName string) (*ChatConversation, error) {
	if userID != nil {
		conv, err := s.convRepo.GetOpenByUserID(ctx, *userID)
		if err == nil && conv != nil {
			return conv, nil
		}
	} else if guestToken != nil && *guestToken != "" {
		conv, err := s.convRepo.GetOpenByGuestToken(ctx, *guestToken)
		if err == nil && conv != nil {
			return conv, nil
		}
	}

	if visitorName == "" {
		visitorName = "访客"
	}

	conv := &ChatConversation{
		GuestToken:  guestToken,
		UserID:      userID,
		VisitorName: visitorName,
		Status:      ChatConversationStatusOpen,
	}
	if err := s.convRepo.Create(ctx, conv); err != nil {
		return nil, err
	}
	return conv, nil
}

func (s *ChatService) SendMessage(ctx context.Context, conversationID int64, senderType, content string) (*ChatMessage, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrChatMessageEmpty
	}
	if utf8.RuneCountInString(content) > domain.ChatMessageMaxLength {
		return nil, ErrChatMessageTooLong
	}
	if senderType != ChatSenderTypeVisitor && senderType != ChatSenderTypeAdmin {
		return nil, ErrChatInvalidSenderType
	}

	conv, err := s.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if conv.Status == ChatConversationStatusClosed {
		return nil, ErrChatConversationClosed
	}

	now := time.Now()
	msg := &ChatMessage{
		ConversationID: conversationID,
		SenderType:     senderType,
		Content:        content,
		CreatedAt:      now,
	}
	if err := s.msgRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	preview := content
	if utf8.RuneCountInString(preview) > 100 {
		runes := []rune(preview)
		preview = string(runes[:100]) + "..."
	}
	if err := s.convRepo.UpdateLastMessage(ctx, conversationID, preview, now); err != nil {
		slog.Warn("chat: failed to update last message metadata", "conversation_id", conversationID, "error", err)
	}

	if senderType == ChatSenderTypeVisitor {
		if err := s.convRepo.IncrementAdminUnread(ctx, conversationID); err != nil {
			slog.Warn("chat: failed to increment unread count", "conversation_id", conversationID, "error", err)
		}
	}

	return msg, nil
}

func (s *ChatService) ListConversations(ctx context.Context, params pagination.PaginationParams, filters ChatConversationListFilters) ([]ChatConversation, *pagination.PaginationResult, error) {
	filters.ExcludeUserIDs = s.adminUserIDs(ctx)
	return s.convRepo.List(ctx, params, filters)
}

// adminUserIDs 返回管理员用户 ID(用于把管理员自己的会话排除出客服列表与未读统计),
// 带 60s 内存缓存以避免高频轮询反复查库;出错时回退到上次缓存,避免误把管理员会话放出来。
func (s *ChatService) adminUserIDs(ctx context.Context) []int64 {
	if s.userRepo == nil {
		return nil
	}
	s.adminIDsMu.Lock()
	defer s.adminIDsMu.Unlock()
	if time.Now().Before(s.adminIDsExpireAt) {
		return s.adminIDs
	}
	ids, err := s.userRepo.ListIDsByRole(ctx, RoleAdmin)
	if err != nil {
		slog.Warn("chat: failed to list admin user ids for exclusion", "error", err)
		return s.adminIDs // 复用上次结果(可能为 nil),不刷新过期时间,下次重试
	}
	s.adminIDs = ids
	s.adminIDsExpireAt = time.Now().Add(adminIDsCacheTTL)
	return ids
}

func (s *ChatService) GetConversation(ctx context.Context, id int64) (*ChatConversation, error) {
	return s.convRepo.GetByID(ctx, id)
}

func (s *ChatService) GetMessages(ctx context.Context, conversationID int64, limit, offset int) ([]ChatMessage, int, error) {
	return s.msgRepo.ListByConversation(ctx, conversationID, limit, offset)
}

func (s *ChatService) CloseConversation(ctx context.Context, id int64) error {
	return s.convRepo.UpdateStatus(ctx, id, ChatConversationStatusClosed)
}

func (s *ChatService) ReopenConversation(ctx context.Context, id int64) error {
	return s.convRepo.UpdateStatus(ctx, id, ChatConversationStatusOpen)
}

func (s *ChatService) MarkAdminRead(ctx context.Context, id int64) error {
	return s.convRepo.ResetAdminUnread(ctx, id)
}

func (s *ChatService) CountUnread(ctx context.Context) (int64, error) {
	return s.convRepo.CountUnread(ctx, s.adminUserIDs(ctx))
}

// ResolveDisplayName 返回管理员会话列表展示用的名称:
// 登录用户 → 用户名(优先)或邮箱; 未登录访客 → "访客"。
// 用户已被软删时仍按 ID 取名(管理员视角),取不到则回退为 "用户"。
func (s *ChatService) ResolveDisplayName(ctx context.Context, userID *int64) string {
	if userID == nil {
		return "访客"
	}
	if s.userRepo != nil {
		if u, err := s.userRepo.GetByIDIncludeDeleted(ctx, *userID); err == nil && u != nil {
			if u.Username != "" {
				return u.Username
			}
			if u.Email != "" {
				return u.Email
			}
		}
	}
	return "用户"
}
