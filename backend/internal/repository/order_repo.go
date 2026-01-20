package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/order"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type orderRepository struct {
	client *dbent.Client
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(client *dbent.Client) service.OrderRepository {
	return &orderRepository{client: client}
}

func (r *orderRepository) Create(ctx context.Context, o *service.Order) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Order.Create().
		SetOrderNo(o.OrderNo).
		SetUserID(o.UserID).
		SetGroupID(o.GroupID).
		SetAmount(o.Amount).
		SetStatus(o.Status)

	if o.TradeNo != nil {
		builder.SetTradeNo(*o.TradeNo)
	}
	if o.PayType != nil {
		builder.SetPayType(*o.PayType)
	}
	if o.PaidAt != nil {
		builder.SetPaidAt(*o.PaidAt)
	}
	if o.SubscriptionID != nil {
		builder.SetSubscriptionID(*o.SubscriptionID)
	}
	if o.ExpiredAt != nil {
		builder.SetExpiredAt(*o.ExpiredAt)
	}

	created, err := builder.Save(ctx)
	if err == nil {
		applyOrderEntityToService(o, created)
	}
	return translatePersistenceError(err, nil, nil)
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*service.Order, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.Order.Query().
		Where(order.IDEQ(id)).
		WithUser().
		WithGroup().
		WithSubscription().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrOrderNotFound, nil)
	}
	return orderEntityToService(m), nil
}

func (r *orderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*service.Order, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.Order.Query().
		Where(order.OrderNoEQ(orderNo)).
		WithUser().
		WithGroup().
		WithSubscription().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrOrderNotFound, nil)
	}
	return orderEntityToService(m), nil
}

func (r *orderRepository) Update(ctx context.Context, o *service.Order) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Order.UpdateOneID(o.ID).
		SetStatus(o.Status)

	if o.TradeNo != nil {
		builder.SetTradeNo(*o.TradeNo)
	}
	if o.PayType != nil {
		builder.SetPayType(*o.PayType)
	}
	if o.PaidAt != nil {
		builder.SetPaidAt(*o.PaidAt)
	}
	if o.SubscriptionID != nil {
		builder.SetSubscriptionID(*o.SubscriptionID)
	}

	updated, err := builder.Save(ctx)
	if err == nil {
		applyOrderEntityToService(o, updated)
		return nil
	}
	return translatePersistenceError(err, service.ErrOrderNotFound, nil)
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id int64, status string, tradeNo *string, payType *string) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Order.UpdateOneID(id).SetStatus(status)
	if tradeNo != nil {
		builder.SetTradeNo(*tradeNo)
	}
	if payType != nil {
		builder.SetPayType(*payType)
	}
	_, err := builder.Save(ctx)
	return translatePersistenceError(err, service.ErrOrderNotFound, nil)
}

func (r *orderRepository) SetPaid(ctx context.Context, id int64, tradeNo string, payType string, subscriptionID int64) error {
	client := clientFromContext(ctx, r.client)
	now := time.Now()
	_, err := client.Order.UpdateOneID(id).
		SetStatus(service.OrderStatusPaid).
		SetTradeNo(tradeNo).
		SetPayType(payType).
		SetPaidAt(now).
		SetSubscriptionID(subscriptionID).
		Save(ctx)
	return translatePersistenceError(err, service.ErrOrderNotFound, nil)
}

func (r *orderRepository) ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.Order, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	q := client.Order.Query().Where(order.UserIDEQ(userID))

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		WithGroup().
		WithSubscription().
		Order(dbent.Desc(order.FieldCreatedAt)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return orderEntitiesToService(orders), paginationResultFromTotal(int64(total), params), nil
}

func (r *orderRepository) ListAll(ctx context.Context, params pagination.PaginationParams, status string, search string) ([]service.Order, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	q := client.Order.Query()

	// 过滤状态
	if status != "" {
		q = q.Where(order.StatusEQ(status))
	}

	// 搜索订单号
	if search != "" {
		q = q.Where(order.OrderNoContains(search))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		WithUser().
		WithGroup().
		WithSubscription().
		Order(dbent.Desc(order.FieldCreatedAt)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return orderEntitiesToService(orders), paginationResultFromTotal(int64(total), params), nil
}

func (r *orderRepository) ListPending(ctx context.Context) ([]service.Order, error) {
	client := clientFromContext(ctx, r.client)
	orders, err := client.Order.Query().
		Where(
			order.StatusEQ(service.OrderStatusPending),
			order.ExpiredAtNotNil(),
			order.ExpiredAtLTE(time.Now()),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return orderEntitiesToService(orders), nil
}

func (r *orderRepository) BatchUpdateExpired(ctx context.Context) (int64, error) {
	client := clientFromContext(ctx, r.client)
	n, err := client.Order.Update().
		Where(
			order.StatusEQ(service.OrderStatusPending),
			order.ExpiredAtNotNil(),
			order.ExpiredAtLTE(time.Now()),
		).
		SetStatus(service.OrderStatusExpired).
		Save(ctx)
	return int64(n), err
}

func orderEntityToService(m *dbent.Order) *service.Order {
	if m == nil {
		return nil
	}
	out := &service.Order{
		ID:             m.ID,
		OrderNo:        m.OrderNo,
		TradeNo:        m.TradeNo,
		UserID:         m.UserID,
		GroupID:        m.GroupID,
		Amount:         m.Amount,
		Status:         m.Status,
		PayType:        m.PayType,
		PaidAt:         m.PaidAt,
		SubscriptionID: m.SubscriptionID,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		ExpiredAt:      m.ExpiredAt,
	}
	if m.Edges.User != nil {
		out.User = userEntityToService(m.Edges.User)
	}
	if m.Edges.Group != nil {
		out.Group = groupEntityToService(m.Edges.Group)
	}
	if m.Edges.Subscription != nil {
		out.Subscription = userSubscriptionEntityToService(m.Edges.Subscription)
	}
	return out
}

func orderEntitiesToService(models []*dbent.Order) []service.Order {
	out := make([]service.Order, 0, len(models))
	for i := range models {
		if o := orderEntityToService(models[i]); o != nil {
			out = append(out, *o)
		}
	}
	return out
}

func applyOrderEntityToService(dst *service.Order, src *dbent.Order) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}
