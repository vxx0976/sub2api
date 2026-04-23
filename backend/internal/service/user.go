package service

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	Notes          string    `json:"notes,omitempty"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	AvatarSource   string    `json:"avatar_source,omitempty"`
	AvatarMIME     string    `json:"avatar_mime,omitempty"`
	AvatarByteSize int       `json:"avatar_byte_size,omitempty"`
	AvatarSHA256   string    `json:"avatar_sha256,omitempty"`
	PasswordHash   string    `json:"-"` // Never expose password hash
	Role           string    `json:"role"`
	Balance        float64   `json:"balance"`
	Concurrency    int       `json:"concurrency"`
	Status         string    `json:"status"`
	AllowedGroups  []int64   `json:"allowed_groups,omitempty"`
	TokenVersion   int64     `json:"-"` // Incremented on password change to invalidate existing tokens
	// TokenVersionResolved indicates TokenVersion already contains the fingerprint-derived
	// value expected in JWT claims and refresh-token state.
	TokenVersionResolved bool       `json:"-"`
	RoleVersion          int64      `json:"-"` // Internal use only — 角色变更时递增
	SignupSource         string     `json:"signup_source,omitempty"`
	LastLoginAt          *time.Time `json:"last_login_at,omitempty"`
	LastActiveAt         *time.Time `json:"last_active_at,omitempty"`
	LastUsedAt           *time.Time `json:"last_used_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Referral system fields（来自 dev：分销/推荐）
	ReferralCode     *string `json:"referral_code,omitempty"`
	ReferredBy       *int64  `json:"-"` // Internal use only
	ReferralRewarded bool    `json:"-"` // Internal use only

	// 注册来源域名（来自 dev）
	RegisterDomain string `json:"register_domain,omitempty"`

	// 分销商层级（来自 dev）
	ParentID *int64 `json:"parent_id,omitempty"` // 上级分销商用户 ID

	// GroupRates 用户专属分组倍率配置
	// map[groupID]rateMultiplier
	GroupRates map[int64]float64

	// TOTP 双因素认证字段
	TotpSecretEncrypted *string    // AES-256-GCM 加密的 TOTP 密钥
	TotpEnabled         bool       // 是否启用 TOTP
	TotpEnabledAt       *time.Time // TOTP 启用时间

	// 余额不足通知（来自 main）
	BalanceNotifyEnabled       bool
	BalanceNotifyThresholdType string // "fixed" (default) | "percentage"
	BalanceNotifyThreshold     *float64
	BalanceNotifyExtraEmails   []NotifyEmailEntry
	TotalRecharged             float64

	// RPMLimit 用户级每分钟请求数上限（0 = 不限制）。仅在所用分组未设置 rpm_limit
	// 且该 (用户, 分组) 无 rpm_override 时作为全局兜底生效，计数键 rpm:u:{userID}:{min}。
	RPMLimit int `json:"rpm_limit,omitempty"`

	// UserGroupRPMOverride 来自 auth cache snapshot 的 (user, group) RPM 覆盖值。
	// nil = 该 API Key 对应的 (user, group) 无 override；非 nil 时 checkRPM 直接使用，
	// 避免每请求查 DB。字段不持久化到数据库。
	UserGroupRPMOverride *int `json:"-"`

	APIKeys       []APIKey           `json:"api_keys,omitempty"`
	Subscriptions []UserSubscription `json:"subscriptions,omitempty"`
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsReseller() bool {
	return u.Role == RoleReseller
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// CanBindGroup checks whether a user can bind to a given group.
// For standard groups:
// - Public groups (non-exclusive): all users can bind
// - Exclusive groups: only users with the group in AllowedGroups can bind
func (u *User) CanBindGroup(groupID int64, isExclusive bool) bool {
	// 公开分组（非专属）：所有用户都可以绑定
	if !isExclusive {
		return true
	}
	// 专属分组：需要在 AllowedGroups 中
	for _, id := range u.AllowedGroups {
		if id == groupID {
			return true
		}
	}
	return false
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}
