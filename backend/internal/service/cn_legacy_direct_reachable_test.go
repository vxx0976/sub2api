package service

import "testing"

// 这条路径曾经「代码在、测试绿、但运行时永远走不到」——上游的 rawCC 分流排在它前面，
// 而那个分流对 CN 账号看 GetAPIProtocol()，凭证缺失时回落 chat_completions → 恒 true。
// 单测直接调 forwardAnthropicDirect 的下游函数照样能过，给出虚假信心。
// 所以这里钉的是**分流判定本身**：按 ForwardAsAnthropic 的实际 gate 顺序重放一遍。
func TestLegacyCNAccountReachesAnthropicDirect(t *testing.T) {
	// 生产两个账号的真实形态：apikey + api_key + base_url，无 account_mode、无 api_protocol。
	prod := func(platform, baseURL string) *Account {
		return &Account{
			Platform:    platform,
			Type:        AccountTypeAPIKey,
			Credentials: map[string]any{"api_key": "k", "base_url": baseURL},
		}
	}
	// 按 ForwardAsAnthropic 的 gate 顺序判定最终落到哪条分支。
	route := func(a *Account) string {
		if a.IsAnthropicProtocol() || a.IsAdaptiveAPIProtocol() {
			return "native-anthropic"
		}
		if !usesLegacyCNAnthropicDirect(a) && shouldForwardOpenAIResponsesViaRawChatCompletions(a) {
			return "raw-chat-completions"
		}
		if usesLegacyCNAnthropicDirect(a) {
			return "legacy-direct"
		}
		return "responses"
	}

	for _, tc := range []struct {
		name string
		acc  *Account
		want string
	}{
		{"生产 deepseek（无 api_protocol）", prod(PlatformDeepseek, "https://api.deepseek.com"), "legacy-direct"},
		{"生产 kimi（无 api_protocol）", prod(PlatformKimi, "https://api.kimi.com/coding/v1"), "legacy-direct"},
		{"zhipu（无 api_protocol）", prod(PlatformZhipu, "https://open.bigmodel.cn"), "legacy-direct"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := route(tc.acc); got != tc.want {
				t.Errorf("路由到 %q，期望 %q —— 存量 CN 账号的原生直通被截走了", got, tc.want)
			}
		})
	}

	// 显式配了协议就必须交给上游分流，不能再被存量直通劫持。
	withProto := func(platform, proto string) *Account {
		a := prod(platform, "https://api.deepseek.com")
		a.Credentials["api_protocol"] = proto
		return a
	}
	for _, tc := range []struct {
		name string
		acc  *Account
		want string
	}{
		{"显式 anthropic", withProto(PlatformDeepseek, APIProtocolAnthropic), "native-anthropic"},
		{"显式 adaptive", withProto(PlatformDeepseek, APIProtocolAdaptive), "native-anthropic"},
		{"显式 chat_completions", withProto(PlatformDeepseek, APIProtocolChatCompletions), "raw-chat-completions"},
		{"显式 responses（deepseek 独有）", withProto(PlatformDeepseek, APIProtocolResponses), "responses"},
		{"kimi 显式 responses → 回落 CC", withProto(PlatformKimi, APIProtocolResponses), "raw-chat-completions"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := route(tc.acc); got != tc.want {
				t.Errorf("路由到 %q，期望 %q", got, tc.want)
			}
		})
	}
}
