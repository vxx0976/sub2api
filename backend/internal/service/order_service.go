package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/payment"
)

// AdminAliMPayConfig AliMPay 后台配置 DTO
// 敏感字段（PrivateKey / AlipayPublicKey）GET 时只返回 has_* 指示，PUT 时空字符串表示保留原值
type AdminAliMPayConfig struct {
	Enabled                bool    `json:"enabled"`
	Mode                   string  `json:"mode"` // business_qr | transfer
	AppID                  string  `json:"app_id"`
	HasPrivateKey          bool    `json:"has_private_key"`
	HasAlipayPublicKey     bool    `json:"has_alipay_public_key"`
	ServerURL              string  `json:"server_url"`
	TransferUserID         string  `json:"transfer_user_id"`
	BusinessQRURL          string  `json:"business_qr_url"`
	BusinessQRPath         string  `json:"business_qr_path"`
	AmountOffset           float64 `json:"amount_offset"`
	MatchToleranceSeconds  int     `json:"match_tolerance_seconds"`
	MonitorIntervalSeconds int     `json:"monitor_interval_seconds"`
	QueryMinutesBack       int     `json:"query_minutes_back"`
	OrderTimeoutSeconds    int     `json:"order_timeout_seconds"`
}

// AdminAliMPayConfigUpdate 更新请求（nil 指针表示不变；字符串空表示保留原值，避免覆盖敏感字段）
type AdminAliMPayConfigUpdate struct {
	Enabled                *bool    `json:"enabled"`
	Mode                   *string  `json:"mode"`
	AppID                  *string  `json:"app_id"`
	PrivateKey             *string  `json:"private_key"`       // 空字符串 = 保留原值
	AlipayPublicKey        *string  `json:"alipay_public_key"` // 空字符串 = 保留原值
	ServerURL              *string  `json:"server_url"`
	TransferUserID         *string  `json:"transfer_user_id"`
	BusinessQRURL          *string  `json:"business_qr_url"`
	BusinessQRPath         *string  `json:"business_qr_path"`
	AmountOffset           *float64 `json:"amount_offset"`
	MatchToleranceSeconds  *int     `json:"match_tolerance_seconds"`
	MonitorIntervalSeconds *int     `json:"monitor_interval_seconds"`
	QueryMinutesBack       *int     `json:"query_minutes_back"`
	OrderTimeoutSeconds    *int     `json:"order_timeout_seconds"`
}

// Setting keys for AliMPay Order（运行时开关；倍率/限额复用 EPAY 的 SettingKeyRecharge* key）
const (
	SettingKeyAliMPayEnabled = "alimpay_enabled"
)

// OrderService AliMPay 充值业务服务
//
// 与 RechargeService 业务语义一致（充值到账余额），差异在于支付通道：
//   - RechargeService：EPAY 聚合支付，有 webhook
//   - OrderService：AliMPay 个人免签，走支付宝账单轮询（无 webhook，由 AlipayMonitor 主动匹配）
//
// 同时实现 payment.OrderMatcher 接口，供 AlipayMonitor 回调确认支付。
type OrderService struct {
	orderRepo      OrderRepository
	settingRepo    SettingRepository
	adminService   AdminService
	settingService *SettingService
	alipay         *payment.AlipayPayment
	locker         *LeaderLocker
	expireInterval time.Duration

	stopCh    chan struct{}
	stopOnce  sync.Once
	startOnce sync.Once
	wg        sync.WaitGroup
}

// NewOrderService 创建 OrderService
func NewOrderService(
	orderRepo OrderRepository,
	settingRepo SettingRepository,
	adminService AdminService,
	settingService *SettingService,
	alipay *payment.AlipayPayment,
	locker *LeaderLocker,
) *OrderService {
	svc := &OrderService{
		orderRepo:      orderRepo,
		settingRepo:    settingRepo,
		adminService:   adminService,
		settingService: settingService,
		alipay:         alipay,
		locker:         locker,
		expireInterval: time.Minute,
		stopCh:         make(chan struct{}),
	}
	svc.Start()
	return svc
}

// Start 启动过期订单清理后台任务
func (s *OrderService) Start() {
	if s == nil || s.orderRepo == nil || s.expireInterval <= 0 {
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

// Stop 停止后台任务
func (s *OrderService) Stop() {
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

// GetConfig 获取 AliMPay 下单页面配置（1:1 到账，不走倍率/汇率换算）
func (s *OrderService) GetConfig(ctx context.Context) (*OrderPublicConfig, error) {
	cfg := &OrderPublicConfig{}

	enabled, _ := s.settingRepo.GetValue(ctx, SettingKeyAliMPayEnabled)
	cfg.Enabled = enabled == "true"

	if minStr, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeMinAmount); minStr != "" {
		cfg.MinAmount, _ = strconv.ParseFloat(minStr, 64)
	}
	if maxStr, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeMaxAmount); maxStr != "" {
		cfg.MaxAmount, _ = strconv.ParseFloat(maxStr, 64)
	}

	// AliMPay 走 1:1（CNY 支付金额 = 到账平台余额）
	cfg.SellingPrice = 1
	if s.alipay != nil {
		cfg.Mode = s.alipay.Mode()
	}

	return cfg, nil
}

// CreateOrderAliMPayResult AliMPay 下单返回
type CreateOrderAliMPayResult struct {
	Order     *Order
	QRCodeURL string // 支付二维码链接（business_qr 模式是商家收款码 URL；transfer 模式是 alipays:// 深链）
	Mode      string // business_qr | transfer
	ExpiresIn int    // 剩余支付秒数
}

// CreateOrder 创建 AliMPay 充值订单
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, amount float64, sourceDomain string) (*CreateOrderAliMPayResult, error) {
	if s.alipay == nil {
		return nil, fmt.Errorf("alimpay is not configured")
	}

	// 从 settings 热加载 AliMPay 配置（商户信息/模式/金额偏移等可在 SettingsView 运行时调整）
	s.alipay.Reload(ctx)

	// 1. 检查启用开关
	enabled, _ := s.settingRepo.GetValue(ctx, SettingKeyAliMPayEnabled)
	if enabled != "true" {
		return nil, fmt.Errorf("alimpay is not enabled")
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

	// 3. 到账金额：AliMPay 走 1:1（CNY 支付金额直接等于平台余额），不走倍率/汇率换算
	multiplier := 1.0
	creditAmount := amount

	// 4. 生成订单号（AliMPay 专用前缀 A）
	randN, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	orderNo := fmt.Sprintf("A%s%06d", time.Now().Format("20060102150405"), randN.Int64())

	// 5. 落库。business_qr 的唯一支付金额必须基于数据库里的活跃订单分配，
	// 避免进程重启或多实例部署后复用未完成订单的金额。
	expiresIn := s.alipay.OrderTimeoutSeconds()
	if expiresIn <= 0 {
		expiresIn = 600 // 默认 10 分钟
	}
	expiredAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	mode := s.alipay.Mode()
	order := &Order{
		OrderNo:       orderNo,
		UserID:        userID,
		Amount:        amount,
		PaymentAmount: amount,
		CreditAmount:  creditAmount,
		Multiplier:    multiplier,
		Status:        "pending",
		PayType:       "alipay",
		SourceDomain:  sourceDomain,
		ExpiredAt:     &expiredAt,
	}
	if mode == "transfer" {
		// transfer 模式靠订单号 memo 匹配账单，不需要唯一支付金额；
		// DB 里 payment_amount 留 NULL 避免和 business_qr 订单抢唯一 partial index。
		order.PaymentAmount = 0
		if err := s.orderRepo.Create(ctx, order); err != nil {
			return nil, fmt.Errorf("create order: %w", err)
		}
		order.PaymentAmount = amount // 回填给前端展示
	} else {
		if err := s.orderRepo.CreateWithUniquePaymentAmount(ctx, order, amount, s.alipay.AmountOffset(), s.alipay.AmountReuseWindow()); err != nil {
			return nil, fmt.Errorf("create order: %w", err)
		}
	}

	// 6. 生成支付信息。business_qr 模式使用数据库分配后的 payment_amount。
	payInfo, err := s.alipay.GeneratePaymentInfo(orderNo, order.PaymentAmount)
	if err != nil {
		return nil, fmt.Errorf("generate payment info: %w", err)
	}

	return &CreateOrderAliMPayResult{
		Order:     order,
		QRCodeURL: payInfo.QRCodeURL,
		Mode:      payInfo.Mode,
		ExpiresIn: expiresIn,
	}, nil
}

// GetOrderStatus 查询订单状态（只读，不主动查 AliMPay——Monitor 在后台轮询）
func (s *OrderService) GetOrderStatus(ctx context.Context, orderNo string, userID int64) (*Order, error) {
	o, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if o == nil || o.UserID != userID {
		return nil, fmt.Errorf("order not found")
	}
	return o, nil
}

// ListUserOrders 查询用户充值记录
func (s *OrderService) ListUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]*Order, int, error) {
	offset := (page - 1) * pageSize
	return s.orderRepo.ListByUserID(ctx, userID, pageSize, offset)
}

// ListAllOrders 管理端查询所有 AliMPay 订单
func (s *OrderService) ListAllOrders(ctx context.Context, status string, userID *int64, page, pageSize int) ([]*Order, int, error) {
	offset := (page - 1) * pageSize
	return s.orderRepo.ListAll(ctx, status, userID, pageSize, offset)
}

// ExpirePendingOrders 清理过期订单（被 runExpireCycle 和 admin 触发）
func (s *OrderService) ExpirePendingOrders(ctx context.Context) (int, error) {
	return s.orderRepo.ExpirePendingOrders(ctx)
}

func (s *OrderService) runExpireCycle() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if s.locker != nil {
		release, ok := s.locker.TryAcquire(ctx, "leader:alimpay_expire", 2*time.Minute)
		if !ok {
			return
		}
		if release != nil {
			defer release()
		}
	}

	n, err := s.ExpirePendingOrders(ctx)
	if err != nil {
		logger.LegacyPrintf("service.order", "expire pending orders failed: err=%v", err)
		return
	}
	if n > 0 {
		logger.LegacyPrintf("service.order", "expired orders: n=%d", n)
	}
}

// ===== AliMPay Admin 配置 =====

// GetAdminAliMPayConfig 读取当前配置（敏感字段脱敏）
func (s *OrderService) GetAdminAliMPayConfig(ctx context.Context) (*AdminAliMPayConfig, error) {
	get := func(k string) string {
		v, _ := s.settingRepo.GetValue(ctx, k)
		return v
	}
	parseFloat := func(v string) float64 {
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	parseInt := func(v string) int {
		i, _ := strconv.Atoi(v)
		return i
	}

	cfg := &AdminAliMPayConfig{
		Enabled:                get(payment.SettingKeyAliMPayEnabled) == "true",
		Mode:                   get(payment.SettingKeyAliMPayMode),
		AppID:                  get(payment.SettingKeyAliMPayAppID),
		HasPrivateKey:          get(payment.SettingKeyAliMPayPrivateKey) != "",
		HasAlipayPublicKey:     get(payment.SettingKeyAliMPayAlipayPublicKey) != "",
		ServerURL:              get(payment.SettingKeyAliMPayServerURL),
		TransferUserID:         get(payment.SettingKeyAliMPayTransferUserID),
		BusinessQRURL:          get(payment.SettingKeyAliMPayBusinessQRURL),
		BusinessQRPath:         get(payment.SettingKeyAliMPayBusinessQRPath),
		AmountOffset:           parseFloat(get(payment.SettingKeyAliMPayAmountOffset)),
		MatchToleranceSeconds:  parseInt(get(payment.SettingKeyAliMPayMatchToleranceSeconds)),
		MonitorIntervalSeconds: parseInt(get(payment.SettingKeyAliMPayMonitorIntervalSeconds)),
		QueryMinutesBack:       parseInt(get(payment.SettingKeyAliMPayQueryMinutesBack)),
		OrderTimeoutSeconds:    parseInt(get(payment.SettingKeyAliMPayOrderTimeoutSeconds)),
	}
	return cfg, nil
}

// UpdateAdminAliMPayConfig 更新配置（字段为 nil 则不变；PrivateKey/AlipayPublicKey 空字符串保留原值）
// 保存后立刻 Reload 让 SDK 感知新值
func (s *OrderService) UpdateAdminAliMPayConfig(ctx context.Context, req *AdminAliMPayConfigUpdate) error {
	if req == nil {
		return nil
	}
	set := func(key, value string) error {
		return s.settingRepo.Set(ctx, key, value)
	}

	if req.Enabled != nil {
		if err := set(payment.SettingKeyAliMPayEnabled, strconv.FormatBool(*req.Enabled)); err != nil {
			return err
		}
	}
	if req.Mode != nil {
		if err := set(payment.SettingKeyAliMPayMode, *req.Mode); err != nil {
			return err
		}
	}
	if req.AppID != nil {
		if err := set(payment.SettingKeyAliMPayAppID, *req.AppID); err != nil {
			return err
		}
	}
	if req.PrivateKey != nil && *req.PrivateKey != "" {
		if err := set(payment.SettingKeyAliMPayPrivateKey, *req.PrivateKey); err != nil {
			return err
		}
	}
	if req.AlipayPublicKey != nil && *req.AlipayPublicKey != "" {
		if err := set(payment.SettingKeyAliMPayAlipayPublicKey, *req.AlipayPublicKey); err != nil {
			return err
		}
	}
	if req.ServerURL != nil {
		if err := set(payment.SettingKeyAliMPayServerURL, *req.ServerURL); err != nil {
			return err
		}
	}
	if req.TransferUserID != nil {
		if err := set(payment.SettingKeyAliMPayTransferUserID, *req.TransferUserID); err != nil {
			return err
		}
	}
	if req.BusinessQRURL != nil {
		if err := set(payment.SettingKeyAliMPayBusinessQRURL, *req.BusinessQRURL); err != nil {
			return err
		}
	}
	if req.BusinessQRPath != nil {
		if err := set(payment.SettingKeyAliMPayBusinessQRPath, *req.BusinessQRPath); err != nil {
			return err
		}
	}
	if req.AmountOffset != nil {
		if err := set(payment.SettingKeyAliMPayAmountOffset, strconv.FormatFloat(*req.AmountOffset, 'f', -1, 64)); err != nil {
			return err
		}
	}
	if req.MatchToleranceSeconds != nil {
		if err := set(payment.SettingKeyAliMPayMatchToleranceSeconds, strconv.Itoa(*req.MatchToleranceSeconds)); err != nil {
			return err
		}
	}
	if req.MonitorIntervalSeconds != nil {
		if err := set(payment.SettingKeyAliMPayMonitorIntervalSeconds, strconv.Itoa(*req.MonitorIntervalSeconds)); err != nil {
			return err
		}
	}
	if req.QueryMinutesBack != nil {
		if err := set(payment.SettingKeyAliMPayQueryMinutesBack, strconv.Itoa(*req.QueryMinutesBack)); err != nil {
			return err
		}
	}
	if req.OrderTimeoutSeconds != nil {
		if err := set(payment.SettingKeyAliMPayOrderTimeoutSeconds, strconv.Itoa(*req.OrderTimeoutSeconds)); err != nil {
			return err
		}
	}

	// 立刻 reload 到 SDK 让新配置生效
	if s.alipay != nil {
		s.alipay.Reload(ctx)
	}
	return nil
}

// ===== payment.OrderMatcher 接口实现 =====

// GetPendingOrders AliMPay Monitor 调用：拉取待支付订单供账单匹配
func (s *OrderService) GetPendingOrders(ctx context.Context) ([]payment.PendingOrder, error) {
	orders, err := s.orderRepo.ListPending(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]payment.PendingOrder, 0, len(orders))
	for _, o := range orders {
		out = append(out, payment.PendingOrder{
			OrderNo:          o.OrderNo,
			Amount:           o.Amount,
			PaymentAmount:    o.PaymentAmount,
			HasPaymentAmount: o.PaymentAmount > 0,
			CreatedAt:        o.CreatedAt,
			ExpiredAt:        o.ExpiredAt,
		})
	}
	return out, nil
}

// ConfirmOrderPaid AliMPay Monitor 调用：匹配到账单后确认支付成功并入账
func (s *OrderService) ConfirmOrderPaid(ctx context.Context, orderNo, tradeNo, payType string) error {
	o, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
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
	if err := s.orderRepo.UpdateStatus(ctx, orderNo, fromStatus, "paid", &tradeNo, &now); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	_, err = s.adminService.UpdateUserBalance(ctx, o.UserID, o.CreditAmount, "add",
		fmt.Sprintf("AliMPay order %s, paid ¥%.2f, credited $%.2f", orderNo, o.Amount, o.CreditAmount))
	if err != nil {
		logger.LegacyPrintf("service.order", "credit balance failed, rolling back: order=%s err=%v", orderNo, err)
		if rbErr := s.orderRepo.UpdateStatus(ctx, orderNo, "paid", fromStatus, nil, nil); rbErr != nil {
			logger.LegacyPrintf("service.order", "CRITICAL: rollback failed: order=%s err=%v", orderNo, rbErr)
		}
		return fmt.Errorf("credit balance failed: %w", err)
	}

	logger.LegacyPrintf("service.order", "alimpay success: order=%s user=%d amount=¥%.2f credit=$%.2f", orderNo, o.UserID, o.Amount, o.CreditAmount)
	return nil
}
