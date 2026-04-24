package repository

import (
	"context"
	"fmt"
	"math"
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
	return createOrder(ctx, r.client, o)
}

func (r *orderRepo) CreateWithUniquePaymentAmount(ctx context.Context, o *service.Order, baseAmount float64, amountOffset float64, reuseWindow time.Duration) error {
	if amountOffset <= 0 {
		amountOffset = 0.01
	}
	if reuseWindow <= 0 {
		reuseWindow = 2 * time.Hour
	}

	const maxAttempts = 3
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		tx, err := r.client.Tx(ctx)
		if err != nil {
			return err
		}

		err = createWithUniquePaymentAmountTx(ctx, tx.Client(), o, baseAmount, amountOffset, reuseWindow)
		if err != nil {
			_ = tx.Rollback()
			if ent.IsConstraintError(err) {
				lastErr = err
				continue
			}
			return err
		}
		if err := tx.Commit(); err != nil {
			if ent.IsConstraintError(err) {
				lastErr = err
				continue
			}
			return err
		}
		return nil
	}
	if lastErr != nil {
		return fmt.Errorf("allocate payment amount: %w", lastErr)
	}
	return service.ErrOrderPaymentAmountUnavailable
}

func createWithUniquePaymentAmountTx(ctx context.Context, client *ent.Client, o *service.Order, baseAmount float64, amountOffset float64, reuseWindow time.Duration) error {
	candidates := paymentAmountCandidates(baseAmount, amountOffset, 100)
	cutoff := time.Now().Add(-reuseWindow)

	rows, err := client.Order.Query().
		Where(
			order.PaymentAmountIn(candidates...),
			order.Or(
				order.StatusEQ("pending"),
				order.And(
					order.StatusEQ("expired"),
					order.ExpiredAtGT(cutoff),
				),
			),
		).
		All(ctx)
	if err != nil {
		return err
	}

	used := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		used[formatPaymentAmount(row.PaymentAmount)] = struct{}{}
	}
	for _, candidate := range candidates {
		if _, ok := used[formatPaymentAmount(candidate)]; ok {
			continue
		}
		o.PaymentAmount = candidate
		// 到账金额 = 实际支付金额：用户因金额偏移多付的部分（比如 +0.01）也一并入账，不让用户亏
		o.CreditAmount = candidate
		return createOrder(ctx, client, o)
	}
	return service.ErrOrderPaymentAmountUnavailable
}

func paymentAmountCandidates(baseAmount float64, amountOffset float64, limit int) []float64 {
	baseCents := int64(math.Round(baseAmount * 100))
	stepCents := int64(math.Round(amountOffset * 100))
	if stepCents < 1 {
		stepCents = 1
	}
	out := make([]float64, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, float64(baseCents+int64(i)*stepCents)/100)
	}
	return out
}

func formatPaymentAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

func createOrder(ctx context.Context, client *ent.Client, o *service.Order) error {
	builder := client.Order.Create().
		SetOrderNo(o.OrderNo).
		SetUserID(o.UserID).
		SetAmount(o.Amount).
		SetCreditAmount(o.CreditAmount).
		SetMultiplier(o.Multiplier).
		SetStatus(o.Status)

	// transfer 模式无需唯一金额匹配；payment_amount 留 NULL 避免命中唯一 partial index
	if o.PaymentAmount > 0 {
		builder.SetPaymentAmount(o.PaymentAmount)
	}

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
