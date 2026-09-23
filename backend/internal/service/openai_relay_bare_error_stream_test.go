package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// fork 回归：上游 277aa1411 让流在 terminal 事件后提前结束，只豁免了 OAuth 的
// bare error 序列。API-key 中继（GPT 号池主力）同样会在输出开始后发
// error + response.failed 成对事件；若在 bare error 处提前结束，response.failed
// 永远读不到，流内失败就不再记 stream_failed，监控里静默消失。
func TestOpenAIRelayBareErrorMidStreamStillReadsResponseFailed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{MaxLineSize: defaultMaxLineSize}}}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"X-Request-Id": []string{"rid-relay-bare-error"}},
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			`data: {"type":"response.created","response":{"id":"resp_relay"}}`,
			"",
			`data: {"type":"response.output_text.delta","delta":"hello"}`,
			"",
			"event: error",
			`data: {"type":"error","error":{"type":"service_unavailable_error","code":"server_is_overloaded","message":"Our servers are currently overloaded. Please try again later."}}`,
			"",
			"event: response.failed",
			`data: {"type":"response.failed","response":{"id":"resp_relay","status":"failed","error":{"code":"server_is_overloaded","message":"Our servers are currently overloaded."},"usage":{"input_tokens":11,"output_tokens":7}}}`,
			"",
			"",
		}, "\n"))),
	}
	account := &Account{ID: 7, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Name: "relay"}

	result, err := svc.handleStreamingResponse(c.Request.Context(), resp, c, account, time.Now(), "model", "model")
	require.Error(t, err)
	var failoverErr *UpstreamFailoverError
	require.NotErrorAs(t, err, &failoverErr, "输出已开始，不能再 failover")
	require.NotNil(t, result)
	require.NotNil(t, result.usage)
	require.Equal(t, 11, result.usage.InputTokens, "response.failed 必须被读到")
	require.Equal(t, 7, result.usage.OutputTokens)

	raw, _ := c.Get(OpsUpstreamErrorsKey)
	events, _ := raw.([]*OpsUpstreamErrorEvent)
	streamFailed := 0
	for _, ev := range events {
		if ev.Kind == "stream_failed" {
			streamFailed++
		}
	}
	require.Equal(t, 1, streamFailed, "流内失败必须记一条 stream_failed: %+v", events)
}
