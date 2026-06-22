package service

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

// SettingKeyModelPricingOverrides 是 settings 表里存放“通用模型价格覆盖表”的 key。
const SettingKeyModelPricingOverrides = "model_pricing_overrides"

// modelPricingOverride 是单条用户自定义模型价格覆盖（价格为每百万 token，币种 CNY/USD）。
type modelPricingOverride struct {
	Model      string  `json:"model"`
	Currency   string  `json:"currency"` // CurrencyCNY | CurrencyUSD
	InputPerM  float64 `json:"input"`
	OutputPerM float64 `json:"output"`
	CachePerM  float64 `json:"cache"`
	HasCache   bool    `json:"has_cache"`
	Enabled    bool    `json:"enabled"`
}

// ModelPricingOverridesDTO 是 admin GET/PUT 的载荷（整张表覆盖式读写）。
type ModelPricingOverridesDTO struct {
	Entries []modelPricingOverride `json:"entries"`
}

// currentPricingService 让包级函数 ModelPriceCurrency 也能查覆盖币种（进程内单例，
// 由 SetSettingRepository 登记）。
var currentPricingService atomic.Pointer[PricingService]

// SetSettingRepository 注入 settings 仓储，并把本实例登记为进程级定价服务。
func (s *PricingService) SetSettingRepository(repo SettingRepository) {
	if s == nil {
		return
	}
	s.settingRepo = repo
	currentPricingService.Store(s)
}

// loadOverrides 读取覆盖表，带独立 overrideMu + 60s TTL 进程缓存。
//
// ⚠️ 绝不能复用 s.mu：GetModelPricing 全程持有 s.mu.RLock 并在其中调用本方法，
// 若复用 s.mu 刷新缓存（写锁）会重入死锁。
func (s *PricingService) loadOverrides() []modelPricingOverride {
	if s == nil || s.settingRepo == nil {
		return nil
	}
	now := time.Now().UnixNano()

	s.overrideMu.RLock()
	if s.overrideLoadedAt != 0 && now < s.overrideLoadedAt {
		cached := s.overrideCache
		s.overrideMu.RUnlock()
		return cached
	}
	s.overrideMu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyModelPricingOverrides)
	var list []modelPricingOverride
	if err == nil && strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &list) // 未配置 / 坏 JSON => 空表，安全回退原链
	}

	s.overrideMu.Lock()
	s.overrideCache = list
	s.overrideLoadedAt = now + int64(60*time.Second)
	s.overrideMu.Unlock()
	return list
}

// InvalidateOverrideCache 置缓存失效，使下次读立即重载（admin 写后调用，即时生效）。
func (s *PricingService) InvalidateOverrideCache() {
	if s == nil {
		return
	}
	s.overrideMu.Lock()
	s.overrideLoadedAt = 0
	s.overrideMu.Unlock()
}

// matchOverride 精确匹配优先，无则按最长前缀回退（仅 enabled 项）。modelLower 须已小写。
func (s *PricingService) matchOverride(modelLower string) (modelPricingOverride, bool) {
	list := s.loadOverrides()
	if len(list) == 0 {
		return modelPricingOverride{}, false
	}
	// 第一遍：精确匹配
	for _, ov := range list {
		if ov.Enabled && strings.ToLower(strings.TrimSpace(ov.Model)) == modelLower {
			return ov, true
		}
	}
	// 第二遍：最长前缀回退（按 Model 长度降序，避免短前缀 "glm" 抢先于 "glm-4.6"）
	cand := make([]modelPricingOverride, 0, len(list))
	for _, ov := range list {
		if ov.Enabled && strings.TrimSpace(ov.Model) != "" {
			cand = append(cand, ov)
		}
	}
	sort.SliceStable(cand, func(i, j int) bool { return len(cand[i].Model) > len(cand[j].Model) })
	for _, ov := range cand {
		if strings.HasPrefix(modelLower, strings.ToLower(strings.TrimSpace(ov.Model))) {
			return ov, true
		}
	}
	return modelPricingOverride{}, false
}

// overrideToLiteLLM 把覆盖项折算成内部“每 token 美元价”：CNY 按运行时汇率折算，
// USD 直接用（不除汇率）。与 qwenPricingOverride 口径一致。
func (s *PricingService) overrideToLiteLLM(ov modelPricingOverride) *LiteLLMModelPricing {
	const perToken = 1.0 / 1_000_000.0
	rate := 1.0
	if ov.Currency == CurrencyCNY {
		rate = s.cnyToUSDRate()
	}
	p := &LiteLLMModelPricing{LiteLLMProvider: "override", Mode: "chat"}
	p.InputCostPerToken = ov.InputPerM / rate * perToken
	p.OutputCostPerToken = ov.OutputPerM / rate * perToken
	if ov.HasCache {
		p.CacheReadInputTokenCost = ov.CachePerM / rate * perToken
		p.SupportsPromptCaching = true
	}
	return p
}

// GetModelPricingOverrides 返回当前覆盖表（供 admin API）。
func (s *PricingService) GetModelPricingOverrides(_ context.Context) (*ModelPricingOverridesDTO, error) {
	list := s.loadOverrides()
	if list == nil {
		list = []modelPricingOverride{}
	}
	return &ModelPricingOverridesDTO{Entries: list}, nil
}

// UpdateModelPricingOverrides 整表覆盖写入 settings 并即时生效（供 admin API）。
func (s *PricingService) UpdateModelPricingOverrides(ctx context.Context, in *ModelPricingOverridesDTO) (*ModelPricingOverridesDTO, error) {
	if s == nil || s.settingRepo == nil {
		return nil, errors.New("setting repository not initialized")
	}
	if in == nil {
		in = &ModelPricingOverridesDTO{}
	}
	norm := make([]modelPricingOverride, 0, len(in.Entries))
	for _, e := range in.Entries {
		e.Model = strings.TrimSpace(e.Model)
		if e.Model == "" {
			continue
		}
		if e.Currency != CurrencyCNY && e.Currency != CurrencyUSD {
			e.Currency = CurrencyCNY
		}
		if e.InputPerM < 0 {
			e.InputPerM = 0
		}
		if e.OutputPerM < 0 {
			e.OutputPerM = 0
		}
		if e.CachePerM < 0 {
			e.CachePerM = 0
		}
		norm = append(norm, e)
	}
	b, err := json.Marshal(norm)
	if err != nil {
		return nil, err
	}
	if err := s.settingRepo.Set(ctx, SettingKeyModelPricingOverrides, string(b)); err != nil {
		return nil, err
	}
	s.InvalidateOverrideCache()
	return &ModelPricingOverridesDTO{Entries: norm}, nil
}
