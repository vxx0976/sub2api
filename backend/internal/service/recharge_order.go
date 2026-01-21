package service

import (
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// 充值订单状态
const (
	RechargeOrderStatusPending  = "pending"
	RechargeOrderStatusPaid     = "paid"
	RechargeOrderStatusExpired  = "expired"
	RechargeOrderStatusRefunded = "refunded"
)

// 错误定义
var (
	ErrRechargeDisabled      = infraerrors.BadRequest("RECHARGE_DISABLED", "recharge feature is disabled")
	ErrInvalidRechargeAmount = infraerrors.BadRequest("INVALID_RECHARGE_AMOUNT", "recharge amount must be between min and max")
	ErrRechargeOrderNotFound = infraerrors.NotFound("RECHARGE_ORDER_NOT_FOUND", "recharge order not found")
	ErrRechargeOrderExpired  = infraerrors.BadRequest("RECHARGE_ORDER_EXPIRED", "recharge order has expired")
	ErrRechargeOrderPaid     = infraerrors.Conflict("RECHARGE_ORDER_PAID", "recharge order has already been paid")
)

// 充值订单模型
type RechargeOrder struct {
	ID           int64      `json:"id"`
	OrderNo      string     `json:"order_no"`
	TradeNo      *string    `json:"trade_no,omitempty"`
	UserID       int64      `json:"user_id"`
	Amount       float64    `json:"amount"`
	CreditAmount float64    `json:"credit_amount"`
	Multiplier   float64    `json:"multiplier"`
	Status       string     `json:"status"`
	PayType      *string    `json:"pay_type,omitempty"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ExpiredAt    *time.Time `json:"expired_at,omitempty"`

	User *User `json:"user,omitempty"`
}

// 充值阶梯配置
type RechargeTier struct {
	Min        float64  `json:"min"`
	Max        *float64 `json:"max"`
	Multiplier float64  `json:"multiplier"`
}

// 充值配置
type RechargeConfig struct {
	Enabled   bool           `json:"enabled"`
	MinAmount float64        `json:"min_amount"`
	MaxAmount float64        `json:"max_amount"`
	Tiers     []RechargeTier `json:"tiers"`
}
