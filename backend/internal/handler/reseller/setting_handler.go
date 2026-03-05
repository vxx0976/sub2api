package reseller

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingHandler handles reseller settings management
type SettingHandler struct {
	resellerService      *service.ResellerService
	invalidateDomainFunc func(domain string)
}

// NewSettingHandler creates a new SettingHandler
func NewSettingHandler(resellerService *service.ResellerService) *SettingHandler {
	return &SettingHandler{resellerService: resellerService}
}

// SetCacheInvalidator sets the function to call when reseller settings change
func (h *SettingHandler) SetCacheInvalidator(fn func(domain string)) {
	h.invalidateDomainFunc = fn
}

// invalidateAllDomainCaches invalidates the HTML cache for all domains owned by this reseller
func (h *SettingHandler) invalidateAllDomainCaches(c *gin.Context, resellerID int64) {
	if h.invalidateDomainFunc == nil {
		return
	}
	domains, _, err := h.resellerService.ListDomains(c.Request.Context(), resellerID, 1, 100)
	if err != nil {
		return
	}
	for _, d := range domains {
		h.invalidateDomainFunc(d.Domain)
	}
}

// Get returns all settings for the reseller
// GET /api/v1/reseller/settings
func (h *SettingHandler) Get(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)

	settings, err := h.resellerService.GetSettings(c.Request.Context(), resellerID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, settings)
}

// Update updates settings for the reseller
// PUT /api/v1/reseller/settings
func (h *SettingHandler) Update(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)

	var settings map[string]string
	if err := c.ShouldBindJSON(&settings); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.resellerService.UpdateSettings(c.Request.Context(), resellerID, settings); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Invalidate HTML cache for all reseller domains so new settings take effect
	h.invalidateAllDomainCaches(c, resellerID)

	response.Success(c, gin.H{"message": "Settings updated successfully"})
}
