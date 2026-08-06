package service

import (
	"context"
	"sort"
	"strings"
)

// 本文件是 fork 扩展，不在上游 channel_plaza.go 中。
//
// 上游模型广场的模型目录完全来自「渠道」：channels → channel_groups → channel_model_pricing。
// 本 fork 的部署从没用过那条链路（渠道表只用来做供应商余额监控），模型目录实际由
// **账号池** 承载：账号挂在分组下，账号的 model_mapping 声明它支持哪些模型。
// 于是上游广场在本 fork 上会把每个分组都算成 0 模型并整组丢弃，页面全空。
//
// 这里补一层回落：某分组从渠道侧一个模型都没拿到时，改用账号池推导模型名，
// 定价走 LiteLLM 全局表（与 fork 原公开定价页 /pricing/public/groups 同口径）。
// 只在「渠道侧为空」时触发，配了渠道的部署行为与上游完全一致。

// plazaAccountModelNames 推导某分组在账号侧支持的模型名。
//
// 口径与 fork 原公开定价页一致（并集，不是交集）：
//   - 只看该分组下平台匹配的活跃账号；
//   - 账号显式配了 model_mapping 时，取其 from 键（通配符 from 不算具体模型）；
//   - 一个显式声明的账号都没有（全是透传）→ 回落到该平台 LiteLLM 全表。
//
// 交集口径会在账号各自映射不相交时得出空集（历史上 GPT 分组显示 0 模型即此因），
// 故此处恒用并集。
func (s *ChannelService) plazaAccountModelNames(ctx context.Context, groupID int64, platform string) []string {
	var union map[string]struct{} // nil = 还没遇到任何「显式列模型」的账号

	if s.accountRepo != nil {
		accounts, err := s.accountRepo.ListByGroup(ctx, groupID)
		if err == nil {
			for i := range accounts {
				acc := &accounts[i]
				if acc.Status != StatusActive {
					continue
				}
				if acc.Platform != platform {
					// 防御性：account.platform 理论上应与 group.platform 一致。
					continue
				}
				mapping := acc.GetModelMapping()
				if len(mapping) == 0 {
					continue // 透传账号不参与并集
				}
				for from := range mapping {
					if strings.HasSuffix(from, "*") {
						continue
					}
					if union == nil {
						union = map[string]struct{}{}
					}
					union[from] = struct{}{}
				}
			}
		}
	}

	if union == nil {
		return s.plazaLiteLLMModelNames(platform)
	}
	out := make([]string, 0, len(union))
	for name := range union {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// plazaLiteLLMModelNames 返回某平台 LiteLLM 全表的模型名（已排序）。
func (s *ChannelService) plazaLiteLLMModelNames(platform string) []string {
	if s.pricingService == nil {
		return nil
	}
	entries := s.pricingService.ListAll(platform)
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Model == "" {
			continue
		}
		out = append(out, e.Model)
	}
	sort.Strings(out)
	return out
}

// plazaAccountFallbackModels 为「渠道侧零模型」的分组构造广场模型条目。
// 定价一律走 LiteLLM 全局表合成（账号侧没有渠道级单价），拿不到价的模型仍然展示，
// 前端显示 "-"：能看到分组支持哪些模型本身就是广场的主要价值。
func (s *ChannelService) plazaAccountFallbackModels(ctx context.Context, groupID int64, platform string) []PlazaModel {
	names := s.plazaAccountModelNames(ctx, groupID, platform)
	if len(names) == 0 {
		return nil
	}
	out := make([]PlazaModel, 0, len(names))
	for _, name := range names {
		var pricing *ChannelModelPricing
		if s.pricingService != nil {
			if lp := s.pricingService.GetModelPricing(name); lp != nil {
				pricing = synthesizePricingFromLiteLLM(lp, nil)
			}
		}
		out = append(out, PlazaModel{
			Name:     name,
			Platform: platform,
			Pricing:  pricing,
		})
	}
	return out
}
