package repository

import (
	"context"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/chatconversation"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
)

type chatConversationRepository struct {
	client *dbent.Client
}

func NewChatConversationRepository(client *dbent.Client) service.ChatConversationRepository {
	return &chatConversationRepository{client: client}
}

func (r *chatConversationRepository) Create(ctx context.Context, c *service.ChatConversation) error {
	client := clientFromContext(ctx, r.client)
	builder := client.ChatConversation.Create().
		SetVisitorName(c.VisitorName).
		SetStatus(c.Status)

	if c.GuestToken != nil {
		builder.SetGuestToken(*c.GuestToken)
	}
	if c.UserID != nil {
		builder.SetUserID(*c.UserID)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	applyChatConversationEntity(c, created)
	return nil
}

func (r *chatConversationRepository) GetByID(ctx context.Context, id int64) (*service.ChatConversation, error) {
	m, err := r.client.ChatConversation.Get(ctx, id)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrChatConversationNotFound, nil)
	}
	return chatConversationToService(m), nil
}

func (r *chatConversationRepository) GetOpenByGuestToken(ctx context.Context, token string) (*service.ChatConversation, error) {
	m, err := r.client.ChatConversation.Query().
		Where(
			chatconversation.GuestTokenEQ(token),
			chatconversation.StatusEQ(service.ChatConversationStatusOpen),
		).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrChatConversationNotFound, nil)
	}
	return chatConversationToService(m), nil
}

func (r *chatConversationRepository) GetOpenByUserID(ctx context.Context, userID int64) (*service.ChatConversation, error) {
	m, err := r.client.ChatConversation.Query().
		Where(
			chatconversation.UserIDEQ(userID),
			chatconversation.StatusEQ(service.ChatConversationStatusOpen),
		).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrChatConversationNotFound, nil)
	}
	return chatConversationToService(m), nil
}

func (r *chatConversationRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	client := clientFromContext(ctx, r.client)
	return client.ChatConversation.UpdateOneID(id).
		SetStatus(status).
		Exec(ctx)
}

func (r *chatConversationRepository) IncrementAdminUnread(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	return client.ChatConversation.UpdateOneID(id).
		AddAdminUnreadCount(1).
		Exec(ctx)
}

func (r *chatConversationRepository) ResetAdminUnread(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	return client.ChatConversation.UpdateOneID(id).
		SetAdminUnreadCount(0).
		Exec(ctx)
}

func (r *chatConversationRepository) UpdateLastMessage(ctx context.Context, id int64, preview string, at time.Time) error {
	client := clientFromContext(ctx, r.client)
	return client.ChatConversation.UpdateOneID(id).
		SetLastMessagePreview(preview).
		SetLastMessageAt(at).
		Exec(ctx)
}

func (r *chatConversationRepository) List(
	ctx context.Context,
	params pagination.PaginationParams,
	filters service.ChatConversationListFilters,
) ([]service.ChatConversation, *pagination.PaginationResult, error) {
	q := r.client.ChatConversation.Query().
		// 只展示真正聊过的会话:访客点开气泡就会建会话,一句话没发的空会话
		// 对客服没有价值,且 last_message_at 为 NULL 在 Postgres 的
		// `ORDER BY ... DESC` 下默认 NULLS FIRST,会把这些空壳顶到列表最前面。
		Where(chatconversation.LastMessageAtNotNil())

	if filters.Status != "" {
		q = q.Where(chatconversation.StatusEQ(filters.Status))
	}
	if filters.Search != "" {
		q = q.Where(
			chatconversation.Or(
				chatconversation.VisitorNameContainsFold(filters.Search),
				chatconversation.LastMessagePreviewContainsFold(filters.Search),
			),
		)
	}
	if len(filters.ExcludeUserIDs) > 0 {
		// 排除管理员自己的会话;访客会话(user_id 为 NULL)需显式保留
		// (SQL `NOT IN` 会过滤掉 NULL 行)
		q = q.Where(chatconversation.Or(
			chatconversation.UserIDIsNil(),
			chatconversation.UserIDNotIn(filters.ExcludeUserIDs...),
		))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	itemsQuery := q.
		Offset(params.Offset()).
		Limit(params.Limit())
	for _, order := range chatConversationListOrders(params) {
		itemsQuery = itemsQuery.Order(order)
	}
	items, err := itemsQuery.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return chatConversationEntitiesToService(items), paginationResultFromTotal(int64(total), params), nil
}

func (r *chatConversationRepository) CountUnread(ctx context.Context, excludeUserIDs []int64) (int64, error) {
	q := r.client.ChatConversation.Query().
		Where(
			chatconversation.StatusEQ(service.ChatConversationStatusOpen),
			chatconversation.AdminUnreadCountGT(0),
		)
	if len(excludeUserIDs) > 0 {
		// 与 List 一致:排除管理员会话,保留访客(user_id NULL)
		q = q.Where(chatconversation.Or(
			chatconversation.UserIDIsNil(),
			chatconversation.UserIDNotIn(excludeUserIDs...),
		))
	}
	count, err := q.Count(ctx)
	return int64(count), err
}

func chatConversationListOrders(params pagination.PaginationParams) []func(*entsql.Selector) {
	field, sortOrder := chatConversationListOrder(params)
	if sortOrder == pagination.SortOrderAsc {
		return []func(*entsql.Selector){
			dbent.Asc(field),
			dbent.Asc(chatconversation.FieldID),
		}
	}
	return []func(*entsql.Selector){
		dbent.Desc(field),
		dbent.Desc(chatconversation.FieldID),
	}
}

func chatConversationListOrder(params pagination.PaginationParams) (string, string) {
	sortBy := strings.ToLower(strings.TrimSpace(params.SortBy))
	sortOrder := params.NormalizedSortOrder(pagination.SortOrderDesc)

	switch sortBy {
	case "visitor_name":
		return chatconversation.FieldVisitorName, sortOrder
	case "status":
		return chatconversation.FieldStatus, sortOrder
	case "admin_unread_count":
		return chatconversation.FieldAdminUnreadCount, sortOrder
	case "created_at":
		return chatconversation.FieldCreatedAt, sortOrder
	case "", "last_message_at":
		return chatconversation.FieldLastMessageAt, sortOrder
	default:
		return chatconversation.FieldLastMessageAt, sortOrder
	}
}

func applyChatConversationEntity(dst *service.ChatConversation, src *dbent.ChatConversation) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.GuestToken = src.GuestToken
	dst.UserID = src.UserID
	dst.VisitorName = src.VisitorName
	dst.Status = src.Status
	dst.AdminUnreadCount = src.AdminUnreadCount
	dst.LastMessageAt = src.LastMessageAt
	dst.LastMessagePreview = src.LastMessagePreview
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}

func chatConversationToService(m *dbent.ChatConversation) *service.ChatConversation {
	if m == nil {
		return nil
	}
	return &service.ChatConversation{
		ID:                 m.ID,
		GuestToken:         m.GuestToken,
		UserID:             m.UserID,
		VisitorName:        m.VisitorName,
		Status:             m.Status,
		AdminUnreadCount:   m.AdminUnreadCount,
		LastMessageAt:      m.LastMessageAt,
		LastMessagePreview: m.LastMessagePreview,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

func chatConversationEntitiesToService(models []*dbent.ChatConversation) []service.ChatConversation {
	out := make([]service.ChatConversation, 0, len(models))
	for i := range models {
		if s := chatConversationToService(models[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
