package repository

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/rechargeorder"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type rechargeOrderRepo struct {
	client *ent.Client
}

func NewRechargeOrderRepo(client *ent.Client) service.RechargeOrderRepository {
	return &rechargeOrderRepo{client: client}
}

func (r *rechargeOrderRepo) Create(ctx context.Context, order *service.RechargeOrder) error {
	builder := r.client.RechargeOrder.Create().
		SetOrderNo(order.OrderNo).
		SetUserID(order.UserID).
		SetAmount(order.Amount).
		SetCreditAmount(order.CreditAmount).
		SetMultiplier(order.Multiplier).
		SetStatus(order.Status)

	if order.PayType != "" {
		builder.SetPayType(order.PayType)
	}
	if order.ExpiredAt != nil {
		builder.SetExpiredAt(*order.ExpiredAt)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	order.ID = created.ID
	order.CreatedAt = created.CreatedAt
	order.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *rechargeOrderRepo) GetByOrderNo(ctx context.Context, orderNo string) (*service.RechargeOrder, error) {
	row, err := r.client.RechargeOrder.Query().
		Where(rechargeorder.OrderNoEQ(orderNo)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toServiceRechargeOrder(row), nil
}

func (r *rechargeOrderRepo) UpdateStatus(ctx context.Context, orderNo, fromStatus, toStatus string, tradeNo *string, paidAt *time.Time) error {
	builder := r.client.RechargeOrder.Update().
		Where(
			rechargeorder.OrderNoEQ(orderNo),
			rechargeorder.StatusEQ(fromStatus),
		).
		SetStatus(toStatus).
		SetUpdatedAt(time.Now())

	if tradeNo != nil {
		builder.SetTradeNo(*tradeNo)
	}
	if paidAt != nil {
		builder.SetPaidAt(*paidAt)
	}

	n, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return service.ErrRechargeOrderStatusConflict
	}
	return nil
}

func (r *rechargeOrderRepo) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*service.RechargeOrder, int, error) {
	query := r.client.RechargeOrder.Query().
		Where(rechargeorder.UserIDEQ(userID)).
		Order(ent.Desc(rechargeorder.FieldCreatedAt))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := query.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	orders := make([]*service.RechargeOrder, len(rows))
	for i, row := range rows {
		orders[i] = toServiceRechargeOrder(row)
	}
	return orders, total, nil
}

func (r *rechargeOrderRepo) ExpirePendingOrders(ctx context.Context) (int, error) {
	n, err := r.client.RechargeOrder.Update().
		Where(
			rechargeorder.StatusEQ("pending"),
			rechargeorder.ExpiredAtLT(time.Now()),
		).
		SetStatus("expired").
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return n, err
}

func (r *rechargeOrderRepo) ListAll(ctx context.Context, status string, userID *int64, limit, offset int) ([]*service.RechargeOrder, int, error) {
	query := r.client.RechargeOrder.Query().
		Order(ent.Desc(rechargeorder.FieldCreatedAt))

	if status != "" {
		query = query.Where(rechargeorder.StatusEQ(status))
	}
	if userID != nil {
		query = query.Where(rechargeorder.UserIDEQ(*userID))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := query.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	orders := make([]*service.RechargeOrder, len(rows))
	for i, row := range rows {
		orders[i] = toServiceRechargeOrder(row)
	}
	return orders, total, nil
}

func toServiceRechargeOrder(row *ent.RechargeOrder) *service.RechargeOrder {
	o := &service.RechargeOrder{
		ID:           row.ID,
		OrderNo:      row.OrderNo,
		UserID:       row.UserID,
		Amount:       row.Amount,
		CreditAmount: row.CreditAmount,
		Multiplier:   row.Multiplier,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if row.TradeNo != nil {
		o.TradeNo = *row.TradeNo
	}
	if row.PayType != nil {
		o.PayType = *row.PayType
	}
	if row.PaidAt != nil {
		o.PaidAt = row.PaidAt
	}
	if row.ExpiredAt != nil {
		o.ExpiredAt = row.ExpiredAt
	}
	return o
}
