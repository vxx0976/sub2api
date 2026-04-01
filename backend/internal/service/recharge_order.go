package service

import (
	"context"
	"errors"
	"time"
)

// RechargeOrder 充值订单领域模型
type RechargeOrder struct {
	ID           int64      `json:"id"`
	OrderNo      string     `json:"order_no"`
	TradeNo      string     `json:"trade_no"`
	UserID       int64      `json:"user_id"`
	Amount       float64    `json:"amount"`        // 实付金额（CNY）
	CreditAmount float64    `json:"credit_amount"` // 到账余额（USD）
	Multiplier   float64    `json:"multiplier"`    // 倍率快照
	Status       string     `json:"status"`        // pending/paid/expired/refunded
	PayType      string     `json:"pay_type"`      // alipay/wxpay
	PaidAt       *time.Time `json:"paid_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ExpiredAt    *time.Time `json:"expired_at"`
}

// RechargeTier 充值档位
type RechargeTier struct {
	Min        float64  `json:"min"`
	Max        *float64 `json:"max"`
	Multiplier float64  `json:"multiplier"`
}

// RechargePublicConfig 充值页面配置（返回给前端）
type RechargePublicConfig struct {
	Enabled      bool           `json:"enabled"`
	MinAmount    float64        `json:"min_amount"`
	MaxAmount    float64        `json:"max_amount"`
	Tiers        []RechargeTier `json:"tiers"`
	PayTypes     []string       `json:"pay_types"`
	SellingPrice float64        `json:"selling_price"` // ¥/USD
}

// RechargeOrderRepository 充值订单数据访问接口
type RechargeOrderRepository interface {
	Create(ctx context.Context, order *RechargeOrder) error
	GetByOrderNo(ctx context.Context, orderNo string) (*RechargeOrder, error)
	UpdateStatus(ctx context.Context, orderNo, fromStatus, toStatus string, tradeNo *string, paidAt *time.Time) error
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*RechargeOrder, int, error)
	ExpirePendingOrders(ctx context.Context) (int, error)
	ListAll(ctx context.Context, status string, userID *int64, limit, offset int) ([]*RechargeOrder, int, error)
	SumPaidCreditByUserIDs(ctx context.Context, userIDs []int64) (float64, error)
	ListPaidByUserIDs(ctx context.Context, userIDs []int64, limit, offset int) ([]*RechargeOrder, int, error)
}

var ErrRechargeOrderStatusConflict = errors.New("recharge order status conflict")
