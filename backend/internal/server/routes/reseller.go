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

		// 兑换码管理
		registerResellerRedeemRoutes(reseller, h)

		// 用户列表
		registerResellerUserRoutes(reseller, h)

		// 公告管理
		registerResellerAnnouncementRoutes(reseller, h)
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
		settings.DELETE("/tg-bind", h.Reseller.Setting.UnbindTelegram)
	}
}

func registerResellerDomainRoutes(reseller *gin.RouterGroup, h *handler.Handlers) {
	domains := reseller.Group("/domains")
	{
		domains.GET("", h.Reseller.Domain.List)
		domains.GET("/server-info", h.Reseller.Domain.ServerInfo)
		domains.POST("", h.Reseller.Domain.Create)
		domains.PUT("/:id", h.Reseller.Domain.Update)
		domains.DELETE("/:id", h.Reseller.Domain.Delete)
		domains.POST("/:id/verify", h.Reseller.Domain.Verify)
	}
}

func registerResellerRedeemRoutes(reseller *gin.RouterGroup, h *handler.Handlers) {
	redeem := reseller.Group("/redeem")
	{
		redeem.GET("", h.Reseller.Redeem.List)
		redeem.POST("/generate", h.Reseller.Redeem.Generate)
		redeem.DELETE("/:id", h.Reseller.Redeem.Delete)
	}
}

func registerResellerUserRoutes(reseller *gin.RouterGroup, h *handler.Handlers) {
	users := reseller.Group("/users")
	{
		users.GET("", h.Reseller.User.List)
	}
}

func registerResellerAnnouncementRoutes(reseller *gin.RouterGroup, h *handler.Handlers) {
	announcements := reseller.Group("/announcements")
	{
		announcements.GET("", h.Reseller.Announcement.List)
		announcements.POST("", h.Reseller.Announcement.Create)
		announcements.PUT("/:id", h.Reseller.Announcement.Update)
		announcements.DELETE("/:id", h.Reseller.Announcement.Delete)
	}
}
