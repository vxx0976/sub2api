package payment

import (
	"math"
	"testing"
)

func TestHumanFromAtomic(t *testing.T) {
	cases := []struct {
		atomic   string
		decimals int
		want     float64
	}{
		{"13891200000000000000", 18, 13.8912}, // BEP20 (18-dec wei)
		{"13891200", 6, 13.8912},              // TRC20 / TON (6-dec)
		{"1000000000000000000", 18, 1},
		{"1000000", 6, 1},
		{"500000", 6, 0.5},
	}
	for _, c := range cases {
		got, ok := humanFromAtomic(c.atomic, c.decimals)
		if !ok || math.Abs(got-c.want) > 1e-9 {
			t.Errorf("humanFromAtomic(%q,%d) = %v ok=%v, want %v", c.atomic, c.decimals, got, ok, c.want)
		}
	}
	if _, ok := humanFromAtomic("notnum", 18); ok {
		t.Errorf("humanFromAtomic should reject non-numeric")
	}
}

func TestBscValidateAddress(t *testing.T) {
	a := &bscAdapter{}
	if a.Chain() != ChainBEP20 || a.Decimals() != 18 {
		t.Fatalf("unexpected chain/decimals: %s/%d", a.Chain(), a.Decimals())
	}
	valid := []string{
		UsdtBEP20Contract,
		"0x55d398326f99059fF775485246999027B3197955",
		"0x0000000000000000000000000000000000000000",
	}
	for _, x := range valid {
		if !a.ValidateAddress(x) {
			t.Errorf("expected %s valid", x)
		}
	}
	invalid := []string{"", "0x123", "55d398326f99059fF775485246999027B3197955", "0xZZd398326f99059fF775485246999027B3197955", "TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC"}
	for _, x := range invalid {
		if a.ValidateAddress(x) {
			t.Errorf("expected %s invalid", x)
		}
	}
}

func TestTonValidateAddress(t *testing.T) {
	a := &tonAdapter{}
	if a.Chain() != ChainTON || a.Decimals() != 6 {
		t.Fatalf("unexpected chain/decimals: %s/%d", a.Chain(), a.Decimals())
	}
	valid := []string{
		UsdtTONJettonMaster, // EQ... user-friendly
		"UQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_aBc",
		"0:b113a994b5024a16719f69139328eb759596c38a25f59028b146fecdc3621dfe", // raw
	}
	for _, x := range valid {
		if !a.ValidateAddress(x) {
			t.Errorf("expected %s valid", x)
		}
	}
	invalid := []string{"", "0x55d398326f99059fF775485246999027B3197955", "EQshort", "notbase64!!!"}
	for _, x := range invalid {
		if a.ValidateAddress(x) {
			t.Errorf("expected %s invalid", x)
		}
	}
}

func TestSupportedChains(t *testing.T) {
	for _, c := range []string{ChainTRC20, ChainBEP20, ChainTON} {
		if !IsSupportedChain(c) {
			t.Errorf("%s should be supported", c)
		}
	}
	if IsSupportedChain("erc20") {
		t.Errorf("erc20 should not be supported")
	}
	if UsdtChainSettingKey("trc20", "address") != "usdt_trc20_address" {
		t.Errorf("unexpected setting key: %s", UsdtChainSettingKey("trc20", "address"))
	}
}
