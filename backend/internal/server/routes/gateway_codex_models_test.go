package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGatewayRoutesCodexModelsManifestPathIsRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()

	registered := make(map[string]string)
	for _, route := range router.Routes() {
		if route.Method == http.MethodGet {
			registered[route.Path] = route.Handler
		}
	}

	require.NotEmpty(t, registered["/backend-api/codex/models"], "GET /backend-api/codex/models should be registered")
	require.NotEmpty(t, registered["/v1/models"], "GET /v1/models should be registered")
	require.NotEmpty(t, registered["/models"], "GET /models should be registered")
	require.Equal(t, registered["/v1/models"], registered["/models"], "root alias should use the same platform-aware handler")
}

func TestDispatchCodexModelsGatewayKeepsOnlyOpenAIOnLiveManifestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		platform   string
		wantOpenAI bool
	}{
		{platform: service.PlatformOpenAI, wantOpenAI: true},
		{platform: service.PlatformComposite},
		{platform: service.PlatformGrok},
		{platform: service.PlatformDeepseek},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/models?client_version=0.147.0", nil)
			c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{
				Group: &service.Group{Platform: tt.platform},
			})
			called := ""

			dispatchCodexModelsGateway(c,
				func(c *gin.Context) { called = "openai" },
				func(c *gin.Context) { called = "generated" },
			)

			if tt.wantOpenAI {
				require.Equal(t, "openai", called)
			} else {
				require.Equal(t, "generated", called)
			}
		})
	}
}

// fork 回归：OpenAI 兼容的国产平台（kimi/zhipu/deepseek）带 client_version 请求模型列表时，
// 必须落到生成式 manifest（Gateway.CodexModels），不能进 OpenAIGateway.CodexModels ——
// 后者是 openai-only，会以 404 拒绝，正是 fork 之前用「回落 Gateway.Models」绕开的问题。
// 上游的分发表只钉了 deepseek，这里把 fork 另外两个国产平台一并钉住。
func TestDispatchCodexModelsGatewayUsesGeneratedManifestForCNCompatPlatforms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, platform := range []string{
		service.PlatformKimi,
		service.PlatformZhipu,
		service.PlatformDeepseek,
	} {
		t.Run(platform, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1/models?client_version=0.29.0", nil)
			c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{
				Group: &service.Group{Platform: platform},
			})
			called := ""

			dispatchCodexModelsGateway(c,
				func(c *gin.Context) { called = "openai" },
				func(c *gin.Context) { called = "generated" },
			)

			require.Equal(t, "generated", called,
				"platform=%s must not hit the openai-only live manifest handler", platform)
		})
	}
}
