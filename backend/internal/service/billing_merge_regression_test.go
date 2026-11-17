//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Exercise Resolve -> CalculateCostUnified -> the source/model eligibility gate.
// Calling computeTokenBreakdown directly would miss the channel/group regression.
func TestBillingMergeRegression_AnthropicFastConfiguredPricing(t *testing.T) {
	models := []struct {
		name string
		fast bool
	}{
		{"claude-opus-4.8", true},
		{"claude-opus-4-8", true},
		{"claude-opus-5", true},
		{"claude-opus-5-5", true},
		{"claude-fable-5-2", true},
		{"claude-mythos-5-2", true},
		{"claude-opus-4-5", false},
		{"claude-sonnet-4-6", false},
	}
	for _, model := range models {
		for _, source := range []string{PricingSourceChannel, PricingSourceGroup} {
			for _, override := range []struct {
				name       string
				multiplier *float64
			}{
				{"default", nil},
				{"explicit_standard", pricingMultiplier(1)},
				{"explicit_premium", pricingMultiplier(3.25)},
			} {
				t.Run(model.name+"/"+source+"/"+override.name, func(t *testing.T) {
					svc := newTestBillingService()
					card := ChannelModelPricing{
						Models: []string{model.name}, Platform: PlatformAnthropic, BillingMode: BillingModeToken,
						InputPrice: pricingMultiplier(7e-6), OutputPrice: pricingMultiplier(11e-6),
						CacheWritePrice: pricingMultiplier(13e-6), CacheReadPrice: pricingMultiplier(17e-6),
						FastMultiplier: override.multiplier,
					}
					originalCard := card.Clone()
					group := &Group{ID: 42, Platform: PlatformAnthropic}
					var channels *ChannelService
					if source == PricingSourceGroup {
						group.ModelPricing = []ChannelModelPricing{card}
					} else {
						cache := newEmptyChannelCache()
						channel := &Channel{ID: 1, Status: StatusActive, ModelPricing: []ChannelModelPricing{card}}
						expandPricingToCache(cache, channel, group.ID, group.Platform)
						cache.channelByGroupID[group.ID] = channel
						cache.groupPlatform[group.ID] = group.Platform
						cache.loadedAt = time.Now()
						channels = &ChannelService{}
						channels.cache.Store(cache)
					}
					resolver := NewModelPricingResolver(channels, svc)
					resolved := resolver.Resolve(context.Background(), PricingInput{Model: model.name, GroupID: &group.ID, Group: group})
					require.Equal(t, source, resolved.Source)
					require.NotNil(t, resolved.BasePricing)
					originalPricing := *resolved.BasePricing
					for _, tier := range []string{"", "fast", "priority", "flex", "default"} {
						t.Run("tier_"+tier, func(t *testing.T) {
							multiplier := 1.0
							if tier == "fast" || tier == "priority" {
								if override.multiplier != nil {
									multiplier = *override.multiplier
								} else if model.fast {
									multiplier = 2
								}
							}
							cost, err := svc.CalculateCostUnified(CostInput{
								Ctx: context.Background(), Model: model.name, GroupID: &group.ID, Group: group,
								Resolver: resolver, Resolved: resolved, ServiceTier: tier, RateMultiplier: 1.25,
								Tokens: UsageTokens{InputTokens: 1000, OutputTokens: 200, CacheCreationTokens: 300, CacheReadTokens: 400},
							})
							require.NoError(t, err)
							require.InDelta(t, 0.007*multiplier, cost.InputCost, 1e-12)
							require.InDelta(t, 0.0022*multiplier, cost.OutputCost, 1e-12)
							require.InDelta(t, 0.0039*multiplier, cost.CacheCreationCost, 1e-12)
							require.InDelta(t, 0.0068*multiplier, cost.CacheReadCost, 1e-12)
							require.InDelta(t, 0.0199*multiplier, cost.TotalCost, 1e-12)
							require.InDelta(t, 0.0199*multiplier*1.25, cost.ActualCost, 1e-12)
							require.Equal(t, string(BillingModeToken), cost.BillingMode)
						})
					}
					// Reusing the resolved card must not leak an injected multiplier into it.
					require.Equal(t, originalPricing, *resolved.BasePricing)
					require.Equal(t, originalCard, card)
				})
			}
		}
	}
}
