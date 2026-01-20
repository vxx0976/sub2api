// Package payment provides payment gateway integrations.
package payment

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// MusePayment 暮色聚合支付客户端
type MusePayment struct {
	cfg config.MusePaymentConfig
}

// NewMusePayment 创建暮色支付客户端
func NewMusePayment(cfg config.MusePaymentConfig) *MusePayment {
	return &MusePayment{cfg: cfg}
}

// CreatePayParams 创建支付参数
type CreatePayParams struct {
	OrderNo string  // 商户订单号
	Money   float64 // 金额
	Name    string  // 商品名称
	Type    string  // 支付类型: alipay/wxpay，空则跳转收银台
}

// PaymentURL 生成支付跳转URL
func (m *MusePayment) PaymentURL(params CreatePayParams) string {
	data := map[string]string{
		"pid":          m.cfg.PID,
		"out_trade_no": params.OrderNo,
		"notify_url":   m.cfg.NotifyURL,
		"return_url":   m.cfg.ReturnURL,
		"name":         params.Name,
		"money":        fmt.Sprintf("%.2f", params.Money),
		"sign_type":    "MD5",
	}
	if params.Type != "" {
		data["type"] = params.Type
	}
	data["sign"] = m.Sign(data)

	values := url.Values{}
	for k, v := range data {
		values.Set(k, v)
	}
	return m.cfg.SubmitURL + "?" + values.Encode()
}

// Sign 生成签名
// 签名规则：
// 1. 过滤空值和签名字段(sign/sign_type)
// 2. 按key字母排序
// 3. 拼接为 key=value&key=value 格式
// 4. 末尾追加商户密钥
// 5. MD5 加密
func (m *MusePayment) Sign(params map[string]string) string {
	// 过滤空值和签名字段
	filtered := make(map[string]string)
	for k, v := range params {
		if k != "sign" && k != "sign_type" && v != "" {
			filtered[k] = v
		}
	}

	// 按key排序
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, filtered[k]))
	}
	str := strings.Join(parts, "&") + m.cfg.Key

	// MD5
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

// VerifySign 验证签名
func (m *MusePayment) VerifySign(params map[string]string) bool {
	sign := params["sign"]
	if sign == "" {
		return false
	}
	return m.Sign(params) == sign
}

// NotifyParams 支付回调参数
type NotifyParams struct {
	PID         string `form:"pid" json:"pid"`
	TradeNo     string `form:"trade_no" json:"trade_no"`           // 支付平台订单号
	OutTradeNo  string `form:"out_trade_no" json:"out_trade_no"`   // 商户订单号
	Type        string `form:"type" json:"type"`                   // 支付类型
	Name        string `form:"name" json:"name"`                   // 商品名称
	Money       string `form:"money" json:"money"`                 // 金额
	TradeStatus string `form:"trade_status" json:"trade_status"`   // 交易状态
	Sign        string `form:"sign" json:"sign"`                   // 签名
	SignType    string `form:"sign_type" json:"sign_type"`         // 签名类型
}

// IsSuccess 检查支付是否成功
func (n *NotifyParams) IsSuccess() bool {
	return n.TradeStatus == "TRADE_SUCCESS"
}

// ToMap 转换为map用于验签
func (n *NotifyParams) ToMap() map[string]string {
	return map[string]string{
		"pid":          n.PID,
		"trade_no":     n.TradeNo,
		"out_trade_no": n.OutTradeNo,
		"type":         n.Type,
		"name":         n.Name,
		"money":        n.Money,
		"trade_status": n.TradeStatus,
		"sign":         n.Sign,
		"sign_type":    n.SignType,
	}
}
