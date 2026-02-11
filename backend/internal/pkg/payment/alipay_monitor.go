package payment

import (
	"context"
	"log"
	"math"
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

	m.matchBills(ctx, orders, bills)
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

	for _, bill := range bills {
		if bill.TransAmount <= 0 {
			continue
		}

		billTime, err := time.ParseInLocation("2006-01-02 15:04:05", bill.TransDt, time.FixedZone("CST", 8*3600))
		if err != nil {
			continue
		}

		for _, order := range orders {
			if math.Abs(bill.TransAmount-order.PaymentAmount) > 0.001 {
				continue
			}

			if billTime.Before(order.CreatedAt.Add(-tolerance)) {
				continue
			}

			if order.ExpiredAt != nil && billTime.After(*order.ExpiredAt) {
				continue
			}

			log.Printf("[AlipayMonitor] Matched order %s: amount=%.2f, bill=%s", order.OrderNo, bill.TransAmount, bill.AlipayOrderNo)
			if err := m.orderMatcher.ConfirmOrderPaid(ctx, order.OrderNo, bill.AlipayOrderNo, "alipay"); err != nil {
				log.Printf("[AlipayMonitor] Failed to confirm order %s: %v", order.OrderNo, err)
			} else {
				m.alipay.ReleaseAmount(order.PaymentAmount)
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
		if bill.TransAmount <= 0 {
			continue
		}

		memo := bill.TransMemo
		for orderNo, order := range orderMap {
			if !strings.Contains(memo, orderNo) {
				continue
			}

			if math.Abs(bill.TransAmount-order.Amount) > 0.1 {
				continue
			}

			log.Printf("[AlipayMonitor] Matched transfer order %s: memo=%s, bill=%s", orderNo, memo, bill.AlipayOrderNo)
			if err := m.orderMatcher.ConfirmOrderPaid(ctx, orderNo, bill.AlipayOrderNo, "alipay"); err != nil {
				log.Printf("[AlipayMonitor] Failed to confirm order %s: %v", orderNo, err)
			}
			delete(orderMap, orderNo)
			break
		}
	}
}
