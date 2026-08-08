package handler

import (
	"encoding/json"
	"html"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingHandler 公开设置处理器（无需认证）
type SettingHandler struct {
	settingService           *service.SettingService
	notificationEmailService *service.NotificationEmailService
	version                  string
}

// NewSettingHandler 创建公开设置处理器
func NewSettingHandler(settingService *service.SettingService, version string) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
		version:        version,
	}
}

// SetNotificationEmailService attaches the public notification email service without
// changing the constructor signature used by existing tests.
func (h *SettingHandler) SetNotificationEmailService(notificationEmailService *service.NotificationEmailService) {
	h.notificationEmailService = notificationEmailService
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
			Title:   a.Title,
			Content: a.Content,
			Date:    a.Date,
		})
	}

	result := dto.PublicSettings{
		RegistrationEnabled:              settings.RegistrationEnabled,
		EmailVerifyEnabled:               settings.EmailVerifyEnabled,
		ForceEmailOnThirdPartySignup:     settings.ForceEmailOnThirdPartySignup,
		RegistrationEmailSuffixWhitelist: settings.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                 settings.PromoCodeEnabled,
		PasswordResetEnabled:             settings.PasswordResetEnabled,
		InvitationCodeEnabled:            settings.InvitationCodeEnabled,
		TotpEnabled:                      settings.TotpEnabled,
		PasskeyEnabled:                   settings.PasskeyEnabled,
		LoginAgreementEnabled:            settings.LoginAgreementEnabled,
		LoginAgreementMode:               settings.LoginAgreementMode,
		LoginAgreementUpdatedAt:          settings.LoginAgreementUpdatedAt,
		LoginAgreementRevision:           settings.LoginAgreementRevision,
		LoginAgreementDocuments:          publicLoginAgreementDocumentsToDTO(settings.LoginAgreementDocuments),
		TurnstileEnabled:                 settings.TurnstileEnabled,
		TurnstileSiteKey:                 settings.TurnstileSiteKey,
		TencentCaptchaEnabled:            settings.TencentCaptchaEnabled,
		TencentCaptchaAppID:              settings.TencentCaptchaAppID,
		TencentCaptchaRegion:             settings.TencentCaptchaRegion,
		AliyunCaptchaEnabled:             settings.AliyunCaptchaEnabled,
		AliyunCaptchaSceneID:             settings.AliyunCaptchaSceneID,
		AliyunCaptchaPrefix:              settings.AliyunCaptchaPrefix,
		AliyunCaptchaRegion:              settings.AliyunCaptchaRegion,
		SiteName:                         settings.SiteName,
		SiteLogo:                         settings.SiteLogo,
		SiteSubtitle:                     settings.SiteSubtitle,
		APIBaseURL:                       settings.APIBaseURL,
		ContactInfo:                      settings.ContactInfo,
		DocURL:                           settings.DocURL,
		HomeContent:                      settings.HomeContent,
		CompactHomeEnabled:               settings.CompactHomeEnabled,
		HideCcsImportButton:              settings.HideCcsImportButton,
		TableDefaultPageSize:             settings.TableDefaultPageSize,
		TablePageSizeOptions:             settings.TablePageSizeOptions,
		CustomMenuItems:                  dto.ParseUserVisibleMenuItems(settings.CustomMenuItems),
		CustomEndpoints:                  dto.ParseCustomEndpoints(settings.CustomEndpoints),
		RechargeEnabled:                  settings.RechargeEnabled,
		AliMPayEnabled:                   settings.AliMPayEnabled,
		UsdtEnabled:                      settings.UsdtEnabled,
		DingTalkOAuthEnabled:             settings.DingTalkOAuthEnabled,
		LinuxDoOAuthEnabled:              settings.LinuxDoOAuthEnabled,
		WeChatOAuthEnabled:               settings.WeChatOAuthEnabled,
		WeChatOAuthOpenEnabled:           settings.WeChatOAuthOpenEnabled,
		WeChatOAuthMPEnabled:             settings.WeChatOAuthMPEnabled,
		WeChatOAuthMobileEnabled:         settings.WeChatOAuthMobileEnabled,
		OIDCOAuthEnabled:                 settings.OIDCOAuthEnabled,
		OIDCOAuthProviderName:            settings.OIDCOAuthProviderName,
		GitHubOAuthEnabled:               settings.GitHubOAuthEnabled,
		GoogleOAuthEnabled:               settings.GoogleOAuthEnabled,
		BackendModeEnabled:               settings.BackendModeEnabled,
		PaymentEnabled:                   settings.PaymentEnabled,
		Version:                          h.version,
		Announcements:                    announcements,
		ContactWechat:                    settings.ContactWechat,
		ContactTelegram:                  settings.ContactTelegram,
		ContactQQ:                        settings.ContactQQ,
		ServerTimezone:                   timezone.Name(),
		ServerUTCOffset:                  timezone.UTCOffset(),
		BalanceLowNotifyEnabled:          settings.BalanceLowNotifyEnabled,
		AccountQuotaNotifyEnabled:        settings.AccountQuotaNotifyEnabled,
		BalanceLowNotifyThreshold:        settings.BalanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:      settings.BalanceLowNotifyRechargeURL,

		ChannelMonitorEnabled:                settings.ChannelMonitorEnabled,
		ChannelMonitorDefaultIntervalSeconds: settings.ChannelMonitorDefaultIntervalSeconds,

		AvailableChannelsEnabled: settings.AvailableChannelsEnabled,

		ModelPlazaEnabled:     settings.ModelPlazaEnabled,
		ModelPlazaRequireAuth: settings.ModelPlazaRequireAuth,

		AffiliateEnabled: settings.AffiliateEnabled,

		RiskControlEnabled: settings.RiskControlEnabled,

		AllowUserViewErrorRequests: settings.AllowUserViewErrorRequests, // 来自 main
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

		// Purchase: use reseller's setting only (no system-level fallback)
		if info.PurchaseEnabled {
			result.PurchaseEnabled = true
			result.PurchaseURL = info.PurchaseURL
		}

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

		// Reseller sites are isolated from main site contact info
		result.ContactInfo = ""
		result.ContactWechat = ""
		result.ContactTelegram = ""
		result.ContactQQ = ""

		// Override from reseller-global settings
		if rs := info.ResellerSettings; rs != nil {
			// Expose merchant_mode to regular users so sidebar can show subscription menus
			if rs["merchant_mode"] == "enabled" {
				result.ResellerAgentEnabled = true
				// When merchant mode is enabled, expose the payment page if reseller has its own pay_url.
				if payURL := rs["pay_url"]; payURL != "" {
					result.PurchaseEnabled = true
					result.PurchaseURL = payURL
				}
			}
			if v := rs["contact_info"]; v != "" {
				result.ContactInfo = v
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
			if v := rs["contact_qq"]; v != "" {
				result.ContactQQ = v
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

// UnsubscribeNotificationEmail handles optional notification email opt-outs.
// GET /api/v1/settings/email-unsubscribe?token=...
func (h *SettingHandler) UnsubscribeNotificationEmail(c *gin.Context) {
	if h.notificationEmailService == nil {
		response.InternalError(c, "notification email service is not configured")
		return
	}
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		response.BadRequest(c, "token is required")
		return
	}
	result, err := h.notificationEmailService.Unsubscribe(c.Request.Context(), token)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	body := "<!doctype html><html><head><meta charset=\"utf-8\"><title>Unsubscribed</title></head><body style=\"font-family:-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif;padding:32px;\"><h1>Unsubscribed</h1><p>You have unsubscribed <strong>" + html.EscapeString(result.Email) + "</strong> from <strong>" + html.EscapeString(result.Event) + "</strong> emails.</p></body></html>"
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(body))
}

func publicLoginAgreementDocumentsToDTO(items []service.LoginAgreementDocument) []dto.LoginAgreementDocument {
	result := make([]dto.LoginAgreementDocument, 0, len(items))
	for _, item := range items {
		result = append(result, dto.LoginAgreementDocument{
			ID:        item.ID,
			Title:     item.Title,
			ContentMD: item.ContentMD,
		})
	}
	return result
}
