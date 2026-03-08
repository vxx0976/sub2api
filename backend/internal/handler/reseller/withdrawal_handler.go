package reseller

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// WithdrawalHandler handles withdrawal-related endpoints for resellers
type WithdrawalHandler struct {
	commissionService *service.CommissionService
}

// NewWithdrawalHandler creates a new WithdrawalHandler
func NewWithdrawalHandler(commissionService *service.CommissionService) *WithdrawalHandler {
	return &WithdrawalHandler{commissionService: commissionService}
}

// List GET /api/v1/reseller/withdrawals
func (h *WithdrawalHandler) List(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
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
	status := c.Query("status")
	offset := (page - 1) * pageSize

	items, total, err := h.commissionService.ListWithdrawals(c.Request.Context(), resellerID, status, pageSize, offset)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, int64(total), page, pageSize)
}

// Create POST /api/v1/reseller/withdrawals
func (h *WithdrawalHandler) Create(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req struct {
		Amount         float64 `json:"amount" binding:"required,gt=0"`
		PaymentMethod  string  `json:"payment_method" binding:"required"`
		PaymentAccount string  `json:"payment_account" binding:"required"`
		PaymentName    string  `json:"payment_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	w, err := h.commissionService.CreateWithdrawal(c.Request.Context(), resellerID, req.Amount, req.PaymentMethod, req.PaymentAccount, req.PaymentName)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, w)
}

// Cancel DELETE /api/v1/reseller/withdrawals/:id
func (h *WithdrawalHandler) Cancel(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid withdrawal ID")
		return
	}
	if err := h.commissionService.CancelWithdrawal(c.Request.Context(), resellerID, id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Withdrawal cancelled"})
}
