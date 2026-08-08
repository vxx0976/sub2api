package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifyImageBillingTier(t *testing.T) {
	tests := []struct {
		name     string
		size     string
		wantTier string
		wantOK   bool
	}{
		{name: "explicit 2k square", size: "2048x2048", wantTier: "2K", wantOK: true},
		{name: "explicit 2k landscape", size: "2048x1152", wantTier: "2K", wantOK: true},
		{name: "explicit 4k landscape", size: "3840x2160", wantTier: "4K", wantOK: true},
		{name: "explicit 4k portrait", size: "2160x3840", wantTier: "4K", wantOK: true},
		{name: "long edge 1k", size: "1024X768", wantTier: "1K", wantOK: true},
		{name: "long edge 2k", size: "1280x768", wantTier: "2K", wantOK: true},
		{name: "long edge 4k", size: "2560x1600", wantTier: "4K", wantOK: true},
		{name: "tier string 1k", size: "1k", wantTier: "1K", wantOK: true},
		{name: "empty", size: "", wantOK: false},
		{name: "auto", size: "auto", wantOK: false},
		{name: "invalid", size: "not-a-size", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTier, gotOK := ClassifyImageBillingTier(tt.size)
			require.Equal(t, tt.wantOK, gotOK)
			require.Equal(t, tt.wantTier, gotTier)
		})
	}
}

func TestResolveImageBillingSize(t *testing.T) {
	tests := []struct {
		name             string
		inputSize        string
		outputSizes      []string
		noCapFromRequest bool
		wantBilling      string
		wantOutput       string
		wantSource       string
		wantBreakdown    map[string]int
	}{
		{
			name:          "output wins over input when within requested tier",
			inputSize:     "3840x2160",
			outputSizes:   []string{"1024x1024"},
			wantBilling:   "1K",
			wantOutput:    "1024x1024",
			wantSource:    ImageSizeSourceOutput,
			wantBreakdown: map[string]int{"1K": 1},
		},
		{
			name:          "output above requested tier is capped to request",
			inputSize:     "1024x1024",
			outputSizes:   []string{"3840x2160"},
			wantBilling:   "1K",
			wantOutput:    "3840x2160",
			wantSource:    ImageSizeSourceCapped,
			wantBreakdown: map[string]int{"1K": 1},
		},
		{
			// ChatGPT OAuth 把 size 归一成 auto，请求 1024x1024 实际出 1254x1254。
			name:          "chatgpt auto normalization capped to requested 1k",
			inputSize:     "1024x1024",
			outputSizes:   []string{"1254x1254"},
			wantBilling:   "1K",
			wantOutput:    "1254x1254",
			wantSource:    ImageSizeSourceCapped,
			wantBreakdown: map[string]int{"1K": 1},
		},
		{
			name:          "output within requested 2k is not capped",
			inputSize:     "2048x2048",
			outputSizes:   []string{"1254x1254"},
			wantBilling:   "2K",
			wantOutput:    "1254x1254",
			wantSource:    ImageSizeSourceOutput,
			wantBreakdown: map[string]int{"2K": 1},
		},
		{
			name:          "no request tier means no cap",
			inputSize:     "auto",
			outputSizes:   []string{"1254x1254"},
			wantBilling:   "2K",
			wantOutput:    "1254x1254",
			wantSource:    ImageSizeSourceOutput,
			wantBreakdown: map[string]int{"2K": 1},
		},
		{
			// 尺寸不是透传给上游的字段（如 /responses 请求体顶层 size）时不许封顶，
			// 否则客户随手加一行就能把大图按小图收费。
			name:             "size not forwarded upstream never caps",
			inputSize:        "1024x1024",
			outputSizes:      []string{"3840x2160"},
			noCapFromRequest: true,
			wantBilling:      "4K",
			wantOutput:       "3840x2160",
			wantSource:       ImageSizeSourceOutput,
			wantBreakdown:    map[string]int{"4K": 1},
		},
		{
			name:        "input fallback",
			inputSize:   "1024x1024",
			wantBilling: "1K",
			wantSource:  ImageSizeSourceInput,
		},
		{
			name:        "auto defaults",
			inputSize:   "auto",
			wantBilling: "2K",
			wantSource:  ImageSizeSourceDefault,
		},
		{
			name:        "empty defaults",
			inputSize:   "",
			wantBilling: "2K",
			wantSource:  ImageSizeSourceDefault,
		},
		{
			name:        "invalid defaults",
			inputSize:   "largest",
			wantBilling: "2K",
			wantSource:  ImageSizeSourceDefault,
		},
		{
			name:          "mixed output chooses highest tier",
			inputSize:     "",
			outputSizes:   []string{"1024x1024", "3840x2160", "1280x720"},
			wantBilling:   "4K",
			wantOutput:    "1024x1024",
			wantSource:    ImageSizeSourceOutput,
			wantBreakdown: map[string]int{"1K": 1, "2K": 1, "4K": 1},
		},
		{
			name:          "mixed output capped per image",
			inputSize:     "1024x1024",
			outputSizes:   []string{"1024x1024", "3840x2160", "1280x720"},
			wantBilling:   "1K",
			wantOutput:    "1024x1024",
			wantSource:    ImageSizeSourceCapped,
			wantBreakdown: map[string]int{"1K": 3},
		},
		{
			name:          "mixed output capped only above requested tier",
			inputSize:     "2048x2048",
			outputSizes:   []string{"1024x1024", "3840x2160"},
			wantBilling:   "2K",
			wantOutput:    "1024x1024",
			wantSource:    ImageSizeSourceCapped,
			wantBreakdown: map[string]int{"1K": 1, "2K": 1},
		},
		{
			name:        "unparseable output falls back to parseable input",
			inputSize:   "2048x1152",
			outputSizes: []string{"auto"},
			wantBilling: "2K",
			wantOutput:  "auto",
			wantSource:  ImageSizeSourceInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveImageBillingSize(tt.inputSize, tt.outputSizes, !tt.noCapFromRequest)
			require.Equal(t, tt.wantBilling, got.BillingSize)
			require.Equal(t, tt.inputSize, got.InputSize)
			require.Equal(t, tt.wantOutput, got.OutputSize)
			require.Equal(t, tt.wantSource, got.Source)
			require.Equal(t, tt.wantBreakdown, got.Breakdown)
		})
	}
}
