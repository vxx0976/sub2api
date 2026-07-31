//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// ipPreserveRepoStub returns a fixed existing key from GetByID and captures the
// persisted key passed to Update, so we can assert IP rules survive a partial update.
type ipPreserveRepoStub struct {
	quotaBaseAPIKeyRepoStub
	existing *APIKey
	saved    *APIKey
}

func (s *ipPreserveRepoStub) GetByID(_ context.Context, _ int64) (*APIKey, error) {
	cp := *s.existing
	// fresh copies of the slices so each call sees the original
	cp.IPWhitelist = append([]string(nil), s.existing.IPWhitelist...)
	cp.IPBlacklist = append([]string(nil), s.existing.IPBlacklist...)
	return &cp, nil
}

func (s *ipPreserveRepoStub) Update(_ context.Context, k *APIKey, _ APIKeyUpdateFields) error {
	cp := *k
	s.saved = &cp
	return nil
}

// Regression: a status-only (or reset-quota) update must NOT wipe a key's IP
// whitelist/blacklist. Previously Update unconditionally overwrote them with the
// request's nil slices, silently stripping the key's IP ACL.
func TestAPIKeyService_Update_PreservesIPRulesOnPartialUpdate(t *testing.T) {
	repo := &ipPreserveRepoStub{
		existing: &APIKey{
			ID:          7,
			UserID:      42,
			Key:         "sk-ip-preserve",
			Status:      StatusActive,
			IPWhitelist: []string{"203.0.113.7"},
			IPBlacklist: []string{"198.51.100.0/24"},
		},
	}
	svc := &APIKeyService{apiKeyRepo: repo, cache: &quotaStateCacheStub{}}

	// Status-only update (disable).
	disabled := "disabled"
	_, err := svc.Update(context.Background(), 7, 42, UpdateAPIKeyRequest{Status: &disabled})
	require.NoError(t, err)
	require.NotNil(t, repo.saved)
	require.Equal(t, []string{"203.0.113.7"}, repo.saved.IPWhitelist, "status-only update must preserve IP whitelist")
	require.Equal(t, []string{"198.51.100.0/24"}, repo.saved.IPBlacklist, "status-only update must preserve IP blacklist")

	// Reset-quota only — also must preserve IP rules.
	repo.saved = nil
	resetTrue := true
	_, err = svc.Update(context.Background(), 7, 42, UpdateAPIKeyRequest{ResetQuota: &resetTrue})
	require.NoError(t, err)
	require.Equal(t, []string{"203.0.113.7"}, repo.saved.IPWhitelist, "reset-quota must preserve IP whitelist")
}

// An explicit empty slice still clears the list (the full-update semantics are unchanged).
func TestAPIKeyService_Update_ExplicitEmptyClearsIPRules(t *testing.T) {
	repo := &ipPreserveRepoStub{
		existing: &APIKey{
			ID:          7,
			UserID:      42,
			Key:         "sk-ip-clear",
			Status:      StatusActive,
			IPWhitelist: []string{"203.0.113.7"},
		},
	}
	svc := &APIKeyService{apiKeyRepo: repo, cache: &quotaStateCacheStub{}}

	empty := []string{}
	_, err := svc.Update(context.Background(), 7, 42, UpdateAPIKeyRequest{IPWhitelist: &empty})
	require.NoError(t, err)
	require.NotNil(t, repo.saved)
	require.Empty(t, repo.saved.IPWhitelist, "explicit empty slice must clear the whitelist")
}
