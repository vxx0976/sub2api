package payment

import (
	"context"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// OrderMatcher is the interface that services implement to allow the monitor to match and confirm payments
type OrderMatcher interface {
	GetPendingOrders(ctx context.Context) ([]PendingOrder, error)
	ConfirmOrderPaid(ctx context.Context, orderNo string, tradeNo string, payType string) error
}

// PendingOrder represents a pending order for matching
type PendingOrder struct {
	OrderNo       string
	Amount        float64
	PaymentAmount float64
	CreatedAt     time.Time
	ExpiredAt     *time.Time
}

// AlipayMonitor monitors Alipay account for incoming payments
type AlipayMonitor struct {
	alipay       *AlipayPayment
	orderMatcher OrderMatcher
	interval     time.Duration
	queryMinutes int

	stopCh chan struct{}
	wg     sync.WaitGroup

	// Track matched bill IDs to prevent duplicate matching
	mu          sync.Mutex
	matchedBills map[string]time.Time // alipayOrderNo -> matchTime
}

// NewAlipayMonitor creates a new payment monitor
func NewAlipayMonitor(alipay *AlipayPayment, orderMatcher OrderMatcher) *AlipayMonitor {
	interval := time.Duration(alipay.cfg.MonitorIntervalSeconds) * time.Second
	if interval < 5*time.Second {
		interval = 10 * time.Second
	}

	queryMinutes := alipay.cfg.QueryMinutesBack
	if queryMinutes <= 0 {
		queryMinutes = 30
	}

	return &AlipayMonitor{
		alipay:       alipay,
		orderMatcher: orderMatcher,
		interval:     interval,
		queryMinutes: queryMinutes,
		stopCh:       make(chan struct{}),
		matchedBills: make(map[string]time.Time),
	}
}

// Start starts the monitor goroutine
func (m *AlipayMonitor) Start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		log.Println("[AlipayMonitor] Started")

		m.runCycle()

		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.runCycle()
			case <-m.stopCh:
				log.Println("[AlipayMonitor] Stopped")
				return
			}
		}
	}()
}

// Stop stops the monitor
func (m *AlipayMonitor) Stop() {
	close(m.stopCh)
	m.wg.Wait()
}

func (m *AlipayMonitor) runCycle() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orders, err := m.orderMatcher.GetPendingOrders(ctx)
	if err != nil {
		log.Printf("[AlipayMonitor] Failed to get pending orders: %v", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	loc := time.FixedZone("CST", 8*3600)
	now := time.Now().In(loc)
	endTime := now
	startTime := now.Add(-time.Duration(m.queryMinutes)*time.Minute - 5*time.Minute)

	bills, err := m.alipay.QueryAccountBills(ctx, startTime, endTime)
	if err != nil {
		log.Printf("[AlipayMonitor] Failed to query bills: %v", err)
		return
	}

	if len(bills) == 0 {
		return
	}

	// Clean up old matched bill records (older than query window)
	m.cleanupMatchedBills()

	m.matchBills(ctx, orders, bills)
}

func (m *AlipayMonitor) cleanupMatchedBills() {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().Add(-2 * time.Hour)
	for k, t := range m.matchedBills {
		if t.Before(cutoff) {
			delete(m.matchedBills, k)
		}
	}
}

func (m *AlipayMonitor) isMatched(alipayOrderNo string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.matchedBills[alipayOrderNo]
	return ok
}

func (m *AlipayMonitor) markMatched(alipayOrderNo string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.matchedBills[alipayOrderNo] = time.Now()
}

func (m *AlipayMonitor) matchBills(ctx context.Context, orders []PendingOrder, bills []AccountBill) {
	if m.alipay.cfg.Mode == "transfer" {
		m.matchTransferBills(ctx, orders, bills)
	} else {
		m.matchBusinessQRBills(ctx, orders, bills)
	}
}

func (m *AlipayMonitor) matchBusinessQRBills(ctx context.Context, orders []PendingOrder, bills []AccountBill) {
	tolerance := time.Duration(m.alipay.cfg.MatchToleranceSeconds) * time.Second
	if tolerance == 0 {
		tolerance = 5 * time.Minute
	}

	// Sort orders by CreatedAt descending - prefer matching newer orders first
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})

	for _, bill := range bills {
		// Skip already matched bills
		if m.isMatched(bill.AlipayOrderNo) {
			continue
		}

		// Only process income bills
		if bill.Direction != "" && bill.Direction != "收入" {
			continue
		}

		amount := bill.TransAmountFloat()
		if amount <= 0 {
			continue
		}

		// Parse bill time without timezone - DB stores CST times as +0000
		billTime, err := time.Parse("2006-01-02 15:04:05", bill.TransDt)
		if err != nil {
			continue
		}

		for _, order := range orders {
			if math.Abs(amount-order.PaymentAmount) > 0.001 {
				continue
			}

			if billTime.Before(order.CreatedAt.Add(-tolerance)) {
				continue
			}

			if order.ExpiredAt != nil && billTime.After(*order.ExpiredAt) {
				continue
			}

			log.Printf("[AlipayMonitor] Matched order %s: amount=%.2f, bill=%s, billTime=%s", order.OrderNo, amount, bill.AlipayOrderNo, bill.TransDt)
			if err := m.orderMatcher.ConfirmOrderPaid(ctx, order.OrderNo, bill.AlipayOrderNo, "alipay"); err != nil {
				log.Printf("[AlipayMonitor] Failed to confirm order %s: %v", order.OrderNo, err)
			} else {
				m.alipay.ReleaseAmount(order.PaymentAmount)
				m.markMatched(bill.AlipayOrderNo)
			}
			break
		}
	}
}

func (m *AlipayMonitor) matchTransferBills(ctx context.Context, orders []PendingOrder, bills []AccountBill) {
	orderMap := make(map[string]PendingOrder)
	for _, o := range orders {
		orderMap[o.OrderNo] = o
	}

	for _, bill := range bills {
		if m.isMatched(bill.AlipayOrderNo) {
			continue
		}

		amount := bill.TransAmountFloat()
		if amount <= 0 {
			continue
		}

		memo := bill.TransMemo
		for orderNo, order := range orderMap {
			if !strings.Contains(memo, orderNo) {
				continue
			}

			if math.Abs(amount-order.Amount) > 0.1 {
				continue
			}

			log.Printf("[AlipayMonitor] Matched transfer order %s: memo=%s, bill=%s", orderNo, memo, bill.AlipayOrderNo)
			if err := m.orderMatcher.ConfirmOrderPaid(ctx, orderNo, bill.AlipayOrderNo, "alipay"); err != nil {
				log.Printf("[AlipayMonitor] Failed to confirm order %s: %v", orderNo, err)
			} else {
				m.markMatched(bill.AlipayOrderNo)
			}
			delete(orderMap, orderNo)
			break
		}
	}
}
