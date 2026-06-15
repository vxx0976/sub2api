package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	ChatSenderTypeVisitor        = domain.ChatSenderTypeVisitor
	ChatSenderTypeAdmin          = domain.ChatSenderTypeAdmin
	ChatConversationStatusOpen   = domain.ChatConversationStatusOpen
	ChatConversationStatusClosed = domain.ChatConversationStatusClosed
)

var (
	ErrChatConversationNotFound = domain.ErrChatConversationNotFound
	ErrChatMessageEmpty         = domain.ErrChatMessageEmpty
	ErrChatMessageTooLong       = domain.ErrChatMessageTooLong
	ErrChatConversationClosed   = domain.ErrChatConversationClosed
	ErrChatUnauthorized         = domain.ErrChatUnauthorized
	ErrChatInvalidGuestToken    = domain.ErrChatInvalidGuestToken
	ErrChatInvalidSenderType    = domain.ErrChatInvalidSenderType
)

type ChatConversation = domain.ChatConversation
type ChatMessage = domain.ChatMessage

type ChatConversationListFilters struct {
	Status string
	Search string
}

type ChatConversationRepository interface {
	Create(ctx context.Context, c *ChatConversation) error
	GetByID(ctx context.Context, id int64) (*ChatConversation, error)
	GetOpenByGuestToken(ctx context.Context, token string) (*ChatConversation, error)
	GetOpenByUserID(ctx context.Context, userID int64) (*ChatConversation, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	IncrementAdminUnread(ctx context.Context, id int64) error
	ResetAdminUnread(ctx context.Context, id int64) error
	UpdateLastMessage(ctx context.Context, id int64, preview string, at time.Time) error
	List(ctx context.Context, params pagination.PaginationParams, filters ChatConversationListFilters) ([]ChatConversation, *pagination.PaginationResult, error)
	CountUnread(ctx context.Context) (int64, error)
}

type ChatMessageRepository interface {
	Create(ctx context.Context, m *ChatMessage) error
	ListByConversation(ctx context.Context, conversationID int64, limit, offset int) ([]ChatMessage, int, error)
}
