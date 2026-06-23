package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// MergedOrderHandler 跨三个自建充值通道（EPAY 充值 / AliMPay / USDT）的合并订单列表接口。
// 返回按 created_at 倒序合并后的归一化订单，复用各通道既有 service 的查询能力。
type MergedOrderHandler struct {
	mergedService *service.MergedOrderService
}

// NewMergedOrderHandler creates a new MergedOrderHandler.
func NewMergedOrderHandler(mergedService *service.MergedOrderService) *MergedOrderHandler {
	return &MergedOrderHandler{mergedService: mergedService}
}

// 合法的 channel / status 取值（空串表示全部）。
var (
	mergedValidChannels = map[string]bool{"": true, "recharge": true, "alimpay": true, "usdt": true, "manual": true}
	mergedValidStatuses = map[string]bool{"": true, "pending": true, "paid": true, "expired": true, "refunded": true}
)

func parseMergedPaging(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func parseMergedFilters(c *gin.Context) (channel, status string, ok bool) {
	channel = c.Query("channel")
	status = c.Query("status")
	if !mergedValidChannels[channel] || !mergedValidStatuses[status] {
		return "", "", false
	}
	return channel, status, true
}

// ListUserOrders GET /api/v1/topup/orders
// 返回当前登录用户在三个自建充值通道的订单，按 created_at 倒序合并分页。
func (h *MergedOrderHandler) ListUserOrders(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := parseMergedPaging(c)
	channel, status, valid := parseMergedFilters(c)
	if !valid {
		response.BadRequest(c, "invalid channel or status")
		return
	}

	uid := subject.UserID
	filter := service.MergedOrderFilter{
		Channel: channel,
		Status:  status,
		UserID:  &uid,
	}

	items, total, err := h.mergedService.ListMerged(c.Request.Context(), filter, page, pageSize, false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

// AdminListOrders GET /api/v1/admin/topup/orders
// 返回所有用户在三个自建充值通道的订单（含 user_email），支持 user_id 过滤。
func (h *MergedOrderHandler) AdminListOrders(c *gin.Context) {
	page, pageSize := parseMergedPaging(c)
	channel, status, valid := parseMergedFilters(c)
	if !valid {
		response.BadRequest(c, "invalid channel or status")
		return
	}

	var userID *int64
	if uid := c.Query("user_id"); uid != "" {
		if id, err := strconv.ParseInt(uid, 10, 64); err == nil {
			userID = &id
		}
	}

	filter := service.MergedOrderFilter{
		Channel: channel,
		Status:  status,
		UserID:  userID,
	}

	items, total, err := h.mergedService.ListMerged(c.Request.Context(), filter, page, pageSize, true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}
