package handler

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// RechargeHandler 充值相关接口
type RechargeHandler struct {
	rechargeService *service.RechargeService
	adminService    service.AdminService
}

// NewRechargeHandler creates a new RechargeHandler
func NewRechargeHandler(rechargeService *service.RechargeService, adminService service.AdminService) *RechargeHandler {
	return &RechargeHandler{rechargeService: rechargeService, adminService: adminService}
}

// GetConfig GET /api/v1/recharge/config
func (h *RechargeHandler) GetConfig(c *gin.Context) {
	cfg, err := h.rechargeService.GetConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// CreateOrder POST /api/v1/recharge/create
func (h *RechargeHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Amount  float64 `json:"amount" binding:"required,gt=0"`
		PayType string  `json:"pay_type" binding:"required,oneof=alipay wxpay"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 构造 baseURL（用于回调地址）
	scheme := "https"
	if c.Request.TLS == nil {
		if fwd := c.GetHeader("X-Forwarded-Proto"); fwd != "" {
			scheme = fwd
		} else {
			scheme = "http"
		}
	}
	baseURL := scheme + "://" + c.Request.Host

	clientIP := c.ClientIP()
	result, err := h.rechargeService.CreateOrder(c.Request.Context(), subject.UserID, req.Amount, req.PayType, baseURL, clientIP)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"order_no":      result.Order.OrderNo,
		"amount":        result.Order.Amount,
		"credit_amount": result.Order.CreditAmount,
		"multiplier":    result.Order.Multiplier,
		"pay_url":       result.PayURL,
		"qrcode":        result.QRCode,
		"expired_at":    result.Order.ExpiredAt,
	})
}

// GetOrderStatus GET /api/v1/recharge/status/:order_no
func (h *RechargeHandler) GetOrderStatus(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	orderNo := c.Param("order_no")
	order, err := h.rechargeService.GetOrderStatus(c.Request.Context(), orderNo, subject.UserID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"order_no":      order.OrderNo,
		"amount":        order.Amount,
		"credit_amount": order.CreditAmount,
		"status":        order.Status,
		"paid_at":       order.PaidAt,
	})
}

// ListOrders GET /api/v1/recharge/orders
func (h *RechargeHandler) ListOrders(c *gin.Context) {
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

	orders, total, err := h.rechargeService.ListUserOrders(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 显式构造响应，避免 sonic 省略零值字段
	items := make([]gin.H, len(orders))
	for i, o := range orders {
		items[i] = gin.H{
			"id":            o.ID,
			"order_no":      o.OrderNo,
			"trade_no":      o.TradeNo,
			"user_id":       o.UserID,
			"amount":        o.Amount,
			"credit_amount": o.CreditAmount,
			"multiplier":    o.Multiplier,
			"status":        o.Status,
			"pay_type":      o.PayType,
			"paid_at":       o.PaidAt,
			"created_at":    o.CreatedAt,
			"updated_at":    o.UpdatedAt,
			"expired_at":    o.ExpiredAt,
		}
	}
	response.Paginated(c, items, int64(total), page, pageSize)
}

// HandleNotify GET /api/v1/recharge/notify — 异步回调（注意：彩虹易支付回调是 GET）
func (h *RechargeHandler) HandleNotify(c *gin.Context) {
	params := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	if err := h.rechargeService.HandleNotify(c.Request.Context(), params); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}

// HandleReturn GET /api/v1/recharge/return — 同步跳转（重定向回前端充值页）
func (h *RechargeHandler) HandleReturn(c *gin.Context) {
	c.Redirect(http.StatusFound, "/#/recharge?from=payment")
}

// AdminListOrders GET /api/v1/admin/recharge/orders
func (h *RechargeHandler) AdminListOrders(c *gin.Context) {
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

	orders, total, err := h.rechargeService.ListAllOrders(c.Request.Context(), status, userID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 批量查用户邮箱
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

	// 显式构造响应，避免 sonic JSON 编码器省略零值字段
	items := make([]gin.H, len(orders))
	for i, o := range orders {
		items[i] = gin.H{
			"id":            o.ID,
			"order_no":      o.OrderNo,
			"trade_no":      o.TradeNo,
			"user_id":       o.UserID,
			"user_email":    emailMap[o.UserID],
			"amount":        o.Amount,
			"credit_amount": o.CreditAmount,
			"multiplier":    o.Multiplier,
			"status":        o.Status,
			"pay_type":      o.PayType,
			"paid_at":       o.PaidAt,
			"created_at":    o.CreatedAt,
			"updated_at":    o.UpdatedAt,
			"expired_at":    o.ExpiredAt,
		}
	}

	response.Paginated(c, items, int64(total), page, pageSize)
}
