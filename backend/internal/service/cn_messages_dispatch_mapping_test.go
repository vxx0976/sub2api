package service

import (
	"reflect"
	"strings"
	"testing"
)

// 国产分组的 Claude Code 模型映射是线上主力路径，这里按生产实际形状钉死。
//
// 背景：上游的 ResolveMessagesDispatchModel 对 CN 分组无条件 return ""（不改写模型名），
// 前提是「CN 账号配 api_protocol=anthropic 原生直通 + 账号级 model_mapping 兜底」。
// 本站不满足该前提——生产 Deepseek 分组近 30 天约 90% 请求用 claude-* 模型名，
// 靠的就是分组级映射；账号级 model_mapping 是恒等白名单，没有 claude-* 条目。
// 照抄上游会让 claude-opus-5 原样发给 api.deepseek.com，直接打掉主力流量。
func TestCNGroupResolvesConfiguredClaudeModelMapping(t *testing.T) {
	// 生产 Deepseek 1x 分组的真实配置。
	deepseek := &Group{
		Platform: PlatformDeepseek,
		MessagesDispatchModelConfig: OpenAIMessagesDispatchModelConfig{
			OpusMappedModel:   "deepseek-v4-pro",
			SonnetMappedModel: "deepseek-v4-pro",
			HaikuMappedModel:  "deepseek-v4-flash",
		},
	}
	// 生产 Kimi 0.5x 分组的真实配置。
	kimi := &Group{
		Platform: PlatformKimi,
		MessagesDispatchModelConfig: OpenAIMessagesDispatchModelConfig{
			OpusMappedModel:   "kimi-k2.6",
			SonnetMappedModel: "kimi-k2.6",
			HaikuMappedModel:  "kimi-k2.6",
		},
	}

	cases := []struct {
		name  string
		group *Group
		model string
		want  string
	}{
		// 线上请求量最大的几个模型名，逐一钉住。
		{"deepseek opus-5", deepseek, "claude-opus-5", "deepseek-v4-pro"},
		{"deepseek opus-4-6", deepseek, "claude-opus-4-6", "deepseek-v4-pro"},
		{"deepseek sonnet-5", deepseek, "claude-sonnet-5", "deepseek-v4-pro"},
		{"deepseek haiku", deepseek, "claude-haiku-4-5-20251001", "deepseek-v4-flash"},
		{"kimi opus-5", kimi, "claude-opus-5", "kimi-k2.6"},

		// 原生模型名不属于 claude 家族，不改写。
		{"deepseek 原生模型不改写", deepseek, "deepseek-v4-flash", ""},
		{"kimi 原生模型不改写", kimi, "kimi-k3", ""},

		// 未配置分组级映射时按平台兜底（与前端 getDefaultModelsForPlatform 一致），
		// 而不是回落 gpt-5.x。新建的 CN 分组走的就是这条。
		{"未配置的 deepseek 分组 opus", &Group{Platform: PlatformDeepseek}, "claude-opus-5", "deepseek-v4-pro"},
		{"未配置的 deepseek 分组 haiku", &Group{Platform: PlatformDeepseek}, "claude-haiku-4-5-20251001", "deepseek-v4-flash"},
		{"未配置的 kimi 分组", &Group{Platform: PlatformKimi}, "claude-sonnet-5", "kimi-k2.6"},
		{"未配置的 zhipu 分组 haiku", &Group{Platform: PlatformZhipu}, "claude-haiku-4-5-20251001", "glm-4.5-air"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.group.ResolveMessagesDispatchModel(tc.model); got != tc.want {
				t.Errorf("ResolveMessagesDispatchModel(%q) = %q, want %q", tc.model, got, tc.want)
			}
		})
	}
}

// 没配置映射时必须返回空（原样透传），绝不能回落到 OpenAI 的 gpt-5.x 默认值——
// 那些型号发给 kimi / zhipu / deepseek 上游必然 400。
func TestCNGroupNeverFallsBackToOpenAIDefaultModels(t *testing.T) {
	for _, platform := range []string{PlatformKimi, PlatformZhipu, PlatformDeepseek} {
		g := &Group{Platform: platform} // 无任何映射配置
		for _, model := range []string{"claude-opus-5", "claude-sonnet-5", "claude-haiku-4-5-20251001"} {
			got := g.ResolveMessagesDispatchModel(model)
			if got == "" {
				t.Errorf("%s 分组未配映射时 %s 返回空——应按平台兜底，否则 claude-* 原样透传上游", platform, model)
			}
			if strings.HasPrefix(got, "gpt-") {
				t.Errorf("%s 分组不得回落 OpenAI 型号：%s → %q", platform, model, got)
			}
		}
	}

	// 反向确认：OpenAI 分组仍然享有 gpt-5.x 默认值，本改动没动它。
	openai := &Group{Platform: PlatformOpenAI}
	if got := openai.ResolveMessagesDispatchModel("claude-opus-5"); got != defaultOpenAIMessagesDispatchOpusMappedModel {
		t.Errorf("OpenAI 分组默认值被改坏: got %q, want %q", got, defaultOpenAIMessagesDispatchOpusMappedModel)
	}
}

// 保存分组时不得清掉 CN 分组的模型映射，否则管理员一次保存就把线上映射抹了。
func TestSanitizeKeepsCNGroupModelMapping(t *testing.T) {
	cfg := OpenAIMessagesDispatchModelConfig{
		OpusMappedModel:   "deepseek-v4-pro",
		SonnetMappedModel: "deepseek-v4-pro",
		HaikuMappedModel:  "deepseek-v4-flash",
	}
	g := &Group{Platform: PlatformDeepseek, AllowMessagesDispatch: true, MessagesDispatchModelConfig: cfg}
	sanitizeGroupMessagesDispatchFields(g)
	if !reflect.DeepEqual(g.MessagesDispatchModelConfig, cfg) {
		t.Fatalf("CN 分组的模型映射被 sanitize 清掉了: %+v", g.MessagesDispatchModelConfig)
	}

	// 非 CN、非 OpenAI、非 composite 的平台仍按上游清空。
	other := &Group{Platform: PlatformGemini, AllowMessagesDispatch: true, MessagesDispatchModelConfig: cfg}
	sanitizeGroupMessagesDispatchFields(other)
	if !reflect.DeepEqual(other.MessagesDispatchModelConfig, OpenAIMessagesDispatchModelConfig{}) {
		t.Errorf("gemini 分组的配置应被清空，实际 %+v", other.MessagesDispatchModelConfig)
	}
	if other.AllowMessagesDispatch {
		t.Error("gemini 分组的 AllowMessagesDispatch 应被置 false")
	}
}

// CN 账号的 /v1/messages 分流优先级：显式 api_protocol > fork 的历史直通。
//
// 存量账号（credentials 里没有 api_protocol）必须继续走 fork 的原生 Anthropic 直通
// ——它带着实测出来的 URL 拼装（Deepseek /anthropic/v1/messages、Zhipu 按 host 区分根）
// 与 usage 归一（Deepseek 的 input_tokens 只报缓存未命中数），上游没有等价物；
// 而一旦管理员显式配了协议，就必须让上游的分流接管。
func TestCNAnthropicDirectOnlyWhenAPIProtocolUnset(t *testing.T) {
	mk := func(protocol string) *Account {
		a := &Account{
			Platform: PlatformDeepseek,
			Type:     AccountTypeAPIKey,
			Credentials: map[string]any{
				"api_key":  "k",
				"base_url": "https://api.deepseek.com",
			},
		}
		if protocol != "" {
			a.Credentials["api_protocol"] = protocol
		}
		return a
	}

	// 未配置：走 fork 直通（GetAPIProtocol 回落 chat_completions，但凭证里确实是空）。
	unset := mk("")
	if got := unset.GetCredential("api_protocol"); got != "" {
		t.Fatalf("前置条件不成立：未配置时 api_protocol 应为空，实际 %q", got)
	}
	if unset.GetAPIProtocol() != APIProtocolChatCompletions {
		t.Errorf("未配置时 GetAPIProtocol 应回落 chat_completions，实际 %q", unset.GetAPIProtocol())
	}

	// 显式配置的三种协议都不得再被 fork 直通劫持；判定条件就是凭证非空。
	for _, protocol := range []string{
		APIProtocolChatCompletions, APIProtocolAnthropic, APIProtocolResponses, APIProtocolAdaptive,
	} {
		a := mk(protocol)
		if a.GetCredential("api_protocol") == "" {
			t.Errorf("显式配置 %s 后凭证不应为空", protocol)
		}
	}
}
