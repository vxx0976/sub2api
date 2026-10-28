//go:build unit

package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func runPreambleStream(t *testing.T, write func(pw *io.PipeWriter)) (*httptest.ResponseRecorder, *gin.Context, *streamingResult, error) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc := newMinimalGatewayService()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	ResetStreamPreamble(c)

	pr, pw := io.Pipe()
	resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: pr}
	go func() {
		defer func() { _ = pw.Close() }()
		write(pw)
	}()
	result, err := svc.handleStreamingResponse(context.Background(), resp, c, &Account{ID: 1}, time.Now(), "model", "model", false)
	_ = pr.Close()
	return rec, c, result, err
}

// 中转排队等并发：先回 200 持续发 ping（中间夹未计量的 message_start），最后流内报并发超限。
// 客户端只收到 ping，不算提交，handler 可换号。
func TestHandleStreamingResponse_PreambleOnlyErrorAllowsFailover(t *testing.T) {
	rec, c, result, err := runPreambleStream(t, func(pw *io.PipeWriter) {
		_, _ = pw.Write([]byte("event: ping\ndata: {\"type\":\"ping\"}\n\n"))
		_, _ = pw.Write([]byte(`event: message_start` + "\n" + `data: {"type":"message_start","message":{"usage":{"input_tokens":0}}}` + "\n\n"))
		_, _ = pw.Write([]byte("event: ping\ndata: {\"type\":\"ping\"}\n\n"))
		_, _ = pw.Write([]byte("event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"rate_limit_error\",\"message\":\"Concurrency limit exceeded for account, please retry later\"}}\n\n"))
	})

	var sseErr *sseStreamErrorEventError
	require.True(t, errors.As(err, &sseErr))
	require.Contains(t, rec.Body.String(), "ping")
	require.NotContains(t, rec.Body.String(), "message_start")
	require.True(t, c.Writer.Written())
	require.False(t, StreamOutputCommitted(c, -1), "只写过 ping 不应视为已提交")
	require.NotNil(t, result)
	require.False(t, shouldBillPartialStream(result))
}

// 首个内容事件到达即提交：暂存的 message_start 按原顺序先写出，usage 正常合并。
func TestHandleStreamingResponse_PreambleCommitsOnContentInOrder(t *testing.T) {
	rec, c, result, err := runPreambleStream(t, func(pw *io.PipeWriter) {
		_, _ = pw.Write([]byte("event: ping\ndata: {\"type\":\"ping\"}\n\n"))
		_, _ = pw.Write([]byte(`event: message_start` + "\n" + `data: {"type":"message_start","message":{"usage":{"input_tokens":0}}}` + "\n\n"))
		_, _ = pw.Write([]byte(`event: content_block_start` + "\n" + `data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}` + "\n\n"))
		_, _ = pw.Write([]byte(`event: content_block_delta` + "\n" + `data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hi"}}` + "\n\n"))
		_, _ = pw.Write([]byte(`event: message_delta` + "\n" + `data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":5,"cache_read_input_tokens":100,"output_tokens":7}}` + "\n\n"))
		_, _ = pw.Write([]byte(`event: message_stop` + "\n" + `data: {"type":"message_stop"}` + "\n\n"))
	})

	require.NoError(t, err)
	body := rec.Body.String()
	iPing := strings.Index(body, "ping")
	iStart := strings.Index(body, "message_start")
	iBlock := strings.Index(body, "content_block_start")
	require.True(t, iPing >= 0 && iStart > iPing && iBlock > iStart, "顺序应为 ping → message_start → content_block_start: %s", body)
	require.True(t, StreamOutputCommitted(c, -1))
	require.Equal(t, 5, result.usage.InputTokens)
	require.Equal(t, 100, result.usage.CacheReadInputTokens)
	require.Equal(t, 7, result.usage.OutputTokens)
}

// 内容已写出后再报错：已提交，不得换号，按部分交付计费。
func TestHandleStreamingResponse_ErrorAfterContentIsCommitted(t *testing.T) {
	rec, c, result, err := runPreambleStream(t, func(pw *io.PipeWriter) {
		_, _ = pw.Write([]byte(`event: message_start` + "\n" + `data: {"type":"message_start","message":{"usage":{"input_tokens":5}}}` + "\n\n"))
		_, _ = pw.Write([]byte(`event: content_block_start` + "\n" + `data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}` + "\n\n"))
		_, _ = pw.Write([]byte("event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"overloaded_error\",\"message\":\"Overloaded\"}}\n\n"))
	})

	var sseErr *sseStreamErrorEventError
	require.True(t, errors.As(err, &sseErr))
	require.Contains(t, rec.Body.String(), "message_start")
	require.True(t, StreamOutputCommitted(c, -1))
	require.True(t, shouldBillPartialStream(result))
}

// 携带 token 的 message_start 立即提交（上游已开始计量），不暂存。
func TestHandleStreamingResponse_MeteredMessageStartCommitsImmediately(t *testing.T) {
	rec, c, _, err := runPreambleStream(t, func(pw *io.PipeWriter) {
		_, _ = pw.Write([]byte(`event: message_start` + "\n" + `data: {"type":"message_start","message":{"usage":{"input_tokens":5}}}` + "\n\n"))
		_, _ = pw.Write([]byte("event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"rate_limit_error\",\"message\":\"x\"}}\n\n"))
	})
	var sseErr *sseStreamErrorEventError
	require.True(t, errors.As(err, &sseErr))
	require.Contains(t, rec.Body.String(), "message_start")
	require.True(t, StreamOutputCommitted(c, -1))
}

// 只发过 ping 就断流（读错误）：仍按未输出处理，返回可换号的 failover 错误。
func TestHandleStreamingResponse_ReadErrorAfterPingFailsOver(t *testing.T) {
	rec, c, _, err := runPreambleStream(t, func(pw *io.PipeWriter) {
		_, _ = pw.Write([]byte("event: ping\ndata: {\"type\":\"ping\"}\n\n"))
		_ = pw.CloseWithError(io.ErrUnexpectedEOF)
	})

	var failoverErr *UpstreamFailoverError
	require.True(t, errors.As(err, &failoverErr), "期望 failover 错误, got %v", err)
	require.Contains(t, rec.Body.String(), "ping")
	require.False(t, StreamOutputCommitted(c, -1))
}

func TestStreamOutputCommitted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	require.False(t, StreamOutputCommitted(c, c.Writer.Size()), "未写出任何内容")

	_, _ = c.Writer.WriteString("data: {}\n\n")
	require.True(t, StreamOutputCommitted(c, -1), "未经流式前导处理的写出一律视为已提交")

	before := c.Writer.Size()
	setStreamPreambleOnly(c, true)
	_, _ = c.Writer.WriteString("event: ping\n\n")
	require.False(t, StreamOutputCommitted(c, before))

	ResetStreamPreamble(c)
	require.True(t, StreamOutputCommitted(c, before), "Reset 后不应残留上一次尝试的前导标记")
}

func TestRelayQueueRejectStreamCooldown(t *testing.T) {
	require.Equal(t, relayQueueRejectCooldown, relayQueueRejectStreamCooldown([]byte(`{"type":"error","error":{"type":"rate_limit_error","message":"Concurrency limit exceeded for account, please retry later"}}`)))
	require.Equal(t, relayQueueRejectCooldown, relayQueueRejectStreamCooldown([]byte(`{"type":"error","error":{"type":"api_error","message":"Concurrency limit exceeded"}}`)))
	// 非并发类限流（RPM、级联上游 429）：只短冷却。
	require.Equal(t, relayRateLimitRejectCooldown, relayQueueRejectStreamCooldown([]byte(`{"type":"error","error":{"type":"rate_limit_error","message":"Upstream rate limit exceeded, please retry later"}}`)))
	// 可能只是单个模型无容量：仅换号，不冷却整个账号。
	require.Zero(t, relayQueueRejectStreamCooldown([]byte(`{"type":"error","error":{"type":"api_error","message":"No available accounts: no available accounts"}}`)))
	require.Zero(t, relayQueueRejectStreamCooldown([]byte(`{"type":"error","error":{"type":"api_error","message":"Internal server error"}}`)))
	require.Zero(t, relayQueueRejectStreamCooldown([]byte(`not json`)))
}

type relayQueueRejectRepoStub struct {
	internal500AccountRepoStub
	peers      []Account
	groupPeers map[int64][]Account
}

func (r *relayQueueRejectRepoStub) ListSchedulableByGroupIDAndPlatform(_ context.Context, groupID int64, _ string) ([]Account, error) {
	if peers, ok := r.groupPeers[groupID]; ok {
		return peers, nil
	}
	return r.peers, nil
}

func relayQueueRejectContext(groupID int64) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("api_key", &APIKey{ID: 1, GroupID: &groupID})
	return c
}

func TestTempUnscheduleRelayQueueReject(t *testing.T) {
	reject := []byte(`{"type":"error","error":{"type":"rate_limit_error","message":"Concurrency limit exceeded for account"}}`)
	relay := &Account{ID: 232, Name: "Max-hh", Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer := Account{ID: 172, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}

	repo := &relayQueueRejectRepoStub{peers: []Account{*relay, peer}}
	svc := &GatewayService{accountRepo: repo}
	svc.tempUnscheduleRelayQueueReject(context.Background(), relayQueueRejectContext(86), relay, reject, "claude-opus-5-5")
	require.Len(t, repo.tempUnschedCalls, 1)
	require.Equal(t, int64(232), repo.tempUnschedCalls[0].accountID)
	require.WithinDuration(t, time.Now().Add(relayQueueRejectCooldown), repo.tempUnschedCalls[0].until, 5*time.Second)

	// 账号同时在另一个只有它自己的分组：冷却作用于所有分组，不冷却。
	multi := *relay
	multi.GroupIDs = []int64{86, 168}
	multiRepo := &relayQueueRejectRepoStub{peers: []Account{*relay, peer}, groupPeers: map[int64][]Account{168: {*relay}}}
	(&GatewayService{accountRepo: multiRepo}).tempUnscheduleRelayQueueReject(context.Background(), relayQueueRejectContext(86), &multi, reject, "claude-opus-5-5")
	require.Empty(t, multiRepo.tempUnschedCalls)

	// 其他账号不支持该模型：不算替补，不冷却。
	narrow := peer
	narrow.Credentials = map[string]any{"model_mapping": map[string]any{"claude-sonnet-5": "claude-sonnet-5"}}
	narrowRepo := &relayQueueRejectRepoStub{peers: []Account{*relay, narrow}}
	(&GatewayService{accountRepo: narrowRepo}).tempUnscheduleRelayQueueReject(context.Background(), relayQueueRejectContext(86), relay, reject, "claude-opus-5-5")
	require.Empty(t, narrowRepo.tempUnschedCalls)

	// 分组内没有其他可调度账号：不冷却，避免整组不可用。
	solo := &relayQueueRejectRepoStub{peers: []Account{*relay}}
	(&GatewayService{accountRepo: solo}).tempUnscheduleRelayQueueReject(context.Background(), relayQueueRejectContext(168), relay, reject, "claude-opus-5-5")
	require.Empty(t, solo.tempUnschedCalls)

	// 已在冷却中、OAuth 账号、非排队类错误、无分组上下文：均不冷却。
	repo2 := &relayQueueRejectRepoStub{peers: []Account{*relay, peer}}
	svc2 := &GatewayService{accountRepo: repo2}
	future := time.Now().Add(time.Minute)
	cooling := *relay
	cooling.TempUnschedulableUntil = &future
	svc2.tempUnscheduleRelayQueueReject(context.Background(), relayQueueRejectContext(86), &cooling, reject, "claude-opus-5-5")
	svc2.tempUnscheduleRelayQueueReject(context.Background(), relayQueueRejectContext(86), &Account{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeOAuth}, reject, "claude-opus-5-5")
	svc2.tempUnscheduleRelayQueueReject(context.Background(), relayQueueRejectContext(86), relay, []byte(`{"type":"error","error":{"type":"api_error","message":"boom"}}`), "claude-opus-5-5")
	gin.SetMode(gin.TestMode)
	bare, _ := gin.CreateTestContext(httptest.NewRecorder())
	svc2.tempUnscheduleRelayQueueReject(context.Background(), bare, relay, reject, "claude-opus-5-5")
	require.Empty(t, repo2.tempUnschedCalls)
}

func TestWriteClaudeErrorBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 未开始流：照常写 JSON 与状态码。
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	writeClaudeErrorBody(c, http.StatusBadRequest, []byte(`{"type":"error","error":{"type":"invalid_request_error","message":"prompt is too long"}}`))
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	require.True(t, strings.HasPrefix(rec.Body.String(), `{"type":"error"`))

	// 已以 200 SSE 发过 ping：改写为 SSE error 帧，保留上游错误类型与文案（客户端靠它触发自动压缩）。
	rec = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(rec)
	c.Header("Content-Type", "text/event-stream")
	_, _ = c.Writer.WriteString("event: ping\ndata: {\"type\":\"ping\"}\n\n")
	writeClaudeErrorBody(c, http.StatusBadRequest, []byte("{\n  \"type\": \"error\",\n  \"error\": {\"type\":\"invalid_request_error\",\"message\":\"prompt is too long\"}\n}"))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"invalid_request_error\",\"message\":\"prompt is too long\"}}\n\n")

	// 非 Anthropic 形状的错误体：包装成标准 error 帧。
	rec = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(rec)
	c.Header("Content-Type", "text/event-stream")
	_, _ = c.Writer.WriteString("event: ping\ndata: {}\n\n")
	writeClaudeErrorBody(c, http.StatusBadRequest, []byte(`upstream exploded`))
	body := rec.Body.String()
	require.Contains(t, body, "event: error\ndata: ")
	require.Contains(t, body, `"type":"api_error"`)
}
