package service

import (
	"context"
	"errors"
	"time"
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
}

var ErrOrderStatusConflict = errors.New("order status conflict")
var ErrOrderPaymentAmountUnavailable = errors.New("no available payment amount")
