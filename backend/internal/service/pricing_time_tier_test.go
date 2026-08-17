//go:build unit

package service

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 本文件的用例不触碰任何全局状态（尤其**不调用 timezone.Init**，它会写全局 time.Local
// 并与遗留后台 goroutine 的 time.Now() 构成 -race 竞态）。
//
// 包级 init 在 group_peak_rate_test.go 里已把全局时区固定为 UTC；本文件的判定全部基于
// 北京时间，UTC 环境下依然全绿正是「档位判定不受服务器时区影响」的证明。

// 时刻构造 helper bj() 定义在 pricing_service_test.go（无 build tag，两种构建下都可用）。

func TestBandAt_OfficialDeepSeekBoundaries(t *testing.T) {
	tests := []struct {
		name string
		h, m,
		s int
		band   string
		factor float64
	}{
		{"00:00 空闲", 0, 0, 0, PricingBandOffPeak, 0.5},
		{"08:59:59 仍是空闲（秒被忽略）", 8, 59, 59, PricingBandOffPeak, 0.5},
		{"09:00:00 高峰起点（左闭）", 9, 0, 0, PricingBandPeak, 1.0},
		{"09:00:01 高峰", 9, 0, 1, PricingBandPeak, 1.0},
		{"10:30 高峰", 10, 30, 0, PricingBandPeak, 1.0},
		{"11:59:59 高峰末刻", 11, 59, 59, PricingBandPeak, 1.0},
		{"12:00:00 高峰终点（右开）→ 空闲", 12, 0, 0, PricingBandOffPeak, 0.5},
		{"12:00:01 空闲", 12, 0, 1, PricingBandOffPeak, 0.5},
		{"13:59:59 空闲", 13, 59, 59, PricingBandOffPeak, 0.5},
		{"14:00:00 第二段高峰起点", 14, 0, 0, PricingBandPeak, 1.0},
		{"16:00 高峰", 16, 0, 0, PricingBandPeak, 1.0},
		{"17:59:59 高峰末刻", 17, 59, 59, PricingBandPeak, 1.0},
		{"18:00:00 高峰终点 → 空闲", 18, 0, 0, PricingBandOffPeak, 0.5},
		{"23:59:59 空闲", 23, 59, 59, PricingBandOffPeak, 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			band, factor := deepSeekOfficialSchedule.bandAt(bj(t, tt.h, tt.m, tt.s))
			require.Equal(t, tt.band, band)
			require.InDelta(t, tt.factor, factor, 1e-15)
		})
	}
}

// 空闲段 18:00 → 次日 09:00 跨天连续：不需要跨天区间表达，枚举高峰、其余即空闲。
func TestBandAt_OffPeakSpansMidnight(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	nextDay := time.Date(2026, 8, 18, 0, 0, 0, 0, loc)
	band, factor := deepSeekOfficialSchedule.bandAt(nextDay)
	require.Equal(t, PricingBandOffPeak, band)
	require.InDelta(t, 0.5, factor, 1e-15)
}

// 传入 UTC 时刻也必须按北京时间判定：UTC 04:00 == 北京 12:00 → 空闲。
func TestBandAt_UTCInputRespectsBeijing(t *testing.T) {
	band, _ := deepSeekOfficialSchedule.bandAt(time.Date(2026, 8, 17, 4, 0, 0, 0, time.UTC))
	require.Equal(t, PricingBandOffPeak, band, "UTC 04:00 = 北京 12:00，应为空闲档")

	band2, _ := deepSeekOfficialSchedule.bandAt(time.Date(2026, 8, 17, 3, 59, 0, 0, time.UTC))
	require.Equal(t, PricingBandPeak, band2, "UTC 03:59 = 北京 11:59，应为高峰档")
}

// 档位判定只看「绝对时刻」，与该时刻用哪个时区表示、以及进程全局时区都无关
// （时区硬锚 Asia/Shanghai，不跟 timezone.Location）。
//
// ⚠️ 刻意**不调用 timezone.Init** 来验证这一点：Init 会写进程全局 time.Local，
// 而本包早先用例遗留的后台 goroutine 仍在调 time.Now()，`go test -race` 下会真报
// DATA RACE（实测 Write=timezone.go 的 time.Local=loc，Read=content_moderation 的
// worker goroutine）。bandAt 本来就不读 timezone.Location()，把同一绝对时刻换成
// 不同时区表示即可证明锚定性，无需触碰任何全局状态。
func TestBandAt_IgnoresServerTimezone(t *testing.T) {
	at := bj(t, 10, 0, 0) // 北京 10:00 → 高峰
	wantBand, wantFactor := deepSeekOfficialSchedule.bandAt(at)
	require.Equal(t, PricingBandPeak, wantBand)

	for _, name := range []string{"America/New_York", "UTC", "Australia/Sydney", "Asia/Kolkata"} {
		loc, err := time.LoadLocation(name)
		require.NoError(t, err)
		// 同一绝对时刻，换个时区表示（Unix 时间戳不变）
		band, factor := deepSeekOfficialSchedule.bandAt(at.In(loc))
		require.Equalf(t, wantBand, band, "以 %s 表示同一时刻不应改变档位判定", name)
		require.InDelta(t, wantFactor, factor, 1e-15)
	}

	// 一个只在错误实现下才会分叉的样本：北京 20:00（空闲）在纽约是同一天早上，
	// 若实现误用了本地/系统时区就会判成高峰。
	off := bj(t, 20, 0, 0)
	ny, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	band, _ := deepSeekOfficialSchedule.bandAt(off.In(ny))
	require.Equal(t, PricingBandOffPeak, band)
}

// 一切异常输入都必须降级到「基准价、不标档位」，且不 panic。
func TestBandAt_SafeDegradation(t *testing.T) {
	at := bj(t, 20, 0, 0) // 正常应为空闲档

	var nilSchedule *timeTierSchedule
	band, factor := nilSchedule.bandAt(at)
	require.Equal(t, "", band)
	require.Equal(t, 1.0, factor)

	// 零值时刻 = 未接线的调用方 → 基准价
	band, factor = deepSeekOfficialSchedule.bandAt(time.Time{})
	require.Equal(t, "", band)
	require.Equal(t, 1.0, factor)

	bad := []struct {
		name string
		s    *timeTierSchedule
	}{
		{"空窗口", &timeTierSchedule{locFn: deepSeekPricingLocation, offPeakFactor: 0.5}},
		{"start>=end", &timeTierSchedule{
			locFn: deepSeekPricingLocation, offPeakFactor: 0.5,
			peakWindows: []timeTierWindow{{startMin: 600, endMin: 600}},
		}},
		{"越界窗口", &timeTierSchedule{
			locFn: deepSeekPricingLocation, offPeakFactor: 0.5,
			peakWindows: []timeTierWindow{{startMin: -1, endMin: 600}},
		}},
		{"系数 0", &timeTierSchedule{
			locFn: deepSeekPricingLocation, offPeakFactor: 0,
			peakWindows: []timeTierWindow{{startMin: 540, endMin: 720}},
		}},
		{"系数负数", &timeTierSchedule{
			locFn: deepSeekPricingLocation, offPeakFactor: -1,
			peakWindows: []timeTierWindow{{startMin: 540, endMin: 720}},
		}},
		{"系数 >1（会导致多收）", &timeTierSchedule{
			locFn: deepSeekPricingLocation, offPeakFactor: 1.5,
			peakWindows: []timeTierWindow{{startMin: 540, endMin: 720}},
		}},
		{"系数 NaN", &timeTierSchedule{
			locFn: deepSeekPricingLocation, offPeakFactor: math.NaN(),
			peakWindows: []timeTierWindow{{startMin: 540, endMin: 720}},
		}},
		{"locFn 为 nil", &timeTierSchedule{
			offPeakFactor: 0.5,
			peakWindows:   []timeTierWindow{{startMin: 540, endMin: 720}},
		}},
		{"locFn 返回 nil", &timeTierSchedule{
			locFn: func() *time.Location { return nil }, offPeakFactor: 0.5,
			peakWindows: []timeTierWindow{{startMin: 540, endMin: 720}},
		}},
	}
	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			band, factor := tt.s.bandAt(at)
			require.Equal(t, "", band, "异常配置不得标注档位")
			require.Equal(t, 1.0, factor, "异常配置必须降级到基准价（最贵档）")
		})
	}
}

// 🔒 核心不变量：¥ 表里的数字恒为最贵档，任何 schedule 的系数都不得 > 1。
// 它守的是「系数整体调大/表价不再是最贵档」这一类改动（走零值时刻的既有断言对此免疫）。
// 注意它**不**覆盖「peak/offpeak 两个返回值互换」——那由
// TestBandAt_OfficialDeepSeekBoundaries 的 band+factor 成对断言负责。
func TestTimeTierFactorNeverExceedsOne(t *testing.T) {
	tables := map[string]map[string]cnyModelPricing{
		"deepseek": deepSeekPricingTable,
		"kimi":     kimiMoonshotPricingTable,
		"qwen":     qwenPricingTable,
	}
	// 覆盖一整天的每个整点与半点，确保所有窗口都被走到。
	for name, table := range tables {
		for model, row := range table {
			if row.schedule == nil {
				continue
			}
			require.Greaterf(t, row.schedule.factor(), 0.0, "%s/%s 系数必须为正", name, model)
			require.LessOrEqualf(t, row.schedule.factor(), 1.0, "%s/%s 系数不得 > 1（表价必须是最贵档）", name, model)
			for minutes := 0; minutes < 24*60; minutes += 30 {
				at := bj(t, minutes/60, minutes%60, 0)
				band, factor := row.schedule.bandAt(at)
				require.Greaterf(t, factor, 0.0, "%s/%s @%02d:%02d 系数必须为正", name, model, minutes/60, minutes%60)
				require.LessOrEqualf(t, factor, 1.0,
					"%s/%s @%02d:%02d 系数 %v > 1，表价不再是最贵档", name, model, minutes/60, minutes%60, factor)
				require.Containsf(t, []string{PricingBandPeak, PricingBandOffPeak}, band,
					"%s/%s @%02d:%02d 档位取值非法: %q", name, model, minutes/60, minutes%60, band)
			}
		}
	}
}

func TestDeepSeekPricingLocation_IsUTC8(t *testing.T) {
	loc := deepSeekPricingLocation()
	require.NotNil(t, loc)
	// 中国自 1991 年后不使用夏令时，全年恒为 UTC+8，故无需 DST 用例。
	_, offset := time.Date(2026, 1, 15, 12, 0, 0, 0, loc).Zone()
	require.Equal(t, 8*60*60, offset, "冬季应为 UTC+8")
	_, offset = time.Date(2026, 7, 15, 12, 0, 0, 0, loc).Zone()
	require.Equal(t, 8*60*60, offset, "夏季应为 UTC+8（中国无夏令时）")
}

func TestTimeTierSchedule_DisplayHelpers(t *testing.T) {
	require.Equal(t, []string{"09:00-12:00", "14:00-18:00"}, deepSeekOfficialSchedule.windowLabels())
	require.Equal(t, "UTC+08:00", deepSeekOfficialSchedule.timezoneLabel())
	require.InDelta(t, 0.5, deepSeekOfficialSchedule.factor(), 1e-15)

	var nilSchedule *timeTierSchedule
	require.Nil(t, nilSchedule.windowLabels())
	require.Equal(t, "", nilSchedule.timezoneLabel())
	require.Equal(t, 0.0, nilSchedule.factor())
}

// ---- 展示层：ModelTimeTierInfo ----

func TestModelTimeTierInfo_DeepSeekReportsWindowsAndCurrentBand(t *testing.T) {
	info := ModelTimeTierInfo("deepseek-v4-flash", bj(t, 20, 0, 0))
	require.NotNil(t, info)
	require.Equal(t, []string{"09:00-12:00", "14:00-18:00"}, info.PeakWindows)
	require.Equal(t, "UTC+08:00", info.Timezone)
	require.InDelta(t, 0.5, info.OffPeakFactor, 1e-15)
	require.Equal(t, PricingBandOffPeak, info.CurrentBand, "当前档必须由后端按官方时区判定")

	peak := ModelTimeTierInfo("deepseek/DeepSeek-V4-Pro", bj(t, 10, 0, 0))
	require.NotNil(t, peak, "大小写与 provider 前缀变体都应识别")
	require.Equal(t, PricingBandPeak, peak.CurrentBand)
}

func TestModelTimeTierInfo_NilForModelsWithoutSchedule(t *testing.T) {
	for _, model := range []string{"kimi-k3", "qwen3-max", "claude-sonnet-4-6", "gpt-5.4", ""} {
		require.Nilf(t, ModelTimeTierInfo(model, bj(t, 20, 0, 0)), "%s 无时段分档，应返回 nil", model)
	}
}

// 🔒 展示口径必须与计费口径一致：admin 覆盖价是绝对价，官方峰谷对该模型已失效，
// 展示层就不能再画两档，否则页面价与实际成交价对不上（计费侧有
// TestPricingBandPrecedence_OverrideTableWins 守，展示侧靠本用例）。
func TestModelTimeTierInfo_NilWhenOverriddenByAdmin(t *testing.T) {
	ps := newCNYPricingService(1.0)
	ps.SetSettingRepository(&writableSettingRepo{store: map[string]string{
		SettingKeyModelPricingOverrides: `[{"model":"deepseek-v4","currency":"CNY","input":1,"output":2,"enabled":true}]`,
	}})
	prev := currentPricingService.Swap(ps)
	t.Cleanup(func() { currentPricingService.Store(prev) })

	require.Nil(t, ModelTimeTierInfo("deepseek-v4-pro", bj(t, 20, 0, 0)),
		"命中覆盖表的模型不得再对外展示官方峰谷两档")
	require.NotNil(t, ModelTimeTierInfo("deepseek-chat", bj(t, 20, 0, 0)),
		"未被覆盖命中的 deepseek 模型仍应展示分档")
}

// ---- 展示层：内置价目表（admin「模型定价」页）----

// 🔒 内置表的数值必须恒为基准价（最贵档），与查询时刻无关。
// 否则 admin 页「用内置价预填覆盖表」会在空闲时段把半价固化成**永久**覆盖价。
func TestListBuiltinPricing_DeepSeekAlwaysPeakPriceWithTierMeta(t *testing.T) {
	svc := newCNYPricingService(1.0)
	entries := svc.ListBuiltinPricing()

	byModel := make(map[string]BuiltinPricingEntry, len(entries))
	for _, e := range entries {
		byModel[e.Model] = e
	}

	flash, ok := byModel["deepseek-v4-flash"]
	require.True(t, ok)
	require.Equal(t, CurrencyCNY, flash.Currency)
	require.InDelta(t, 3.0, flash.InputPerM, 1e-12, "内置表恒为高峰价，不随时刻变化")
	require.InDelta(t, 9.0, flash.OutputPerM, 1e-12)
	require.InDelta(t, 0.10, flash.CachePerM, 1e-12)
	require.InDelta(t, 0.5, flash.OffPeakFactor, 1e-12)
	require.Equal(t, []string{"09:00-12:00", "14:00-18:00"}, flash.PeakWindows)
	require.Equal(t, "UTC+08:00", flash.TimeTierTZ)

	// 无分档的国产模型不得凭空长出时段字段
	kimi, ok := byModel["kimi-k3"]
	require.True(t, ok)
	require.Zero(t, kimi.OffPeakFactor)
	require.Nil(t, kimi.PeakWindows)
	require.Equal(t, "", kimi.TimeTierTZ)
}
