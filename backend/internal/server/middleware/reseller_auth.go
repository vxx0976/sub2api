package middleware

import (
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// NewResellerAuthMiddleware 创建分销商认证中间件
func NewResellerAuthMiddleware(authService *service.AuthService, userService *service.UserService) ResellerAuthMiddleware {
	return ResellerAuthMiddleware(resellerAuth(authService, userService))
}

// resellerAuth 分销商认证中间件实现
// JWT Token: Authorization: Bearer <jwt-token> (需要分销商或管理员角色)
func resellerAuth(authService *service.AuthService, userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			AbortWithError(c, 401, "UNAUTHORIZED", "Authorization header is required")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			AbortWithError(c, 401, "INVALID_AUTH_HEADER", "Authorization header format must be 'Bearer {token}'")
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			AbortWithError(c, 401, "EMPTY_TOKEN", "Token cannot be empty")
			return
		}

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, service.ErrTokenExpired) {
				AbortWithError(c, 401, "TOKEN_EXPIRED", "Token has expired")
				return
			}
			AbortWithError(c, 401, "INVALID_TOKEN", "Invalid token")
			return
		}

		user, err := userService.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			AbortWithError(c, 401, "USER_NOT_FOUND", "User not found")
			return
		}

		if !user.IsActive() {
			AbortWithError(c, 401, "USER_INACTIVE", "User account is not active")
			return
		}

		// 检查分销商或管理员权限（管理员可以访问分销商接口）
		if !user.IsReseller() && !user.IsAdmin() {
			AbortWithError(c, 403, "FORBIDDEN", "Reseller access required")
			return
		}

		// Security: Validate TokenVersion
		if claims.TokenVersion != user.TokenVersion {
			AbortWithError(c, 401, "TOKEN_REVOKED", "Token has been revoked (password changed)")
			return
		}

		// Security: Validate RoleVersion
		if claims.RoleVersion != user.RoleVersion {
			AbortWithError(c, 401, "TOKEN_REVOKED", "Token has been revoked (role changed)")
			return
		}

		c.Set(string(ContextKeyUser), AuthSubject{
			UserID:      user.ID,
			Concurrency: user.Concurrency,
		})
		c.Set(string(ContextKeyUserRole), user.Role)
		c.Set("auth_method", "jwt")

		c.Next()
	}
}
