package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrNotReseller        = infraerrors.Forbidden("NOT_RESELLER", "reseller access required")
	ErrOwnershipViolation = infraerrors.Forbidden("OWNERSHIP_VIOLATION", "you do not own this resource")
	ErrSelfOperation      = infraerrors.BadRequest("SELF_OPERATION", "cannot perform this operation on yourself")
	ErrDomainExists       = infraerrors.Conflict("DOMAIN_EXISTS", "domain already registered")
	ErrDomainNotFound     = infraerrors.NotFound("DOMAIN_NOT_FOUND", "domain not found")
	ErrDomainNotVerified  = infraerrors.BadRequest("DOMAIN_NOT_VERIFIED", "domain is not verified")
	ErrVerifyFailed       = infraerrors.BadRequest("VERIFY_FAILED", "domain verification failed")
	ErrTemplateNotFound   = infraerrors.NotFound("TEMPLATE_NOT_FOUND", "template group not found")
	ErrGroupNotOwned      = infraerrors.Forbidden("GROUP_NOT_OWNED", "group does not belong to this reseller")
)

// Reseller Telegram setting keys
const (
	ResellerSettingTgBotToken = "tg_bot_token"
	ResellerSettingTgChatID   = "tg_chat_id"    // reseller's own chat_id
	ResellerSettingTgBindCode = "tg_bind_code"   // temporary bind code
	ResellerSettingTgFeatures = "tg_features"    // comma-separated feature flags
)

// Default Telegram features (all enabled)
const DefaultTgFeatures = "admin_keys,admin_stats,admin_notify,user_query,user_notify"

// Protected reseller setting keys that cannot be set directly via API
var protectedResellerSettingKeys = map[string]bool{
	ResellerSettingTgChatID:   true,
	ResellerSettingTgBindCode: true,
}

// IsProtectedResellerSetting returns true if the setting key is protected.
func IsProtectedResellerSetting(key string) bool {
	return protectedResellerSettingKeys[key]
}

// ResellerSettingRepository defines the interface for reseller settings data access
type ResellerSettingRepository interface {
	GetAll(ctx context.Context, resellerID int64) (map[string]string, error)
	Get(ctx context.Context, resellerID int64, key string) (string, error)
	Set(ctx context.Context, resellerID int64, key, value string) error
	SetAll(ctx context.Context, resellerID int64, settings map[string]string) error
	// ListByKey returns a map of resellerID -> value for all resellers that have the given key set
	ListByKey(ctx context.Context, key string) (map[int64]string, error)
}

// ResellerDomain represents a reseller's custom domain
type ResellerDomain struct {
	ID              int64      `json:"id"`
	ResellerID      int64      `json:"reseller_id"`
	Domain          string     `json:"domain"`
	SiteName        string     `json:"site_name"`
	SiteLogo        string     `json:"site_logo"`
	BrandColor      string     `json:"brand_color"`
	CustomCSS       string     `json:"custom_css,omitempty"`
	Subtitle        string     `json:"subtitle"`
	HomeContent     string     `json:"home_content,omitempty"`
	DocURL          string     `json:"doc_url,omitempty"`
	HomeTemplate    string     `json:"home_template,omitempty"`
	PurchaseEnabled bool       `json:"purchase_enabled"`
	PurchaseURL     string     `json:"purchase_url,omitempty"`
	DefaultLocale   string     `json:"default_locale,omitempty"`
	SEOTitle        string     `json:"seo_title,omitempty"`
	SEODescription  string     `json:"seo_description,omitempty"`
	SEOKeywords     string     `json:"seo_keywords,omitempty"`
	LoginRedirect   string     `json:"login_redirect,omitempty"`
	VerifyToken     string     `json:"verify_token,omitempty"`
	Verified        bool       `json:"verified"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ResellerDomainRepository defines the interface for reseller domain data access
type ResellerDomainRepository interface {
	Create(ctx context.Context, domain *ResellerDomain) error
	GetByID(ctx context.Context, id int64) (*ResellerDomain, error)
	GetByDomain(ctx context.Context, domain string) (*ResellerDomain, error)
	Update(ctx context.Context, domain *ResellerDomain) error
	Delete(ctx context.Context, id int64) error
	ListByResellerID(ctx context.Context, resellerID int64, params pagination.PaginationParams) ([]ResellerDomain, *pagination.PaginationResult, error)
}

// ResellerDashboardStats contains reseller dashboard statistics
type ResellerDashboardStats struct {
	MyBalance       float64 `json:"my_balance"`
	DomainCount     int     `json:"domain_count"`
	VerifiedDomains int     `json:"verified_domains"`
	GroupCount      int     `json:"group_count"`
	KeyCount        int     `json:"key_count"`
	ActiveKeyCount  int     `json:"active_key_count"`
	TotalQuotaUsed  float64 `json:"total_quota_used"`
}

// CreateDomainInput represents the input for creating a domain
type CreateDomainInput struct {
	Domain          string `json:"domain" binding:"required"`
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
}

// UpdateDomainInput represents the input for updating a domain
type UpdateDomainInput struct {
	SiteName        *string `json:"site_name"`
	SiteLogo        *string `json:"site_logo"`
	BrandColor      *string `json:"brand_color"`
	CustomCSS       *string `json:"custom_css"`
	Subtitle        *string `json:"subtitle"`
	HomeContent     *string `json:"home_content"`
	DocURL          *string `json:"doc_url"`
	HomeTemplate    *string `json:"home_template"`
	PurchaseEnabled *bool   `json:"purchase_enabled"`
	PurchaseURL     *string `json:"purchase_url"`
	DefaultLocale   *string `json:"default_locale"`
	SEOTitle        *string `json:"seo_title"`
	SEODescription  *string `json:"seo_description"`
	SEOKeywords     *string `json:"seo_keywords"`
	LoginRedirect   *string `json:"login_redirect"`
}

// CreateResellerGroupInput represents the input for creating a reseller group (package)
type CreateResellerGroupInput struct {
	Name                string   `json:"name" binding:"required"`
	Description         string   `json:"description"`
	SourceGroupID       int64    `json:"source_group_id" binding:"required"`
	Price               *float64 `json:"price"`
	DailyLimitUSD       *float64 `json:"daily_limit_usd"`
	WeeklyLimitUSD      *float64 `json:"weekly_limit_usd"`
	MonthlyLimitUSD     *float64 `json:"monthly_limit_usd"`
	DefaultValidityDays int      `json:"default_validity_days"`
	IsPurchasable       bool     `json:"is_purchasable"`
	SortOrder           int      `json:"sort_order"`
}

// UpdateResellerGroupInput represents the input for updating a reseller group
type UpdateResellerGroupInput struct {
	Name                *string  `json:"name"`
	Description         *string  `json:"description"`
	Price               *float64 `json:"price"`
	DailyLimitUSD       *float64 `json:"daily_limit_usd"`
	WeeklyLimitUSD      *float64 `json:"weekly_limit_usd"`
	MonthlyLimitUSD     *float64 `json:"monthly_limit_usd"`
	DefaultValidityDays *int     `json:"default_validity_days"`
	IsPurchasable       *bool    `json:"is_purchasable"`
	SortOrder           *int     `json:"sort_order"`
	Status              *string  `json:"status"`
}

// CreateResellerKeyInput represents the input for creating an API key for the reseller
type CreateResellerKeyInput struct {
	Name          string   `json:"name"`
	GroupID       *int64   `json:"group_id"`
	CustomKey     *string  `json:"custom_key"`
	Notes         string   `json:"notes"`
	Quota         float64  `json:"quota"`
	ExpiresInDays *int     `json:"expires_in_days"`
	IPWhitelist   []string `json:"ip_whitelist"`
	IPBlacklist   []string `json:"ip_blacklist"`
}

// UpdateResellerKeyInput represents the input for updating a reseller's API key
type UpdateResellerKeyInput struct {
	Name        *string    `json:"name"`
	Notes       *string    `json:"notes"`
	GroupID     *int64     `json:"group_id"`
	Status      *string    `json:"status"`
	Quota       *float64   `json:"quota"`
	ExpiresAt   *time.Time `json:"expires_at"`
	IPWhitelist []string   `json:"ip_whitelist"`
	IPBlacklist []string   `json:"ip_blacklist"`
}

// ResellerService provides reseller-specific business logic
type ResellerService struct {
	userRepo      UserRepository
	domainRepo    ResellerDomainRepository
	groupRepo     GroupRepository
	settingRepo   ResellerSettingRepository
	apiKeyRepo    APIKeyRepository
	apiKeyService *APIKeyService
}

// NewResellerService creates a new ResellerService
func NewResellerService(
	userRepo UserRepository,
	domainRepo ResellerDomainRepository,
	groupRepo GroupRepository,
	settingRepo ResellerSettingRepository,
	apiKeyRepo APIKeyRepository,
	apiKeyService *APIKeyService,
) *ResellerService {
	return &ResellerService{
		userRepo:      userRepo,
		domainRepo:    domainRepo,
		groupRepo:     groupRepo,
		settingRepo:   settingRepo,
		apiKeyRepo:    apiKeyRepo,
		apiKeyService: apiKeyService,
	}
}

// --- Dashboard ---

// GetDashboardStats returns dashboard statistics for a reseller
func (s *ResellerService) GetDashboardStats(ctx context.Context, resellerID int64) (*ResellerDashboardStats, error) {
	reseller, err := s.userRepo.GetByID(ctx, resellerID)
	if err != nil {
		return nil, fmt.Errorf("get reseller: %w", err)
	}

	// Count domains
	domains, _, err := s.domainRepo.ListByResellerID(ctx, resellerID, pagination.PaginationParams{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}
	var verifiedCount int
	for _, d := range domains {
		if d.Verified {
			verifiedCount++
		}
	}

	// Count groups owned by reseller
	groups, _, err := s.groupRepo.ListWithFilters(ctx, pagination.PaginationParams{Page: 1, PageSize: 10000}, "", "", "", nil, nil)
	var groupCount int
	if err == nil {
		for _, g := range groups {
			if g.OwnerID != nil && *g.OwnerID == resellerID {
				groupCount++
			}
		}
	}

	// Count API keys owned by reseller
	keys, _, err := s.apiKeyRepo.ListByUserID(ctx, resellerID, pagination.PaginationParams{Page: 1, PageSize: 10000})
	var keyCount, activeKeyCount int
	var totalQuotaUsed float64
	if err == nil {
		keyCount = len(keys)
		for _, k := range keys {
			if k.Status == StatusActive {
				activeKeyCount++
			}
			totalQuotaUsed += k.QuotaUsed
		}
	}

	return &ResellerDashboardStats{
		MyBalance:       reseller.Balance,
		DomainCount:     len(domains),
		VerifiedDomains: verifiedCount,
		GroupCount:      groupCount,
		KeyCount:        keyCount,
		ActiveKeyCount:  activeKeyCount,
		TotalQuotaUsed:  totalQuotaUsed,
	}, nil
}

// --- Domain management ---

// ListDomains returns the reseller's domains
func (s *ResellerService) ListDomains(ctx context.Context, resellerID int64, page, pageSize int) ([]ResellerDomain, *pagination.PaginationResult, error) {
	return s.domainRepo.ListByResellerID(ctx, resellerID, pagination.PaginationParams{Page: page, PageSize: pageSize})
}

// CreateDomain creates a new domain for the reseller
func (s *ResellerService) CreateDomain(ctx context.Context, resellerID int64, input *CreateDomainInput) (*ResellerDomain, error) {
	// Check if domain already exists
	existing, _ := s.domainRepo.GetByDomain(ctx, input.Domain)
	if existing != nil {
		return nil, ErrDomainExists
	}

	// Generate verification token
	token, err := generateVerifyToken()
	if err != nil {
		return nil, fmt.Errorf("generate verify token: %w", err)
	}

	domain := &ResellerDomain{
		ResellerID:      resellerID,
		Domain:          input.Domain,
		SiteName:        input.SiteName,
		SiteLogo:        input.SiteLogo,
		BrandColor:      input.BrandColor,
		CustomCSS:       input.CustomCSS,
		Subtitle:        input.Subtitle,
		HomeContent:     input.HomeContent,
		DocURL:          input.DocURL,
		HomeTemplate:    input.HomeTemplate,
		PurchaseEnabled: input.PurchaseEnabled,
		PurchaseURL:     input.PurchaseURL,
		DefaultLocale:   input.DefaultLocale,
		SEOTitle:        input.SEOTitle,
		SEODescription:  input.SEODescription,
		SEOKeywords:     input.SEOKeywords,
		LoginRedirect:   input.LoginRedirect,
		VerifyToken:     token,
	}

	if err := s.domainRepo.Create(ctx, domain); err != nil {
		return nil, fmt.Errorf("create domain: %w", err)
	}

	return domain, nil
}

// UpdateDomain updates a domain, validating ownership
func (s *ResellerService) UpdateDomain(ctx context.Context, resellerID, domainID int64, input *UpdateDomainInput) (*ResellerDomain, error) {
	domain, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return nil, ErrDomainNotFound
	}
	if domain.ResellerID != resellerID {
		return nil, ErrOwnershipViolation
	}

	if input.SiteName != nil {
		domain.SiteName = *input.SiteName
	}
	if input.SiteLogo != nil {
		domain.SiteLogo = *input.SiteLogo
	}
	if input.BrandColor != nil {
		domain.BrandColor = *input.BrandColor
	}
	if input.CustomCSS != nil {
		domain.CustomCSS = *input.CustomCSS
	}
	if input.Subtitle != nil {
		domain.Subtitle = *input.Subtitle
	}
	if input.HomeContent != nil {
		domain.HomeContent = *input.HomeContent
	}
	if input.DocURL != nil {
		domain.DocURL = *input.DocURL
	}
	if input.HomeTemplate != nil {
		domain.HomeTemplate = *input.HomeTemplate
	}
	if input.PurchaseEnabled != nil {
		domain.PurchaseEnabled = *input.PurchaseEnabled
	}
	if input.PurchaseURL != nil {
		domain.PurchaseURL = *input.PurchaseURL
	}
	if input.DefaultLocale != nil {
		domain.DefaultLocale = *input.DefaultLocale
	}
	if input.SEOTitle != nil {
		domain.SEOTitle = *input.SEOTitle
	}
	if input.SEODescription != nil {
		domain.SEODescription = *input.SEODescription
	}
	if input.SEOKeywords != nil {
		domain.SEOKeywords = *input.SEOKeywords
	}
	if input.LoginRedirect != nil {
		domain.LoginRedirect = *input.LoginRedirect
	}

	if err := s.domainRepo.Update(ctx, domain); err != nil {
		return nil, fmt.Errorf("update domain: %w", err)
	}

	return domain, nil
}

// DeleteDomain deletes a domain, validating ownership
func (s *ResellerService) DeleteDomain(ctx context.Context, resellerID, domainID int64) error {
	domain, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return ErrDomainNotFound
	}
	if domain.ResellerID != resellerID {
		return ErrOwnershipViolation
	}
	return s.domainRepo.Delete(ctx, domainID)
}

// VerifyDomain initiates or checks domain verification via DNS TXT record
func (s *ResellerService) VerifyDomain(ctx context.Context, resellerID, domainID int64) (*ResellerDomain, error) {
	domain, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return nil, ErrDomainNotFound
	}
	if domain.ResellerID != resellerID {
		return nil, ErrOwnershipViolation
	}
	if domain.Verified {
		return domain, nil
	}

	// Perform DNS TXT verification
	verified, err := verifyDNSTXT(domain.Domain, domain.VerifyToken)
	if err != nil {
		return nil, fmt.Errorf("DNS verification: %w", err)
	}
	if !verified {
		return nil, ErrVerifyFailed
	}

	now := time.Now()
	domain.Verified = true
	domain.VerifiedAt = &now

	if err := s.domainRepo.Update(ctx, domain); err != nil {
		return nil, fmt.Errorf("update domain: %w", err)
	}

	return domain, nil
}

// --- Group (Package) management ---

// ListTemplates returns admin groups marked as reseller templates
func (s *ResellerService) ListTemplates(ctx context.Context) ([]Group, error) {
	groups, _, err := s.groupRepo.ListWithFilters(ctx, pagination.PaginationParams{Page: 1, PageSize: 1000}, "", "", "", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	var templates []Group
	for _, g := range groups {
		if g.ResellerTemplate && g.OwnerID == nil {
			templates = append(templates, Group{
				ID:          g.ID,
				Name:        g.Name,
				Description: g.Description,
			})
		}
	}
	return templates, nil
}

// CreateGroup creates a new group (package) for the reseller
func (s *ResellerService) CreateGroup(ctx context.Context, resellerID int64, input *CreateResellerGroupInput) (*Group, error) {
	// Validate source template
	template, err := s.groupRepo.GetByID(ctx, input.SourceGroupID)
	if err != nil || template == nil {
		return nil, ErrTemplateNotFound
	}
	if !template.ResellerTemplate || template.OwnerID != nil {
		return nil, ErrTemplateNotFound
	}

	// Create a new group based on the template
	newGroup := &Group{
		Name:                 input.Name,
		Description:          input.Description,
		Platform:             template.Platform,
		SubscriptionType:     template.SubscriptionType,
		RateMultiplier:       template.RateMultiplier,
		ModelRouting:         template.ModelRouting,
		ModelRoutingEnabled:  template.ModelRoutingEnabled,
		MCPXMLInject:         template.MCPXMLInject,
		SupportedModelScopes: template.SupportedModelScopes,
		Status:               StatusActive,
		OwnerID:              &resellerID,
		SourceGroupID:        &input.SourceGroupID,
		Price:                input.Price,
		DailyLimitUSD:        input.DailyLimitUSD,
		WeeklyLimitUSD:       input.WeeklyLimitUSD,
		MonthlyLimitUSD:      input.MonthlyLimitUSD,
		DefaultValidityDays:  input.DefaultValidityDays,
		IsPurchasable:        input.IsPurchasable,
		SortOrder:            input.SortOrder,
		IsExclusive:          template.IsExclusive,
		ClaudeCodeOnly:       template.ClaudeCodeOnly,
	}
	if newGroup.DefaultValidityDays <= 0 {
		newGroup.DefaultValidityDays = 30
	}

	if err := s.groupRepo.Create(ctx, newGroup); err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}

	// Clone account bindings from template
	accountIDs, err := s.groupRepo.GetAccountIDsByGroupIDs(ctx, []int64{input.SourceGroupID})
	if err == nil && len(accountIDs) > 0 {
		_ = s.groupRepo.BindAccountsToGroup(ctx, newGroup.ID, accountIDs)
	}

	return newGroup, nil
}

// ListGroups returns groups owned by the reseller
func (s *ResellerService) ListGroups(ctx context.Context, resellerID int64, page, pageSize int) ([]Group, *pagination.PaginationResult, error) {
	groups, _, err := s.groupRepo.ListWithFilters(ctx, pagination.PaginationParams{Page: 1, PageSize: 10000}, "", "", "", nil, nil)
	if err != nil {
		return nil, nil, err
	}
	// Filter by owner
	var filtered []Group
	for _, g := range groups {
		if g.OwnerID != nil && *g.OwnerID == resellerID {
			filtered = append(filtered, g)
		}
	}
	total := int64(len(filtered))
	// Apply pagination on filtered results
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	pages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return filtered[start:end], &pagination.PaginationResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}, nil
}

// GetGroup returns a group owned by the reseller
func (s *ResellerService) GetGroup(ctx context.Context, resellerID, groupID int64) (*Group, error) {
	g, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if g.OwnerID == nil || *g.OwnerID != resellerID {
		return nil, ErrGroupNotOwned
	}
	return g, nil
}

// UpdateGroup updates a group owned by the reseller
func (s *ResellerService) UpdateGroup(ctx context.Context, resellerID, groupID int64, input *UpdateResellerGroupInput) (*Group, error) {
	g, err := s.GetGroup(ctx, resellerID, groupID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		g.Name = *input.Name
	}
	if input.Description != nil {
		g.Description = *input.Description
	}
	if input.Price != nil {
		g.Price = input.Price
	}
	if input.DailyLimitUSD != nil {
		g.DailyLimitUSD = input.DailyLimitUSD
	}
	if input.WeeklyLimitUSD != nil {
		g.WeeklyLimitUSD = input.WeeklyLimitUSD
	}
	if input.MonthlyLimitUSD != nil {
		g.MonthlyLimitUSD = input.MonthlyLimitUSD
	}
	if input.DefaultValidityDays != nil {
		g.DefaultValidityDays = *input.DefaultValidityDays
	}
	if input.IsPurchasable != nil {
		g.IsPurchasable = *input.IsPurchasable
	}
	if input.SortOrder != nil {
		g.SortOrder = *input.SortOrder
	}
	if input.Status != nil {
		g.Status = *input.Status
	}

	if err := s.groupRepo.Update(ctx, g); err != nil {
		return nil, fmt.Errorf("update group: %w", err)
	}
	return g, nil
}

// DeleteGroup deletes a group owned by the reseller
func (s *ResellerService) DeleteGroup(ctx context.Context, resellerID, groupID int64) error {
	if _, err := s.GetGroup(ctx, resellerID, groupID); err != nil {
		return err
	}
	_, err := s.groupRepo.DeleteCascade(ctx, groupID)
	return err
}

// --- API Key management ---

// ListKeys returns API keys owned by the reseller
func (s *ResellerService) ListKeys(ctx context.Context, resellerID int64, page, pageSize int) ([]APIKey, *pagination.PaginationResult, error) {
	return s.apiKeyRepo.ListByUserID(ctx, resellerID, pagination.PaginationParams{Page: page, PageSize: pageSize})
}

// CreateKey creates a new API key for the reseller
func (s *ResellerService) CreateKey(ctx context.Context, resellerID int64, input *CreateResellerKeyInput) (*APIKey, error) {
	// Validate group ownership if specified, and read group defaults
	var group *Group
	if input.GroupID != nil {
		g, err := s.groupRepo.GetByID(ctx, *input.GroupID)
		if err != nil {
			return nil, fmt.Errorf("get group: %w", err)
		}
		// Allow binding to reseller's own groups or public non-exclusive groups
		if g.OwnerID != nil && *g.OwnerID != resellerID {
			return nil, ErrGroupNotOwned
		}
		group = g
	}

	// Apply group defaults if not explicitly provided
	expiresInDays := input.ExpiresInDays
	quota := input.Quota
	if group != nil {
		if (expiresInDays == nil || *expiresInDays == 0) && group.DefaultValidityDays > 0 {
			expiresInDays = &group.DefaultValidityDays
		}
		if quota == 0 && group.MonthlyLimitUSD != nil && *group.MonthlyLimitUSD > 0 {
			quota = *group.MonthlyLimitUSD
		}
	}

	// Delegate to the existing API key service
	req := CreateAPIKeyRequest{
		Name:          input.Name,
		GroupID:       input.GroupID,
		CustomKey:     input.CustomKey,
		Quota:         quota,
		ExpiresInDays: expiresInDays,
		IPWhitelist:   input.IPWhitelist,
		IPBlacklist:   input.IPBlacklist,
	}

	key, err := s.apiKeyService.Create(ctx, resellerID, req)
	if err != nil {
		return nil, err
	}

	// Update notes if provided
	if input.Notes != "" {
		key.Notes = input.Notes
		if err := s.apiKeyRepo.Update(ctx, key); err != nil {
			return nil, fmt.Errorf("update notes: %w", err)
		}
	}

	return key, nil
}

// UpdateKey updates an API key owned by the reseller
func (s *ResellerService) UpdateKey(ctx context.Context, resellerID, keyID int64, input *UpdateResellerKeyInput) (*APIKey, error) {
	// Build update request
	req := UpdateAPIKeyRequest{
		Name:        input.Name,
		Status:      input.Status,
		Quota:       input.Quota,
		ExpiresAt:   input.ExpiresAt,
		IPWhitelist: input.IPWhitelist,
		IPBlacklist: input.IPBlacklist,
	}

	if input.GroupID != nil {
		// Validate group ownership
		g, err := s.groupRepo.GetByID(ctx, *input.GroupID)
		if err != nil {
			return nil, fmt.Errorf("get group: %w", err)
		}
		if g.OwnerID != nil && *g.OwnerID != resellerID {
			return nil, ErrGroupNotOwned
		}
		req.GroupID = input.GroupID
	}

	key, err := s.apiKeyService.Update(ctx, keyID, resellerID, req)
	if err != nil {
		return nil, err
	}

	// Update notes if provided
	if input.Notes != nil {
		key.Notes = *input.Notes
		if err := s.apiKeyRepo.Update(ctx, key); err != nil {
			return nil, fmt.Errorf("update notes: %w", err)
		}
	}

	return key, nil
}

// DeleteKey deletes an API key owned by the reseller
func (s *ResellerService) DeleteKey(ctx context.Context, resellerID, keyID int64) error {
	return s.apiKeyService.Delete(ctx, keyID, resellerID)
}

// ResetKeyQuota resets the quota of an API key owned by the reseller
func (s *ResellerService) ResetKeyQuota(ctx context.Context, resellerID, keyID int64) (*APIKey, error) {
	resetTrue := true
	req := UpdateAPIKeyRequest{
		ResetQuota: &resetTrue,
	}
	return s.apiKeyService.Update(ctx, keyID, resellerID, req)
}

// --- Settings management ---

// GetSettings returns all settings for the reseller
func (s *ResellerService) GetSettings(ctx context.Context, resellerID int64) (map[string]string, error) {
	return s.settingRepo.GetAll(ctx, resellerID)
}

// UpdateSettings updates settings for the reseller
func (s *ResellerService) UpdateSettings(ctx context.Context, resellerID int64, settings map[string]string) error {
	return s.settingRepo.SetAll(ctx, resellerID, settings)
}

// --- Internal helpers ---

func generateVerifyToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
