package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterRechargeRoutes 注册充值相关路由
func RegisterRechargeRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
) {
	// 公开接口（无需登录）
	rechargePublic := v1.Group("/recharge")
	{
		rechargePublic.GET("/notify", h.Recharge.HandleNotify)
		rechargePublic.POST("/notify", h.Recharge.HandleNotify)
		rechargePublic.GET("/return", h.Recharge.HandleReturn)
		rechargePublic.GET("/config", h.Recharge.GetConfig)
	}

	// 需要登录的接口
	rechargeAuth := v1.Group("/recharge")
	rechargeAuth.Use(gin.HandlerFunc(jwtAuth))
	rechargeAuth.Use(middleware.BackendModeUserGuard(settingService))
	{
		rechargeAuth.POST("/create", h.Recharge.CreateOrder)
		rechargeAuth.GET("/status/:order_no", h.Recharge.GetOrderStatus)
		rechargeAuth.GET("/orders", h.Recharge.ListOrders)
	}
}
