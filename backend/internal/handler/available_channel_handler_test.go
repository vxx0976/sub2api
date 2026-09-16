//go:build unit

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUserAvailableChannel_Unauthenticated401(t *testing.T) {
	// 没有 AuthSubject 注入时，handler 应返回 401 且不触达 service 依赖。
	gin.SetMode(gin.TestMode)
	h := &AvailableChannelHandler{} // nil services — 401 路径不会调用它们
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/channels/available", nil)

	h.List(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFilterUserVisibleGroups_IntersectionOnly(t *testing.T) {
	// 渠道挂在 {g1, g2, g3}，用户只允许 {g1, g3} —— 响应必须仅含 g1/g3。
	groups := []service.AvailableGroupRef{
		{ID: 1, Name: "g1", Platform: "anthropic"},
		{ID: 2, Name: "g2", Platform: "anthropic"},
		{ID: 3, Name: "g3", Platform: "openai"},
	}
	allowed := map[int64]struct{}{1: {}, 3: {}}

	visible := filterUserVisibleGroups(groups, allowed)
	require.Len(t, visible, 2)
	ids := []int64{visible[0].ID, visible[1].ID}
	require.ElementsMatch(t, []int64{1, 3}, ids)
}

func TestToUserSupportedModels_FiltersByAllowedPlatforms(t *testing.T) {
	// 用户可访问分组只覆盖 anthropic；anthropic 平台的模型保留，openai 模型被剔除。
	src := []service.SupportedModel{
		{Name: "claude-sonnet-4-6", Platform: "anthropic", Pricing: nil},
		{Name: "gpt-4o", Platform: "openai", Pricing: nil},
	}
	allowed := map[string]struct{}{"anthropic": {}}
	out := toUserSupportedModels(src, allowed)
	require.Len(t, out, 1)
	require.Equal(t, "claude-sonnet-4-6", out[0].Name)
}

func TestToUserSupportedModels_NilAllowedPlatformsKeepsAll(t *testing.T) {
	// 显式传 nil allowedPlatforms 表示不做过滤。
	src := []service.SupportedModel{
		{Name: "a", Platform: "anthropic"},
		{Name: "b", Platform: "openai"},
	}
	require.Len(t, toUserSupportedModels(src, nil), 2)
}

func TestUserAvailableChannel_FieldWhitelist(t *testing.T) {
	// 通过序列化 userAvailableChannel 结构体验证响应形状：
	// 只有 name / description / platforms；不含管理端字段。
	row := userAvailableChannel{
		Name:        "ch",
		Description: "d",
		Platforms: []userChannelPlatformSection{
			{
				Platform:        "anthropic",
				Groups:          []userAvailableGroup{{ID: 1, Name: "g1", Platform: "anthropic"}},
				SupportedModels: []userSupportedModel{},
			},
		},
	}
	raw, err := json.Marshal(row)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	for _, key := range []string{"id", "status", "billing_model_source", "restrict_models"} {
		_, exists := decoded[key]
		require.Falsef(t, exists, "user DTO must not expose %q", key)
	}
	for _, key := range []string{"name", "description", "platforms"} {
		_, exists := decoded[key]
		require.Truef(t, exists, "user DTO must expose %q", key)
	}

	// 验证 section 的字段（platform / groups / supported_models）。
	rawSection, err := json.Marshal(row.Platforms[0])
	require.NoError(t, err)
	var sectionDecoded map[string]any
	require.NoError(t, json.Unmarshal(rawSection, &sectionDecoded))
	for _, key := range []string{"platform", "groups", "supported_models"} {
		_, exists := sectionDecoded[key]
		require.Truef(t, exists, "platform section must expose %q", key)
	}

	// Group DTO 暴露区分专属/公开、订阅类型、默认倍率和高峰倍率规则所需的字段，
	// 前端据此渲染 GroupBadge 并与 API 密钥页保持一致的视觉。
	rawGroup, err := json.Marshal(row.Platforms[0].Groups[0])
	require.NoError(t, err)
	var groupDecoded map[string]any
	require.NoError(t, json.Unmarshal(rawGroup, &groupDecoded))
	for _, key := range []string{"id", "name", "platform", "subscription_type", "rate_multiplier", "peak_rate_enabled", "peak_start", "peak_end", "peak_rate_multiplier", "is_exclusive"} {
		_, exists := groupDecoded[key]
		require.Truef(t, exists, "group DTO must expose %q", key)
	}

	// pricing interval 白名单：不应暴露 id / sort_order。
	inputMultiplier := 2.0
	outputMultiplier := 1.5
	cacheWriteMultiplier := 2.0
	cacheReadMultiplier := 2.0
	pricing := toUserPricing(&service.ChannelModelPricing{
		BillingMode: service.BillingModeToken,
		Intervals: []service.PricingInterval{
			{
				ID: 7, MinTokens: 0, MaxTokens: nil, SortOrder: 3,
				InputMultiplier: &inputMultiplier, OutputMultiplier: &outputMultiplier,
				CacheWriteMultiplier: &cacheWriteMultiplier, CacheReadMultiplier: &cacheReadMultiplier,
			},
		},
	})
	require.NotNil(t, pricing)
	require.Len(t, pricing.Intervals, 1)
	rawIv, err := json.Marshal(pricing.Intervals[0])
	require.NoError(t, err)
	var ivDecoded map[string]any
	require.NoError(t, json.Unmarshal(rawIv, &ivDecoded))
	for _, key := range []string{"id", "pricing_id", "sort_order"} {
		_, exists := ivDecoded[key]
		require.Falsef(t, exists, "user pricing interval must not expose %q", key)
	}
	for key, want := range map[string]float64{
		"input_multiplier": inputMultiplier, "output_multiplier": outputMultiplier,
		"cache_write_multiplier": cacheWriteMultiplier, "cache_read_multiplier": cacheReadMultiplier,
	} {
		got, exists := ivDecoded[key]
		require.Truef(t, exists, "user pricing interval must expose %q", key)
		require.InDelta(t, want, got.(float64), 1e-12)
	}
}

func TestBuildPlatformSections_GroupsByPlatform(t *testing.T) {
	// 一个渠道横跨 anthropic / openai / 空平台：应该生成 2 个 section，
	// 按 platform 字母序排序，各自 groups 和 supported_models 只含同平台条目。
	ch := service.AvailableChannel{
		Name: "ch",
		SupportedModels: []service.SupportedModel{
			{Name: "claude-sonnet-4-6", Platform: "anthropic"},
			{Name: "gpt-4o", Platform: "openai"},
		},
	}
	visible := []userAvailableGroup{
		{ID: 1, Name: "g-openai", Platform: "openai"},
		{ID: 2, Name: "g-ant", Platform: "anthropic"},
		{ID: 3, Name: "g-empty", Platform: ""},
	}
	sections := buildPlatformSections(ch, visible)
	require.Len(t, sections, 2)
	require.Equal(t, "anthropic", sections[0].Platform)
	require.Equal(t, "openai", sections[1].Platform)
	require.Len(t, sections[0].Groups, 1)
	require.Equal(t, int64(2), sections[0].Groups[0].ID)
	require.Len(t, sections[0].SupportedModels, 1)
	require.Equal(t, "claude-sonnet-4-6", sections[0].SupportedModels[0].Name)
}

func TestBuildPlatformSections_CompositeGroupExpandsAcrossConfiguredModelPlatforms(t *testing.T) {
	anthropicPrice := 3e-6
	openAIPrice := 2.5e-6
	ch := service.AvailableChannel{
		Name: "composite-channel",
		SupportedModels: []service.SupportedModel{
			{
				Name:     "claude-sonnet-4-6",
				Platform: service.PlatformAnthropic,
				Pricing:  &service.ChannelModelPricing{InputPrice: &anthropicPrice},
			},
			{
				Name:     "gpt-5",
				Platform: service.PlatformOpenAI,
				Pricing:  &service.ChannelModelPricing{InputPrice: &openAIPrice},
			},
		},
	}
	visible := []userAvailableGroup{
		{ID: 9, Name: "composite", Platform: service.PlatformComposite},
	}

	sections := buildPlatformSections(ch, visible)

	require.Len(t, sections, 2)
	require.Equal(t, service.PlatformAnthropic, sections[0].Platform)
	require.Equal(t, service.PlatformOpenAI, sections[1].Platform)
	for _, section := range sections {
		require.Len(t, section.Groups, 1)
		require.Equal(t, int64(9), section.Groups[0].ID)
		require.Equal(t, service.PlatformComposite, section.Groups[0].Platform)
		require.Len(t, section.SupportedModels, 1)
		require.Equal(t, section.Platform, section.SupportedModels[0].Platform)
		require.NotNil(t, section.SupportedModels[0].Pricing)
	}
	require.Equal(t, "claude-sonnet-4-6", sections[0].SupportedModels[0].Name)
	require.Equal(t, "gpt-5", sections[1].SupportedModels[0].Name)
}

func TestBuildPlatformSections_OrdinaryGroupRemainsPlatformIsolated(t *testing.T) {
	ch := service.AvailableChannel{
		SupportedModels: []service.SupportedModel{
			{Name: "claude-sonnet-4-6", Platform: service.PlatformAnthropic},
			{Name: "gpt-5", Platform: service.PlatformOpenAI},
		},
	}
	visible := []userAvailableGroup{
		{ID: 1, Name: "anthropic-only", Platform: service.PlatformAnthropic},
	}

	sections := buildPlatformSections(ch, visible)

	require.Len(t, sections, 1)
	require.Equal(t, service.PlatformAnthropic, sections[0].Platform)
	require.Len(t, sections[0].SupportedModels, 1)
	require.Equal(t, "claude-sonnet-4-6", sections[0].SupportedModels[0].Name)
}

func TestBuildPlatformSections_CompositeAndOrdinaryGroupsShareConcreteSection(t *testing.T) {
	ch := service.AvailableChannel{
		SupportedModels: []service.SupportedModel{
			{Name: "claude-sonnet-4-6", Platform: service.PlatformAnthropic},
			{Name: "gpt-5", Platform: service.PlatformOpenAI},
		},
	}
	visible := []userAvailableGroup{
		{ID: 1, Name: "anthropic-only", Platform: service.PlatformAnthropic},
		{ID: 9, Name: "composite", Platform: service.PlatformComposite},
	}

	sections := buildPlatformSections(ch, visible)

	require.Len(t, sections, 2)
	require.Equal(t, service.PlatformAnthropic, sections[0].Platform)
	require.Equal(t, []int64{1, 9}, []int64{
		sections[0].Groups[0].ID,
		sections[0].Groups[1].ID,
	})
	require.Equal(t, service.PlatformOpenAI, sections[1].Platform)
	require.Len(t, sections[1].Groups, 1)
	require.Equal(t, int64(9), sections[1].Groups[0].ID)
}

func TestBuildPlatformSections_CompositeWithoutModelsKeepsEmptyCompositeSection(t *testing.T) {
	visible := []userAvailableGroup{
		{ID: 9, Name: "composite", Platform: service.PlatformComposite},
	}

	sections := buildPlatformSections(service.AvailableChannel{
		SupportedModels: []service.SupportedModel{{Name: "invalid-without-platform"}},
	}, visible)

	require.Len(t, sections, 1)
	require.Equal(t, service.PlatformComposite, sections[0].Platform)
	require.Len(t, sections[0].Groups, 1)
	require.Empty(t, sections[0].SupportedModels)
}

// 分组价卡是计费链路的最高优先级来源（service.ModelPricingResolver.Resolve：
// Group → Channel → LiteLLM → Fallback）。定价展示端必须同源，否则被分组价卡改过价的
// 模型会按官方价展示，与实际扣费不符（线上真实案例：分组 29 把 gpt-5.6-luna 调到
// terra 价后，/pricing/groups 仍返回官方 $0.2/MTok）。
func TestBuildPricingModel_GroupCardOverridesOfficialPrice(t *testing.T) {
	inPrice, outPrice := 2e-06, 1.2e-05
	group := &service.Group{
		ID:       29,
		Platform: service.PlatformOpenAI,
		ModelPricing: []service.ChannelModelPricing{{
			Platform:    service.PlatformOpenAI,
			Models:      []string{"gpt-5.6-luna"},
			BillingMode: service.BillingModeToken,
			InputPrice:  &inPrice,
			OutputPrice: &outPrice,
		}},
	}
	officialIn, officialOut := 2e-07, 1.2e-06
	lookupOfficial := func(string) *service.ModelPricing {
		return &service.ModelPricing{
			InputPricePerToken:  officialIn,
			OutputPricePerToken: officialOut,
		}
	}

	// channelService 为 nil：分组价卡必须独立生效，不依赖分组是否绑定了渠道。
	h := &AvailableChannelHandler{}
	m := h.buildPricingModel(t.Context(), group, "gpt-5.6-luna", lookupOfficial)

	require.NotNil(t, m.InputPrice)
	require.InDelta(t, inPrice, *m.InputPrice, 1e-12)
	require.NotNil(t, m.OutputPrice)
	require.InDelta(t, outPrice, *m.OutputPrice, 1e-12)
	// official_* 保持真实官方价：前端的「本站价 vs 官方价」对比依赖它。
	require.NotNil(t, m.OfficialInputPrice)
	require.InDelta(t, officialIn, *m.OfficialInputPrice, 1e-12)
}

// 同一分组里没有价卡的模型不受影响：显式价留空，前端回退 official_*。
func TestBuildPricingModel_UnmatchedModelKeepsOfficialOnly(t *testing.T) {
	inPrice := 2e-06
	group := &service.Group{
		ID:       29,
		Platform: service.PlatformOpenAI,
		ModelPricing: []service.ChannelModelPricing{{
			Platform:   service.PlatformOpenAI,
			Models:     []string{"gpt-5.6-luna"},
			InputPrice: &inPrice,
		}},
	}
	lookupOfficial := func(string) *service.ModelPricing {
		return &service.ModelPricing{InputPricePerToken: 2e-06}
	}

	h := &AvailableChannelHandler{}
	m := h.buildPricingModel(t.Context(), group, "gpt-5.6-terra", lookupOfficial)

	require.Nil(t, m.InputPrice)
	require.NotNil(t, m.OfficialInputPrice)
}

// token 模式的分组价卡只覆盖首档/平价，长上下文阶梯走官方预设（Resolve 里
// stripped.Intervals = nil）。展示端保留区间会显示一套永不成交的阶梯。
func TestBuildPricingModel_GroupTokenCardDropsIntervals(t *testing.T) {
	inPrice, tierPrice := 2e-06, 9e-06
	group := &service.Group{
		ID:       29,
		Platform: service.PlatformOpenAI,
		ModelPricing: []service.ChannelModelPricing{{
			Platform:    service.PlatformOpenAI,
			Models:      []string{"gpt-5.6-luna"},
			BillingMode: service.BillingModeToken,
			InputPrice:  &inPrice,
			Intervals: []service.PricingInterval{{
				MinTokens:  272000,
				TierLabel:  ">272K",
				InputPrice: &tierPrice,
			}},
		}},
	}

	h := &AvailableChannelHandler{}
	m := h.buildPricingModel(t.Context(), group, "gpt-5.6-luna", func(string) *service.ModelPricing { return nil })

	require.NotNil(t, m.InputPrice)
	require.Empty(t, m.Intervals)
}

// 按次/图片模式的分组价卡区间确实参与计费（Resolve 只对 token 模式剥离），必须展示。
func TestBuildPricingModel_GroupPerRequestCardKeepsIntervals(t *testing.T) {
	tierPrice := 0.02
	group := &service.Group{
		ID:       29,
		Platform: service.PlatformOpenAI,
		ModelPricing: []service.ChannelModelPricing{{
			Platform:    service.PlatformOpenAI,
			Models:      []string{"gpt-image-2"},
			BillingMode: service.BillingModeImage,
			Intervals: []service.PricingInterval{{
				TierLabel:       "2K",
				PerRequestPrice: &tierPrice,
			}},
		}},
	}

	h := &AvailableChannelHandler{}
	m := h.buildPricingModel(t.Context(), group, "gpt-image-2", func(string) *service.ModelPricing { return nil })

	require.Len(t, m.Intervals, 1)
	require.Equal(t, "2K", m.Intervals[0].TierLabel)
}
