package service

import (
	"context"
	"log/slog"
	"sort"
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
	// ErrMerchantModeDisabled 当 reseller 未开通代理模式（merchant_mode != "enabled"）却访问佣金/提现能力时返回。
	ErrMerchantModeDisabled = infraerrors.Forbidden("MERCHANT_MODE_DISABLED", "merchant mode is not enabled for this reseller")
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

// RechargeDetailRecord is one row of sub-user recharge history for merchant commission
type RechargeDetailRecord struct {
	UserID       int64     `json:"user_id"`
	UserEmail    string    `json:"user_email,omitempty"`
	OrderNo      string    `json:"order_no"`
	CreditAmount float64   `json:"credit_amount"`
	PaidAt       time.Time `json:"paid_at"`
}

// CommissionSummary is the response for /reseller/commissions/summary
type CommissionSummary struct {
	CommissionRate  float64 `json:"commission_rate"`
	TotalCommission float64 `json:"total_commission"`
	TotalRecharge   float64 `json:"total_recharge"`
	TotalUsers      int64   `json:"total_users"`
	TodayNewUsers   int64   `json:"today_new_users"`
	Withdrawn       float64 `json:"withdrawn"`
	Pending         float64 `json:"pending"`
	Available       float64 `json:"available"`
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
	domainRepo        ResellerDomainRepository
	rechargeOrderRepo RechargeOrderRepository
	orderRepo         OrderRepository
	usdtRepo          UsdtOrderRepository
}

func NewCommissionService(
	withdrawalRepo ResellerWithdrawalRepository,
	usageLogRepo UsageLogRepository,
	userRepo UserRepository,
	settingRepo ResellerSettingRepository,
	domainRepo ResellerDomainRepository,
	rechargeOrderRepo RechargeOrderRepository,
	orderRepo OrderRepository,
	usdtRepo UsdtOrderRepository,
) *CommissionService {
	return &CommissionService{
		withdrawalRepo:    withdrawalRepo,
		usageLogRepo:      usageLogRepo,
		userRepo:          userRepo,
		settingRepo:       settingRepo,
		domainRepo:        domainRepo,
		rechargeOrderRepo: rechargeOrderRepo,
		orderRepo:         orderRepo,
		usdtRepo:          usdtRepo,
	}
}

// getSubUserIDs returns all user IDs with parent_id = resellerID
func (s *CommissionService) getSubUserIDs(ctx context.Context, resellerID int64) ([]int64, error) {
	return s.userRepo.ListIDsByParentID(ctx, resellerID)
}

// requireMerchantMode 校验该 reseller 已开通代理模式（merchant_mode == "enabled"）。
// 未开通则返回 ErrMerchantModeDisabled，作为佣金/提现相关入口的门槛。
func (s *CommissionService) requireMerchantMode(ctx context.Context, resellerID int64) error {
	mode, err := s.settingRepo.Get(ctx, resellerID, "merchant_mode")
	if err != nil {
		return err
	}
	if mode != "enabled" {
		return ErrMerchantModeDisabled
	}
	return nil
}

// getCommissionRate returns the commission_rate for the reseller.
// 未配置/解析失败时默认返回 0（不发佣金）；rate>1 视为误配，返回 0 并记 warn 日志。
// DB 错误必须传播：吞成 0 会让 GetSummary 把 available 算成深度负值。
func (s *CommissionService) getCommissionRate(ctx context.Context, resellerID int64) (float64, error) {
	val, err := s.settingRepo.Get(ctx, resellerID, "commission_rate")
	if err != nil {
		return 0, err
	}
	if val == "" {
		return 0, nil // 未配置：默认不发佣金
	}
	rate, err := strconv.ParseFloat(val, 64)
	if err != nil || rate < 0 {
		return 0, nil
	}
	if rate > 1 {
		// 误配（如把百分比 10 当成 0.1 填成 10），按 0 处理避免超额发放。
		slog.Warn("commission_rate misconfigured (>1), treating as 0",
			"reseller_id", resellerID, "rate", rate)
		return 0, nil
	}
	return rate, nil
}

// GetSummary returns commission summary for a reseller
func (s *CommissionService) GetSummary(ctx context.Context, resellerID int64) (*CommissionSummary, error) {
	// 门槛：仅开通代理模式的 reseller 可访问佣金能力。
	if err := s.requireMerchantMode(ctx, resellerID); err != nil {
		return nil, err
	}

	userIDs, err := s.getSubUserIDs(ctx, resellerID)
	if err != nil {
		return nil, err
	}

	rate, err := s.getCommissionRate(ctx, resellerID)
	if err != nil {
		return nil, err
	}

	// 充值总额：原生充值订单的到账金额（EPAY RechargeOrder + AliMPay Order 两路并行）
	// fail-fast：任一汇总查询出错都直接返回，避免把错误吞成 0 导致 available 算成深度负值。
	var totalRecharge float64
	if s.rechargeOrderRepo != nil && len(userIDs) > 0 {
		nativeTotal, err := s.rechargeOrderRepo.SumPaidCreditByUserIDs(ctx, userIDs)
		if err != nil {
			return nil, err
		}
		totalRecharge += nativeTotal
	}
	if s.orderRepo != nil && len(userIDs) > 0 {
		alimpayTotal, err := s.orderRepo.SumPaidCreditByUserIDs(ctx, userIDs)
		if err != nil {
			return nil, err
		}
		totalRecharge += alimpayTotal
	}
	// USDT 通道：usdt_orders status='paid' 的 credit_amount（credit 即 CNY 等值余额，计佣一致口径）。
	if s.usdtRepo != nil && len(userIDs) > 0 {
		usdtTotal, err := s.usdtRepo.SumPaidCreditByUserIDs(ctx, userIDs)
		if err != nil {
			return nil, err
		}
		totalRecharge += usdtTotal
	}

	// 分成 = 充值总额 × 分成比例
	totalCommission := totalRecharge * rate

	totalUsers := int64(len(userIDs))

	todayNewUsers, err := s.userRepo.CountByParentIDToday(ctx, resellerID)
	if err != nil {
		return nil, err
	}

	withdrawn, err := s.withdrawalRepo.SumPaidByResellerID(ctx, resellerID)
	if err != nil {
		return nil, err
	}

	pending, err := s.withdrawalRepo.SumPendingByResellerID(ctx, resellerID)
	if err != nil {
		return nil, err
	}

	available := totalCommission - withdrawn - pending

	return &CommissionSummary{
		CommissionRate:  rate,
		TotalCommission: totalCommission,
		TotalRecharge:   totalRecharge,
		TotalUsers:      totalUsers,
		TodayNewUsers:   todayNewUsers,
		Withdrawn:       withdrawn,
		Pending:         pending,
		Available:       available,
	}, nil
}

// CreateWithdrawal creates a new withdrawal request
func (s *CommissionService) CreateWithdrawal(ctx context.Context, resellerID int64, amount float64, paymentMethod, paymentAccount, paymentName string) (*ResellerWithdrawal, error) {
	// 门槛：仅开通代理模式的 reseller 可发起提现。
	if err := s.requireMerchantMode(ctx, resellerID); err != nil {
		return nil, err
	}

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
// 注意：这里不做 merchant_mode 门槛——已存在的 pending 提现是商户自己的数据，
// 管理员事后关闭代理模式（或老商户从未写入 merchant_mode key）不应让其
// 看不到/取消不了自己挂起的提现，否则提现被"冻结"而管理端仍可打款。
// 创建提现（CreateWithdrawal）仍受门槛保护。
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
// 只读自有数据，不做 merchant_mode 门槛（理由见 CancelWithdrawal）。
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
	// 门槛：仅开通代理模式的 reseller 可查看充值明细。
	if err := s.requireMerchantMode(ctx, resellerID); err != nil {
		return nil, 0, err
	}

	// Collect all records from both sources, then sort and paginate
	var allRecords []*RechargeDetailRecord

	// Native recharges from recharge_orders table
	if s.rechargeOrderRepo != nil {
		userIDs, err := s.getSubUserIDs(ctx, resellerID)
		if err == nil && len(userIDs) > 0 {
			// 批量查用户邮箱
			emailMap := make(map[int64]string)
			for _, uid := range userIDs {
				if u, err := s.userRepo.GetByID(ctx, uid); err == nil && u != nil {
					emailMap[uid] = u.Email
				}
			}
			if orders, _, err := s.rechargeOrderRepo.ListPaidByUserIDs(ctx, userIDs, 10000, 0); err == nil {
				for _, o := range orders {
					rec := &RechargeDetailRecord{
						UserID:       o.UserID,
						UserEmail:    emailMap[o.UserID],
						OrderNo:      o.OrderNo,
						CreditAmount: o.CreditAmount,
					}
					if o.PaidAt != nil {
						rec.PaidAt = *o.PaidAt
					} else {
						rec.PaidAt = o.CreatedAt
					}
					allRecords = append(allRecords, rec)
				}
			}
		}
	}

	// Native AliMPay orders from orders table
	if s.orderRepo != nil {
		userIDs, err := s.getSubUserIDs(ctx, resellerID)
		if err == nil && len(userIDs) > 0 {
			emailMap := make(map[int64]string)
			for _, uid := range userIDs {
				if u, err := s.userRepo.GetByID(ctx, uid); err == nil && u != nil {
					emailMap[uid] = u.Email
				}
			}
			if orders, _, err := s.orderRepo.ListPaidByUserIDs(ctx, userIDs, 10000, 0); err == nil {
				for _, o := range orders {
					rec := &RechargeDetailRecord{
						UserID:       o.UserID,
						UserEmail:    emailMap[o.UserID],
						OrderNo:      o.OrderNo,
						CreditAmount: o.CreditAmount,
					}
					if o.PaidAt != nil {
						rec.PaidAt = *o.PaidAt
					} else {
						rec.PaidAt = o.CreatedAt
					}
					allRecords = append(allRecords, rec)
				}
			}
		}
	}

	// USDT recharges from usdt_orders table
	if s.usdtRepo != nil {
		userIDs, err := s.getSubUserIDs(ctx, resellerID)
		if err == nil && len(userIDs) > 0 {
			emailMap := make(map[int64]string)
			for _, uid := range userIDs {
				if u, err := s.userRepo.GetByID(ctx, uid); err == nil && u != nil {
					emailMap[uid] = u.Email
				}
			}
			if orders, _, err := s.usdtRepo.ListPaidByUserIDs(ctx, userIDs, 10000, 0); err == nil {
				for _, o := range orders {
					rec := &RechargeDetailRecord{
						UserID:       o.UserID,
						UserEmail:    emailMap[o.UserID],
						OrderNo:      o.OrderNo,
						CreditAmount: o.CreditAmount,
					}
					if o.PaidAt != nil {
						rec.PaidAt = *o.PaidAt
					} else {
						rec.PaidAt = o.CreatedAt
					}
					allRecords = append(allRecords, rec)
				}
			}
		}
	}

	// Sort merged results by PaidAt desc
	sort.Slice(allRecords, func(i, j int) bool {
		return allRecords[i].PaidAt.After(allRecords[j].PaidAt)
	})

	// Apply pagination
	total := len(allRecords)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return allRecords[offset:end], total, nil
}
