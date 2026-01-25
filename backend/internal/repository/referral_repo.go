package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/referralreward"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type referralRepository struct {
	client *dbent.Client
}

// NewReferralRepository creates a new referral repository
func NewReferralRepository(client *dbent.Client) service.ReferralRepository {
	return &referralRepository{client: client}
}

// CreateReward creates a new referral reward record
func (r *referralRepository) CreateReward(ctx context.Context, reward *service.ReferralReward) error {
	builder := r.client.ReferralReward.Create().
		SetReferrerID(reward.ReferrerID).
		SetInviteeID(reward.InviteeID).
		SetTriggerOrderID(reward.TriggerOrderID).
		SetReferrerReward(reward.ReferrerReward).
		SetInviteeReward(reward.InviteeReward)

	if reward.SkipReferrerReason != nil {
		builder = builder.SetSkipReferrerReason(*reward.SkipReferrerReason)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	reward.ID = created.ID
	reward.CreatedAt = created.CreatedAt
	reward.UpdatedAt = created.UpdatedAt
	return nil
}

// GetRewardByInviteeID gets reward record by invitee ID
func (r *referralRepository) GetRewardByInviteeID(ctx context.Context, inviteeID int64) (*service.ReferralReward, error) {
	reward, err := r.client.ReferralReward.Query().
		Where(referralreward.InviteeIDEQ(inviteeID)).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrUserNotFound, nil)
	}
	return referralRewardEntityToService(reward), nil
}

// GetRewardsByReferrerID gets all rewards for a referrer
func (r *referralRepository) GetRewardsByReferrerID(ctx context.Context, referrerID int64, params pagination.PaginationParams) ([]service.ReferralReward, *pagination.PaginationResult, error) {
	query := r.client.ReferralReward.Query().
		Where(referralreward.ReferrerIDEQ(referrerID))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	rewards, err := query.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(referralreward.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	result := make([]service.ReferralReward, len(rewards))
	for i, rw := range rewards {
		result[i] = *referralRewardEntityToService(rw)
	}

	return result, paginationResultFromTotal(int64(total), params), nil
}

// GetReferralStats gets statistics for a referrer
func (r *referralRepository) GetReferralStats(ctx context.Context, referrerID int64) (*service.ReferralStats, error) {
	// Count total invited (users referred by this user)
	totalInvited, err := r.client.User.Query().
		Where(dbuser.ReferredByEQ(referrerID)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	// Count rewarded (users who have triggered rewards)
	totalRewarded, err := r.client.ReferralReward.Query().
		Where(referralreward.ReferrerIDEQ(referrerID)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	// Count pending (invited but not yet rewarded)
	pendingPayment := totalInvited - totalRewarded

	// Calculate total earnings
	var totalEarnings float64
	rewards, err := r.client.ReferralReward.Query().
		Where(referralreward.ReferrerIDEQ(referrerID)).
		Select(referralreward.FieldReferrerReward).
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, rw := range rewards {
		totalEarnings += rw.ReferrerReward
	}

	return &service.ReferralStats{
		TotalInvited:   totalInvited,
		TotalRewarded:  totalRewarded,
		PendingPayment: pendingPayment,
		TotalEarnings:  totalEarnings,
	}, nil
}

// GetInvitees gets all invitees for a referrer
func (r *referralRepository) GetInvitees(ctx context.Context, referrerID int64, params pagination.PaginationParams) ([]service.InviteeInfo, *pagination.PaginationResult, error) {
	query := r.client.User.Query().
		Where(dbuser.ReferredByEQ(referrerID))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	users, err := query.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(dbuser.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get reward info for these users
	userIDs := make([]int64, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	rewardMap := make(map[int64]*dbent.ReferralReward)
	if len(userIDs) > 0 {
		rewards, err := r.client.ReferralReward.Query().
			Where(referralreward.InviteeIDIn(userIDs...)).
			All(ctx)
		if err != nil {
			return nil, nil, err
		}
		for _, rw := range rewards {
			rewardMap[rw.InviteeID] = rw
		}
	}

	result := make([]service.InviteeInfo, len(users))
	for i, u := range users {
		info := service.InviteeInfo{
			ID:       u.ID,
			Email:    service.MaskEmail(u.Email),
			Status:   "pending",
			JoinedAt: u.CreatedAt,
		}

		if reward, ok := rewardMap[u.ID]; ok {
			info.Status = "rewarded"
			info.RewardedAt = &reward.CreatedAt
			info.RewardAmount = &reward.InviteeReward
			info.ReferrerEarning = &reward.ReferrerReward
		}

		result[i] = info
	}

	return result, paginationResultFromTotal(int64(total), params), nil
}

// GetUserByReferralCode gets user by referral code
func (r *referralRepository) GetUserByReferralCode(ctx context.Context, code string) (*service.User, error) {
	user, err := r.client.User.Query().
		Where(dbuser.ReferralCodeEQ(code)).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrUserNotFound, nil)
	}
	return userEntityToService(user), nil
}

// UpdateUserReferralRewarded marks user as having received referral reward
func (r *referralRepository) UpdateUserReferralRewarded(ctx context.Context, userID int64) error {
	_, err := r.client.User.UpdateOneID(userID).
		SetReferralRewarded(true).
		Save(ctx)
	return err
}

// referralRewardEntityToService converts an ent ReferralReward to service ReferralReward
func referralRewardEntityToService(e *dbent.ReferralReward) *service.ReferralReward {
	if e == nil {
		return nil
	}
	return &service.ReferralReward{
		ID:                 e.ID,
		ReferrerID:         e.ReferrerID,
		InviteeID:          e.InviteeID,
		TriggerOrderID:     e.TriggerOrderID,
		ReferrerReward:     e.ReferrerReward,
		InviteeReward:      e.InviteeReward,
		SkipReferrerReason: e.SkipReferrerReason,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
}
