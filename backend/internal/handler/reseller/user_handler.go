package reseller

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler handles reseller user list endpoints
type UserHandler struct {
	resellerService *service.ResellerService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(resellerService *service.ResellerService) *UserHandler {
	return &UserHandler{resellerService: resellerService}
}

// List returns users belonging to the reseller (parent_id = resellerID)
// GET /api/v1/reseller/users
func (h *UserHandler) List(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)

	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 100 {
		search = search[:100]
	}

	filters := service.UserListFilters{
		ParentID: &resellerID,
		Search:   search,
	}

	users, pag, err := h.resellerService.ListUsers(c.Request.Context(), resellerID, page, pageSize, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]*dto.User, 0, len(users))
	for i := range users {
		out = append(out, dto.UserFromServiceShallow(&users[i]))
	}

	response.PaginatedWithResult(c, out, &response.PaginationResult{
		Total:    pag.Total,
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Pages:    pag.Pages,
	})
}

// UpdateBalanceRequest represents reseller balance transfer request
type UpdateBalanceRequest struct {
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Operation string  `json:"operation" binding:"required,oneof=add subtract"`
	Notes     string  `json:"notes"`
}

// UpdateBalance handles reseller transferring balance to/from a user
// POST /api/v1/reseller/users/:id/balance
func (h *UserHandler) UpdateBalance(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req UpdateBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	user, err := h.resellerService.TransferBalance(c.Request.Context(), resellerID, userID, req.Amount, req.Operation, req.Notes)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.UserFromServiceShallow(user))
}

// BalanceHistory returns balance history for a user belonging to the reseller
// GET /api/v1/reseller/users/:id/balance-history
func (h *UserHandler) BalanceHistory(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	page, pageSize := response.ParsePagination(c)

	codes, total, totalRecharged, err := h.resellerService.GetUserBalanceHistory(c.Request.Context(), resellerID, userID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.AdminRedeemCode, 0, len(codes))
	for i := range codes {
		out = append(out, *dto.RedeemCodeFromServiceAdmin(&codes[i]))
	}

	pages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if pages < 1 {
		pages = 1
	}
	response.Success(c, gin.H{
		"items":           out,
		"total":           total,
		"page":            page,
		"page_size":       pageSize,
		"pages":           pages,
		"total_recharged": totalRecharged,
	})
}
