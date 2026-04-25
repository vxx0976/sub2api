package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/epay"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// Setting keys for recharge
const (
	SettingKeyRechargeEnabled   = "recharge_enabled"
	SettingKeyRechargeMinAmount = "recharge_min_amount"
	SettingKeyRechargeMaxAmount = "recharge_max_amount"
	SettingKeyRechargePayTypes  = "recharge_pay_types" // JSON array: ["alipay","wxpay"]
	SettingKeyEpayAPIURL = "epay_api_url"
	SettingKeyEpayPID    = "epay_pid"
	SettingKeyEpayKey    = "epay_key"
)

// RechargeService 充值业务服务
type RechargeService struct {
	orderRepo    RechargeOrderRepository
	settingRepo  SettingRepository
	adminService AdminService
	locker       *LeaderLocker
	syncInterval time.Duration
	stopCh       chan struct{}
	stopOnce     sync.Once
	startOnce    sync.Once
	wg           sync.WaitGroup
}

// NewRechargeService creates a new RechargeService
func NewRechargeService(
	orderRepo RechargeOrderRepository,
	settingRepo SettingRepository,
	adminService AdminService,
	locker *LeaderLocker,
) *RechargeService {
	svc := &RechargeService{
		orderRepo:    orderRepo,
		settingRepo:  settingRepo,
		adminService: adminService,
		locker:       locker,
		syncInterval: time.Minute,
		stopCh:       make(chan struct{}),
	}
	svc.Start()
	return svc
}

// Start 启动待支付订单后台补偿同步。
func (s *RechargeService) Start() {
	if s == nil || s.orderRepo == nil || s.syncInterval <= 0 {
		return
	}

	s.startOnce.Do(func() {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()

			ticker := time.NewTicker(s.syncInterval)
			defer ticker.Stop()

			s.runPendingOrderSweep()
			for {
				select {
				case <-ticker.C:
					s.runPendingOrderSweep()
				case <-s.stopCh:
					return
				}
			}
		}()
	})
}

// Stop 停止后台补偿同步。
func (s *RechargeService) Stop() {
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

// GetConfig 获取充值配置
func (s *RechargeService) GetConfig(ctx context.Context) (*RechargePublicConfig, error) {
	cfg := &RechargePublicConfig{}

	enabled, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeEnabled)
	cfg.Enabled = enabled == "true"

	if minStr, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeMinAmount); minStr != "" {
		cfg.MinAmount, _ = strconv.ParseFloat(minStr, 64)
	}
	if maxStr, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeMaxAmount); maxStr != "" {
		cfg.MaxAmount, _ = strconv.ParseFloat(maxStr, 64)
	}

	if payTypesJSON, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargePayTypes); payTypesJSON != "" {
		_ = json.Unmarshal([]byte(payTypesJSON), &cfg.PayTypes)
	}
	if len(cfg.PayTypes) == 0 {
		cfg.PayTypes = []string{"alipay", "wxpay"}
	}

	// 易支付走 1:1（¥1 = $1），与 AliMPay 保持统一
	cfg.SellingPrice = 1

	return cfg, nil
}

// CreateOrderResult 创建充值订单结果
type CreateOrderResult struct {
	Order  *RechargeOrder
	PayURL string // 支付页面链接
	QRCode string // 二维码链接
}

// CreateOrder 创建充值订单
func (s *RechargeService) CreateOrder(ctx context.Context, userID int64, amount float64, payType string, baseURL string, clientIP string) (*CreateOrderResult, error) {
	// 1. 检查充值是否启用
	enabled, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeEnabled)
	if enabled != "true" {
		return nil, fmt.Errorf("recharge is not enabled")
	}

	// 2. 校验金额范围
	minAmount := 10.0
	maxAmount := 10000.0
	if v, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeMinAmount); v != "" {
		minAmount, _ = strconv.ParseFloat(v, 64)
	}
	if v, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeMaxAmount); v != "" {
		maxAmount, _ = strconv.ParseFloat(v, 64)
	}
	if amount < minAmount || amount > maxAmount {
		return nil, fmt.Errorf("amount must be between %.2f and %.2f", minAmount, maxAmount)
	}

	// 3. 到账金额：易支付走 1:1（CNY 支付金额 = 到账平台余额），不走倍率换算，与 AliMPay 保持统一
	multiplier := 1.0
	creditAmount := amount

	// 4. 生成订单号（时间戳 + 6位随机数）
	randN, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	orderNo := fmt.Sprintf("R%s%06d", time.Now().Format("20060102150405"), randN.Int64())

	// 5. 创建易支付客户端
	epayClient, err := s.getEpayClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("payment gateway not configured: %w", err)
	}

	// 6. 构造通知和跳转 URL
	baseURL = strings.TrimRight(baseURL, "/")
	notifyURL := baseURL + "/api/v1/recharge/notify"
	returnURL := baseURL + "/api/v1/recharge/return"

	// 7. 调用易支付 API 创建支付（API 模式，返回二维码）
	payResp, err := epayClient.CreatePayment(epay.CreatePaymentRequest{
		OutTradeNo: orderNo,
		Type:       payType,
		Name:       "Balance Recharge",
		Money:      fmt.Sprintf("%.2f", amount),
		NotifyURL:  notifyURL,
		ReturnURL:  returnURL,
		ClientIP:   clientIP,
	})
	if err != nil {
		return nil, fmt.Errorf("create payment failed: %w", err)
	}

	// 8. 写入数据库
	expiredAt := time.Now().Add(5 * time.Minute)
	order := &RechargeOrder{
		OrderNo:      orderNo,
		TradeNo:      payResp.TradeNo,
		UserID:       userID,
		Amount:       amount,
		CreditAmount: creditAmount,
		Multiplier:   multiplier,
		Status:       "pending",
		PayType:      payType,
		ExpiredAt:    &expiredAt,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return &CreateOrderResult{
		Order:  order,
		PayURL: payResp.PayURL,
		QRCode: payResp.QRCode,
	}, nil
}

// HandleNotify 处理异步回调
func (s *RechargeService) HandleNotify(ctx context.Context, params map[string]string) error {
	// 1. 验签
	epayClient, err := s.getEpayClient(ctx)
	if err != nil {
		return fmt.Errorf("payment gateway not configured")
	}

	if !epayClient.VerifyNotify(params) {
		return fmt.Errorf("signature verification failed")
	}

	// 2. 检查交易状态
	if params["trade_status"] != "TRADE_SUCCESS" {
		return nil // 非成功状态忽略
	}

	orderNo := params["out_trade_no"]
	tradeNo := params["trade_no"]

	// 3. 查询订单
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return fmt.Errorf("query order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("order not found: %s", orderNo)
	}

	// 幂等：已支付则跳过
	if order.Status == "paid" {
		return nil
	}
	// 允许 pending 和 expired 订单通过回调完成支付（回调可能在订单过期后才到达）
	if order.Status != "pending" && order.Status != "expired" {
		return fmt.Errorf("order status is %s, cannot pay", order.Status)
	}

	// 4. 校验金额
	if money := params["money"]; money != "" {
		paidAmount, _ := strconv.ParseFloat(money, 64)
		if math.Abs(paidAmount-order.Amount) > 0.01 {
			return fmt.Errorf("amount mismatch: expected %.2f, got %s", order.Amount, money)
		}
	}

	// 5. CAS 更新订单状态
	now := time.Now()
	if err := s.orderRepo.UpdateStatus(ctx, orderNo, order.Status, "paid", &tradeNo, &now); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	// 6. 充值到账
	_, err = s.adminService.UpdateUserBalance(ctx, order.UserID, order.CreditAmount, "add",
		fmt.Sprintf("Recharge order %s, paid ¥%.2f, credited $%.2f", orderNo, order.Amount, order.CreditAmount))
	if err != nil {
		logger.LegacyPrintf("service.recharge", "credit balance failed, rolling back order status: order=%s user=%d amount=%.2f err=%v", orderNo, order.UserID, order.CreditAmount, err)
		// 回滚订单状态，避免订单标记 paid 但余额未到账
		if rbErr := s.orderRepo.UpdateStatus(ctx, orderNo, "paid", "pending", nil, nil); rbErr != nil {
			logger.LegacyPrintf("service.recharge", "CRITICAL: rollback order status failed: order=%s err=%v", orderNo, rbErr)
		}
		return fmt.Errorf("credit balance failed: %w", err)
	}

	logger.LegacyPrintf("service.recharge", "recharge success: order=%s user=%d amount=¥%.2f credit=$%.2f", orderNo, order.UserID, order.Amount, order.CreditAmount)
	return nil
}

// runPendingOrderSweep 周期性地清理过期的 pending 订单。
// 注意：易支付到账只走 ZPAY 异步回调（POST /api/v1/recharge/notify），
// 不主动轮询 ZPAY 查单——上游 trade_no 字段语义不稳定，回放反而易踩坑。
func (s *RechargeService) runPendingOrderSweep() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	release, ok := s.locker.TryAcquire(ctx, "leader:recharge_expire", 2*time.Minute)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}

	expired, err := s.ExpirePendingOrders(ctx)
	if err != nil {
		logger.LegacyPrintf("service.recharge", "expire pending orders failed: err=%v", err)
		return
	}

	if expired > 0 {
		logger.LegacyPrintf("service.recharge", "pending order sweep completed: expired=%d", expired)
	}
}

// GetOrderStatus 查询订单状态（仅读本地，依赖 ZPAY 回调更新）
func (s *RechargeService) GetOrderStatus(ctx context.Context, orderNo string, userID int64) (*RechargeOrder, error) {
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order == nil || order.UserID != userID {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

// ListUserOrders 查询用户充值记录
func (s *RechargeService) ListUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]*RechargeOrder, int, error) {
	offset := (page - 1) * pageSize
	return s.orderRepo.ListByUserID(ctx, userID, pageSize, offset)
}

// ListAllOrders 管理端查询所有充值订单
func (s *RechargeService) ListAllOrders(ctx context.Context, status string, userID *int64, page, pageSize int) ([]*RechargeOrder, int, error) {
	offset := (page - 1) * pageSize
	return s.orderRepo.ListAll(ctx, status, userID, pageSize, offset)
}

// ExpirePendingOrders 清理过期订单
func (s *RechargeService) ExpirePendingOrders(ctx context.Context) (int, error) {
	return s.orderRepo.ExpirePendingOrders(ctx)
}

func (s *RechargeService) getEpayClient(ctx context.Context) (*epay.Client, error) {
	apiURL, _ := s.settingRepo.GetValue(ctx, SettingKeyEpayAPIURL)
	pid, _ := s.settingRepo.GetValue(ctx, SettingKeyEpayPID)
	key, _ := s.settingRepo.GetValue(ctx, SettingKeyEpayKey)

	if apiURL == "" || pid == "" || key == "" {
		return nil, fmt.Errorf("epay configuration incomplete")
	}

	return epay.NewClient(apiURL, pid, key)
}
