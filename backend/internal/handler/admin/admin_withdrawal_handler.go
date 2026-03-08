package admin

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// AdminWithdrawalHandler handles admin withdrawal management endpoints
type AdminWithdrawalHandler struct {
	commissionService *service.CommissionService
}

// NewAdminWithdrawalHandler creates a new AdminWithdrawalHandler
func NewAdminWithdrawalHandler(commissionService *service.CommissionService) *AdminWithdrawalHandler {
	return &AdminWithdrawalHandler{commissionService: commissionService}
}

// List GET /api/v1/admin/withdrawals
func (h *AdminWithdrawalHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	status := c.Query("status")
	var resellerID int64
	if rid := c.Query("reseller_id"); rid != "" {
		resellerID, _ = strconv.ParseInt(rid, 10, 64)
	}
	offset := (page - 1) * pageSize

	items, total, err := h.commissionService.AdminListWithdrawals(c.Request.Context(), status, resellerID, pageSize, offset)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, int64(total), page, pageSize)
}

// Pay PUT /api/v1/admin/withdrawals/:id/pay
func (h *AdminWithdrawalHandler) Pay(c *gin.Context) {
	adminID := getAdminIDFromContext(c)
	if adminID == 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid withdrawal ID")
		return
	}
	var req struct {
		AdminNotes string `json:"admin_notes"`
	}
	_ = c.ShouldBindJSON(&req)

	if err := h.commissionService.AdminPayWithdrawal(c.Request.Context(), adminID, id, req.AdminNotes); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	logger.LegacyPrintf("handler.admin.withdrawal", "[Withdrawal] admin=%d marked withdrawal=%d as paid notes=%q", adminID, id, req.AdminNotes)
	response.Success(c, gin.H{"message": "Withdrawal marked as paid"})
}

// Reject PUT /api/v1/admin/withdrawals/:id/reject
func (h *AdminWithdrawalHandler) Reject(c *gin.Context) {
	adminID := getAdminIDFromContext(c)
	if adminID == 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid withdrawal ID")
		return
	}
	var req struct {
		AdminNotes string `json:"admin_notes"`
	}
	_ = c.ShouldBindJSON(&req)

	if err := h.commissionService.AdminRejectWithdrawal(c.Request.Context(), adminID, id, req.AdminNotes); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	logger.LegacyPrintf("handler.admin.withdrawal", "[Withdrawal] admin=%d rejected withdrawal=%d notes=%q", adminID, id, req.AdminNotes)
	response.Success(c, gin.H{"message": "Withdrawal rejected"})
}
