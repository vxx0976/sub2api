//go:build unit

package service

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func newRelayBlockedTestResp(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// 中转（如 zhima）对单条请求的 422 风控拦截应切换到下一个账号，
// 而不是直接包成 502 返回给用户。
func TestRelayRequestBlockedFailover_422BlockedSwitchesAccount(t *testing.T) {
	s := &GatewayService{}
	c := newTransportErrorTestGin(t)
	account := &Account{ID: 172, Name: "relay", Platform: PlatformAnthropic, Type: AccountTypeAPIKey}
	body := `{"error":{"message":"request blocked: client identity could not be fully anonymized","type":"invalid_request_error"},"type":"error"}`
	resp := newRelayBlockedTestResp(http.StatusUnprocessableEntity, body)

	failoverErr := s.relayRequestBlockedFailover(c, resp, account, false)
	if failoverErr == nil {
		t.Fatal("expected failover error for relay-blocked 422")
	}
	if failoverErr.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want 422", failoverErr.StatusCode)
	}
	if failoverErr.RetryableOnSameAccount {
		t.Fatal("relay block must not retry on the same account")
	}
	if !failoverErr.ShouldRetryNextAccount() {
		t.Fatal("relay block must switch to the next account")
	}
	if string(failoverErr.ResponseBody) != body {
		t.Fatalf("ResponseBody = %q, want original body", failoverErr.ResponseBody)
	}

	v, _ := c.Get(OpsUpstreamErrorsKey)
	events, _ := v.([]*OpsUpstreamErrorEvent)
	if len(events) != 1 || events[0].Kind != "failover" || events[0].AccountID != 172 {
		t.Fatalf("unexpected ops events: %+v", events)
	}
}

// 其他 422（如普通参数校验错误）和非 422 状态码保持原有处理：
// 不切换账号，且响应体要放回去供常规错误处理读取。
func TestRelayRequestBlockedFailover_OtherErrorsUnchanged(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"422 普通校验错误", http.StatusUnprocessableEntity, `{"error":{"message":"max_tokens: must be positive","type":"invalid_request_error"}}`},
		{"400 含 request blocked", http.StatusBadRequest, `{"error":{"message":"request blocked","type":"invalid_request_error"}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &GatewayService{}
			c := newTransportErrorTestGin(t)
			account := &Account{ID: 1, Name: "acc", Platform: PlatformAnthropic}
			resp := newRelayBlockedTestResp(tc.status, tc.body)

			if failoverErr := s.relayRequestBlockedFailover(c, resp, account, false); failoverErr != nil {
				t.Fatalf("unexpected failover: %+v", failoverErr)
			}
			got, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if string(got) != tc.body {
				t.Fatalf("body after check = %q, want %q", got, tc.body)
			}
			if v, ok := c.Get(OpsUpstreamErrorsKey); ok {
				if events, _ := v.([]*OpsUpstreamErrorEvent); len(events) != 0 {
					t.Fatalf("unexpected ops events: %+v", events)
				}
			}
		})
	}
}

// 中转的 Claude Code 客户端版本门槛（400）在 /v1/messages 路径上换号：
// hanhan 要求 opus-5-5 用 2.1.280+、Max-hh 拒 2.1.161，而同组其他中转可接。
func TestRelayMessagesRejectedFailover_ClientVersionGateSwitchesAccount(t *testing.T) {
	bodies := []string{
		`{"error":{"message":"Your Claude Code version (2.1.161) is below the minimum required version (2.1.220). Please update: npm update -g @anthropic-ai/claude-code","type":"invalid_request_error"},"type":"error"}`,
		`{"error":{"message":"Claude Code 2.1.274 does not support this model; version 2.1.280 or newer is required. Run 'claude update', or update the Claude desktop app, then try again. (request id: 202609290202258726400428268d9d6ldiD6WX0)","type":"invalid_request_error"},"type":"error"}`,
	}
	for _, body := range bodies {
		s := &GatewayService{}
		c := newTransportErrorTestGin(t)
		account := &Account{ID: 232, Name: "relay", Platform: PlatformAnthropic, Type: AccountTypeAPIKey}
		resp := newRelayBlockedTestResp(http.StatusBadRequest, body)

		failoverErr := s.relayMessagesRejectedFailover(c, resp, account, false)
		if failoverErr == nil {
			t.Fatalf("expected failover for version gate: %s", body)
		}
		if failoverErr.StatusCode != http.StatusBadRequest || failoverErr.RetryableOnSameAccount || !failoverErr.ShouldRetryNextAccount() {
			t.Fatalf("unexpected failover error: %+v", failoverErr)
		}
		if string(failoverErr.ResponseBody) != body {
			t.Fatalf("ResponseBody = %q, want original body", failoverErr.ResponseBody)
		}
	}
}

// 版本门槛只在 /v1/messages 启用；桥接端点（relayRequestBlockedFailover）与普通 400 维持原处理。
func TestRelayMessagesRejectedFailover_OtherErrorsUnchanged(t *testing.T) {
	versionGate := `{"error":{"message":"Your Claude Code version (2.1.161) is below the minimum required version (2.1.220).","type":"invalid_request_error"}}`
	cases := []struct {
		name     string
		status   int
		body     string
		messages bool
	}{
		{"桥接端点不对版本门槛换号", http.StatusBadRequest, versionGate, false},
		{"普通 400 不换号", http.StatusBadRequest, `{"error":{"message":"max_tokens: Field required","type":"invalid_request_error"}}`, true},
		{"不带 Claude Code 字样的版本提示不换号", http.StatusBadRequest, `{"error":{"message":"SDK version is below the minimum required version","type":"invalid_request_error"}}`, true},
		{"非 400 的版本提示不换号", http.StatusForbidden, versionGate, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &GatewayService{}
			c := newTransportErrorTestGin(t)
			account := &Account{ID: 1, Name: "acc", Platform: PlatformAnthropic}
			resp := newRelayBlockedTestResp(tc.status, tc.body)

			var failoverErr *UpstreamFailoverError
			if tc.messages {
				failoverErr = s.relayMessagesRejectedFailover(c, resp, account, false)
			} else {
				failoverErr = s.relayRequestBlockedFailover(c, resp, account, false)
			}
			if failoverErr != nil {
				t.Fatalf("unexpected failover: %+v", failoverErr)
			}
			got, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if string(got) != tc.body {
				t.Fatalf("body after check = %q, want %q", got, tc.body)
			}
		})
	}
}

// 官方 OAuth 号的版本拒绝源自网关自身的 CLI 版本伪装，换号无济于事，不应换号。
func TestRelayMessagesRejectedFailover_OAuthVersionGateUnchanged(t *testing.T) {
	body := `{"error":{"message":"Claude Code 2.1.274 does not support this model; version 2.1.280 or newer is required.","type":"invalid_request_error"}}`
	s := &GatewayService{}
	c := newTransportErrorTestGin(t)
	account := &Account{ID: 9, Name: "oauth", Platform: PlatformAnthropic, Type: AccountTypeOAuth}
	resp := newRelayBlockedTestResp(http.StatusBadRequest, body)
	if failoverErr := s.relayMessagesRejectedFailover(c, resp, account, false); failoverErr != nil {
		t.Fatalf("unexpected failover for OAuth account: %+v", failoverErr)
	}
	if got, _ := io.ReadAll(resp.Body); string(got) != body {
		t.Fatalf("body after check = %q, want %q", got, body)
	}
}
