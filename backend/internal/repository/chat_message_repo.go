package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/chatmessage"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type chatMessageRepository struct {
	client *dbent.Client
}

func NewChatMessageRepository(client *dbent.Client) service.ChatMessageRepository {
	return &chatMessageRepository{client: client}
}

func (r *chatMessageRepository) Create(ctx context.Context, m *service.ChatMessage) error {
	client := clientFromContext(ctx, r.client)
	created, err := client.ChatMessage.Create().
		SetConversationID(m.ConversationID).
		SetSenderType(m.SenderType).
		SetContent(m.Content).
		Save(ctx)
	if err != nil {
		return err
	}
	m.ID = created.ID
	m.CreatedAt = created.CreatedAt
	return nil
}

func (r *chatMessageRepository) ListByConversation(ctx context.Context, conversationID int64, limit, offset int) ([]service.ChatMessage, int, error) {
	q := r.client.ChatMessage.Query().
		Where(chatmessage.ConversationIDEQ(conversationID))

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// offset 为负表示"最后一页":会话可能积压上千条消息,默认要展示最近的 limit 条。
	// 否则列表预览显示的是最新消息、点进去却只有最早的 50 条,对不上。
	if offset < 0 {
		offset = total - limit
		if offset < 0 {
			offset = 0
		}
	}

	items, err := q.
		Order(dbent.Asc(chatmessage.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return chatMessageEntitiesToService(items), total, nil
}

func chatMessageToService(m *dbent.ChatMessage) *service.ChatMessage {
	if m == nil {
		return nil
	}
	return &service.ChatMessage{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderType:     m.SenderType,
		Content:        m.Content,
		CreatedAt:      m.CreatedAt,
	}
}

func chatMessageEntitiesToService(models []*dbent.ChatMessage) []service.ChatMessage {
	out := make([]service.ChatMessage, 0, len(models))
	for i := range models {
		if s := chatMessageToService(models[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
