//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"

	"github.com/stretchr/testify/require"
)

// 本文件专门守「请求冻结时刻 → 计费分档 → 档位落库」这条接线。
//
// 🔴 为什么必须在网关层测：BillingService/PricingService 层的用例只证明"给了时刻就会分档"，
// 完全证明不了"网关真的把时刻给进去了、并把结果写进了 usage_logs"。删掉
// gateway_usage_billing.go 的 `opts.PricingAt = pricingAt`、buildRecordUsageLog 里的
// PricingTimeBand/PricedAt、或 openai_gateway_service.go 的同名两处，
// 全量测试曾经**照样全绿**——DeepSeek 会全天按高峰价多收一倍而无人察觉。

// bandedBillingService 返回一个真正装配了官方 ¥ 表（含 DeepSeek 峰谷 schedule）的计费服务。
// 网关测试的默认 harness 用 NewBillingService(cfg, nil)，没有 PricingService 就永远走
// fallbackPrices（无分档），测不出档位。
func bandedBillingService() *BillingService {
	cfg := &config.Config{}
	cfg.Default.RateMultiplier = 1.0
	return NewBillingService(cfg, newCNYPricingService(1.0))
}

func TestGatewayRecordUsage_PersistsPricingTimeBandAndPricedAt(t *testing.T) {
	cases := []struct {
		name      string
		pricingAt time.Time
		wantBand  string
		// 1M input tokens 的期望成本（¥ 表 1:1 汇率）：高峰 3.0 / 空闲 1.5
		wantCost float64
	}{
		{"空闲时段发起 → 谷价 + offpeak", bjTestTime(20, 0, 0), PricingBandOffPeak, 1.5},
		{"高峰时段发起 → 表价 + peak", bjTestTime(10, 0, 0), PricingBandPeak, 3.0},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
			svc := newGatewayRecordUsageServiceForTest(usageRepo,
				&openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{})
			svc.billingService = bandedBillingService()

			err := svc.RecordUsage(context.Background(), &RecordUsageInput{
				Result: &ForwardResult{
					RequestID: "band_" + tt.wantBand,
					Model:     "deepseek-v4-flash",
					Usage:     ClaudeUsage{InputTokens: 1_000_000},
				},
				APIKey:    &APIKey{ID: 811},
				User:      &User{ID: 611},
				Account:   &Account{ID: 711, Platform: PlatformDeepSeek},
				PricingAt: tt.pricingAt,
			})
			require.NoError(t, err)
			require.NotNil(t, usageRepo.lastLog)

			require.NotNilf(t, usageRepo.lastLog.PricingTimeBand,
				"档位必须落库，否则「band 为空 = 漏接线」的监控信号会被淹没")
			require.Equal(t, tt.wantBand, *usageRepo.lastLog.PricingTimeBand)

			require.NotNil(t, usageRepo.lastLog.PricedAt, "定档时刻必须落库，供回放校验档位")
			require.True(t, usageRepo.lastLog.PricedAt.Equal(tt.pricingAt),
				"落库的 priced_at 必须正是请求开始冻结的那个时刻，got=%v want=%v",
				*usageRepo.lastLog.PricedAt, tt.pricingAt)

			require.InDelta(t, tt.wantCost, usageRepo.lastLog.TotalCost, 1e-9,
				"档位必须真的作用到金额上，而不只是打了个标签")
		})
	}
}

func TestOpenAIRecordUsage_PersistsPricingTimeBandAndPricedAt(t *testing.T) {
	cases := []struct {
		name      string
		pricingAt time.Time
		wantBand  string
		wantCost  float64
	}{
		{"空闲时段发起 → 谷价 + offpeak", bjTestTime(3, 0, 0), PricingBandOffPeak, 1.5},
		{"高峰时段发起 → 表价 + peak", bjTestTime(15, 0, 0), PricingBandPeak, 3.0},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
			svc := newOpenAIRecordUsageServiceForTest(usageRepo,
				&openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
			svc.billingService = bandedBillingService()
			svc.resolver = NewModelPricingResolver(nil, svc.billingService)

			err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
				Result: &OpenAIForwardResult{
					RequestID: "openai_band_" + tt.wantBand,
					Model:     "deepseek-v4-flash",
					Usage:     OpenAIUsage{InputTokens: 1_000_000},
				},
				APIKey:    &APIKey{ID: 812},
				User:      &User{ID: 612},
				Account:   &Account{ID: 712, Platform: PlatformDeepSeek},
				PricingAt: tt.pricingAt,
			})
			require.NoError(t, err)
			require.NotNil(t, usageRepo.lastLog)

			require.NotNilf(t, usageRepo.lastLog.PricingTimeBand, "OpenAI 链同样必须落档位")
			require.Equal(t, tt.wantBand, *usageRepo.lastLog.PricingTimeBand)
			require.NotNil(t, usageRepo.lastLog.PricedAt)
			require.True(t, usageRepo.lastLog.PricedAt.Equal(tt.pricingAt),
				"got=%v want=%v", *usageRepo.lastLog.PricedAt, tt.pricingAt)
			require.InDelta(t, tt.wantCost, usageRepo.lastLog.TotalCost, 1e-9)
		})
	}
}

// bjTestTime 按北京时间构造时刻（本文件里 *testing.T 不总在手边，故与 bj() 分开）。
func bjTestTime(hour, min, sec int) time.Time {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*60*60)
	}
	return time.Date(2026, 8, 17, hour, min, sec, 0, loc)
}
