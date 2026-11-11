package apicompat

// NOTE(fork): gpt-6.1-sol 在 fork 的兼容路径上会把 none/minimal 强制改成 medium（见
// CoerceGPT61SolReasoningEffort 与 fork 自己的 TestGPT61SolCoercesDisabledReasoningToMedium），
// 因此这里断言 fork 口径：不报错、effort=medium、summary=auto。延后校验行为由 terra 等其他
// 6.1 模型与 service 侧用例覆盖。

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnthropicToResponsesDeferredEffortValidation(t *testing.T) {
	for _, effort := range []string{"", "low", "high", "max"} {
		t.Run(effort, func(t *testing.T) {
			req := &AnthropicRequest{
				Model:        "gpt-6.1-sol",
				MaxTokens:    1,
				Thinking:     &AnthropicThinking{Type: "disabled"},
				OutputConfig: &AnthropicOutputConfig{Effort: effort},
			}
			// fork：gpt-6.1-sol 的 none/minimal 被强制成 medium，不报错。
			_, err := AnthropicToResponses(req)
			require.NoError(t, err)
			_, err = AnthropicToResponsesWithMinOutputTokens(req, 32)
			require.NoError(t, err)
			converted, err := AnthropicToResponsesWithOptions(req, AnthropicToResponsesOptions{
				MinOutputTokens:                32,
				DeferReasoningEffortValidation: true,
			})
			require.NoError(t, err)
			require.Equal(t, "gpt-6.1-sol", converted.Model)
			require.Equal(t, "medium", converted.Reasoning.Effort)
			require.Equal(t, "auto", converted.Reasoning.Summary) // fork 强制 medium 时会带 summary=auto
			require.Equal(t, 32, *converted.MaxOutputTokens)
			require.Equal(t, "disabled", req.Thinking.Type)
			require.Equal(t, effort, req.OutputConfig.Effort)
		})
	}
}
