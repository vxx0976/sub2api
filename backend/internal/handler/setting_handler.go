package handler

import (
	"encoding/json"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingHandler 公开设置处理器（无需认证）
type SettingHandler struct {
	settingService *service.SettingService
	version        string
}

// NewSettingHandler 创建公开设置处理器
func NewSettingHandler(settingService *service.SettingService, version string) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
		version:        version,
	}
}

// GetPublicSettings 获取公开设置
// GET /api/v1/settings/public
func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.settingService.GetPublicSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Convert service announcements to DTO announcements
	announcements := make([]dto.SimpleAnnouncement, 0, len(settings.Announcements))
	for _, a := range settings.Announcements {
		announcements = append(announcements, dto.SimpleAnnouncement{
			Title: a.Title,
			Date:  a.Date,
		})
	}

	result := dto.PublicSettings{
		RegistrationEnabled:         settings.RegistrationEnabled,
		EmailVerifyEnabled:          settings.EmailVerifyEnabled,
		PromoCodeEnabled:            settings.PromoCodeEnabled,
		PasswordResetEnabled:        settings.PasswordResetEnabled,
		InvitationCodeEnabled:       settings.InvitationCodeEnabled,
		TotpEnabled:                 settings.TotpEnabled,
		TurnstileEnabled:            settings.TurnstileEnabled,
		TurnstileSiteKey:            settings.TurnstileSiteKey,
		SiteName:                    settings.SiteName,
		SiteLogo:                    settings.SiteLogo,
		SiteSubtitle:                settings.SiteSubtitle,
		APIBaseURL:                  settings.APIBaseURL,
		ContactInfo:                 settings.ContactInfo,
		DocURL:                      settings.DocURL,
		HomeContent:                 settings.HomeContent,
		HideCcsImportButton:         settings.HideCcsImportButton,
		PurchaseSubscriptionEnabled: settings.PurchaseSubscriptionEnabled,
		PurchaseSubscriptionURL:     settings.PurchaseSubscriptionURL,
		LinuxDoOAuthEnabled:         settings.LinuxDoOAuthEnabled,
		Version:                     h.version,
		Announcements:               announcements,
		CryptoAddresses:             settings.CryptoAddresses,
		QueryDomain:                 settings.QueryDomain,
	}

	// If accessed via a reseller's custom domain, overlay reseller branding
	// This mirrors the logic in web/embed_on.go mergeResellerBranding
	if info := middleware.GetResellerDomainFromContext(c); info != nil {
		result.ResellerID = info.ResellerID
		result.ResellerDomain = info.Domain

		if info.SiteName != "" {
			result.SiteName = info.SiteName
		}
		if info.SiteLogo != "" {
			result.SiteLogo = info.SiteLogo
		}
		if info.BrandColor != "" {
			result.BrandColor = info.BrandColor
		}
		if info.CustomCSS != "" {
			result.CustomCSS = info.CustomCSS
		}
		if info.Subtitle != "" {
			result.SiteSubtitle = info.Subtitle
		}
		if info.HomeContent != "" {
			result.HomeContent = info.HomeContent
		}
		if info.HomeTemplate != "" {
			result.HomeTemplate = info.HomeTemplate
		}

		// Doc URL: use reseller's or clear system default
		if info.DocURL != "" {
			result.DocURL = info.DocURL
		} else {
			result.DocURL = ""
		}

		// Purchase: use reseller's or clear system defaults
		if info.PurchaseEnabled {
			result.PurchaseEnabled = true
			result.PurchaseURL = info.PurchaseURL
		}
		result.PurchaseSubscriptionEnabled = false
		result.PurchaseSubscriptionURL = ""

		if info.SEOTitle != "" {
			result.SEOTitle = info.SEOTitle
		}
		if info.SEODescription != "" {
			result.SEODescription = info.SEODescription
		}
		if info.SEOKeywords != "" {
			result.SEOKeywords = info.SEOKeywords
		}
		if info.LoginRedirect != "" {
			result.LoginRedirect = info.LoginRedirect
		}

		// Override from reseller-global settings
		if rs := info.ResellerSettings; rs != nil {
			if v := rs["contact_info"]; v != "" {
				result.ContactInfo = v
			}
			if v := rs["crypto_addresses"]; v != "" {
				result.CryptoAddresses = v
			}
			if v := rs["default_locale"]; v != "" {
				result.DefaultLocale = v
			}
			if v := rs["contact_wechat"]; v != "" {
				result.ContactWechat = v
			}
			if v := rs["contact_telegram"]; v != "" {
				result.ContactTelegram = v
			}
			// Replace system announcements with reseller's own
			if v := rs["announcements"]; v != "" {
				var arr []dto.SimpleAnnouncement
				if json.Unmarshal([]byte(v), &arr) == nil && len(arr) > 0 {
					result.Announcements = arr
				} else {
					result.Announcements = nil
				}
			} else {
				result.Announcements = nil
			}
		} else {
			result.Announcements = nil
		}

		// Domain-level default_locale overrides reseller-global
		if info.DefaultLocale != "" {
			result.DefaultLocale = info.DefaultLocale
		}
	}

	response.Success(c, result)
}
