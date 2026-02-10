package reseller

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingHandler handles reseller settings management
type SettingHandler struct {
	resellerService *service.ResellerService
	botManager      *service.TelegramBotManager
}

// NewSettingHandler creates a new SettingHandler
func NewSettingHandler(resellerService *service.ResellerService, botManager *service.TelegramBotManager) *SettingHandler {
	return &SettingHandler{resellerService: resellerService, botManager: botManager}
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

	// Filter out protected keys
	for key := range settings {
		if service.IsProtectedResellerSetting(key) {
			delete(settings, key)
		}
	}

	if err := h.resellerService.UpdateSettings(c.Request.Context(), resellerID, settings); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Notify bot manager about settings change
	if h.botManager != nil {
		allSettings, err := h.resellerService.GetSettings(c.Request.Context(), resellerID)
		if err == nil {
			h.botManager.OnSettingsUpdated(resellerID, allSettings)
		}
	}

	response.Success(c, gin.H{"message": "Settings updated successfully"})
}

// GenerateBindCode generates a temporary bind code for Telegram bot binding.
// POST /api/v1/reseller/settings/tg-bind-code
func (h *SettingHandler) GenerateBindCode(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)

	// Generate 6-byte random hex code
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		response.InternalError(c, "Failed to generate bind code")
		return
	}
	code := hex.EncodeToString(b)

	if err := h.resellerService.UpdateSettings(c.Request.Context(), resellerID, map[string]string{
		service.ResellerSettingTgBindCode: code,
	}); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"bind_code": code})
}
