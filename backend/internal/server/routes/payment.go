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
	// 公开接口（无需认证）
	// 支付回调通知（异步回调，服务端对服务端）
	v1.GET("/payment/notify", h.Payment.PaymentNotify)
	// 支付完成跳转（同步回调）
	v1.GET("/payment/return", h.Payment.PaymentReturn)
	// 支付结果验证（前端调用，验证签名并处理订单）
	v1.POST("/payment/verify", h.Payment.VerifyPaymentReturn)

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
		// 继续支付（重新获取支付链接）
		authenticated.POST("/orders/:order_no/repay", h.Payment.RepayOrder)
	}
}
