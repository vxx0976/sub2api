package service

import (
	"context"
	"errors"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// UsdtOrder USDT(TRC20) 充值订单领域模型（自建链上收款）。
// 与 AliMPay 的 Order 并列：CNY 计价（Amount/CreditAmount，1:1 入账），
// UsdtAmount 是按 UsdtRate 冻结汇率换算的、带唯一尾数的链上应付金额。
type UsdtOrder struct {
	ID               int64      `json:"id"`
	OrderNo          string     `json:"order_no"`
	TradeNo          string     `json:"trade_no"` // 链上交易哈希
	UserID           int64      `json:"user_id"`
	Amount           float64    `json:"amount"`        // 订单金额（CNY）
	CreditAmount     float64    `json:"credit_amount"` // 到账余额（= Amount）
	Multiplier       float64    `json:"multiplier"`
	Chain            string     `json:"chain"`
	ReceivingAddress string     `json:"receiving_address"`
	UsdtRate         float64    `json:"usdt_rate"`   // 冻结汇率：1 USDT = ? CNY
	UsdtAmount       float64    `json:"usdt_amount"` // 应付 USDT（含唯一尾数）
	PaidUsdtAmount   *float64   `json:"paid_usdt_amount"`
	FromAddress      string     `json:"from_address"`
	BlockNumber      *int64     `json:"block_number"`
	Status           string     `json:"status"` // pending/paid/expired/refunded
	PayType          string     `json:"pay_type"`
	PaidAt           *time.Time `json:"paid_at"`
	SourceDomain     string     `json:"source_domain"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ExpiredAt        *time.Time `json:"expired_at"`
}

// UsdtOrderPublicConfig USDT 充值页面配置。
type UsdtOrderPublicConfig struct {
	Enabled   bool     `json:"enabled"`
	MinAmount float64  `json:"min_amount"`
	MaxAmount float64  `json:"max_amount"`
	Chains    []string `json:"chains"` // 当前可收款的链（trc20/bep20/ton）
	Rate      float64  `json:"rate"`   // 当前换算汇率（1 USDT = ? CNY），供前端预估
}

// UsdtPaidUpdate 确认到账时写入订单的链上信息。
type UsdtPaidUpdate struct {
	TxHash         string
	FromAddress    string
	PaidUsdtAmount float64
	BlockNumber    *int64
}

// UsdtOrderRepository USDT 订单数据访问接口。
type UsdtOrderRepository interface {
	// CreateWithUniqueUsdtAmount 在 (chain, usdt_amount) 上分配唯一应付金额并落库。
	CreateWithUniqueUsdtAmount(ctx context.Context, order *UsdtOrder, baseUsdt, amountOffset float64, reuseWindow time.Duration) error
	GetByOrderNo(ctx context.Context, orderNo string) (*UsdtOrder, error)
	// UpdateStatus CAS 状态流转（过期/退款），不写链上字段。
	UpdateStatus(ctx context.Context, orderNo, fromStatus, toStatus string, paidAt *time.Time) error
	// MarkPaid CAS fromStatus→paid 并写入链上到账信息（trade_no 唯一索引保证幂等）。
	MarkPaid(ctx context.Context, orderNo, fromStatus string, upd UsdtPaidUpdate, paidAt time.Time) error
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*UsdtOrder, int, error)
	ExpirePendingOrders(ctx context.Context) (int, error)
	ListAll(ctx context.Context, status string, userID *int64, limit, offset int) ([]*UsdtOrder, int, error)
	ListPending(ctx context.Context) ([]*UsdtOrder, error)
	// ListMatchable 返回 pending + 宽限期内刚过期的订单，供 watcher 匹配（防过期后到账孤儿单）。
	ListMatchable(ctx context.Context, graceCutoff time.Time) ([]*UsdtOrder, error)
	// SumPaidCreditByUserIDs 汇总指定用户集合 status='paid' 的 credit_amount（经销商佣金基数）。
	SumPaidCreditByUserIDs(ctx context.Context, userIDs []int64) (float64, error)
	// ListPaidByUserIDs 分页列出指定用户集合 status='paid' 的订单（佣金充值明细）。
	ListPaidByUserIDs(ctx context.Context, userIDs []int64, limit, offset int) ([]*UsdtOrder, int, error)
	// SumPaidCreditByDay 汇总指定时区下每日 status='paid' 的 credit_amount（仪表盘财务趋势）。
	SumPaidCreditByDay(ctx context.Context, startTime, endTime time.Time, tzName string) (map[string]float64, error)
}

// SettingKeyUsdtEnabled USDT 运行时开关 key（与 payment.SettingKeyUsdtEnabled 同值，
// 供 setting_service 暴露公开设置时引用，避免 setting_service 反向依赖 pkg/payment）。
const SettingKeyUsdtEnabled = "usdt_enabled"

var (
	// ErrUsdtOrderNotFound 退款时指定订单号不存在。
	ErrUsdtOrderNotFound = infraerrors.NotFound("USDT_ORDER_NOT_FOUND", "order not found")
	// ErrUsdtOrderNotRefundable 仅 status='paid' 的 USDT 订单可退款。
	ErrUsdtOrderNotRefundable = infraerrors.BadRequest("USDT_ORDER_NOT_REFUNDABLE", "only paid orders can be refunded")
	// ErrUsdtAmountUnavailable 唯一金额池在重试后仍无可用值。
	ErrUsdtAmountUnavailable = errors.New("no available usdt amount")
)
