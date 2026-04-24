package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// OrderHandler AliMPay 个人免签充值接口
// 和 RechargeHandler 平级：业务都是"CNY 入账 → USD 余额"，差异是支付通道
type OrderHandler struct {
	orderService *service.OrderService
	adminService service.AdminService
}

func NewOrderHandler(orderService *service.OrderService, adminService service.AdminService) *OrderHandler {
	return &OrderHandler{orderService: orderService, adminService: adminService}
}

// GetConfig GET /api/v1/alimpay/config
func (h *OrderHandler) GetConfig(c *gin.Context) {
	cfg, err := h.orderService.GetConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// CreateOrder POST /api/v1/alimpay/create
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 来源域名：优先取 X-Forwarded-Host（反代场景），否则 Request.Host
	sourceDomain := c.GetHeader("X-Forwarded-Host")
	if sourceDomain == "" {
		sourceDomain = c.Request.Host
	}

	result, err := h.orderService.CreateOrder(c.Request.Context(), subject.UserID, req.Amount, sourceDomain)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"order_no":       result.Order.OrderNo,
		"amount":         result.Order.Amount,
		"payment_amount": result.Order.PaymentAmount,
		"credit_amount":  result.Order.CreditAmount,
		"multiplier":     result.Order.Multiplier,
		"qrcode_url":     result.QRCodeURL,
		"mode":           result.Mode,
		"expires_in":     result.ExpiresIn,
		"expired_at":     result.Order.ExpiredAt,
	})
}

// GetOrderStatus GET /api/v1/alimpay/status/:order_no
func (h *OrderHandler) GetOrderStatus(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	orderNo := c.Param("order_no")
	o, err := h.orderService.GetOrderStatus(c.Request.Context(), orderNo, subject.UserID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"order_no":       o.OrderNo,
		"amount":         o.Amount,
		"payment_amount": o.PaymentAmount,
		"credit_amount":  o.CreditAmount,
		"status":         o.Status,
		"paid_at":        o.PaidAt,
	})
}

// ListOrders GET /api/v1/alimpay/orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orders, total, err := h.orderService.ListUserOrders(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	items := make([]gin.H, len(orders))
	for i, o := range orders {
		items[i] = gin.H{
			"id":             o.ID,
			"order_no":       o.OrderNo,
			"trade_no":       o.TradeNo,
			"user_id":        o.UserID,
			"amount":         o.Amount,
			"payment_amount": o.PaymentAmount,
			"credit_amount":  o.CreditAmount,
			"multiplier":     o.Multiplier,
			"status":         o.Status,
			"pay_type":       o.PayType,
			"paid_at":        o.PaidAt,
			"created_at":     o.CreatedAt,
			"updated_at":     o.UpdatedAt,
			"expired_at":     o.ExpiredAt,
		}
	}
	response.Paginated(c, items, int64(total), page, pageSize)
}

// AdminGetConfig GET /api/v1/admin/alimpay/config
func (h *OrderHandler) AdminGetConfig(c *gin.Context) {
	cfg, err := h.orderService.GetAdminAliMPayConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// AdminUpdateConfig PUT /api/v1/admin/alimpay/config
func (h *OrderHandler) AdminUpdateConfig(c *gin.Context) {
	var req service.AdminAliMPayConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.orderService.UpdateAdminAliMPayConfig(c.Request.Context(), &req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	cfg, err := h.orderService.GetAdminAliMPayConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// AdminListOrders GET /api/v1/admin/alimpay/orders
func (h *OrderHandler) AdminListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	status := c.Query("status")
	var userID *int64
	if uid := c.Query("user_id"); uid != "" {
		if id, err := strconv.ParseInt(uid, 10, 64); err == nil {
			userID = &id
		}
	}

	orders, total, err := h.orderService.ListAllOrders(c.Request.Context(), status, userID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	uidSet := make(map[int64]bool)
	for _, o := range orders {
		uidSet[o.UserID] = true
	}
	emailMap := make(map[int64]string)
	for uid := range uidSet {
		if u, err := h.adminService.GetUser(c.Request.Context(), uid); err == nil && u != nil {
			emailMap[uid] = u.Email
		}
	}

	items := make([]gin.H, len(orders))
	for i, o := range orders {
		items[i] = gin.H{
			"id":             o.ID,
			"order_no":       o.OrderNo,
			"trade_no":       o.TradeNo,
			"user_id":        o.UserID,
			"user_email":     emailMap[o.UserID],
			"amount":         o.Amount,
			"payment_amount": o.PaymentAmount,
			"credit_amount":  o.CreditAmount,
			"multiplier":     o.Multiplier,
			"status":         o.Status,
			"pay_type":       o.PayType,
			"paid_at":        o.PaidAt,
			"source_domain":  o.SourceDomain,
			"created_at":     o.CreatedAt,
			"updated_at":     o.UpdatedAt,
			"expired_at":     o.ExpiredAt,
		}
	}
	response.Paginated(c, items, int64(total), page, pageSize)
}
