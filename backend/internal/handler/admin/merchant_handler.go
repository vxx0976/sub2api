package admin

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// MerchantHandler handles admin merchant management endpoints
type MerchantHandler struct {
	commissionService *service.CommissionService
	adminService      service.AdminService
}

// NewMerchantHandler creates a new MerchantHandler
func NewMerchantHandler(commissionService *service.CommissionService, adminService service.AdminService) *MerchantHandler {
	return &MerchantHandler{commissionService: commissionService, adminService: adminService}
}

// List GET /api/v1/admin/merchants
func (h *MerchantHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	search := c.Query("search")

	items, total, err := h.commissionService.AdminGetMerchants(c.Request.Context(), page, pageSize, search)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, int64(total), page, pageSize)
}

// GetSettings GET /api/v1/admin/merchants/:id/settings
func (h *MerchantHandler) GetSettings(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid merchant ID")
		return
	}
	settings, err := h.commissionService.AdminGetMerchantSettings(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}

// UpdateSettings PUT /api/v1/admin/merchants/:id/settings
func (h *MerchantHandler) UpdateSettings(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid merchant ID")
		return
	}
	// Accept both string and number values from frontend
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	settings := make(map[string]string, len(raw))
	for k, v := range raw {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			settings[k] = s
		} else {
			// number or other scalar → convert to string
			settings[k] = strings.TrimSpace(string(v))
		}
	}
	if err := h.commissionService.AdminUpdateMerchantSettings(c.Request.Context(), id, settings); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Settings updated"})
}

// UpdateBalance handles updating merchant balance
// POST /api/v1/admin/merchants/:id/balance
func (h *MerchantHandler) UpdateBalance(c *gin.Context) {
	merchantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid merchant ID")
		return
	}

	var req UpdateBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	idempotencyPayload := struct {
		MerchantID int64                `json:"merchant_id"`
		Body       UpdateBalanceRequest `json:"body"`
	}{
		MerchantID: merchantID,
		Body:       req,
	}
	executeAdminIdempotentJSON(c, "admin.merchants.balance.update", idempotencyPayload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		user, execErr := h.adminService.UpdateUserBalance(ctx, merchantID, req.Balance, req.Operation, req.Notes)
		if execErr != nil {
			return nil, execErr
		}
		return dto.UserFromServiceAdmin(user), nil
	})
}
