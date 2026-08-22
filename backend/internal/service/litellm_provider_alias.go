package service

import "strings"

// litellmProviderAliases 把本站的平台标识映射到 LiteLLM 价目表里的 litellm_provider。
//
// 两套 ID 是各自独立演进的命名空间，不保证同名：
//   - 本站平台标识随上游 sub2api 走（见 AllowedQuotaPlatforms）；
//   - litellm_provider 随 LiteLLM 上游价目表走，我们只是消费方，改不了。
//
// 目前唯一对不上的是 Kimi：LiteLLM 至今把月之暗面的模型标为 "moonshot"
// （运行时同步缓存里 moonshot/kimi-* 与 moonshot/moonshot-v1-* 共 20+ 条，
// litellm_provider="kimi" 一条都没有）。迁移 226 把本站平台从 moonshot 改名成 kimi 之后，
// 没有这层别名，ListAll("kimi") 会恒返回空 —— 模型广场与公开定价页的兜底链
// （渠道为空 → 账号全透传 → 回落平台全表）拿不到任何模型名，
// `if len(pg.Models) == 0 { continue }` 会把整个 Kimi 分组从公开页面上抹掉。
//
// zhipu / deepseek 两侧同名，无需别名；这里只登记真正有分歧的。
// 今后 LiteLLM 若补上 "kimi" provider，这条别名仍然安全：它只放宽匹配，不排除新值。
var litellmProviderAliases = map[string]string{
	PlatformKimi: "moonshot",
}

// litellmProviderMatches 报告价目表条目的 litellm_provider 是否命中给定的平台过滤值。
// 平台标识本身与其 LiteLLM 别名都算命中，均大小写不敏感。
func litellmProviderMatches(entryProvider, filter string) bool {
	if strings.EqualFold(entryProvider, filter) {
		return true
	}
	if alias, ok := litellmProviderAliases[strings.ToLower(strings.TrimSpace(filter))]; ok {
		return strings.EqualFold(entryProvider, alias)
	}
	return false
}
