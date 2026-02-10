package middleware

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// ResellerDomainContext contains reseller branding info extracted from the custom domain
type ResellerDomainContext struct {
	ResellerID      int64  `json:"reseller_id"`
	Domain          string `json:"domain"`
	SiteName        string `json:"site_name"`
	SiteLogo        string `json:"site_logo"`
	BrandColor      string `json:"brand_color"`
	CustomCSS       string `json:"custom_css"`
	Subtitle        string `json:"subtitle"`
	HomeContent     string `json:"home_content"`
	DocURL          string `json:"doc_url"`
	HomeTemplate    string `json:"home_template"`
	PurchaseEnabled bool   `json:"purchase_enabled"`
	PurchaseURL     string `json:"purchase_url"`
	DefaultLocale   string `json:"default_locale"`
	SEOTitle        string `json:"seo_title"`
	SEODescription  string `json:"seo_description"`
	SEOKeywords     string `json:"seo_keywords"`
	LoginRedirect   string `json:"login_redirect"`
	// Reseller-global settings from reseller_settings KV table
	ResellerSettings map[string]string `json:"reseller_settings,omitempty"`
}

const (
	// ContextKeyResellerDomain is the gin context key for reseller domain info
	ContextKeyResellerDomain = "reseller_domain"
)

// GetResellerDomainFromContext retrieves the reseller domain context from gin.Context
func GetResellerDomainFromContext(c *gin.Context) *ResellerDomainContext {
	if v, exists := c.Get(ContextKeyResellerDomain); exists {
		if info, ok := v.(*ResellerDomainContext); ok {
			return info
		}
	}
	return nil
}

// DomainDetect creates middleware that detects reseller custom domains from the Host header.
// If a verified reseller domain is found, it stores branding info in gin.Context.
func DomainDetect(domainRepo service.ResellerDomainRepository, settingRepo service.ResellerSettingRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host

		// Strip port if present
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}

		// Normalize to lowercase
		host = strings.ToLower(strings.TrimSpace(host))

		// Skip common non-custom-domain cases
		if host == "" || host == "localhost" || net.ParseIP(host) != nil {
			c.Next()
			return
		}

		// Look up domain in database
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		domain, err := domainRepo.GetByDomain(ctx, host)
		if err != nil || domain == nil || !domain.Verified {
			// Not a reseller domain — continue normally
			c.Next()
			return
		}

		// Load reseller-level settings
		var resellerSettings map[string]string
		if settingRepo != nil {
			resellerSettings, _ = settingRepo.GetAll(ctx, domain.ResellerID)
		}

		// Set reseller domain info in context for downstream handlers
		c.Set(ContextKeyResellerDomain, &ResellerDomainContext{
			ResellerID:       domain.ResellerID,
			Domain:           domain.Domain,
			SiteName:         domain.SiteName,
			SiteLogo:         domain.SiteLogo,
			BrandColor:       domain.BrandColor,
			CustomCSS:        domain.CustomCSS,
			Subtitle:         domain.Subtitle,
			HomeContent:      domain.HomeContent,
			DocURL:           domain.DocURL,
			HomeTemplate:     domain.HomeTemplate,
			PurchaseEnabled:  domain.PurchaseEnabled,
			PurchaseURL:      domain.PurchaseURL,
			DefaultLocale:    domain.DefaultLocale,
			SEOTitle:         domain.SEOTitle,
			SEODescription:   domain.SEODescription,
			SEOKeywords:      domain.SEOKeywords,
			LoginRedirect:    domain.LoginRedirect,
			ResellerSettings: resellerSettings,
		})

		c.Next()
	}
}
