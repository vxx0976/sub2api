//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type relayBlockedBridgeCase struct {
	name string
	path string
	body []byte
	call func(*GatewayService, context.Context, *gin.Context, *Account, []byte) (*ForwardResult, error)
}

func relayBlockedBridgeCases() []relayBlockedBridgeCase {
	return []relayBlockedBridgeCase{
		{
			name: "chat completions",
			path: "/v1/chat/completions",
			body: []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hello"}]}`),
			call: func(svc *GatewayService, ctx context.Context, c *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
				return svc.ForwardAsChatCompletions(ctx, c, account, body, nil)
			},
		},
		{
			name: "responses",
			path: "/v1/responses",
			body: []byte(`{"model":"claude-sonnet-4-5","input":"hello"}`),
			call: func(svc *GatewayService, ctx context.Context, c *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
				return svc.ForwardAsResponses(ctx, c, account, body, nil)
			},
		},
	}
}

func runRelayBlockedBridge(t *testing.T, tc relayBlockedBridgeCase, status int, upstreamBody string) (*httptest.ResponseRecorder, *gin.Context, *ForwardResult, error) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{{
		StatusCode: status,
		Header:     http.Header{"X-Request-Id": []string{"relay-req"}},
		Body:       io.NopCloser(strings.NewReader(upstreamBody)),
	}}}
	svc := &GatewayService{
		cfg:                 &config.Config{},
		httpUpstream:        upstream,
		tlsFPProfileService: &TLSFingerprintProfileService{},
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, tc.path, nil)
	account := &Account{
		ID:          172,
		Name:        "relay",
		Platform:    PlatformAnthropic,
		Type:        AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "test-key"},
	}
	result, err := tc.call(svc, context.Background(), c, account, tc.body)
	require.Equal(t, 1, upstream.callCount)
	return recorder, c, result, err
}

func relayBlockedOpsEvents(c *gin.Context) []*OpsUpstreamErrorEvent {
	v, _ := c.Get(OpsUpstreamErrorsKey)
	events, _ := v.([]*OpsUpstreamErrorEvent)
	return events
}

// Responses / Chat Completions 桥接到 Anthropic 上游时，中转 422 风控拦截
// 也要与 Forward 同口径换号，而不是映射成 5xx 直接回给客户端。
func TestGatewayBridges_RelayBlocked422FailsOver(t *testing.T) {
	body := `{"error":{"message":"request blocked: client identity could not be fully anonymized","type":"invalid_request_error"},"type":"error"}`
	for _, tc := range relayBlockedBridgeCases() {
		t.Run(tc.name, func(t *testing.T) {
			recorder, c, result, err := runRelayBlockedBridge(t, tc, http.StatusUnprocessableEntity, body)
			require.Nil(t, result)
			var failoverErr *UpstreamFailoverError
			require.ErrorAs(t, err, &failoverErr)
			require.Equal(t, http.StatusUnprocessableEntity, failoverErr.StatusCode)
			require.Equal(t, body, string(failoverErr.ResponseBody))
			require.False(t, failoverErr.RetryableOnSameAccount)
			require.True(t, failoverErr.ShouldRetryNextAccount())
			require.Empty(t, recorder.Body.String())
			require.False(t, c.Writer.Written(), "nothing may be committed before failover")

			events := relayBlockedOpsEvents(c)
			require.Len(t, events, 1)
			require.Equal(t, "failover", events[0].Kind)
			require.Equal(t, int64(172), events[0].AccountID)
			require.Equal(t, http.StatusUnprocessableEntity, events[0].UpstreamStatusCode)
		})
	}
}

// 其他 422 与非 422 状态码维持原处理：不换号、写桥接格式的错误给客户端。
func TestGatewayBridges_OtherErrorsKeepExistingHandling(t *testing.T) {
	statusCases := []struct {
		name   string
		status int
		body   string
	}{
		{"422 普通校验错误", http.StatusUnprocessableEntity, `{"error":{"message":"max_tokens: must be positive","type":"invalid_request_error"}}`},
		{"400 含 request blocked", http.StatusBadRequest, `{"error":{"message":"request blocked","type":"invalid_request_error"}}`},
	}
	for _, tc := range relayBlockedBridgeCases() {
		for _, sc := range statusCases {
			t.Run(tc.name+"/"+sc.name, func(t *testing.T) {
				recorder, c, result, err := runRelayBlockedBridge(t, tc, sc.status, sc.body)
				require.Nil(t, result)
				require.Error(t, err)
				var failoverErr *UpstreamFailoverError
				require.NotErrorAs(t, err, &failoverErr)
				require.Equal(t, mapUpstreamStatusCode(sc.status), recorder.Code)
				require.NotEmpty(t, recorder.Body.String())
				require.Empty(t, relayBlockedOpsEvents(c))
			})
		}
	}
}
