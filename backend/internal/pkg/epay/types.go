package epay

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
type CreatePaymentResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	TradeNo   string `json:"trade_no"`
	PayURL    string `json:"payurl"`
	QRCode    string `json:"qrcode"`
	URLScheme string `json:"urlscheme"`
}

// QueryOrderResponse 订单查询响应
type QueryOrderResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	TradeNo string `json:"trade_no"`
	Money   string `json:"money"`
	Status  int    `json:"status"` // 1=已支付
	Type    string `json:"type"`
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
