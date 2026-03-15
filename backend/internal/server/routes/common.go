package routes

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// StatusDependencies contains dependencies for status checks
type StatusDependencies struct {
	DB               *sql.DB
	RedisClient      *redis.Client
	GroupService     *service.GroupService
	GroupStatusCache service.GroupStatusCache
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

	// 分组实时状态接口（公开，无需认证）
	r.GET("/api/v1/group-status", func(c *gin.Context) {
		if deps.GroupService == nil || deps.GroupStatusCache == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    1,
				"message": "group status not available",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		groups, err := deps.GroupService.ListActive(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    1,
				"message": "failed to list groups",
			})
			return
		}

		groupIDs := make([]int64, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.ID
		}

		statuses, err := deps.GroupStatusCache.GetGroupStatuses(ctx, groupIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    1,
				"message": "failed to get group statuses",
			})
			return
		}

		type groupStatusItem struct {
			ID            int64   `json:"id"`
			Name          string  `json:"name"`
			Status        string  `json:"status"`
			SuccessRate   float64 `json:"success_rate"`
			TotalRequests int64   `json:"total_requests"`
		}

		items := make([]groupStatusItem, 0, len(groups))
		for _, g := range groups {
			data := statuses[g.ID]
			item := groupStatusItem{
				ID:            g.ID,
				Name:          g.Name,
				TotalRequests: data.Total,
			}
			if data.Total == 0 {
				item.Status = "unknown"
				item.SuccessRate = 0
			} else {
				rate := float64(data.Success) / float64(data.Total) * 100
				item.SuccessRate = rate
				switch {
				case rate >= 99:
					item.Status = "operational"
				case rate >= 95:
					item.Status = "degraded"
				default:
					item.Status = "down"
				}
			}
			items = append(items, item)
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"groups":     items,
				"updated_at": time.Now().UTC().Format(time.RFC3339),
			},
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
