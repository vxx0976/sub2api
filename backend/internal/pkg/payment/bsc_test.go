package payment

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestPadTopicAddress(t *testing.T) {
	addr := "0x55d398326f99059fF775485246999027B3197955"
	got := padTopicAddress(addr)
	want := "0x000000000000000000000000" + "55d398326f99059ff775485246999027b3197955"
	if got != want {
		t.Fatalf("padTopicAddress = %s, want %s", got, want)
	}
	// 反向提取（watcher 解码用的 topics[2][len-40:]）应还原地址（小写）。
	back := "0x" + strings.ToLower(got[len(got)-40:])
	if back != strings.ToLower(addr) {
		t.Errorf("extract back = %s, want %s", back, strings.ToLower(addr))
	}
}

func TestHexConv(t *testing.T) {
	if hexToUint64("0x10") != 16 {
		t.Errorf("hexToUint64(0x10) != 16")
	}
	if hexToUint64("0x0") != 0 {
		t.Errorf("hexToUint64(0x0) != 0")
	}
	// 1 USDT(18-dec) = 1e18 = 0xde0b6b3a7640000
	if got := hexToDecString("0xde0b6b3a7640000"); got != "1000000000000000000" {
		t.Errorf("hexToDecString = %s, want 1000000000000000000", got)
	}
	// 解码到人类金额（BEP20 18 精度）
	human, ok := humanFromAtomic(hexToDecString("0xde0b6b3a7640000"), 18)
	if !ok || math.Abs(human-1.0) > 1e-9 {
		t.Errorf("decode 1 USDT failed: %v ok=%v", human, ok)
	}
	if hexToDecString("0xnothex") != "0" {
		t.Errorf("hexToDecString should return 0 on bad input")
	}
}

func TestBscFromBlock(t *testing.T) {
	// minTs<=0 → head - cap（head 足够大）
	if got := bscFromBlock(100000, 0); got != 100000-bscLookbackCapBlocks {
		t.Errorf("bscFromBlock(100000,0) = %d, want %d", got, 100000-bscLookbackCapBlocks)
	}
	// head 小于 cap → 0
	if got := bscFromBlock(100, 0); got != 0 {
		t.Errorf("bscFromBlock(100,0) = %d, want 0", got)
	}
	// 35 分钟回看 ≈ 705 块，但被 cap(300) 钳制 → 实际回看 = cap
	head := uint64(50_000_000)
	from := bscFromBlock(head, NowMinus(35))
	if blocks := head - from; blocks != bscLookbackCapBlocks {
		t.Errorf("35min lookback blocks = %d, want cap %d", blocks, bscLookbackCapBlocks)
	}
	// 小回看(2分钟≈40块)不触发 cap
	from2 := bscFromBlock(head, NowMinus(2))
	if blocks := head - from2; blocks < 40 || blocks > 60 {
		t.Errorf("2min lookback blocks = %d, want ~45", blocks)
	}
}

// NowMinus 返回 minutes 分钟前的毫秒时间戳（测试辅助）。
func NowMinus(minutes int64) int64 {
	return time.Now().UnixMilli() - minutes*60*1000
}

func TestBscEndpoints(t *testing.T) {
	eps := bscEndpoints("https://custom.rpc/")
	if eps[0] != "https://custom.rpc" {
		t.Errorf("configured baseURL should be first, got %s", eps[0])
	}
	if eps[1] != DefaultBSCRPCURL {
		t.Errorf("second should be default, got %s", eps[1])
	}
	// 配置的就是默认时应去重
	eps2 := bscEndpoints(DefaultBSCRPCURL)
	count := 0
	for _, e := range eps2 {
		if e == DefaultBSCRPCURL {
			count++
		}
	}
	if count != 1 {
		t.Errorf("default endpoint should appear once, got %d", count)
	}
	// 空 baseURL → 直接用 fallback 列表
	if got := bscEndpoints(""); got[0] != DefaultBSCRPCURL {
		t.Errorf("empty baseURL first should be default, got %s", got[0])
	}
}
