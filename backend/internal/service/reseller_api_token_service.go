package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// ResellerServiceTokenPrefix 是分销商 M2M 服务令牌的明文前缀。
const ResellerServiceTokenPrefix = "rst-"

const (
	resellerTokenStatusActive  = "active"
	resellerTokenStatusRevoked = "revoked"
)

// ErrInvalidServiceToken 表示令牌不存在、已吊销或已过期。
// 中间件据此返回 401，不向调用方泄露具体原因。
var ErrInvalidServiceToken = errors.New("invalid reseller service token")

// ResellerAPIToken 是分销商服务令牌的领域模型。
type ResellerAPIToken struct {
	ID          int64
	ResellerID  int64
	Name        string
	TokenPrefix string
	TokenHash   string
	Status      string
	LastUsedAt  *time.Time
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsActive 判断令牌当前是否可用（状态为 active 且未过期）。
func (t *ResellerAPIToken) IsActive(now time.Time) bool {
	if t.Status != resellerTokenStatusActive {
		return false
	}
	if t.ExpiresAt != nil && !t.ExpiresAt.After(now) {
		return false
	}
	return true
}

// ResellerAPITokenRepository 是分销商服务令牌的存储接口。
type ResellerAPITokenRepository interface {
	Create(ctx context.Context, token *ResellerAPIToken) (*ResellerAPIToken, error)
	GetByHash(ctx context.Context, hash string) (*ResellerAPIToken, error)
	GetByIDForReseller(ctx context.Context, id, resellerID int64) (*ResellerAPIToken, error)
	ListByReseller(ctx context.Context, resellerID int64) ([]*ResellerAPIToken, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	TouchLastUsed(ctx context.Context, id int64, at time.Time) error
}

// resellerTokenLastUsedMinTouch 限制同一令牌刷新 last_used_at 的最小间隔，
// 避免高频 M2M 调用造成每请求一次 UPDATE 的写放大。
const resellerTokenLastUsedMinTouch = 30 * time.Second

// ResellerAPITokenService 负责分销商服务令牌的生成、校验与生命周期管理。
type ResellerAPITokenService struct {
	repo ResellerAPITokenRepository
	// lastUsedTouchL1 记录每个 tokenID 下一次允许写 last_used_at 的时间，做轻量去抖。
	lastUsedTouchL1 sync.Map // map[int64]time.Time
}

// NewResellerAPITokenService 创建 ResellerAPITokenService。
func NewResellerAPITokenService(repo ResellerAPITokenRepository) *ResellerAPITokenService {
	return &ResellerAPITokenService{repo: repo}
}

// hashServiceToken 返回令牌明文的 SHA-256 十六进制摘要（小写）。
func hashServiceToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// GenerateToken 为分销商生成一个新的服务令牌。
// 返回的明文仅此一次可见，调用方必须立即返回给用户；库内只存哈希。
// ttl 为 nil 表示永不过期。
func (s *ResellerAPITokenService) GenerateToken(ctx context.Context, resellerID int64, name string, ttl *time.Duration) (plaintext string, token *ResellerAPIToken, err error) {
	name = strings.TrimSpace(name)
	if len(name) > 100 {
		return "", nil, infraerrors.BadRequest("INVALID_NAME", "name must be at most 100 characters")
	}

	// 生成 32 字节随机数 = 64 位十六进制字符，与 GenerateAdminAPIKey 一致。
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", nil, fmt.Errorf("generate random bytes: %w", err)
	}
	plaintext = ResellerServiceTokenPrefix + hex.EncodeToString(buf)

	var expiresAt *time.Time
	if ttl != nil {
		if *ttl <= 0 {
			return "", nil, infraerrors.BadRequest("INVALID_TTL", "ttl must be positive")
		}
		exp := time.Now().Add(*ttl)
		expiresAt = &exp
	}

	model := &ResellerAPIToken{
		ResellerID:  resellerID,
		Name:        name,
		TokenPrefix: serviceTokenDisplayPrefix(plaintext),
		TokenHash:   hashServiceToken(plaintext),
		Status:      resellerTokenStatusActive,
		ExpiresAt:   expiresAt,
	}

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return "", nil, fmt.Errorf("create service token: %w", err)
	}
	return plaintext, created, nil
}

// ValidateToken 校验明文令牌并返回对应记录。
// 令牌不存在 / 已吊销 / 已过期时统一返回 ErrInvalidServiceToken。
func (s *ResellerAPITokenService) ValidateToken(ctx context.Context, raw string) (*ResellerAPIToken, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, ResellerServiceTokenPrefix) {
		return nil, ErrInvalidServiceToken
	}

	token, err := s.repo.GetByHash(ctx, hashServiceToken(raw))
	if err != nil {
		return nil, ErrInvalidServiceToken
	}
	if token == nil || !token.IsActive(time.Now()) {
		return nil, ErrInvalidServiceToken
	}
	return token, nil
}

// ListTokens 返回分销商的全部服务令牌（不含明文，token_hash 字段由上层屏蔽）。
func (s *ResellerAPITokenService) ListTokens(ctx context.Context, resellerID int64) ([]*ResellerAPIToken, error) {
	return s.repo.ListByReseller(ctx, resellerID)
}

// RevokeToken 吊销分销商自己的某个服务令牌。
func (s *ResellerAPITokenService) RevokeToken(ctx context.Context, resellerID, tokenID int64) error {
	token, err := s.repo.GetByIDForReseller(ctx, tokenID, resellerID)
	if err != nil {
		return err
	}
	if token == nil {
		return infraerrors.NotFound("TOKEN_NOT_FOUND", "service token not found")
	}
	if token.Status == resellerTokenStatusRevoked {
		return nil
	}
	return s.repo.UpdateStatus(ctx, tokenID, resellerTokenStatusRevoked)
}

// TouchLastUsedAsync 以"尽力而为"方式异步刷新令牌的最近使用时间，
// 不阻塞请求主路径，失败仅忽略。带 30s 去抖，避免每请求一次 UPDATE 的写放大。
func (s *ResellerAPITokenService) TouchLastUsedAsync(tokenID int64) {
	now := time.Now()
	if v, ok := s.lastUsedTouchL1.Load(tokenID); ok {
		if next, ok2 := v.(time.Time); ok2 && now.Before(next) {
			return // 仍在去抖窗口内，跳过本次（连 goroutine 都不创建）
		}
	}
	s.lastUsedTouchL1.Store(tokenID, now.Add(resellerTokenLastUsedMinTouch))
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.repo.TouchLastUsed(ctx, tokenID, now); err != nil {
			// 写失败时缩短下次允许时间，便于尽快重试。
			s.lastUsedTouchL1.Store(tokenID, time.Now().Add(5*time.Second))
		}
	}()
}

// serviceTokenDisplayPrefix 取明文的前若干位用于 UI 展示（如 rst-1a2b）。
func serviceTokenDisplayPrefix(plaintext string) string {
	const showAfterPrefix = 4
	limit := len(ResellerServiceTokenPrefix) + showAfterPrefix
	if len(plaintext) < limit {
		return plaintext
	}
	return plaintext[:limit]
}
