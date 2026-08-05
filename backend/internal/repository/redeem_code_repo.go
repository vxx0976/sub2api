package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/redeemcode"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/lib/pq"
)

type redeemCodeRepository struct {
	client *dbent.Client
	sql    *sql.DB
}

func NewRedeemCodeRepository(client *dbent.Client, sqlDB *sql.DB) service.RedeemCodeRepository {
	return &redeemCodeRepository{client: client, sql: sqlDB}
}

func (r *redeemCodeRepository) Create(ctx context.Context, code *service.RedeemCode) error {
	created, err := r.client.RedeemCode.Create().
		SetCode(code.Code).
		SetType(code.Type).
		SetValue(code.Value).
		SetStatus(code.Status).
		SetNotes(code.Notes).
		SetValidityDays(code.ValidityDays).
		SetNillableExpiresAt(code.ExpiresAt).
		SetNillableUsedBy(code.UsedBy).
		SetNillableUsedAt(code.UsedAt).
		SetNillableGroupID(code.GroupID).
		SetNillableOwnerID(code.OwnerID).
		Save(ctx)
	if err == nil {
		code.ID = created.ID
		code.CreatedAt = created.CreatedAt
	}
	return err
}

func (r *redeemCodeRepository) CreateBatch(ctx context.Context, codes []service.RedeemCode) error {
	if len(codes) == 0 {
		return nil
	}

	builders := make([]*dbent.RedeemCodeCreate, 0, len(codes))
	for i := range codes {
		c := &codes[i]
		b := r.client.RedeemCode.Create().
			SetCode(c.Code).
			SetType(c.Type).
			SetValue(c.Value).
			SetStatus(c.Status).
			SetNotes(c.Notes).
			SetValidityDays(c.ValidityDays).
			SetNillableExpiresAt(c.ExpiresAt).
			SetNillableUsedBy(c.UsedBy).
			SetNillableUsedAt(c.UsedAt).
			SetNillableGroupID(c.GroupID).
			SetNillableOwnerID(c.OwnerID)
		builders = append(builders, b)
	}

	return r.client.RedeemCode.CreateBulk(builders...).Exec(ctx)
}

func (r *redeemCodeRepository) GetByID(ctx context.Context, id int64) (*service.RedeemCode, error) {
	m, err := r.client.RedeemCode.Query().
		Where(redeemcode.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrRedeemCodeNotFound
		}
		return nil, err
	}
	return redeemCodeEntityToService(m), nil
}

func (r *redeemCodeRepository) GetByCode(ctx context.Context, code string) (*service.RedeemCode, error) {
	m, err := r.client.RedeemCode.Query().
		Where(redeemcode.CodeEQ(code)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrRedeemCodeNotFound
		}
		return nil, err
	}
	return redeemCodeEntityToService(m), nil
}

// Delete 无条件按 ID 删除（管理员路径：允许删除 expired/disabled 等任意状态的码；
// 管理员删除不触发退款，无双花风险）。删除不存在的 ID 视为幂等成功——
// DeleteOneID 的 NotFoundError 若原样上抛，管理端重复删除（双击/并发）会变 500。
func (r *redeemCodeRepository) Delete(ctx context.Context, id int64) error {
	if err := clientFromContext(ctx, r.client).RedeemCode.DeleteOneID(id).Exec(ctx); err != nil {
		if dbent.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

// DeleteIfUnused 条件删除：仅当兑换码仍处于 unused 状态时才删除，供商户"删除退款"
// 路径使用，防止与用户兑换并发导致的双倍退款/净亏面值。tx-aware：若 ctx 中带有事务
// 则在事务内执行。返回 0 行（已被兑换或不存在）时返回 ErrRedeemCodeUsed，由上层回滚事务。
func (r *redeemCodeRepository) DeleteIfUnused(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	affected, err := client.RedeemCode.Delete().
		Where(redeemcode.IDEQ(id), redeemcode.StatusEQ(service.StatusUnused)).
		Exec(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrRedeemCodeUsed
	}
	return nil
}

func (r *redeemCodeRepository) List(ctx context.Context, params pagination.PaginationParams) ([]service.RedeemCode, *pagination.PaginationResult, error) {
	return r.ListWithFilters(ctx, params, "", "", "")
}

func (r *redeemCodeRepository) ListWithFilters(ctx context.Context, params pagination.PaginationParams, codeType, status, search string) ([]service.RedeemCode, *pagination.PaginationResult, error) {
	q := r.client.RedeemCode.Query()

	if codeType != "" {
		q = q.Where(redeemcode.TypeEQ(codeType))
	}
	if status != "" {
		now := time.Now()
		switch status {
		case service.StatusExpired:
			q = q.Where(redeemcode.Or(
				redeemcode.StatusEQ(service.StatusExpired),
				redeemcode.And(
					redeemcode.StatusEQ(service.StatusUnused),
					redeemcode.ExpiresAtNotNil(),
					redeemcode.ExpiresAtLTE(now),
				),
			))
		case service.StatusUnused:
			q = q.Where(
				redeemcode.StatusEQ(service.StatusUnused),
				redeemcode.Or(
					redeemcode.ExpiresAtIsNil(),
					redeemcode.ExpiresAtGT(now),
				),
			)
		default:
			q = q.Where(redeemcode.StatusEQ(status))
		}
	}
	if search != "" {
		q = q.Where(
			redeemcode.Or(
				redeemcode.CodeContainsFold(search),
				redeemcode.HasUserWith(user.EmailContainsFold(search)),
			),
		)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	codesQuery := q.
		WithUser().
		WithGroup().
		Offset(params.Offset()).
		Limit(params.Limit())
	for _, order := range redeemCodeListOrder(params) {
		codesQuery = codesQuery.Order(order)
	}

	codes, err := codesQuery.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	outCodes := redeemCodeEntitiesToService(codes)

	return outCodes, paginationResultFromTotal(int64(total), params), nil
}

func redeemCodeListOrder(params pagination.PaginationParams) []func(*entsql.Selector) {
	sortBy := strings.ToLower(strings.TrimSpace(params.SortBy))
	sortOrder := params.NormalizedSortOrder(pagination.SortOrderDesc)

	var field string
	switch sortBy {
	case "type":
		field = redeemcode.FieldType
	case "value":
		field = redeemcode.FieldValue
	case "status":
		field = redeemcode.FieldStatus
	case "used_at":
		field = redeemcode.FieldUsedAt
	case "created_at":
		field = redeemcode.FieldCreatedAt
	case "expires_at":
		field = redeemcode.FieldExpiresAt
	case "code":
		field = redeemcode.FieldCode
	default:
		field = redeemcode.FieldID
	}

	if sortOrder == pagination.SortOrderAsc {
		return []func(*entsql.Selector){dbent.Asc(field), dbent.Asc(redeemcode.FieldID)}
	}
	return []func(*entsql.Selector){dbent.Desc(field), dbent.Desc(redeemcode.FieldID)}
}

func (r *redeemCodeRepository) Update(ctx context.Context, code *service.RedeemCode) error {
	up := r.client.RedeemCode.UpdateOneID(code.ID).
		SetCode(code.Code).
		SetType(code.Type).
		SetValue(code.Value).
		SetStatus(code.Status).
		SetNotes(code.Notes).
		SetValidityDays(code.ValidityDays)

	if code.UsedBy != nil {
		up.SetUsedBy(*code.UsedBy)
	} else {
		up.ClearUsedBy()
	}
	if code.UsedAt != nil {
		up.SetUsedAt(*code.UsedAt)
	} else {
		up.ClearUsedAt()
	}
	if code.GroupID != nil {
		up.SetGroupID(*code.GroupID)
	} else {
		up.ClearGroupID()
	}
	if code.ExpiresAt != nil {
		up.SetExpiresAt(*code.ExpiresAt)
	} else {
		up.ClearExpiresAt()
	}

	updated, err := up.Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrRedeemCodeNotFound
		}
		return err
	}
	code.CreatedAt = updated.CreatedAt
	return nil
}

func (r *redeemCodeRepository) BatchUpdate(ctx context.Context, ids []int64, fields service.RedeemCodeBatchUpdateFields) (int64, error) {
	uniqueIDs := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	if len(uniqueIDs) == 0 {
		return 0, nil
	}

	if tx := dbent.TxFromContext(ctx); tx != nil {
		return r.batchUpdate(ctx, tx.Client(), uniqueIDs, fields)
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return 0, err
	}
	txCtx := dbent.NewTxContext(ctx, tx)
	defer func() { _ = tx.Rollback() }()

	updated, err := r.batchUpdate(txCtx, tx.Client(), uniqueIDs, fields)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return updated, nil
}

func (r *redeemCodeRepository) batchUpdate(ctx context.Context, client *dbent.Client, ids []int64, fields service.RedeemCodeBatchUpdateFields) (int64, error) {
	existing, err := client.RedeemCode.Query().
		Where(redeemcode.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return 0, err
	}
	if len(existing) != len(ids) {
		return 0, service.ErrRedeemCodeNotFound
	}
	if fields.TouchesUsedSensitiveFields() {
		for _, code := range existing {
			if code.Status == service.StatusUsed {
				return 0, service.ErrRedeemCodeUsed
			}
		}
	}

	up := client.RedeemCode.Update().Where(redeemcode.IDIn(ids...))
	if fields.Status != nil {
		up.SetStatus(*fields.Status)
	}
	if fields.Notes != nil {
		up.SetNotes(*fields.Notes)
	}
	if fields.ExpiresAt.Set {
		if fields.ExpiresAt.Value != nil {
			up.SetExpiresAt(*fields.ExpiresAt.Value)
		} else {
			up.ClearExpiresAt()
		}
	}
	if fields.GroupID.Set {
		if fields.GroupID.Value != nil {
			up.SetGroupID(*fields.GroupID.Value)
		} else {
			up.ClearGroupID()
		}
	}

	affected, err := up.Save(ctx)
	if err != nil {
		return 0, err
	}
	if affected != len(ids) {
		return 0, service.ErrRedeemCodeNotFound
	}
	return int64(affected), nil
}

func (r *redeemCodeRepository) Use(ctx context.Context, id, userID int64) error {
	now := time.Now()
	client := clientFromContext(ctx, r.client)
	affected, err := client.RedeemCode.Update().
		Where(redeemcode.IDEQ(id), redeemcode.StatusEQ(service.StatusUnused)).
		SetStatus(service.StatusUsed).
		SetUsedBy(userID).
		SetUsedAt(now).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrRedeemCodeUsed
	}
	return nil
}

func (r *redeemCodeRepository) ListByUser(ctx context.Context, userID int64, limit int) ([]service.RedeemCode, error) {
	if limit <= 0 {
		limit = 10
	}

	codes, err := r.client.RedeemCode.Query().
		Where(redeemcode.UsedByEQ(userID)).
		WithGroup().
		Order(dbent.Desc(redeemcode.FieldUsedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return redeemCodeEntitiesToService(codes), nil
}

// ListByUserPaginated returns paginated balance/concurrency history for a user.
// Supports optional type filter (e.g. "balance", "admin_balance", "concurrency", "admin_concurrency", "subscription").
func (r *redeemCodeRepository) ListByUserPaginated(ctx context.Context, userID int64, params pagination.PaginationParams, codeType string) ([]service.RedeemCode, *pagination.PaginationResult, error) {
	q := r.client.RedeemCode.Query().
		Where(redeemcode.UsedByEQ(userID))

	// Optional type filter
	if codeType != "" {
		q = q.Where(redeemcode.TypeEQ(codeType))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	codes, err := q.
		WithGroup().
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(redeemcode.FieldUsedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return redeemCodeEntitiesToService(codes), paginationResultFromTotal(int64(total), params), nil
}

// manualBalanceShadowPrefixes 是 adminService.UpdateUserBalance / RefundUserBalance 为每笔充值/退款订单
// 成功时自动写入的 type='admin_balance' 审计影子记录的 notes 前缀（见 recharge_service.go / order_service.go /
// usdt_order_service.go 中对应的 fmt.Sprintf）。列"手工调整"时必须排除这些影子，否则每笔真实订单都会再多出
// 一条重复的"手工"行并把分页 total 翻倍——仅保留管理员后台直接加/扣余额的真实记录。
var manualBalanceShadowPrefixes = []string{
	"Recharge order ",        // recharge_service.go 充值到账
	"AliMPay order ",         // order_service.go 充值到账
	"USDT order ",            // usdt_order_service.go 充值到账
	"Refund recharge order ", // recharge_service.go 退款回冲
	"Refund AliMPay order ",  // order_service.go 退款回冲
	"Refund USDT order ",     // usdt_order_service.go 退款回冲
}

// ListManualBalanceAdjustments 见 service.RedeemCodeRepository 接口注释。
func (r *redeemCodeRepository) ListManualBalanceAdjustments(ctx context.Context, userID *int64, limit int) ([]service.RedeemCode, int64, error) {
	if limit <= 0 {
		limit = 20
	}

	q := r.client.RedeemCode.Query().
		Where(redeemcode.TypeEQ(service.AdjustmentTypeAdminBalance))
	if userID != nil {
		q = q.Where(redeemcode.UsedByEQ(*userID))
	}
	// 排除 audit shadow：notes 为 NULL（真实手工调整可能不带备注）始终保留，否则要求不以任一影子前缀开头。
	for _, prefix := range manualBalanceShadowPrefixes {
		q = q.Where(redeemcode.Or(
			redeemcode.NotesIsNil(),
			redeemcode.Not(redeemcode.NotesHasPrefix(prefix)),
		))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	codes, err := q.
		WithGroup().
		Order(dbent.Desc(redeemcode.FieldUsedAt), dbent.Desc(redeemcode.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return redeemCodeEntitiesToService(codes), int64(total), nil
}

// ListByOwnerID returns paginated balance redeem codes owned by the given reseller.
func (r *redeemCodeRepository) ListByOwnerID(ctx context.Context, ownerID int64, params pagination.PaginationParams) ([]service.RedeemCode, *pagination.PaginationResult, error) {
	q := r.client.RedeemCode.Query().
		Where(
			redeemcode.OwnerIDEQ(ownerID),
			redeemcode.TypeEQ(service.RedeemTypeBalance),
		)

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	codes, err := q.
		WithUser().
		WithGroup().
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(redeemcode.FieldID)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return redeemCodeEntitiesToService(codes), paginationResultFromTotal(int64(total), params), nil
}

// SumPositiveBalanceByUser returns total recharged amount (sum of value > 0 where type is balance/admin_balance).
func (r *redeemCodeRepository) SumPositiveBalanceByUser(ctx context.Context, userID int64) (float64, error) {
	var result []struct {
		Sum float64 `json:"sum"`
	}
	err := r.client.RedeemCode.Query().
		Where(
			redeemcode.UsedByEQ(userID),
			redeemcode.ValueGT(0),
			redeemcode.TypeIn("balance", "admin_balance"),
		).
		Aggregate(dbent.As(dbent.Sum(redeemcode.FieldValue), "sum")).
		Scan(ctx, &result)
	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, nil
	}
	return result[0].Sum, nil
}

// SumPositiveValueByDayForTypes 按时区分桶汇总指定 type 列表中、value > 0、status='used'
// 且 owner_id IS NULL（仅平台码）的兑换码 value 总和。常用于把"通过平台兑换码加余额"
// 并入资金流入曲线；商户自费码已计入其它口径，此处排除以免双计。
// 返回的 map key 为 "YYYY-MM-DD"（按 tzName 解释 used_at），value 为当日 value 合计（USD）。
func (r *redeemCodeRepository) SumPositiveValueByDayForTypes(ctx context.Context, startTime, endTime time.Time, tzName string, types []string) (map[string]float64, error) {
	if r.sql == nil || len(types) == 0 {
		return map[string]float64{}, nil
	}
	if tzName == "" {
		tzName = "UTC"
	}
	// 用 ANY($4) 安全地传入字符串数组，避免拼 SQL。
	// owner_id IS NULL：只统计平台码作为"资金流入"。商户自费生成的码（owner_id 非空）
	// 其资金已计入 admin_manual / 充值口径，再计入此处会被双计。
	query := `
		SELECT
			TO_CHAR(used_at AT TIME ZONE $3, 'YYYY-MM-DD') AS day,
			COALESCE(SUM(value), 0) AS total
		FROM redeem_codes
		WHERE status = 'used'
		  AND used_at IS NOT NULL
		  AND used_at >= $1
		  AND used_at < $2
		  AND value > 0
		  AND owner_id IS NULL
		  AND type = ANY($4)
		GROUP BY 1
		ORDER BY 1
	`
	rows, err := r.sql.QueryContext(ctx, query, startTime, endTime, tzName, pq.Array(types))
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

// SumManualAdminBalanceByDay 按时区分桶汇总管理员"真实"手工加余额（type='admin_balance' / value>0）。
// 由于 adminService.UpdateUserBalance / RefundUserBalance 会在每笔充值/退款订单成功时也写一条
// type='admin_balance' 的审计影子记录（notes 以 manualBalanceShadowPrefixes 中的前缀开头，含
// AliMPay / Recharge / USDT 及各自 Refund），这里通过 notes 前缀过滤掉这些 audit shadow，
// 仅保留管理员从后台直接调整的真实记录。
func (r *redeemCodeRepository) SumManualAdminBalanceByDay(ctx context.Context, startTime, endTime time.Time, tzName string) (map[string]float64, error) {
	if r.sql == nil {
		return map[string]float64{}, nil
	}
	if tzName == "" {
		tzName = "UTC"
	}
	// 排除 audit shadow：每笔充值/退款订单成功时都会写一条 type='admin_balance' 的影子记录，
	// notes 以 manualBalanceShadowPrefixes 中某个前缀开头。这里用共享前缀表动态拼 NOT LIKE，
	// 与 ListManualBalanceAdjustments 保持同源，避免新增支付通道时再漏（历史漏了 'USDT order '）。
	args := []any{startTime, endTime, tzName}
	var notLike strings.Builder
	for _, prefix := range manualBalanceShadowPrefixes {
		args = append(args, prefix+"%")
		fmt.Fprintf(&notLike, "\n\t\t  AND COALESCE(notes, '') NOT LIKE $%d", len(args))
	}
	query := fmt.Sprintf(`
		SELECT
			TO_CHAR(used_at AT TIME ZONE $3, 'YYYY-MM-DD') AS day,
			COALESCE(SUM(value), 0) AS total
		FROM redeem_codes
		WHERE type = 'admin_balance'
		  AND status = 'used'
		  AND used_at IS NOT NULL
		  AND used_at >= $1
		  AND used_at < $2
		  AND value > 0%s
		GROUP BY 1
		ORDER BY 1
	`, notLike.String())
	rows, err := r.sql.QueryContext(ctx, query, args...)
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

func redeemCodeEntityToService(m *dbent.RedeemCode) *service.RedeemCode {
	if m == nil {
		return nil
	}
	out := &service.RedeemCode{
		ID:           m.ID,
		Code:         m.Code,
		Type:         m.Type,
		Value:        m.Value,
		Status:       m.Status,
		UsedBy:       m.UsedBy,
		UsedAt:       m.UsedAt,
		Notes:        derefString(m.Notes),
		CreatedAt:    m.CreatedAt,
		ExpiresAt:    m.ExpiresAt,
		GroupID:      m.GroupID,
		ValidityDays: m.ValidityDays,
		OwnerID:      m.OwnerID,
	}
	if m.Edges.User != nil {
		out.User = userEntityToService(m.Edges.User)
	}
	if m.Edges.Group != nil {
		out.Group = groupEntityToService(m.Edges.Group)
	}
	return out
}

func redeemCodeEntitiesToService(models []*dbent.RedeemCode) []service.RedeemCode {
	out := make([]service.RedeemCode, 0, len(models))
	for i := range models {
		if s := redeemCodeEntityToService(models[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
