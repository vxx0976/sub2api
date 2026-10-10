//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// 2026-10-09：pigcode 对单条请求回 403「内容审计命中风险规则」，网关按账号鉴权失败
// SetError，整个 Max-pig 停调度数小时。中转号的内容审核 403 只换号、不处罚账号。
func TestHandleUpstreamError_RelayContentModeration403SkipsAccountPenalty(t *testing.T) {
	bodies := []string{
		`{"error":{"message":"内容审计命中风险规则，请调整输入后重试","type":"forbidden"},"type":"error"}`,
		`{"error":{"message":"Request rejected by content moderation, please rephrase and retry","type":"forbidden"}}`,
	}
	for _, platform := range []string{PlatformAnthropic, PlatformOpenAI, PlatformGrok} {
		for _, body := range bodies {
			repo := &rateLimitAccountRepoStub{}
			blocker := &runtimeBlockRecorder{}
			svc := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
			svc.SetOpenAI403CounterCache(&openAI403CounterCacheStub{})
			svc.SetAccountRuntimeBlocker(blocker)
			account := &Account{ID: 281, Platform: platform, Type: AccountTypeAPIKey}

			shouldDisable := svc.HandleUpstreamError(context.Background(), account, http.StatusForbidden, http.Header{}, []byte(body))

			require.False(t, shouldDisable, "不得发出停调度信号（OpenAI/Grok 侧会转成运行时阻断）: %s %s", platform, body)
			require.Equal(t, 0, repo.setErrorCalls, "内容审核 403 不得永久禁用中转号: %s %s", platform, body)
			require.Equal(t, 0, repo.tempCalls, "内容审核 403 不得临时停调度中转号: %s %s", platform, body)
			require.Empty(t, blocker.accounts, "内容审核 403 不得触发运行时调度阻断: %s %s", platform, body)
		}
	}
}

// 作用域守卫：账号级 403（含审核字样的封号通知）、普通 403、官方 OAuth 号维持原有停号行为。
func TestHandleUpstreamError_NonModerationOrOAuth403Unchanged(t *testing.T) {
	relay := func() *Account { return &Account{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeAPIKey} }
	cases := []struct {
		name    string
		account *Account
		body    string
	}{
		{"中转号普通 403", relay(), `{"error":{"message":"invalid api key, access denied"}}`},
		{"风险规则封号", relay(), `{"error":{"message":"账号触发风险规则已封禁"}}`},
		{"令牌命中风险规则被禁用", relay(), `{"error":{"message":"您的令牌命中风险规则已被禁用，请调整输入后重试"}}`},
		{"审核违规封号(英文)", relay(), `{"error":{"message":"Your account has been suspended due to content moderation violations"}}`},
		{"只有审核词无请求级提示", relay(), `{"error":{"message":"内容审计未通过"}}`},
		{"OAuth 号即使带审核字样", &Account{ID: 2, Platform: PlatformAnthropic, Type: AccountTypeOAuth}, `{"error":{"message":"内容审计命中风险规则，请调整输入后重试"}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &rateLimitAccountRepoStub{}
			svc := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
			svc.HandleUpstreamError(context.Background(), tc.account, http.StatusForbidden, http.Header{}, []byte(tc.body))
			require.Equal(t, 1, repo.setErrorCalls)
		})
	}
}
