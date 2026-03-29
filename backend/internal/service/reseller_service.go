package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
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
	ErrVerifyFailed  = infraerrors.BadRequest("VERIFY_FAILED", "domain verification failed")
	ErrGroupNotOwned = infraerrors.Forbidden("GROUP_NOT_OWNED", "group does not belong to this reseller")
)


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
	// ListAllDomainNamesByResellerID returns all non-deleted domain name strings for the given reseller (no pagination).
	ListAllDomainNamesByResellerID(ctx context.Context, resellerID int64) ([]string, error)
	// CountByResellerID returns total and verified domain counts for the given reseller.
	CountByResellerID(ctx context.Context, resellerID int64) (total int, verified int, err error)
	// GetDomainsByResellerIDs returns a map of resellerID -> first verified domain (or first domain if none verified).
	GetDomainsByResellerIDs(ctx context.Context, resellerIDs []int64) (map[int64]string, error)
	// PurgeSoftDeletedByDomain physically deletes soft-deleted records for a given domain name.
	PurgeSoftDeletedByDomain(ctx context.Context, domain string)
}

// ResellerDashboardStats contains reseller dashboard statistics
type ResellerDashboardStats struct {
	MyBalance        float64 `json:"my_balance"`
	DomainCount      int     `json:"domain_count"`
	VerifiedDomains  int     `json:"verified_domains"`
	GroupCount       int     `json:"group_count"`
	KeyCount         int     `json:"key_count"`
	ActiveKeyCount   int     `json:"active_key_count"`
	TotalQuotaUsed   float64 `json:"total_quota_used"`
	TotalRechargeAmt float64 `json:"total_recharge_amount"`
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
	userRepo          UserRepository
	domainRepo        ResellerDomainRepository
	groupRepo         GroupRepository
	settingRepo       ResellerSettingRepository
	apiKeyRepo        APIKeyRepository
	apiKeyService     *APIKeyService
	redeemRepo        RedeemCodeRepository
	announcementRepo  AnnouncementRepository
	sub2apipayService *Sub2apipayService
	entClient         *dbent.Client
}

// NewResellerService creates a new ResellerService
func NewResellerService(
	userRepo UserRepository,
	domainRepo ResellerDomainRepository,
	groupRepo GroupRepository,
	settingRepo ResellerSettingRepository,
	apiKeyRepo APIKeyRepository,
	apiKeyService *APIKeyService,
	redeemRepo RedeemCodeRepository,
	announcementRepo AnnouncementRepository,
	sub2apipayService *Sub2apipayService,
	client *dbent.Client,
) *ResellerService {
	return &ResellerService{
		userRepo:          userRepo,
		domainRepo:        domainRepo,
		groupRepo:         groupRepo,
		settingRepo:       settingRepo,
		apiKeyRepo:        apiKeyRepo,
		apiKeyService:     apiKeyService,
		redeemRepo:        redeemRepo,
		announcementRepo:  announcementRepo,
		sub2apipayService: sub2apipayService,
		entClient:         client,
	}
}

// --- Dashboard ---

// GetDashboardStats returns dashboard statistics for a reseller
func (s *ResellerService) GetDashboardStats(ctx context.Context, resellerID int64) (*ResellerDashboardStats, error) {
	reseller, err := s.userRepo.GetByID(ctx, resellerID)
	if err != nil {
		return nil, fmt.Errorf("get reseller: %w", err)
	}

	// Count domains (two COUNT queries instead of full scan)
	domainTotal, domainVerified, err := s.domainRepo.CountByResellerID(ctx, resellerID)
	if err != nil {
		return nil, fmt.Errorf("count domains: %w", err)
	}

	// Count groups owned by reseller (single COUNT query)
	groupCount, err := s.groupRepo.CountByOwnerID(ctx, resellerID)
	if err != nil {
		groupCount = 0
	}

	// Count and sum API keys owned by reseller (three COUNT/SUM queries instead of full scan)
	keyCount, err := s.apiKeyRepo.CountByUserID(ctx, resellerID)
	if err != nil {
		keyCount = 0
	}
	activeKeyCount, err := s.apiKeyRepo.CountActiveByUserID(ctx, resellerID)
	if err != nil {
		activeKeyCount = 0
	}
	totalQuotaUsed, err := s.apiKeyRepo.SumQuotaUsedByUserID(ctx, resellerID)
	if err != nil {
		totalQuotaUsed = 0
	}

	// Get total recharge amount for reseller's users from sub2apipay
	userIDs, err := s.userRepo.ListIDsByParentID(ctx, resellerID)
	if err != nil {
		userIDs = []int64{}
	}
	totalRechargeAmt := 0.0
	if len(userIDs) > 0 && s.sub2apipayService != nil {
		totalRechargeAmt, err = s.sub2apipayService.SumRechargeByUserIDs(ctx, userIDs)
		if err != nil {
			totalRechargeAmt = 0
		}
	}

	return &ResellerDashboardStats{
		MyBalance:        reseller.Balance,
		DomainCount:      domainTotal,
		VerifiedDomains:  domainVerified,
		GroupCount:       int(groupCount),
		KeyCount:         int(keyCount),
		ActiveKeyCount:   int(activeKeyCount),
		TotalQuotaUsed:   totalQuotaUsed,
		TotalRechargeAmt: totalRechargeAmt,
	}, nil
}

// --- Domain management ---

// ListDomains returns the reseller's domains
func (s *ResellerService) ListDomains(ctx context.Context, resellerID int64, page, pageSize int) ([]ResellerDomain, *pagination.PaginationResult, error) {
	return s.domainRepo.ListByResellerID(ctx, resellerID, pagination.PaginationParams{Page: page, PageSize: pageSize})
}

// ListAllDomainNames returns all domain name strings for the reseller without pagination.
func (s *ResellerService) ListAllDomainNames(ctx context.Context, resellerID int64) ([]string, error) {
	return s.domainRepo.ListAllDomainNamesByResellerID(ctx, resellerID)
}

// CreateDomain creates a new domain for the reseller
func (s *ResellerService) CreateDomain(ctx context.Context, resellerID int64, input *CreateDomainInput) (*ResellerDomain, error) {
	// Check if domain already exists (active, not soft-deleted)
	existing, _ := s.domainRepo.GetByDomain(ctx, input.Domain)
	if existing != nil {
		return nil, ErrDomainExists
	}

	// Purge any soft-deleted record with the same domain to avoid unique constraint violation
	s.domainRepo.PurgeSoftDeletedByDomain(ctx, input.Domain)

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

// DeleteDomain deletes a domain, validating ownership. Returns the domain name for cache invalidation.
func (s *ResellerService) DeleteDomain(ctx context.Context, resellerID, domainID int64) (string, error) {
	domain, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return "", ErrDomainNotFound
	}
	if domain.ResellerID != resellerID {
		return "", ErrOwnershipViolation
	}
	return domain.Domain, s.domainRepo.Delete(ctx, domainID)
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

// --- API Key management ---

// ListKeys returns API keys owned by the reseller
func (s *ResellerService) ListKeys(ctx context.Context, resellerID int64, page, pageSize int) ([]APIKey, *pagination.PaginationResult, error) {
	return s.apiKeyRepo.ListByUserID(ctx, resellerID, pagination.PaginationParams{Page: page, PageSize: pageSize}, APIKeyListFilters{})
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

// --- Redeem Code management ---

// GenerateRedeemCodes creates redeem codes funded from the reseller's balance
func (s *ResellerService) GenerateRedeemCodes(ctx context.Context, resellerID int64, count int, value float64) ([]RedeemCode, error) {
	if count <= 0 || count > 100 {
		return nil, infraerrors.BadRequest("INVALID_COUNT", "count must be between 1 and 100")
	}
	if value <= 0 {
		return nil, infraerrors.BadRequest("INVALID_VALUE", "value must be positive")
	}

	totalCost := value * float64(count)

	// Generate code strings before starting the transaction (no DB I/O needed)
	codes := make([]RedeemCode, count)
	for i := 0; i < count; i++ {
		code, err := GenerateRedeemCode()
		if err != nil {
			return nil, fmt.Errorf("generate code: %w", err)
		}
		codes[i] = RedeemCode{
			Code:    code,
			Type:    RedeemTypeBalance,
			Value:   value,
			Status:  StatusUnused,
			OwnerID: &resellerID,
		}
	}

	// Use a transaction: atomically deduct balance and create codes
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)

	if err := s.userRepo.DeductBalanceIfSufficient(txCtx, resellerID, totalCost); err != nil {
		return nil, fmt.Errorf("deduct balance: %w", err)
	}

	if err := s.redeemRepo.CreateBatch(txCtx, codes); err != nil {
		return nil, fmt.Errorf("create codes: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return codes, nil
}

// ListRedeemCodes returns redeem codes owned by the reseller
func (s *ResellerService) ListRedeemCodes(ctx context.Context, resellerID int64, page, pageSize int) ([]RedeemCode, *pagination.PaginationResult, error) {
	return s.redeemRepo.ListByOwnerID(ctx, resellerID, pagination.PaginationParams{Page: page, PageSize: pageSize})
}

// DeleteRedeemCode deletes a redeem code owned by the reseller
func (s *ResellerService) DeleteRedeemCode(ctx context.Context, resellerID, codeID int64) error {
	code, err := s.redeemRepo.GetByID(ctx, codeID)
	if err != nil {
		return infraerrors.NotFound("REDEEM_CODE_NOT_FOUND", "redeem code not found")
	}
	if code.OwnerID == nil || *code.OwnerID != resellerID {
		return ErrOwnershipViolation
	}
	if code.Status != StatusUnused {
		return infraerrors.BadRequest("REDEEM_CODE_USED", "cannot delete a used redeem code")
	}

	// Atomically refund balance (ADD operation)
	if err := s.userRepo.UpdateBalance(ctx, resellerID, code.Value); err != nil {
		return fmt.Errorf("refund balance: %w", err)
	}

	return s.redeemRepo.Delete(ctx, codeID)
}

// --- Announcement management ---

// ListAnnouncements returns announcements owned by the reseller
func (s *ResellerService) ListAnnouncements(ctx context.Context, resellerID int64, page, pageSize int) ([]Announcement, *pagination.PaginationResult, error) {
	return s.announcementRepo.ListByOwnerID(ctx, resellerID, pagination.PaginationParams{Page: page, PageSize: pageSize})
}

// CreateAnnouncement creates an announcement owned by the reseller
func (s *ResellerService) CreateAnnouncement(ctx context.Context, resellerID int64, input *CreateAnnouncementInput) (*Announcement, error) {
	a := &Announcement{
		Title:   input.Title,
		Content: input.Content,
		Status:  input.Status,
		OwnerID: &resellerID,
	}
	if a.Status == "" {
		a.Status = AnnouncementStatusDraft
	}
	if input.StartsAt != nil {
		a.StartsAt = input.StartsAt
	}
	if input.EndsAt != nil {
		a.EndsAt = input.EndsAt
	}

	if err := s.announcementRepo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("create announcement: %w", err)
	}
	return a, nil
}

// UpdateAnnouncement updates an announcement owned by the reseller
func (s *ResellerService) UpdateAnnouncement(ctx context.Context, resellerID, announcementID int64, input *UpdateAnnouncementInput) (*Announcement, error) {
	a, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil {
		return nil, infraerrors.NotFound("ANNOUNCEMENT_NOT_FOUND", "announcement not found")
	}
	if a.OwnerID == nil || *a.OwnerID != resellerID {
		return nil, ErrOwnershipViolation
	}

	if input.Title != nil {
		a.Title = *input.Title
	}
	if input.Content != nil {
		a.Content = *input.Content
	}
	if input.Status != nil {
		a.Status = *input.Status
	}
	if input.StartsAt != nil {
		a.StartsAt = *input.StartsAt
	}
	if input.EndsAt != nil {
		a.EndsAt = *input.EndsAt
	}

	if err := s.announcementRepo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("update announcement: %w", err)
	}
	return a, nil
}

// DeleteAnnouncement deletes an announcement owned by the reseller
func (s *ResellerService) DeleteAnnouncement(ctx context.Context, resellerID, announcementID int64) error {
	a, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil {
		return infraerrors.NotFound("ANNOUNCEMENT_NOT_FOUND", "announcement not found")
	}
	if a.OwnerID == nil || *a.OwnerID != resellerID {
		return ErrOwnershipViolation
	}
	return s.announcementRepo.Delete(ctx, announcementID)
}

// --- User management ---

// ListUsers returns users belonging to the reseller (parent_id = resellerID)
func (s *ResellerService) ListUsers(ctx context.Context, resellerID int64, page, pageSize int, filters UserListFilters) ([]User, *pagination.PaginationResult, error) {
	// Force ParentID to resellerID to prevent privilege escalation
	filters.ParentID = &resellerID
	return s.userRepo.ListWithFilters(ctx, pagination.PaginationParams{Page: page, PageSize: pageSize}, filters)
}

// TransferBalance transfers balance between the reseller and a user.
// operation "add" (充值): deducts from reseller, adds to user.
// operation "subtract" (退款): deducts from user, adds back to reseller.
func (s *ResellerService) TransferBalance(ctx context.Context, resellerID, userID int64, amount float64, operation string, notes string) (*User, error) {
	// Validate user belongs to this reseller
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.ParentID == nil || *user.ParentID != resellerID {
		return nil, ErrOwnershipViolation
	}

	var balanceDiff float64
	switch operation {
	case "add", "subtract":
		// validated below inside transaction
	default:
		return nil, infraerrors.BadRequest("INVALID_OPERATION", "operation must be 'add' or 'subtract'")
	}

	// Use a transaction to atomically transfer balance
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)

	switch operation {
	case "add":
		// Deduct from reseller, add to user
		if err := s.userRepo.DeductBalanceIfSufficient(txCtx, resellerID, amount); err != nil {
			return nil, fmt.Errorf("deduct reseller balance: %w", err)
		}
		if err := s.userRepo.UpdateBalance(txCtx, userID, amount); err != nil {
			return nil, fmt.Errorf("add user balance: %w", err)
		}
		balanceDiff = amount
	case "subtract":
		// Deduct from user, add back to reseller
		if err := s.userRepo.DeductBalanceIfSufficient(txCtx, userID, amount); err != nil {
			return nil, fmt.Errorf("deduct user balance: %w", err)
		}
		if err := s.userRepo.UpdateBalance(txCtx, resellerID, amount); err != nil {
			return nil, fmt.Errorf("add reseller balance: %w", err)
		}
		balanceDiff = -amount
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	// Reload updated user to return accurate balance
	updatedUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		updatedUser = user
	}

	// Create audit record outside transaction; failure only logs a warning
	code, err := GenerateRedeemCode()
	if err == nil {
		now := time.Now()
		adjustmentRecord := &RedeemCode{
			Code:    code,
			Type:    AdjustmentTypeResellerTransfer,
			Value:   balanceDiff,
			Status:  StatusUsed,
			UsedBy:  &userID,
			OwnerID: &resellerID,
			Notes:   notes,
		}
		adjustmentRecord.UsedAt = &now
		if err := s.redeemRepo.Create(ctx, adjustmentRecord); err != nil {
			slog.Warn("failed to create reseller transfer audit record", "error", err)
		}
	}

	return updatedUser, nil
}

// GetUserBalanceHistory returns paginated balance history for a user belonging to the reseller.
func (s *ResellerService) GetUserBalanceHistory(ctx context.Context, resellerID, userID int64, page, pageSize int) ([]RedeemCode, int64, float64, error) {
	// Validate user belongs to this reseller
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, 0, 0, err
	}
	if user.ParentID == nil || *user.ParentID != resellerID {
		return nil, 0, 0, ErrOwnershipViolation
	}

	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	codes, result, err := s.redeemRepo.ListByUserPaginated(ctx, userID, params, "")
	if err != nil {
		return nil, 0, 0, err
	}

	totalRecharged, err := s.redeemRepo.SumPositiveBalanceByUser(ctx, userID)
	if err != nil {
		return nil, 0, 0, err
	}

	return codes, result.Total, totalRecharged, nil
}

// --- Internal helpers ---

func generateVerifyToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
