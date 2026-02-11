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
}
