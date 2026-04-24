package repository

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/order"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type orderRepo struct {
	client *ent.Client
}

func NewOrderRepo(client *ent.Client) service.OrderRepository {
	return &orderRepo{client: client}
}

func (r *orderRepo) Create(ctx context.Context, o *service.Order) error {
	builder := r.client.Order.Create().
		SetOrderNo(o.OrderNo).
		SetUserID(o.UserID).
		SetAmount(o.Amount).
		SetPaymentAmount(o.PaymentAmount).
		SetCreditAmount(o.CreditAmount).
		SetMultiplier(o.Multiplier).
		SetStatus(o.Status)

	if o.TradeNo != "" {
		builder.SetTradeNo(o.TradeNo)
	}
	if o.PayType != "" {
		builder.SetPayType(o.PayType)
	}
	if o.SourceDomain != "" {
		builder.SetSourceDomain(o.SourceDomain)
	}
	if o.GroupID != 0 {
		builder.SetGroupID(o.GroupID)
	}
	if o.ExpiredAt != nil {
		builder.SetExpiredAt(*o.ExpiredAt)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	o.ID = created.ID
	o.CreatedAt = created.CreatedAt
	o.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *orderRepo) GetByOrderNo(ctx context.Context, orderNo string) (*service.Order, error) {
	row, err := r.client.Order.Query().
		Where(order.OrderNoEQ(orderNo)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toServiceOrder(row), nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, orderNo, fromStatus, toStatus string, tradeNo *string, paidAt *time.Time) error {
	builder := r.client.Order.Update().
		Where(
			order.OrderNoEQ(orderNo),
			order.StatusEQ(fromStatus),
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
		return service.ErrOrderStatusConflict
	}
	return nil
}

func (r *orderRepo) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*service.Order, int, error) {
	query := r.client.Order.Query().
		Where(order.UserIDEQ(userID)).
		Order(ent.Desc(order.FieldCreatedAt))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := query.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	out := make([]*service.Order, len(rows))
	for i, row := range rows {
		out[i] = toServiceOrder(row)
	}
	return out, total, nil
}

func (r *orderRepo) ExpirePendingOrders(ctx context.Context) (int, error) {
	n, err := r.client.Order.Update().
		Where(
			order.StatusEQ("pending"),
			order.ExpiredAtLT(time.Now()),
		).
		SetStatus("expired").
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return n, err
}

func (r *orderRepo) ListAll(ctx context.Context, status string, userID *int64, limit, offset int) ([]*service.Order, int, error) {
	query := r.client.Order.Query().
		Order(ent.Desc(order.FieldCreatedAt))

	if status != "" {
		query = query.Where(order.StatusEQ(status))
	}
	if userID != nil {
		query = query.Where(order.UserIDEQ(*userID))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := query.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	out := make([]*service.Order, len(rows))
	for i, row := range rows {
		out[i] = toServiceOrder(row)
	}
	return out, total, nil
}

func (r *orderRepo) ListPending(ctx context.Context) ([]*service.Order, error) {
	rows, err := r.client.Order.Query().
		Where(order.StatusEQ("pending")).
		Order(ent.Desc(order.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*service.Order, len(rows))
	for i, row := range rows {
		out[i] = toServiceOrder(row)
	}
	return out, nil
}

func (r *orderRepo) SumPaidCreditByUserIDs(ctx context.Context, userIDs []int64) (float64, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	var result []struct {
		Sum float64 `json:"sum"`
	}
	err := r.client.Order.Query().
		Where(order.StatusEQ("paid"), order.UserIDIn(userIDs...)).
		Aggregate(ent.As(ent.Sum(order.FieldCreditAmount), "sum")).
		Scan(ctx, &result)
	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, nil
	}
	return result[0].Sum, nil
}

func (r *orderRepo) ListPaidByUserIDs(ctx context.Context, userIDs []int64, limit, offset int) ([]*service.Order, int, error) {
	if len(userIDs) == 0 {
		return nil, 0, nil
	}
	query := r.client.Order.Query().
		Where(order.StatusEQ("paid"), order.UserIDIn(userIDs...)).
		Order(ent.Desc(order.FieldCreatedAt))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := query.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	out := make([]*service.Order, len(rows))
	for i, row := range rows {
		out[i] = toServiceOrder(row)
	}
	return out, total, nil
}

func toServiceOrder(row *ent.Order) *service.Order {
	o := &service.Order{
		ID:            row.ID,
		OrderNo:       row.OrderNo,
		UserID:        row.UserID,
		Amount:        row.Amount,
		PaymentAmount: row.PaymentAmount,
		CreditAmount:  row.CreditAmount,
		Multiplier:    row.Multiplier,
		Status:        row.Status,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	if row.TradeNo != nil {
		o.TradeNo = *row.TradeNo
	}
	if row.PayType != nil {
		o.PayType = *row.PayType
	}
	if row.GroupID != nil {
		o.GroupID = *row.GroupID
	}
	if row.PaidAt != nil {
		o.PaidAt = row.PaidAt
	}
	if row.SourceDomain != nil {
		o.SourceDomain = *row.SourceDomain
	}
	if row.ExpiredAt != nil {
		o.ExpiredAt = row.ExpiredAt
	}
	return o
}
