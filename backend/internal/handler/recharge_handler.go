package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type RechargeHandler struct {
	rechargeOrderService *service.RechargeOrderService
	settingService       *service.SettingService
}

func NewRechargeHandler(
	rechargeOrderService *service.RechargeOrderService,
	settingService *service.SettingService,
) *RechargeHandler {
	return &RechargeHandler{
		rechargeOrderService: rechargeOrderService,
		settingService:       settingService,
	}
}

// 创建充值订单请求
type CreateRechargeOrderRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// 创建充值订单响应
type CreateRechargeOrderResponse struct {
	OrderNo      string  `json:"order_no"`
	PayURL       string  `json:"pay_url"`
	Amount       float64 `json:"amount"`
	CreditAmount float64 `json:"credit_amount"`
	Multiplier   float64 `json:"multiplier"`
}

// CreateRechargeOrder 创建充值订单
func (h *RechargeHandler) CreateRechargeOrder(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	var req CreateRechargeOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.rechargeOrderService.CreateRechargeOrder(
		c.Request.Context(),
		subject.UserID,
		req.Amount,
		getBaseURL(c),
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	resp := CreateRechargeOrderResponse{
		OrderNo: result.OrderNo,
		PayURL:  result.PayURL,
		Amount:  result.Amount,
	}
	if result.CreditAmount != nil {
		resp.CreditAmount = *result.CreditAmount
	}
	if result.Multiplier != nil {
		resp.Multiplier = *result.Multiplier
	}

	response.Success(c, resp)
}

// GetRechargeOrders 获取充值订单列表
func (h *RechargeHandler) GetRechargeOrders(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	orders, paginationResult, err := h.rechargeOrderService.GetRechargeOrders(
		c.Request.Context(),
		subject.UserID,
		params,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, orders, paginationResult.Total, page, pageSize)
}

// RepayRechargeOrder 继续支付充值订单
func (h *RechargeHandler) RepayRechargeOrder(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	orderNo := c.Param("order_no")
	if orderNo == "" {
		response.BadRequest(c, "order_no is required")
		return
	}

	result, err := h.rechargeOrderService.RepayRechargeOrder(
		c.Request.Context(),
		subject.UserID,
		orderNo,
		getBaseURL(c),
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	resp := CreateRechargeOrderResponse{
		OrderNo: result.OrderNo,
		PayURL:  result.PayURL,
		Amount:  result.Amount,
	}
	if result.CreditAmount != nil {
		resp.CreditAmount = *result.CreditAmount
	}
	if result.Multiplier != nil {
		resp.Multiplier = *result.Multiplier
	}

	response.Success(c, resp)
}

// GetRechargeConfig 获取充值配置（用户端）
func (h *RechargeHandler) GetRechargeConfig(c *gin.Context) {
	config, err := h.settingService.GetRechargeConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, config)
}

// GetRechargeConfigAdmin 获取充值配置（管理员）
func (h *RechargeHandler) GetRechargeConfigAdmin(c *gin.Context) {
	config, err := h.settingService.GetRechargeConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, config)
}

// UpdateRechargeConfig 更新充值配置（管理员）
func (h *RechargeHandler) UpdateRechargeConfig(c *gin.Context) {
	var config service.RechargeConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.settingService.UpdateRechargeConfig(c.Request.Context(), &config); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, config)
}

// GetAllRechargeOrders 获取所有充值订单（管理员）
func (h *RechargeHandler) GetAllRechargeOrders(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)

	// Get optional filters
	keyword := c.Query("keyword")
	status := c.Query("status")

	params := pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	orders, paginationResult, err := h.rechargeOrderService.GetAllRechargeOrders(
		c.Request.Context(),
		params,
		keyword,
		status,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, orders, paginationResult.Total, page, pageSize)
}
