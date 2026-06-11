package reseller

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// CommissionHandler handles commission-related endpoints for resellers
type CommissionHandler struct {
	commissionService *service.CommissionService
}

// NewCommissionHandler creates a new CommissionHandler
func NewCommissionHandler(commissionService *service.CommissionService) *CommissionHandler {
	return &CommissionHandler{commissionService: commissionService}
}

// GetSummary GET /api/v1/reseller/commissions/summary
func (h *CommissionHandler) GetSummary(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	summary, err := h.commissionService.GetSummary(c.Request.Context(), resellerID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, summary)
}

func parsePage(c *gin.Context) (page, pageSize, offset int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset = (page - 1) * pageSize
	return
}

// GetRecharges GET /api/v1/reseller/commissions/recharges
func (h *CommissionHandler) GetRecharges(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	page, pageSize, offset := parsePage(c)
	items, total, err := h.commissionService.ListRechargeDetail(c.Request.Context(), resellerID, pageSize, offset)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, int64(total), page, pageSize)
}
