package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// 共享 setting keys（与链无关；前端 SettingsView / 后台可配置）
const (
	SettingKeyUsdtEnabled                = "usdt_enabled" // 主开关
	SettingKeyUsdtMinAmount              = "usdt_min_amount"
	SettingKeyUsdtMaxAmount              = "usdt_max_amount"
	SettingKeyUsdtManualRate             = "usdt_manual_rate"
	SettingKeyUsdtRateAutoFetch          = "usdt_rate_auto_fetch"
	SettingKeyUsdtRateMarkup             = "usdt_rate_markup"
	SettingKeyUsdtAmountOffset           = "usdt_amount_offset"
	SettingKeyUsdtAmountTolerance        = "usdt_amount_tolerance"
	SettingKeyUsdtConfirmSeconds         = "usdt_confirm_seconds"
	SettingKeyUsdtMonitorIntervalSeconds = "usdt_monitor_interval_seconds"
	SettingKeyUsdtQueryMinutesBack       = "usdt_query_minutes_back"
	SettingKeyUsdtOrderTimeoutSeconds    = "usdt_order_timeout_seconds"
)

// UsdtChainSettingKey 返回某条链某字段的 setting key，如 usdt_trc20_address。
// suffix ∈ {enabled,address,api_key,api_base_url}
func UsdtChainSettingKey(chain, suffix string) string {
	return "usdt_" + chain + "_" + suffix
}

// chainRuntime 是一条链的运行时配置快照。
type chainRuntime struct {
	Enabled bool
	Address string
	APIKey  string
	BaseURL string
}

// SettingStore 在 SettingGetter(只读)基础上增加写能力，
// 供 USDT 链上扫描游标(usdt_<chain>_scan_block)持久化到 settings 表。
type SettingStore interface {
	SettingGetter
	Set(ctx context.Context, key, value string) error
}

// UsdtPayment 是多链 USDT 自建收款的「配置持有者」。
// 共享配置优先级：settings 表（动态）> config.yaml（fallback）；
// per-chain 配置（地址/api key/启用）只来自 settings（在后台配置）。
type UsdtPayment struct {
	mu          sync.Mutex
	cfg         config.UsdtPaymentConfig // 共享字段
	fallbackCfg config.UsdtPaymentConfig
	settings    SettingStore
	chains      map[string]*chainRuntime // chain -> 运行时配置

	adapters map[string]ChainAdapter

	httpClient *http.Client

	rateMu        sync.Mutex
	cachedRate    float64
	cachedRateExp time.Time
}

// NewUsdtPayment 创建 UsdtPayment。
func NewUsdtPayment(fallback config.UsdtPaymentConfig, settings SettingStore) (*UsdtPayment, error) {
	up := &UsdtPayment{
		cfg:         fallback,
		fallbackCfg: fallback,
		settings:    settings,
		chains:      make(map[string]*chainRuntime),
		adapters:    newAdapters(),
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
	for _, c := range SupportedChains {
		up.chains[c] = &chainRuntime{}
	}
	return up, nil
}

// Reload 从 settings 表读取最新共享配置 + per-chain 配置。
func (u *UsdtPayment) Reload(ctx context.Context) {
	if u == nil || u.settings == nil {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	get := func(k string) string {
		v, _ := u.settings.GetValue(ctx, k)
		return v
	}

	// 共享配置
	cfg := u.fallbackCfg
	if v := get(SettingKeyUsdtMinAmount); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			cfg.MinAmount = f
		}
	}
	if v := get(SettingKeyUsdtMaxAmount); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			cfg.MaxAmount = f
		}
	}
	if v := get(SettingKeyUsdtManualRate); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			cfg.ManualRate = f
		}
	}
	if v := get(SettingKeyUsdtRateAutoFetch); v != "" {
		cfg.RateAutoFetch = v == "true"
	}
	if v := get(SettingKeyUsdtRateMarkup); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 && f < 1 {
			cfg.RateMarkup = f
		}
	}
	if v := get(SettingKeyUsdtAmountOffset); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			cfg.AmountOffset = f
		}
	}
	if v := get(SettingKeyUsdtAmountTolerance); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			cfg.AmountTolerance = f
		}
	}
	if v := get(SettingKeyUsdtConfirmSeconds); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			cfg.ConfirmSeconds = i
		}
	}
	if v := get(SettingKeyUsdtMonitorIntervalSeconds); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			cfg.MonitorIntervalSeconds = i
		}
	}
	if v := get(SettingKeyUsdtQueryMinutesBack); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			cfg.QueryMinutesBack = i
		}
	}
	if v := get(SettingKeyUsdtOrderTimeoutSeconds); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			cfg.OrderTimeoutSeconds = i
		}
	}
	u.cfg = cfg

	// per-chain 配置
	for _, chain := range SupportedChains {
		u.chains[chain] = &chainRuntime{
			Enabled: get(UsdtChainSettingKey(chain, "enabled")) == "true",
			Address: get(UsdtChainSettingKey(chain, "address")),
			APIKey:  get(UsdtChainSettingKey(chain, "api_key")),
			BaseURL: get(UsdtChainSettingKey(chain, "api_base_url")),
		}
	}
}

// IsEnabled 主开关（查 settings.usdt_enabled，nil settings 恒 false）。
func (u *UsdtPayment) IsEnabled(ctx context.Context) bool {
	if u == nil || u.settings == nil {
		return false
	}
	v, _ := u.settings.GetValue(ctx, SettingKeyUsdtEnabled)
	return v == "true"
}

func (u *UsdtPayment) chainSnapshot(chain string) chainRuntime {
	u.mu.Lock()
	defer u.mu.Unlock()
	if cr, ok := u.chains[chain]; ok && cr != nil {
		return *cr
	}
	return chainRuntime{}
}

func (u *UsdtPayment) sharedSnapshot() config.UsdtPaymentConfig {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.cfg
}

// Adapter 返回某条链的适配器。
func (u *UsdtPayment) Adapter(chain string) ChainAdapter {
	return u.adapters[chain]
}

// ChainAddress 返回某条链的收款地址。
func (u *UsdtPayment) ChainAddress(chain string) string {
	return u.chainSnapshot(chain).Address
}

// IsChainUsable 判断某条链当前是否可收款：主开关开 + 该链启用 + 地址合法。
func (u *UsdtPayment) IsChainUsable(ctx context.Context, chain string) bool {
	if !u.IsEnabled(ctx) {
		return false
	}
	adapter, ok := u.adapters[chain]
	if !ok {
		return false
	}
	cr := u.chainSnapshot(chain)
	return cr.Enabled && cr.Address != "" && adapter.ValidateAddress(cr.Address)
}

// UsableChains 返回当前可收款的链（按 SupportedChains 顺序）。
func (u *UsdtPayment) UsableChains(ctx context.Context) []string {
	out := make([]string, 0, len(SupportedChains))
	for _, c := range SupportedChains {
		if u.IsChainUsable(ctx, c) {
			out = append(out, c)
		}
	}
	return out
}

// AmountOffset 唯一金额尾数步长（USDT），默认 0.0001。
// MinUsdt 充值最小 USDT 数量，默认 0.1。
func (u *UsdtPayment) MinUsdt() float64 {
	if v := u.sharedSnapshot().MinAmount; v > 0 {
		return v
	}
	return 0.1
}

// MaxUsdt 充值最大 USDT 数量；0 = 不限。
func (u *UsdtPayment) MaxUsdt() float64 {
	if v := u.sharedSnapshot().MaxAmount; v > 0 {
		return v
	}
	return 0
}

// AmountTolerance 到账金额容差（USDT），默认 0.01。实收与应付在 ±容差内即算匹配成功。
func (u *UsdtPayment) AmountTolerance() float64 {
	if t := u.sharedSnapshot().AmountTolerance; t > 0 {
		return t
	}
	return 0.01
}

func (u *UsdtPayment) AmountOffset() float64 {
	off := u.sharedSnapshot().AmountOffset
	if off <= 0 {
		off = 0.05
	}
	// 自动保证间隔 > 2*容差：否则容差匹配时相邻订单金额落入同一容差带导致归属歧义。
	if minOff := 2*u.AmountTolerance() + 0.001; off < minOff {
		off = minOff
	}
	return off
}

// ConfirmDuration 到账交易需达到的最小链上时长才入账，默认 60s。
func (u *UsdtPayment) ConfirmDuration() time.Duration {
	if v := u.sharedSnapshot().ConfirmSeconds; v > 0 {
		return time.Duration(v) * time.Second
	}
	return 60 * time.Second
}

// MonitorInterval 轮询间隔，默认 15s，最低 5s。
func (u *UsdtPayment) MonitorInterval() time.Duration {
	v := u.sharedSnapshot().MonitorIntervalSeconds
	if v < 5 {
		return 15 * time.Second
	}
	return time.Duration(v) * time.Second
}

// QueryMinutesBack 链上回看窗口（分钟），默认 30。
func (u *UsdtPayment) QueryMinutesBack() int {
	if v := u.sharedSnapshot().QueryMinutesBack; v > 0 {
		return v
	}
	return 30
}

// OrderTimeoutSeconds 订单超时（秒），默认 1800。
func (u *UsdtPayment) OrderTimeoutSeconds() int {
	if v := u.sharedSnapshot().OrderTimeoutSeconds; v > 0 {
		return v
	}
	return 1800
}

// AmountReuseWindow 唯一金额释放回池前的等待 = max(OrderTimeout, QueryMinutesBack)。
func (u *UsdtPayment) AmountReuseWindow() time.Duration {
	snap := u.sharedSnapshot()
	timeout := time.Duration(snap.OrderTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	queryBack := time.Duration(u.QueryMinutesBack()) * time.Minute
	if timeout > queryBack {
		return timeout
	}
	return queryBack
}

// GraceWindow 订单过期后仍可被链上到账匹配并补入账的宽限期。
// 防"用户临近/略超截止才转账→已付却永不入账"的孤儿单（链上确认有延迟，加密资金不可逆）。
// = clamp( max(10min, 2*ConfirmDuration), 上限=AmountReuseWindow )。
// 必须 <= AmountReuseWindow，否则唯一金额可能已被新订单复用，导致匹配歧义。
func (u *UsdtPayment) GraceWindow() time.Duration {
	grace := 2 * u.ConfirmDuration()
	if grace < 10*time.Minute {
		grace = 10 * time.Minute
	}
	if reuse := u.AmountReuseWindow(); grace > reuse {
		grace = reuse
	}
	return grace
}

// usdt 换算汇率(CNY/USDT)合理性硬上下限：防异常/被污染的市场价让用户近乎零成本充值。
const (
	usdtRateFloor = 1.0   // 1 USDT < 1 CNY 不可能
	usdtRateCeil  = 100.0 // 1 USDT > 100 CNY 不可能
)

// QueryRate 返回换算用汇率（1 USDT = ? CNY，已叠加加价 markup），所有链共用。
func (u *UsdtPayment) QueryRate(ctx context.Context) (float64, error) {
	snap := u.sharedSnapshot()
	base := snap.ManualRate
	usedMarket := false
	if snap.RateAutoFetch {
		if mkt, err := u.fetchMarketRate(ctx); err == nil && mkt > 0 {
			base = mkt
			usedMarket = true
		} else if base <= 0 {
			return 0, fmt.Errorf("usdt rate unavailable: auto-fetch failed and no manual rate: %v", err)
		} else {
			log.Printf("[UsdtPayment] market rate fetch failed, falling back to manual rate %.4f: %v", base, err)
		}
	}
	if base <= 0 {
		return 0, fmt.Errorf("usdt manual rate not configured")
	}

	// 合理性护栏：市场价超出硬区间、或与手动价偏离 >±30% 时，回退手动价；手动价本身也校验硬区间。
	if usedMarket {
		bad := base < usdtRateFloor || base > usdtRateCeil
		if !bad && snap.ManualRate > 0 && (base < snap.ManualRate*0.7 || base > snap.ManualRate*1.3) {
			bad = true
		}
		if bad {
			if snap.ManualRate >= usdtRateFloor && snap.ManualRate <= usdtRateCeil {
				log.Printf("[UsdtPayment] market rate %.4f rejected (out of band / >±30%% vs manual), using manual %.4f", base, snap.ManualRate)
				base = snap.ManualRate
			} else {
				return 0, fmt.Errorf("usdt market rate %.4f out of sane band [%.0f,%.0f] and no valid manual rate", base, usdtRateFloor, usdtRateCeil)
			}
		}
	} else if base < usdtRateFloor || base > usdtRateCeil {
		return 0, fmt.Errorf("usdt manual rate %.4f out of sane band [%.0f,%.0f]", base, usdtRateFloor, usdtRateCeil)
	}

	markup := snap.RateMarkup
	if markup < 0 || markup >= 1 {
		markup = 0
	}
	return base * (1 - markup), nil
}

func (u *UsdtPayment) fetchMarketRate(ctx context.Context) (float64, error) {
	u.rateMu.Lock()
	if u.cachedRate > 0 && time.Now().Before(u.cachedRateExp) {
		r := u.cachedRate
		u.rateMu.Unlock()
		return r, nil
	}
	u.rateMu.Unlock()

	const api = "https://api.coingecko.com/api/v3/simple/price?ids=tether&vs_currencies=cny"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api, nil)
	if err != nil {
		return 0, err
	}
	resp, err := u.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("coingecko status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	var parsed struct {
		Tether struct {
			CNY float64 `json:"cny"`
		} `json:"tether"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return 0, err
	}
	if parsed.Tether.CNY <= 0 {
		return 0, fmt.Errorf("coingecko returned non-positive rate")
	}
	u.rateMu.Lock()
	u.cachedRate = parsed.Tether.CNY
	u.cachedRateExp = time.Now().Add(60 * time.Second)
	u.rateMu.Unlock()
	return parsed.Tether.CNY, nil
}

// UsdtPaymentInfo 返回给前端展示的收款信息。
type UsdtPaymentInfo struct {
	Chain         string  `json:"chain"`
	Address       string  `json:"address"`
	UsdtAmount    float64 `json:"usdt_amount"`
	UsdtAmountStr string  `json:"usdt_amount_str"`
	Rate          float64 `json:"rate"`
}

// GeneratePaymentInfo 生成某条链的收款展示信息（QR 内容用纯地址）。
func (u *UsdtPayment) GeneratePaymentInfo(chain string, usdtAmount, rate float64) *UsdtPaymentInfo {
	return &UsdtPaymentInfo{
		Chain:         chain,
		Address:       u.ChainAddress(chain),
		UsdtAmount:    usdtAmount,
		UsdtAmountStr: FormatUsdt(usdtAmount),
		Rate:          rate,
	}
}

// cursoredAdapter 是支持「持久化区块游标」扫描的链适配器（目前仅 BSC/EVM）。
// 相比时间窗口的 QueryIncoming，游标扫描从上次已扫块续扫，不受出块提速/重启影响、绝不漏块。
type cursoredAdapter interface {
	// QueryIncomingCursor 从 fromCursor+1 扫到「已确认链头」，返回转账 + 新游标(已成功扫完的最高块)。
	// confirmDur 用于换算确认滞后；matchWindow(订单可匹配最长时长)用于给冷启动/追赶回补定下限，
	// 确保绝不把仍可入账的历史块跳过/钳掉。
	QueryIncomingCursor(ctx context.Context, address, apiKey, baseURL string, fromCursor uint64, confirmDur, matchWindow time.Duration) (transfers []IncomingTransfer, newCursor uint64, err error)
}

// QueryIncoming 拉取某条链收款地址收到的 USDT 转账。
// 支持游标的链(BSC)走持久化区块游标路径（不漏块）；其余链走时间窗口路径。
func (u *UsdtPayment) QueryIncoming(ctx context.Context, chain string, minTimestampMs int64) ([]IncomingTransfer, error) {
	adapter, ok := u.adapters[chain]
	if !ok {
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}
	cr := u.chainSnapshot(chain)
	if cr.Address == "" {
		return nil, fmt.Errorf("chain %s receiving address not configured", chain)
	}
	if ca, ok := adapter.(cursoredAdapter); ok && u.settings != nil {
		return u.queryIncomingCursored(ctx, chain, ca, cr)
	}
	return adapter.QueryIncoming(ctx, cr.Address, cr.APIKey, cr.BaseURL, minTimestampMs)
}

// queryIncomingCursored 用持久化区块游标扫描一条支持游标的链，并把新游标写回 settings。
// 出错时不推进游标（下轮整体重扫，幂等由 usdt_orders.trade_no 唯一索引兜底），保证绝不漏块。
func (u *UsdtPayment) queryIncomingCursored(ctx context.Context, chain string, ca cursoredAdapter, cr chainRuntime) ([]IncomingTransfer, error) {
	key := UsdtChainSettingKey(chain, "scan_block")
	cursor := u.readScanCursor(ctx, key)
	// matchWindow = 一笔到账最长仍可被匹配入账的时长(订单超时 + 宽限)，随 OrderTimeout 配置自适应，
	// 供适配器给冷启动/追赶回补定下限，确保绝不跳过/钳掉仍可入账的历史块。
	matchWindow := time.Duration(u.OrderTimeoutSeconds())*time.Second + u.GraceWindow()
	transfers, newCursor, err := ca.QueryIncomingCursor(ctx, cr.Address, cr.APIKey, cr.BaseURL, cursor, u.ConfirmDuration(), matchWindow)
	if err != nil {
		return nil, err
	}
	if newCursor > cursor {
		if setErr := u.settings.Set(ctx, key, strconv.FormatUint(newCursor, 10)); setErr != nil {
			// 游标写失败：本轮结果照常返回入账（幂等安全），下轮从旧游标重扫补回，不丢块。
			log.Printf("[UsdtPayment] persist scan cursor %s=%d failed: %v", key, newCursor, setErr)
		}
	}
	return transfers, nil
}

// readScanCursor 读取某链持久化的扫描游标（块高）；缺失/非法一律回 0（触发有界冷启动回补）。
func (u *UsdtPayment) readScanCursor(ctx context.Context, key string) uint64 {
	v, err := u.settings.GetValue(ctx, key)
	if err != nil || v == "" {
		return 0
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
