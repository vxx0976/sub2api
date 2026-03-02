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

const platformUsdRate = 10.0 // 1 USDT = $10 platform balance

type RechargeOrderService struct {
	rechargeOrderRepo   RechargeOrderRepository
	redeemCodeRepo      RedeemCodeRepository
	userRepo            UserRepository
	settingService      *SettingService
	exchangeRateService *ExchangeRateService
	alipayPayment       *payment.AlipayPayment
	cfg                 *config.Config
}

func NewRechargeOrderService(
	rechargeOrderRepo RechargeOrderRepository,
	redeemCodeRepo RedeemCodeRepository,
	userRepo UserRepository,
	settingService *SettingService,
	exchangeRateService *ExchangeRateService,
	alipayPayment *payment.AlipayPayment,
	cfg *config.Config,
) *RechargeOrderService {
	return &RechargeOrderService{
		rechargeOrderRepo:   rechargeOrderRepo,
		redeemCodeRepo:      redeemCodeRepo,
		userRepo:            userRepo,
		settingService:      settingService,
		exchangeRateService: exchangeRateService,
		alipayPayment:       alipayPayment,
		cfg:                 cfg,
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

	// 2. 检查是否为 ¥1 特惠订单（所有用户可用一次）
	var multiplier float64
	var creditAmount float64
	isFirstRecharge := false

	const firstRechargePrice = 1.0
	const firstRechargeCredit = 5.0

	if amount == firstRechargePrice {
		hasPaid, err := s.rechargeOrderRepo.HasPaidOrder(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("check paid orders: %w", err)
		}
		if !hasPaid {
			isFirstRecharge = true
			creditAmount = firstRechargeCredit
			multiplier = firstRechargeCredit / firstRechargePrice
		}
	}

	if !isFirstRecharge {
		// 验证金额范围
		if amount < config.MinAmount || amount > config.MaxAmount {
			return nil, ErrInvalidRechargeAmount
		}
		// 获取实时汇率，计算到账金额：CNY → USD → 平台余额
		usdCnyRate := s.exchangeRateService.GetUsdCnyRate()
		if usdCnyRate <= 0 {
			usdCnyRate = config.UsdCnyRate // fallback to DB config
		}
		creditAmount = (amount / usdCnyRate) * platformUsdRate
		multiplier = creditAmount / amount // 记录实际换算比率
	}

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

	// 7. 生成支付信息（QR码）
	payInfo, err := s.alipayPayment.GeneratePaymentInfo(orderNo, amount)
	if err != nil {
		return nil, fmt.Errorf("generate payment info: %w", err)
	}

	qrCodeBase64, err := payment.GenerateQRCodeBase64(payInfo.QRCodeURL, 256)
	if err != nil {
		return nil, fmt.Errorf("generate QR code: %w", err)
	}

	return &CreateOrderOutput{
		OrderNo:       orderNo,
		Amount:        amount,
		PaymentAmount: payInfo.PaymentAmount,
		QRCodeURL:     payInfo.QRCodeURL,
		QRCode:        qrCodeBase64,
		Mode:          payInfo.Mode,
		CreditAmount:  &creditAmount,
		Multiplier:    &multiplier,
	}, nil
}

// HandleRechargeNotify 处理充值回调（由监控服务调用）
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

	// 5. 创建余额变更记录（用于管理员查看充值历史）
	code, err := GenerateRedeemCode()
	if err == nil {
		now := time.Now()
		notes := fmt.Sprintf("充值订单 %s，支付 ¥%.2f", orderNo, order.Amount)
		record := &RedeemCode{
			Code:   code,
			Type:   AdjustmentTypeRecharge,
			Value:  order.CreditAmount,
			Status: StatusUsed,
			UsedBy: &order.UserID,
			UsedAt: &now,
			Notes:  notes,
		}
		if err := s.redeemCodeRepo.Create(ctx, record); err != nil {
			// 不影响主流程，仅记录日志
			fmt.Printf("[RechargeOrder] failed to create balance history record: %v\n", err)
		}
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

	// 生成支付信息
	payInfo, err := s.alipayPayment.GeneratePaymentInfo(order.OrderNo, order.Amount)
	if err != nil {
		return nil, fmt.Errorf("generate payment info: %w", err)
	}

	qrCodeBase64, err := payment.GenerateQRCodeBase64(payInfo.QRCodeURL, 256)
	if err != nil {
		return nil, fmt.Errorf("generate QR code: %w", err)
	}

	return &CreateOrderOutput{
		OrderNo:       order.OrderNo,
		Amount:        order.Amount,
		PaymentAmount: payInfo.PaymentAmount,
		QRCodeURL:     payInfo.QRCodeURL,
		QRCode:        qrCodeBase64,
		Mode:          payInfo.Mode,
		CreditAmount:  &order.CreditAmount,
		Multiplier:    &order.Multiplier,
	}, nil
}

// GetRechargeOrderPaymentStatus returns the payment status of a recharge order
func (s *RechargeOrderService) GetRechargeOrderPaymentStatus(ctx context.Context, userID int64, orderNo string) (string, error) {
	order, err := s.rechargeOrderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return "", err
	}

	if order.UserID != userID {
		return "", ErrRechargeOrderNotFound
	}

	if order.Status == RechargeOrderStatusPending && order.ExpiredAt != nil && time.Now().After(*order.ExpiredAt) {
		_ = s.rechargeOrderRepo.UpdateStatus(ctx, order.ID, RechargeOrderStatusExpired, nil, nil)
		return RechargeOrderStatusExpired, nil
	}

	return order.Status, nil
}

// GetPendingRechargeOrdersForMonitor returns pending recharge orders for payment matching
func (s *RechargeOrderService) GetPendingRechargeOrdersForMonitor(ctx context.Context) ([]payment.PendingOrder, error) {
	orders, err := s.rechargeOrderRepo.ListPending(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := make([]payment.PendingOrder, 0, len(orders))
	for _, o := range orders {
		// Skip expired orders to prevent amount collisions
		if o.ExpiredAt != nil && o.ExpiredAt.Before(now) {
			continue
		}
		result = append(result, payment.PendingOrder{
			OrderNo:       o.OrderNo,
			Amount:        o.Amount,
			PaymentAmount: o.Amount,
			CreatedAt:     o.CreatedAt,
			ExpiredAt:     o.ExpiredAt,
		})
	}
	return result, nil
}

// GetRechargeOrders 获取用户充值订单列表
func (s *RechargeOrderService) GetRechargeOrders(ctx context.Context, userID int64, params pagination.PaginationParams) ([]RechargeOrder, *pagination.PaginationResult, error) {
	return s.rechargeOrderRepo.ListByUserID(ctx, userID, params)
}

// GetAllRechargeOrders 获取所有充值订单（管理员）
func (s *RechargeOrderService) GetAllRechargeOrders(ctx context.Context, params pagination.PaginationParams, keyword string, status string) ([]RechargeOrder, *pagination.PaginationResult, error) {
	return s.rechargeOrderRepo.ListAll(ctx, params, keyword, status)
}

// DeleteRechargeOrder deletes a pending recharge order (admin only)
func (s *RechargeOrderService) DeleteRechargeOrder(ctx context.Context, id int64) error {
	order, err := s.rechargeOrderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order.Status != RechargeOrderStatusPending {
		return ErrRechargeOrderNotPending
	}
	return s.rechargeOrderRepo.Delete(ctx, id)
}

// AdminMarkRechargeOrderPaid marks a pending recharge order as paid manually (admin only, no business logic triggered)
func (s *RechargeOrderService) AdminMarkRechargeOrderPaid(ctx context.Context, id int64) error {
	order, err := s.rechargeOrderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order.Status != RechargeOrderStatusPending {
		return ErrRechargeOrderNotPending
	}
	return s.rechargeOrderRepo.MarkManualPaid(ctx, id)
}

// FirstRechargeStatus 首充资格状态
type FirstRechargeStatus struct {
	Eligible bool    `json:"eligible"`
	Price    float64 `json:"price"`
	Credit   float64 `json:"credit"`
}

// CheckFirstRechargeEligibility 检查用户 ¥1 特惠资格（所有用户可用一次）
func (s *RechargeOrderService) CheckFirstRechargeEligibility(ctx context.Context, userID int64) (*FirstRechargeStatus, error) {
	hasPaid, err := s.rechargeOrderRepo.HasPaidOrder(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check paid orders: %w", err)
	}

	return &FirstRechargeStatus{
		Eligible: !hasPaid,
		Price:    1.0,
		Credit:   5.0,
	}, nil
}

// generateRechargeOrderNo generates a unique recharge order number
func generateRechargeOrderNo() string {
	now := time.Now()
	uid := uuid.New().String()[:8]
	return fmt.Sprintf("%s%s", now.Format("20060102150405"), uid)
}
