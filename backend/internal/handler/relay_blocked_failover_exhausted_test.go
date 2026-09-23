//go:build unit

package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// fork: 中转 422 风控拦截在 Responses / Chat Completions 桥接上换号耗尽时，
// 与 /v1/messages 一致回 502 上游错误：不回 422+server_error，拦截原因只进 ops。
func TestBridgeFailoverExhaustedRelayBlockedReturns502(t *testing.T) {
	gin.SetMode(gin.TestMode)
	blocked := &service.UpstreamFailoverError{
		StatusCode:   http.StatusUnprocessableEntity,
		ResponseBody: []byte(`{"error":{"message":"request blocked: client identity could not be fully anonymized","type":"invalid_request_error"},"type":"error"}`),
	}
	handlers := map[string]func(*GatewayHandler, *gin.Context){
		"responses":        func(h *GatewayHandler, c *gin.Context) { h.handleResponsesFailoverExhausted(c, blocked, false) },
		"chat_completions": func(h *GatewayHandler, c *gin.Context) { h.handleCCFailoverExhausted(c, blocked, false) },
	}
	for name, run := range handlers {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			run(&GatewayHandler{}, c)

			require.Equal(t, http.StatusBadGateway, recorder.Code)
			require.Equal(t, "Upstream request failed", gjson.Get(recorder.Body.String(), "error.message").String())
			require.NotContains(t, recorder.Body.String(), "request blocked")
			recorded, ok := c.Get(service.OpsUpstreamErrorMessageKey)
			require.True(t, ok)
			require.Contains(t, recorded, "request blocked")
		})
	}
}

// 其他 422 仍按原逻辑回原状态码，不被误判为中转风控。
func TestBridgeFailoverExhaustedOther422Unchanged(t *testing.T) {
	gin.SetMode(gin.TestMode)
	other := &service.UpstreamFailoverError{
		StatusCode:   http.StatusUnprocessableEntity,
		ResponseBody: []byte(`{"error":{"message":"messages: field required","type":"invalid_request_error"}}`),
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	(&GatewayHandler{}).handleCCFailoverExhausted(c, other, false)
	require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	recorder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recorder)
	(&GatewayHandler{}).handleResponsesFailoverExhausted(c, other, false)
	require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}
