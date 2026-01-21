package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/rechargeorder"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type rechargeOrderRepository struct {
	client *dbent.Client
}

// NewRechargeOrderRepository creates a new recharge order repository
func NewRechargeOrderRepository(client *dbent.Client) service.RechargeOrderRepository {
	return &rechargeOrderRepository{client: client}
}

func (r *rechargeOrderRepository) Create(ctx context.Context, o *service.RechargeOrder) error {
	client := clientFromContext(ctx, r.client)
	builder := client.RechargeOrder.Create().
		SetOrderNo(o.OrderNo).
		SetUserID(o.UserID).
		SetAmount(o.Amount).
		SetCreditAmount(o.CreditAmount).
		SetMultiplier(o.Multiplier).
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
	if o.ExpiredAt != nil {
		builder.SetExpiredAt(*o.ExpiredAt)
	}

	created, err := builder.Save(ctx)
	if err == nil {
		applyRechargeOrderEntityToService(o, created)
	}
	return translatePersistenceError(err, nil, nil)
}

func (r *rechargeOrderRepository) GetByID(ctx context.Context, id int64) (*service.RechargeOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.RechargeOrder.Query().
		Where(rechargeorder.IDEQ(id)).
		WithUser().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrRechargeOrderNotFound, nil)
	}
	return rechargeOrderEntityToService(m), nil
}

func (r *rechargeOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*service.RechargeOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.RechargeOrder.Query().
		Where(rechargeorder.OrderNoEQ(orderNo)).
		WithUser().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrRechargeOrderNotFound, nil)
	}
	return rechargeOrderEntityToService(m), nil
}

func (r *rechargeOrderRepository) Update(ctx context.Context, o *service.RechargeOrder) error {
	client := clientFromContext(ctx, r.client)
	builder := client.RechargeOrder.UpdateOneID(o.ID).
		SetStatus(o.Status).
		SetAmount(o.Amount).
		SetCreditAmount(o.CreditAmount).
		SetMultiplier(o.Multiplier)

	if o.TradeNo != nil {
		builder.SetTradeNo(*o.TradeNo)
	}
	if o.PayType != nil {
		builder.SetPayType(*o.PayType)
	}
	if o.PaidAt != nil {
		builder.SetPaidAt(*o.PaidAt)
	}
	if o.ExpiredAt != nil {
		builder.SetExpiredAt(*o.ExpiredAt)
	}

	updated, err := builder.Save(ctx)
	if err == nil {
		applyRechargeOrderEntityToService(o, updated)
		return nil
	}
	return translatePersistenceError(err, service.ErrRechargeOrderNotFound, nil)
}

func (r *rechargeOrderRepository) UpdateStatus(ctx context.Context, id int64, status string, tradeNo *string, payType *string) error {
	client := clientFromContext(ctx, r.client)
	builder := client.RechargeOrder.UpdateOneID(id).SetStatus(status)
	if tradeNo != nil {
		builder.SetTradeNo(*tradeNo)
	}
	if payType != nil {
		builder.SetPayType(*payType)
	}
	_, err := builder.Save(ctx)
	return translatePersistenceError(err, service.ErrRechargeOrderNotFound, nil)
}

func (r *rechargeOrderRepository) SetPaid(ctx context.Context, id int64, tradeNo string, payType string) error {
	client := clientFromContext(ctx, r.client)
	now := time.Now()
	_, err := client.RechargeOrder.UpdateOneID(id).
		SetStatus(service.RechargeOrderStatusPaid).
		SetTradeNo(tradeNo).
		SetPayType(payType).
		SetPaidAt(now).
		Save(ctx)
	return translatePersistenceError(err, service.ErrRechargeOrderNotFound, nil)
}

func (r *rechargeOrderRepository) ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.RechargeOrder, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	q := client.RechargeOrder.Query().Where(rechargeorder.UserIDEQ(userID))

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		Order(dbent.Desc(rechargeorder.FieldCreatedAt)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return rechargeOrderEntitiesToService(orders), paginationResultFromTotal(int64(total), params), nil
}

func (r *rechargeOrderRepository) ListAll(ctx context.Context, params pagination.PaginationParams, keyword string, status string) ([]service.RechargeOrder, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	q := client.RechargeOrder.Query()

	// Filter by status if provided
	if status != "" {
		q = q.Where(rechargeorder.StatusEQ(status))
	}

	// Search by keyword in order_no
	if keyword != "" {
		q = q.Where(rechargeorder.OrderNoContains(keyword))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		WithUser().
		Order(dbent.Desc(rechargeorder.FieldCreatedAt)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return rechargeOrderEntitiesToService(orders), paginationResultFromTotal(int64(total), params), nil
}

func rechargeOrderEntityToService(m *dbent.RechargeOrder) *service.RechargeOrder {
	if m == nil {
		return nil
	}
	out := &service.RechargeOrder{
		ID:           m.ID,
		OrderNo:      m.OrderNo,
		TradeNo:      m.TradeNo,
		UserID:       m.UserID,
		Amount:       m.Amount,
		CreditAmount: m.CreditAmount,
		Multiplier:   m.Multiplier,
		Status:       m.Status,
		PayType:      m.PayType,
		PaidAt:       m.PaidAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		ExpiredAt:    m.ExpiredAt,
	}
	if m.Edges.User != nil {
		out.User = userEntityToService(m.Edges.User)
	}
	return out
}

func rechargeOrderEntitiesToService(models []*dbent.RechargeOrder) []service.RechargeOrder {
	out := make([]service.RechargeOrder, 0, len(models))
	for i := range models {
		if o := rechargeOrderEntityToService(models[i]); o != nil {
			out = append(out, *o)
		}
	}
	return out
}

func applyRechargeOrderEntityToService(dst *service.RechargeOrder, src *dbent.RechargeOrder) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}
