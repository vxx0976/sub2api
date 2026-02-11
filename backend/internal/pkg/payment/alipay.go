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
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// AlipayPayment is the native Alipay payment client
type AlipayPayment struct {
	cfg        config.AlipayPaymentConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey

	// Amount allocation for business QR mode
	mu               sync.Mutex
	allocatedAmounts map[string]time.Time // amount string -> allocation time
}

// NewAlipayPayment creates a new Alipay payment client
func NewAlipayPayment(cfg config.AlipayPaymentConfig) (*AlipayPayment, error) {
	ap := &AlipayPayment{
		cfg:              cfg,
		allocatedAmounts: make(map[string]time.Time),
	}

	if cfg.AppID != "" {
		if cfg.PrivateKey != "" {
			pk, err := parseRSAPrivateKey(cfg.PrivateKey)
			if err != nil {
				return nil, fmt.Errorf("parse private key: %w", err)
			}
			ap.privateKey = pk
		}
		if cfg.AlipayPublicKey != "" {
			pub, err := parseRSAPublicKey(cfg.AlipayPublicKey)
			if err != nil {
				return nil, fmt.Errorf("parse public key: %w", err)
			}
			ap.publicKey = pub
		}
	}

	return ap, nil
}

// PaymentInfo contains all payment info returned to frontend
type PaymentInfo struct {
	PaymentAmount float64 `json:"payment_amount"`
	QRCodeURL     string  `json:"qr_code_url"`
	Mode          string  `json:"mode"`
}

// GeneratePaymentInfo generates payment info for an order
func (ap *AlipayPayment) GeneratePaymentInfo(orderNo string, amount float64) (*PaymentInfo, error) {
	if ap.cfg.Mode == "transfer" {
		return ap.generateTransferPayment(orderNo, amount)
	}
	return ap.generateBusinessQRPayment(orderNo, amount)
}

func (ap *AlipayPayment) generateBusinessQRPayment(orderNo string, amount float64) (*PaymentInfo, error) {
	paymentAmount, err := ap.allocateUniqueAmount(amount)
	if err != nil {
		return nil, fmt.Errorf("allocate unique amount: %w", err)
	}

	qrURL := ap.cfg.BusinessQRURL
	if qrURL == "" {
		qrURL = fmt.Sprintf("https://qr.alipay.com/%s", ap.cfg.BusinessQRPath)
	}

	return &PaymentInfo{
		PaymentAmount: paymentAmount,
		QRCodeURL:     qrURL,
		Mode:          "business_qr",
	}, nil
}

func (ap *AlipayPayment) generateTransferPayment(orderNo string, amount float64) (*PaymentInfo, error) {
	transferURL := ap.generateTransferLink(orderNo, amount)
	return &PaymentInfo{
		PaymentAmount: amount,
		QRCodeURL:     transferURL,
		Mode:          "transfer",
	}, nil
}

func (ap *AlipayPayment) generateTransferLink(orderNo string, amount float64) string {
	bizData := fmt.Sprintf(`{"s":"money","u":"%s","a":"%.2f","m":"%s"}`, ap.cfg.TransferUserID, amount, orderNo)
	params := url.Values{}
	params.Set("appId", "20000123")
	params.Set("actionType", "scan")
	params.Set("biz_data", bizData)
	return "alipays://platformapi/startapp?" + params.Encode()
}

func (ap *AlipayPayment) allocateUniqueAmount(baseAmount float64) (float64, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	timeout := time.Duration(ap.cfg.OrderTimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Minute
	}
	now := time.Now()
	for k, t := range ap.allocatedAmounts {
		if now.Sub(t) > timeout {
			delete(ap.allocatedAmounts, k)
		}
	}

	offset := ap.cfg.AmountOffset
	if offset <= 0 {
		offset = 0.01
	}

	for i := 0; i < 100; i++ {
		candidate := baseAmount + float64(i)*offset
		key := fmt.Sprintf("%.2f", candidate)
		if _, exists := ap.allocatedAmounts[key]; !exists {
			ap.allocatedAmounts[key] = now
			return candidate, nil
		}
	}

	return 0, fmt.Errorf("could not allocate unique amount after 100 attempts")
}

// ReleaseAmount releases an allocated amount
func (ap *AlipayPayment) ReleaseAmount(amount float64) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	delete(ap.allocatedAmounts, fmt.Sprintf("%.2f", amount))
}

// AccountBill represents a single account log entry from Alipay
type AccountBill struct {
	TransDt         string  `json:"trans_dt"`
	AccountLogID    string  `json:"account_log_id"`
	AlipayOrderNo   string  `json:"alipay_order_no"`
	MerchantOrderNo string  `json:"merchant_order_no"`
	TransAmount     float64 `json:"trans_amount"`
	Balance         float64 `json:"balance"`
	Type            string  `json:"type"`
	OtherAccount    string  `json:"other_account"`
	TransMemo       string  `json:"trans_memo"`
}

// QueryAccountBills queries Alipay account logs via alipay.data.bill.accountlog.query
func (ap *AlipayPayment) QueryAccountBills(ctx context.Context, startTime, endTime time.Time) ([]AccountBill, error) {
	if ap.privateKey == nil {
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
		"app_id":      ap.cfg.AppID,
		"method":      "alipay.data.bill.accountlog.query",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContentJSON),
	}

	signStr := buildSignString(params)
	signature, err := rsaSign(signStr, ap.privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}
	params["sign"] = signature

	serverURL := ap.cfg.ServerURL
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
			TotalSize  string        `json:"total_size"`
			DetailList []AccountBill `json:"detail_list"`
		} `json:"alipay_data_bill_accountlog_query_response"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(data))
	}

	resp := result.AlipayDataBillAccountlogQueryResponse
	if resp.Code != "10000" {
		return nil, fmt.Errorf("alipay API error: code=%s, msg=%s", resp.Code, resp.Msg)
	}

	return resp.DetailList, nil
}

// RSA2 signing helpers

func buildSignString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || k == "sign_type" {
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
