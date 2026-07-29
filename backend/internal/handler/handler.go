package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/handler/reseller"
	"github.com/Wei-Shaw/sub2api/internal/securityaudit"
)

// AdminHandlers contains all admin-related HTTP handlers
type AdminHandlers struct {
	Dashboard              *admin.DashboardHandler
	User                   *admin.UserHandler
	Group                  *admin.GroupHandler
	Account                *admin.AccountHandler
	Announcement           *admin.AnnouncementHandler
	DataManagement         *admin.DataManagementHandler
	Backup                 *admin.BackupHandler
	OAuth                  *admin.OAuthHandler
	OpenAIOAuth            *admin.OpenAIOAuthHandler
	GeminiOAuth            *admin.GeminiOAuthHandler
	AntigravityOAuth       *admin.AntigravityOAuthHandler
	GrokOAuth              *admin.GrokOAuthHandler
	Proxy                  *admin.ProxyHandler
	Redeem                 *admin.RedeemHandler
	Promo                  *admin.PromoHandler
	Setting                *admin.SettingHandler
	Ops                    *admin.OpsHandler
	System                 *admin.SystemHandler
	Subscription           *admin.SubscriptionHandler
	Usage                  *admin.UsageHandler
	UserAttribute          *admin.UserAttributeHandler
	Referral               *admin.ReferralHandler
	Channel                *admin.ChannelHandler
	ChannelMonitor         *admin.ChannelMonitorHandler
	ChannelMonitorTemplate *admin.ChannelMonitorRequestTemplateHandler
	ErrorPassthrough       *admin.ErrorPassthroughHandler
	TLSFingerprintProfile  *admin.TLSFingerprintProfileHandler
	APIKey                 *admin.AdminAPIKeyHandler
	ScheduledTest          *admin.ScheduledTestHandler
	Merchant               *admin.MerchantHandler
	AdminWithdrawal        *admin.AdminWithdrawalHandler
	ContentModeration      *admin.ContentModerationHandler
	ModelPricing           *admin.ModelPricingHandler
	PromptAudit            *securityaudit.PromptAdminHandler
	Payment                *admin.PaymentHandler
	Affiliate              *admin.AffiliateHandler
	Translation            *admin.TranslationHandler
	Chat                   *admin.ChatHandler
	Compliance             *admin.ComplianceHandler
	AuditLog               *admin.AuditLogHandler
}

// ResellerHandlers contains all reseller-related HTTP handlers
type ResellerHandlers struct {
	Dashboard    *reseller.DashboardHandler
	Domain       *reseller.DomainHandler
	Setting      *reseller.SettingHandler
	Key          *reseller.KeyHandler
	Redeem       *reseller.RedeemHandler
	Announcement *reseller.AnnouncementHandler
	User         *reseller.UserHandler
	Commission   *reseller.CommissionHandler
	Withdrawal   *reseller.WithdrawalHandler
	ServiceToken *reseller.ServiceTokenHandler
}

// Handlers contains all HTTP handlers
type Handlers struct {
	Auth             *AuthHandler
	User             *UserHandler
	APIKey           *APIKeyHandler
	Usage            *UsageHandler
	Redeem           *RedeemHandler
	Subscription     *SubscriptionHandler
	Announcement     *AnnouncementHandler
	ChannelMonitor   *ChannelMonitorUserHandler
	Admin            *AdminHandlers
	Reseller         *ResellerHandlers
	Gateway          *GatewayHandler
	OpenAIGateway    *OpenAIGatewayHandler
	Setting          *SettingHandler
	Referral         *ReferralHandler
	Totp             *TotpHandler
	Passkey          *PasskeyHandler
	KeyQuery         *KeyQueryHandler
	Recharge         *RechargeHandler
	Order            *OrderHandler
	Usdt             *UsdtHandler
	MergedOrder      *MergedOrderHandler
	Payment          *PaymentHandler
	PaymentWebhook   *PaymentWebhookHandler
	AvailableChannel *AvailableChannelHandler
	ModelPlaza       *ModelPlazaHandler
	Chat             *ChatHandler
	AsyncImage       *AsyncImageHandler
	BatchImage       *BatchImageHandler
}

// BuildInfo contains build-time information
type BuildInfo struct {
	Version   string
	BuildType string // "source" for manual builds, "release" for CI builds
}
