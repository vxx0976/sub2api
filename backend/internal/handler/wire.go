package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/handler/reseller"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/google/wire"
)

// ProvideAdminHandlers creates the AdminHandlers struct
func ProvideAdminHandlers(
	dashboardHandler *admin.DashboardHandler,
	userHandler *admin.UserHandler,
	groupHandler *admin.GroupHandler,
	accountHandler *admin.AccountHandler,
	announcementHandler *admin.AnnouncementHandler,
	dataManagementHandler *admin.DataManagementHandler,
	backupHandler *admin.BackupHandler,
	oauthHandler *admin.OAuthHandler,
	openaiOAuthHandler *admin.OpenAIOAuthHandler,
	geminiOAuthHandler *admin.GeminiOAuthHandler,
	antigravityOAuthHandler *admin.AntigravityOAuthHandler,
	proxyHandler *admin.ProxyHandler,
	redeemHandler *admin.RedeemHandler,
	promoHandler *admin.PromoHandler,
	settingHandler *admin.SettingHandler,
	opsHandler *admin.OpsHandler,
	systemHandler *admin.SystemHandler,
	subscriptionHandler *admin.SubscriptionHandler,
	usageHandler *admin.UsageHandler,
	userAttributeHandler *admin.UserAttributeHandler,
	referralHandler *admin.ReferralHandler,
	channelHandler *admin.ChannelHandler,
	errorPassthroughHandler *admin.ErrorPassthroughHandler,
	tlsFingerprintProfileHandler *admin.TLSFingerprintProfileHandler,
	apiKeyHandler *admin.AdminAPIKeyHandler,
	scheduledTestHandler *admin.ScheduledTestHandler,
	merchantHandler *admin.MerchantHandler,
	adminWithdrawalHandler *admin.AdminWithdrawalHandler,
) *AdminHandlers {
	return &AdminHandlers{
		Dashboard:             dashboardHandler,
		User:                  userHandler,
		Group:                 groupHandler,
		Account:               accountHandler,
		Announcement:          announcementHandler,
		DataManagement:        dataManagementHandler,
		Backup:                backupHandler,
		OAuth:                 oauthHandler,
		OpenAIOAuth:           openaiOAuthHandler,
		GeminiOAuth:           geminiOAuthHandler,
		AntigravityOAuth:      antigravityOAuthHandler,
		Proxy:                 proxyHandler,
		Redeem:                redeemHandler,
		Promo:                 promoHandler,
		Setting:               settingHandler,
		Ops:                   opsHandler,
		System:                systemHandler,
		Subscription:          subscriptionHandler,
		Usage:                 usageHandler,
		UserAttribute:         userAttributeHandler,
		Referral:              referralHandler,
		Channel:               channelHandler,
		ErrorPassthrough:      errorPassthroughHandler,
		TLSFingerprintProfile: tlsFingerprintProfileHandler,
		APIKey:                apiKeyHandler,
		ScheduledTest:         scheduledTestHandler,
		Merchant:              merchantHandler,
		AdminWithdrawal:       adminWithdrawalHandler,
	}
}

// ProvideResellerHandlers creates the ResellerHandlers struct
func ProvideResellerHandlers(
	dashboardHandler *reseller.DashboardHandler,
	domainHandler *reseller.DomainHandler,
	settingHandler *reseller.SettingHandler,
	keyHandler *reseller.KeyHandler,
	redeemHandler *reseller.RedeemHandler,
	announcementHandler *reseller.AnnouncementHandler,
	userHandler *reseller.UserHandler,
	commissionHandler *reseller.CommissionHandler,
	withdrawalHandler *reseller.WithdrawalHandler,
) *ResellerHandlers {
	return &ResellerHandlers{
		Dashboard:    dashboardHandler,
		Domain:       domainHandler,
		Setting:      settingHandler,
		Key:          keyHandler,
		Redeem:       redeemHandler,
		Announcement: announcementHandler,
		User:         userHandler,
		Commission:   commissionHandler,
		Withdrawal:   withdrawalHandler,
	}
}

// ProvideSystemHandler creates admin.SystemHandler with UpdateService
func ProvideSystemHandler(updateService *service.UpdateService, lockService *service.SystemOperationLockService) *admin.SystemHandler {
	return admin.NewSystemHandler(updateService, lockService)
}

// ProvideSettingHandler creates SettingHandler with version from BuildInfo
func ProvideSettingHandler(settingService *service.SettingService, buildInfo BuildInfo) *SettingHandler {
	return NewSettingHandler(settingService, buildInfo.Version)
}

// ProvideHandlers creates the Handlers struct
func ProvideHandlers(
	authHandler *AuthHandler,
	userHandler *UserHandler,
	apiKeyHandler *APIKeyHandler,
	usageHandler *UsageHandler,
	redeemHandler *RedeemHandler,
	subscriptionHandler *SubscriptionHandler,
	announcementHandler *AnnouncementHandler,
	adminHandlers *AdminHandlers,
	resellerHandlers *ResellerHandlers,
	gatewayHandler *GatewayHandler,
	openaiGatewayHandler *OpenAIGatewayHandler,
	soraGatewayHandler *SoraGatewayHandler,
	soraClientHandler *SoraClientHandler,
	settingHandler *SettingHandler,
	referralHandler *ReferralHandler,
	totpHandler *TotpHandler,
	keyQueryHandler *KeyQueryHandler,
	rechargeHandler *RechargeHandler,
	_ *service.IdempotencyCoordinator,
	_ *service.IdempotencyCleanupService,
) *Handlers {
	return &Handlers{
		Auth:          authHandler,
		User:          userHandler,
		APIKey:        apiKeyHandler,
		Usage:         usageHandler,
		Redeem:        redeemHandler,
		Subscription:  subscriptionHandler,
		Announcement:  announcementHandler,
		Admin:         adminHandlers,
		Reseller:      resellerHandlers,
		Gateway:       gatewayHandler,
		OpenAIGateway: openaiGatewayHandler,
		SoraGateway:   soraGatewayHandler,
		SoraClient:    soraClientHandler,
		Setting:       settingHandler,
		Referral:      referralHandler,
		Totp:          totpHandler,
		KeyQuery:      keyQueryHandler,
		Recharge:      rechargeHandler,
	}
}

// ProviderSet is the Wire provider set for all handlers
var ProviderSet = wire.NewSet(
	// Top-level handlers
	NewAuthHandler,
	NewUserHandler,
	NewAPIKeyHandler,
	NewUsageHandler,
	NewRedeemHandler,
	NewSubscriptionHandler,
	NewAnnouncementHandler,
	NewGatewayHandler,
	NewOpenAIGatewayHandler,
	NewSoraGatewayHandler,
	NewSoraClientHandler,
	NewTotpHandler,
	ProvideSettingHandler,
	NewReferralHandler,
	NewKeyQueryHandler,
	NewRechargeHandler,

	// Admin handlers
	admin.NewDashboardHandler,
	admin.NewUserHandler,
	admin.NewGroupHandler,
	admin.NewAccountHandler,
	admin.NewAnnouncementHandler,
	admin.NewDataManagementHandler,
	admin.NewBackupHandler,
	admin.NewOAuthHandler,
	admin.NewOpenAIOAuthHandler,
	admin.NewGeminiOAuthHandler,
	admin.NewAntigravityOAuthHandler,
	admin.NewProxyHandler,
	admin.NewRedeemHandler,
	admin.NewPromoHandler,
	admin.NewSettingHandler,
	admin.NewOpsHandler,
	ProvideSystemHandler,
	admin.NewSubscriptionHandler,
	admin.NewUsageHandler,
	admin.NewUserAttributeHandler,
	admin.NewReferralHandler,
	admin.NewChannelHandler,
	admin.NewErrorPassthroughHandler,
	admin.NewTLSFingerprintProfileHandler,
	admin.NewAdminAPIKeyHandler,
	admin.NewScheduledTestHandler,
	admin.NewMerchantHandler,
	admin.NewAdminWithdrawalHandler,

	// Reseller handlers
	reseller.NewDashboardHandler,
	reseller.NewDomainHandler,
	reseller.NewSettingHandler,
	reseller.NewKeyHandler,
	reseller.NewRedeemHandler,
	reseller.NewAnnouncementHandler,
	reseller.NewUserHandler,
	reseller.NewCommissionHandler,
	reseller.NewWithdrawalHandler,

	// AdminHandlers, ResellerHandlers and Handlers constructors
	ProvideAdminHandlers,
	ProvideResellerHandlers,
	ProvideHandlers,
)
