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
