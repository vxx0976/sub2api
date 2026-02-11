package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/payment"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PurchasablePlan represents a plan available for purchase
type PurchasablePlan struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Price           float64  `json:"price"`
	ValidityDays    int      `json:"validity_days"`
	DailyLimitUSD   *float64 `json:"daily_limit_usd,omitempty"`
	WeeklyLimitUSD  *float64 `json:"weekly_limit_usd,omitempty"`
	MonthlyLimitUSD *float64 `json:"monthly_limit_usd,omitempty"`
	SortOrder       int      `json:"sort_order"`
	IsRecommended   bool     `json:"is_recommended"`
	ExternalBuyURL  *string  `json:"external_buy_url,omitempty"`
}

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	GroupID int64 `json:"group_id" binding:"required"`
}

// CreateOrderResponse represents the response for creating an order
type CreateOrderResponse struct {
	OrderNo       string  `json:"order_no"`
	Amount        float64 `json:"amount"`
	PaymentAmount float64 `json:"payment_amount"`
	QRCodeURL     string  `json:"qr_code_url"`
	QRCode        string  `json:"qr_code"`
	Mode          string  `json:"mode"`
}

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	orderService         *service.OrderService
	rechargeOrderService *service.RechargeOrderService
	alipayPayment        *payment.AlipayPayment
	cfg                  *config.Config
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(
	orderService *service.OrderService,
	rechargeOrderService *service.RechargeOrderService,
	alipayPayment *payment.AlipayPayment,
	cfg *config.Config,
) *PaymentHandler {
	return &PaymentHandler{
		orderService:         orderService,
		rechargeOrderService: rechargeOrderService,
		alipayPayment:        alipayPayment,
		cfg:                  cfg,
	}
}

// GetPlans handles getting purchasable plans
// GET /api/v1/plans
func (h *PaymentHandler) GetPlans(c *gin.Context) {
	var plans []service.Group
	var err error

	// Check if accessing from a reseller domain
	if domainCtx := middleware2.GetResellerDomainFromContext(c); domainCtx != nil {
		plans, err = h.orderService.GetPurchasablePlansForReseller(c.Request.Context(), domainCtx.ResellerID)
	} else {
		plans, err = h.orderService.GetPurchasablePlans(c.Request.Context())
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	result := make([]PurchasablePlan, 0, len(plans))
	for _, p := range plans {
		plan := PurchasablePlan{
			ID:            p.ID,
			Name:          p.Name,
			Description:   p.Description,
			ValidityDays:  p.DefaultValidityDays,
			SortOrder:     p.SortOrder,
			IsRecommended: p.IsRecommended,
		}
		if p.Price != nil {
			plan.Price = *p.Price
		}
		plan.DailyLimitUSD = p.DailyLimitUSD
		plan.WeeklyLimitUSD = p.WeeklyLimitUSD
		plan.MonthlyLimitUSD = p.MonthlyLimitUSD
		plan.ExternalBuyURL = p.ExternalBuyURL
		result = append(result, plan)
	}

	response.Success(c, result)
}

// CreateOrder handles creating a new order
// POST /api/v1/orders
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.orderService.CreateOrder(c.Request.Context(), &service.CreateOrderInput{
		UserID:  subject.UserID,
		GroupID: req.GroupID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, CreateOrderResponse{
		OrderNo:       result.OrderNo,
		Amount:        result.Amount,
		PaymentAmount: result.PaymentAmount,
		QRCodeURL:     result.QRCodeURL,
		QRCode:        result.QRCode,
		Mode:          result.Mode,
	})
}

// GetUserOrders handles getting current user's orders
// GET /api/v1/orders
func (h *PaymentHandler) GetUserOrders(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	page, pageSize := response.ParsePagination(c)

	orders, pag, err := h.orderService.GetUserOrders(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, orders, pag.Total, page, pageSize)
}

// RepayOrder handles repaying an existing pending order
// POST /api/v1/orders/:order_no/repay
func (h *PaymentHandler) RepayOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	orderNo := c.Param("order_no")
	if orderNo == "" {
		response.BadRequest(c, "order_no is required")
		return
	}

	result, err := h.orderService.RepayOrder(c.Request.Context(), subject.UserID, orderNo)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, CreateOrderResponse{
		OrderNo:       result.OrderNo,
		Amount:        result.Amount,
		PaymentAmount: result.PaymentAmount,
		QRCodeURL:     result.QRCodeURL,
		QRCode:        result.QRCode,
		Mode:          result.Mode,
	})
}

// GetOrderPaymentStatus handles polling for order payment status
// GET /api/v1/orders/:order_no/payment-status
func (h *PaymentHandler) GetOrderPaymentStatus(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	orderNo := c.Param("order_no")
	if orderNo == "" {
		response.BadRequest(c, "order_no is required")
		return
	}

	var status string
	var err error

	if len(orderNo) > 0 && orderNo[0] == 'R' {
		status, err = h.rechargeOrderService.GetRechargeOrderPaymentStatus(c.Request.Context(), subject.UserID, orderNo)
	} else {
		status, err = h.orderService.GetOrderPaymentStatus(c.Request.Context(), subject.UserID, orderNo)
	}

	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"status":   status,
		"order_no": orderNo,
	})
}
