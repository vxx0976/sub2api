package repository

import (
	"context"
	"log"

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

// GetAllReferralRecords gets all referral records for admin with user emails
// This includes both rewarded and pending referrals
func (r *referralRepository) GetAllReferralRecords(ctx context.Context, params pagination.PaginationParams, search string) ([]service.AdminReferralRecord, *pagination.PaginationResult, error) {
	// Query all users who were referred (have referred_by set)
	query := r.client.User.Query().
		Where(dbuser.ReferredByNotNil())

	// Apply search filter if provided
	if search != "" {
		query = query.Where(dbuser.EmailContainsFold(search))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		log.Printf("[AdminReferral] Count error: %v", err)
		return nil, nil, err
	}

	log.Printf("[AdminReferral] Total users with referred_by: %d, page: %d, pageSize: %d", total, params.Page, params.PageSize)

	invitees, err := query.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(dbuser.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		log.Printf("[AdminReferral] Query error: %v", err)
		return nil, nil, err
	}

	log.Printf("[AdminReferral] Found %d invitees in current page", len(invitees))

	if len(invitees) == 0 {
		return []service.AdminReferralRecord{}, paginationResultFromTotal(int64(total), params), nil
	}

	// Collect invitee IDs and referrer IDs
	inviteeIDs := make([]int64, len(invitees))
	referrerIDSet := make(map[int64]bool)
	for i, u := range invitees {
		inviteeIDs[i] = u.ID
		if u.ReferredBy != nil {
			referrerIDSet[*u.ReferredBy] = true
		}
	}

	// Fetch reward records for these invitees
	rewardMap := make(map[int64]*dbent.ReferralReward)
	rewards, err := r.client.ReferralReward.Query().
		Where(referralreward.InviteeIDIn(inviteeIDs...)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, rw := range rewards {
		rewardMap[rw.InviteeID] = rw
	}

	// Fetch referrer emails
	referrerIDs := make([]int64, 0, len(referrerIDSet))
	for id := range referrerIDSet {
		referrerIDs = append(referrerIDs, id)
	}

	referrerEmailMap := make(map[int64]string)
	if len(referrerIDs) > 0 {
		referrers, err := r.client.User.Query().
			Where(dbuser.IDIn(referrerIDs...)).
			Select(dbuser.FieldID, dbuser.FieldEmail).
			All(ctx)
		if err != nil {
			return nil, nil, err
		}
		for _, u := range referrers {
			referrerEmailMap[u.ID] = u.Email
		}
	}

	// Build result
	result := make([]service.AdminReferralRecord, len(invitees))
	for i, invitee := range invitees {
		referrerID := int64(0)
		if invitee.ReferredBy != nil {
			referrerID = *invitee.ReferredBy
		}

		record := service.AdminReferralRecord{
			ReferrerID:    referrerID,
			ReferrerEmail: referrerEmailMap[referrerID],
			InviteeID:     invitee.ID,
			InviteeEmail:  invitee.Email,
			CreatedAt:     invitee.CreatedAt,
			Status:        "pending",
		}

		// If there's a reward record, use its data
		if rw, ok := rewardMap[invitee.ID]; ok {
			record.ID = rw.ID
			record.TriggerOrderID = rw.TriggerOrderID
			record.ReferrerReward = rw.ReferrerReward
			record.InviteeReward = rw.InviteeReward
			record.SkipReferrerReason = rw.SkipReferrerReason
			record.CreatedAt = rw.CreatedAt
			record.Status = "rewarded"
		}

		result[i] = record
	}

	return result, paginationResultFromTotal(int64(total), params), nil
}

// GetAdminReferralStats gets overall referral statistics for admin
func (r *referralRepository) GetAdminReferralStats(ctx context.Context) (*service.AdminReferralStats, error) {
	// Count total invitees (users with referred_by set)
	totalInvitees, err := r.client.User.Query().
		Where(dbuser.ReferredByNotNil()).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	// Count rewarded invitees
	totalRewarded, err := r.client.ReferralReward.Query().Count(ctx)
	if err != nil {
		return nil, err
	}

	// Get unique referrers from users table
	users, err := r.client.User.Query().
		Where(dbuser.ReferredByNotNil()).
		Select(dbuser.FieldReferredBy).
		All(ctx)
	if err != nil {
		return nil, err
	}

	referrerSet := make(map[int64]bool)
	for _, u := range users {
		if u.ReferredBy != nil {
			referrerSet[*u.ReferredBy] = true
		}
	}

	// Calculate total paid amounts
	rewards, err := r.client.ReferralReward.Query().
		Select(referralreward.FieldReferrerReward, referralreward.FieldInviteeReward).
		All(ctx)
	if err != nil {
		return nil, err
	}

	var totalReferrerPaid, totalInviteePaid float64
	for _, rw := range rewards {
		totalReferrerPaid += rw.ReferrerReward
		totalInviteePaid += rw.InviteeReward
	}

	return &service.AdminReferralStats{
		TotalRecords:      totalInvitees,
		TotalReferrers:    len(referrerSet),
		TotalInvitees:     totalInvitees,
		TotalPending:      totalInvitees - totalRewarded,
		TotalRewarded:     totalRewarded,
		TotalReferrerPaid: totalReferrerPaid,
		TotalInviteePaid:  totalInviteePaid,
	}, nil
}
