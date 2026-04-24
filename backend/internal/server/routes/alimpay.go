package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterAliMPayRoutes 注册 AliMPay 个人免签充值路由
// 和 /api/v1/recharge/* 并列，没有 webhook（走轮询匹配账单）
func RegisterAliMPayRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
) {
	// 公开接口（获取配置无需登录）
	public := v1.Group("/alimpay")
	{
		public.GET("/config", h.Order.GetConfig)
	}

	// 需要登录的接口
	auth := v1.Group("/alimpay")
	auth.Use(gin.HandlerFunc(jwtAuth))
	auth.Use(middleware.BackendModeUserGuard(settingService))
	{
		auth.POST("/create", h.Order.CreateOrder)
		auth.GET("/status/:order_no", h.Order.GetOrderStatus)
		auth.GET("/orders", h.Order.ListOrders)
	}
}
