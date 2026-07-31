package handler

import (
	"context"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AvailableChannelHandler 处理用户侧「可用渠道」查询。
//
// 用户侧接口委托 ChannelService.ListAvailable，并在返回前做四层过滤：
//  1. 行过滤：只保留状态为 Active 且与当前用户可访问分组有交集的渠道；
//  2. 分组过滤：渠道的 Groups 只保留用户可访问的那些；
//  3. 平台过滤：普通分组只保留自身平台模型；Composite 分组按渠道已配置的具体模型平台
//     展开。这样既防止普通分组跨平台泄漏，也让 Composite 正确展示其多平台能力；
//  4. 字段白名单：仅返回用户需要的字段（省略 BillingModelSource / RestrictModels
//     / 内部 ID / Status 等管理字段）。
type AvailableChannelHandler struct {
	channelService *service.ChannelService
	apiKeyService  *service.APIKeyService
	settingService *service.SettingService
	billingService *service.BillingService
	accountService *service.AccountService
}

// NewAvailableChannelHandler 创建用户侧可用渠道 handler。
// billingService / accountService 供公开「模型广场」端点解析模型集合与官方价；
// 可为 nil（此时官方价与账号交集兜底部分留空）。
func NewAvailableChannelHandler(
	channelService *service.ChannelService,
	apiKeyService *service.APIKeyService,
	settingService *service.SettingService,
	billingService *service.BillingService,
	accountService *service.AccountService,
) *AvailableChannelHandler {
	return &AvailableChannelHandler{
		channelService: channelService,
		apiKeyService:  apiKeyService,
		settingService: settingService,
		billingService: billingService,
		accountService: accountService,
	}
}

// featureEnabled 返回 available-channels 开关是否启用。默认关闭（opt-in）。
func (h *AvailableChannelHandler) featureEnabled(c *gin.Context) bool {
	if h.settingService == nil {
		return false
	}
	return h.settingService.GetAvailableChannelsRuntime(c.Request.Context()).Enabled
}

// userAvailableGroup 用户可见的分组概要（白名单字段）。
//
// 前端据此区分专属 vs 公开分组（IsExclusive）、订阅 vs 标准分组（SubscriptionType，
// 订阅视觉加深），并展示默认倍率与高峰倍率规则；用户专属倍率前端走
// /groups/rates，和 API 密钥页面保持一致。
type userAvailableGroup struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	Platform           string  `json:"platform"`
	SubscriptionType   string  `json:"subscription_type"`
	RateMultiplier     float64 `json:"rate_multiplier"`
	PeakRateEnabled    bool    `json:"peak_rate_enabled"`
	PeakStart          string  `json:"peak_start"`
	PeakEnd            string  `json:"peak_end"`
	PeakRateMultiplier float64 `json:"peak_rate_multiplier"`
	IsExclusive        bool    `json:"is_exclusive"`
}

// userSupportedModelPricing 用户可见的定价字段白名单。
type userSupportedModelPricing struct {
	BillingMode      string                   `json:"billing_mode"`
	InputPrice       *float64                 `json:"input_price"`
	OutputPrice      *float64                 `json:"output_price"`
	CacheWritePrice  *float64                 `json:"cache_write_price"`
	CacheReadPrice   *float64                 `json:"cache_read_price"`
	ImageInputPrice  *float64                 `json:"image_input_price"`
	ImageOutputPrice *float64                 `json:"image_output_price"`
	PerRequestPrice  *float64                 `json:"per_request_price"`
	Intervals        []userPricingIntervalDTO `json:"intervals"`
}

// userPricingIntervalDTO 定价区间白名单（去掉内部 ID、SortOrder 等前端不渲染的字段）。
type userPricingIntervalDTO struct {
	MinTokens       int      `json:"min_tokens"`
	MaxTokens       *int     `json:"max_tokens"`
	TierLabel       string   `json:"tier_label,omitempty"`
	InputPrice      *float64 `json:"input_price"`
	OutputPrice     *float64 `json:"output_price"`
	CacheWritePrice *float64 `json:"cache_write_price"`
	CacheReadPrice  *float64 `json:"cache_read_price"`
	PerRequestPrice *float64 `json:"per_request_price"`
}

// userSupportedModel 用户可见的支持模型条目。
type userSupportedModel struct {
	Name     string                     `json:"name"`
	Platform string                     `json:"platform"`
	Pricing  *userSupportedModelPricing `json:"pricing"`
}

// userChannelPlatformSection 单渠道内某个平台的子视图：用户可见的分组 + 该平台
// 支持的模型。按 platform 聚合后让前端可以把渠道名作为 row-group 一次渲染，
// 后面的平台行按 sections 顺序铺开。
type userChannelPlatformSection struct {
	Platform        string               `json:"platform"`
	Groups          []userAvailableGroup `json:"groups"`
	SupportedModels []userSupportedModel `json:"supported_models"`
}

// userAvailableChannel 用户可见的渠道条目（白名单字段）。
//
// 每个渠道聚合为一条记录，内嵌 platforms 子数组：每个 section 对应一个平台，
// 包含该平台的 groups 和 supported_models。
type userAvailableChannel struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Platforms   []userChannelPlatformSection `json:"platforms"`
}

// List 列出当前用户可见的「可用渠道」。
// GET /api/v1/channels/available
func (h *AvailableChannelHandler) List(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Feature 未启用时返回空数组（不暴露渠道信息）。检查放在认证之后，
	// 保持与未开关前的 401 行为一致：未登录先 401，登录后再按开关决定。
	if !h.featureEnabled(c) {
		response.Success(c, []userAvailableChannel{})
		return
	}

	userGroups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	allowedGroupIDs := make(map[int64]struct{}, len(userGroups))
	for i := range userGroups {
		allowedGroupIDs[userGroups[i].ID] = struct{}{}
	}

	channels, err := h.channelService.ListAvailable(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]userAvailableChannel, 0, len(channels))
	for _, ch := range channels {
		if ch.Status != service.StatusActive {
			continue
		}
		visibleGroups := filterUserVisibleGroups(ch.Groups, allowedGroupIDs)
		if len(visibleGroups) == 0 {
			continue
		}
		sections := buildPlatformSections(ch, visibleGroups)
		if len(sections) == 0 {
			continue
		}
		out = append(out, userAvailableChannel{
			Name:        ch.Name,
			Description: ch.Description,
			Platforms:   sections,
		})
	}

	response.Success(c, out)
}

// buildPlatformSections 把一个渠道按 visibleGroups 的平台集合拆成有序的 section 列表：
// 每个 section 对应一个具体平台，只包含该平台的 groups 和 supported_models。
//
// Composite 分组可访问渠道中所有已配置的具体平台，因此会被展开到每个有支持模型的
// 平台 section。普通分组仍严格留在自身平台，避免跨平台模型信息泄漏。Composite 渠道
// 尚未配置任何模型时保留 composite section，以便前端继续展示该分组和“未配置模型”状态。
// 输出按 platform 字母序稳定排序，便于前端等效比较与回归测试。
func buildPlatformSections(
	ch service.AvailableChannel,
	visibleGroups []userAvailableGroup,
) []userChannelPlatformSection {
	groupsByPlatform := make(map[string][]userAvailableGroup, 4)
	compositeGroups := make([]userAvailableGroup, 0, 1)
	for _, g := range visibleGroups {
		if g.Platform == "" {
			continue
		}
		if g.Platform == service.PlatformComposite {
			compositeGroups = append(compositeGroups, g)
			continue
		}
		groupsByPlatform[g.Platform] = append(groupsByPlatform[g.Platform], g)
	}

	if len(compositeGroups) > 0 {
		modelPlatforms := make(map[string]struct{}, len(ch.SupportedModels))
		for i := range ch.SupportedModels {
			if platform := ch.SupportedModels[i].Platform; platform != "" {
				modelPlatforms[platform] = struct{}{}
			}
		}
		if len(modelPlatforms) == 0 {
			groupsByPlatform[service.PlatformComposite] = append(
				groupsByPlatform[service.PlatformComposite],
				compositeGroups...,
			)
		} else {
			for platform := range modelPlatforms {
				groupsByPlatform[platform] = append(groupsByPlatform[platform], compositeGroups...)
			}
		}
	}
	if len(groupsByPlatform) == 0 {
		return nil
	}

	platforms := make([]string, 0, len(groupsByPlatform))
	for p := range groupsByPlatform {
		platforms = append(platforms, p)
	}
	sort.Strings(platforms)

	sections := make([]userChannelPlatformSection, 0, len(platforms))
	for _, platform := range platforms {
		platformSet := map[string]struct{}{platform: {}}
		sections = append(sections, userChannelPlatformSection{
			Platform:        platform,
			Groups:          groupsByPlatform[platform],
			SupportedModels: toUserSupportedModels(ch.SupportedModels, platformSet),
		})
	}
	return sections
}

// filterUserVisibleGroups 仅保留用户可访问的分组。
func filterUserVisibleGroups(
	groups []service.AvailableGroupRef,
	allowed map[int64]struct{},
) []userAvailableGroup {
	visible := make([]userAvailableGroup, 0, len(groups))
	for _, g := range groups {
		if _, ok := allowed[g.ID]; !ok {
			continue
		}
		visible = append(visible, userAvailableGroup{
			ID:                 g.ID,
			Name:               g.Name,
			Platform:           g.Platform,
			SubscriptionType:   g.SubscriptionType,
			RateMultiplier:     g.RateMultiplier,
			PeakRateEnabled:    g.PeakRateEnabled,
			PeakStart:          g.PeakStart,
			PeakEnd:            g.PeakEnd,
			PeakRateMultiplier: g.PeakRateMultiplier,
			IsExclusive:        g.IsExclusive,
		})
	}
	return visible
}

// toUserSupportedModels 将 service 层支持模型转换为用户 DTO（字段白名单）。
// 仅保留平台在 allowedPlatforms 中的条目，防止跨平台模型信息泄漏。
// allowedPlatforms 为 nil 时不做平台过滤（保留全部，供测试或明确无过滤场景使用）。
func toUserSupportedModels(
	src []service.SupportedModel,
	allowedPlatforms map[string]struct{},
) []userSupportedModel {
	out := make([]userSupportedModel, 0, len(src))
	for i := range src {
		m := src[i]
		if allowedPlatforms != nil {
			if _, ok := allowedPlatforms[m.Platform]; !ok {
				continue
			}
		}
		out = append(out, userSupportedModel{
			Name:     m.Name,
			Platform: m.Platform,
			Pricing:  toUserPricing(m.Pricing),
		})
	}
	return out
}

// toUserPricing 将 service 层定价转换为用户 DTO；入参为 nil 时返回 nil。
func toUserPricing(p *service.ChannelModelPricing) *userSupportedModelPricing {
	if p == nil {
		return nil
	}
	intervals := make([]userPricingIntervalDTO, 0, len(p.Intervals))
	for _, iv := range p.Intervals {
		intervals = append(intervals, userPricingIntervalDTO{
			MinTokens:       iv.MinTokens,
			MaxTokens:       iv.MaxTokens,
			TierLabel:       iv.TierLabel,
			InputPrice:      iv.InputPrice,
			OutputPrice:     iv.OutputPrice,
			CacheWritePrice: iv.CacheWritePrice,
			CacheReadPrice:  iv.CacheReadPrice,
			PerRequestPrice: iv.PerRequestPrice,
		})
	}
	billingMode := string(p.BillingMode)
	if billingMode == "" {
		billingMode = string(service.BillingModeToken)
	}
	return &userSupportedModelPricing{
		BillingMode:      billingMode,
		InputPrice:       p.InputPrice,
		OutputPrice:      p.OutputPrice,
		CacheWritePrice:  p.CacheWritePrice,
		CacheReadPrice:   p.CacheReadPrice,
		ImageInputPrice:  p.ImageInputPrice,
		ImageOutputPrice: p.ImageOutputPrice,
		PerRequestPrice:  p.PerRequestPrice,
		Intervals:        intervals,
	}
}

// ============================ 模型广场（公开定价端点）============================

// userPricingModel 模型广场端点（group）下的单条模型定价。
//
// 价格字段两套，均为"基础单价"（USD / per token）语义：
//   - input_price / output_price / cache_write_price / cache_read_price：渠道
//     管理员显式配置的基础单价；nil 表示该字段未配置，前端回退到 official_*。
//   - official_*：LiteLLM 官方价。
//
// 前端在 site 模式按 (group.rate_multiplier / cny_per_usd) 乘出本站价，与计费链路
// actualCost = totalCost × RateMultiplier 一致。billing_mode / intervals 用于
// per_request / image 等非 token 模式。
type userPricingModel struct {
	Name                    string                   `json:"name"`
	BillingMode             string                   `json:"billing_mode,omitempty"`
	InputPrice              *float64                 `json:"input_price,omitempty"`
	OutputPrice             *float64                 `json:"output_price,omitempty"`
	CacheWritePrice         *float64                 `json:"cache_write_price,omitempty"`
	CacheReadPrice          *float64                 `json:"cache_read_price,omitempty"`
	PerRequestPrice         *float64                 `json:"per_request_price,omitempty"`
	Intervals               []userPricingIntervalDTO `json:"intervals,omitempty"`
	OfficialInputPrice      *float64                 `json:"official_input_price,omitempty"`
	OfficialOutputPrice     *float64                 `json:"official_output_price,omitempty"`
	OfficialCacheWritePrice *float64                 `json:"official_cache_write_price,omitempty"`
	OfficialCacheReadPrice  *float64                 `json:"official_cache_read_price,omitempty"`
	// PriceCurrency 价格币种口径："CNY" 表示国产官方人民币计价模型（前端显示 ¥），
	// 其余为 "USD"。与用量页 price_currency 同源（service.ModelPriceCurrency）。
	PriceCurrency string `json:"price_currency,omitempty"`
}

// userPricingGroup 模型广场展示页的端点 = 一个 group。折扣由 rate_multiplier 决定。
type userPricingGroup struct {
	ID             int64              `json:"id"`
	Name           string             `json:"name"`
	Platform       string             `json:"platform"`
	RateMultiplier float64            `json:"rate_multiplier"`
	IsExclusive    bool               `json:"is_exclusive"`
	Models         []userPricingModel `json:"models"`
}

// buildPricingGroups 把一批分组渲染成模型广场/定价 DTO：按账号交集(+LiteLLM 兜底)
// 解析每组模型集合，并填充渠道显式价与 LiteLLM 官方价。公开端点与认证端点共用。
func (h *AvailableChannelHandler) buildPricingGroups(ctx context.Context, groups []service.Group) []userPricingGroup {
	// 官方价 lookup 缓存
	priceCache := make(map[string]*service.ModelPricing, 32)
	lookupOfficial := func(model string) *service.ModelPricing {
		if p, ok := priceCache[model]; ok {
			return p
		}
		if h.billingService == nil {
			priceCache[model] = nil
			return nil
		}
		p, err := h.billingService.GetModelPricing(model)
		if err != nil {
			priceCache[model] = nil
			return nil
		}
		priceCache[model] = p
		return p
	}

	// LiteLLM 兜底：每个 platform 一次拉全表缓存。
	litellmByPlatform := map[string][]string{}
	getLiteLLMModels := func(platform string) []string {
		if v, ok := litellmByPlatform[platform]; ok {
			return v
		}
		var out []string
		if h.billingService != nil {
			for _, e := range h.billingService.ListAllModelPricings(platform) {
				out = append(out, e.Model)
			}
		}
		litellmByPlatform[platform] = out
		return out
	}

	out := make([]userPricingGroup, 0, len(groups))
	for i := range groups {
		g := groups[i]
		modelNames := h.resolveGroupModelsByAccount(ctx, g.ID, g.Platform, getLiteLLMModels)
		item := userPricingGroup{
			ID:             g.ID,
			Name:           g.Name,
			Platform:       g.Platform,
			RateMultiplier: g.RateMultiplier,
			IsExclusive:    g.IsExclusive,
			Models:         []userPricingModel{},
		}
		for _, name := range modelNames {
			item.Models = append(item.Models, h.buildPricingModel(ctx, g.ID, name, lookupOfficial))
		}
		out = append(out, item)
	}
	return out
}

// PricingGroupListPublic 返回所有公开（非专属、非订阅）的活跃分组，用于未登录访客
// 的"模型广场"展示。不应用任何用户上下文，不受 available_channels_enabled 开关限制。
//
// GET /api/v1/pricing/public/groups
func (h *AvailableChannelHandler) PricingGroupListPublic(c *gin.Context) {
	ctx := c.Request.Context()
	groups, err := h.apiKeyService.ListPublicGroups(ctx)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if len(groups) == 0 {
		response.Success(c, []userPricingGroup{})
		return
	}
	response.Success(c, h.buildPricingGroups(ctx, groups))
}

// PricingGroupList 返回当前登录用户可访问的分组及其模型定价（含专属/订阅组），用于
// 登录后的「模型定价」页。与公开端点共享模型解析与官方价逻辑，按 group.rate_multiplier
// 展示倍率（本 fork 无商户 sell_rate 概念）。
//
// GET /api/v1/pricing/groups
func (h *AvailableChannelHandler) PricingGroupList(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	ctx := c.Request.Context()
	groups, err := h.apiKeyService.GetAvailableGroups(ctx, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if len(groups) == 0 {
		response.Success(c, []userPricingGroup{})
		return
	}
	response.Success(c, h.buildPricingGroups(ctx, groups))
}

// GetFXRate 返回展示用汇率 cny_per_usd，供模型广场「本站价 vs 官方价」换算。
// 值即 pricing.cny_to_usd_rate（语义 = 1 USD ≈ 几 CNY，计费侧 CNY价/rate=USD）；
// 1¥=1$ 余额模型下默认为 1.0。公开无需认证。
//
// GET /api/v1/pricing/public/fx-rate
func (h *AvailableChannelHandler) GetFXRate(c *gin.Context) {
	rate := 1.0
	if h.billingService != nil {
		if v := h.billingService.CNYPerUSD(); v > 0 {
			rate = v
		}
	}
	response.Success(c, gin.H{"cny_per_usd": rate, "last_updated": nil})
}

// buildPricingModel 拼装单条模型的定价 DTO：渠道显式配置的基础单价覆盖到独立字段，
// LiteLLM 官方价填到 official_*。两套字段都保持"基础单价"语义。
// channelService 未注入或 group 未绑定 channel 时，channel 价部分留空。
func (h *AvailableChannelHandler) buildPricingModel(
	ctx context.Context,
	groupID int64,
	name string,
	lookupOfficial func(string) *service.ModelPricing,
) userPricingModel {
	m := userPricingModel{Name: name, PriceCurrency: service.ModelPriceCurrency(name)}
	if p := lookupOfficial(name); p != nil {
		m.OfficialInputPrice = positiveFloatPtr(p.InputPricePerToken)
		m.OfficialOutputPrice = positiveFloatPtr(p.OutputPricePerToken)
		m.OfficialCacheWritePrice = positiveFloatPtr(p.CacheCreationPricePerToken)
		m.OfficialCacheReadPrice = positiveFloatPtr(p.CacheReadPricePerToken)
	}
	if h.channelService != nil {
		if cp := h.channelService.GetChannelModelPricing(ctx, groupID, name); cp != nil {
			m.BillingMode = string(cp.BillingMode)
			m.InputPrice = cp.InputPrice
			m.OutputPrice = cp.OutputPrice
			m.CacheWritePrice = cp.CacheWritePrice
			m.CacheReadPrice = cp.CacheReadPrice
			m.PerRequestPrice = cp.PerRequestPrice
			if len(cp.Intervals) > 0 {
				m.Intervals = make([]userPricingIntervalDTO, 0, len(cp.Intervals))
				for _, iv := range cp.Intervals {
					m.Intervals = append(m.Intervals, userPricingIntervalDTO{
						MinTokens:       iv.MinTokens,
						MaxTokens:       iv.MaxTokens,
						TierLabel:       iv.TierLabel,
						InputPrice:      iv.InputPrice,
						OutputPrice:     iv.OutputPrice,
						CacheWritePrice: iv.CacheWritePrice,
						CacheReadPrice:  iv.CacheReadPrice,
						PerRequestPrice: iv.PerRequestPrice,
					})
				}
			}
		}
	}
	return m
}

// resolveGroupModelsByAccount 按"account 并集 + LiteLLM 兜底"算法计算 group 的模型列表。
//
//   - 拉该 group 下所有 active account（accountService.ListByGroup）
//   - 对每个 account 取 GetModelMapping()：空 mapping 或全通配符 → 透传，不参与并集；
//     含非通配符 from → 这些 from 就是该 account 显式支持的模型
//   - 所有参与的 account 之间取**并集**（与 /v1/models 的 GetAvailableModels 口径一致：
//     组内任一账号支持的模型都可能被调度命中，展示上应全部列出）
//   - 无参与 / 无 account → LiteLLM 按 platform 兜底
func (h *AvailableChannelHandler) resolveGroupModelsByAccount(
	ctx context.Context,
	groupID int64,
	platform string,
	getLiteLLMModels func(platform string) []string,
) []string {
	var union map[string]struct{} // nil 表示还没遇到任何"显式列模型"的账号

	if h.accountService != nil {
		accounts, err := h.accountService.ListByGroup(ctx, groupID)
		if err == nil {
			for i := range accounts {
				acc := accounts[i]
				if acc.Status != service.StatusActive {
					continue
				}
				if acc.Platform != platform {
					// 防御性：理论上 account.platform 跟 group.platform 应该一致
					continue
				}
				mapping := acc.GetModelMapping()
				if len(mapping) == 0 {
					// 透传，不参与并集
					continue
				}
				for from := range mapping {
					if strings.HasSuffix(from, "*") {
						continue // 通配符 from 不算具体模型
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
		return append([]string(nil), getLiteLLMModels(platform)...)
	}
	out := make([]string, 0, len(union))
	for k := range union {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func positiveFloatPtr(v float64) *float64 {
	if v <= 0 {
		return nil
	}
	return &v
}
