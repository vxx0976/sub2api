package service

import (
	"context"
	"errors"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// Order AliMPay 充值订单领域模型（个人免签支付）
// 与 EPAY RechargeOrder 并列存在，业务语义一致：CNY 支付 → USD 余额入账
// 差异：支付通道走支付宝账单轮询（无 webhook），订单多一个 payment_amount（唯一金额用于账单匹配）
type Order struct {
	ID            int64      `json:"id"`
	OrderNo       string     `json:"order_no"`
	TradeNo       string     `json:"trade_no"`
	UserID        int64      `json:"user_id"`
	GroupID       int64      `json:"group_id"`
	Amount        float64    `json:"amount"`         // 基础金额（CNY）
	PaymentAmount float64    `json:"payment_amount"` // 实际支付金额（金额偏移后，用于账单匹配）
	CreditAmount  float64    `json:"credit_amount"`  // 到账余额（USD）
	Multiplier    float64    `json:"multiplier"`
	Status        string     `json:"status"` // pending/paid/expired/refunded
	PayType       string     `json:"pay_type"`
	PaidAt        *time.Time `json:"paid_at"`
	SourceDomain  string     `json:"source_domain"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ExpiredAt     *time.Time `json:"expired_at"`
}

// OrderPublicConfig 充值页面配置（AliMPay 下单页返回给前端）
// 字段含义与 RechargePublicConfig 一致，限额从同一 Setting key 读，和 EPAY 共享
type OrderPublicConfig struct {
	Enabled      bool    `json:"enabled"`
	MinAmount    float64 `json:"min_amount"`
	MaxAmount    float64 `json:"max_amount"`
	SellingPrice float64 `json:"selling_price"`
	Mode         string  `json:"mode"` // business_qr | transfer
}

// OrderRepository AliMPay 订单数据访问接口
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	CreateWithUniquePaymentAmount(ctx context.Context, order *Order, baseAmount float64, amountOffset float64, reuseWindow time.Duration) error
	GetByOrderNo(ctx context.Context, orderNo string) (*Order, error)
	UpdateStatus(ctx context.Context, orderNo, fromStatus, toStatus string, tradeNo *string, paidAt *time.Time) error
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*Order, int, error)
	ExpirePendingOrders(ctx context.Context) (int, error)
	ListAll(ctx context.Context, status string, userID *int64, limit, offset int) ([]*Order, int, error)
	ListPending(ctx context.Context) ([]*Order, error)
	SumPaidCreditByUserIDs(ctx context.Context, userIDs []int64) (float64, error)
	ListPaidByUserIDs(ctx context.Context, userIDs []int64, limit, offset int) ([]*Order, int, error)
	// SumPaidCreditByDay 汇总指定时区下每日 status='paid' 订单的 credit_amount（USD）。
	// 返回 map key 为 "YYYY-MM-DD"，value 为当日到账余额合计。
	SumPaidCreditByDay(ctx context.Context, startTime, endTime time.Time, tzName string) (map[string]float64, error)
	// TradeNoUsedByPaidOrder 判断某支付平台账单号（trade_no）是否已被其它 paid 订单占用。
	// 用于入账前的二次入账兜底校验：内存去重（AlipayMonitor.matchedBills）在多实例/重启后失效，
	// 这里从 DB 侧兜底，防止同一笔账单被两个订单先后入账。excludeOrderNo 为当前订单自身。
	TradeNoUsedByPaidOrder(ctx context.Context, tradeNo, excludeOrderNo string) (bool, error)
}

var ErrOrderStatusConflict = errors.New("order status conflict")
var ErrOrderPaymentAmountUnavailable = errors.New("no available payment amount")

// ErrAliMPayOrderNotFound 退款时指定订单号不存在。
var ErrAliMPayOrderNotFound = infraerrors.NotFound("ALIMPAY_ORDER_NOT_FOUND", "order not found")

// ErrOrderNotRefundable 仅 status='paid' 的 AliMPay 订单可退款。
var ErrOrderNotRefundable = infraerrors.BadRequest("ALIMPAY_ORDER_NOT_REFUNDABLE", "only paid orders can be refunded")
