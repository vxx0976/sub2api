package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/payment"
	"github.com/google/uuid"
)

type RechargeOrderService struct {
	rechargeOrderRepo RechargeOrderRepository
	userRepo          UserRepository
	settingService    *SettingService
	musePayment       *payment.MusePayment
	cfg               *config.Config
}

func NewRechargeOrderService(
	rechargeOrderRepo RechargeOrderRepository,
	userRepo UserRepository,
	settingService *SettingService,
	musePayment *payment.MusePayment,
	cfg *config.Config,
) *RechargeOrderService {
	return &RechargeOrderService{
		rechargeOrderRepo: rechargeOrderRepo,
		userRepo:          userRepo,
		settingService:    settingService,
		musePayment:       musePayment,
		cfg:               cfg,
	}
}

// CreateRechargeOrder 创建充值订单
func (s *RechargeOrderService) CreateRechargeOrder(ctx context.Context, userID int64, amount float64) (*CreateOrderOutput, error) {
	// 1. 获取充值配置
	config, err := s.settingService.GetRechargeConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get recharge config: %w", err)
	}

	if !config.Enabled {
		return nil, ErrRechargeDisabled
	}

	// 2. 验证金额范围
	if amount < config.MinAmount || amount > config.MaxAmount {
		return nil, ErrInvalidRechargeAmount
	}

	// 3. 计算倍率和到账金额
	multiplier := s.calculateMultiplier(amount, config.Tiers)
	creditAmount := amount * multiplier

	// 4. 生成订单号（加 R 前缀区分充值订单）
	orderNo := fmt.Sprintf("R%s", generateRechargeOrderNo())

	// 5. 设置过期时间
	expiredAt := time.Now().Add(30 * time.Minute)

	// 6. 创建订单记录
	order := &RechargeOrder{
		OrderNo:      orderNo,
		UserID:       userID,
		Amount:       amount,
		CreditAmount: creditAmount,
		Multiplier:   multiplier,
		Status:       RechargeOrderStatusPending,
		ExpiredAt:    &expiredAt,
	}

	if err := s.rechargeOrderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create recharge order: %w", err)
	}

	// 7. 生成支付URL
	payURL := s.musePayment.PaymentURL(payment.CreatePayParams{
		OrderNo: orderNo,
		Money:   amount, // 用户实际支付金额
		Name:    "账户充值",
	})

	return &CreateOrderOutput{
		OrderNo:      orderNo,
		PayURL:       payURL,
		Amount:       amount,
		CreditAmount: &creditAmount,
		Multiplier:   &multiplier,
	}, nil
}

// calculateMultiplier 根据金额计算倍率
func (s *RechargeOrderService) calculateMultiplier(amount float64, tiers []RechargeTier) float64 {
	for _, tier := range tiers {
		if amount >= tier.Min {
			if tier.Max == nil || amount <= *tier.Max {
				return tier.Multiplier
			}
		}
	}
	return 1.0 // 默认倍率
}

// HandleRechargeNotify 处理充值回调
func (s *RechargeOrderService) HandleRechargeNotify(ctx context.Context, orderNo string, tradeNo string, payType string) error {
	// 1. 获取订单
	order, err := s.rechargeOrderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return err
	}

	// 2. 检查订单状态（幂等性）
	if order.Status == RechargeOrderStatusPaid {
		return nil // 已处理
	}

	if order.Status == RechargeOrderStatusExpired {
		return ErrRechargeOrderExpired
	}

	// 3. 增加用户余额（使用 credit_amount）
	if err := s.userRepo.UpdateBalance(ctx, order.UserID, order.CreditAmount); err != nil {
		return fmt.Errorf("increase balance: %w", err)
	}

	// 4. 更新订单状态
	if err := s.rechargeOrderRepo.SetPaid(ctx, order.ID, tradeNo, payType); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	return nil
}

// RepayRechargeOrder 继续支付
func (s *RechargeOrderService) RepayRechargeOrder(ctx context.Context, userID int64, orderNo string) (*CreateOrderOutput, error) {
	// 验证功能是否启用
	config, err := s.settingService.GetRechargeConfig(ctx)
	if err != nil {
		return nil, err
	}
	if !config.Enabled {
		return nil, ErrRechargeDisabled
	}

	// 获取订单
	order, err := s.rechargeOrderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}

	// 验证所有权
	if order.UserID != userID {
		return nil, ErrRechargeOrderNotFound
	}

	// 检查订单状态
	if order.Status != RechargeOrderStatusPending {
		if order.Status == RechargeOrderStatusPaid {
			return nil, ErrRechargeOrderPaid
		}
		return nil, ErrRechargeOrderExpired
	}

	// 检查是否过期
	if order.ExpiredAt != nil && time.Now().After(*order.ExpiredAt) {
		_ = s.rechargeOrderRepo.UpdateStatus(ctx, order.ID, RechargeOrderStatusExpired, nil, nil)
		return nil, ErrRechargeOrderExpired
	}

	// 生成新的支付URL
	payURL := s.musePayment.PaymentURL(payment.CreatePayParams{
		OrderNo: order.OrderNo,
		Money:   order.Amount,
		Name:    "账户充值",
	})

	return &CreateOrderOutput{
		OrderNo:      order.OrderNo,
		PayURL:       payURL,
		Amount:       order.Amount,
		CreditAmount: &order.CreditAmount,
		Multiplier:   &order.Multiplier,
	}, nil
}

// GetRechargeOrders 获取用户充值订单列表
func (s *RechargeOrderService) GetRechargeOrders(ctx context.Context, userID int64, params pagination.PaginationParams) ([]RechargeOrder, *pagination.PaginationResult, error) {
	return s.rechargeOrderRepo.ListByUserID(ctx, userID, params)
}

// GetAllRechargeOrders 获取所有充值订单（管理员）
func (s *RechargeOrderService) GetAllRechargeOrders(ctx context.Context, params pagination.PaginationParams, keyword string, status string) ([]RechargeOrder, *pagination.PaginationResult, error) {
	return s.rechargeOrderRepo.ListAll(ctx, params, keyword, status)
}

// generateRechargeOrderNo generates a unique recharge order number
func generateRechargeOrderNo() string {
	now := time.Now()
	uid := uuid.New().String()[:8]
	return fmt.Sprintf("%s%s", now.Format("20060102150405"), uid)
}
