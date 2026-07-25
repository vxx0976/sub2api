package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
	gocache "github.com/patrickmn/go-cache"
)

const rawUsageLogModelColumn = "model"

// rawUsageLogModelColumn preserves the exact stored usage_logs.model semantics for direct filters.
// Historical rows may contain upstream/billing model values, while newer rows store requested_model.
// Requested/upstream/mapping analytics must use resolveModelDimensionExpression instead.

// usageLogSuccessFilterUL 用于把"失败请求 usage log"（tokens=0、cost=0、不计费的占位记录）
// 从统计性聚合中排除，避免污染 Dashboard / 用量拆分等指标。
//
// schema 中没有 success bool 列；新增列要做迁移，风险大；这里用 actual_cost > 0 作为代理：
// 任何成功落账的请求都会产生 actual_cost（包括 token 计费、纯图片 token 计费、按次/按图计费），
// 反之 failed-request usage log 的 actual_cost 为 0。
// 早期版本用 4 项 token 和 > 0 判定会把"按次/按图计费"与"image_output_tokens 独立计费"的纯图片
// 请求误判为失败，导致这部分请求从用量统计里消失，故改用 actual_cost。
// 配合 `FROM usage_logs ul` JOIN 查询使用。
const usageLogSuccessFilterUL = "ul.actual_cost > 0"

// usageLogEffectivePlatformExpr 用于按"有效平台"维度聚合 usage_logs：
// 优先取请求实际走的分组 platform，若分组未设置 platform 再 fallback 到 account.platform。
// Composite groups are a routing layer, so platform analytics must use the
// resolved concrete account platform instead of grouping spend under "composite".
// 配套要求查询里 LEFT JOIN groups g ON g.id = ul.group_id 与 LEFT JOIN accounts a ON a.id = ul.account_id。
const usageLogEffectivePlatformExpr = "CASE WHEN g.platform = 'composite' THEN a.platform ELSE COALESCE(NULLIF(g.platform,''), a.platform) END"

// dateFormatWhitelist 将 granularity 参数映射为 PostgreSQL TO_CHAR 格式字符串，防止外部输入直接拼入 SQL
var dateFormatWhitelist = map[string]string{
	"hour":  "YYYY-MM-DD HH24:00",
	"day":   "YYYY-MM-DD",
	"week":  "IYYY-IW",
	"month": "YYYY-MM",
}

// safeDateFormat 根据白名单获取 dateFormat，未匹配时返回默认值
func safeDateFormat(granularity string) string {
	if f, ok := dateFormatWhitelist[granularity]; ok {
		return f
	}
	return "YYYY-MM-DD"
}

// appendRawUsageLogModelWhereCondition keeps direct model filters on the raw model column for backward
// compatibility with historical rows. Requested/upstream analytics must use
// resolveModelDimensionExpression instead.
func appendRawUsageLogModelWhereCondition(conditions []string, args []any, model string) ([]string, []any) {
	if strings.TrimSpace(model) == "" {
		return conditions, args
	}
	conditions = append(conditions, fmt.Sprintf("%s = $%d", rawUsageLogModelColumn, len(args)+1))
	args = append(args, model)
	return conditions, args
}

func appendUsageLogBillingModeWhereCondition(conditions []string, args []any, billingMode string) ([]string, []any) {
	return appendUsageLogBillingModeWhereConditionWithAlias(conditions, args, billingMode, "")
}

func appendUsageLogBillingModeWhereConditionWithAlias(conditions []string, args []any, billingMode string, alias string) ([]string, []any) {
	mode := strings.TrimSpace(billingMode)
	if mode == "" {
		return conditions, args
	}
	column := func(name string) string {
		if alias == "" {
			return name
		}
		return alias + "." + name
	}
	placeholder := fmt.Sprintf("$%d", len(args)+1)
	switch service.BillingMode(mode) {
	case service.BillingModeImage:
		conditions = append(conditions, fmt.Sprintf("(%s = %s OR ((%s IS NULL OR %s = '') AND COALESCE(%s, 0) > 0))", column("billing_mode"), placeholder, column("billing_mode"), column("billing_mode"), column("image_count")))
	case service.BillingModeVideo:
		conditions = append(conditions, fmt.Sprintf("%s = %s", column("billing_mode"), placeholder))
	case service.BillingModeToken:
		conditions = append(conditions, fmt.Sprintf("(%s = %s OR ((%s IS NULL OR %s = '') AND COALESCE(%s, 0) <= 0))", column("billing_mode"), placeholder, column("billing_mode"), column("billing_mode"), column("image_count")))
	default:
		conditions = append(conditions, fmt.Sprintf("%s = %s", column("billing_mode"), placeholder))
	}
	args = append(args, mode)
	return conditions, args
}

func appendUsageLogBillingModeQueryFilter(query string, args []any, billingMode string, alias string) (string, []any) {
	conditions, args := appendUsageLogBillingModeWhereConditionWithAlias(nil, args, billingMode, alias)
	if len(conditions) == 0 {
		return query, args
	}
	return query + " AND " + conditions[0], args
}

func appendUsageLogModelWhereCondition(conditions []string, args []any, model string, source string) ([]string, []any) {
	if strings.TrimSpace(source) == "" {
		return appendRawUsageLogModelWhereCondition(conditions, args, model)
	}
	if strings.TrimSpace(model) == "" {
		return conditions, args
	}
	conditions = append(conditions, fmt.Sprintf("%s = $%d", resolveModelDimensionExpression(source), len(args)+1))
	args = append(args, model)
	return conditions, args
}

// appendRawUsageLogModelQueryFilter keeps direct model filters on the raw model column for backward
// compatibility with historical rows. Requested/upstream analytics must use
// resolveModelDimensionExpression instead.
func appendRawUsageLogModelQueryFilter(query string, args []any, model string) (string, []any) {
	if strings.TrimSpace(model) == "" {
		return query, args
	}
	query += fmt.Sprintf(" AND %s = $%d", rawUsageLogModelColumn, len(args)+1)
	args = append(args, model)
	return query, args
}

func appendUsageLogModelQueryFilter(query string, args []any, model string, source string) (string, []any) {
	if strings.TrimSpace(source) == "" {
		return appendRawUsageLogModelQueryFilter(query, args, model)
	}
	if strings.TrimSpace(model) == "" {
		return query, args
	}
	query += fmt.Sprintf(" AND %s = $%d", resolveModelDimensionExpression(source), len(args)+1)
	args = append(args, model)
	return query, args
}

type usageLogRepository struct {
	client *dbent.Client
	sql    sqlExecutor
	db     *sql.DB

	createBatchOnce     sync.Once
	createBatchCh       chan usageLogCreateRequest
	bestEffortBatchOnce sync.Once
	bestEffortBatchCh   chan usageLogBestEffortRequest
	bestEffortRecent    *gocache.Cache
}

func NewUsageLogRepository(client *dbent.Client, sqlDB *sql.DB) service.UsageLogRepository {
	return newUsageLogRepositoryWithSQL(client, sqlDB)
}

func newUsageLogRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *usageLogRepository {
	// 使用 scanSingleRow 替代 QueryRowContext，保证 ent.Tx 作为 sqlExecutor 可用。
	repo := &usageLogRepository{client: client, sql: sqlq}
	if db, ok := sqlq.(*sql.DB); ok {
		repo.db = db
	}
	repo.bestEffortRecent = gocache.New(usageLogBestEffortRecentTTL, time.Minute)
	return repo
}

func buildWhere(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(conditions, " AND ")
}

func appendRequestTypeOrStreamWhereCondition(conditions []string, args []any, requestType *int16, stream *bool) ([]string, []any) {
	if requestType != nil {
		condition, conditionArgs := buildRequestTypeFilterCondition(len(args)+1, *requestType)
		conditions = append(conditions, condition)
		args = append(args, conditionArgs...)
		return conditions, args
	}
	if stream != nil {
		conditions = append(conditions, fmt.Sprintf("stream = $%d", len(args)+1))
		args = append(args, *stream)
	}
	return conditions, args
}

func appendRequestTypeOrStreamQueryFilter(query string, args []any, requestType *int16, stream *bool) (string, []any) {
	if requestType != nil {
		condition, conditionArgs := buildRequestTypeFilterCondition(len(args)+1, *requestType)
		query += " AND " + condition
		args = append(args, conditionArgs...)
		return query, args
	}
	if stream != nil {
		query += fmt.Sprintf(" AND stream = $%d", len(args)+1)
		args = append(args, *stream)
	}
	return query, args
}

// buildRequestTypeFilterCondition 在 request_type 过滤时兼容 legacy 字段，避免历史数据漏查。
func buildRequestTypeFilterCondition(startArgIndex int, requestType int16) (string, []any) {
	return buildRequestTypeFilterConditionWithAlias(startArgIndex, requestType, "")
}

func buildRequestTypeFilterConditionWithAlias(startArgIndex int, requestType int16, alias string) (string, []any) {
	normalized := service.RequestTypeFromInt16(requestType)
	requestTypeArg := int16(normalized)
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	switch normalized {
	case service.RequestTypeSync:
		return fmt.Sprintf("(%srequest_type = $%d OR (%srequest_type = %d AND %sstream = FALSE AND %sopenai_ws_mode = FALSE))", prefix, startArgIndex, prefix, int16(service.RequestTypeUnknown), prefix, prefix), []any{requestTypeArg}
	case service.RequestTypeStream:
		return fmt.Sprintf("(%srequest_type = $%d OR (%srequest_type = %d AND %sstream = TRUE AND %sopenai_ws_mode = FALSE))", prefix, startArgIndex, prefix, int16(service.RequestTypeUnknown), prefix, prefix), []any{requestTypeArg}
	case service.RequestTypeWSV2:
		return fmt.Sprintf("(%srequest_type = $%d OR (%srequest_type = %d AND %sopenai_ws_mode = TRUE))", prefix, startArgIndex, prefix, int16(service.RequestTypeUnknown), prefix), []any{requestTypeArg}
	default:
		return fmt.Sprintf("%srequest_type = $%d", prefix, startArgIndex), []any{requestTypeArg}
	}
}

// ---- 以下为 dev fork 专有方法（上游拆分重构后仍保留在本文件）----

func (r *usageLogRepository) GetGeoDistribution(ctx context.Context, startTime, endTime time.Time) ([]service.GeoDistributionItem, error) {
	query := `
		SELECT country_code, COUNT(*) as count
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2 AND country_code IS NOT NULL AND country_code != ''
		GROUP BY country_code
		ORDER BY count DESC
	`
	rows, err := r.sql.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []service.GeoDistributionItem
	for rows.Next() {
		var item service.GeoDistributionItem
		if err := rows.Scan(&item.CountryCode, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *usageLogRepository) GetDistinctIPsWithoutCountry(ctx context.Context, limit int) ([]string, error) {
	query := `
		SELECT DISTINCT ip_address FROM usage_logs
		WHERE ip_address IS NOT NULL AND ip_address != ''
		AND (country_code IS NULL OR country_code = '')
		LIMIT $1
	`
	rows, err := r.sql.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ips []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

func (r *usageLogRepository) BackfillCountryCode(ctx context.Context, ip, countryCode string) (int64, error) {
	result, err := r.sql.ExecContext(ctx,
		`UPDATE usage_logs SET country_code = $1 WHERE ip_address = $2 AND (country_code IS NULL OR country_code = '')`,
		countryCode, ip,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// SumUsageCostsByDay 按用户时区分桶同时聚合 actual_cost（从用户余额扣的钱）和
// account_cost（平台向上游 AI 服务实际支付的金额，口径与 GetDashboardStats 的 total_account_cost 一致）。
// 直接走 usage_logs 实时聚合，**用 AT TIME ZONE 按用户时区分桶**——不能用预聚合
// 表 usage_dashboard_daily，因为它的 bucket_date 在 PG session(UTC)tz 下被强制
// 截取过，跨日边界会偏一天；且预聚合表没有 account_cost 字段。
func (r *usageLogRepository) SumUsageCostsByDay(ctx context.Context, startTime, endTime time.Time, tzName string) (map[string]usagestats.DailyUsageCost, error) {
	if tzName == "" {
		tzName = "UTC"
	}
	query := `
		SELECT
			TO_CHAR(created_at AT TIME ZONE $3, 'YYYY-MM-DD') AS day,
			COALESCE(SUM(actual_cost), 0) AS actual_cost,
			COALESCE(SUM(COALESCE(account_stats_cost, total_cost) * COALESCE(account_rate_multiplier, 1)), 0) AS account_cost
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2
		GROUP BY 1
		ORDER BY 1
	`
	rows, err := r.sql.QueryContext(ctx, query, startTime, endTime, tzName)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make(map[string]usagestats.DailyUsageCost)
	for rows.Next() {
		var day string
		var actual, account float64
		if err := rows.Scan(&day, &actual, &account); err != nil {
			return nil, err
		}
		result[day] = usagestats.DailyUsageCost{ActualCost: actual, AccountCost: account}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// SumCommissionByUserIDs returns total_cost for given user IDs.
// Used by commission service to calculate reseller commission (totalCost * commission_rate).
// Note: Returns all records regardless of merchant_rate_snapshot value.
func (r *usageLogRepository) SumCommissionByUserIDs(ctx context.Context, userIDs []int64) (totalCost float64, err error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	query := `
		SELECT COALESCE(SUM(total_cost), 0)
		FROM usage_logs
		WHERE user_id = ANY($1)
	`
	if err := scanSingleRow(ctx, r.sql, query, []any{pq.Array(userIDs)}, &totalCost); err != nil {
		return 0, err
	}
	return totalCost, nil
}

// SumTodayCostByUserIDs returns today's total_cost for the given user IDs.
func (r *usageLogRepository) SumTodayCostByUserIDs(ctx context.Context, userIDs []int64) (float64, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	query := `
		SELECT COALESCE(SUM(total_cost), 0)
		FROM usage_logs
		WHERE user_id = ANY($1)
		  AND created_at >= CURRENT_DATE
	`
	var total float64
	if err := scanSingleRow(ctx, r.sql, query, []any{pq.Array(userIDs)}, &total); err != nil {
		return 0, err
	}
	return total, nil
}

// BackfillMerchantRateSnapshot fills in NULL merchant_rate_snapshot and platform_cost_snapshot values
// for users with a parent reseller that has valid price_multiplier and platform_cost settings.
func (r *usageLogRepository) BackfillMerchantRateSnapshot(ctx context.Context) (int64, error) {
	query := `
		UPDATE usage_logs ul
		SET merchant_rate_snapshot = rs.mult,
		    platform_cost_snapshot = rs.pcost
		FROM users u
		JOIN (
			SELECT r1.reseller_id,
			       r1.value::double precision AS mult,
			       r2.value::double precision AS pcost
			FROM reseller_settings r1
			JOIN reseller_settings r2
			    ON r2.reseller_id = r1.reseller_id AND r2.key = 'platform_cost'
			WHERE r1.key = 'price_multiplier'
			  AND r1.value ~ '^\d+\.?\d*$' AND r1.value::double precision > 0
			  AND r2.value ~ '^\d+\.?\d*$' AND r2.value::double precision > 0
		) rs ON rs.reseller_id = u.parent_id
		WHERE ul.user_id = u.id
		  AND u.parent_id IS NOT NULL
		  AND (ul.merchant_rate_snapshot IS NULL OR ul.platform_cost_snapshot IS NULL)
	`
	result, err := r.sql.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
