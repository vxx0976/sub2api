package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterResellerRoutes 注册分销商路由
func RegisterResellerRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	resellerAuth middleware.ResellerAuthMiddleware,
) {
	reseller := v1.Group("/reseller")
	reseller.Use(gin.HandlerFunc(resellerAuth))
	{
		// 仪表盘
		reseller.GET("/dashboard", h.Reseller.Dashboard.GetStats)

		// 套餐（分组）管理
		registerResellerGroupRoutes(reseller, h)

		// API Key 管理
		registerResellerKeyRoutes(reseller, h)

		// 域名管理
		registerResellerDomainRoutes(reseller, h)

		// 设置管理
		registerResellerSettingRoutes(reseller, h)
	}
}

func registerResellerGroupRoutes(reseller *gin.RouterGroup, h *handler.Handlers) {
	groups := reseller.Group("/groups")
	{
		groups.GET("", h.Reseller.Group.List)
		groups.GET("/templates", h.Reseller.Group.ListTemplates)
		groups.GET("/:id", h.Reseller.Group.GetByID)
		groups.POST("", h.Reseller.Group.Create)
		groups.PUT("/:id", h.Reseller.Group.Update)
		groups.DELETE("/:id", h.Reseller.Group.Delete)
	}
}

func registerResellerKeyRoutes(reseller *gin.RouterGroup, h *handler.Handlers) {
	keys := reseller.Group("/keys")
	{
		keys.GET("", h.Reseller.Key.List)
		keys.POST("", h.Reseller.Key.Create)
		keys.PUT("/:id", h.Reseller.Key.Update)
		keys.DELETE("/:id", h.Reseller.Key.Delete)
		keys.POST("/:id/reset-quota", h.Reseller.Key.ResetQuota)
	}
}

func registerResellerSettingRoutes(reseller *gin.RouterGroup, h *handler.Handlers) {
	settings := reseller.Group("/settings")
	{
		settings.GET("", h.Reseller.Setting.Get)
		settings.PUT("", h.Reseller.Setting.Update)
		settings.POST("/tg-bind-code", h.Reseller.Setting.GenerateBindCode)
	}
}

func registerResellerDomainRoutes(reseller *gin.RouterGroup, h *handler.Handlers) {
	domains := reseller.Group("/domains")
	{
		domains.GET("", h.Reseller.Domain.List)
		domains.POST("", h.Reseller.Domain.Create)
		domains.PUT("/:id", h.Reseller.Domain.Update)
		domains.DELETE("/:id", h.Reseller.Domain.Delete)
		domains.POST("/:id/verify", h.Reseller.Domain.Verify)
	}
}
