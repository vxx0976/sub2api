package service

import (
	"testing"
	"time"
)

// DeepSeek 官方口径：高峰窗口仅工作日生效，周六/周日（北京时间）全天低谷。
//
// 上游在本轮也实现了同一业务规则（deepseekPeakMultiplierAt），但基准价口径相反——
// 它的常量是低谷价、高峰 = ×2.0，漏接线会静默少收；本 fork 保持「表价 = 最贵档、
// 空闲 = ×0.5」，漏接线只会多收（可发现可退款，见 pricing_time_tier.go 头注释）。
// 站长决定保留 fork 的 ¥ 价与口径，只吸收上游「周末全天低谷」这一项正确性改进。
func TestDeepSeekScheduleWeekendIsAlwaysOffPeak(t *testing.T) {
	loc := deepSeekPricingLocation()
	if loc == nil {
		t.Fatal("deepSeekPricingLocation 返回 nil")
	}
	bj := func(y int, m time.Month, d, hh, mm int) time.Time {
		return time.Date(y, m, d, hh, mm, 0, 0, loc)
	}

	// 2026-08-24 周一、08-22 周六、08-23 周日
	for _, tc := range []struct {
		name string
		at   time.Time
		band string
		fac  float64
	}{
		// 工作日：高峰窗口内为峰价
		{"周一 09:00 高峰起点", bj(2026, 8, 24, 9, 0), PricingBandPeak, 1.0},
		{"周一 11:59 高峰内", bj(2026, 8, 24, 11, 59), PricingBandPeak, 1.0},
		{"周一 15:00 第二个高峰", bj(2026, 8, 24, 15, 0), PricingBandPeak, 1.0},
		// 工作日：窗口外为空闲
		{"周一 08:59 高峰前", bj(2026, 8, 24, 8, 59), PricingBandOffPeak, 0.5},
		{"周一 12:00 高峰后", bj(2026, 8, 24, 12, 0), PricingBandOffPeak, 0.5},
		{"周一 23:00 深夜", bj(2026, 8, 24, 23, 0), PricingBandOffPeak, 0.5},
		// 周末：即便落在高峰时刻也必须是空闲
		{"周六 10:00（高峰时刻但周末）", bj(2026, 8, 22, 10, 0), PricingBandOffPeak, 0.5},
		{"周六 15:00（高峰时刻但周末）", bj(2026, 8, 22, 15, 0), PricingBandOffPeak, 0.5},
		{"周日 09:00（高峰起点但周末）", bj(2026, 8, 23, 9, 0), PricingBandOffPeak, 0.5},
		{"周日 17:59（高峰内但周末）", bj(2026, 8, 23, 17, 59), PricingBandOffPeak, 0.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			band, fac := deepSeekOfficialSchedule.bandAt(tc.at)
			if band != tc.band || fac != tc.fac {
				t.Errorf("bandAt = (%q, %v)，期望 (%q, %v)", band, fac, tc.band, tc.fac)
			}
		})
	}
}

// 周末判定必须按 schedule 自己的时区（北京时间），不能跟随部署时区。
// 同一个 UTC 时刻在 UTC 下是周五、在北京时间已是周六，必须按后者判成周末。
func TestDeepSeekWeekendUsesScheduleTimezoneNotDeployTimezone(t *testing.T) {
	// 2026-08-21 20:00 UTC = 2026-08-22 04:00 北京时间（周六）
	at := time.Date(2026, 8, 21, 20, 0, 0, 0, time.UTC)
	if at.UTC().Weekday() != time.Friday {
		t.Fatalf("前置条件不成立：UTC 侧应为周五，实际 %v", at.UTC().Weekday())
	}
	band, fac := deepSeekOfficialSchedule.bandAt(at)
	if band != PricingBandOffPeak || fac != 0.5 {
		t.Errorf("北京时间已是周六，应判空闲档；bandAt = (%q, %v)", band, fac)
	}
}

// 未声明 peakWeekdaysOnly 的 schedule 行为不得改变（周末仍按窗口判定）。
func TestScheduleWithoutWeekdayOnlyFlagUnchangedOnWeekend(t *testing.T) {
	s := &timeTierSchedule{
		locFn:         deepSeekPricingLocation,
		tzLabel:       "UTC+08:00",
		peakWindows:   []timeTierWindow{{startMin: 9 * 60, endMin: 12 * 60}},
		offPeakFactor: 0.5,
		// 刻意不设 peakWeekdaysOnly
	}
	loc := deepSeekPricingLocation()
	sat := time.Date(2026, 8, 22, 10, 0, 0, 0, loc) // 周六 10:00，落在高峰窗口
	band, fac := s.bandAt(sat)
	if band != PricingBandPeak || fac != 1.0 {
		t.Errorf("未开 peakWeekdaysOnly 的 schedule 周末应仍按窗口判峰；得到 (%q, %v)", band, fac)
	}
}
