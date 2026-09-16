//go:build unit

package service

import (
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// OpenCode Go 的模型目录里每一个 ID 都必须能解析出价格。
// 漏一个 = 该型号的请求零成本落账（GetModelPricingAt 报 ErrModelPricingUnavailable，
// 网关落「零成本 + 告警」分支），而日志里显示的正是这个型号名，对账时看不出异常。
// longcat / mimo / muse-spark / hy3|hy4 / omen 是 OpenCode 自有品牌，官方与 LiteLLM
// 都没有公开价，由 getFallbackPricing 末尾的 opencode-go-unpriced 条目按最贵档兜住。
func TestOpenCodeGoCatalogModelsAllPriced(t *testing.T) {
	svc := NewBillingService(&config.Config{}, NewPricingService(&config.Config{}, nil))
	for _, model := range DefaultOpenCodeGoModelIDs() {
		pricing, err := svc.GetModelPricingAt(strings.ToLower(model), time.Time{})
		if err != nil || pricing == nil {
			t.Errorf("OpenCode 目录型号 %q 解析不出价格（会零成本落账）: %v", model, err)
			continue
		}
		if pricing.InputPricePerToken <= 0 || pricing.OutputPricePerToken <= 0 {
			t.Errorf("OpenCode 目录型号 %q 价格为 0: in=%g out=%g",
				model, pricing.InputPricePerToken, pricing.OutputPricePerToken)
		}
	}
}
