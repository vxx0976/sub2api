package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// fakeResellerAPITokenRepo 是 ResellerAPITokenRepository 的内存实现，用于单测。
type fakeResellerAPITokenRepo struct {
	rows   map[int64]*ResellerAPIToken
	nextID int64
}

func newFakeResellerAPITokenRepo() *fakeResellerAPITokenRepo {
	return &fakeResellerAPITokenRepo{rows: map[int64]*ResellerAPIToken{}, nextID: 0}
}

func (f *fakeResellerAPITokenRepo) Create(_ context.Context, token *ResellerAPIToken) (*ResellerAPIToken, error) {
	f.nextID++
	cp := *token
	cp.ID = f.nextID
	cp.CreatedAt = time.Now()
	cp.UpdatedAt = cp.CreatedAt
	f.rows[cp.ID] = &cp
	out := cp
	return &out, nil
}

func (f *fakeResellerAPITokenRepo) GetByHash(_ context.Context, hash string) (*ResellerAPIToken, error) {
	for _, r := range f.rows {
		if r.TokenHash == hash {
			out := *r
			return &out, nil
		}
	}
	return nil, nil
}

func (f *fakeResellerAPITokenRepo) GetByIDForReseller(_ context.Context, id, resellerID int64) (*ResellerAPIToken, error) {
	r, ok := f.rows[id]
	if !ok || r.ResellerID != resellerID {
		return nil, nil
	}
	out := *r
	return &out, nil
}

func (f *fakeResellerAPITokenRepo) ListByReseller(_ context.Context, resellerID int64) ([]*ResellerAPIToken, error) {
	var out []*ResellerAPIToken
	for _, r := range f.rows {
		if r.ResellerID == resellerID {
			cp := *r
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakeResellerAPITokenRepo) UpdateStatus(_ context.Context, id int64, status string) error {
	if r, ok := f.rows[id]; ok {
		r.Status = status
	}
	return nil
}

func (f *fakeResellerAPITokenRepo) TouchLastUsed(_ context.Context, id int64, at time.Time) error {
	if r, ok := f.rows[id]; ok {
		r.LastUsedAt = &at
	}
	return nil
}

func TestResellerAPITokenService_GenerateAndValidate(t *testing.T) {
	repo := newFakeResellerAPITokenRepo()
	svc := NewResellerAPITokenService(repo)
	ctx := context.Background()

	plaintext, token, err := svc.GenerateToken(ctx, 42, "client-backend", nil)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(plaintext, ResellerServiceTokenPrefix), "plaintext must carry the rst- prefix")
	require.Equal(t, int64(42), token.ResellerID)
	require.Equal(t, resellerTokenStatusActive, token.Status)
	require.Nil(t, token.ExpiresAt, "nil ttl means never expires")

	// 明文不落库：存储的是 SHA-256 哈希，且与明文不同。
	require.NotEqual(t, plaintext, token.TokenHash)
	require.Equal(t, hashServiceToken(plaintext), token.TokenHash)
	require.Len(t, token.TokenHash, 64)

	// 正确明文可校验通过。
	got, err := svc.ValidateToken(ctx, plaintext)
	require.NoError(t, err)
	require.Equal(t, token.ID, got.ID)
}

func TestResellerAPITokenService_ValidateRejects(t *testing.T) {
	repo := newFakeResellerAPITokenRepo()
	svc := NewResellerAPITokenService(repo)
	ctx := context.Background()

	plaintext, _, err := svc.GenerateToken(ctx, 7, "", nil)
	require.NoError(t, err)

	// 错误前缀 / 空串 / 随机串都应被拒。
	for _, bad := range []string{"", "garbage", "sk-deadbeef", plaintext + "tampered"} {
		_, err := svc.ValidateToken(ctx, bad)
		require.ErrorIs(t, err, ErrInvalidServiceToken, "bad token %q must be rejected", bad)
	}
}

func TestResellerAPITokenService_RevokeInvalidatesToken(t *testing.T) {
	repo := newFakeResellerAPITokenRepo()
	svc := NewResellerAPITokenService(repo)
	ctx := context.Background()

	plaintext, token, err := svc.GenerateToken(ctx, 9, "to-revoke", nil)
	require.NoError(t, err)

	// 他人无法吊销（归属隔离）：用错误的 reseller_id 吊销返回 404，且不影响该令牌。
	err = svc.RevokeToken(ctx, 999, token.ID)
	require.Error(t, err, "non-owner revoke must be rejected")
	_, err = svc.ValidateToken(ctx, plaintext)
	require.NoError(t, err, "token must still be valid after a non-owner revoke attempt")

	// 本人吊销后，令牌立即失效。
	require.NoError(t, svc.RevokeToken(ctx, 9, token.ID))
	_, err = svc.ValidateToken(ctx, plaintext)
	require.ErrorIs(t, err, ErrInvalidServiceToken)
}

func TestResellerAPITokenService_ExpiredTokenRejected(t *testing.T) {
	repo := newFakeResellerAPITokenRepo()
	svc := NewResellerAPITokenService(repo)
	ctx := context.Background()

	ttl := time.Hour
	plaintext, token, err := svc.GenerateToken(ctx, 5, "ttl", &ttl)
	require.NoError(t, err)
	require.NotNil(t, token.ExpiresAt)

	// 校验通过（未过期）。
	_, err = svc.ValidateToken(ctx, plaintext)
	require.NoError(t, err)

	// 把过期时间改到过去 → 校验失败。
	past := time.Now().Add(-time.Minute)
	repo.rows[token.ID].ExpiresAt = &past
	_, err = svc.ValidateToken(ctx, plaintext)
	require.ErrorIs(t, err, ErrInvalidServiceToken)
}
