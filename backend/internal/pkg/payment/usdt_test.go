package payment

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func newTestUsdt(cfg config.UsdtPaymentConfig) *UsdtPayment {
	u, _ := NewUsdtPayment(cfg, nil)
	return u
}

func TestGraceWindowBounded(t *testing.T) {
	// 默认：grace = max(10min, 2*60s) capped at reuse(30min) = 10min
	u := newTestUsdt(config.UsdtPaymentConfig{})
	if g := u.GraceWindow(); g != 10*time.Minute {
		t.Errorf("default grace = %v, want 10m", g)
	}
	if u.GraceWindow() > u.AmountReuseWindow() {
		t.Errorf("grace must be <= reuse window (else amount-reuse ambiguity)")
	}
	// 大 confirm + 小 reuse：grace 被 reuse 上限钳制
	u2 := newTestUsdt(config.UsdtPaymentConfig{ConfirmSeconds: 3600, OrderTimeoutSeconds: 600, QueryMinutesBack: 5})
	if u2.GraceWindow() != u2.AmountReuseWindow() {
		t.Errorf("grace should be capped at reuse window, grace=%v reuse=%v", u2.GraceWindow(), u2.AmountReuseWindow())
	}
}

func TestQueryRateSanity(t *testing.T) {
	ctx := context.Background()
	// 手动价在合理区间
	if r, err := newTestUsdt(config.UsdtPaymentConfig{ManualRate: 7.2}).QueryRate(ctx); err != nil || math.Abs(r-7.2) > 1e-9 {
		t.Errorf("rate 7.2 → r=%v err=%v", r, err)
	}
	// 手动价超硬上限被拒（防异常汇率近零成本充值）
	if _, err := newTestUsdt(config.UsdtPaymentConfig{ManualRate: 200}).QueryRate(ctx); err == nil {
		t.Errorf("rate 200 should be rejected (out of [1,100] band)")
	}
	// 手动价低于硬下限被拒
	if _, err := newTestUsdt(config.UsdtPaymentConfig{ManualRate: 0.5}).QueryRate(ctx); err == nil {
		t.Errorf("rate 0.5 should be rejected")
	}
	// markup 生效：10 * (1-0.1) = 9
	if r, err := newTestUsdt(config.UsdtPaymentConfig{ManualRate: 10, RateMarkup: 0.1}).QueryRate(ctx); err != nil || math.Abs(r-9) > 1e-9 {
		t.Errorf("markup: want 9, got r=%v err=%v", r, err)
	}
}

func TestPinEndpointFirst(t *testing.T) {
	eps := []string{"a", "b", "c"}
	got := pinEndpointFirst(eps, "b")
	if len(got) != 3 || got[0] != "b" {
		t.Errorf("pin b first: %v", got)
	}
	// 去重：原列表里 b 不再重复
	count := 0
	for _, e := range got {
		if e == "b" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("pinned endpoint should appear once, got %d in %v", count, got)
	}
	if pinEndpointFirst(eps, "")[0] != "a" {
		t.Errorf("empty pin → unchanged order")
	}
}
