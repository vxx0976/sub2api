package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterTopupRoutes 注册跨通道（EPAY 充值 / AliMPay / USDT）合并充值订单路由。
// 复用与各单通道 /orders 相同的用户鉴权中间件。
func RegisterTopupRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
) {
	auth := v1.Group("/topup")
	auth.Use(gin.HandlerFunc(jwtAuth))
	auth.Use(middleware.BackendModeUserGuard(settingService))
	{
		auth.GET("/orders", h.MergedOrder.ListUserOrders)
	}
}
