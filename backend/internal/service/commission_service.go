package service

import (
	"context"
	"strconv"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrWithdrawalNotFound      = infraerrors.NotFound("WITHDRAWAL_NOT_FOUND", "withdrawal not found")
	ErrWithdrawalPending       = infraerrors.Conflict("WITHDRAWAL_PENDING", "you already have a pending withdrawal")
	ErrWithdrawalAmountTooLow  = infraerrors.BadRequest("WITHDRAWAL_AMOUNT_TOO_LOW", "withdrawal amount is below minimum")
	ErrWithdrawalAmountTooHigh = infraerrors.BadRequest("WITHDRAWAL_AMOUNT_TOO_HIGH", "withdrawal amount exceeds available balance")
	ErrWithdrawalExceedsMax    = infraerrors.BadRequest("WITHDRAWAL_EXCEEDS_MAX", "withdrawal amount exceeds maximum allowed")
	ErrWithdrawalNotPending    = infraerrors.BadRequest("WITHDRAWAL_NOT_PENDING", "only pending withdrawals can be cancelled")
	ErrWithdrawalOwnership     = infraerrors.Forbidden("WITHDRAWAL_OWNERSHIP", "you do not own this withdrawal")
	ErrInvalidWithdrawalStatus = infraerrors.BadRequest("INVALID_WITHDRAWAL_STATUS", "invalid withdrawal status, must be pending, paid, or rejected")
)

// validWithdrawalStatuses 提现状态白名单
var validWithdrawalStatuses = map[string]struct{}{
	"pending":  {},
	"paid":     {},
	"rejected": {},
}

// isValidWithdrawalStatus 校验提现状态是否合法
func isValidWithdrawalStatus(status string) bool {
	if status == "" {
		return true // 空字符串表示不过滤
	}
	_, ok := validWithdrawalStatuses[status]
	return ok
}

// ResellerWithdrawal represents a withdrawal request
type ResellerWithdrawal struct {
	ID             int64      `json:"id"`
	ResellerID     int64      `json:"reseller_id"`
	Amount         float64    `json:"amount"`
	Status         string     `json:"status"`
	PaymentMethod  string     `json:"payment_method"`
	PaymentAccount string     `json:"payment_account"`
	PaymentName    string     `json:"payment_name"`
	AdminNotes     string     `json:"admin_notes"`
	AdminID        *int64     `json:"admin_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	RejectedAt     *time.Time `json:"rejected_at,omitempty"`
}

// CommissionSummary is the response for /reseller/commissions/summary
type CommissionSummary struct {
	CommissionRate  float64 `json:"commission_rate"`
	TotalCost       float64 `json:"total_cost"`
	TotalCommission float64 `json:"total_commission"`
	TotalRecharge   float64 `json:"total_recharge"`
	TotalUsers      int64   `json:"total_users"`
	TodayNewUsers   int64   `json:"today_new_users"`
	TodayCost       float64 `json:"today_cost"`
	Withdrawn       float64 `json:"withdrawn"`
	Pending         float64 `json:"pending"`
	Available       float64 `json:"available"`
}

// CommissionDetailItem is one row in /reseller/commissions/detail
type CommissionDetailItem struct {
	UserID               int64     `json:"user_id"`
	Model                string    `json:"model"`
	TotalCost            float64   `json:"total_cost"`
	MerchantRateSnapshot *float64  `json:"merchant_rate_snapshot"`
	PlatformCostSnapshot *float64  `json:"platform_cost_snapshot"`
	Commission           float64   `json:"commission"`
	CreatedAt            time.Time `json:"created_at"`
}

// MerchantInfo is information about a reseller user
type MerchantInfo struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Balance   float64 `json:"balance"`
	Domain    string  `json:"domain"`
	UserCount int64   `json:"user_count"`
	KeyCount  int64   `json:"key_count"`
}

// ResellerWithdrawalRepository interface
type ResellerWithdrawalRepository interface {
	// CreateIfNoPending 原子性地检查是否有待审核提现，若无则创建新提现。防止并发竞态。
	CreateIfNoPending(ctx context.Context, w *ResellerWithdrawal) error
	GetByID(ctx context.Context, id int64) (*ResellerWithdrawal, error)
	List(ctx context.Context, resellerID int64, status string, limit, offset int) ([]*ResellerWithdrawal, int, error)
	ListAll(ctx context.Context, status string, resellerID int64, limit, offset int) ([]*ResellerWithdrawal, int, error)
	SumPaidByResellerID(ctx context.Context, resellerID int64) (float64, error)
	SumPendingByResellerID(ctx context.Context, resellerID int64) (float64, error)
	UpdateStatus(ctx context.Context, id int64, status string, adminID int64, adminNotes string) error
	Delete(ctx context.Context, id int64) error
}

// CommissionService handles commission calculation and withdrawal operations
type CommissionService struct {
	withdrawalRepo    ResellerWithdrawalRepository
	usageLogRepo      UsageLogRepository
	userRepo          UserRepository
	settingRepo       ResellerSettingRepository
	rechargeOrderRepo RechargeOrderRepository
	domainRepo        ResellerDomainRepository
	sub2apipayService *Sub2apipayService
}

func NewCommissionService(
	withdrawalRepo ResellerWithdrawalRepository,
	usageLogRepo UsageLogRepository,
	userRepo UserRepository,
	settingRepo ResellerSettingRepository,
	rechargeOrderRepo RechargeOrderRepository,
	domainRepo ResellerDomainRepository,
	sub2apipayService *Sub2apipayService,
) *CommissionService {
	return &CommissionService{
		withdrawalRepo:    withdrawalRepo,
		usageLogRepo:      usageLogRepo,
		userRepo:          userRepo,
		settingRepo:       settingRepo,
		rechargeOrderRepo: rechargeOrderRepo,
		domainRepo:        domainRepo,
		sub2apipayService: sub2apipayService,
	}
}

// getSubUserIDs returns all user IDs with parent_id = resellerID
func (s *CommissionService) getSubUserIDs(ctx context.Context, resellerID int64) ([]int64, error) {
	return s.userRepo.ListIDsByParentID(ctx, resellerID)
}

// getCommissionRate returns the commission_rate for the reseller (default 0.1 = 10%)
func (s *CommissionService) getCommissionRate(ctx context.Context, resellerID int64) (float64, error) {
	val, err := s.settingRepo.Get(ctx, resellerID, "commission_rate")
	if err != nil || val == "" {
		return 0.1, nil // 默认10%
	}
	rate, err := strconv.ParseFloat(val, 64)
	if err != nil || rate < 0 {
		return 0.1, nil
	}
	return rate, nil
}

// GetSummary returns commission summary for a reseller
func (s *CommissionService) GetSummary(ctx context.Context, resellerID int64) (*CommissionSummary, error) {
	userIDs, err := s.getSubUserIDs(ctx, resellerID)
	if err != nil {
		return nil, err
	}

	rate, err := s.getCommissionRate(ctx, resellerID)
	if err != nil {
		return nil, err
	}

	var totalCost float64
	if len(userIDs) > 0 {
		totalCost, err = s.usageLogRepo.SumCommissionByUserIDs(ctx, userIDs)
		if err != nil {
			return nil, err
		}
	}

	// 分成 = 用户总消费 × 分成比例
	totalCommission := totalCost * rate

	// 查询充值总额：通过 sub2apipay HTTP API 按商户域名汇总
	var totalRecharge float64
	if s.sub2apipayService != nil {
		domains, _ := s.domainRepo.ListAllDomainNamesByResellerID(ctx, resellerID)
		if len(domains) > 0 {
			// sub2apipay srcHost 带 https:// 前缀
			hosts := make([]string, len(domains))
			for i, d := range domains {
				hosts[i] = "https://" + d
			}
			if totals, err := s.sub2apipayService.SumRechargeByHosts(ctx, hosts); err == nil {
				for _, v := range totals {
					totalRecharge += v
				}
			}
		}
	}

	totalUsers := int64(len(userIDs))

	todayNewUsers, err := s.userRepo.CountByParentIDToday(ctx, resellerID)
	if err != nil {
		return nil, err
	}

	var todayCost float64
	if len(userIDs) > 0 {
		todayCost, err = s.usageLogRepo.SumTodayCostByUserIDs(ctx, userIDs)
		if err != nil {
			return nil, err
		}
	}

	withdrawn, err := s.withdrawalRepo.SumPaidByResellerID(ctx, resellerID)
	if err != nil {
		return nil, err
	}

	// 待审核提现也应从可用余额中扣除，避免重复申请
	pending, err := s.withdrawalRepo.SumPendingByResellerID(ctx, resellerID)
	if err != nil {
		return nil, err
	}

	available := totalCommission - withdrawn - pending
	if available < 0 {
		available = 0
	}

	return &CommissionSummary{
		CommissionRate:  rate,
		TotalCost:       totalCost,
		TotalCommission: totalCommission,
		TotalRecharge:   totalRecharge,
		TotalUsers:      totalUsers,
		TodayNewUsers:   todayNewUsers,
		TodayCost:       todayCost,
		Withdrawn:       withdrawn,
		Pending:         pending,
		Available:       available,
	}, nil
}

// GetDetail returns paginated commission detail for reseller's sub-users
func (s *CommissionService) GetDetail(ctx context.Context, resellerID int64, startDate, endDate *time.Time, userIDFilter *int64, limit, offset int) ([]*CommissionDetailItem, int, error) {
	userIDs, err := s.getSubUserIDs(ctx, resellerID)
	if err != nil {
		return nil, 0, err
	}
	if len(userIDs) == 0 {
		return nil, 0, nil
	}
	return s.usageLogRepo.ListCommissionDetail(ctx, userIDs, startDate, endDate, userIDFilter, limit, offset)
}

// CreateWithdrawal creates a new withdrawal request
func (s *CommissionService) CreateWithdrawal(ctx context.Context, resellerID int64, amount float64, paymentMethod, paymentAccount, paymentName string) (*ResellerWithdrawal, error) {
	// 校验支付方式白名单
	validMethods := map[string]struct{}{"alipay": {}, "wechat": {}, "bank": {}}
	if _, ok := validMethods[paymentMethod]; !ok {
		return nil, infraerrors.BadRequest("INVALID_PAYMENT_METHOD", "invalid payment method")
	}

	// 校验最低提现金额
	minVal, err := s.settingRepo.Get(ctx, resellerID, "min_withdrawal")
	if err == nil && minVal != "" {
		if minAmount, parseErr := strconv.ParseFloat(minVal, 64); parseErr == nil && amount < minAmount {
			return nil, ErrWithdrawalAmountTooLow
		}
	}

	// 校验最高提现金额（可选配置）
	maxVal, err := s.settingRepo.Get(ctx, resellerID, "max_withdrawal")
	if err == nil && maxVal != "" {
		if maxAmount, parseErr := strconv.ParseFloat(maxVal, 64); parseErr == nil && maxAmount > 0 && amount > maxAmount {
			return nil, ErrWithdrawalExceedsMax
		}
	}

	// 校验可用余额（含待审核金额）
	summary, err := s.GetSummary(ctx, resellerID)
	if err != nil {
		return nil, err
	}
	if amount > summary.Available {
		return nil, ErrWithdrawalAmountTooHigh
	}

	w := &ResellerWithdrawal{
		ResellerID:     resellerID,
		Amount:         amount,
		PaymentMethod:  paymentMethod,
		PaymentAccount: paymentAccount,
		PaymentName:    paymentName,
	}
	// CreateIfNoPending 在事务中原子性地检查并创建，防止并发重复提交
	if err := s.withdrawalRepo.CreateIfNoPending(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

// CancelWithdrawal cancels a pending withdrawal (reseller-initiated)
func (s *CommissionService) CancelWithdrawal(ctx context.Context, resellerID, withdrawalID int64) error {
	w, err := s.withdrawalRepo.GetByID(ctx, withdrawalID)
	if err != nil {
		return err
	}
	if w.ResellerID != resellerID {
		return ErrWithdrawalOwnership
	}
	if w.Status != "pending" {
		return ErrWithdrawalNotPending
	}
	return s.withdrawalRepo.Delete(ctx, withdrawalID)
}

// ListWithdrawals lists withdrawals for a reseller
func (s *CommissionService) ListWithdrawals(ctx context.Context, resellerID int64, status string, limit, offset int) ([]*ResellerWithdrawal, int, error) {
	if !isValidWithdrawalStatus(status) {
		return nil, 0, ErrInvalidWithdrawalStatus
	}
	return s.withdrawalRepo.List(ctx, resellerID, status, limit, offset)
}

// AdminListWithdrawals lists all withdrawals (admin view)
func (s *CommissionService) AdminListWithdrawals(ctx context.Context, status string, resellerID int64, limit, offset int) ([]*ResellerWithdrawal, int, error) {
	if !isValidWithdrawalStatus(status) {
		return nil, 0, ErrInvalidWithdrawalStatus
	}
	return s.withdrawalRepo.ListAll(ctx, status, resellerID, limit, offset)
}

// AdminPayWithdrawal marks a withdrawal as paid
func (s *CommissionService) AdminPayWithdrawal(ctx context.Context, adminID, withdrawalID int64, adminNotes string) error {
	w, err := s.withdrawalRepo.GetByID(ctx, withdrawalID)
	if err != nil {
		return err
	}
	if w.Status != "pending" {
		return ErrWithdrawalNotPending
	}
	return s.withdrawalRepo.UpdateStatus(ctx, withdrawalID, "paid", adminID, adminNotes)
}

// AdminRejectWithdrawal marks a withdrawal as rejected
func (s *CommissionService) AdminRejectWithdrawal(ctx context.Context, adminID, withdrawalID int64, adminNotes string) error {
	w, err := s.withdrawalRepo.GetByID(ctx, withdrawalID)
	if err != nil {
		return err
	}
	if w.Status != "pending" {
		return ErrWithdrawalNotPending
	}
	return s.withdrawalRepo.UpdateStatus(ctx, withdrawalID, "rejected", adminID, adminNotes)
}

// AdminGetMerchants returns paginated list of merchants (role=reseller) with info
func (s *CommissionService) AdminGetMerchants(ctx context.Context, page, pageSize int, search string) ([]*MerchantInfo, int, error) {
	return s.userRepo.ListResellerUsers(ctx, page, pageSize, search)
}

// AdminGetMerchantSettings returns all settings for a merchant
func (s *CommissionService) AdminGetMerchantSettings(ctx context.Context, merchantID int64) (map[string]string, error) {
	return s.settingRepo.GetAll(ctx, merchantID)
}

// AdminUpdateMerchantSettings updates merchant mode and pricing settings
func (s *CommissionService) AdminUpdateMerchantSettings(ctx context.Context, merchantID int64, settings map[string]string) error {
	return s.settingRepo.SetAll(ctx, merchantID, settings)
}

// BackfillMerchantRateSnapshot fills NULL merchant_rate_snapshot values in usage_logs
// using the current price_multiplier from reseller_settings. Returns updated row count.
func (s *CommissionService) BackfillMerchantRateSnapshot(ctx context.Context) (int64, error) {
	return s.usageLogRepo.BackfillMerchantRateSnapshot(ctx)
}

// ListRechargeDetail returns paginated recharge history for a reseller's sub-users
func (s *CommissionService) ListRechargeDetail(ctx context.Context, resellerID int64, limit, offset int) ([]*RechargeDetailRecord, int, error) {
	userIDs, err := s.getSubUserIDs(ctx, resellerID)
	if err != nil {
		return nil, 0, err
	}
	if len(userIDs) == 0 {
		return nil, 0, nil
	}
	return s.rechargeOrderRepo.ListPaidByUserIDs(ctx, userIDs, limit, offset)
}
