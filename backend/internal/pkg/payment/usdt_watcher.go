package payment

import (
	"context"
	"log"
	"sync"
	"time"
)

// UsdtOrderMatcher 是 UsdtOrderService 实现的接口，供 UsdtMonitor 匹配并确认链上到账。
type UsdtOrderMatcher interface {
	GetPendingUsdtOrders(ctx context.Context) ([]UsdtPendingOrder, error)
	ConfirmUsdtOrderPaid(ctx context.Context, orderNo string, deposit UsdtDeposit) error
}

// UsdtPendingOrder 是一笔待支付的 USDT 订单（供链上金额匹配）。
type UsdtPendingOrder struct {
	OrderNo    string
	Chain      string
	UsdtAmount float64 // 期望的精确应付金额
	UsdtAtomic int64   // 期望金额的 micro-USDT（6 位定点 = UsdtToAtomic(UsdtAmount)），跨链统一匹配
	CreatedAt  time.Time
	ExpiredAt  *time.Time
}

// UsdtDeposit 是匹配成功的链上转账信息。
type UsdtDeposit struct {
	TxID           string
	Chain          string
	FromAddress    string
	PaidUsdtAtomic int64
	PaidUsdt       float64
	BlockTimeMs    int64
}

// UsdtMonitor 轮询各启用链，按唯一金额把 USDT 到账匹配到 pending 订单。
// 始终启动，运行时按 settings.usdt_enabled + per-chain 开关决定是否真正工作。
type UsdtMonitor struct {
	usdt    *UsdtPayment
	matcher UsdtOrderMatcher

	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup

	mu        sync.Mutex
	matchedTx map[string]time.Time // key = chain:txid
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
		log.Println("[UsdtMonitor] Started (gated by settings.usdt_enabled + per-chain switches)")

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
	// 兜底：任一适配器(畸形响应/解码)即便 panic 也不拖垮进程，下轮继续。
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[UsdtMonitor] recovered from panic in runCycle: %v", r)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	m.usdt.Reload(ctx)
	if !m.usdt.IsEnabled(ctx) {
		return
	}
	usable := m.usdt.UsableChains(ctx)
	if len(usable) == 0 {
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

	m.cleanupMatched()

	byChain := make(map[string][]UsdtPendingOrder)
	for _, o := range orders {
		byChain[o.Chain] = append(byChain[o.Chain], o)
	}

	minTs := time.Now().Add(-time.Duration(m.usdt.QueryMinutesBack())*time.Minute - 5*time.Minute).UnixMilli()
	confirmCutoff := time.Now().Add(-m.usdt.ConfirmDuration())
	grace := m.usdt.GraceWindow()

	for _, chain := range usable {
		pend := byChain[chain]
		if len(pend) == 0 {
			continue
		}
		transfers, err := m.usdt.QueryIncoming(ctx, chain, minTs)
		if err != nil {
			log.Printf("[UsdtMonitor] [%s] query transfers failed: %v", chain, err)
			continue
		}
		if len(transfers) == 0 {
			continue
		}

		byAmt := make(map[int64]UsdtPendingOrder, len(pend))
		for _, o := range pend {
			byAmt[o.UsdtAtomic] = o
		}

		for _, tr := range transfers {
			txKey := chain + ":" + tr.TxID
			if m.isMatched(txKey) {
				continue
			}
			hasTime := tr.BlockTimeMs > 0
			blockTime := time.UnixMilli(tr.BlockTimeMs)
			// 等待链上确认：交易需足够「老」（粗粒度 finality 保护）。无时间戳则视为已确认。
			if hasTime && blockTime.After(confirmCutoff) {
				continue
			}
			micro := UsdtToAtomic(tr.AmountHuman)
			order, ok := byAmt[micro]
			if !ok {
				continue
			}
			if hasTime && blockTime.Before(order.CreatedAt.Add(-5*time.Minute)) {
				continue
			}
			// 允许过期后 grace 宽限内的到账仍匹配（订单也已在 GetPendingUsdtOrders 的宽限窗内）。
			if hasTime && order.ExpiredAt != nil && blockTime.After(order.ExpiredAt.Add(grace)) {
				continue
			}

			deposit := UsdtDeposit{
				TxID:           tr.TxID,
				Chain:          chain,
				FromAddress:    tr.From,
				PaidUsdtAtomic: micro,
				PaidUsdt:       tr.AmountHuman,
				BlockTimeMs:    tr.BlockTimeMs,
			}
			log.Printf("[UsdtMonitor] [%s] matched order %s ← tx %s (%.6f USDT)", chain, order.OrderNo, tr.TxID, tr.AmountHuman)
			if err := m.matcher.ConfirmUsdtOrderPaid(ctx, order.OrderNo, deposit); err != nil {
				log.Printf("[UsdtMonitor] [%s] confirm order %s failed: %v", chain, order.OrderNo, err)
				continue
			}
			m.markMatched(txKey)
		}
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

func (m *UsdtMonitor) isMatched(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.matchedTx[key]
	return ok
}

func (m *UsdtMonitor) markMatched(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.matchedTx[key] = time.Now()
}
