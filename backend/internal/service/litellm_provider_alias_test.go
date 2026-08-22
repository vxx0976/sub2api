package service

import "testing"

// 迁移 226 把本站平台 moonshot 改名成 kimi，但 LiteLLM 价目表里 Kimi 系模型的
// litellm_provider 仍是 "moonshot"。没有别名层，ListAll("kimi") 恒空 →
// 模型广场 / 公开定价页的兜底链拿不到模型名 → 整个 Kimi 分组从公开页面消失。
func TestLiteLLMProviderMatchesKimiAlias(t *testing.T) {
	cases := []struct {
		name          string
		entryProvider string
		filter        string
		want          bool
	}{
		{"kimi 过滤命中 LiteLLM 的 moonshot", "moonshot", "kimi", true},
		{"大小写不敏感", "MoonShot", "KIMI", true},
		{"带空白的过滤值", "moonshot", "  kimi  ", true},
		{"LiteLLM 将来补上 kimi 也命中", "kimi", "kimi", true},
		{"zhipu 两侧同名", "zhipu", "zhipu", true},
		{"deepseek 两侧同名", "deepseek", "deepseek", true},

		// 别名只放宽 kimi 一侧，不能把 moonshot 变成通配。
		{"别名不反向生效：moonshot 过滤不吃 kimi 条目", "kimi", "moonshot", false},
		{"别名不泄漏到别的平台", "moonshot", "zhipu", false},
		{"无关 provider 不命中", "anthropic", "kimi", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := litellmProviderMatches(tc.entryProvider, tc.filter); got != tc.want {
				t.Errorf("litellmProviderMatches(%q, %q) = %v, want %v",
					tc.entryProvider, tc.filter, got, tc.want)
			}
		})
	}
}

// 别名表的键必须是真实平台标识，否则是死条目（改名/下线后没人清理）。
func TestLiteLLMProviderAliasKeysAreRealPlatforms(t *testing.T) {
	for platform := range litellmProviderAliases {
		if !IsAllowedQuotaPlatform(platform) {
			t.Errorf("别名表里的 %q 不是 AllowedQuotaPlatforms 里的平台，是死条目", platform)
		}
	}
}
