package payment

import (
	"math"
	"testing"
)

func TestValidateTronAddress(t *testing.T) {
	valid := []string{
		UsdtTRC20Contract,                    // 已知合法的 TRC20 合约地址
		"TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC", // 已知合法的普通地址
	}
	for _, a := range valid {
		if !ValidateTronAddress(a) {
			t.Errorf("expected %s to be a valid TRON address", a)
		}
	}

	invalid := []string{
		"",
		"0x1234567890abcdef",
		"TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeX", // 校验和被篡改
		"BJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC", // 非 T 开头
		"TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDe",  // 长度不足
	}
	for _, a := range invalid {
		if ValidateTronAddress(a) {
			t.Errorf("expected %s to be invalid", a)
		}
	}
}

func TestUsdtAtomicRoundTrip(t *testing.T) {
	cases := []struct {
		usdt   float64
		atomic int64
	}{
		{1, 1_000_000},
		{0.000001, 1},
		{13.891234, 13_891_234},
		{100, 100_000_000},
	}
	for _, c := range cases {
		if got := UsdtToAtomic(c.usdt); got != c.atomic {
			t.Errorf("UsdtToAtomic(%v) = %d, want %d", c.usdt, got, c.atomic)
		}
	}

	// AtomicToUsdt 反向
	f, ok := AtomicToUsdt("13891234")
	if !ok || math.Abs(f-13.891234) > 1e-9 {
		t.Errorf("AtomicToUsdt(13891234) = %v, ok=%v", f, ok)
	}
	if _, ok := AtomicToUsdt("not-a-number"); ok {
		t.Errorf("AtomicToUsdt should reject non-numeric input")
	}
}

func TestFormatUsdt(t *testing.T) {
	if got := FormatUsdt(13.8912); got != "13.891200" {
		t.Errorf("FormatUsdt(13.8912) = %q, want 13.891200", got)
	}
}
