package repository

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type RechargeOrderRepository interface {
	Create(ctx context.Context, order *service.RechargeOrder) error
	GetByID(ctx context.Context, id int64) (*service.RechargeOrder, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*service.RechargeOrder, error)
	Update(ctx context.Context, order *service.RechargeOrder) error
	UpdateStatus(ctx context.Context, id int64, status string, tradeNo *string, payType *string) error
	SetPaid(ctx context.Context, id int64, tradeNo string, payType string) error
	ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.RechargeOrder, *pagination.PaginationResult, error)
	ListAll(ctx context.Context, params pagination.PaginationParams, keyword string, status string) ([]service.RechargeOrder, *pagination.PaginationResult, error)
}
