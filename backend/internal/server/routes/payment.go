package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterPaymentRoutes 注册支付相关路由
func RegisterPaymentRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
) {
	// 认证接口
	authenticated := v1.Group("")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	{
		// 获取可购买套餐列表
		authenticated.GET("/plans", h.Payment.GetPlans)
		// 创建订单
		authenticated.POST("/orders", h.Payment.CreateOrder)
		// 获取用户订单列表
		authenticated.GET("/orders", h.Payment.GetUserOrders)
		// 继续支付（重新获取支付二维码）
		authenticated.POST("/orders/:order_no/repay", h.Payment.RepayOrder)
		// 查询订单支付状态（前端轮询）
		authenticated.GET("/orders/:order_no/payment-status", h.Payment.GetOrderPaymentStatus)
	}
}
