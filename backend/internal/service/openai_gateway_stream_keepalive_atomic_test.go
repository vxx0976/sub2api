package service

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// fork：上游 0.2.22 的心跳直通是逐行写（event 行、data 行、空行各一次 Write），
// 首输出超时 / 首输出前的读错误 / 心跳错误帧都可能落在这几次 Write 之间并触发换号，
// 下一次尝试的 response.created 就会拼进未终止的心跳帧，客户端 SSE 解析失败。
// 本 fork 改为整帧缓冲后一次写出，因此被中断的帧必须一字节都不留给客户端。
func TestOpenAINativeStreamDropsUnterminatedKeepaliveFrameOnFailover(t *testing.T) {
	svc := newKeepaliveForwardService(1)
	c, rec := newKeepaliveForwardContext()

	// 只发 event 行就卡住：既没有 data 行也没有空行，帧永远不可能完整。
	_, firstErr := svc.handleStreamingResponse(c.Request.Context(), stalledKeepaliveResponse("event: keepalive\n"),
		c, &Account{ID: 1, Platform: PlatformOpenAI}, time.Now(), "model", "model")

	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, firstErr, &failoverErr, "no semantic output: the first-output timeout must fail over")
	require.Zero(t, rec.Body.Len(), "an unterminated keepalive frame must not reach the client")

	secondResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(strings.Join([]string{
			`data: {"type":"response.created","response":{"id":"resp_second"}}`,
			"",
			`data: {"type":"response.output_text.delta","delta":"hello"}`,
			"",
			`data: {"type":"response.completed","response":{"id":"resp_second","usage":{"input_tokens":1,"output_tokens":1}}}`,
			"",
		}, "\n"))),
	}
	result, secondErr := svc.handleStreamingResponse(c.Request.Context(), secondResp, c,
		&Account{ID: 2, Platform: PlatformOpenAI}, time.Now(), "model", "model")
	require.NoError(t, secondErr)
	require.NotNil(t, result)

	body := rec.Body.String()
	require.True(t, strings.HasPrefix(body, `data: {"type":"response.created"`),
		"the next attempt must start a clean SSE frame, got %q", body)
	require.NotContains(t, body, "keepalive")
}

// fork：不合规中转可能省掉事件之间的空行，此时后续事件会一直落进心跳帧里，
// usage 与终止事件都会丢失。帧类型一旦不是 keepalive，必须先把当前帧收尾再按正常事件处理。
func TestOpenAINativeStreamEndsKeepaliveFrameWhenNextEventIsNotKeepalive(t *testing.T) {
	svc := newKeepaliveForwardService(0)
	c, rec := newKeepaliveForwardContext()

	stream := strings.Join([]string{
		`event: keepalive`,
		`data: {"type":"keepalive","sequence_number":1}`,
		// 没有空行，直接跟下一个事件。
		`event: response.output_text.delta`,
		`data: {"type":"response.output_text.delta","delta":"hi"}`,
		"",
		`data: {"type":"response.completed","response":{"id":"resp_no_blank","usage":{"input_tokens":2,"output_tokens":1}}}`,
		"",
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(stream)),
	}

	result, err := svc.handleStreamingResponse(c.Request.Context(), resp, c,
		&Account{ID: 1, Platform: PlatformOpenAI}, time.Now(), "model", "model")
	require.NoError(t, err)
	require.NotNil(t, result)

	body := rec.Body.String()
	require.Contains(t, body, `"type":"keepalive"`, "the keepalive frame still reaches the client")
	require.Contains(t, body, `"delta":"hi"`, "the following event must not be swallowed as keepalive")
	require.Contains(t, body, `"type":"response.completed"`, "the terminal event must not be swallowed as keepalive")
	require.True(t, strings.HasPrefix(body, "event: keepalive\ndata: {\"type\":\"keepalive\",\"sequence_number\":1}\n\n"),
		"the keepalive must end before the next event, got %q", body)
	require.Contains(t, body, "event: response.output_text.delta\ndata:")
}
