package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterUsdtRoutes 注册 USDT(TRC20) 自建收款充值路由。
// 和 /api/v1/alimpay/* 并列，无 webhook（走链上轮询匹配）。
func RegisterUsdtRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
) {
	// 公开接口（获取配置无需登录）
	public := v1.Group("/usdt")
	{
		public.GET("/config", h.Usdt.GetConfig)
	}

	// 需要登录的接口
	auth := v1.Group("/usdt")
	auth.Use(gin.HandlerFunc(jwtAuth))
	auth.Use(middleware.BackendModeUserGuard(settingService))
	{
		auth.POST("/create", h.Usdt.CreateOrder)
		auth.GET("/status/:order_no", h.Usdt.GetOrderStatus)
		auth.GET("/orders", h.Usdt.ListOrders)
	}
}
