package routes

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	ratelimit "github.com/Wei-Shaw/sub2api/internal/middleware"
	servermw "github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterResellerServiceTokenAPIRoutes 注册分销商 M2M 接口（/api/v1/reseller-api/*）。
//
// 鉴权：X-Reseller-Token 服务令牌（非 JWT 登录），仅分销商后端可调用。
// 这些端点复用分销商已有的 key handler/service，归属校验在 service 层强制：
// 一个令牌只能操作 user_id == reseller_id 的 key。
//
// 限流：按服务令牌维度 60 次/分钟，Redis 故障时放行（fail-open），
// 避免偶发抖动阻塞"客户已付款 → 重置配额"这类关键操作。
func RegisterResellerServiceTokenAPIRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	serviceTokenAuth servermw.ResellerServiceTokenAuthMiddleware,
	redisClient *redis.Client,
) {
	limiter := ratelimit.NewRateLimiter(redisClient)
	// 鉴权前的按 IP 粗限流：拦住无效令牌的洪水（每次都会触发一次 token_hash 查库），
	// 在它们到达鉴权/DB 之前就削峰。
	preAuthByIP := limiter.LimitWithOptions("reseller-api-ip", 120, time.Minute, ratelimit.RateLimitOptions{
		FailureMode: ratelimit.RateLimitFailOpen,
	})
	// 鉴权后的按令牌限流：正常业务调用的主要约束。
	perToken := limiter.LimitWithOptions("reseller-api", 60, time.Minute, ratelimit.RateLimitOptions{
		FailureMode: ratelimit.RateLimitFailOpen,
		KeyFunc: func(c *gin.Context) string {
			if id := c.GetInt64(servermw.ContextKeyServiceTokenID); id > 0 {
				return "stoken:" + strconv.FormatInt(id, 10)
			}
			return ""
		},
	})

	api := v1.Group("/reseller-api")
	api.Use(preAuthByIP)
	api.Use(gin.HandlerFunc(serviceTokenAuth))
	api.Use(perToken)
	{
		keys := api.Group("/keys")
		{
			keys.GET("", h.Reseller.Key.List)
			keys.POST("", h.Reseller.Key.Create)
			keys.POST("/:id/reset-quota", h.Reseller.Key.ResetQuota)
			keys.POST("/:id/enable", h.Reseller.Key.Enable)
			keys.POST("/:id/disable", h.Reseller.Key.Disable)
		}
	}
}
