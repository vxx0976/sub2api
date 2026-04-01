package epay

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client 彩虹易支付 SDK 客户端
type Client struct {
	apiURL     string
	pid        string
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	httpClient *http.Client
}

// NewClient 创建易支付客户端
func NewClient(apiURL, pid, platformPublicKey, merchantPrivateKey string) (*Client, error) {
	pubKey, err := parsePublicKey(platformPublicKey)
	if err != nil {
		return nil, fmt.Errorf("epay: parse platform public key: %w", err)
	}
	privKey, err := parsePrivateKey(merchantPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("epay: parse merchant private key: %w", err)
	}
	apiURL = strings.TrimRight(apiURL, "/") + "/"
	return &Client{
		apiURL:     apiURL,
		pid:        pid,
		publicKey:  pubKey,
		privateKey: privKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// GetPayLink 生成支付跳转链接（页面跳转模式）
func (c *Client) GetPayLink(req CreatePaymentRequest) (string, error) {
	params := map[string]string{
		"type":         req.Type,
		"out_trade_no": req.OutTradeNo,
		"notify_url":   req.NotifyURL,
		"return_url":   req.ReturnURL,
		"name":         req.Name,
		"money":        req.Money,
	}
	signed, err := c.buildRequestParams(params)
	if err != nil {
		return "", err
	}

	vals := url.Values{}
	for k, v := range signed {
		vals.Set(k, v)
	}
	return c.apiURL + "api/pay/submit?" + vals.Encode(), nil
}

// CreatePayment 通过 API 创建支付订单
func (c *Client) CreatePayment(req CreatePaymentRequest) (*CreatePaymentResponse, error) {
	params := map[string]string{
		"type":         req.Type,
		"out_trade_no": req.OutTradeNo,
		"notify_url":   req.NotifyURL,
		"return_url":   req.ReturnURL,
		"name":         req.Name,
		"money":        req.Money,
	}

	body, err := c.doPost("api/pay/create", params)
	if err != nil {
		return nil, err
	}

	var resp CreatePaymentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("epay: decode response: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("epay: create payment failed: %s", resp.Msg)
	}
	return &resp, nil
}

// QueryOrder 查询订单状态
func (c *Client) QueryOrder(tradeNo string) (*QueryOrderResponse, error) {
	params := map[string]string{
		"trade_no": tradeNo,
	}

	body, err := c.doPost("api/pay/query", params)
	if err != nil {
		return nil, err
	}

	var resp QueryOrderResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("epay: decode query response: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("epay: query order failed: %s", resp.Msg)
	}
	return &resp, nil
}

// VerifyNotify 验证异步回调签名
func (c *Client) VerifyNotify(params map[string]string) bool {
	sign := params["sign"]
	if sign == "" {
		return false
	}

	// 时间戳校验（5 分钟窗口）
	if ts := params["timestamp"]; ts != "" {
		t, err := strconv.ParseInt(ts, 10, 64)
		if err != nil || math.Abs(float64(time.Now().Unix()-t)) > 300 {
			return false
		}
	}

	content := buildSignContent(params)
	return rsaVerify(c.publicKey, content, sign)
}

func (c *Client) buildRequestParams(params map[string]string) (map[string]string, error) {
	params["pid"] = c.pid
	params["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)

	content := buildSignContent(params)
	sign, err := rsaSign(c.privateKey, content)
	if err != nil {
		return nil, fmt.Errorf("epay: sign request: %w", err)
	}
	params["sign"] = sign
	params["sign_type"] = "RSA"
	return params, nil
}

func (c *Client) doPost(path string, params map[string]string) ([]byte, error) {
	signed, err := c.buildRequestParams(params)
	if err != nil {
		return nil, err
	}

	vals := url.Values{}
	for k, v := range signed {
		vals.Set(k, v)
	}

	resp, err := c.httpClient.Post(c.apiURL+path, "application/x-www-form-urlencoded", strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, fmt.Errorf("epay: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("epay: read response: %w", err)
	}

	// 验签响应（仅在有 sign 字段时验证，失败不阻断）
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err == nil {
		if sign, ok := raw["sign"].(string); ok && sign != "" {
			strParams := make(map[string]string)
			for k, v := range raw {
				switch val := v.(type) {
				case string:
					strParams[k] = val
				case float64:
					if val == float64(int64(val)) {
						strParams[k] = strconv.FormatInt(int64(val), 10)
					} else {
						strParams[k] = strconv.FormatFloat(val, 'f', -1, 64)
					}
				}
			}
			if !rsaVerify(c.publicKey, buildSignContent(strParams), sign) {
				// 响应验签失败仅记录日志，不阻断（部分平台签名格式不一致）
				fmt.Printf("[epay] warning: response signature verification failed\n")
			}
		}
	}

	return body, nil
}
