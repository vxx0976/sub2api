package service

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	Notes         string    `json:"notes,omitempty"`
	PasswordHash  string    `json:"-"` // Never expose password hash
	Role          string    `json:"role"`
	Balance       float64   `json:"balance"`
	Concurrency   int       `json:"concurrency"`
	Status        string    `json:"status"`
	AllowedGroups []int64   `json:"allowed_groups,omitempty"`
	TokenVersion  int64     `json:"-"` // Internal use only
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Referral system fields
	ReferralCode     *string `json:"referral_code,omitempty"`
	ReferredBy       *int64  `json:"-"` // Internal use only
	ReferralRewarded bool    `json:"-"` // Internal use only

	APIKeys       []APIKey           `json:"api_keys,omitempty"`
	Subscriptions []UserSubscription `json:"subscriptions,omitempty"`
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// CanBindGroup checks whether a user can bind to a given group.
// For standard groups:
// - If AllowedGroups is non-empty, only allow binding to IDs in that list.
// - If AllowedGroups is empty (nil or length 0), allow binding to any non-exclusive group.
func (u *User) CanBindGroup(groupID int64, isExclusive bool) bool {
	if len(u.AllowedGroups) > 0 {
		for _, id := range u.AllowedGroups {
			if id == groupID {
				return true
			}
		}
		return false
	}
	return !isExclusive
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
