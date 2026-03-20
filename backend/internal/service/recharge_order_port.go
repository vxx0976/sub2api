package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type RechargeOrderRepository interface {
	Create(ctx context.Context, order *RechargeOrder) error
	GetByID(ctx context.Context, id int64) (*RechargeOrder, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*RechargeOrder, error)
	Update(ctx context.Context, order *RechargeOrder) error
	UpdateStatus(ctx context.Context, id int64, status string, tradeNo *string, payType *string) error
	SetPaid(ctx context.Context, id int64, tradeNo string, payType string) error
	ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams) ([]RechargeOrder, *pagination.PaginationResult, error)
	ListAll(ctx context.Context, params pagination.PaginationParams, keyword string, status string) ([]RechargeOrder, *pagination.PaginationResult, error)
	ListPending(ctx context.Context) ([]RechargeOrder, error)
	HasPaidOrder(ctx context.Context, userID int64) (bool, error)

	// Delete deletes a pending recharge order
	Delete(ctx context.Context, id int64) error

	// MarkManualPaid marks a recharge order as paid manually (without triggering business logic)
	MarkManualPaid(ctx context.Context, id int64) error

	// SumCreditByUserIDs returns the total credit_amount for paid orders of given user IDs
	SumCreditByUserIDs(ctx context.Context, userIDs []int64) (float64, error)

	// ListPaidByUserIDs returns paginated paid recharge orders for given user IDs, newest first
	ListPaidByUserIDs(ctx context.Context, userIDs []int64, limit, offset int) ([]*RechargeDetailRecord, int, error)
}
