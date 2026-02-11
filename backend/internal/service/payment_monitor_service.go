package service

import (
	"context"
	"log"

	"github.com/Wei-Shaw/sub2api/internal/pkg/payment"
)

// PaymentMonitorService implements payment.OrderMatcher
type PaymentMonitorService struct {
	orderService         *OrderService
	rechargeOrderService *RechargeOrderService
}

// NewPaymentMonitorService creates a new payment monitor service
func NewPaymentMonitorService(
	orderService *OrderService,
	rechargeOrderService *RechargeOrderService,
) *PaymentMonitorService {
	return &PaymentMonitorService{
		orderService:         orderService,
		rechargeOrderService: rechargeOrderService,
	}
}

// GetPendingOrders returns all pending orders for payment matching
func (s *PaymentMonitorService) GetPendingOrders(ctx context.Context) ([]payment.PendingOrder, error) {
	subscriptionOrders, err := s.orderService.GetPendingOrdersForMonitor(ctx)
	if err != nil {
		log.Printf("[PaymentMonitor] Failed to get pending subscription orders: %v", err)
	}

	rechargeOrders, err := s.rechargeOrderService.GetPendingRechargeOrdersForMonitor(ctx)
	if err != nil {
		log.Printf("[PaymentMonitor] Failed to get pending recharge orders: %v", err)
	}

	result := make([]payment.PendingOrder, 0, len(subscriptionOrders)+len(rechargeOrders))
	result = append(result, subscriptionOrders...)
	result = append(result, rechargeOrders...)
	return result, nil
}

// ConfirmOrderPaid marks an order as paid by the monitor
func (s *PaymentMonitorService) ConfirmOrderPaid(ctx context.Context, orderNo string, tradeNo string, payType string) error {
	if len(orderNo) > 0 && orderNo[0] == 'R' {
		return s.rechargeOrderService.HandleRechargeNotify(ctx, orderNo, tradeNo, payType)
	}
	return s.orderService.ConfirmOrderPaid(ctx, orderNo, tradeNo, payType)
}
