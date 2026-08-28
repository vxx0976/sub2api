package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// 按图片计费的模型在广场必须仍以图片档位展示，不得被 token 阶梯表覆盖。
//
// 本 fork 的模型目录挂在账号上（model_plaza_account_fallback.go →
// synthesizePricingFromLiteLLM 判出 BillingModeImage）。上游新增的 fillDisplayPricing
// 会先试 ResolveContextPricingSchedule，而绝大多数 image_generation 模型在 LiteLLM
// 目录里同时带 token 价，阶梯表因此会成功返回，plazaPricingFromSchedule 再无条件把
// BillingMode 写成 token —— 公开定价页显示每百万 token 价，实收却按每张图收，
// 两者差一个数量级。
func TestPlazaKeepsImageBillingModeAndTiers(t *testing.T) {
	svc := &ModelPlazaService{}
	group := &Group{ID: 1, Platform: PlatformOpenAI}

	imgIn, imgOut := 0.011, 0.042
	tokIn, tokOut := 5e-06, 1e-05
	imagePricing := &ChannelModelPricing{
		BillingMode:      BillingModeImage,
		ImageInputPrice:  &imgIn,
		ImageOutputPrice: &imgOut,
		InputPrice:       &tokIn, // LiteLLM 目录里同时带的 token 价，正是诱发覆盖的那部分
		OutputPrice:      &tokOut,
	}
	m := &PlazaModel{Name: "gpt-image-2", Platform: PlatformOpenAI, Pricing: imagePricing}

	// billingService / resolver 为 nil 时也必须早退（不能依赖它们缺席才正确）
	svc.fillDisplayPricing(context.Background(), m, group)

	require.NotNil(t, m.Pricing)
	require.Equal(t, BillingModeImage, m.Pricing.BillingMode,
		"图片计费模型的 BillingMode 不得被改写成 token")
	require.Empty(t, m.LongContextBasis, "图片模型不应带长上下文档位说明")
	require.Nil(t, m.TimePricing, "图片模型不应带分时倍率")
}

// 非图片模型不受影响：仍按原路径处理。
func TestPlazaTokenModelUnaffectedByImageGuard(t *testing.T) {
	svc := &ModelPlazaService{} // billingService/resolver 为 nil → 落到 plazaImageDisplayPricing
	group := &Group{ID: 1, Platform: PlatformOpenAI}
	tIn, tOut := 2.5e-06, 1e-05
	m := &PlazaModel{
		Name:     "gpt-5.4",
		Platform: PlatformOpenAI,
		Pricing:  &ChannelModelPricing{BillingMode: BillingModeToken, InputPrice: &tIn, OutputPrice: &tOut},
	}
	svc.fillDisplayPricing(context.Background(), m, group)
	require.NotNil(t, m.Pricing)
	require.Equal(t, BillingModeToken, m.Pricing.BillingMode)
	require.NotNil(t, m.Pricing.InputPrice)
	require.InDelta(t, 2.5e-06, *m.Pricing.InputPrice, 1e-12)
}
