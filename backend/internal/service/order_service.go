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
	alipayPayment       *payment.AlipayPayment
	cfg                 *config.Config
	referralService     *ReferralService
}

// NewOrderService creates a new order service
func NewOrderService(
	orderRepo OrderRepository,
	groupRepo GroupRepository,
	subscriptionService *SubscriptionService,
	alipayPayment *payment.AlipayPayment,
	cfg *config.Config,
	referralService *ReferralService,
) *OrderService {
	return &OrderService{
		orderRepo:           orderRepo,
		groupRepo:           groupRepo,
		subscriptionService: subscriptionService,
		alipayPayment:       alipayPayment,
		cfg:                 cfg,
		referralService:     referralService,
	}
}

// CreateOrderInput input for creating an order
type CreateOrderInput struct {
	UserID  int64
	GroupID int64
}

// CreateOrderOutput output for creating an order
type CreateOrderOutput struct {
	OrderNo       string
	Amount        float64
	PaymentAmount float64  // actual amount to pay (may differ for unique amount matching)
	QRCodeURL     string   // URL encoded as QR code
	QRCode        string   // base64 QR code image
	Mode          string   // "business_qr" or "transfer"
	CreditAmount  *float64 // 充值订单实际到账金额（可选）
	Multiplier    *float64 // 充值订单倍率（可选）
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

	// Generate payment info (QR code)
	payInfo, err := s.alipayPayment.GeneratePaymentInfo(orderNo, *group.Price)
	if err != nil {
		return nil, fmt.Errorf("generate payment info: %w", err)
	}

	qrCodeBase64, err := payment.GenerateQRCodeBase64(payInfo.QRCodeURL, 256)
	if err != nil {
		return nil, fmt.Errorf("generate QR code: %w", err)
	}

	return &CreateOrderOutput{
		OrderNo:       orderNo,
		Amount:        *group.Price,
		PaymentAmount: payInfo.PaymentAmount,
		QRCodeURL:     payInfo.QRCodeURL,
		QRCode:        qrCodeBase64,
		Mode:          payInfo.Mode,
	}, nil
}

// ConfirmOrderPaid is called by the payment monitor when a matching payment is detected
func (s *OrderService) ConfirmOrderPaid(ctx context.Context, orderNo string, tradeNo string, payType string) error {
	// Get order
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
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

	sub, isNewSubscription, err := s.subscriptionService.AssignOrExtendSubscription(ctx, &AssignSubscriptionInput{
		UserID:       order.UserID,
		GroupID:      order.GroupID,
		ValidityDays: validityDays,
		Notes:        fmt.Sprintf("订单支付: %s", order.OrderNo),
	})
	if err != nil {
		return fmt.Errorf("assign subscription: %w", err)
	}

	// Update order status
	if err := s.orderRepo.SetPaid(ctx, order.ID, tradeNo, payType, sub.ID); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	// Process referral reward for first-time subscription payments
	if s.referralService != nil && isNewSubscription {
		go func() {
			bgCtx := context.Background()
			referrerRewarded, inviteeRewarded, err := s.referralService.ProcessReferralReward(bgCtx, order.UserID, order.ID, order.Amount)
			if err != nil {
				fmt.Printf("[Referral] Failed to process reward for order %d: %v\n", order.ID, err)
			} else if referrerRewarded || inviteeRewarded {
				fmt.Printf("[Referral] Processed reward for order %d: referrer=%v, invitee=%v\n", order.ID, referrerRewarded, inviteeRewarded)
			}
		}()
	}

	return nil
}

// GetOrderPaymentStatus returns the payment status of an order
func (s *OrderService) GetOrderPaymentStatus(ctx context.Context, userID int64, orderNo string) (string, error) {
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return "", err
	}

	if order.UserID != userID {
		return "", ErrOrderNotFound
	}

	if order.Status == OrderStatusPending && order.ExpiredAt != nil && time.Now().After(*order.ExpiredAt) {
		_ = s.orderRepo.UpdateStatus(ctx, order.ID, OrderStatusExpired, nil, nil)
		return OrderStatusExpired, nil
	}

	return order.Status, nil
}

// GetPendingOrdersForMonitor returns all pending orders for payment matching
func (s *OrderService) GetPendingOrdersForMonitor(ctx context.Context) ([]payment.PendingOrder, error) {
	orders, err := s.orderRepo.ListPending(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]payment.PendingOrder, 0, len(orders))
	for _, o := range orders {
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

// GetPurchasablePlans retrieves all purchasable plans (admin-owned, for main site)
func (s *OrderService) GetPurchasablePlans(ctx context.Context) ([]Group, error) {
	return s.groupRepo.ListPurchasable(ctx)
}

// GetPurchasablePlansForReseller retrieves purchasable plans owned by the given reseller
func (s *OrderService) GetPurchasablePlansForReseller(ctx context.Context, ownerID int64) ([]Group, error) {
	return s.groupRepo.ListPurchasableByOwnerID(ctx, ownerID)
}

// RepayOrder generates a new payment QR code for an existing pending order
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

	// Generate payment info
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
	}, nil
}

// generateOrderNo generates a unique order number
func generateOrderNo() string {
	now := time.Now()
	uid := uuid.New().String()[:8]
	return fmt.Sprintf("%s%s", now.Format("20060102150405"), uid)
}
