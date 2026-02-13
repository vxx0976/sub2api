package dto

// SystemSettings represents the admin settings API response payload.
type SystemSettings struct {
	RegistrationEnabled         bool `json:"registration_enabled"`
	EmailVerifyEnabled          bool `json:"email_verify_enabled"`
	PromoCodeEnabled            bool `json:"promo_code_enabled"`
	PasswordResetEnabled        bool `json:"password_reset_enabled"`
	InvitationCodeEnabled       bool `json:"invitation_code_enabled"`
	TotpEnabled                 bool `json:"totp_enabled"`                   // TOTP 双因素认证
	TotpEncryptionKeyConfigured bool `json:"totp_encryption_key_configured"` // TOTP 加密密钥是否已配置

	SMTPHost               string `json:"smtp_host"`
	SMTPPort               int    `json:"smtp_port"`
	SMTPUsername           string `json:"smtp_username"`
	SMTPPasswordConfigured bool   `json:"smtp_password_configured"`
	SMTPFrom               string `json:"smtp_from_email"`
	SMTPFromName           string `json:"smtp_from_name"`
	SMTPUseTLS             bool   `json:"smtp_use_tls"`

	TurnstileEnabled             bool   `json:"turnstile_enabled"`
	TurnstileSiteKey             string `json:"turnstile_site_key"`
	TurnstileSecretKeyConfigured bool   `json:"turnstile_secret_key_configured"`

	LinuxDoConnectEnabled                bool   `json:"linuxdo_connect_enabled"`
	LinuxDoConnectClientID               string `json:"linuxdo_connect_client_id"`
	LinuxDoConnectClientSecretConfigured bool   `json:"linuxdo_connect_client_secret_configured"`
	LinuxDoConnectRedirectURL            string `json:"linuxdo_connect_redirect_url"`

	SiteName                    string `json:"site_name"`
	SiteLogo                    string `json:"site_logo"`
	SiteSubtitle                string `json:"site_subtitle"`
	APIBaseURL                  string `json:"api_base_url"`
	ContactInfo                 string `json:"contact_info"`
	DocURL                      string `json:"doc_url"`
	HomeContent                 string `json:"home_content"`
	HideCcsImportButton         bool   `json:"hide_ccs_import_button"`
	PurchaseSubscriptionEnabled bool   `json:"purchase_subscription_enabled"`
	PurchaseSubscriptionURL     string `json:"purchase_subscription_url"`
	CryptoAddresses             string `json:"crypto_addresses"`
	QueryDomain                 string `json:"query_domain"`

	DefaultConcurrency int     `json:"default_concurrency"`
	DefaultBalance     float64 `json:"default_balance"`

	// Model fallback configuration
	EnableModelFallback      bool   `json:"enable_model_fallback"`
	FallbackModelAnthropic   string `json:"fallback_model_anthropic"`
	FallbackModelOpenAI      string `json:"fallback_model_openai"`
	FallbackModelGemini      string `json:"fallback_model_gemini"`
	FallbackModelAntigravity string `json:"fallback_model_antigravity"`

	// Identity patch configuration (Claude -> Gemini)
	EnableIdentityPatch bool   `json:"enable_identity_patch"`
	IdentityPatchPrompt string `json:"identity_patch_prompt"`

	// Ops monitoring (vNext)
	OpsMonitoringEnabled         bool   `json:"ops_monitoring_enabled"`
	OpsRealtimeMonitoringEnabled bool   `json:"ops_realtime_monitoring_enabled"`
	OpsQueryModeDefault          string `json:"ops_query_mode_default"`
	OpsMetricsIntervalSeconds    int    `json:"ops_metrics_interval_seconds"`
}

// SimpleAnnouncement represents a single announcement item for public display
type SimpleAnnouncement struct {
	Title string `json:"title"`
	Date  string `json:"date,omitempty"`
}

type PublicSettings struct {
	RegistrationEnabled         bool           `json:"registration_enabled"`
	EmailVerifyEnabled          bool           `json:"email_verify_enabled"`
	PromoCodeEnabled            bool           `json:"promo_code_enabled"`
	PasswordResetEnabled        bool           `json:"password_reset_enabled"`
	InvitationCodeEnabled       bool           `json:"invitation_code_enabled"`
	TotpEnabled                 bool           `json:"totp_enabled"` // TOTP 双因素认证
	TurnstileEnabled            bool           `json:"turnstile_enabled"`
	TurnstileSiteKey            string         `json:"turnstile_site_key"`
	SiteName                    string         `json:"site_name"`
	SiteLogo                    string         `json:"site_logo"`
	SiteSubtitle                string         `json:"site_subtitle"`
	APIBaseURL                  string         `json:"api_base_url"`
	ContactInfo                 string         `json:"contact_info"`
	DocURL                      string         `json:"doc_url"`
	HomeContent                 string         `json:"home_content"`
	HomeTemplate                string         `json:"home_template,omitempty"`
	HideCcsImportButton         bool           `json:"hide_ccs_import_button"`
	PurchaseSubscriptionEnabled bool           `json:"purchase_subscription_enabled"`
	PurchaseSubscriptionURL     string         `json:"purchase_subscription_url"`
	PurchaseEnabled             bool           `json:"purchase_enabled,omitempty"`
	PurchaseURL                 string         `json:"purchase_url,omitempty"`
	LinuxDoOAuthEnabled         bool           `json:"linuxdo_oauth_enabled"`
	Version                     string         `json:"version"`
	Announcements               []SimpleAnnouncement `json:"announcements,omitempty"`
	CryptoAddresses             string         `json:"crypto_addresses"`
	QueryDomain                 string         `json:"query_domain"`
	DefaultLocale               string         `json:"default_locale,omitempty"`
	ContactWechat               string         `json:"contact_wechat,omitempty"`
	ContactTelegram             string         `json:"contact_telegram,omitempty"`

	// Reseller domain branding (populated when accessed via a reseller's custom domain)
	ResellerID     int64  `json:"reseller_id,omitempty"`
	ResellerDomain string `json:"reseller_domain,omitempty"`
	BrandColor     string `json:"brand_color,omitempty"`
	CustomCSS      string `json:"custom_css,omitempty"`
	LoginRedirect  string `json:"login_redirect,omitempty"`
	SEOTitle       string `json:"seo_title,omitempty"`
	SEODescription string `json:"seo_description,omitempty"`
	SEOKeywords    string `json:"seo_keywords,omitempty"`
}

// StreamTimeoutSettings 流超时处理配置 DTO
type StreamTimeoutSettings struct {
	Enabled                bool   `json:"enabled"`
	Action                 string `json:"action"`
	TempUnschedMinutes     int    `json:"temp_unsched_minutes"`
	ThresholdCount         int    `json:"threshold_count"`
	ThresholdWindowMinutes int    `json:"threshold_window_minutes"`
}
