package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// 自动透传入口收到 image-only model 的 /responses 请求时，要把顶层 size 补进
// image_generation 工具——上游只认工具里的尺寸，计费也只认「会透传给上游」的尺寸
// 才能按请求档封顶（与 native 入口的计费口径对齐）。
func TestOpenAIGatewayService_Passthrough_ImageOnlyModelNormalizedAndCapped(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 上游忽略请求的 1024x1024，实际出 1254x1254（最长边 >1024，按实际出图档算是 2K）。
	encoded := encodeOpenAIImageTestPNG(t, 1254, 1254)
	upstreamBody := fmt.Sprintf(
		`{"id":"resp_passthrough_image","model":"gpt-image-2","output":[{"id":"ig_1","type":"image_generation_call","size":"1254x1254","result":%q}]}`,
		encoded,
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(nil))
	c.Set("api_key", &APIKey{ID: 7})

	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"x-request-id": []string{"rid-passthrough-image"},
		},
		Body: io.NopCloser(strings.NewReader(upstreamBody)),
	}}
	svc := &OpenAIGatewayService{
		cfg:          &config.Config{},
		httpUpstream: upstream,
	}
	account := &Account{
		ID:             321,
		Name:           "passthrough-apikey",
		Platform:       PlatformOpenAI,
		Type:           AccountTypeAPIKey,
		Concurrency:    1,
		Status:         StatusActive,
		Schedulable:    true,
		RateMultiplier: f64p(1),
		Extra:          map[string]any{"openai_passthrough": true},
		Credentials:    map[string]any{"api_key": "sk-test"},
	}

	body := []byte(`{"model":"gpt-image-2","stream":false,"prompt":"draw a cat","size":"1024x1024"}`)
	result, err := svc.Forward(context.Background(), c, account, body)
	require.NoError(t, err)
	require.NotNil(t, result)

	// 尺寸被补进 image_generation 工具（上游只认这里），其余报文保持客户端原样。
	require.Equal(t, "1024x1024", gjson.GetBytes(upstream.lastBody, "tools.0.size").String())
	require.Equal(t, "gpt-image-2", gjson.GetBytes(upstream.lastBody, "tools.0.model").String())
	require.Equal(t, "gpt-image-2", gjson.GetBytes(upstream.lastBody, "model").String(),
		"透传入口不改写 model，避免牵动下游模型回写与账号冷却记账口径")
	require.Equal(t, "draw a cat", gjson.GetBytes(upstream.lastBody, "prompt").String())

	// 计费按客户端请求的 1K 档封顶，而不是实际出图的 1254x1254（2K）。
	require.Equal(t, 1, result.ImageCount)
	require.Equal(t, "gpt-image-2", result.BillingModel)
	require.Equal(t, "1024x1024", result.ImageInputSize)
	ApplyOpenAIImageBillingResolution(result)
	require.Equal(t, ImageBillingSize1K, result.ImageSize)
	require.Equal(t, ImageSizeSourceCapped, result.ImageSizeSource)

	// 模型身份没被动过：下游看到的、记账用的都还是客户端模型。
	require.Equal(t, "gpt-image-2", result.Model)
	require.Contains(t, rec.Body.String(), "gpt-image-2")
	require.NotContains(t, rec.Body.String(), openAIImagesResponsesMainModel)
}
