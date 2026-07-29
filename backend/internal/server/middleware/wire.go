package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// JWTAuthMiddleware JWT 认证中间件类型
type JWTAuthMiddleware gin.HandlerFunc

// OptionalJWTAuthMiddleware 可选 JWT 认证中间件类型：匿名放行，带 token 严格校验
type OptionalJWTAuthMiddleware gin.HandlerFunc

// AdminAuthMiddleware 管理员认证中间件类型
type AdminAuthMiddleware gin.HandlerFunc

// APIKeyAuthMiddleware API Key 认证中间件类型
type APIKeyAuthMiddleware gin.HandlerFunc

// ResellerAuthMiddleware 分销商认证中间件类型
type ResellerAuthMiddleware gin.HandlerFunc

// ResellerServiceTokenAuthMiddleware 分销商服务令牌（M2M）认证中间件类型
type ResellerServiceTokenAuthMiddleware gin.HandlerFunc

// ProviderSet 中间件层的依赖注入
var ProviderSet = wire.NewSet(
	NewJWTAuthMiddleware,
	NewOptionalJWTAuthMiddleware,
	NewAdminAuthMiddleware,
	NewAPIKeyAuthMiddleware,
	NewResellerAuthMiddleware,
	NewResellerServiceTokenAuthMiddleware,
	NewAuditLogMiddleware,
	NewStepUpAuthMiddleware,
)
