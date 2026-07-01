package modelname

import "testing"

func TestFormatDisplayName(t *testing.T) {
	cases := []struct {
		id   string
		want string
	}{
		// The reported case: a model pulled from upstream that is not in the
		// curated catalog should still render like the built-ins.
		{"claude-sonnet-5", "Claude Sonnet 5"},
		{"claude-opus-5", "Claude Opus 5"},

		// Version pairs must render with a dot, matching the built-in catalog.
		{"claude-sonnet-4-6", "Claude Sonnet 4.6"},
		{"claude-opus-4-8", "Claude Opus 4.8"},
		{"claude-3-5-sonnet", "Claude 3.5 Sonnet"},

		// Trailing compact YYYYMMDD date suffix is dropped (catalog convention).
		{"claude-opus-4-5-20251101", "Claude Opus 4.5"},
		{"claude-3-5-sonnet-20241022", "Claude 3.5 Sonnet"},
		{"claude-opus-4-20250514", "Claude Opus 4"},

		// Trailing dashed YYYY-MM-DD date suffix is dropped, NOT turned into a
		// bogus "04.09" version by the numeric coalescing rule.
		{"gpt-4-turbo-2024-04-09", "GPT 4 Turbo"},
		{"gpt-4o-2024-08-06", "GPT 4o"},
		{"o3-2025-04-16", "O3"},
		{"o4-mini-2025-04-16", "O4 Mini"},
		{"qwen-plus-2025-09-11", "Qwen Plus"},
		// A bare date with no name token is left alone (degenerate input).
		{"2024-04-09", "2024 04.09"},
		// Date-drop requires zero-padded MM/DD; a single-digit tail is a version,
		// not a date, so it is preserved rather than stripped.
		{"gpt-2024-1-5", "GPT 2024 1.5"},

		// Brand acronyms for platforms this deployment actually syncs.
		{"glm-4.6", "GLM 4.6"},
		{"deepseek-v4-flash", "DeepSeek V4 Flash"},

		// Other platforms.
		{"gpt-5.4", "GPT 5.4"},
		{"gemini-2.5-flash", "Gemini 2.5 Flash"},
		{"gemini-2.5-flash-image", "Gemini 2.5 Flash Image"},
		{"grok-4-fast", "Grok 4 Fast"},

		// 4-digit numeric fragments must NOT be coalesced into a bogus version.
		{"gpt-4-1106-preview", "GPT 4 1106 Preview"},

		// Digit-leading sub-tokens are left as-is.
		{"gpt-4o", "GPT 4o"},

		// Edge cases.
		{"", ""},
		{"   ", "   "},
		{"claude--sonnet", "Claude Sonnet"},
		{"20250101", "20250101"}, // single date-only token is preserved
		{"grok", "Grok"},
	}

	for _, tc := range cases {
		if got := FormatDisplayName(tc.id); got != tc.want {
			t.Errorf("FormatDisplayName(%q) = %q, want %q", tc.id, got, tc.want)
		}
	}
}
