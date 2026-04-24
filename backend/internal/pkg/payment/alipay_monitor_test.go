package payment

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type testOrderMatcher struct {
	confirmed []string
}

func (m *testOrderMatcher) GetPendingOrders(context.Context) ([]PendingOrder, error) {
	return nil, nil
}

func (m *testOrderMatcher) ConfirmOrderPaid(_ context.Context, orderNo string, _, _ string) error {
	m.confirmed = append(m.confirmed, orderNo)
	return nil
}

func TestAlipayMonitorTransferMatchRequiresIncomeAndOrderWindow(t *testing.T) {
	alipay, err := NewAlipayPayment(config.AlipayPaymentConfig{
		Mode:                  "transfer",
		MatchToleranceSeconds: 300,
	}, nil)
	require.NoError(t, err)

	matcher := &testOrderMatcher{}
	monitor := NewAlipayMonitor(alipay, matcher)
	loc := time.FixedZone("CST", 8*3600)
	createdAt := time.Date(2026, 4, 24, 10, 0, 0, 0, loc)
	expiredAt := createdAt.Add(10 * time.Minute)
	order := PendingOrder{
		OrderNo:   "A20260424100000000001",
		Amount:    10,
		CreatedAt: createdAt,
		ExpiredAt: &expiredAt,
	}

	monitor.matchTransferBills(context.Background(), []PendingOrder{order}, []AccountBill{
		{
			AlipayOrderNo: "expense",
			TransAmount:   "10.00",
			Direction:     "支出",
			TransMemo:     order.OrderNo,
			TransDt:       createdAt.Add(time.Minute).Format("2006-01-02 15:04:05"),
		},
		{
			AlipayOrderNo: "too-early",
			TransAmount:   "10.00",
			Direction:     "收入",
			TransMemo:     order.OrderNo,
			TransDt:       createdAt.Add(-10 * time.Minute).Format("2006-01-02 15:04:05"),
		},
		{
			AlipayOrderNo: "too-late",
			TransAmount:   "10.00",
			Direction:     "收入",
			TransMemo:     order.OrderNo,
			TransDt:       expiredAt.Add(time.Second).Format("2006-01-02 15:04:05"),
		},
		{
			AlipayOrderNo: "valid",
			TransAmount:   "10.00",
			Direction:     "收入",
			TransMemo:     order.OrderNo,
			TransDt:       createdAt.Add(time.Minute).Format("2006-01-02 15:04:05"),
		},
	})

	require.Equal(t, []string{order.OrderNo}, matcher.confirmed)
}

type mapSettingGetter struct {
	mu     sync.RWMutex
	values map[string]string
}

func (g *mapSettingGetter) GetValue(_ context.Context, key string) (string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.values[key], nil
}

func (g *mapSettingGetter) set(key, value string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.values[key] = value
}

func TestAlipayPaymentSnapshotAvoidsConcurrentReloadRaces(t *testing.T) {
	settings := &mapSettingGetter{values: map[string]string{
		SettingKeyAliMPayMode:           "business_qr",
		SettingKeyAliMPayBusinessQRURL:  "https://qr.alipay.com/example",
		SettingKeyAliMPayTransferUserID: "2088000000000000",
	}}
	alipay, err := NewAlipayPayment(config.AlipayPaymentConfig{
		Mode:          "business_qr",
		BusinessQRURL: "https://qr.alipay.com/fallback",
	}, settings)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				settings.set(SettingKeyAliMPayMode, "business_qr")
			} else {
				settings.set(SettingKeyAliMPayMode, "transfer")
			}
			alipay.Reload(context.Background())
		}(i)
		go func() {
			defer wg.Done()
			_, _ = alipay.GeneratePaymentInfo("A20260424100000000001", 10)
			_ = alipay.Mode()
			_ = alipay.AmountOffset()
			_ = alipay.QueryMinutesBack()
		}()
	}
	wg.Wait()
}
