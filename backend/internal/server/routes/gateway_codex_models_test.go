package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
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

// Codex manifest 分支的路由门必须与 CodexModels handler 的 openai-only 检查一致：
// OpenAI 兼容的国产平台分组带 client_version 请求 /v1/models 应落回通用模型列表，
// 而不是被 CodexModels 以 404 拒绝。
func TestGatewayRoutesModelsWithClientVersionFallsBackForCompatPlatforms(t *testing.T) {
	for _, platform := range []string{
		service.PlatformKimi,
		service.PlatformZhipu,
		service.PlatformDeepseek,
	} {
		router := newGatewayRoutesTestRouter(platform)

		req := httptest.NewRequest(http.MethodGet, "/v1/models?client_version=0.29.0", nil)
		w := httptest.NewRecorder()

		// 测试路由器的 handler 依赖为零值：真正落到 Gateway.Models 会因 nil 服务
		// panic，这恰好证明分发正确；被 CodexModels 拒绝则会干净地写出 404。
		fellBackToModels := func() (panicked bool) {
			defer func() {
				if recover() != nil {
					panicked = true
				}
			}()
			router.ServeHTTP(w, req)
			return false
		}()

		if !fellBackToModels {
			require.NotEqual(t, http.StatusNotFound, w.Code,
				"platform=%s GET /v1/models?client_version=... should fall back to Gateway.Models, not Codex 404", platform)
		}
	}
}
