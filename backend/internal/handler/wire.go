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
	grokOAuthHandler *admin.GrokOAuthHandler,
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
	channelMonitorHandler *admin.ChannelMonitorHandler,
	channelMonitorTemplateHandler *admin.ChannelMonitorRequestTemplateHandler,
	errorPassthroughHandler *admin.ErrorPassthroughHandler,
	tlsFingerprintProfileHandler *admin.TLSFingerprintProfileHandler,
	apiKeyHandler *admin.AdminAPIKeyHandler,
	scheduledTestHandler *admin.ScheduledTestHandler,
	merchantHandler *admin.MerchantHandler,
	adminWithdrawalHandler *admin.AdminWithdrawalHandler,
	contentModerationHandler *admin.ContentModerationHandler,
	paymentHandler *admin.PaymentHandler,
	affiliateHandler *admin.AffiliateHandler,
	translationHandler *admin.TranslationHandler,
	chatHandler *admin.ChatHandler,
	complianceHandler *admin.ComplianceHandler,
	modelPricingHandler *admin.ModelPricingHandler,
	auditLogHandler *admin.AuditLogHandler,
	upstreamBillingProbe *service.UpstreamBillingProbeService,
) *AdminHandlers {
	accountHandler.SetUpstreamBillingProbeService(upstreamBillingProbe)
	return &AdminHandlers{
		Dashboard:              dashboardHandler,
		User:                   userHandler,
		Group:                  groupHandler,
		Account:                accountHandler,
		Announcement:           announcementHandler,
		DataManagement:         dataManagementHandler,
		Backup:                 backupHandler,
		OAuth:                  oauthHandler,
		OpenAIOAuth:            openaiOAuthHandler,
		GeminiOAuth:            geminiOAuthHandler,
		AntigravityOAuth:       antigravityOAuthHandler,
		GrokOAuth:              grokOAuthHandler,
		Proxy:                  proxyHandler,
		Redeem:                 redeemHandler,
		Promo:                  promoHandler,
		Setting:                settingHandler,
		Ops:                    opsHandler,
		System:                 systemHandler,
		Subscription:           subscriptionHandler,
		Usage:                  usageHandler,
		UserAttribute:          userAttributeHandler,
		Referral:               referralHandler,
		Channel:                channelHandler,
		ChannelMonitor:         channelMonitorHandler,
		ChannelMonitorTemplate: channelMonitorTemplateHandler,
		ErrorPassthrough:       errorPassthroughHandler,
		TLSFingerprintProfile:  tlsFingerprintProfileHandler,
		APIKey:                 apiKeyHandler,
		ScheduledTest:          scheduledTestHandler,
		Merchant:               merchantHandler,
		AdminWithdrawal:        adminWithdrawalHandler,
		ContentModeration:      contentModerationHandler,
		Payment:                paymentHandler,
		Affiliate:              affiliateHandler,
		Translation:            translationHandler,
		Chat:                   chatHandler,
		Compliance:             complianceHandler,
		ModelPricing:           modelPricingHandler,
		AuditLog:               auditLogHandler,
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
	serviceTokenHandler *reseller.ServiceTokenHandler,
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
		ServiceToken: serviceTokenHandler,
	}
}

// ProvideSystemHandler creates admin.SystemHandler with UpdateService
func ProvideSystemHandler(updateService *service.UpdateService, lockService *service.SystemOperationLockService) *admin.SystemHandler {
	return admin.NewSystemHandler(updateService, lockService)
}

// ProvideSettingHandler creates SettingHandler with version from BuildInfo
func ProvideSettingHandler(settingService *service.SettingService, buildInfo BuildInfo, notificationEmailService *service.NotificationEmailService) *SettingHandler {
	h := NewSettingHandler(settingService, buildInfo.Version)
	h.SetNotificationEmailService(notificationEmailService)
	return h
}

// ProvideAdminSettingHandler creates admin.SettingHandler with notification template APIs.
func ProvideAdminSettingHandler(settingService *service.SettingService, emailService *service.EmailService, turnstileService *service.TurnstileService, opsService *service.OpsService, paymentConfigService *service.PaymentConfigService, paymentService *service.PaymentService, userAttributeService *service.UserAttributeService, notificationEmailService *service.NotificationEmailService) *admin.SettingHandler {
	h := admin.NewSettingHandler(settingService, emailService, turnstileService, opsService, paymentConfigService, paymentService, userAttributeService)
	h.SetNotificationEmailService(notificationEmailService)
	return h
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
	channelMonitorUserHandler *ChannelMonitorUserHandler,
	adminHandlers *AdminHandlers,
	resellerHandlers *ResellerHandlers,
	gatewayHandler *GatewayHandler,
	openaiGatewayHandler *OpenAIGatewayHandler,
	settingHandler *SettingHandler,
	referralHandler *ReferralHandler,
	totpHandler *TotpHandler,
	keyQueryHandler *KeyQueryHandler,
	rechargeHandler *RechargeHandler,
	orderHandler *OrderHandler,
	usdtHandler *UsdtHandler,
	mergedOrderHandler *MergedOrderHandler,
	paymentHandler *PaymentHandler,
	paymentWebhookHandler *PaymentWebhookHandler,
	availableChannelHandler *AvailableChannelHandler,
	chatHandler *ChatHandler,
	asyncImageHandler *AsyncImageHandler,
	batchImageHandler *BatchImageHandler,
	_ *service.IdempotencyCoordinator,
	_ *service.IdempotencyCleanupService,
) *Handlers {
	return &Handlers{
		Auth:             authHandler,
		User:             userHandler,
		APIKey:           apiKeyHandler,
		Usage:            usageHandler,
		Redeem:           redeemHandler,
		Subscription:     subscriptionHandler,
		Announcement:     announcementHandler,
		ChannelMonitor:   channelMonitorUserHandler,
		Admin:            adminHandlers,
		Reseller:         resellerHandlers,
		Gateway:          gatewayHandler,
		OpenAIGateway:    openaiGatewayHandler,
		Setting:          settingHandler,
		Referral:         referralHandler,
		Totp:             totpHandler,
		KeyQuery:         keyQueryHandler,
		Recharge:         rechargeHandler,
		Order:            orderHandler,
		Usdt:             usdtHandler,
		MergedOrder:      mergedOrderHandler,
		Payment:          paymentHandler,
		PaymentWebhook:   paymentWebhookHandler,
		AvailableChannel: availableChannelHandler,
		Chat:             chatHandler,
		AsyncImage:       asyncImageHandler,
		BatchImage:       batchImageHandler,
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
	NewChannelMonitorUserHandler,
	NewGatewayHandler,
	NewOpenAIGatewayHandler,
	NewTotpHandler,
	ProvideSettingHandler,
	NewReferralHandler,
	NewKeyQueryHandler,
	NewRechargeHandler,
	NewOrderHandler,
	NewUsdtHandler,
	NewMergedOrderHandler,
	NewPaymentHandler,
	NewPaymentWebhookHandler,
	NewAvailableChannelHandler,
	NewChatHandler,
	NewAsyncImageHandler,
	NewBatchImageHandler,

	// Admin handlers
	admin.NewDashboardHandler,
	admin.NewUserHandler,
	admin.NewGroupHandler,
	admin.ProvideAccountHandler,
	admin.NewAnnouncementHandler,
	admin.NewDataManagementHandler,
	admin.NewBackupHandler,
	admin.NewOAuthHandler,
	admin.NewOpenAIOAuthHandler,
	admin.NewGeminiOAuthHandler,
	admin.NewAntigravityOAuthHandler,
	admin.NewGrokOAuthHandler,
	admin.NewProxyHandler,
	admin.NewRedeemHandler,
	admin.NewPromoHandler,
	ProvideAdminSettingHandler,
	admin.NewOpsHandler,
	ProvideSystemHandler,
	admin.NewSubscriptionHandler,
	admin.NewUsageHandler,
	admin.NewUserAttributeHandler,
	admin.NewReferralHandler,
	admin.NewChannelHandler,
	admin.NewChannelMonitorHandler,
	admin.NewChannelMonitorRequestTemplateHandler,
	admin.NewErrorPassthroughHandler,
	admin.NewTLSFingerprintProfileHandler,
	admin.NewAdminAPIKeyHandler,
	admin.NewScheduledTestHandler,
	admin.NewMerchantHandler,
	admin.NewAdminWithdrawalHandler,
	admin.NewContentModerationHandler,
	admin.NewModelPricingHandler,
	admin.NewPaymentHandler,
	admin.NewAffiliateHandler,
	admin.NewTranslationHandler,
	admin.NewChatHandler,
	admin.NewComplianceHandler,
	admin.NewAuditLogHandler,

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
	reseller.NewServiceTokenHandler,

	// AdminHandlers, ResellerHandlers and Handlers constructors
	ProvideAdminHandlers,
	ProvideResellerHandlers,
	ProvideHandlers,
)
