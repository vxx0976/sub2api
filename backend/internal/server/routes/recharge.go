package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRechargeRoutes 注册充值相关路由
func RegisterRechargeRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	adminAuth middleware.AdminAuthMiddleware,
) {
	// 认证接口
	authenticated := v1.Group("")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	{
		// 用户充值接口
		authenticated.POST("/recharge/orders", h.Recharge.CreateRechargeOrder)
		authenticated.GET("/recharge/orders", h.Recharge.GetRechargeOrders)
		authenticated.POST("/recharge/orders/:order_no/repay", h.Recharge.RepayRechargeOrder)
		authenticated.GET("/recharge/config", h.Recharge.GetRechargeConfig)
	}

	// 管理员接口
	admin := v1.Group("/admin")
	admin.Use(gin.HandlerFunc(adminAuth))
	{
		admin.GET("/recharge/orders", h.Recharge.GetAllRechargeOrders)
		admin.GET("/settings/recharge", h.Recharge.GetRechargeConfigAdmin)
		admin.PUT("/settings/recharge", h.Recharge.UpdateRechargeConfig)
	}
}
