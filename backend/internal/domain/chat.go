package domain

import (
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	ChatSenderTypeVisitor = "visitor"
	ChatSenderTypeAdmin   = "admin"
)

const (
	ChatConversationStatusOpen   = "open"
	ChatConversationStatusClosed = "closed"
)

const ChatMessageMaxLength = 5000

var (
	ErrChatConversationNotFound = infraerrors.NotFound("CHAT_CONVERSATION_NOT_FOUND", "conversation not found")
	ErrChatMessageEmpty         = infraerrors.BadRequest("CHAT_MESSAGE_EMPTY", "message content is required")
	ErrChatMessageTooLong       = infraerrors.BadRequest("CHAT_MESSAGE_TOO_LONG", "message content exceeds maximum length")
	ErrChatConversationClosed   = infraerrors.BadRequest("CHAT_CONVERSATION_CLOSED", "conversation is closed")
	ErrChatUnauthorized         = infraerrors.Unauthorized("CHAT_UNAUTHORIZED", "not authorized to access this conversation")
	ErrChatInvalidGuestToken    = infraerrors.BadRequest("CHAT_INVALID_GUEST_TOKEN", "invalid guest token format")
	ErrChatInvalidSenderType    = infraerrors.BadRequest("CHAT_INVALID_SENDER_TYPE", "invalid sender type")
)

type ChatConversation struct {
	ID                 int64
	GuestToken         *string
	UserID             *int64
	VisitorName        string
	Status             string
	AdminUnreadCount   int
	LastMessageAt      *time.Time
	LastMessagePreview string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ChatMessage struct {
	ID             int64
	ConversationID int64
	SenderType     string
	Content        string
	CreatedAt      time.Time
}
