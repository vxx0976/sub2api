package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// 生产回归（2026-09-16，GPT Pro 号池）：Codex 流的真实降载序列是
// created → in_progress → output_item.added(reasoning, 只带 encrypted_content)
// → response.failed(server_is_overloaded)。
// 修复前 reasoning item 会把流提交为 200，随后的 failed 只能带内透传，
// failover 彻底失效（ops_error_logs 里 attempts=1、kind=stream_failed、零 usage 行）。
// 期望：不向客户端写出任何字节，并抛出 pre-output failover 错误。
func TestOpenAIStreamOpaqueReasoningThenCapacityShedStillFailsOver(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stream := strings.Join([]string{
		"event: response.created",
		`data: {"type":"response.created","response":{"id":"resp_1"},"sequence_number":0}`,
		"",
		"event: response.in_progress",
		`data: {"type":"response.in_progress","response":{"id":"resp_1"},"sequence_number":1}`,
		"",
		"event: response.output_item.added",
		`data: {"type":"response.output_item.added","output_index":0,"item":{"id":"rs_1","type":"reasoning","summary":[],"encrypted_content":"gAAAAABo_opaque"},"sequence_number":2}`,
		"",
		"event: response.failed",
		`data: {"type":"response.failed","response":{"id":"resp_1","status":"failed","error":{"code":"server_is_overloaded","message":"Our servers are currently overloaded. Please try again later."}},"sequence_number":3}`,
		"",
	}, "\n")

	tests := []struct {
		name         string
		stagesHeader bool
		run          func(*OpenAIGatewayService, *gin.Context, *http.Response, *Account) error
	}{
		{
			name:         "native",
			stagesHeader: true,
			run: func(svc *OpenAIGatewayService, c *gin.Context, resp *http.Response, account *Account) error {
				_, err := svc.handleStreamingResponse(c.Request.Context(), resp, c, account, time.Now(), "model", "model")
				return err
			},
		},
		{
			// 透传路径用 Header().Set 直接回写 x-request-id（覆盖而非追加），
			// 不走 attempt 暂存头，因此不做 header 断言。
			name: "passthrough",
			run: func(svc *OpenAIGatewayService, c *gin.Context, resp *http.Response, account *Account) error {
				_, err := svc.handleStreamingResponsePassthrough(c.Request.Context(), resp, c, account, time.Now(), "model", "model")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{MaxLineSize: defaultMaxLineSize}}}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(stream)),
				Header:     http.Header{"X-Request-Id": []string{"rid-opaque-reasoning-shed"}},
			}
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Name: "acc"}

			err := tt.run(svc, c, resp, account)
			require.Error(t, err)
			var failoverErr *UpstreamFailoverError
			require.ErrorAs(t, err, &failoverErr)
			require.True(t, failoverErr.RetryableOnSameAccount)
			require.True(t, failoverErr.RequestScopedTransient)
			require.Equal(t, http.StatusServiceUnavailable, failoverErr.ClientStatusCode)
			require.False(t, c.Writer.Written(), "不得提交响应头，否则 failover 无法重放")
			require.Empty(t, rec.Body.String())
			if tt.stagesHeader {
				require.Empty(t, c.Writer.Header().Values("X-Request-Id"), "未提交的 attempt 不应把自己的 x-request-id 叠加给下一次尝试")
			}
		})
	}
}

// 正常流：不透明 reasoning item 只是被暂存，首个可见输出到达时必须按原顺序
// 一并送出，客户端不能丢事件。
func TestOpenAIStreamOpaqueReasoningIsReplayedOnFirstVisibleOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stream := strings.Join([]string{
		"event: response.created",
		`data: {"type":"response.created","response":{"id":"resp_1"},"sequence_number":0}`,
		"",
		"event: response.output_item.added",
		`data: {"type":"response.output_item.added","output_index":0,"item":{"id":"rs_1","type":"reasoning","summary":[],"encrypted_content":"gAAAAABo_opaque"},"sequence_number":1}`,
		"",
		"event: response.output_text.delta",
		`data: {"type":"response.output_text.delta","delta":"hi","sequence_number":2}`,
		"",
		"event: response.completed",
		`data: {"type":"response.completed","response":{"id":"resp_1","status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"hi"}]}],"usage":{"input_tokens":5,"output_tokens":1}},"sequence_number":3}`,
		"",
	}, "\n")

	tests := []struct {
		name string
		run  func(*OpenAIGatewayService, *gin.Context, *http.Response, *Account) error
	}{
		{
			name: "native",
			run: func(svc *OpenAIGatewayService, c *gin.Context, resp *http.Response, account *Account) error {
				_, err := svc.handleStreamingResponse(c.Request.Context(), resp, c, account, time.Now(), "model", "model")
				return err
			},
		},
		{
			name: "passthrough",
			run: func(svc *OpenAIGatewayService, c *gin.Context, resp *http.Response, account *Account) error {
				_, err := svc.handleStreamingResponsePassthrough(c.Request.Context(), resp, c, account, time.Now(), "model", "model")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{MaxLineSize: defaultMaxLineSize}}}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(stream)),
				Header:     http.Header{"X-Request-Id": []string{"rid-opaque-reasoning-ok"}},
			}
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Name: "acc"}

			require.NoError(t, tt.run(svc, c, resp, account))
			body := rec.Body.String()
			require.Contains(t, body, `"encrypted_content":"gAAAAABo_opaque"`, "暂存的 reasoning item 必须补发")
			require.Contains(t, body, `"type":"response.output_text.delta"`)
			require.Less(t, strings.Index(body, `"encrypted_content"`), strings.Index(body, `response.output_text.delta`), "补发必须保持原始顺序")
		})
	}
}
