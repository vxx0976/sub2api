package routes

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// StatusDependencies contains dependencies for status checks
type StatusDependencies struct {
	DB          *sql.DB
	RedisClient *redis.Client
}

// RegisterCommonRoutes 注册通用路由（健康检查、状态等）
func RegisterCommonRoutes(r *gin.Engine, deps *StatusDependencies) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Claude Code 遥测日志（忽略，直接返回200）
	r.POST("/api/event_logging/batch", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup status endpoint (always returns needs_setup: false in normal mode)
	// This is used by the frontend to detect when the service has restarted after setup
	r.GET("/setup/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"needs_setup": false,
				"step":        "completed",
			},
		})
	})

	// System status endpoint for public status page
	r.GET("/api/v1/status", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		response := gin.H{
			"api": gin.H{
				"status":  "healthy",
				"latency": 0,
			},
			"database": checkDatabaseStatus(ctx, deps.DB),
			"cache":    checkRedisStatus(ctx, deps.RedisClient),
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": response,
		})
	})
}

// checkDatabaseStatus checks the database connection status
func checkDatabaseStatus(ctx context.Context, db *sql.DB) gin.H {
	if db == nil {
		return gin.H{
			"status": "unknown",
		}
	}

	start := time.Now()
	err := db.PingContext(ctx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return gin.H{
			"status":  "down",
			"latency": latency,
		}
	}

	return gin.H{
		"status":  "healthy",
		"latency": latency,
	}
}

// checkRedisStatus checks the Redis connection status
func checkRedisStatus(ctx context.Context, client *redis.Client) gin.H {
	if client == nil {
		return gin.H{
			"status": "unknown",
		}
	}

	start := time.Now()
	_, err := client.Ping(ctx).Result()
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return gin.H{
			"status":  "down",
			"latency": latency,
		}
	}

	return gin.H{
		"status":  "healthy",
		"latency": latency,
	}
}
