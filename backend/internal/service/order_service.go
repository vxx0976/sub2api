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

// OrderService handles order-related business logic
type OrderService struct {
	orderRepo           OrderRepository
	groupRepo           GroupRepository
	subscriptionService *SubscriptionService
	musePayment         *payment.MusePayment
	cfg                 *config.Config
}

// NewOrderService creates a new order service
func NewOrderService(
	orderRepo OrderRepository,
	groupRepo GroupRepository,
	subscriptionService *SubscriptionService,
	musePayment *payment.MusePayment,
	cfg *config.Config,
) *OrderService {
	return &OrderService{
		orderRepo:           orderRepo,
		groupRepo:           groupRepo,
		subscriptionService: subscriptionService,
		musePayment:         musePayment,
		cfg:                 cfg,
	}
}

// CreateOrderInput input for creating an order
type CreateOrderInput struct {
	UserID  int64
	GroupID int64
}

// CreateOrderOutput output for creating an order
type CreateOrderOutput struct {
	OrderNo      string
	PayURL       string
	Amount       float64
	CreditAmount *float64 // 充值订单实际到账金额（可选）
	Multiplier   *float64 // 充值订单倍率（可选）
}

// CreateOrder creates a new order for purchasing a subscription
func (s *OrderService) CreateOrder(ctx context.Context, input *CreateOrderInput) (*CreateOrderOutput, error) {
	// Check if payment is enabled
	if !s.cfg.Payment.Enabled {
		return nil, ErrPaymentDisabled
	}

	// Get the group and check if it's purchasable
	group, err := s.groupRepo.GetByID(ctx, input.GroupID)
	if err != nil {
		return nil, err
	}

	if !group.IsPurchasable || group.Price == nil || *group.Price <= 0 {
		return nil, ErrGroupNotPurchasable
	}

	// Generate order number
	orderNo := generateOrderNo()

	// Calculate expiration time (30 minutes)
	expiredAt := time.Now().Add(30 * time.Minute)

	// Create order
	order := &Order{
		OrderNo:   orderNo,
		UserID:    input.UserID,
		GroupID:   input.GroupID,
		Amount:    *group.Price,
		Status:    OrderStatusPending,
		ExpiredAt: &expiredAt,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// Generate payment URL
	payURL := s.musePayment.PaymentURL(payment.CreatePayParams{
		OrderNo: orderNo,
		Money:   *group.Price,
		Name:    fmt.Sprintf("%s 订阅套餐", group.Name),
	})

	return &CreateOrderOutput{
		OrderNo: orderNo,
		PayURL:  payURL,
		Amount:  *group.Price,
	}, nil
}

// HandlePaymentNotify handles payment callback notification
func (s *OrderService) HandlePaymentNotify(ctx context.Context, params *payment.NotifyParams) error {
	// Verify signature
	if !s.musePayment.VerifySign(params.ToMap()) {
		return fmt.Errorf("invalid signature")
	}

	// Check payment status
	if !params.IsSuccess() {
		return nil // Not a success notification, ignore
	}

	// Get order
	order, err := s.orderRepo.GetByOrderNo(ctx, params.OutTradeNo)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	// Check if already paid
	if order.Status == OrderStatusPaid {
		return nil // Already processed
	}

	// Check if expired
	if order.Status == OrderStatusExpired {
		return ErrOrderExpired
	}

	// Get group to determine validity days
	group, err := s.groupRepo.GetByID(ctx, order.GroupID)
	if err != nil {
		return fmt.Errorf("get group: %w", err)
	}

	// Assign subscription
	validityDays := group.DefaultValidityDays
	if validityDays <= 0 {
		validityDays = 30 // Default to 30 days
	}

	sub, _, err := s.subscriptionService.AssignOrExtendSubscription(ctx, &AssignSubscriptionInput{
		UserID:       order.UserID,
		GroupID:      order.GroupID,
		ValidityDays: validityDays,
		Notes:        fmt.Sprintf("订单支付: %s", order.OrderNo),
	})
	if err != nil {
		return fmt.Errorf("assign subscription: %w", err)
	}

	// Update order status
	if err := s.orderRepo.SetPaid(ctx, order.ID, params.TradeNo, params.Type, sub.ID); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	return nil
}

// GetUserOrders retrieves orders for a user
func (s *OrderService) GetUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]Order, *pagination.PaginationResult, error) {
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	return s.orderRepo.ListByUserID(ctx, userID, params)
}

// GetAllOrders retrieves all orders with pagination and filters (admin use)
func (s *OrderService) GetAllOrders(ctx context.Context, page, pageSize int, status, search string) ([]Order, *pagination.PaginationResult, error) {
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	return s.orderRepo.ListAll(ctx, params, status, search)
}

// GetOrderByNo retrieves an order by order number
func (s *OrderService) GetOrderByNo(ctx context.Context, orderNo string) (*Order, error) {
	return s.orderRepo.GetByOrderNo(ctx, orderNo)
}

// GetPurchasablePlans retrieves all purchasable plans
func (s *OrderService) GetPurchasablePlans(ctx context.Context) ([]Group, error) {
	if !s.cfg.Payment.Enabled {
		return nil, ErrPaymentDisabled
	}
	return s.groupRepo.ListPurchasable(ctx)
}

// RepayOrder generates a new payment URL for an existing pending order
func (s *OrderService) RepayOrder(ctx context.Context, userID int64, orderNo string) (*CreateOrderOutput, error) {
	// Check if payment is enabled
	if !s.cfg.Payment.Enabled {
		return nil, ErrPaymentDisabled
	}

	// Get order
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}

	// Check if order belongs to user
	if order.UserID != userID {
		return nil, ErrOrderNotFound
	}

	// Check if order is pending
	if order.Status != OrderStatusPending {
		if order.Status == OrderStatusPaid {
			return nil, ErrOrderAlreadyPaid
		}
		return nil, ErrOrderExpired
	}

	// Check if order has expired
	if order.ExpiredAt != nil && time.Now().After(*order.ExpiredAt) {
		// Update status to expired
		_ = s.orderRepo.UpdateStatus(ctx, order.ID, OrderStatusExpired, nil, nil)
		return nil, ErrOrderExpired
	}

	// Get group for name
	group, err := s.groupRepo.GetByID(ctx, order.GroupID)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}

	// Generate payment URL
	payURL := s.musePayment.PaymentURL(payment.CreatePayParams{
		OrderNo: order.OrderNo,
		Money:   order.Amount,
		Name:    fmt.Sprintf("%s 订阅套餐", group.Name),
	})

	return &CreateOrderOutput{
		OrderNo: order.OrderNo,
		PayURL:  payURL,
		Amount:  order.Amount,
	}, nil
}

// generateOrderNo generates a unique order number
func generateOrderNo() string {
	now := time.Now()
	uid := uuid.New().String()[:8]
	return fmt.Sprintf("%s%s", now.Format("20060102150405"), uid)
}
