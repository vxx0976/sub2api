package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/payment"
	"github.com/google/uuid"
)

// UsdtOrderService USDT(TRC20) 充值业务服务。
//
// 与 OrderService(AliMPay) 平级：业务都是「CNY 计价 → 平台余额入账」，
// 差异在支付通道——USDT 走链上收款，无 webhook，由 UsdtMonitor 轮询 TronGrid
// 按唯一金额匹配后回调 ConfirmUsdtOrderPaid。
// 实现 payment.UsdtOrderMatcher 接口。
type UsdtOrderService struct {
	usdtRepo       UsdtOrderRepository
	settingRepo    SettingRepository
	adminService   AdminService
	settingService *SettingService
	usdt           *payment.UsdtPayment
	expireInterval time.Duration

	stopCh    chan struct{}
	stopOnce  sync.Once
	startOnce sync.Once
	wg        sync.WaitGroup

	lockCache  LeaderLockCache
	db         *sql.DB
	instanceID string
}

// NewUsdtOrderService 创建 UsdtOrderService 并启动过期清理后台任务。
func NewUsdtOrderService(
	usdtRepo UsdtOrderRepository,
	settingRepo SettingRepository,
	adminService AdminService,
	settingService *SettingService,
	usdt *payment.UsdtPayment,
	lockCache LeaderLockCache,
	db *sql.DB,
) *UsdtOrderService {
	svc := &UsdtOrderService{
		usdtRepo:       usdtRepo,
		settingRepo:    settingRepo,
		adminService:   adminService,
		settingService: settingService,
		usdt:           usdt,
		expireInterval: time.Minute,
		stopCh:         make(chan struct{}),
		lockCache:      lockCache,
		db:             db,
		instanceID:     uuid.NewString(),
	}
	svc.Start()
	return svc
}

// Start 启动过期订单清理后台任务。
func (s *UsdtOrderService) Start() {
	if s == nil || s.usdtRepo == nil || s.expireInterval <= 0 {
		return
	}
	s.startOnce.Do(func() {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			ticker := time.NewTicker(s.expireInterval)
			defer ticker.Stop()
			s.runExpireCycle()
			for {
				select {
				case <-ticker.C:
					s.runExpireCycle()
				case <-s.stopCh:
					return
				}
			}
		}()
	})
}

// Stop 停止后台任务。
func (s *UsdtOrderService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.stopCh != nil {
			close(s.stopCh)
		}
	})
	s.wg.Wait()
}

func (s *UsdtOrderService) runExpireCycle() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	release, ok := tryAcquireSingletonLeaderLock(ctx, s.lockCache, s.db, "leader:usdt_expire", s.instanceID, 2*time.Minute)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}

	n, err := s.usdtRepo.ExpirePendingOrders(ctx)
	if err != nil {
		logger.LegacyPrintf("service.usdt_order", "expire pending orders failed: err=%v", err)
		return
	}
	if n > 0 {
		logger.LegacyPrintf("service.usdt_order", "expired usdt orders: n=%d", n)
	}
}

// GetConfig 获取 USDT 下单页面配置。
func (s *UsdtOrderService) GetConfig(ctx context.Context) (*UsdtOrderPublicConfig, error) {
	s.usdt.Reload(ctx)
	cfg := &UsdtOrderPublicConfig{Chains: s.usdt.UsableChains(ctx)}

	v, _ := s.settingRepo.GetValue(ctx, payment.SettingKeyUsdtEnabled)
	cfg.Enabled = v == "true"

	if minStr, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeMinAmount); minStr != "" {
		cfg.MinAmount, _ = strconv.ParseFloat(minStr, 64)
	}
	if maxStr, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeMaxAmount); maxStr != "" {
		cfg.MaxAmount, _ = strconv.ParseFloat(maxStr, 64)
	}
	if rate, err := s.usdt.QueryRate(ctx); err == nil {
		cfg.Rate = rate
	}
	return cfg, nil
}

// CreateUsdtOrderResult USDT 下单返回。
type CreateUsdtOrderResult struct {
	Order         *UsdtOrder
	Chain         string
	Address       string
	UsdtAmount    float64
	UsdtAmountStr string
	Rate          float64
	ExpiresIn     int
}

// CreateOrder 创建 USDT 充值订单（指定链）。
// 入参 usdtAmount 是用户填写的 USDT 数量（填多少付多少）；到账余额 = usdtAmount × 汇率。
func (s *UsdtOrderService) CreateOrder(ctx context.Context, userID int64, usdtAmount float64, chain, sourceDomain string) (*CreateUsdtOrderResult, error) {
	if s.usdt == nil {
		return nil, fmt.Errorf("usdt is not configured")
	}
	s.usdt.Reload(ctx)

	if !payment.IsSupportedChain(chain) {
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}
	if !s.usdt.IsChainUsable(ctx, chain) {
		return nil, fmt.Errorf("chain %s is not enabled or not configured", chain)
	}
	addr := s.usdt.ChainAddress(chain)

	if usdtAmount <= 0 {
		return nil, fmt.Errorf("usdt amount must be positive")
	}

	// 汇率快照：1 USDT = rate CNY（含加价 markup）
	rate, err := s.usdt.QueryRate(ctx)
	if err != nil || rate <= 0 {
		return nil, fmt.Errorf("usdt rate unavailable: %v", err)
	}
	// 到账余额 = 用户填写的 USDT × 汇率（按下单金额入账，与实收无关，容差内即成功）。
	credit := math.Round(usdtAmount*rate*100) / 100

	// 限额按到账余额校验（与 EPAY/AliMPay 共享 recharge 限额，单位是余额/CNY）。
	minAmount := 10.0
	maxAmount := 10000.0
	if v, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeMinAmount); v != "" {
		minAmount, _ = strconv.ParseFloat(v, 64)
	}
	if v, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeMaxAmount); v != "" {
		maxAmount, _ = strconv.ParseFloat(v, 64)
	}
	if credit < minAmount || credit > maxAmount {
		return nil, fmt.Errorf("credited amount %.2f must be between %.2f and %.2f", credit, minAmount, maxAmount)
	}

	expiresIn := s.usdt.OrderTimeoutSeconds()
	if expiresIn <= 0 {
		expiresIn = 1800
	}
	expiredAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	randN, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	orderNo := fmt.Sprintf("U%s%06d", time.Now().Format("20060102150405"), randN.Int64())

	order := &UsdtOrder{
		OrderNo:          orderNo,
		UserID:           userID,
		Amount:           credit, // 余额价值（admin 展示）
		CreditAmount:     credit, // 入账余额 = USDT × 汇率
		Multiplier:       1.0,
		Chain:            chain,
		ReceivingAddress: addr,
		UsdtRate:         rate,
		Status:           "pending",
		PayType:          "usdt",
		SourceDomain:     sourceDomain,
		ExpiredAt:        &expiredAt,
	}
	// 链上应付基数 = 用户填写的 USDT（再叠加唯一尾数用于归属匹配）。
	if err := s.usdtRepo.CreateWithUniqueUsdtAmount(ctx, order, usdtAmount, s.usdt.AmountOffset(), s.usdt.AmountReuseWindow()); err != nil {
		return nil, fmt.Errorf("create usdt order: %w", err)
	}

	return &CreateUsdtOrderResult{
		Order:         order,
		Chain:         order.Chain,
		Address:       addr,
		UsdtAmount:    order.UsdtAmount,
		UsdtAmountStr: payment.FormatUsdt(order.UsdtAmount),
		Rate:          rate,
		ExpiresIn:     expiresIn,
	}, nil
}

// GetOrderStatus 查询订单状态（只读）。
func (s *UsdtOrderService) GetOrderStatus(ctx context.Context, orderNo string, userID int64) (*UsdtOrder, error) {
	o, err := s.usdtRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if o == nil || o.UserID != userID {
		return nil, fmt.Errorf("order not found")
	}
	return o, nil
}

// ListUserOrders 查询用户充值记录。
func (s *UsdtOrderService) ListUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]*UsdtOrder, int, error) {
	offset := (page - 1) * pageSize
	return s.usdtRepo.ListByUserID(ctx, userID, pageSize, offset)
}

// ListAllOrders 管理端查询所有 USDT 订单。
func (s *UsdtOrderService) ListAllOrders(ctx context.Context, status string, userID *int64, page, pageSize int) ([]*UsdtOrder, int, error) {
	offset := (page - 1) * pageSize
	return s.usdtRepo.ListAll(ctx, status, userID, pageSize, offset)
}

// ExpirePendingOrders 清理过期订单。
func (s *UsdtOrderService) ExpirePendingOrders(ctx context.Context) (int, error) {
	return s.usdtRepo.ExpirePendingOrders(ctx)
}

// RefundOrder 管理员退款：已支付订单 paid→refunded 并扣回余额。语义与 AliMPay 一致。
func (s *UsdtOrderService) RefundOrder(ctx context.Context, orderNo, reason string) (*UsdtOrder, error) {
	o, err := s.usdtRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("query order: %w", err)
	}
	if o == nil {
		return nil, ErrUsdtOrderNotFound
	}
	if o.Status != "paid" {
		return nil, ErrUsdtOrderNotRefundable
	}

	if err := s.usdtRepo.UpdateStatus(ctx, orderNo, "paid", "refunded", nil); err != nil {
		if errors.Is(err, ErrOrderStatusConflict) {
			return nil, ErrUsdtOrderNotRefundable
		}
		return nil, fmt.Errorf("mark order refunded: %w", err)
	}

	notes := fmt.Sprintf("Refund USDT order %s (credited $%.2f)", orderNo, o.CreditAmount)
	if reason != "" {
		notes = notes + ": " + reason
	}
	if err := s.adminService.RefundUserBalance(ctx, o.UserID, o.CreditAmount, notes); err != nil {
		logger.LegacyPrintf("service.usdt_order", "refund clawback failed, rolling back: order=%s err=%v", orderNo, err)
		if rbErr := s.usdtRepo.UpdateStatus(ctx, orderNo, "refunded", "paid", nil); rbErr != nil {
			logger.LegacyPrintf("service.usdt_order", "CRITICAL: rollback refunded->paid failed, manual fix required: order=%s err=%v", orderNo, rbErr)
		}
		return nil, fmt.Errorf("refund clawback failed: %w", err)
	}
	o.Status = "refunded"
	return o, nil
}

// ===== payment.UsdtOrderMatcher 接口实现 =====

// GetPendingUsdtOrders UsdtMonitor 调用：拉取可匹配订单（pending + 宽限期内刚过期）供链上金额匹配。
// 含宽限期是为了让"临近/略超截止才到账"的订单仍能补入账（链上确认有延迟，加密资金不可逆）。
func (s *UsdtOrderService) GetPendingUsdtOrders(ctx context.Context) ([]payment.UsdtPendingOrder, error) {
	graceCutoff := time.Now().Add(-s.usdt.GraceWindow())
	orders, err := s.usdtRepo.ListMatchable(ctx, graceCutoff)
	if err != nil {
		return nil, err
	}
	out := make([]payment.UsdtPendingOrder, 0, len(orders))
	for _, o := range orders {
		out = append(out, payment.UsdtPendingOrder{
			OrderNo:    o.OrderNo,
			Chain:      o.Chain,
			UsdtAmount: o.UsdtAmount,
			UsdtAtomic: payment.UsdtToAtomic(o.UsdtAmount),
			CreatedAt:  o.CreatedAt,
			ExpiredAt:  o.ExpiredAt,
		})
	}
	return out, nil
}

// ConfirmUsdtOrderPaid UsdtMonitor 调用：匹配到链上转账后确认并入账。
func (s *UsdtOrderService) ConfirmUsdtOrderPaid(ctx context.Context, orderNo string, deposit payment.UsdtDeposit) error {
	o, err := s.usdtRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return fmt.Errorf("query order: %w", err)
	}
	if o == nil {
		return fmt.Errorf("order not found: %s", orderNo)
	}
	if o.Status == "paid" {
		return nil
	}
	if o.Status != "pending" && o.Status != "expired" {
		return fmt.Errorf("order status is %s, cannot pay", o.Status)
	}

	now := time.Now()
	fromStatus := o.Status
	upd := UsdtPaidUpdate{
		TxHash:         deposit.TxID,
		FromAddress:    deposit.FromAddress,
		PaidUsdtAmount: deposit.PaidUsdt,
	}
	if err := s.usdtRepo.MarkPaid(ctx, orderNo, fromStatus, upd, now); err != nil {
		if errors.Is(err, ErrOrderStatusConflict) {
			// 另一个实例已确认（CAS 失败）：视为幂等成功。
			return nil
		}
		return fmt.Errorf("mark order paid: %w", err)
	}

	_, err = s.adminService.UpdateUserBalance(ctx, o.UserID, o.CreditAmount, "add",
		fmt.Sprintf("USDT order %s, paid %.6f USDT, credited $%.2f", orderNo, deposit.PaidUsdt, o.CreditAmount))
	if err != nil {
		logger.LegacyPrintf("service.usdt_order", "credit balance failed, rolling back: order=%s err=%v", orderNo, err)
		if rbErr := s.usdtRepo.UpdateStatus(ctx, orderNo, "paid", fromStatus, nil); rbErr != nil {
			logger.LegacyPrintf("service.usdt_order", "CRITICAL: rollback paid->%s failed: order=%s err=%v", fromStatus, orderNo, rbErr)
		}
		return fmt.Errorf("credit balance failed: %w", err)
	}
	logger.LegacyPrintf("service.usdt_order", "usdt success: order=%s user=%d paid=%.6f USDT credit=$%.2f tx=%s",
		orderNo, o.UserID, deposit.PaidUsdt, o.CreditAmount, deposit.TxID)
	return nil
}
