package service

import (
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrOrderNotFound      = infraerrors.NotFound("ORDER_NOT_FOUND", "order not found")
	ErrOrderAlreadyPaid   = infraerrors.Conflict("ORDER_ALREADY_PAID", "order has already been paid")
	ErrOrderExpired       = infraerrors.BadRequest("ORDER_EXPIRED", "order has expired")
	ErrPaymentDisabled    = infraerrors.BadRequest("PAYMENT_DISABLED", "payment feature is disabled")
	ErrGroupNotPurchasable = infraerrors.BadRequest("GROUP_NOT_PURCHASABLE", "group is not available for purchase")
)

// Order status constants
const (
	OrderStatusPending  = "pending"
	OrderStatusPaid     = "paid"
	OrderStatusExpired  = "expired"
	OrderStatusRefunded = "refunded"
)

// Order represents a payment order
type Order struct {
	ID             int64      `json:"id"`
	OrderNo        string     `json:"order_no"`
	TradeNo        *string    `json:"trade_no,omitempty"`
	UserID         int64      `json:"user_id"`
	GroupID        int64      `json:"group_id"`
	Amount         float64    `json:"amount"`
	Status         string     `json:"status"`
	PayType        *string    `json:"pay_type,omitempty"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	SubscriptionID *int64     `json:"subscription_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ExpiredAt      *time.Time `json:"expired_at,omitempty"`

	User         *User             `json:"user,omitempty"`
	Group        *Group            `json:"group,omitempty"`
	Subscription *UserSubscription `json:"subscription,omitempty"`
}

// IsPending checks if the order is pending payment
func (o *Order) IsPending() bool {
	return o.Status == OrderStatusPending
}

// IsPaid checks if the order is paid
func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusPaid
}

// IsExpired checks if the order has expired
func (o *Order) IsExpired() bool {
	if o.ExpiredAt == nil {
		return false
	}
	return time.Now().After(*o.ExpiredAt)
}
