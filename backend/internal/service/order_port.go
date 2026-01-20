package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// OrderRepository defines the interface for order persistence operations
type OrderRepository interface {
	// Create creates a new order
	Create(ctx context.Context, order *Order) error

	// GetByID retrieves an order by ID
	GetByID(ctx context.Context, id int64) (*Order, error)

	// GetByOrderNo retrieves an order by order number
	GetByOrderNo(ctx context.Context, orderNo string) (*Order, error)

	// Update updates an order
	Update(ctx context.Context, order *Order) error

	// UpdateStatus updates the order status and related fields
	UpdateStatus(ctx context.Context, id int64, status string, tradeNo *string, payType *string) error

	// SetPaid marks the order as paid
	SetPaid(ctx context.Context, id int64, tradeNo string, payType string, subscriptionID int64) error

	// ListByUserID retrieves orders for a user with pagination
	ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Order, *pagination.PaginationResult, error)

	// ListAll retrieves all orders with pagination and filters (admin use)
	ListAll(ctx context.Context, params pagination.PaginationParams, status string, search string) ([]Order, *pagination.PaginationResult, error)

	// ListPending retrieves pending orders that are about to expire
	ListPending(ctx context.Context) ([]Order, error)

	// BatchUpdateExpired updates expired pending orders
	BatchUpdateExpired(ctx context.Context) (int64, error)
}
