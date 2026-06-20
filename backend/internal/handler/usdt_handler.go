package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// UsdtHandler USDT(TRC20) 自建收款充值接口。
// 与 OrderHandler(AliMPay) 平级：CNY 计价入账，差异是链上收款通道。
type UsdtHandler struct {
	usdtService  *service.UsdtOrderService
	adminService service.AdminService
}

func NewUsdtHandler(usdtService *service.UsdtOrderService, adminService service.AdminService) *UsdtHandler {
	return &UsdtHandler{usdtService: usdtService, adminService: adminService}
}

func fmtUsdt(amount float64) string { return strconv.FormatFloat(amount, 'f', 6, 64) }

// GetConfig GET /api/v1/usdt/config
func (h *UsdtHandler) GetConfig(c *gin.Context) {
	cfg, err := h.usdtService.GetConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// CreateOrder POST /api/v1/usdt/create
func (h *UsdtHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
		Chain  string  `json:"chain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	sourceDomain := c.GetHeader("X-Forwarded-Host")
	if sourceDomain == "" {
		sourceDomain = c.Request.Host
	}

	result, err := h.usdtService.CreateOrder(c.Request.Context(), subject.UserID, req.Amount, req.Chain, sourceDomain)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"order_no":        result.Order.OrderNo,
		"amount":          result.Order.Amount,
		"credit_amount":   result.Order.CreditAmount,
		"chain":           result.Chain,
		"address":         result.Address,
		"usdt_amount":     result.UsdtAmount,
		"usdt_amount_str": result.UsdtAmountStr,
		"rate":            result.Rate,
		"expires_in":      result.ExpiresIn,
		"expired_at":      result.Order.ExpiredAt,
	})
}

// GetOrderStatus GET /api/v1/usdt/status/:order_no
func (h *UsdtHandler) GetOrderStatus(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	orderNo := c.Param("order_no")
	o, err := h.usdtService.GetOrderStatus(c.Request.Context(), orderNo, subject.UserID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"order_no":        o.OrderNo,
		"amount":          o.Amount,
		"credit_amount":   o.CreditAmount,
		"chain":           o.Chain,
		"address":         o.ReceivingAddress,
		"usdt_amount":     o.UsdtAmount,
		"usdt_amount_str": fmtUsdt(o.UsdtAmount),
		"status":          o.Status,
		"paid_at":         o.PaidAt,
		"trade_no":        o.TradeNo,
		"expired_at":      o.ExpiredAt,
	})
}

// ListOrders GET /api/v1/usdt/orders
func (h *UsdtHandler) ListOrders(c *gin.Context) {
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

	orders, total, err := h.usdtService.ListUserOrders(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	items := make([]gin.H, len(orders))
	for i, o := range orders {
		items[i] = usdtOrderToMap(o, false)
	}
	response.Paginated(c, items, int64(total), page, pageSize)
}

// AdminGetConfig GET /api/v1/admin/usdt/config
func (h *UsdtHandler) AdminGetConfig(c *gin.Context) {
	cfg, err := h.usdtService.GetAdminUsdtConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// AdminUpdateConfig PUT /api/v1/admin/usdt/config
func (h *UsdtHandler) AdminUpdateConfig(c *gin.Context) {
	var req service.AdminUsdtConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.usdtService.UpdateAdminUsdtConfig(c.Request.Context(), &req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	cfg, err := h.usdtService.GetAdminUsdtConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// AdminListOrders GET /api/v1/admin/usdt/orders
func (h *UsdtHandler) AdminListOrders(c *gin.Context) {
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

	orders, total, err := h.usdtService.ListAllOrders(c.Request.Context(), status, userID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	emailMap := make(map[int64]string)
	for _, o := range orders {
		if _, seen := emailMap[o.UserID]; seen {
			continue
		}
		if u, err := h.adminService.GetUser(c.Request.Context(), o.UserID); err == nil && u != nil {
			emailMap[o.UserID] = u.Email
		}
	}

	items := make([]gin.H, len(orders))
	for i, o := range orders {
		m := usdtOrderToMap(o, true)
		m["user_email"] = emailMap[o.UserID]
		items[i] = m
	}
	response.Paginated(c, items, int64(total), page, pageSize)
}

// AdminRefundOrder POST /api/v1/admin/usdt/orders/:orderNo/refund
func (h *UsdtHandler) AdminRefundOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		response.Error(c, http.StatusBadRequest, "order_no is required")
		return
	}
	var req AdminRefundOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	order, err := h.usdtService.RefundOrder(c.Request.Context(), orderNo, req.Reason)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"order_no":      order.OrderNo,
		"user_id":       order.UserID,
		"credit_amount": order.CreditAmount,
		"status":        order.Status,
	})
}

func usdtOrderToMap(o *service.UsdtOrder, admin bool) gin.H {
	m := gin.H{
		"id":              o.ID,
		"order_no":        o.OrderNo,
		"trade_no":        o.TradeNo,
		"user_id":         o.UserID,
		"amount":          o.Amount,
		"credit_amount":   o.CreditAmount,
		"chain":           o.Chain,
		"usdt_amount":     o.UsdtAmount,
		"usdt_amount_str": fmtUsdt(o.UsdtAmount),
		"usdt_rate":       o.UsdtRate,
		"status":          o.Status,
		"pay_type":        o.PayType,
		"paid_at":         o.PaidAt,
		"created_at":      o.CreatedAt,
		"updated_at":      o.UpdatedAt,
		"expired_at":      o.ExpiredAt,
	}
	if o.PaidUsdtAmount != nil {
		m["paid_usdt_amount"] = *o.PaidUsdtAmount
		m["paid_usdt_amount_str"] = fmt.Sprintf("%.6f", *o.PaidUsdtAmount)
	}
	if admin {
		m["receiving_address"] = o.ReceivingAddress
		m["from_address"] = o.FromAddress
		m["source_domain"] = o.SourceDomain
	}
	return m
}
