package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/usdtorder"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type usdtOrderRepo struct {
	client *ent.Client
	sql    *sql.DB
}

// NewUsdtOrderRepo 创建 USDT 订单仓储。
func NewUsdtOrderRepo(client *ent.Client, sqlDB *sql.DB) service.UsdtOrderRepository {
	return &usdtOrderRepo{client: client, sql: sqlDB}
}

func (r *usdtOrderRepo) CreateWithUniqueUsdtAmount(ctx context.Context, o *service.UsdtOrder, baseUsdt, amountOffset float64, reuseWindow time.Duration) error {
	if amountOffset <= 0 {
		amountOffset = 0.0001
	}
	if reuseWindow <= 0 {
		reuseWindow = 30 * time.Minute
	}

	const maxAttempts = 3
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		tx, err := r.client.Tx(ctx)
		if err != nil {
			return err
		}
		err = createWithUniqueUsdtAmountTx(ctx, tx.Client(), o, baseUsdt, amountOffset, reuseWindow)
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
		return fmt.Errorf("allocate usdt amount: %w", lastErr)
	}
	return service.ErrUsdtAmountUnavailable
}

func createWithUniqueUsdtAmountTx(ctx context.Context, client *ent.Client, o *service.UsdtOrder, baseUsdt, amountOffset float64, reuseWindow time.Duration) error {
	candidates := usdtAmountCandidates(baseUsdt, amountOffset, 100)
	cutoff := time.Now().Add(-reuseWindow)

	rows, err := client.UsdtOrder.Query().
		Where(
			usdtorder.ChainEQ(o.Chain),
			usdtorder.UsdtAmountIn(candidates...),
			usdtorder.Or(
				usdtorder.StatusEQ("pending"),
				usdtorder.And(
					usdtorder.StatusEQ("expired"),
					usdtorder.ExpiredAtGT(cutoff),
				),
			),
		).
		All(ctx)
	if err != nil {
		return err
	}

	used := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		used[formatUsdtAmount(row.UsdtAmount)] = struct{}{}
	}
	for _, candidate := range candidates {
		if _, ok := used[formatUsdtAmount(candidate)]; ok {
			continue
		}
		o.UsdtAmount = candidate
		return createUsdtOrder(ctx, client, o)
	}
	return service.ErrUsdtAmountUnavailable
}

func usdtAmountCandidates(baseUsdt, amountOffset float64, limit int) []float64 {
	baseAtomic := int64(math.Round(baseUsdt * 1_000_000))
	step := int64(math.Round(amountOffset * 1_000_000))
	if step < 1 {
		step = 1
	}
	out := make([]float64, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, float64(baseAtomic+int64(i)*step)/1_000_000)
	}
	return out
}

func formatUsdtAmount(amount float64) string {
	return fmt.Sprintf("%.6f", amount)
}

func createUsdtOrder(ctx context.Context, client *ent.Client, o *service.UsdtOrder) error {
	builder := client.UsdtOrder.Create().
		SetOrderNo(o.OrderNo).
		SetUserID(o.UserID).
		SetAmount(o.Amount).
		SetCreditAmount(o.CreditAmount).
		SetMultiplier(o.Multiplier).
		SetChain(o.Chain).
		SetReceivingAddress(o.ReceivingAddress).
		SetUsdtRate(o.UsdtRate).
		SetUsdtAmount(o.UsdtAmount).
		SetStatus(o.Status)

	if o.PayType != "" {
		builder.SetPayType(o.PayType)
	}
	if o.SourceDomain != "" {
		builder.SetSourceDomain(o.SourceDomain)
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

func (r *usdtOrderRepo) GetByOrderNo(ctx context.Context, orderNo string) (*service.UsdtOrder, error) {
	row, err := r.client.UsdtOrder.Query().
		Where(usdtorder.OrderNoEQ(orderNo)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toServiceUsdtOrder(row), nil
}

func (r *usdtOrderRepo) UpdateStatus(ctx context.Context, orderNo, fromStatus, toStatus string, paidAt *time.Time) error {
	builder := r.client.UsdtOrder.Update().
		Where(
			usdtorder.OrderNoEQ(orderNo),
			usdtorder.StatusEQ(fromStatus),
		).
		SetStatus(toStatus).
		SetUpdatedAt(time.Now())
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

func (r *usdtOrderRepo) MarkPaid(ctx context.Context, orderNo, fromStatus string, upd service.UsdtPaidUpdate, paidAt time.Time) error {
	builder := r.client.UsdtOrder.Update().
		Where(
			usdtorder.OrderNoEQ(orderNo),
			usdtorder.StatusEQ(fromStatus),
		).
		SetStatus("paid").
		SetPaidAt(paidAt).
		SetUpdatedAt(time.Now())
	if upd.TxHash != "" {
		builder.SetTradeNo(upd.TxHash)
	}
	if upd.FromAddress != "" {
		builder.SetFromAddress(upd.FromAddress)
	}
	if upd.PaidUsdtAmount > 0 {
		builder.SetPaidUsdtAmount(upd.PaidUsdtAmount)
	}
	if upd.BlockNumber != nil {
		builder.SetBlockNumber(*upd.BlockNumber)
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

func (r *usdtOrderRepo) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*service.UsdtOrder, int, error) {
	query := r.client.UsdtOrder.Query().
		Where(usdtorder.UserIDEQ(userID)).
		Order(ent.Desc(usdtorder.FieldCreatedAt))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := query.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*service.UsdtOrder, len(rows))
	for i, row := range rows {
		out[i] = toServiceUsdtOrder(row)
	}
	return out, total, nil
}

func (r *usdtOrderRepo) ExpirePendingOrders(ctx context.Context) (int, error) {
	n, err := r.client.UsdtOrder.Update().
		Where(
			usdtorder.StatusEQ("pending"),
			usdtorder.ExpiredAtLT(time.Now()),
		).
		SetStatus("expired").
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return n, err
}

func (r *usdtOrderRepo) ListAll(ctx context.Context, status string, userID *int64, limit, offset int) ([]*service.UsdtOrder, int, error) {
	query := r.client.UsdtOrder.Query().
		Order(ent.Desc(usdtorder.FieldCreatedAt))
	if status != "" {
		query = query.Where(usdtorder.StatusEQ(status))
	}
	if userID != nil {
		query = query.Where(usdtorder.UserIDEQ(*userID))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := query.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*service.UsdtOrder, len(rows))
	for i, row := range rows {
		out[i] = toServiceUsdtOrder(row)
	}
	return out, total, nil
}

func (r *usdtOrderRepo) ListPending(ctx context.Context) ([]*service.UsdtOrder, error) {
	rows, err := r.client.UsdtOrder.Query().
		Where(usdtorder.StatusEQ("pending")).
		Order(ent.Desc(usdtorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*service.UsdtOrder, len(rows))
	for i, row := range rows {
		out[i] = toServiceUsdtOrder(row)
	}
	return out, nil
}

// ListMatchable 返回 watcher 可匹配的订单：pending + 宽限期内刚过期(expired 且 expired_at > graceCutoff)。
// 让"临近/略超截止才到账"的订单仍能被链上匹配并补入账，避免孤儿单。
func (r *usdtOrderRepo) ListMatchable(ctx context.Context, graceCutoff time.Time) ([]*service.UsdtOrder, error) {
	rows, err := r.client.UsdtOrder.Query().
		Where(
			usdtorder.Or(
				usdtorder.StatusEQ("pending"),
				usdtorder.And(
					usdtorder.StatusEQ("expired"),
					usdtorder.ExpiredAtGT(graceCutoff),
				),
			),
		).
		Order(ent.Desc(usdtorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*service.UsdtOrder, len(rows))
	for i, row := range rows {
		out[i] = toServiceUsdtOrder(row)
	}
	return out, nil
}

// SumPaidCreditByUserIDs 汇总指定用户集合 status='paid' 的 credit_amount（供经销商佣金基数）。
func (r *usdtOrderRepo) SumPaidCreditByUserIDs(ctx context.Context, userIDs []int64) (float64, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	var result []struct {
		Sum float64 `json:"sum"`
	}
	err := r.client.UsdtOrder.Query().
		Where(usdtorder.StatusEQ("paid"), usdtorder.UserIDIn(userIDs...)).
		Aggregate(ent.As(ent.Sum(usdtorder.FieldCreditAmount), "sum")).
		Scan(ctx, &result)
	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, nil
	}
	return result[0].Sum, nil
}

// ListPaidByUserIDs 分页列出指定用户集合 status='paid' 的订单（供佣金充值明细）。
func (r *usdtOrderRepo) ListPaidByUserIDs(ctx context.Context, userIDs []int64, limit, offset int) ([]*service.UsdtOrder, int, error) {
	if len(userIDs) == 0 {
		return nil, 0, nil
	}
	query := r.client.UsdtOrder.Query().
		Where(usdtorder.StatusEQ("paid"), usdtorder.UserIDIn(userIDs...)).
		Order(ent.Desc(usdtorder.FieldCreatedAt))
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := query.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*service.UsdtOrder, len(rows))
	for i, row := range rows {
		out[i] = toServiceUsdtOrder(row)
	}
	return out, total, nil
}

// SumPaidCreditByDay 汇总指定时区下每日 status='paid' 的 credit_amount（供仪表盘财务趋势）。
// 与 orders/recharge_orders 的同名方法口径一致：按 paid_at 分桶，缺 paid_at 的不计。
func (r *usdtOrderRepo) SumPaidCreditByDay(ctx context.Context, startTime, endTime time.Time, tzName string) (map[string]float64, error) {
	if r.sql == nil {
		return map[string]float64{}, nil
	}
	if tzName == "" {
		tzName = "UTC"
	}
	query := `
		SELECT
			TO_CHAR(paid_at AT TIME ZONE $3, 'YYYY-MM-DD') AS day,
			COALESCE(SUM(credit_amount), 0) AS total
		FROM usdt_orders
		WHERE status = 'paid'
		  AND paid_at IS NOT NULL
		  AND paid_at >= $1
		  AND paid_at < $2
		GROUP BY 1
		ORDER BY 1
	`
	rows, err := r.sql.QueryContext(ctx, query, startTime, endTime, tzName)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make(map[string]float64)
	for rows.Next() {
		var day string
		var total float64
		if err := rows.Scan(&day, &total); err != nil {
			return nil, err
		}
		result[day] = total
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func toServiceUsdtOrder(row *ent.UsdtOrder) *service.UsdtOrder {
	o := &service.UsdtOrder{
		ID:               row.ID,
		OrderNo:          row.OrderNo,
		UserID:           row.UserID,
		Amount:           row.Amount,
		CreditAmount:     row.CreditAmount,
		Multiplier:       row.Multiplier,
		Chain:            row.Chain,
		ReceivingAddress: row.ReceivingAddress,
		UsdtRate:         row.UsdtRate,
		UsdtAmount:       row.UsdtAmount,
		Status:           row.Status,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
	if row.TradeNo != nil {
		o.TradeNo = *row.TradeNo
	}
	if row.PayType != nil {
		o.PayType = *row.PayType
	}
	if row.PaidUsdtAmount != nil {
		o.PaidUsdtAmount = row.PaidUsdtAmount
	}
	if row.FromAddress != nil {
		o.FromAddress = *row.FromAddress
	}
	if row.BlockNumber != nil {
		o.BlockNumber = row.BlockNumber
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
