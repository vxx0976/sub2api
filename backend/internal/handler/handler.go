package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/handler/reseller"
)

// AdminHandlers contains all admin-related HTTP handlers
type AdminHandlers struct {
	Dashboard             *admin.DashboardHandler
	User                  *admin.UserHandler
	Group                 *admin.GroupHandler
	Account               *admin.AccountHandler
	Announcement          *admin.AnnouncementHandler
	DataManagement        *admin.DataManagementHandler
	Backup                *admin.BackupHandler
	OAuth                 *admin.OAuthHandler
	OpenAIOAuth           *admin.OpenAIOAuthHandler
	GeminiOAuth           *admin.GeminiOAuthHandler
	AntigravityOAuth      *admin.AntigravityOAuthHandler
	Proxy                 *admin.ProxyHandler
	Redeem                *admin.RedeemHandler
	Promo                 *admin.PromoHandler
	Setting               *admin.SettingHandler
	Ops                   *admin.OpsHandler
	System                *admin.SystemHandler
	Subscription          *admin.SubscriptionHandler
	Usage                 *admin.UsageHandler
	UserAttribute         *admin.UserAttributeHandler
	Referral              *admin.ReferralHandler
	Channel               *admin.ChannelHandler
	ErrorPassthrough      *admin.ErrorPassthroughHandler
	TLSFingerprintProfile *admin.TLSFingerprintProfileHandler
	APIKey                *admin.AdminAPIKeyHandler
	ScheduledTest         *admin.ScheduledTestHandler
	Merchant              *admin.MerchantHandler
	AdminWithdrawal       *admin.AdminWithdrawalHandler
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
}

// Handlers contains all HTTP handlers
type Handlers struct {
	Auth          *AuthHandler
	User          *UserHandler
	APIKey        *APIKeyHandler
	Usage         *UsageHandler
	Redeem        *RedeemHandler
	Subscription  *SubscriptionHandler
	Announcement  *AnnouncementHandler
	Admin         *AdminHandlers
	Reseller      *ResellerHandlers
	Gateway       *GatewayHandler
	OpenAIGateway *OpenAIGatewayHandler
	SoraGateway   *SoraGatewayHandler
	SoraClient    *SoraClientHandler
	Setting       *SettingHandler
	Referral      *ReferralHandler
	Totp          *TotpHandler
	KeyQuery      *KeyQueryHandler
	Recharge      *RechargeHandler
}

// BuildInfo contains build-time information
type BuildInfo struct {
	Version   string
	BuildType string // "source" for manual builds, "release" for CI builds
}
