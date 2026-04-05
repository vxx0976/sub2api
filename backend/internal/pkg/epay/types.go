package epay

import (
	"encoding/json"
	"strconv"
)

// FlexInt 兼容 JSON 中数字和字符串两种格式的整数，如 1 和 "1"
type FlexInt int

func (fi *FlexInt) UnmarshalJSON(data []byte) error {
	// 先尝试数字
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*fi = FlexInt(n)
		return nil
	}
	// 再尝试字符串
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*fi = FlexInt(n)
	return nil
}

// CreatePaymentRequest 创建支付请求参数
type CreatePaymentRequest struct {
	OutTradeNo string  `json:"out_trade_no"` // 商户订单号
	Type       string  `json:"type"`         // 支付方式: alipay/wxpay
	Name       string  `json:"name"`         // 商品名称
	Money      string  `json:"money"`        // 支付金额
	NotifyURL  string  `json:"notify_url"`   // 异步通知地址
	ReturnURL  string  `json:"return_url"`   // 同步跳转地址
	ClientIP   string  `json:"clientip"`     // 用户IP地址（API模式必填）
}

// CreatePaymentResponse API 创建支付响应
// 兼容两种易支付响应格式:
//   - 旧版: payurl, qrcode, urlscheme
//   - 新版: pay_type + pay_info
type CreatePaymentResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	TradeNo   string `json:"trade_no"`
	PayURL    string `json:"payurl"`
	QRCode    string `json:"qrcode"`
	URLScheme string `json:"urlscheme"`
	PayType   string `json:"pay_type"` // 新版: qrcode / url / ...
	PayInfo   string `json:"pay_info"` // 新版: 支付链接
}

// QueryOrderResponse 订单查询响应
type QueryOrderResponse struct {
	Code    FlexInt `json:"code"`
	Msg     string  `json:"msg"`
	TradeNo string  `json:"trade_no"`
	Money   string  `json:"money"`
	Status  FlexInt `json:"status"` // 1=已支付
	Type    string  `json:"type"`
}

// NotifyParams 异步回调参数
type NotifyParams struct {
	OutTradeNo  string `json:"out_trade_no"`
	TradeNo     string `json:"trade_no"`
	TradeStatus string `json:"trade_status"`
	Type        string `json:"type"`
	Money       string `json:"money"`
	Sign        string `json:"sign"`
	SignType    string `json:"sign_type"`
	Timestamp   string `json:"timestamp"`
}
