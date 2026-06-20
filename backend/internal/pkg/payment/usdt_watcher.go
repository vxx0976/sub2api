package payment

import (
	"context"
	"log"
	"sync"
	"time"
)

// UsdtOrderMatcher 是 UsdtOrderService 实现的接口，供 UsdtMonitor 匹配并确认链上到账。
// 单独定义（不复用 OrderMatcher），避免 wire 把它绑成 AliMPay 的 *OrderService。
type UsdtOrderMatcher interface {
	GetPendingUsdtOrders(ctx context.Context) ([]UsdtPendingOrder, error)
	ConfirmUsdtOrderPaid(ctx context.Context, orderNo string, deposit UsdtDeposit) error
}

// UsdtPendingOrder 是一笔待支付的 USDT 订单（供链上金额匹配）。
type UsdtPendingOrder struct {
	OrderNo    string
	UsdtAmount float64 // 期望的精确应付金额
	UsdtAtomic int64   // 期望的链上最小单位（= UsdtToAtomic(UsdtAmount)）
	CreatedAt  time.Time
	ExpiredAt  *time.Time
}

// UsdtDeposit 是匹配成功的链上转账信息（确认入账时写入订单）。
type UsdtDeposit struct {
	TxID           string
	FromAddress    string
	PaidUsdtAtomic int64
	PaidUsdt       float64
	BlockTimeMs    int64
}

// UsdtMonitor 轮询 TronGrid，按唯一金额把 USDT 到账匹配到 pending 订单。
// 结构镜像 AlipayMonitor：始终启动，运行时按 settings.usdt_enabled 决定是否真正工作。
type UsdtMonitor struct {
	usdt    *UsdtPayment
	matcher UsdtOrderMatcher

	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup

	// 已确认 tx 去重（DB trade_no 唯一索引是最终保证，这里减少无谓重试）
	mu        sync.Mutex
	matchedTx map[string]time.Time
}

// NewUsdtMonitor 创建监控器。
func NewUsdtMonitor(usdt *UsdtPayment, matcher UsdtOrderMatcher) *UsdtMonitor {
	return &UsdtMonitor{
		usdt:      usdt,
		matcher:   matcher,
		interval:  usdt.MonitorInterval(),
		stopCh:    make(chan struct{}),
		matchedTx: make(map[string]time.Time),
	}
}

// Start 启动监控协程。
func (m *UsdtMonitor) Start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		log.Println("[UsdtMonitor] Started (gated by settings.usdt_enabled at runtime)")

		m.runCycle()

		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.runCycle()
				if newInterval := m.usdt.MonitorInterval(); newInterval > 0 && newInterval != m.interval {
					m.interval = newInterval
					ticker.Reset(newInterval)
					log.Printf("[UsdtMonitor] Interval updated to %s", newInterval)
				}
			case <-m.stopCh:
				log.Println("[UsdtMonitor] Stopped")
				return
			}
		}
	}()
}

// Stop 停止监控。
func (m *UsdtMonitor) Stop() {
	close(m.stopCh)
	m.wg.Wait()
}

func (m *UsdtMonitor) runCycle() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 热加载配置
	m.usdt.Reload(ctx)
	if !m.usdt.IsEnabled(ctx) {
		return
	}
	if m.usdt.ReceivingAddress() == "" {
		return
	}

	orders, err := m.matcher.GetPendingUsdtOrders(ctx)
	if err != nil {
		log.Printf("[UsdtMonitor] get pending orders failed: %v", err)
		return
	}
	if len(orders) == 0 {
		return
	}

	minTs := time.Now().Add(-time.Duration(m.usdt.QueryMinutesBack())*time.Minute - 5*time.Minute).UnixMilli()
	transfers, err := m.usdt.QueryIncomingTransfers(ctx, minTs)
	if err != nil {
		log.Printf("[UsdtMonitor] query transfers failed: %v", err)
		return
	}
	if len(transfers) == 0 {
		return
	}

	m.cleanupMatched()

	// 唯一金额 → 订单（pending 订单的 usdt_amount 在 (chain,amount) 上唯一）
	byAtomic := make(map[int64]UsdtPendingOrder, len(orders))
	for _, o := range orders {
		byAtomic[o.UsdtAtomic] = o
	}

	confirmCutoff := time.Now().Add(-m.usdt.ConfirmDuration())

	for _, tr := range transfers {
		if m.isMatched(tr.TxID) {
			continue
		}
		// 等待链上确认：交易需足够「老」才入账（粗粒度的 finality 保护）。
		blockTime := time.UnixMilli(tr.BlockTimeMs)
		if blockTime.After(confirmCutoff) {
			continue
		}
		atomic, ok := new64(tr.ValueAtomic)
		if !ok || atomic <= 0 {
			continue
		}
		order, ok := byAtomic[atomic]
		if !ok {
			log.Printf("[UsdtMonitor] transfer %s (%s USDT atomic) no matching pending order", tr.TxID, tr.ValueAtomic)
			continue
		}
		// 时间窗校验：转账不能早于订单创建（留 5min 容差），不能晚于过期。
		if blockTime.Before(order.CreatedAt.Add(-5 * time.Minute)) {
			continue
		}
		if order.ExpiredAt != nil && blockTime.After(*order.ExpiredAt) {
			continue
		}

		paidUsdt, _ := AtomicToUsdt(tr.ValueAtomic)
		deposit := UsdtDeposit{
			TxID:           tr.TxID,
			FromAddress:    tr.From,
			PaidUsdtAtomic: atomic,
			PaidUsdt:       paidUsdt,
			BlockTimeMs:    tr.BlockTimeMs,
		}
		log.Printf("[UsdtMonitor] matched order %s ← tx %s (%.6f USDT)", order.OrderNo, tr.TxID, paidUsdt)
		if err := m.matcher.ConfirmUsdtOrderPaid(ctx, order.OrderNo, deposit); err != nil {
			log.Printf("[UsdtMonitor] confirm order %s failed: %v", order.OrderNo, err)
			continue
		}
		m.markMatched(tr.TxID)
	}
}

func (m *UsdtMonitor) cleanupMatched() {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().Add(-2 * time.Hour)
	for k, t := range m.matchedTx {
		if t.Before(cutoff) {
			delete(m.matchedTx, k)
		}
	}
}

func (m *UsdtMonitor) isMatched(tx string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.matchedTx[tx]
	return ok
}

func (m *UsdtMonitor) markMatched(tx string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.matchedTx[tx] = time.Now()
}

// new64 把十进制字符串解析为 int64（链上 USDT 金额最小单位不会溢出 int64）。
func new64(s string) (int64, bool) {
	var n int64
	if s == "" {
		return 0, false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int64(r-'0')
		if n < 0 { // 溢出
			return 0, false
		}
	}
	return n, true
}
