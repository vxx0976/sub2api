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
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/epay"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// Setting keys for recharge
const (
	SettingKeyRechargeEnabled    = "recharge_enabled"
	SettingKeyRechargeMinAmount  = "recharge_min_amount"
	SettingKeyRechargeMaxAmount  = "recharge_max_amount"
	SettingKeyRechargeTiers      = "recharge_tiers"
	SettingKeyRechargePayTypes   = "recharge_pay_types" // JSON array: ["alipay","wxpay"]
	SettingKeyEpayAPIURL         = "epay_api_url"
	SettingKeyEpayPID            = "epay_pid"
	SettingKeyEpayPublicKey      = "epay_platform_public_key"
	SettingKeyEpayPrivateKey     = "epay_merchant_private_key"
)

// RechargeService 充值业务服务
type RechargeService struct {
	orderRepo      RechargeOrderRepository
	settingRepo    SettingRepository
	adminService   AdminService
	settingService *SettingService
}

// NewRechargeService creates a new RechargeService
func NewRechargeService(
	orderRepo RechargeOrderRepository,
	settingRepo SettingRepository,
	adminService AdminService,
	settingService *SettingService,
) *RechargeService {
	return &RechargeService{
		orderRepo:      orderRepo,
		settingRepo:    settingRepo,
		adminService:   adminService,
		settingService: settingService,
	}
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

	if tiersJSON, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeTiers); tiersJSON != "" {
		_ = json.Unmarshal([]byte(tiersJSON), &cfg.Tiers)
	}

	if payTypesJSON, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargePayTypes); payTypesJSON != "" {
		_ = json.Unmarshal([]byte(payTypesJSON), &cfg.PayTypes)
	}
	if len(cfg.PayTypes) == 0 {
		cfg.PayTypes = []string{"alipay", "wxpay"}
	}

	cfg.SellingPrice = s.settingService.GetPlatformSellingPrice(ctx)

	return cfg, nil
}

// CreateOrder 创建充值订单
func (s *RechargeService) CreateOrder(ctx context.Context, userID int64, amount float64, payType string, baseURL string) (*RechargeOrder, string, error) {
	// 1. 检查充值是否启用
	enabled, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeEnabled)
	if enabled != "true" {
		return nil, "", fmt.Errorf("recharge is not enabled")
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
		return nil, "", fmt.Errorf("amount must be between %.2f and %.2f", minAmount, maxAmount)
	}

	// 3. 计算到账金额
	sellingPrice := s.settingService.GetPlatformSellingPrice(ctx)
	if sellingPrice <= 0 {
		return nil, "", fmt.Errorf("platform selling price not configured")
	}

	// 查找倍率档位
	multiplier := 1.0
	if tiersJSON, _ := s.settingRepo.GetValue(ctx, SettingKeyRechargeTiers); tiersJSON != "" {
		var tiers []RechargeTier
		if err := json.Unmarshal([]byte(tiersJSON), &tiers); err == nil {
			for _, t := range tiers {
				if amount >= t.Min && (t.Max == nil || amount <= *t.Max) {
					multiplier = t.Multiplier
					break
				}
			}
		}
	}

	creditAmount := math.Round(amount/sellingPrice*multiplier*100) / 100

	// 4. 生成订单号（时间戳 + 6位随机数）
	randN, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	orderNo := fmt.Sprintf("R%s%06d", time.Now().Format("20060102150405"), randN.Int64())

	// 5. 创建易支付客户端
	epayClient, err := s.getEpayClient(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("payment gateway not configured: %w", err)
	}

	// 6. 构造通知和跳转 URL
	baseURL = strings.TrimRight(baseURL, "/")
	notifyURL := baseURL + "/api/v1/recharge/notify"
	returnURL := baseURL + "/api/v1/recharge/return"

	// 7. 调用易支付创建支付
	payLink, err := epayClient.GetPayLink(epay.CreatePaymentRequest{
		OutTradeNo: orderNo,
		Type:       payType,
		Name:       "Balance Recharge",
		Money:      fmt.Sprintf("%.2f", amount),
		NotifyURL:  notifyURL,
		ReturnURL:  returnURL,
	})
	if err != nil {
		return nil, "", fmt.Errorf("create payment failed: %w", err)
	}

	// 8. 写入数据库
	expiredAt := time.Now().Add(5 * time.Minute)
	order := &RechargeOrder{
		OrderNo:      orderNo,
		UserID:       userID,
		Amount:       amount,
		CreditAmount: creditAmount,
		Multiplier:   multiplier,
		Status:       "pending",
		PayType:      payType,
		ExpiredAt:    &expiredAt,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, "", fmt.Errorf("create order: %w", err)
	}

	return order, payLink, nil
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
	if order.Status != "pending" {
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
	if err := s.orderRepo.UpdateStatus(ctx, orderNo, "pending", "paid", &tradeNo, &now); err != nil {
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

// GetOrderStatus 查询订单状态
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
	pubKey, _ := s.settingRepo.GetValue(ctx, SettingKeyEpayPublicKey)
	privKey, _ := s.settingRepo.GetValue(ctx, SettingKeyEpayPrivateKey)

	if apiURL == "" || pid == "" || pubKey == "" || privKey == "" {
		return nil, fmt.Errorf("epay configuration incomplete")
	}

	return epay.NewClient(apiURL, pid, pubKey, privKey)
}
