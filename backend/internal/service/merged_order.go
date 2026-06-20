package service

import (
	"context"
	"sort"
	"strconv"
	"time"
)

// mergedFetchCap 单通道拉取上限：合并分页需要先把每个通道的前 page*page_size 行取回内存再排序切片，
// 这里给一个硬上限防止超大 page 把整张表拉进内存。超过则 clamp。
const mergedFetchCap = 5000

// MergedTopupChannel 通道标识。
const (
	MergedChannelRecharge = "recharge"
	MergedChannelAliMPay  = "alimpay"
	MergedChannelUsdt     = "usdt"
)

// MergedOrder 三个自建充值通道（EPAY 充值 / AliMPay 个人免签 / USDT 多链）订单的归一化视图。
// 金额统一格式化为字符串（与各通道处理器保持一致的精度），通道特有字段在不适用时为 nil。
type MergedOrder struct {
	Channel      string  `json:"channel"` // recharge | alimpay | usdt
	ID           int64   `json:"id"`      // 各通道表内行 id（非全局唯一）
	OrderNo      string  `json:"order_no"`
	TradeNo      string  `json:"trade_no"`
	UserID       int64   `json:"user_id"`
	UserEmail    string  `json:"user_email"`    // 仅 admin 端填充，user 端为 ""
	Amount       string  `json:"amount"`        // CNY，十进制字符串
	CreditAmount string  `json:"credit_amount"` // USD 到账，十进制字符串
	Status       string  `json:"status"`        // pending|paid|expired|refunded
	PayType      string  `json:"pay_type"`      // alipay|wxpay|usdt
	CreatedAt    string  `json:"created_at"`    // RFC3339Nano
	PaidAt       *string `json:"paid_at"`
	ExpiredAt    *string `json:"expired_at"`

	// 通道特有字段：不适用时为 nil。
	PaymentAmount *string `json:"payment_amount"`  // 仅 alimpay（唯一偏移金额）
	UsdtAmountStr *string `json:"usdt_amount_str"` // 仅 usdt
	UsdtRate      *string `json:"usdt_rate"`       // 仅 usdt
	UsdtChain     *string `json:"usdt_chain"`      // 仅 usdt（如 trc20）
	SourceDomain  *string `json:"source_domain"`   // alimpay + usdt

	// 排序键：不参与 JSON 序列化。
	createdAt time.Time
}

// MergedOrderFilter 合并列表的过滤条件。
type MergedOrderFilter struct {
	Channel string // "" = 全部；recharge | alimpay | usdt
	Status  string // "" = 全部；pending | paid | expired | refunded
	UserID  *int64 // nil = 全部用户（仅 admin）；非 nil = 限定该用户
}

// MergedOrderService 聚合三个自建充值通道服务，提供按时间倒序合并的订单列表。
type MergedOrderService struct {
	rechargeService *RechargeService
	orderService    *OrderService
	usdtService     *UsdtOrderService
	adminService    AdminService
}

// NewMergedOrderService 构造合并订单服务。
func NewMergedOrderService(
	rechargeService *RechargeService,
	orderService *OrderService,
	usdtService *UsdtOrderService,
	adminService AdminService,
) *MergedOrderService {
	return &MergedOrderService{
		rechargeService: rechargeService,
		orderService:    orderService,
		usdtService:     usdtService,
		adminService:    adminService,
	}
}

// ListMerged 返回符合过滤条件、跨通道按 created_at 倒序合并后的当前页订单及总数。
// withEmail 为 true 时填充 user_email（admin 端）。
func (s *MergedOrderService) ListMerged(ctx context.Context, f MergedOrderFilter, page, pageSize int, withEmail bool) ([]*MergedOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	// 每个在范围内的通道都要先取回前 page*page_size 行，合并后才能正确切片。
	fetchLimit := page * pageSize
	if fetchLimit > mergedFetchCap {
		fetchLimit = mergedFetchCap
	}

	want := func(channel string) bool {
		return f.Channel == "" || f.Channel == channel
	}

	all := make([]*MergedOrder, 0, fetchLimit*3)
	var total int64

	if want(MergedChannelRecharge) {
		orders, cnt, err := s.rechargeService.ListAllOrders(ctx, f.Status, f.UserID, 1, fetchLimit)
		if err != nil {
			return nil, 0, err
		}
		total += int64(cnt)
		for _, o := range orders {
			all = append(all, mapRechargeOrder(o))
		}
	}

	if want(MergedChannelAliMPay) {
		orders, cnt, err := s.orderService.ListAllOrders(ctx, f.Status, f.UserID, 1, fetchLimit)
		if err != nil {
			return nil, 0, err
		}
		total += int64(cnt)
		for _, o := range orders {
			all = append(all, mapAliMPayOrder(o))
		}
	}

	if want(MergedChannelUsdt) {
		orders, cnt, err := s.usdtService.ListAllOrders(ctx, f.Status, f.UserID, 1, fetchLimit)
		if err != nil {
			return nil, 0, err
		}
		total += int64(cnt)
		for _, o := range orders {
			all = append(all, mapUsdtOrder(o))
		}
	}

	// 按 created_at 倒序；同刻用 (channel, id) 兜底稳定排序。
	sort.SliceStable(all, func(i, j int) bool {
		if !all[i].createdAt.Equal(all[j].createdAt) {
			return all[i].createdAt.After(all[j].createdAt)
		}
		if all[i].Channel != all[j].Channel {
			return all[i].Channel < all[j].Channel
		}
		return all[i].ID > all[j].ID
	})

	// 内存切片当前页。
	// 超过 mergedFetchCap 的深翻页无法保证全局正确（每个通道只取回最新 mergedFetchCap 行，
	// 更深的页可能落在被截断的尾部），直接返回空页，避免静默返回错位数据。
	// 这是极端场景（单通道 > 5000 单且翻到很深页）；total 仍为真实计数。
	start := (page - 1) * pageSize
	if start >= len(all) || start >= mergedFetchCap {
		return []*MergedOrder{}, total, nil
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	pageItems := all[start:end]

	if withEmail {
		s.fillEmails(ctx, pageItems)
	}

	return pageItems, total, nil
}

// fillEmails 批量回填当前页订单的 user_email。
func (s *MergedOrderService) fillEmails(ctx context.Context, items []*MergedOrder) {
	emailMap := make(map[int64]string)
	for _, it := range items {
		if _, seen := emailMap[it.UserID]; seen {
			continue
		}
		if u, err := s.adminService.GetUser(ctx, it.UserID); err == nil && u != nil {
			emailMap[it.UserID] = u.Email
		} else {
			emailMap[it.UserID] = ""
		}
	}
	for _, it := range items {
		it.UserEmail = emailMap[it.UserID]
	}
}

// --- 金额/时间格式化 ---

// fmtCNY 格式化 CNY 金额，去除多余尾零（与各通道处理器把 float64 直接序列化的取值一致）。
func fmtCNY(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func fmtTime(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func fmtTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339Nano)
	return &s
}

func mergedStrPtr(s string) *string { return &s }

// --- 各通道 → 归一化映射 ---

func mapRechargeOrder(o *RechargeOrder) *MergedOrder {
	return &MergedOrder{
		Channel:      MergedChannelRecharge,
		ID:           o.ID,
		OrderNo:      o.OrderNo,
		TradeNo:      o.TradeNo,
		UserID:       o.UserID,
		Amount:       fmtCNY(o.Amount),
		CreditAmount: fmtCNY(o.CreditAmount),
		Status:       o.Status,
		PayType:      o.PayType,
		CreatedAt:    fmtTime(o.CreatedAt),
		PaidAt:       fmtTimePtr(o.PaidAt),
		ExpiredAt:    fmtTimePtr(o.ExpiredAt),
		createdAt:    o.CreatedAt,
	}
}

func mapAliMPayOrder(o *Order) *MergedOrder {
	// 与 OrderHandler.effectivePaymentAmount 保持一致：transfer 模式 payment_amount 为 0 时回退到 amount。
	pay := o.PaymentAmount
	if pay <= 0 {
		pay = o.Amount
	}
	return &MergedOrder{
		Channel:       MergedChannelAliMPay,
		ID:            o.ID,
		OrderNo:       o.OrderNo,
		TradeNo:       o.TradeNo,
		UserID:        o.UserID,
		Amount:        fmtCNY(o.Amount),
		CreditAmount:  fmtCNY(o.CreditAmount),
		Status:        o.Status,
		PayType:       o.PayType,
		CreatedAt:     fmtTime(o.CreatedAt),
		PaidAt:        fmtTimePtr(o.PaidAt),
		ExpiredAt:     fmtTimePtr(o.ExpiredAt),
		PaymentAmount: mergedStrPtr(fmtCNY(pay)),
		SourceDomain:  mergedStrPtr(o.SourceDomain),
		createdAt:     o.CreatedAt,
	}
}

func mapUsdtOrder(o *UsdtOrder) *MergedOrder {
	return &MergedOrder{
		Channel:       MergedChannelUsdt,
		ID:            o.ID,
		OrderNo:       o.OrderNo,
		TradeNo:       o.TradeNo,
		UserID:        o.UserID,
		Amount:        fmtCNY(o.Amount),
		CreditAmount:  fmtCNY(o.CreditAmount),
		Status:        o.Status,
		PayType:       o.PayType,
		CreatedAt:     fmtTime(o.CreatedAt),
		PaidAt:        fmtTimePtr(o.PaidAt),
		ExpiredAt:     fmtTimePtr(o.ExpiredAt),
		UsdtAmountStr: mergedStrPtr(strconv.FormatFloat(o.UsdtAmount, 'f', 6, 64)),
		UsdtRate:      mergedStrPtr(fmtCNY(o.UsdtRate)),
		UsdtChain:     mergedStrPtr(o.Chain),
		SourceDomain:  mergedStrPtr(o.SourceDomain),
		createdAt:     o.CreatedAt,
	}
}
