package payment

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// SettingGetter 抽象读取 settings 表的依赖（避免引入 service 包造成循环依赖）
type SettingGetter interface {
	GetValue(ctx context.Context, key string) (string, error)
}

// Setting keys for AliMPay runtime config（前端 SettingsView 可配置）
const (
	SettingKeyAliMPayEnabled                = "alimpay_enabled"
	SettingKeyAliMPayMode                   = "alimpay_mode"
	SettingKeyAliMPayAppID                  = "alimpay_app_id"
	SettingKeyAliMPayPrivateKey             = "alimpay_private_key"
	SettingKeyAliMPayAlipayPublicKey        = "alimpay_alipay_public_key"
	SettingKeyAliMPayServerURL              = "alimpay_server_url"
	SettingKeyAliMPayTransferUserID         = "alimpay_transfer_user_id"
	SettingKeyAliMPayBusinessQRURL          = "alimpay_business_qr_url"
	SettingKeyAliMPayBusinessQRPath         = "alimpay_business_qr_path"
	SettingKeyAliMPayAmountOffset           = "alimpay_amount_offset"
	SettingKeyAliMPayMatchToleranceSeconds  = "alimpay_match_tolerance_seconds"
	SettingKeyAliMPayMonitorIntervalSeconds = "alimpay_monitor_interval_seconds"
	SettingKeyAliMPayQueryMinutesBack       = "alimpay_query_minutes_back"
	SettingKeyAliMPayOrderTimeoutSeconds    = "alimpay_order_timeout_seconds"
)

// AlipayPayment is the native Alipay payment client
// 配置优先级：settings 表（动态） > config.yaml（fallback）
type AlipayPayment struct {
	mu         sync.Mutex
	cfg        config.AlipayPaymentConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey

	// 缓存用于识别 key 是否变化（避免重复解析 RSA）
	privateKeyRaw string
	publicKeyRaw  string

	fallbackCfg config.AlipayPaymentConfig
	settings    SettingGetter // 可为 nil（纯 yaml 模式）
}

// NewAlipayPayment creates a new Alipay payment client
// fallback 是 config.yaml 初始值，settings 若非 nil 则从数据库 setting 动态覆盖
func NewAlipayPayment(fallback config.AlipayPaymentConfig, settings SettingGetter) (*AlipayPayment, error) {
	ap := &AlipayPayment{
		cfg:         fallback,
		fallbackCfg: fallback,
		settings:    settings,
	}
	if err := ap.applyKeys(fallback.PrivateKey, fallback.AlipayPublicKey); err != nil {
		return nil, err
	}
	return ap, nil
}

// Reload 从 settings 表读取最新配置并覆盖 fallback
// 关键字段（enabled/mode/app_id/key 等）若 setting 有值则以 setting 为准
// 建议在 CreateOrder 和 Monitor runCycle 调用
func (ap *AlipayPayment) Reload(ctx context.Context) {
	if ap == nil || ap.settings == nil {
		return
	}
	ap.mu.Lock()
	defer ap.mu.Unlock()

	cfg := ap.fallbackCfg
	get := func(k string) string {
		v, _ := ap.settings.GetValue(ctx, k)
		return v
	}
	if v := get(SettingKeyAliMPayMode); v != "" {
		cfg.Mode = v
	}
	if v := get(SettingKeyAliMPayAppID); v != "" {
		cfg.AppID = v
	}
	if v := get(SettingKeyAliMPayPrivateKey); v != "" {
		cfg.PrivateKey = v
	}
	if v := get(SettingKeyAliMPayAlipayPublicKey); v != "" {
		cfg.AlipayPublicKey = v
	}
	if v := get(SettingKeyAliMPayServerURL); v != "" {
		cfg.ServerURL = v
	}
	if v := get(SettingKeyAliMPayTransferUserID); v != "" {
		cfg.TransferUserID = v
	}
	if v := get(SettingKeyAliMPayBusinessQRURL); v != "" {
		cfg.BusinessQRURL = v
	}
	if v := get(SettingKeyAliMPayBusinessQRPath); v != "" {
		cfg.BusinessQRPath = v
	}
	if v := get(SettingKeyAliMPayAmountOffset); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			cfg.AmountOffset = f
		}
	}
	if v := get(SettingKeyAliMPayMatchToleranceSeconds); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			cfg.MatchToleranceSeconds = i
		}
	}
	if v := get(SettingKeyAliMPayMonitorIntervalSeconds); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			cfg.MonitorIntervalSeconds = i
		}
	}
	if v := get(SettingKeyAliMPayQueryMinutesBack); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			cfg.QueryMinutesBack = i
		}
	}
	if v := get(SettingKeyAliMPayOrderTimeoutSeconds); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			cfg.OrderTimeoutSeconds = i
		}
	}
	ap.cfg = cfg
	_ = ap.applyKeysLocked(cfg.PrivateKey, cfg.AlipayPublicKey)
}

// IsEnabled 返回当前是否启用（查 setting alimpay_enabled，nil settings 下始终 false）
func (ap *AlipayPayment) IsEnabled(ctx context.Context) bool {
	if ap == nil || ap.settings == nil {
		return false
	}
	v, _ := ap.settings.GetValue(ctx, SettingKeyAliMPayEnabled)
	return v == "true"
}

// applyKeys 解析并缓存 RSA key（调用方不持锁）
func (ap *AlipayPayment) applyKeys(private, public string) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	return ap.applyKeysLocked(private, public)
}

// applyKeysLocked 调用方已持锁
func (ap *AlipayPayment) applyKeysLocked(private, public string) error {
	if private != ap.privateKeyRaw {
		if private == "" {
			ap.privateKey = nil
		} else {
			pk, err := parseRSAPrivateKey(private)
			if err != nil {
				return fmt.Errorf("parse private key: %w", err)
			}
			ap.privateKey = pk
		}
		ap.privateKeyRaw = private
	}
	if public != ap.publicKeyRaw {
		if public == "" {
			ap.publicKey = nil
		} else {
			pub, err := parseRSAPublicKey(public)
			if err != nil {
				return fmt.Errorf("parse public key: %w", err)
			}
			ap.publicKey = pub
		}
		ap.publicKeyRaw = public
	}
	return nil
}

type alipaySnapshot struct {
	cfg        config.AlipayPaymentConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func (ap *AlipayPayment) snapshot() alipaySnapshot {
	if ap == nil {
		return alipaySnapshot{}
	}
	ap.mu.Lock()
	defer ap.mu.Unlock()
	return alipaySnapshot{
		cfg:        ap.cfg,
		privateKey: ap.privateKey,
		publicKey:  ap.publicKey,
	}
}

// PaymentInfo contains all payment info returned to frontend
type PaymentInfo struct {
	PaymentAmount float64 `json:"payment_amount"`
	QRCodeURL     string  `json:"qr_code_url"`
	Mode          string  `json:"mode"`
}

// GeneratePaymentInfo generates payment info for an order
func (ap *AlipayPayment) GeneratePaymentInfo(orderNo string, amount float64) (*PaymentInfo, error) {
	snap := ap.snapshot()
	if snap.cfg.Mode == "transfer" {
		return generateTransferPayment(snap.cfg, orderNo, amount)
	}
	return generateBusinessQRPayment(snap.cfg, amount)
}

func generateBusinessQRPayment(cfg config.AlipayPaymentConfig, paymentAmount float64) (*PaymentInfo, error) {
	qrURL := cfg.BusinessQRURL
	if qrURL == "" {
		qrURL = fmt.Sprintf("https://qr.alipay.com/%s", cfg.BusinessQRPath)
	}

	return &PaymentInfo{
		PaymentAmount: paymentAmount,
		QRCodeURL:     qrURL,
		Mode:          "business_qr",
	}, nil
}

func generateTransferPayment(cfg config.AlipayPaymentConfig, orderNo string, amount float64) (*PaymentInfo, error) {
	transferURL := generateTransferLink(cfg, orderNo, amount)
	return &PaymentInfo{
		PaymentAmount: amount,
		QRCodeURL:     transferURL,
		Mode:          "transfer",
	}, nil
}

func generateTransferLink(cfg config.AlipayPaymentConfig, orderNo string, amount float64) string {
	bizData := fmt.Sprintf(`{"s":"money","u":"%s","a":"%.2f","m":"%s"}`, cfg.TransferUserID, amount, orderNo)
	params := url.Values{}
	params.Set("appId", "20000123")
	params.Set("actionType", "scan")
	params.Set("biz_data", bizData)
	return "alipays://platformapi/startapp?" + params.Encode()
}

// Mode returns the configured payment mode (business_qr | transfer).
func (ap *AlipayPayment) Mode() string {
	return ap.snapshot().cfg.Mode
}

// OrderTimeoutSeconds returns the configured order timeout in seconds.
func (ap *AlipayPayment) OrderTimeoutSeconds() int {
	return ap.snapshot().cfg.OrderTimeoutSeconds
}

// AmountOffset returns the amount step used to distinguish business QR orders.
func (ap *AlipayPayment) AmountOffset() float64 {
	offset := ap.snapshot().cfg.AmountOffset
	if offset <= 0 {
		return 0.01
	}
	return offset
}

// MatchTolerance returns the bill/order time tolerance.
func (ap *AlipayPayment) MatchTolerance() time.Duration {
	seconds := ap.snapshot().cfg.MatchToleranceSeconds
	if seconds <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(seconds) * time.Second
}

// QueryMinutesBack returns the bill query lookback window in minutes.
func (ap *AlipayPayment) QueryMinutesBack() int {
	minutes := ap.snapshot().cfg.QueryMinutesBack
	if minutes <= 0 {
		return 30
	}
	return minutes
}

// MonitorInterval returns the configured monitor interval.
func (ap *AlipayPayment) MonitorInterval() time.Duration {
	seconds := ap.snapshot().cfg.MonitorIntervalSeconds
	if seconds < 5 {
		return 10 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

// AmountReuseWindow is the minimum time before a business QR payment amount can be reused.
func (ap *AlipayPayment) AmountReuseWindow() time.Duration {
	snap := ap.snapshot().cfg
	timeout := time.Duration(snap.OrderTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	tolerance := time.Duration(snap.MatchToleranceSeconds) * time.Second
	if tolerance <= 0 {
		tolerance = 5 * time.Minute
	}
	queryBack := time.Duration(snap.QueryMinutesBack) * time.Minute
	if queryBack <= 0 {
		queryBack = 30 * time.Minute
	}
	window := timeout + tolerance + queryBack
	if window < 2*time.Hour {
		window = 2 * time.Hour
	}
	return window
}

// AccountBill represents a single account log entry from Alipay
type AccountBill struct {
	TransDt         string `json:"trans_dt"`
	AccountLogID    string `json:"account_log_id"`
	AlipayOrderNo   string `json:"alipay_order_no"`
	MerchantOrderNo string `json:"merchant_order_no"`
	TransAmount     string `json:"trans_amount"`
	Balance         string `json:"balance"`
	Type            string `json:"type"`
	OtherAccount    string `json:"other_account"`
	TransMemo       string `json:"trans_memo"`
	Direction       string `json:"direction"`
}

// TransAmountFloat returns the transaction amount as float64
func (b AccountBill) TransAmountFloat() float64 {
	v, _ := strconv.ParseFloat(b.TransAmount, 64)
	return v
}

// QueryAccountBills queries Alipay account logs via alipay.data.bill.accountlog.query
func (ap *AlipayPayment) QueryAccountBills(ctx context.Context, startTime, endTime time.Time) ([]AccountBill, error) {
	snap := ap.snapshot()
	if snap.privateKey == nil {
		return nil, fmt.Errorf("alipay private key not configured")
	}

	bizContent := map[string]string{
		"start_time": startTime.Format("2006-01-02 15:04:05"),
		"end_time":   endTime.Format("2006-01-02 15:04:05"),
		"page_no":    "1",
		"page_size":  "2000",
	}
	bizContentJSON, _ := json.Marshal(bizContent)

	params := map[string]string{
		"app_id":      snap.cfg.AppID,
		"method":      "alipay.data.bill.accountlog.query",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContentJSON),
	}

	signStr := buildSignString(params)
	signature, err := rsaSign(signStr, snap.privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}
	params["sign"] = signature

	serverURL := snap.cfg.ServerURL
	if serverURL == "" {
		serverURL = "https://openapi.alipay.com/gateway.do"
	}

	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", serverURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return parseAccountBills(body)
}

func parseAccountBills(data []byte) ([]AccountBill, error) {
	var result struct {
		AlipayDataBillAccountlogQueryResponse struct {
			Code       string        `json:"code"`
			Msg        string        `json:"msg"`
			SubCode    string        `json:"sub_code"`
			SubMsg     string        `json:"sub_msg"`
			TotalSize  string        `json:"total_size"`
			DetailList []AccountBill `json:"detail_list"`
		} `json:"alipay_data_bill_accountlog_query_response"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(data))
	}

	resp := result.AlipayDataBillAccountlogQueryResponse
	if resp.Code != "10000" {
		if resp.SubCode != "" {
			return nil, fmt.Errorf("alipay API error: code=%s, msg=%s, sub_code=%s, sub_msg=%s", resp.Code, resp.Msg, resp.SubCode, resp.SubMsg)
		}
		return nil, fmt.Errorf("alipay API error: code=%s, msg=%s", resp.Code, resp.Msg)
	}

	return resp.DetailList, nil
}

// RSA2 signing helpers

func buildSignString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		if params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	return strings.Join(parts, "&")
}

func rsaSign(content string, privateKey *rsa.PrivateKey) (string, error) {
	h := sha256.New()
	h.Write([]byte(content))
	digest := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func parseRSAPrivateKey(keyStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(keyStr))
	if block == nil {
		keyStr = "-----BEGIN RSA PRIVATE KEY-----\n" + keyStr + "\n-----END RSA PRIVATE KEY-----"
		block, _ = pem.Decode([]byte(keyStr))
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM block")
		}
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		parsed, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse private key: PKCS1: %v, PKCS8: %v", err, err2)
		}
		rsaKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
		return rsaKey, nil
	}
	return key, nil
}

func parseRSAPublicKey(keyStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(keyStr))
	if block == nil {
		keyStr = "-----BEGIN PUBLIC KEY-----\n" + keyStr + "\n-----END PUBLIC KEY-----"
		block, _ = pem.Decode([]byte(keyStr))
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM block")
		}
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaKey, nil
}
