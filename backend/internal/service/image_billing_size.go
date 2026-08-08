package service

import (
	"sort"
	"strconv"
	"strings"
)

const (
	ImageBillingSize1K = "1K"
	ImageBillingSize2K = "2K"
	ImageBillingSize4K = "4K"

	ImageSizeSourceOutput  = "output"
	ImageSizeSourceInput   = "input"
	ImageSizeSourceDefault = "default"
	ImageSizeSourceLegacy  = "legacy"
	// ImageSizeSourceCapped 表示实际出图档位高于请求里显式指定的尺寸档，已按请求档封顶计费。
	// 典型场景：ChatGPT OAuth 后端把 size 归一成 auto，无论请求 1024x1024 与否都恒定出
	// ~1.57MP 的图（1254x1254 / 1122x1402 …），最长边永远 >1024 会被判成 2K。
	ImageSizeSourceCapped = "capped"
)

type ImageBillingSizeResolution struct {
	BillingSize string
	InputSize   string
	OutputSize  string
	Source      string
	Breakdown   map[string]int
}

func ClassifyImageBillingTier(size string) (string, bool) {
	trimmed := strings.TrimSpace(size)
	normalized := strings.ToLower(trimmed)
	switch normalized {
	case "", "auto":
		return "", false
	case "1k":
		return ImageBillingSize1K, true
	case "2k":
		return ImageBillingSize2K, true
	case "4k":
		return ImageBillingSize4K, true
	case "2048x2048", "2048x1152":
		return ImageBillingSize2K, true
	case "3840x2160", "2160x3840":
		return ImageBillingSize4K, true
	}

	width, height, ok := parseImageBillingDimensions(trimmed)
	if !ok {
		return "", false
	}
	maxEdge := width
	if height > maxEdge {
		maxEdge = height
	}
	switch {
	case maxEdge <= 1024:
		return ImageBillingSize1K, true
	case maxEdge <= 2048:
		return ImageBillingSize2K, true
	default:
		return ImageBillingSize4K, true
	}
}

func NormalizeImageBillingTierOrDefault(size string) string {
	if tier, ok := ClassifyImageBillingTier(size); ok {
		return tier
	}
	return ImageBillingSize2K
}

// ResolveImageBillingSize 决定这次出图按哪一档计费。
// capFromRequest 表示 inputSize 是「确实会被转发给上游的请求字段」（如 /v1/images 的 size、
// /v1/responses 里 tools[].image_generation.size）。只有这种尺寸才允许对实际出图档位封顶——
// 不透传的字段（如 /v1/responses 请求体顶层的 size）上游根本看不到，拿它封顶等于让客户
// 随手加一行就把大图按小图收费。
func ResolveImageBillingSize(inputSize string, outputSizes []string, capFromRequest bool) ImageBillingSizeResolution {
	inputSize = strings.TrimSpace(inputSize)
	outputSizes = compactTrimmedStrings(outputSizes)

	requestTier, hasRequestTier := ClassifyImageBillingTier(inputSize)
	hasRequestTier = hasRequestTier && capFromRequest

	breakdown := map[string]int{}
	outputSize := firstDisplayImageOutputSize(outputSizes)
	billedTier := ""
	capped := false
	for _, output := range outputSizes {
		tier, ok := ClassifyImageBillingTier(output)
		if !ok {
			continue
		}
		// 上游忽略请求里的显式尺寸时不加价：单张按 min(实际出图档, 请求档) 计费。
		if hasRequestTier && imageTierRank(tier) > imageTierRank(requestTier) {
			tier = requestTier
			capped = true
		}
		breakdown[tier]++
		if imageTierRank(tier) > imageTierRank(billedTier) {
			billedTier = tier
		}
	}
	if billedTier != "" {
		source := ImageSizeSourceOutput
		if capped {
			source = ImageSizeSourceCapped
		}
		return ImageBillingSizeResolution{
			BillingSize: billedTier,
			InputSize:   inputSize,
			OutputSize:  outputSize,
			Source:      source,
			Breakdown:   normalizeImageSizeBreakdown(breakdown),
		}
	}

	if tier, ok := ClassifyImageBillingTier(inputSize); ok {
		return ImageBillingSizeResolution{
			BillingSize: tier,
			InputSize:   inputSize,
			OutputSize:  outputSize,
			Source:      ImageSizeSourceInput,
		}
	}

	return ImageBillingSizeResolution{
		BillingSize: ImageBillingSize2K,
		InputSize:   inputSize,
		OutputSize:  outputSize,
		Source:      ImageSizeSourceDefault,
	}
}

func ApplyOpenAIImageBillingResolution(result *OpenAIForwardResult) {
	if result == nil || result.ImageCount <= 0 {
		return
	}
	requestSize := strings.TrimSpace(result.ImageInputSize)
	inputSize := requestSize
	if inputSize == "" && strings.TrimSpace(result.ImageSize) != ImageBillingSize2K {
		inputSize = strings.TrimSpace(result.ImageSize)
	}
	outputSizes := result.ImageOutputSizes
	if len(outputSizes) == 0 && strings.TrimSpace(result.ImageOutputSize) != "" {
		outputSizes = []string{result.ImageOutputSize}
	}
	// 只有请求里原样记下来的尺寸能封顶；上面那条回退拿的是 ImageSize（可能是响应侧推导出的
	// 档位），只用于兜底选档，不能当客户的报价依据。
	resolved := ResolveImageBillingSize(inputSize, outputSizes, requestSize != "")
	// 回写请求原值而非兜底值，避免重复调用时把兜底档位当成「客户请求档」而越算越便宜。
	resolved.InputSize = requestSize
	applyImageBillingResolution(
		&result.ImageSize,
		&result.ImageInputSize,
		&result.ImageOutputSize,
		&result.ImageSizeSource,
		&result.ImageSizeBreakdown,
		resolved,
	)
}

func ApplyForwardImageBillingResolution(result *ForwardResult) {
	if result == nil || result.ImageCount <= 0 {
		return
	}
	requestSize := strings.TrimSpace(result.ImageInputSize)
	inputSize := requestSize
	if inputSize == "" && strings.TrimSpace(result.ImageSize) != ImageBillingSize2K {
		inputSize = strings.TrimSpace(result.ImageSize)
	}
	outputSizes := result.ImageOutputSizes
	if len(outputSizes) == 0 && strings.TrimSpace(result.ImageOutputSize) != "" {
		outputSizes = []string{result.ImageOutputSize}
	}
	// 同 ApplyOpenAIImageBillingResolution：兜底值只用于选档，不能当封顶依据，也不回写。
	resolved := ResolveImageBillingSize(inputSize, outputSizes, requestSize != "")
	resolved.InputSize = requestSize
	applyImageBillingResolution(
		&result.ImageSize,
		&result.ImageInputSize,
		&result.ImageOutputSize,
		&result.ImageSizeSource,
		&result.ImageSizeBreakdown,
		resolved,
	)
}

func applyImageBillingResolution(
	billingSize *string,
	inputSize *string,
	outputSize *string,
	source *string,
	breakdown *map[string]int,
	resolved ImageBillingSizeResolution,
) {
	*billingSize = resolved.BillingSize
	*inputSize = resolved.InputSize
	*outputSize = resolved.OutputSize
	*source = resolved.Source
	*breakdown = resolved.Breakdown
}

func parseImageBillingDimensions(size string) (int, int, bool) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(size)), "x")
	if len(parts) != 2 {
		return 0, 0, false
	}
	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, false
	}
	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, false
	}
	if width <= 0 || height <= 0 {
		return 0, 0, false
	}
	return width, height, true
}

func compactTrimmedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func firstDisplayImageOutputSize(outputSizes []string) string {
	for _, output := range outputSizes {
		if trimmed := strings.TrimSpace(output); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func imageTierRank(tier string) int {
	switch strings.ToUpper(strings.TrimSpace(tier)) {
	case ImageBillingSize1K:
		return 1
	case ImageBillingSize2K:
		return 2
	case ImageBillingSize4K:
		return 3
	default:
		return 0
	}
}

func normalizeImageSizeBreakdown(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for _, tier := range []string{ImageBillingSize1K, ImageBillingSize2K, ImageBillingSize4K} {
		if count := in[tier]; count > 0 {
			out[tier] = count
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func SortedImageBillingBreakdownKeys(breakdown map[string]int) []string {
	keys := make([]string, 0, len(breakdown))
	for key := range breakdown {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left, right := imageTierRank(keys[i]), imageTierRank(keys[j])
		if left == right {
			return keys[i] < keys[j]
		}
		return left < right
	})
	return keys
}
