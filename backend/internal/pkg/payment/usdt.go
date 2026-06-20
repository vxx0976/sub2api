package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// Setting keys for USDT(TRC20) runtime config（前端 SettingsView / 后台可配置）
const (
	SettingKeyUsdtEnabled                = "usdt_enabled"
	SettingKeyUsdtReceivingAddress       = "usdt_receiving_address"
	SettingKeyUsdtTronAPIBaseURL         = "usdt_tron_api_base_url"
	SettingKeyUsdtTronAPIKey             = "usdt_tron_api_key"
	SettingKeyUsdtManualRate             = "usdt_manual_rate"
	SettingKeyUsdtRateAutoFetch          = "usdt_rate_auto_fetch"
	SettingKeyUsdtRateMarkup             = "usdt_rate_markup"
	SettingKeyUsdtAmountOffset           = "usdt_amount_offset"
	SettingKeyUsdtConfirmSeconds         = "usdt_confirm_seconds"
	SettingKeyUsdtMonitorIntervalSeconds = "usdt_monitor_interval_seconds"
	SettingKeyUsdtQueryMinutesBack       = "usdt_query_minutes_back"
	SettingKeyUsdtOrderTimeoutSeconds    = "usdt_order_timeout_seconds"
)

// UsdtPayment 是 USDT(TRC20) 自建收款的「SDK / 配置持有者」。
// 配置优先级：settings 表（动态）> config.yaml（fallback）。镜像 AlipayPayment 的结构。
type UsdtPayment struct {
	mu          sync.Mutex
	cfg         config.UsdtPaymentConfig
	fallbackCfg config.UsdtPaymentConfig
	settings    SettingGetter // 可为 nil（纯 yaml 模式）

	httpClient *http.Client

	// 汇率缓存（自动拉取时）
	rateMu        sync.Mutex
	cachedRate    float64
	cachedRateExp time.Time
}

// NewUsdtPayment 创建 UsdtPayment。fallback 是 config.yaml 初始值，settings 非 nil 时动态覆盖。
func NewUsdtPayment(fallback config.UsdtPaymentConfig, settings SettingGetter) (*UsdtPayment, error) {
	return &UsdtPayment{
		cfg:         fallback,
		fallbackCfg: fallback,
		settings:    settings,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// Reload 从 settings 表读取最新配置覆盖 fallback。建议在 CreateOrder 和 Monitor runCycle 调用。
func (u *UsdtPayment) Reload(ctx context.Context) {
	if u == nil || u.settings == nil {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	cfg := u.fallbackCfg
	get := func(k string) string {
		v, _ := u.settings.GetValue(ctx, k)
		return v
	}
	if v := get(SettingKeyUsdtReceivingAddress); v != "" {
		cfg.ReceivingAddress = v
	}
	if v := get(SettingKeyUsdtTronAPIBaseURL); v != "" {
		cfg.TronAPIBaseURL = v
	}
	if v := get(SettingKeyUsdtTronAPIKey); v != "" {
		cfg.TronAPIKey = v
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
}

// IsEnabled 返回当前是否启用（查 setting usdt_enabled，nil settings 下始终 false）。
func (u *UsdtPayment) IsEnabled(ctx context.Context) bool {
	if u == nil || u.settings == nil {
		return false
	}
	v, _ := u.settings.GetValue(ctx, SettingKeyUsdtEnabled)
	return v == "true"
}

func (u *UsdtPayment) snapshot() config.UsdtPaymentConfig {
	if u == nil {
		return config.UsdtPaymentConfig{}
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.cfg
}

// ReceivingAddress 当前收款地址。
func (u *UsdtPayment) ReceivingAddress() string { return u.snapshot().ReceivingAddress }

// Chain 目前固定 trc20。
func (u *UsdtPayment) Chain() string { return "trc20" }

// TronAPIBaseURL 返回 TronGrid 网关（带默认值）。
func (u *UsdtPayment) TronAPIBaseURL() string {
	if v := u.snapshot().TronAPIBaseURL; v != "" {
		return v
	}
	return DefaultTronAPIBaseURL
}

// TronAPIKey 返回 TronGrid API Key（可空）。
func (u *UsdtPayment) TronAPIKey() string { return u.snapshot().TronAPIKey }

// AmountOffset 唯一金额尾数步长（USDT），默认 0.0001。
func (u *UsdtPayment) AmountOffset() float64 {
	if v := u.snapshot().AmountOffset; v > 0 {
		return v
	}
	return 0.0001
}

// ConfirmDuration 到账交易需达到的最小链上时长才入账，默认 60s。
func (u *UsdtPayment) ConfirmDuration() time.Duration {
	if v := u.snapshot().ConfirmSeconds; v > 0 {
		return time.Duration(v) * time.Second
	}
	return 60 * time.Second
}

// MonitorInterval 轮询间隔，默认 15s，最低 5s。
func (u *UsdtPayment) MonitorInterval() time.Duration {
	v := u.snapshot().MonitorIntervalSeconds
	if v < 5 {
		return 15 * time.Second
	}
	return time.Duration(v) * time.Second
}

// QueryMinutesBack 链上回看窗口（分钟），默认 30。
func (u *UsdtPayment) QueryMinutesBack() int {
	if v := u.snapshot().QueryMinutesBack; v > 0 {
		return v
	}
	return 30
}

// OrderTimeoutSeconds 订单超时（秒），默认 1800（30 分钟，链上结算较慢）。
func (u *UsdtPayment) OrderTimeoutSeconds() int {
	if v := u.snapshot().OrderTimeoutSeconds; v > 0 {
		return v
	}
	return 1800
}

// AmountReuseWindow 唯一金额释放回池前的等待时间 = max(OrderTimeout, QueryMinutesBack)。
func (u *UsdtPayment) AmountReuseWindow() time.Duration {
	snap := u.snapshot()
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

// QueryRate 返回换算用汇率（1 USDT = ? CNY，已叠加加价 markup）。
// 自动拉取启用时优先取市场价（缓存 60s）并回退手动价；否则用手动价。
func (u *UsdtPayment) QueryRate(ctx context.Context) (float64, error) {
	snap := u.snapshot()
	base := snap.ManualRate
	if snap.RateAutoFetch {
		if mkt, err := u.fetchMarketRate(ctx); err == nil && mkt > 0 {
			base = mkt
		} else if base <= 0 {
			return 0, fmt.Errorf("usdt rate unavailable: auto-fetch failed and no manual rate: %v", err)
		} else {
			log.Printf("[UsdtPayment] market rate fetch failed, falling back to manual rate %.4f: %v", base, err)
		}
	}
	if base <= 0 {
		return 0, fmt.Errorf("usdt manual rate not configured")
	}
	markup := snap.RateMarkup
	if markup < 0 || markup >= 1 {
		markup = 0
	}
	// 加价：换算用汇率更低 → 用户多付一点 USDT。
	return base * (1 - markup), nil
}

// fetchMarketRate 从 CoinGecko 拉取 USDT/CNY 市场价（缓存 60s）。
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
		return 0, fmt.Errorf("coingecko status %d: %s", resp.StatusCode, string(body))
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

// UsdtPaymentInfo 是返回给前端展示的收款信息。
type UsdtPaymentInfo struct {
	Chain         string  `json:"chain"`
	Address       string  `json:"address"`
	UsdtAmount    float64 `json:"usdt_amount"`
	UsdtAmountStr string  `json:"usdt_amount_str"`
	Rate          float64 `json:"rate"`
}

// GeneratePaymentInfo 生成收款展示信息（QR 内容用纯地址，最兼容各钱包）。
func (u *UsdtPayment) GeneratePaymentInfo(usdtAmount, rate float64) *UsdtPaymentInfo {
	return &UsdtPaymentInfo{
		Chain:         u.Chain(),
		Address:       u.ReceivingAddress(),
		UsdtAmount:    usdtAmount,
		UsdtAmountStr: FormatUsdt(usdtAmount),
		Rate:          rate,
	}
}

// Trc20Transfer 是从 TronGrid 解析出的一笔 TRC20 转账。
type Trc20Transfer struct {
	TxID         string
	From         string
	To           string
	ValueAtomic  string // 链上最小单位（字符串）
	BlockTimeMs  int64  // 区块时间（毫秒）
	ContractAddr string
}

// QueryIncomingTransfers 通过 TronGrid 拉取收款地址在 minTimestampMs 之后收到的 USDT(TRC20) 转账。
func (u *UsdtPayment) QueryIncomingTransfers(ctx context.Context, minTimestampMs int64) ([]Trc20Transfer, error) {
	addr := u.ReceivingAddress()
	if addr == "" {
		return nil, fmt.Errorf("usdt receiving address not configured")
	}

	q := url.Values{}
	q.Set("only_to", "true")
	q.Set("contract_address", UsdtTRC20Contract)
	q.Set("limit", "200")
	q.Set("order_by", "block_timestamp,desc")
	if minTimestampMs > 0 {
		q.Set("min_timestamp", strconv.FormatInt(minTimestampMs, 10))
	}
	endpoint := fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20?%s", u.TronAPIBaseURL(), url.PathEscape(addr), q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if key := u.TronAPIKey(); key != "" {
		req.Header.Set("TRON-PRO-API-KEY", key)
	}
	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("trongrid status %d: %s", resp.StatusCode, truncate(string(body), 300))
	}

	var parsed struct {
		Success bool `json:"success"`
		Data    []struct {
			TransactionID  string `json:"transaction_id"`
			From           string `json:"from"`
			To             string `json:"to"`
			Value          string `json:"value"`
			Type           string `json:"type"`
			BlockTimestamp int64  `json:"block_timestamp"`
			TokenInfo      struct {
				Address string `json:"address"`
			} `json:"token_info"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse trongrid response: %w (body: %s)", err, truncate(string(body), 300))
	}

	out := make([]Trc20Transfer, 0, len(parsed.Data))
	for _, d := range parsed.Data {
		if d.Type != "" && d.Type != "Transfer" {
			continue
		}
		if d.To != addr {
			continue
		}
		if d.TokenInfo.Address != "" && d.TokenInfo.Address != UsdtTRC20Contract {
			continue
		}
		out = append(out, Trc20Transfer{
			TxID:         d.TransactionID,
			From:         d.From,
			To:           d.To,
			ValueAtomic:  d.Value,
			BlockTimeMs:  d.BlockTimestamp,
			ContractAddr: d.TokenInfo.Address,
		})
	}
	return out, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
