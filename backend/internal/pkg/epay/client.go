package epay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// Client 蓝荔支付 SDK 客户端（MD5 签名）
type Client struct {
	apiURL     string
	pid        string
	key        string
	httpClient *http.Client
}

// NewClient 创建支付客户端
func NewClient(apiURL, pid, key string) (*Client, error) {
	if apiURL == "" || pid == "" || key == "" {
		return nil, fmt.Errorf("epay: apiURL, pid, key are required")
	}
	apiURL = strings.TrimRight(apiURL, "/") + "/"
	return &Client{
		apiURL:     apiURL,
		pid:        pid,
		key:        key,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// GetPayLink 生成支付跳转链接（页面跳转模式，submit.php）
func (c *Client) GetPayLink(req CreatePaymentRequest) (string, error) {
	params := map[string]string{
		"pid":          c.pid,
		"type":         req.Type,
		"out_trade_no": req.OutTradeNo,
		"notify_url":   req.NotifyURL,
		"return_url":   req.ReturnURL,
		"name":         req.Name,
		"money":        req.Money,
		"clientip":     req.ClientIP,
	}
	if req.Device != "" {
		params["device"] = req.Device
	}

	params["sign"] = md5Sign(params, c.key)
	params["sign_type"] = "MD5"

	vals := url.Values{}
	for k, v := range params {
		vals.Set(k, v)
	}
	return c.apiURL + "submit.php?" + vals.Encode(), nil
}

// CreatePayment 通过 API 创建支付订单（mapi.php）
func (c *Client) CreatePayment(req CreatePaymentRequest) (*CreatePaymentResponse, error) {
	params := map[string]string{
		"pid":          c.pid,
		"type":         req.Type,
		"out_trade_no": req.OutTradeNo,
		"notify_url":   req.NotifyURL,
		"return_url":   req.ReturnURL,
		"name":         req.Name,
		"money":        req.Money,
		"clientip":     req.ClientIP,
	}
	if req.Device != "" {
		params["device"] = req.Device
	}

	params["sign"] = md5Sign(params, c.key)
	params["sign_type"] = "MD5"

	vals := url.Values{}
	for k, v := range params {
		vals.Set(k, v)
	}

	resp, err := c.httpClient.Post(
		c.apiURL+"mapi.php",
		"application/x-www-form-urlencoded",
		strings.NewReader(vals.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("epay: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("epay: read response: %w", err)
	}

	logger.LegacyPrintf("epay", "create payment response: %s", string(body))

	var result CreatePaymentResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("epay: decode response: %w", err)
	}

	if result.Code != 1 {
		return nil, fmt.Errorf("epay: create payment failed: %s", result.Msg)
	}

	return &result, nil
}

// QueryOrder 查询订单状态
func (c *Client) QueryOrder(tradeNo string) (*QueryOrderResponse, error) {
	u := fmt.Sprintf("%sapi.php?act=order&pid=%s&key=%s&trade_no=%s",
		c.apiURL, c.pid, c.key, url.QueryEscape(tradeNo))

	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("epay: query request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("epay: read response: %w", err)
	}

	logger.LegacyPrintf("epay", "query order response: trade_no=%s body=%s", tradeNo, string(body))

	var result QueryOrderResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("epay: decode query response: %w", err)
	}
	if result.Code != 1 {
		return nil, fmt.Errorf("epay: query order failed: %s", result.Msg)
	}
	return &result, nil
}

// VerifyNotify 验证异步回调签名（MD5）
func (c *Client) VerifyNotify(params map[string]string) bool {
	sign := params["sign"]
	if sign == "" {
		logger.LegacyPrintf("epay", "verify notify failed: sign is empty, params=%v", params)
		return false
	}

	if !md5Verify(params, c.key, sign) {
		logger.LegacyPrintf("epay", "verify notify failed: MD5 signature mismatch, params=%v", params)
		return false
	}

	return true
}
