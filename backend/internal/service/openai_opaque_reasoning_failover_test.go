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

// 对抗 review 补齐的缺口：暂存窗口 × 下游心跳。
// 心跳负载 ": keepalive\n\n" 自带空行，一旦插进补发事件的 event: 行与 data: 行
// 之间就会把事件块提前截断。断言暂存事件补发后 SSE 结构仍完整。
func TestOpenAIPassthroughStagedReplayNotCorruptedByKeepalive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reader, writer := io.Pipe()
	go func() {
		defer func() { _ = writer.Close() }()
		_, _ = io.WriteString(writer, "event: response.created\ndata: {\"type\":\"response.created\",\"response\":{\"id\":\"r1\"}}\n\n")
		_, _ = io.WriteString(writer, "event: response.output_item.added\ndata: {\"type\":\"response.output_item.added\",\"item\":{\"id\":\"rs_1\",\"type\":\"reasoning\",\"summary\":[],\"encrypted_content\":\"gAAAA\"}}\n\n")
		// 思考阶段跨过多个心跳周期
		time.Sleep(2500 * time.Millisecond)
		_, _ = io.WriteString(writer, "event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"hi\"}\n\n")
		_, _ = io.WriteString(writer, "event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"r1\",\"status\":\"completed\",\"output\":[{\"type\":\"message\",\"content\":[{\"type\":\"output_text\",\"text\":\"hi\"}]}],\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\n")
	}()

	svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{
		MaxLineSize:             defaultMaxLineSize,
		StreamKeepaliveInterval: 1,
	}}}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: reader}
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Name: "acc"}

	_, err := svc.handleStreamingResponsePassthrough(c.Request.Context(), resp, c, account, time.Now(), "model", "model")
	require.NoError(t, err)

	body := rec.Body.String()
	// 心跳应该出现过（证明暂存期间连接确实被保活）
	require.Contains(t, body, ":", "暂存期间应有心跳写出")
	// 每个 event: 行后面紧跟的必须是 data: 行，中间不得被心跳截断
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "event: ") {
			continue
		}
		require.Less(t, i+1, len(lines), "event 行后必须还有内容: %q", line)
		require.True(t, strings.HasPrefix(lines[i+1], "data: "),
			"event 行与 data 行之间被插入了内容: %q -> %q", line, lines[i+1])
	}
	require.Contains(t, body, `"encrypted_content":"gAAAA"`)
	require.Contains(t, body, `"type":"response.completed"`)
}

// 对抗 review 发现：passthrough 的 pendingLines 原本无上限，只用 commits=false
// 的事件（只带 encrypted_content 的 reasoning item、text 为空的 content_part.added
// 等）就能把网关内存打爆——native 侧的 openAIFirstOutputStage 有 8MB 上限并会溢写
// 磁盘，透传侧是裸内存切片。
//
// 判据：上游只发 commits=false 的事件、且【不】以终止事件收尾。修复前全部攒在
// 内存里、EOF 时走无输出 failover，下游一个字节都没有；修复后超过上限即提交，
// 下游必然见到字节。
func TestOpenAIPassthroughStagingIsBounded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	blob := strings.Repeat("x", 256*1024)
	reader, writer := io.Pipe()
	go func() {
		defer func() { _ = writer.Close() }()
		_, _ = io.WriteString(writer, "event: response.created\ndata: {\"type\":\"response.created\",\"response\":{\"id\":\"r1\"}}\n\n")
		for i := 0; i < 40; i++ {
			_, _ = io.WriteString(writer, "event: response.output_item.added\ndata: {\"type\":\"response.output_item.added\",\"item\":{\"id\":\"rs\",\"type\":\"reasoning\",\"summary\":[],\"encrypted_content\":\""+blob+"\"}}\n\n")
		}
		// 刻意不发终止事件
	}()

	svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{MaxLineSize: defaultMaxLineSize}}}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: reader}
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Name: "acc"}

	_, _ = svc.handleStreamingResponsePassthrough(c.Request.Context(), resp, c, account, time.Now(), "model", "model")

	require.Greater(t, rec.Body.Len(), 0,
		"暂存量超过 openAIFirstOutputStageMaxBytes 后必须提交下发，否则 pendingLines 无上限增长")
	require.GreaterOrEqual(t, rec.Body.Len(), int(openAIFirstOutputStageMaxBytes)/2,
		"提交后应把已暂存的事件补发出去")
}
