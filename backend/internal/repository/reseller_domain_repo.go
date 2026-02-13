package repository

import (
	"context"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbresellerdomain "github.com/Wei-Shaw/sub2api/ent/resellerdomain"
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"entgo.io/ent/dialect/sql"
)

type resellerDomainRepository struct {
	client *dbent.Client
}

// NewResellerDomainRepository creates a new ResellerDomainRepository
func NewResellerDomainRepository(client *dbent.Client) service.ResellerDomainRepository {
	return &resellerDomainRepository{client: client}
}

func (r *resellerDomainRepository) Create(ctx context.Context, domain *service.ResellerDomain) error {
	builder := r.client.ResellerDomain.Create().
		SetResellerID(domain.ResellerID).
		SetDomain(domain.Domain).
		SetSiteName(domain.SiteName).
		SetSiteLogo(domain.SiteLogo).
		SetBrandColor(domain.BrandColor).
		SetCustomCSS(domain.CustomCSS).
		SetSubtitle(domain.Subtitle).
		SetHomeContent(domain.HomeContent).
		SetDocURL(domain.DocURL).
		SetHomeTemplate(domain.HomeTemplate).
		SetPurchaseEnabled(domain.PurchaseEnabled).
		SetPurchaseURL(domain.PurchaseURL).
		SetDefaultLocale(domain.DefaultLocale).
		SetSeoTitle(domain.SEOTitle).
		SetSeoDescription(domain.SEODescription).
		SetSeoKeywords(domain.SEOKeywords).
		SetLoginRedirect(domain.LoginRedirect).
		SetVerifyToken(domain.VerifyToken).
		SetVerified(domain.Verified)

	if domain.VerifiedAt != nil {
		builder = builder.SetVerifiedAt(*domain.VerifiedAt)
	}

	entity, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create reseller domain: %w", err)
	}

	domain.ID = entity.ID
	domain.CreatedAt = entity.CreatedAt
	domain.UpdatedAt = entity.UpdatedAt
	return nil
}

func (r *resellerDomainRepository) GetByID(ctx context.Context, id int64) (*service.ResellerDomain, error) {
	entity, err := r.client.ResellerDomain.Query().
		Where(
			dbresellerdomain.ID(id),
			dbresellerdomain.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get reseller domain by id: %w", err)
	}
	return resellerDomainEntityToService(entity), nil
}

func (r *resellerDomainRepository) GetByDomain(ctx context.Context, domain string) (*service.ResellerDomain, error) {
	entity, err := r.client.ResellerDomain.Query().
		Where(
			dbresellerdomain.DomainEQ(domain),
			dbresellerdomain.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get reseller domain by domain: %w", err)
	}
	return resellerDomainEntityToService(entity), nil
}

func (r *resellerDomainRepository) Update(ctx context.Context, domain *service.ResellerDomain) error {
	builder := r.client.ResellerDomain.UpdateOneID(domain.ID).
		SetSiteName(domain.SiteName).
		SetSiteLogo(domain.SiteLogo).
		SetBrandColor(domain.BrandColor).
		SetCustomCSS(domain.CustomCSS).
		SetSubtitle(domain.Subtitle).
		SetHomeContent(domain.HomeContent).
		SetDocURL(domain.DocURL).
		SetHomeTemplate(domain.HomeTemplate).
		SetPurchaseEnabled(domain.PurchaseEnabled).
		SetPurchaseURL(domain.PurchaseURL).
		SetDefaultLocale(domain.DefaultLocale).
		SetSeoTitle(domain.SEOTitle).
		SetSeoDescription(domain.SEODescription).
		SetSeoKeywords(domain.SEOKeywords).
		SetLoginRedirect(domain.LoginRedirect).
		SetVerified(domain.Verified)

	if domain.VerifiedAt != nil {
		builder = builder.SetVerifiedAt(*domain.VerifiedAt)
	}

	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("update reseller domain: %w", err)
	}
	return nil
}

func (r *resellerDomainRepository) Delete(ctx context.Context, id int64) error {
	// Soft delete by setting deleted_at
	now := time.Now()
	_, err := r.client.ResellerDomain.UpdateOneID(id).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("delete reseller domain: %w", err)
	}
	return nil
}

func (r *resellerDomainRepository) ListByResellerID(ctx context.Context, resellerID int64, params pagination.PaginationParams) ([]service.ResellerDomain, *pagination.PaginationResult, error) {
	q := r.client.ResellerDomain.Query().
		Where(
			dbresellerdomain.ResellerIDEQ(resellerID),
			dbresellerdomain.DeletedAtIsNil(),
		)

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("count reseller domains: %w", err)
	}

	entities, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(sql.OrderByField(dbresellerdomain.FieldID, sql.OrderDesc()).ToFunc()).
		All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list reseller domains: %w", err)
	}

	domains := make([]service.ResellerDomain, 0, len(entities))
	for _, e := range entities {
		domains = append(domains, *resellerDomainEntityToService(e))
	}

	return domains, paginationResultFromTotal(int64(total), params), nil
}

func (r *resellerDomainRepository) GetDomainsByResellerIDs(ctx context.Context, resellerIDs []int64) (map[int64]string, error) {
	if len(resellerIDs) == 0 {
		return map[int64]string{}, nil
	}

	entities, err := r.client.ResellerDomain.Query().
		Where(
			dbresellerdomain.ResellerIDIn(resellerIDs...),
			dbresellerdomain.DeletedAtIsNil(),
		).
		Order(
			sql.OrderByField(dbresellerdomain.FieldVerified, sql.OrderDesc()).ToFunc(),
			sql.OrderByField(dbresellerdomain.FieldID, sql.OrderAsc()).ToFunc(),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get domains by reseller ids: %w", err)
	}

	// Pick the first domain per reseller (verified domains come first due to ordering)
	result := make(map[int64]string, len(resellerIDs))
	for _, e := range entities {
		if _, exists := result[e.ResellerID]; !exists {
			result[e.ResellerID] = e.Domain
		}
	}
	return result, nil
}

func (r *resellerDomainRepository) PurgeSoftDeletedByDomain(ctx context.Context, domain string) {
	// Physically delete soft-deleted records to avoid unique constraint violation on re-create.
	// Must use SkipSoftDelete to bypass the soft-delete hook, which would otherwise
	// add "deleted_at IS NULL" and convert DELETE to UPDATE, making the purge a no-op.
	_, _ = r.client.ResellerDomain.Delete().
		Where(
			dbresellerdomain.DomainEQ(domain),
			dbresellerdomain.DeletedAtNotNil(),
		).
		Exec(mixins.SkipSoftDelete(ctx))
}

func resellerDomainEntityToService(e *dbent.ResellerDomain) *service.ResellerDomain {
	return &service.ResellerDomain{
		ID:              e.ID,
		ResellerID:      e.ResellerID,
		Domain:          e.Domain,
		SiteName:        e.SiteName,
		SiteLogo:        e.SiteLogo,
		BrandColor:      e.BrandColor,
		CustomCSS:       e.CustomCSS,
		Subtitle:        e.Subtitle,
		HomeContent:     e.HomeContent,
		DocURL:          e.DocURL,
		HomeTemplate:    e.HomeTemplate,
		PurchaseEnabled: e.PurchaseEnabled,
		PurchaseURL:     e.PurchaseURL,
		DefaultLocale:   e.DefaultLocale,
		SEOTitle:        e.SeoTitle,
		SEODescription:  e.SeoDescription,
		SEOKeywords:     e.SeoKeywords,
		LoginRedirect:   e.LoginRedirect,
		VerifyToken:     e.VerifyToken,
		Verified:        e.Verified,
		VerifiedAt:      e.VerifiedAt,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}
