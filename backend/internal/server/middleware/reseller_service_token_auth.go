package middleware

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ResellerServiceTokenHeader 是分销商 M2M 服务令牌的请求头名称。
const ResellerServiceTokenHeader = "X-Reseller-Token"

// ContextKeyServiceTokenID 在 gin context 中存放当前请求所用服务令牌的 ID（用于审计与按令牌限流）。
const ContextKeyServiceTokenID = "service_token_id"

// NewResellerServiceTokenAuthMiddleware 创建分销商服务令牌（M2M）认证中间件。
func NewResellerServiceTokenAuthMiddleware(tokenService *service.ResellerAPITokenService, userService *service.UserService) ResellerServiceTokenAuthMiddleware {
	return ResellerServiceTokenAuthMiddleware(resellerServiceTokenAuth(tokenService, userService))
}

// resellerServiceTokenAuth 分销商服务令牌认证中间件实现。
//
// 鉴权方式：请求头 X-Reseller-Token: rst-xxxx（也兼容 Authorization: Bearer rst-xxxx）。
// 校验通过后写入与 resellerAuth 完全一致的上下文契约（AuthSubject / 角色），
// 因此下游 handler 与 service 的归属校验逻辑无需任何改动。
func resellerServiceTokenAuth(tokenService *service.ResellerAPITokenService, userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractServiceToken(c)
		if raw == "" {
			AbortWithError(c, 401, "UNAUTHORIZED", "Reseller service token is required ("+ResellerServiceTokenHeader+" header)")
			return
		}

		token, err := tokenService.ValidateToken(c.Request.Context(), raw)
		if err != nil {
			AbortWithError(c, 401, "INVALID_TOKEN", "Invalid or revoked service token")
			return
		}

		user, err := userService.GetByID(c.Request.Context(), token.ResellerID)
		if err != nil {
			AbortWithError(c, 401, "USER_NOT_FOUND", "Reseller not found")
			return
		}
		if !user.IsActive() {
			AbortWithError(c, 401, "USER_INACTIVE", "Reseller account is not active")
			return
		}
		if !user.IsReseller() && !user.IsAdmin() {
			AbortWithError(c, 403, "FORBIDDEN", "Reseller access required")
			return
		}

		// 与 resellerAuth 一致的上下文契约：下游据此识别身份并强制 key 归属。
		c.Set(string(ContextKeyUser), AuthSubject{
			UserID:      user.ID,
			Concurrency: user.Concurrency,
		})
		c.Set(string(ContextKeyUserRole), user.Role)
		c.Set("auth_method", "service_token")
		c.Set(ContextKeyServiceTokenID, token.ID)

		// 异步刷新最近使用时间，不阻塞主路径。
		tokenService.TouchLastUsedAsync(token.ID)

		c.Next()
	}
}

// extractServiceToken 优先读取 X-Reseller-Token，其次兼容 Authorization: Bearer rst-xxxx。
func extractServiceToken(c *gin.Context) string {
	if v := strings.TrimSpace(c.GetHeader(ResellerServiceTokenHeader)); v != "" {
		return v
	}
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		candidate := strings.TrimSpace(parts[1])
		if strings.HasPrefix(candidate, service.ResellerServiceTokenPrefix) {
			return candidate
		}
	}
	return ""
}
