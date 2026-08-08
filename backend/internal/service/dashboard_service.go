package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geoip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

const (
	defaultDashboardStatsFreshTTL       = 15 * time.Second
	defaultDashboardStatsCacheTTL       = 30 * time.Second
	defaultDashboardStatsRefreshTimeout = 30 * time.Second
)

// ErrDashboardStatsCacheMiss 标记仪表盘缓存未命中。
var ErrDashboardStatsCacheMiss = errors.New("仪表盘缓存未命中")

// DashboardStatsCache 定义仪表盘统计缓存接口。
type DashboardStatsCache interface {
	GetDashboardStats(ctx context.Context) (string, error)
	SetDashboardStats(ctx context.Context, data string, ttl time.Duration) error
	DeleteDashboardStats(ctx context.Context) error
}

type dashboardStatsRangeFetcher interface {
	GetDashboardStatsWithRange(ctx context.Context, start, end time.Time) (*usagestats.DashboardStats, error)
}

type dashboardStatsCacheEntry struct {
	Stats     *usagestats.DashboardStats `json:"stats"`
	UpdatedAt int64                      `json:"updated_at"`
}

// DashboardService 提供管理员仪表盘统计服务。
type DashboardService struct {
	usageRepo      UsageLogRepository
	aggRepo        DashboardAggregationRepository
	userRepo       UserRepository
	rechargeRepo   RechargeOrderRepository
	orderRepo      OrderRepository
	usdtRepo       UsdtOrderRepository
	redeemRepo     RedeemCodeRepository
	channelRepo    ChannelRepository
	cache          DashboardStatsCache
	cacheFreshTTL  time.Duration
	cacheTTL       time.Duration
	refreshTimeout time.Duration
	refreshing     int32
	aggEnabled     bool
	aggInterval    time.Duration
	aggLookback    time.Duration
	aggUsageDays   int
}

func NewDashboardService(usageRepo UsageLogRepository, aggRepo DashboardAggregationRepository, userRepo UserRepository, rechargeRepo RechargeOrderRepository, orderRepo OrderRepository, usdtRepo UsdtOrderRepository, redeemRepo RedeemCodeRepository, channelRepo ChannelRepository, cache DashboardStatsCache, cfg *config.Config) *DashboardService {
	freshTTL := defaultDashboardStatsFreshTTL
	cacheTTL := defaultDashboardStatsCacheTTL
	refreshTimeout := defaultDashboardStatsRefreshTimeout
	aggEnabled := true
	aggInterval := time.Minute
	aggLookback := 2 * time.Minute
	aggUsageDays := 90
	if cfg != nil {
		if !cfg.Dashboard.Enabled {
			cache = nil
		}
		if cfg.Dashboard.StatsFreshTTLSeconds > 0 {
			freshTTL = time.Duration(cfg.Dashboard.StatsFreshTTLSeconds) * time.Second
		}
		if cfg.Dashboard.StatsTTLSeconds > 0 {
			cacheTTL = time.Duration(cfg.Dashboard.StatsTTLSeconds) * time.Second
		}
		if cfg.Dashboard.StatsRefreshTimeoutSeconds > 0 {
			refreshTimeout = time.Duration(cfg.Dashboard.StatsRefreshTimeoutSeconds) * time.Second
		}
		aggEnabled = cfg.DashboardAgg.Enabled
		if cfg.DashboardAgg.IntervalSeconds > 0 {
			aggInterval = time.Duration(cfg.DashboardAgg.IntervalSeconds) * time.Second
		}
		if cfg.DashboardAgg.LookbackSeconds > 0 {
			aggLookback = time.Duration(cfg.DashboardAgg.LookbackSeconds) * time.Second
		}
		if cfg.DashboardAgg.Retention.UsageLogsDays > 0 {
			aggUsageDays = cfg.DashboardAgg.Retention.UsageLogsDays
		}
	}
	if aggRepo == nil {
		aggEnabled = false
	}
	svc := &DashboardService{
		usageRepo:      usageRepo,
		aggRepo:        aggRepo,
		userRepo:       userRepo,
		rechargeRepo:   rechargeRepo,
		orderRepo:      orderRepo,
		usdtRepo:       usdtRepo,
		redeemRepo:     redeemRepo,
		channelRepo:    channelRepo,
		cache:          cache,
		cacheFreshTTL:  freshTTL,
		cacheTTL:       cacheTTL,
		refreshTimeout: refreshTimeout,
		aggEnabled:     aggEnabled,
		aggInterval:    aggInterval,
		aggLookback:    aggLookback,
		aggUsageDays:   aggUsageDays,
	}

	// Auto-backfill country_code for historical usage logs in background
	go svc.autoBackfillGeoData()

	return svc
}

func (s *DashboardService) autoBackfillGeoData() {
	ctx := context.Background()
	const batchSize = 500
	var totalIPs, totalRows int
	for {
		ips, rows, err := s.BackfillGeoData(ctx, batchSize)
		if err != nil {
			logger.LegacyPrintf("service.dashboard", "[GeoBackfill] error: %v", err)
			return
		}
		totalIPs += ips
		totalRows += int(rows)
		if ips < batchSize {
			break
		}
	}
	if totalIPs > 0 {
		logger.LegacyPrintf("service.dashboard", "[GeoBackfill] completed: %d IPs resolved, %d rows updated", totalIPs, totalRows)
	}
}

func (s *DashboardService) GetDashboardStats(ctx context.Context) (*usagestats.DashboardStats, error) {
	if s.cache != nil {
		cached, fresh, err := s.getCachedDashboardStats(ctx)
		if err == nil && cached != nil {
			s.refreshAggregationStaleness(cached)
			if !fresh {
				s.refreshDashboardStatsAsync()
			}
			return cached, nil
		}
		if err != nil && !errors.Is(err, ErrDashboardStatsCacheMiss) {
			logger.LegacyPrintf("service.dashboard", "[Dashboard] 仪表盘缓存读取失败: %v", err)
		}
	}

	stats, err := s.refreshDashboardStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get dashboard stats: %w", err)
	}
	return stats, nil
}

// FinanceTrendPoint 表示某一天的资金流入/流出。
type FinanceTrendPoint struct {
	Date        string  `json:"date"`
	Recharge    float64 `json:"recharge"`     // 当日到账余额合计（USD）
	Consumption float64 `json:"consumption"`  // 当日实际从用户余额扣除合计（USD），来自 actual_cost
	AccountCost float64 `json:"account_cost"` // 当日平台向上游 AI 服务支付的真实成本合计（USD）
}

// FinanceTrendResult 平台资金趋势聚合结果。
type FinanceTrendResult struct {
	CurrentTotalBalance float64 `json:"current_total_balance"`
	// TotalChannelBalance 所有渠道当前缓存余额之和（cached_balance）。
	// 单位以各渠道 balance_unit 为准，可能不一致，仅做原始求和。
	TotalChannelBalance float64 `json:"total_channel_balance"`
	// TodayRecharge 今日（按 tz）4 类充值来源合计（USD）。与 RechargeBreakdown 无关，
	// 独立于 [startTime, endTime) 过滤范围。
	TodayRecharge float64 `json:"today_recharge"`
	// TodayGrossProfit 今日（按 tz）毛利 = 今日消耗(actual_cost) - 今日上游成本(account_cost)。
	// 与资金趋势图"区间毛利"保持同一口径：充值只是入金不算营收。
	TodayGrossProfit float64             `json:"today_gross_profit"`
	Trend            []FinanceTrendPoint `json:"trend"`
	// RechargeBreakdown 在 [startTime, endTime) 区间内,各类充值来源的总和（USD）。
	// key 固定为以下 4 类，便于前端固定渲染：
	//   - "alipay"        : orders 表 status=paid (AliMPay 通道)
	//   - "wxpay"         : recharge_orders 表 status=paid (EPay 通道)
	//   - "redeem_code"   : redeem_codes type='balance' 已使用且 value>0
	//   - "admin_manual"  : redeem_codes type='admin_balance' 已使用且 value>0，
	//                       排除 notes 以 'AliMPay order '/'Recharge order ' 开头的 audit shadow。
	RechargeBreakdown map[string]float64 `json:"recharge_breakdown"`
}

// GetFinanceTrend 返回平台每日充值/消耗趋势以及当前总余额。
// 时区参数用于 SQL 端按天分桶；与 parseTimeRange 保持一致。
func (s *DashboardService) GetFinanceTrend(ctx context.Context, startTime, endTime time.Time, tzName string) (*FinanceTrendResult, error) {
	// 消耗 / 成本：按用户时区分桶查 usage_logs。
	// **不能复用 GetUsageTrendWithFilters**，它走预聚合表 usage_dashboard_daily，
	// 其 bucket_date 在 PG session(UTC)下被截取，跨日时会偏一天，导致前端图表
	// 不响应顶部日期过滤、显示串日数据。
	var err error
	usageByDay := map[string]usagestats.DailyUsageCost{}
	if s.usageRepo != nil {
		usageByDay, err = s.usageRepo.SumUsageCostsByDay(ctx, startTime, endTime, tzName)
		if err != nil {
			return nil, fmt.Errorf("sum usage costs by day: %w", err)
		}
	}

	// 充值合并：把 5 类来源按天累加到 rechargeByDay，便于趋势线绘制。
	// 5 类来源对应 RechargeBreakdown 的 5 个 key（alipay/wxpay/usdt/redeem_code/admin_manual）。
	rechargeByDay := map[string]float64{}

	// (a) 微信通道：recharge_orders 表 status='paid' 的 credit_amount。
	wxpayByDay := map[string]float64{}
	if s.rechargeRepo != nil {
		wxpayByDay, err = s.rechargeRepo.SumPaidCreditByDay(ctx, startTime, endTime, tzName)
		if err != nil {
			return nil, fmt.Errorf("sum recharge_orders by day: %w", err)
		}
	}
	for d, v := range wxpayByDay {
		rechargeByDay[d] += v
	}

	// (b) 支付宝通道：orders 表 status='paid' 的 credit_amount。
	alipayByDay := map[string]float64{}
	if s.orderRepo != nil {
		alipayByDay, err = s.orderRepo.SumPaidCreditByDay(ctx, startTime, endTime, tzName)
		if err != nil {
			return nil, fmt.Errorf("sum orders by day: %w", err)
		}
	}
	for d, v := range alipayByDay {
		rechargeByDay[d] += v
	}

	// (b2) USDT 通道：usdt_orders 表 status='paid' 的 credit_amount（credit_amount 即 CNY 等值余额，1:1）。
	// USDT 入账只写 usdt_orders + admin_balance 影子行，而影子行被 admin_manual 汇总显式排除，
	// 故必须在此单独并入，否则仪表盘财务口径整体漏计 USDT。
	usdtByDay := map[string]float64{}
	if s.usdtRepo != nil {
		usdtByDay, err = s.usdtRepo.SumPaidCreditByDay(ctx, startTime, endTime, tzName)
		if err != nil {
			return nil, fmt.Errorf("sum usdt_orders by day: %w", err)
		}
	}
	for d, v := range usdtByDay {
		rechargeByDay[d] += v
	}

	// (c) 兑换码：redeem_codes type='balance' value>0 用过的，独立增量（不会重复写 audit）。
	redeemBalanceByDay := map[string]float64{}
	if s.redeemRepo != nil {
		redeemBalanceByDay, err = s.redeemRepo.SumPositiveValueByDayForTypes(
			ctx, startTime, endTime, tzName, []string{RedeemTypeBalance},
		)
		if err != nil {
			return nil, fmt.Errorf("sum redeem balance by day: %w", err)
		}
	}
	for d, v := range redeemBalanceByDay {
		rechargeByDay[d] += v
	}

	// (d) 管理员手工加余额：redeem_codes type='admin_balance' 排除 audit shadow 后的真实增量。
	adminManualByDay := map[string]float64{}
	if s.redeemRepo != nil {
		adminManualByDay, err = s.redeemRepo.SumManualAdminBalanceByDay(ctx, startTime, endTime, tzName)
		if err != nil {
			return nil, fmt.Errorf("sum manual admin balance by day: %w", err)
		}
	}
	for d, v := range adminManualByDay {
		rechargeByDay[d] += v
	}

	sumMap := func(m map[string]float64) float64 {
		var total float64
		for _, v := range m {
			total += v
		}
		return total
	}
	rechargeBreakdown := map[string]float64{
		"alipay":       sumMap(alipayByDay),
		"wxpay":        sumMap(wxpayByDay),
		"usdt":         sumMap(usdtByDay),
		"redeem_code":  sumMap(redeemBalanceByDay),
		"admin_manual": sumMap(adminManualByDay),
	}

	// 合并：以充值日期 ∪ usage 日期为全集。
	dateSet := make(map[string]struct{}, len(usageByDay)+len(rechargeByDay))
	for d := range usageByDay {
		dateSet[d] = struct{}{}
	}
	for d := range rechargeByDay {
		dateSet[d] = struct{}{}
	}

	dates := make([]string, 0, len(dateSet))
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	points := make([]FinanceTrendPoint, 0, len(dates))
	for _, d := range dates {
		uc := usageByDay[d]
		points = append(points, FinanceTrendPoint{
			Date:        d,
			Recharge:    rechargeByDay[d],
			Consumption: uc.ActualCost,
			AccountCost: uc.AccountCost,
		})
	}

	// 当前总余额：所有用户余额之和。
	var totalBalance float64
	if s.userRepo != nil {
		totalBalance, err = s.userRepo.SumTotalBalance(ctx)
		if err != nil {
			return nil, fmt.Errorf("sum total balance: %w", err)
		}
	}

	// 所有渠道缓存余额之和（cached_balance），独立于日期过滤。
	var totalChannelBalance float64
	if s.channelRepo != nil {
		totalChannelBalance, err = s.channelRepo.SumCachedBalance(ctx)
		if err != nil {
			return nil, fmt.Errorf("sum channel balance: %w", err)
		}
	}

	// 今日充值 / 毛利：独立于顶部 [startTime, endTime) 过滤范围，
	// 按用户时区取今日 [00:00, 次日 00:00) 重新聚合 5 类来源 + usage account_cost。
	loc, err := time.LoadLocation(tzName)
	if err != nil || loc == nil {
		loc = time.UTC
	}
	nowInTz := time.Now().In(loc)
	todayStart := time.Date(nowInTz.Year(), nowInTz.Month(), nowInTz.Day(), 0, 0, 0, 0, loc)
	todayEnd := todayStart.Add(24 * time.Hour)

	sumMapValues := func(m map[string]float64) float64 {
		var v float64
		for _, x := range m {
			v += x
		}
		return v
	}

	var todayRecharge float64
	if s.rechargeRepo != nil {
		m, err := s.rechargeRepo.SumPaidCreditByDay(ctx, todayStart, todayEnd, tzName)
		if err != nil {
			return nil, fmt.Errorf("sum today recharge_orders: %w", err)
		}
		todayRecharge += sumMapValues(m)
	}
	if s.orderRepo != nil {
		m, err := s.orderRepo.SumPaidCreditByDay(ctx, todayStart, todayEnd, tzName)
		if err != nil {
			return nil, fmt.Errorf("sum today orders: %w", err)
		}
		todayRecharge += sumMapValues(m)
	}
	if s.usdtRepo != nil {
		m, err := s.usdtRepo.SumPaidCreditByDay(ctx, todayStart, todayEnd, tzName)
		if err != nil {
			return nil, fmt.Errorf("sum today usdt_orders: %w", err)
		}
		todayRecharge += sumMapValues(m)
	}
	if s.redeemRepo != nil {
		m, err := s.redeemRepo.SumPositiveValueByDayForTypes(ctx, todayStart, todayEnd, tzName, []string{RedeemTypeBalance})
		if err != nil {
			return nil, fmt.Errorf("sum today redeem balance: %w", err)
		}
		todayRecharge += sumMapValues(m)
		m2, err := s.redeemRepo.SumManualAdminBalanceByDay(ctx, todayStart, todayEnd, tzName)
		if err != nil {
			return nil, fmt.Errorf("sum today manual admin balance: %w", err)
		}
		todayRecharge += sumMapValues(m2)
	}

	var todayActualCost, todayAccountCost float64
	if s.usageRepo != nil {
		usageToday, err := s.usageRepo.SumUsageCostsByDay(ctx, todayStart, todayEnd, tzName)
		if err != nil {
			return nil, fmt.Errorf("sum today usage costs: %w", err)
		}
		for _, v := range usageToday {
			todayActualCost += v.ActualCost
			todayAccountCost += v.AccountCost
		}
	}
	todayGrossProfit := todayActualCost - todayAccountCost

	return &FinanceTrendResult{
		CurrentTotalBalance: totalBalance,
		TotalChannelBalance: totalChannelBalance,
		TodayRecharge:       todayRecharge,
		TodayGrossProfit:    todayGrossProfit,
		Trend:               points,
		RechargeBreakdown:   rechargeBreakdown,
	}, nil
}

func (s *DashboardService) GetUsageTrendWithFilters(ctx context.Context, startTime, endTime time.Time, granularity string, userID, apiKeyID, accountID, groupID int64, model string, requestType *int16, stream *bool, billingType *int8) ([]usagestats.TrendDataPoint, error) {
	trend, err := s.usageRepo.GetUsageTrendWithFilters(ctx, startTime, endTime, granularity, userID, apiKeyID, accountID, groupID, model, requestType, stream, billingType)
	if err != nil {
		return nil, fmt.Errorf("get usage trend with filters: %w", err)
	}
	return trend, nil
}

func (s *DashboardService) GetUsageTrendWithUsageFilters(ctx context.Context, startTime, endTime time.Time, granularity string, filters usagestats.UsageLogFilters) ([]usagestats.TrendDataPoint, error) {
	type usageTrendWithFiltersRepo interface {
		GetUsageTrendWithUsageFilters(context.Context, time.Time, time.Time, string, usagestats.UsageLogFilters) ([]usagestats.TrendDataPoint, error)
	}
	if repo, ok := s.usageRepo.(usageTrendWithFiltersRepo); ok {
		trend, err := repo.GetUsageTrendWithUsageFilters(ctx, startTime, endTime, granularity, filters)
		if err != nil {
			return nil, fmt.Errorf("get usage trend with usage filters: %w", err)
		}
		return trend, nil
	}
	return s.GetUsageTrendWithFilters(ctx, startTime, endTime, granularity, filters.UserID, filters.APIKeyID, filters.AccountID, filters.GroupID, filters.Model, filters.RequestType, filters.Stream, filters.BillingType)
}

func (s *DashboardService) GetModelStatsWithFilters(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, requestType *int16, stream *bool, billingType *int8) ([]usagestats.ModelStat, error) {
	stats, err := s.usageRepo.GetModelStatsWithFilters(ctx, startTime, endTime, userID, apiKeyID, accountID, groupID, requestType, stream, billingType)
	if err != nil {
		return nil, fmt.Errorf("get model stats with filters: %w", err)
	}
	return stats, nil
}

func (s *DashboardService) GetModelStatsWithFiltersBySource(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, requestType *int16, stream *bool, billingType *int8, modelSource string) ([]usagestats.ModelStat, error) {
	normalizedSource := usagestats.NormalizeModelSource(modelSource)
	if normalizedSource == usagestats.ModelSourceRequested {
		return s.GetModelStatsWithFilters(ctx, startTime, endTime, userID, apiKeyID, accountID, groupID, requestType, stream, billingType)
	}

	type modelStatsBySourceRepo interface {
		GetModelStatsWithFiltersBySource(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, requestType *int16, stream *bool, billingType *int8, source string) ([]usagestats.ModelStat, error)
	}

	if sourceRepo, ok := s.usageRepo.(modelStatsBySourceRepo); ok {
		stats, err := sourceRepo.GetModelStatsWithFiltersBySource(ctx, startTime, endTime, userID, apiKeyID, accountID, groupID, requestType, stream, billingType, normalizedSource)
		if err != nil {
			return nil, fmt.Errorf("get model stats with filters by source: %w", err)
		}
		return stats, nil
	}

	return s.GetModelStatsWithFilters(ctx, startTime, endTime, userID, apiKeyID, accountID, groupID, requestType, stream, billingType)
}

func (s *DashboardService) GetModelStatsWithUsageFiltersBySource(ctx context.Context, startTime, endTime time.Time, filters usagestats.UsageLogFilters, modelSource string) ([]usagestats.ModelStat, error) {
	normalizedSource := usagestats.NormalizeModelSource(modelSource)
	type modelStatsWithFiltersRepo interface {
		GetModelStatsWithUsageFiltersBySource(context.Context, time.Time, time.Time, usagestats.UsageLogFilters, string) ([]usagestats.ModelStat, error)
	}
	if repo, ok := s.usageRepo.(modelStatsWithFiltersRepo); ok {
		stats, err := repo.GetModelStatsWithUsageFiltersBySource(ctx, startTime, endTime, filters, normalizedSource)
		if err != nil {
			return nil, fmt.Errorf("get model stats with usage filters by source: %w", err)
		}
		return stats, nil
	}
	return s.GetModelStatsWithFiltersBySource(ctx, startTime, endTime, filters.UserID, filters.APIKeyID, filters.AccountID, filters.GroupID, filters.RequestType, filters.Stream, filters.BillingType, normalizedSource)
}

func (s *DashboardService) GetGroupStatsWithFilters(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, requestType *int16, stream *bool, billingType *int8) ([]usagestats.GroupStat, error) {
	stats, err := s.usageRepo.GetGroupStatsWithFilters(ctx, startTime, endTime, userID, apiKeyID, accountID, groupID, requestType, stream, billingType)
	if err != nil {
		return nil, fmt.Errorf("get group stats with filters: %w", err)
	}
	return stats, nil
}

func (s *DashboardService) GetGroupStatsWithUsageFilters(ctx context.Context, startTime, endTime time.Time, filters usagestats.UsageLogFilters) ([]usagestats.GroupStat, error) {
	type groupStatsWithFiltersRepo interface {
		GetGroupStatsWithUsageFilters(context.Context, time.Time, time.Time, usagestats.UsageLogFilters) ([]usagestats.GroupStat, error)
	}
	if repo, ok := s.usageRepo.(groupStatsWithFiltersRepo); ok {
		stats, err := repo.GetGroupStatsWithUsageFilters(ctx, startTime, endTime, filters)
		if err != nil {
			return nil, fmt.Errorf("get group stats with usage filters: %w", err)
		}
		return stats, nil
	}
	return s.GetGroupStatsWithFilters(ctx, startTime, endTime, filters.UserID, filters.APIKeyID, filters.AccountID, filters.GroupID, filters.RequestType, filters.Stream, filters.BillingType)
}

// GetGroupUsageSummary returns today's and cumulative cost for all groups.
func (s *DashboardService) GetGroupUsageSummary(ctx context.Context, todayStart time.Time) ([]usagestats.GroupUsageSummary, error) {
	results, err := s.usageRepo.GetAllGroupUsageSummary(ctx, todayStart)
	if err != nil {
		return nil, fmt.Errorf("get group usage summary: %w", err)
	}
	return results, nil
}

func (s *DashboardService) getCachedDashboardStats(ctx context.Context) (*usagestats.DashboardStats, bool, error) {
	data, err := s.cache.GetDashboardStats(ctx)
	if err != nil {
		return nil, false, err
	}

	var entry dashboardStatsCacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		s.evictDashboardStatsCache(err)
		return nil, false, ErrDashboardStatsCacheMiss
	}
	if entry.Stats == nil {
		s.evictDashboardStatsCache(errors.New("仪表盘缓存缺少统计数据"))
		return nil, false, ErrDashboardStatsCacheMiss
	}

	age := time.Since(time.Unix(entry.UpdatedAt, 0))
	return entry.Stats, age <= s.cacheFreshTTL, nil
}

func (s *DashboardService) refreshDashboardStats(ctx context.Context) (*usagestats.DashboardStats, error) {
	stats, err := s.fetchDashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	s.applyAggregationStatus(ctx, stats)
	cacheCtx, cancel := s.cacheOperationContext()
	defer cancel()
	s.saveDashboardStatsCache(cacheCtx, stats)
	return stats, nil
}

func (s *DashboardService) refreshDashboardStatsAsync() {
	if s.cache == nil {
		return
	}
	if !atomic.CompareAndSwapInt32(&s.refreshing, 0, 1) {
		return
	}

	go func() {
		defer atomic.StoreInt32(&s.refreshing, 0)

		ctx, cancel := context.WithTimeout(context.Background(), s.refreshTimeout)
		defer cancel()

		stats, err := s.fetchDashboardStats(ctx)
		if err != nil {
			logger.LegacyPrintf("service.dashboard", "[Dashboard] 仪表盘缓存异步刷新失败: %v", err)
			return
		}
		s.applyAggregationStatus(ctx, stats)
		cacheCtx, cancel := s.cacheOperationContext()
		defer cancel()
		s.saveDashboardStatsCache(cacheCtx, stats)
	}()
}

func (s *DashboardService) fetchDashboardStats(ctx context.Context) (*usagestats.DashboardStats, error) {
	if !s.aggEnabled {
		if fetcher, ok := s.usageRepo.(dashboardStatsRangeFetcher); ok {
			now := time.Now().UTC()
			start := truncateToDayUTC(now.AddDate(0, 0, -s.aggUsageDays))
			return fetcher.GetDashboardStatsWithRange(ctx, start, now)
		}
	}
	return s.usageRepo.GetDashboardStats(ctx)
}

func (s *DashboardService) saveDashboardStatsCache(ctx context.Context, stats *usagestats.DashboardStats) {
	if s.cache == nil || stats == nil {
		return
	}

	entry := dashboardStatsCacheEntry{
		Stats:     stats,
		UpdatedAt: time.Now().Unix(),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		logger.LegacyPrintf("service.dashboard", "[Dashboard] 仪表盘缓存序列化失败: %v", err)
		return
	}

	if err := s.cache.SetDashboardStats(ctx, string(data), s.cacheTTL); err != nil {
		logger.LegacyPrintf("service.dashboard", "[Dashboard] 仪表盘缓存写入失败: %v", err)
	}
}

func (s *DashboardService) evictDashboardStatsCache(reason error) {
	if s.cache == nil {
		return
	}
	cacheCtx, cancel := s.cacheOperationContext()
	defer cancel()

	if err := s.cache.DeleteDashboardStats(cacheCtx); err != nil {
		logger.LegacyPrintf("service.dashboard", "[Dashboard] 仪表盘缓存清理失败: %v", err)
	}
	if reason != nil {
		logger.LegacyPrintf("service.dashboard", "[Dashboard] 仪表盘缓存异常，已清理: %v", reason)
	}
}

func (s *DashboardService) cacheOperationContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.refreshTimeout)
}

func (s *DashboardService) applyAggregationStatus(ctx context.Context, stats *usagestats.DashboardStats) {
	if stats == nil {
		return
	}
	updatedAt := s.fetchAggregationUpdatedAt(ctx)
	stats.StatsUpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	stats.StatsStale = s.isAggregationStale(updatedAt, time.Now().UTC())
}

func (s *DashboardService) refreshAggregationStaleness(stats *usagestats.DashboardStats) {
	if stats == nil {
		return
	}
	updatedAt := parseStatsUpdatedAt(stats.StatsUpdatedAt)
	stats.StatsStale = s.isAggregationStale(updatedAt, time.Now().UTC())
}

func (s *DashboardService) fetchAggregationUpdatedAt(ctx context.Context) time.Time {
	if s.aggRepo == nil {
		return time.Unix(0, 0).UTC()
	}
	updatedAt, err := s.aggRepo.GetAggregationWatermark(ctx)
	if err != nil {
		logger.LegacyPrintf("service.dashboard", "[Dashboard] 读取聚合水位失败: %v", err)
		return time.Unix(0, 0).UTC()
	}
	if updatedAt.IsZero() {
		return time.Unix(0, 0).UTC()
	}
	return updatedAt.UTC()
}

func (s *DashboardService) isAggregationStale(updatedAt, now time.Time) bool {
	if !s.aggEnabled {
		return true
	}
	epoch := time.Unix(0, 0).UTC()
	if !updatedAt.After(epoch) {
		return true
	}
	threshold := s.aggInterval + s.aggLookback
	return now.Sub(updatedAt) > threshold
}

func parseStatsUpdatedAt(raw string) time.Time {
	if raw == "" {
		return time.Unix(0, 0).UTC()
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Unix(0, 0).UTC()
	}
	return parsed.UTC()
}

func (s *DashboardService) GetAPIKeyUsageTrend(ctx context.Context, startTime, endTime time.Time, granularity string, limit int) ([]usagestats.APIKeyUsageTrendPoint, error) {
	trend, err := s.usageRepo.GetAPIKeyUsageTrend(ctx, startTime, endTime, granularity, limit)
	if err != nil {
		return nil, fmt.Errorf("get api key usage trend: %w", err)
	}
	return trend, nil
}

func (s *DashboardService) GetUserUsageTrend(ctx context.Context, startTime, endTime time.Time, granularity string, limit int) ([]usagestats.UserUsageTrendPoint, error) {
	trend, err := s.usageRepo.GetUserUsageTrend(ctx, startTime, endTime, granularity, limit)
	if err != nil {
		return nil, fmt.Errorf("get user usage trend: %w", err)
	}
	return trend, nil
}

func (s *DashboardService) GetUserSpendingRanking(ctx context.Context, startTime, endTime time.Time, limit int) (*usagestats.UserSpendingRankingResponse, error) {
	ranking, err := s.usageRepo.GetUserSpendingRanking(ctx, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("get user spending ranking: %w", err)
	}
	return ranking, nil
}

func (s *DashboardService) GetUserBreakdownStats(ctx context.Context, startTime, endTime time.Time, dim usagestats.UserBreakdownDimension, limit int) ([]usagestats.UserBreakdownItem, error) {
	stats, err := s.usageRepo.GetUserBreakdownStats(ctx, startTime, endTime, dim, limit)
	if err != nil {
		return nil, fmt.Errorf("get user breakdown stats: %w", err)
	}
	return stats, nil
}

func (s *DashboardService) GetBatchUserUsageStats(ctx context.Context, userIDs []int64, startTime, endTime time.Time) (map[int64]*usagestats.BatchUserUsageStats, error) {
	stats, err := s.usageRepo.GetBatchUserUsageStats(ctx, userIDs, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get batch user usage stats: %w", err)
	}
	return stats, nil
}

func (s *DashboardService) GetBatchAPIKeyUsageStats(ctx context.Context, apiKeyIDs []int64, startTime, endTime time.Time) (map[int64]*usagestats.BatchAPIKeyUsageStats, error) {
	stats, err := s.usageRepo.GetBatchAPIKeyUsageStats(ctx, apiKeyIDs, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get batch api key usage stats: %w", err)
	}
	return stats, nil
}

func (s *DashboardService) GetGeoDistribution(ctx context.Context, startTime, endTime time.Time) ([]GeoDistributionItem, error) {
	items, err := s.usageRepo.GetGeoDistribution(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get geo distribution: %w", err)
	}
	return items, nil
}

// BackfillGeoData resolves IPs without country_code using GeoIP service.
// Returns the number of IPs processed and rows updated.
func (s *DashboardService) BackfillGeoData(ctx context.Context, batchSize int) (ipsProcessed int, rowsUpdated int64, err error) {
	geoSvc := geoip.Get()
	if geoSvc == nil || !geoSvc.IsAvailable() {
		return 0, 0, fmt.Errorf("GeoIP service not available")
	}

	ips, err := s.usageRepo.GetDistinctIPsWithoutCountry(ctx, batchSize)
	if err != nil {
		return 0, 0, fmt.Errorf("get distinct IPs: %w", err)
	}

	for _, ip := range ips {
		code := geoSvc.LookupCountry(ip)
		if code == "" {
			code = "XX" // Unknown
		}
		affected, err := s.usageRepo.BackfillCountryCode(ctx, ip, code)
		if err != nil {
			logger.LegacyPrintf("service.dashboard", "[GeoBackfill] Failed to update IP %s: %v", ip, err)
			continue
		}
		ipsProcessed++
		rowsUpdated += affected
	}
	return ipsProcessed, rowsUpdated, nil
}
