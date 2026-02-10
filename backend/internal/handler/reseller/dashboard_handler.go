package reseller

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// DashboardHandler handles reseller dashboard endpoints
type DashboardHandler struct {
	resellerService *service.ResellerService
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(resellerService *service.ResellerService) *DashboardHandler {
	return &DashboardHandler{resellerService: resellerService}
}

// GetStats returns reseller dashboard statistics
// GET /api/v1/reseller/dashboard
func (h *DashboardHandler) GetStats(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Error(c, 401, "Unauthorized")
		return
	}

	stats, err := h.resellerService.GetDashboardStats(c.Request.Context(), resellerID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, stats)
}

// getResellerIDFromContext extracts the reseller user ID from gin context
func getResellerIDFromContext(c *gin.Context) int64 {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		return 0
	}
	return subject.UserID
}
