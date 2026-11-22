package usagestats

import "context"

type businessScopeKey struct{}

// WithBusinessScope marks dashboard business statistics so shared usage queries
// can exclude administrators without changing personal usage or operational metrics.
func WithBusinessScope(ctx context.Context) context.Context {
	return context.WithValue(ctx, businessScopeKey{}, true)
}

// IsBusinessScope reports whether a query belongs to dashboard business statistics.
func IsBusinessScope(ctx context.Context) bool {
	business, _ := ctx.Value(businessScopeKey{}).(bool)
	return business
}
