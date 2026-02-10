package server

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/server/routes"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/web"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// SetupRouter 配置路由器中间件和路由
func SetupRouter(
	r *gin.Engine,
	handlers *handler.Handlers,
	jwtAuth middleware2.JWTAuthMiddleware,
	adminAuth middleware2.AdminAuthMiddleware,
	apiKeyAuth middleware2.APIKeyAuthMiddleware,
	resellerAuth middleware2.ResellerAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	domainRepo service.ResellerDomainRepository,
	settingRepo service.ResellerSettingRepository,
	cfg *config.Config,
	redisClient *redis.Client,
	db *sql.DB,
) *gin.Engine {
	// 应用中间件
	r.Use(middleware2.Logger())
	r.Use(middleware2.CORS(cfg.CORS))
	r.Use(middleware2.SecurityHeaders(cfg.Security.CSP))

	// Domain detection middleware — must run before frontend server
	// so reseller branding can be injected into the HTML
	if domainRepo != nil {
		r.Use(middleware2.DomainDetect(domainRepo, settingRepo))
	}

	// Serve embedded frontend with settings injection if available
	if web.HasEmbeddedFrontend() {
		frontendServer, err := web.NewFrontendServer(settingService)
		if err != nil {
			log.Printf("Warning: Failed to create frontend server with settings injection: %v, using legacy mode", err)
			r.Use(web.ServeEmbeddedFrontend())
		} else {
			// Register cache invalidation callbacks
			settingService.SetOnUpdateCallback(frontendServer.InvalidateCache)
			handlers.Reseller.Domain.SetCacheInvalidator(frontendServer.InvalidateDomainCache)
			r.Use(frontendServer.Middleware())
		}
	}

	// 注册路由
	registerRoutes(r, handlers, jwtAuth, adminAuth, apiKeyAuth, resellerAuth, apiKeyService, subscriptionService, opsService, domainRepo, settingRepo, cfg, redisClient, db)

	return r
}

// registerRoutes 注册所有 HTTP 路由
func registerRoutes(
	r *gin.Engine,
	h *handler.Handlers,
	jwtAuth middleware2.JWTAuthMiddleware,
	adminAuth middleware2.AdminAuthMiddleware,
	apiKeyAuth middleware2.APIKeyAuthMiddleware,
	resellerAuth middleware2.ResellerAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	domainRepo service.ResellerDomainRepository,
	settingRepo service.ResellerSettingRepository,
	cfg *config.Config,
	redisClient *redis.Client,
	db *sql.DB,
) {
	// 通用路由（健康检查、状态等）
	routes.RegisterCommonRoutes(r, &routes.StatusDependencies{
		DB:          db,
		RedisClient: redisClient,
	})

	// API v1
	v1 := r.Group("/api/v1")

	// Caddy on-demand TLS ask endpoint (public, no auth)
	if domainRepo != nil {
		v1.GET("/caddy/ask-tls", caddyAskTLS(domainRepo))
	}

	// 注册各模块路由
	routes.RegisterAuthRoutes(v1, h, jwtAuth, redisClient)
	routes.RegisterUserRoutes(v1, h, jwtAuth)
	routes.RegisterAdminRoutes(v1, h, adminAuth)
	routes.RegisterResellerRoutes(v1, h, resellerAuth)
	routes.RegisterPaymentRoutes(v1, h, jwtAuth)
	routes.RegisterRechargeRoutes(v1, h, jwtAuth, adminAuth)
	routes.RegisterGatewayRoutes(r, h, apiKeyAuth, apiKeyService, subscriptionService, opsService, cfg)
}

// caddyAskTLS returns a handler for Caddy's on-demand TLS ask endpoint.
// Caddy calls GET /api/v1/caddy/ask-tls?domain=<domain> before issuing a certificate.
// Returns 200 if the domain is a verified reseller domain, 404 otherwise.
func caddyAskTLS(domainRepo service.ResellerDomainRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		domain := c.Query("domain")
		if domain == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "domain parameter required"})
			return
		}

		d, err := domainRepo.GetByDomain(c.Request.Context(), domain)
		if err != nil || d == nil || !d.Verified {
			c.JSON(http.StatusNotFound, gin.H{"error": "domain not found or not verified"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"domain": domain, "status": "verified"})
	}
}
