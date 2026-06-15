package service

import (
	"context"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type ChatService struct {
	convRepo ChatConversationRepository
	msgRepo  ChatMessageRepository
}

func NewChatService(convRepo ChatConversationRepository, msgRepo ChatMessageRepository) *ChatService {
	return &ChatService{
		convRepo: convRepo,
		msgRepo:  msgRepo,
	}
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
	return s.convRepo.List(ctx, params, filters)
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
	return s.convRepo.CountUnread(ctx)
}
