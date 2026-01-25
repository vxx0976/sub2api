package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrReferralCodeNotFound    = infraerrors.NotFound("REFERRAL_CODE_NOT_FOUND", "referral code not found")
	ErrReferralCodeInvalid     = infraerrors.BadRequest("REFERRAL_CODE_INVALID", "invalid referral code format")
	ErrReferralSelfReferral    = infraerrors.BadRequest("REFERRAL_SELF_REFERRAL", "cannot use your own referral code")
	ErrReferralAlreadyRewarded = infraerrors.BadRequest("REFERRAL_ALREADY_REWARDED", "referral reward already distributed")
)

// ReferralReward represents a referral reward record
type ReferralReward struct {
	ID                 int64     `json:"id"`
	ReferrerID         int64     `json:"referrer_id"`
	InviteeID          int64     `json:"invitee_id"`
	TriggerOrderID     int64     `json:"trigger_order_id"`
	ReferrerReward     float64   `json:"referrer_reward"`
	InviteeReward      float64   `json:"invitee_reward"`
	SkipReferrerReason *string   `json:"skip_referrer_reason,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ReferralStats represents the referral statistics for a user
type ReferralStats struct {
	TotalInvited   int     `json:"total_invited"`
	TotalRewarded  int     `json:"total_rewarded"`
	PendingPayment int     `json:"pending_payment"`
	TotalEarnings  float64 `json:"total_earnings"`
}

// InviteeInfo represents information about an invitee
type InviteeInfo struct {
	ID              int64      `json:"id"`
	Email           string     `json:"email"` // Will be masked
	Status          string     `json:"status"`
	JoinedAt        time.Time  `json:"joined_at"`
	RewardedAt      *time.Time `json:"rewarded_at,omitempty"`
	RewardAmount    *float64   `json:"reward_amount,omitempty"`
	ReferrerEarning *float64   `json:"referrer_earning,omitempty"`
}

// ReferralRepository defines the interface for referral data access
type ReferralRepository interface {
	// CreateReward creates a new referral reward record
	CreateReward(ctx context.Context, reward *ReferralReward) error
	// GetRewardByInviteeID gets reward record by invitee ID
	GetRewardByInviteeID(ctx context.Context, inviteeID int64) (*ReferralReward, error)
	// GetRewardsByReferrerID gets all rewards for a referrer
	GetRewardsByReferrerID(ctx context.Context, referrerID int64, params pagination.PaginationParams) ([]ReferralReward, *pagination.PaginationResult, error)
	// GetReferralStats gets statistics for a referrer
	GetReferralStats(ctx context.Context, referrerID int64) (*ReferralStats, error)
	// GetInvitees gets all invitees for a referrer
	GetInvitees(ctx context.Context, referrerID int64, params pagination.PaginationParams) ([]InviteeInfo, *pagination.PaginationResult, error)
	// GetUserByReferralCode gets user by referral code
	GetUserByReferralCode(ctx context.Context, code string) (*User, error)
	// UpdateUserReferralRewarded marks user as having received referral reward
	UpdateUserReferralRewarded(ctx context.Context, userID int64) error
}

// ReferralService handles referral-related business logic
type ReferralService struct {
	referralRepo ReferralRepository
	userRepo     UserRepository
}

// NewReferralService creates a new ReferralService
func NewReferralService(referralRepo ReferralRepository, userRepo UserRepository) *ReferralService {
	return &ReferralService{
		referralRepo: referralRepo,
		userRepo:     userRepo,
	}
}

// GenerateReferralCode generates a unique referral code for a user
func (s *ReferralService) GenerateReferralCode() (string, error) {
	// Generate random bytes for the code (7 characters after 'R' prefix)
	bytes := make([]byte, 4) // 4 bytes = 8 hex chars, we'll use 7
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	code := ReferralCodePrefix + strings.ToUpper(hex.EncodeToString(bytes))[:ReferralCodeLength-1]
	return code, nil
}

// IsReferralCode checks if a code is a referral code (starts with 'R')
func (s *ReferralService) IsReferralCode(code string) bool {
	code = strings.ToUpper(strings.TrimSpace(code))
	return len(code) == ReferralCodeLength && strings.HasPrefix(code, ReferralCodePrefix)
}

// ValidateReferralCode validates a referral code and returns the referrer info
func (s *ReferralService) ValidateReferralCode(ctx context.Context, code string, currentUserID int64) (*User, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	// Check format
	if !s.IsReferralCode(code) {
		return nil, ErrReferralCodeInvalid
	}

	// Get user by referral code
	referrer, err := s.referralRepo.GetUserByReferralCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrReferralCodeNotFound
		}
		return nil, err
	}

	// Check if user is trying to use their own code
	if currentUserID > 0 && referrer.ID == currentUserID {
		return nil, ErrReferralSelfReferral
	}

	return referrer, nil
}

// ValidateReferralCodePublic validates a referral code without checking self-referral
// Used for public validation before registration
func (s *ReferralService) ValidateReferralCodePublic(ctx context.Context, code string) (*User, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	// Check format
	if !s.IsReferralCode(code) {
		return nil, ErrReferralCodeInvalid
	}

	// Get user by referral code
	referrer, err := s.referralRepo.GetUserByReferralCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrReferralCodeNotFound
		}
		return nil, err
	}

	return referrer, nil
}

// GetUserReferralCode gets or generates a referral code for a user
func (s *ReferralService) GetUserReferralCode(ctx context.Context, userID int64) (string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	// If user already has a referral code, return it
	if user.ReferralCode != nil && *user.ReferralCode != "" {
		return *user.ReferralCode, nil
	}

	// Generate a new code
	code, err := s.GenerateReferralCode()
	if err != nil {
		return "", err
	}

	// Update user with the new code
	user.ReferralCode = &code
	if err := s.userRepo.Update(ctx, user); err != nil {
		// If update fails due to duplicate, try again with a new code
		log.Printf("[Referral] Failed to save referral code for user %d: %v, retrying...", userID, err)
		code, err = s.GenerateReferralCode()
		if err != nil {
			return "", err
		}
		user.ReferralCode = &code
		if err := s.userRepo.Update(ctx, user); err != nil {
			return "", err
		}
	}

	return code, nil
}

// GetReferralStats gets referral statistics for a user
func (s *ReferralService) GetReferralStats(ctx context.Context, userID int64) (*ReferralStats, error) {
	return s.referralRepo.GetReferralStats(ctx, userID)
}

// GetInvitees gets the list of invitees for a user
func (s *ReferralService) GetInvitees(ctx context.Context, userID int64, page, pageSize int) ([]InviteeInfo, *pagination.PaginationResult, error) {
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	return s.referralRepo.GetInvitees(ctx, userID, params)
}

// ProcessReferralReward processes referral reward when an invitee makes their first qualifying payment
// Returns (referrerRewarded, inviteeRewarded, error)
func (s *ReferralService) ProcessReferralReward(ctx context.Context, inviteeID int64, orderID int64, orderAmountCNY float64) (bool, bool, error) {
	// Check if order amount meets minimum threshold
	if orderAmountCNY < ReferralMinPaymentAmountCNY {
		log.Printf("[Referral] Order amount %.2f CNY is below minimum %.2f CNY, skipping reward for user %d",
			orderAmountCNY, ReferralMinPaymentAmountCNY, inviteeID)
		return false, false, nil
	}

	// Get invitee info
	invitee, err := s.userRepo.GetByID(ctx, inviteeID)
	if err != nil {
		return false, false, err
	}

	// Check if invitee was referred by someone
	if invitee.ReferredBy == nil {
		log.Printf("[Referral] User %d was not referred by anyone, skipping reward", inviteeID)
		return false, false, nil
	}

	// Check if reward has already been distributed
	if invitee.ReferralRewarded {
		log.Printf("[Referral] Reward already distributed for user %d", inviteeID)
		return false, false, nil
	}

	referrerID := *invitee.ReferredBy

	// Get referrer info
	referrer, err := s.userRepo.GetByID(ctx, referrerID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Printf("[Referral] Referrer %d not found, giving reward only to invitee %d", referrerID, inviteeID)
			// Referrer not found, give reward only to invitee
			return s.distributeReward(ctx, referrerID, inviteeID, orderID, false, "referrer_not_found")
		}
		return false, false, err
	}

	// Check if referrer is banned/disabled
	if referrer.Status != StatusActive {
		log.Printf("[Referral] Referrer %d is not active (status: %s), giving reward only to invitee %d",
			referrerID, referrer.Status, inviteeID)
		return s.distributeReward(ctx, referrerID, inviteeID, orderID, false, "referrer_"+referrer.Status)
	}

	// Both get rewards
	return s.distributeReward(ctx, referrerID, inviteeID, orderID, true, "")
}

// distributeReward distributes referral rewards
func (s *ReferralService) distributeReward(ctx context.Context, referrerID, inviteeID, orderID int64, giveToReferrer bool, skipReason string) (bool, bool, error) {
	reward := &ReferralReward{
		ReferrerID:     referrerID,
		InviteeID:      inviteeID,
		TriggerOrderID: orderID,
		InviteeReward:  ReferralRewardAmountUSD,
	}

	if giveToReferrer {
		reward.ReferrerReward = ReferralRewardAmountUSD
	} else if skipReason != "" {
		reward.SkipReferrerReason = &skipReason
	}

	// Create reward record
	if err := s.referralRepo.CreateReward(ctx, reward); err != nil {
		return false, false, err
	}

	// Update invitee balance
	if err := s.userRepo.UpdateBalance(ctx, inviteeID, ReferralRewardAmountUSD); err != nil {
		log.Printf("[Referral] Failed to update invitee %d balance: %v", inviteeID, err)
		// Continue even if balance update fails, the reward record is created
	}

	// Update referrer balance if applicable
	if giveToReferrer {
		if err := s.userRepo.UpdateBalance(ctx, referrerID, ReferralRewardAmountUSD); err != nil {
			log.Printf("[Referral] Failed to update referrer %d balance: %v", referrerID, err)
			// Continue even if balance update fails
		}
	}

	// Mark invitee as rewarded
	if err := s.referralRepo.UpdateUserReferralRewarded(ctx, inviteeID); err != nil {
		log.Printf("[Referral] Failed to mark user %d as rewarded: %v", inviteeID, err)
	}

	log.Printf("[Referral] Reward distributed: referrer=%d (%.2f USD), invitee=%d (%.2f USD), order=%d",
		referrerID, reward.ReferrerReward, inviteeID, reward.InviteeReward, orderID)

	return giveToReferrer, true, nil
}

// MaskEmail masks an email address for privacy
// e.g., "user@example.com" -> "u***@example.com"
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 1 {
		return localPart + "***@" + domain
	}

	return string(localPart[0]) + "***@" + domain
}
