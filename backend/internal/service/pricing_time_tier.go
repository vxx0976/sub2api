package service

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ---- 官方时段分档（peak / off-peak）----
//
// 部分上游按「时段」公布两套价（如 DeepSeek：高峰价 + 空闲时段半价）。本文件提供
// 一套与模型无关的时段判定：内置 ¥ 价表的每一行可以挂一个 *timeTierSchedule，
// 表里的数字恒为**基准价（最贵档 = 高峰价）**，空闲档由 offPeakFactor 折算。
//
// 「表价 = 最贵档」是全套机制的安全底座：所有未接线的路径、全部只读展示面、
// 渠道价自动填充都拿到基准价，因此**漏接线只会多收（可发现、可退款），
// 绝不会静默少收**。TestTimeTierFactorNeverExceedsOne 把这条不变量钉在测试里。
const (
	// PricingBandPeak 高峰档，即价目表原价。
	PricingBandPeak = "peak"
	// PricingBandOffPeak 空闲（波谷）档，按 offPeakFactor 折算。
	PricingBandOffPeak = "offpeak"
)

// timeTierWindow 是当日的一个左闭右开分钟区间 [startMin, endMin)，
// 与分组高峰倍率 Group.PeakMultiplierAt 的区间语义保持一致。
// 仅支持当日区间：空闲段（如 18:00→次日 09:00）通过「枚举高峰、其余即空闲」表达，
// 天然不需要跨天区间。
type timeTierWindow struct {
	startMin int
	endMin   int
}

func (w timeTierWindow) valid() bool {
	return w.startMin >= 0 && w.endMin <= 24*60 && w.startMin < w.endMin
}

func (w timeTierWindow) label() string {
	return fmt.Sprintf("%02d:%02d-%02d:%02d", w.startMin/60, w.startMin%60, w.endMin/60, w.endMin%60)
}

// timeTierSchedule 描述某个模型的官方时段分档规则。
//
// 时区**硬锚到上游厂商公布价格所用的时区**（如 DeepSeek 用北京时间），
// 刻意不跟随 timezone.Location()——后者是部署参数（TZ 环境变量可覆盖），
// 用它会让同一笔请求在不同机器上算出不同档位。同一手法见 geminiQuotaLocation()。
type timeTierSchedule struct {
	locFn         func() *time.Location // 惰性解析并缓存，避免在计费热路径重复 LoadLocation
	tzLabel       string                // 展示用，如 "UTC+08:00"
	peakWindows   []timeTierWindow
	offPeakFactor float64 // 空闲档相对表价的系数，必须落在 (0,1]
}

// bandAt 返回时刻 at 所处的档位与价格系数。纯函数，不读任何外部状态。
//
// 安全降级：receiver 为 nil / at 为零值 / 窗口非法 / 系数非法（含 NaN）/ 时区解析失败
// 一律返回 ("", 1.0)，即**基准价（最贵档）**且不标注档位。因此任何未接线或配置错误
// 的路径都落在「贵的一侧」，且 band 为空可直接作为漏接线的监控信号。
func (s *timeTierSchedule) bandAt(at time.Time) (string, float64) {
	if s == nil || at.IsZero() || len(s.peakWindows) == 0 || s.locFn == nil {
		return "", 1.0
	}
	// 注意 NaN：NaN 的任何比较都为 false，因此非法系数走这条降级分支。
	if !(s.offPeakFactor > 0 && s.offPeakFactor <= 1) {
		return "", 1.0
	}
	for _, w := range s.peakWindows {
		if !w.valid() {
			return "", 1.0
		}
	}
	loc := s.locFn()
	if loc == nil {
		return "", 1.0
	}

	t := at.In(loc)
	cur := t.Hour()*60 + t.Minute() // 秒被忽略：08:59:59 属空闲、09:00:00 起为高峰
	for _, w := range s.peakWindows {
		if cur >= w.startMin && cur < w.endMin {
			return PricingBandPeak, 1.0
		}
	}
	return PricingBandOffPeak, s.offPeakFactor
}

// windowLabels 返回高峰窗口的展示串（如 ["09:00-12:00", "14:00-18:00"]），供定价页展示。
func (s *timeTierSchedule) windowLabels() []string {
	if s == nil || len(s.peakWindows) == 0 {
		return nil
	}
	out := make([]string, 0, len(s.peakWindows))
	for _, w := range s.peakWindows {
		if w.valid() {
			out = append(out, w.label())
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// factor 返回空闲档系数；schedule 不可用时返回 0，调用方据此判断「无时段分档」。
func (s *timeTierSchedule) factor() float64 {
	if s == nil || !(s.offPeakFactor > 0 && s.offPeakFactor <= 1) {
		return 0
	}
	return s.offPeakFactor
}

// timezoneLabel 返回展示用时区标签。
func (s *timeTierSchedule) timezoneLabel() string {
	if s == nil {
		return ""
	}
	return s.tzLabel
}

// ModelTimeTierDTO 是展示层用的「官方时段分档」说明，供定价页/模型广场渲染两档价格。
type ModelTimeTierDTO struct {
	PeakWindows   []string `json:"peak_windows"`    // 高峰窗口，如 ["09:00-12:00","14:00-18:00"]
	Timezone      string   `json:"timezone"`        // 窗口所用时区标签，如 "UTC+08:00"
	OffPeakFactor float64  `json:"off_peak_factor"` // 空闲档系数（如 0.5）
	CurrentBand   string   `json:"current_band"`    // 当前所处档位：peak / offpeak
}

// ModelTimeTierInfo 返回某模型的官方时段分档说明；无分档时返回 nil。
//
// **必须由后端判定 CurrentBand**：档位锚定在上游厂商的时区上，前端按浏览器本地时区
// 算一定会错（同一时刻中美用户会看到不同档位）。
//
// 命中 admin 覆盖表时返回 nil —— 覆盖价是绝对价，官方峰谷分档对该模型已不再生效，
// 与计费口径（GetModelPricingAt 里覆盖表短路在最前面）保持一致。
func ModelTimeTierInfo(model string, at time.Time) *ModelTimeTierDTO {
	ml := strings.ToLower(strings.TrimSpace(model))
	if ml == "" {
		return nil
	}
	if ps := currentPricingService.Load(); ps != nil {
		if _, ok := ps.matchOverride(ml); ok {
			return nil
		}
	}

	var schedule *timeTierSchedule
	if cny, ok := matchKimiMoonshotCNY(ml); ok {
		schedule = cny.schedule
	} else if cny, ok := matchDeepSeekCNY(ml); ok {
		schedule = cny.schedule
	} else if cny, ok := matchQwenCNY(ml); ok {
		schedule = cny.schedule
	}
	if schedule == nil {
		return nil
	}
	windows := schedule.windowLabels()
	if len(windows) == 0 {
		return nil
	}
	band, _ := schedule.bandAt(at)
	return &ModelTimeTierDTO{
		PeakWindows:   windows,
		Timezone:      schedule.timezoneLabel(),
		OffPeakFactor: schedule.factor(),
		CurrentBand:   band,
	}
}

// deepSeekPricingLocation 返回 DeepSeek 官方计价时区（北京时间）。
// sync.OnceValue 保证 LoadLocation 每进程只做一次——Go 对非 UTC/Local 的时区不做缓存，
// 在每请求的计费热路径上重复解析会很贵（同一纪律见 group.go 手写 parseMinutes 的注释）。
// FixedZone 兜底：容器若未装 tzdata（镜像瘦身）也能给出正确的 UTC+8。
var deepSeekPricingLocation = sync.OnceValue(func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil && loc != nil {
		return loc
	}
	return time.FixedZone("CST", 8*60*60)
})

// deepSeekOfficialSchedule 是 DeepSeek 官方公布的分时段规则：
// 高峰时段为北京时间 09:00-12:00、14:00-18:00，其余为空闲时段，空闲价 = 高峰价的一半。
// 来源：https://api-docs.deepseek.com/zh-cn/quick_start/pricing/ （2026-08-17 核对）
var deepSeekOfficialSchedule = &timeTierSchedule{
	locFn:   deepSeekPricingLocation,
	tzLabel: "UTC+08:00",
	peakWindows: []timeTierWindow{
		{startMin: 9 * 60, endMin: 12 * 60},
		{startMin: 14 * 60, endMin: 18 * 60},
	},
	offPeakFactor: 0.5,
}
