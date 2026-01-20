package handler

import (
	"net/http"

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
}

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	GroupID int64 `json:"group_id" binding:"required"`
}

// CreateOrderResponse represents the response for creating an order
type CreateOrderResponse struct {
	OrderNo string  `json:"order_no"`
	PayURL  string  `json:"pay_url"`
	Amount  float64 `json:"amount"`
}

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	orderService *service.OrderService
	cfg          *config.Config
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(orderService *service.OrderService, cfg *config.Config) *PaymentHandler {
	return &PaymentHandler{
		orderService: orderService,
		cfg:          cfg,
	}
}

// GetPlans handles getting purchasable plans
// GET /api/v1/plans
func (h *PaymentHandler) GetPlans(c *gin.Context) {
	plans, err := h.orderService.GetPurchasablePlans(c.Request.Context())
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
		OrderNo: result.OrderNo,
		PayURL:  result.PayURL,
		Amount:  result.Amount,
	})
}

// PaymentNotify handles payment callback notification
// GET /api/v1/payment/notify
func (h *PaymentHandler) PaymentNotify(c *gin.Context) {
	var params payment.NotifyParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	if err := h.orderService.HandlePaymentNotify(c.Request.Context(), &params); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

// PaymentReturn handles payment completion redirect
// GET /api/v1/payment/return
func (h *PaymentHandler) PaymentReturn(c *gin.Context) {
	// Get order number from query
	orderNo := c.Query("out_trade_no")
	if orderNo == "" {
		c.Redirect(http.StatusFound, h.cfg.Payment.Muse.ReturnURL+"?status=error")
		return
	}

	// Check order status
	order, err := h.orderService.GetOrderByNo(c.Request.Context(), orderNo)
	if err != nil {
		c.Redirect(http.StatusFound, h.cfg.Payment.Muse.ReturnURL+"?status=error")
		return
	}

	// Redirect based on status
	status := "pending"
	if order.IsPaid() {
		status = "success"
	} else if order.IsExpired() {
		status = "expired"
	}

	c.Redirect(http.StatusFound, h.cfg.Payment.Muse.ReturnURL+"?status="+status+"&order_no="+orderNo)
}

// VerifyPaymentReturn handles verification of payment return parameters
// POST /api/v1/payment/verify
func (h *PaymentHandler) VerifyPaymentReturn(c *gin.Context) {
	var params payment.NotifyParams
	if err := c.ShouldBindJSON(&params); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Process the payment (verify signature, update order, assign subscription)
	if err := h.orderService.HandlePaymentNotify(c.Request.Context(), &params); err != nil {
		// Check if it's a known error
		if err == service.ErrOrderExpired {
			response.BadRequest(c, "订单已过期")
			return
		}
		response.ErrorFrom(c, err)
		return
	}

	// Get updated order to return status
	order, err := h.orderService.GetOrderByNo(c.Request.Context(), params.OutTradeNo)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"status":   order.Status,
		"order_no": order.OrderNo,
		"paid":     order.IsPaid(),
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
