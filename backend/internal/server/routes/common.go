package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
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

// groupStatusCache 内存缓存，避免公开端点频繁查询 Redis
type groupStatusRespCache struct {
	mu       sync.RWMutex
	data     []byte
	cachedAt time.Time
	ttl      time.Duration
}

func (c *groupStatusRespCache) get() ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.data == nil || time.Since(c.cachedAt) > c.ttl {
		return nil, false
	}
	return c.data, true
}

func (c *groupStatusRespCache) set(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = data
	c.cachedAt = time.Now()
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

	// 分组实时状态接口（公开，无需认证）。
	// 数据源已从“真实调用流量计数器”切换为“定时健康探测历史”：
	//   - 当前状态读 groups.health_status（available → operational，否则 down）；
	//   - 30 天柱状图按 UTC 日聚合 group_health_check_logs。
	// 探测频率较低（默认 30 分钟/组），缓存 TTL 同步放宽到 120s。
	respCache := &groupStatusRespCache{ttl: 120 * time.Second}
	r.GET("/api/v1/group-status", func(c *gin.Context) {
		// 命中缓存直接返回
		if cached, ok := respCache.get(); ok {
			c.Data(http.StatusOK, "application/json; charset=utf-8", cached)
			return
		}

		if deps.GroupService == nil {
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

		const historyDays = 30
		dailyHistory, err := deps.GroupService.ListHealthDailyHistory(ctx, groupIDs, historyDays)
		if err != nil {
			// 历史聚合失败不影响当前状态徽章渲染。
			dailyHistory = nil
		}

		type dailyItem struct {
			Date   string  `json:"date"`
			Status string  `json:"status"`
			Rate   float64 `json:"rate"`
			Total  int64   `json:"total"`
		}

		type groupStatusItem struct {
			ID            int64       `json:"id"`
			Name          string      `json:"name"`
			Platform      string      `json:"platform"`
			Description   string      `json:"description,omitempty"`
			Status        string      `json:"status"`
			SuccessRate   float64     `json:"success_rate"`
			TotalRequests int64       `json:"total_requests"`
			AvgLatency    int64       `json:"avg_latency_ms"`
			Uptime30d     float64     `json:"uptime_30d"`
			DailyHistory  []dailyItem `json:"daily_history"`
		}

		items := make([]groupStatusItem, 0, len(groups))
		for _, g := range groups {
			item := groupStatusItem{
				ID:          g.ID,
				Name:        g.Name,
				Platform:    g.Platform,
				Description: g.Description,
			}

			// 当前状态：直接看健康探测最新结果。空字符串视为“尚未探测”，按正常展示。
			switch g.HealthStatus {
			case service.HealthStatusUnavailable:
				item.Status = "down"
				item.SuccessRate = 0
			default:
				item.Status = "operational"
				item.SuccessRate = 100
			}

			// 30 天柱状图：每天有任何成功探测即视为可用（用户口径：只要绿就好）。
			if history, ok := dailyHistory[g.ID]; ok && len(history) > 0 {
				var sumSuccess, sumTotal int64
				item.DailyHistory = make([]dailyItem, len(history))
				for i, d := range history {
					di := dailyItem{Date: d.Date, Total: d.Total}
					switch {
					case d.Total == 0:
						// 无探测数据的天默认正常占位。
						di.Status = "operational"
						di.Rate = 100
					case d.Success > 0:
						di.Status = "operational"
						di.Rate = float64(d.Success) / float64(d.Total) * 100
					default:
						di.Status = "down"
						di.Rate = 0
					}
					sumSuccess += d.Success
					sumTotal += d.Total
					item.DailyHistory[i] = di
				}
				if sumTotal > 0 {
					item.Uptime30d = float64(sumSuccess) / float64(sumTotal) * 100
				} else {
					item.Uptime30d = 100
				}
			} else {
				item.Uptime30d = 100
				item.DailyHistory = []dailyItem{}
			}

			items = append(items, item)
		}

		resp := gin.H{
			"code": 0,
			"data": gin.H{
				"groups":     items,
				"updated_at": time.Now().UTC().Format(time.RFC3339),
			},
		}
		// 序列化并缓存
		if b, err := json.Marshal(resp); err == nil {
			respCache.set(b)
			c.Data(http.StatusOK, "application/json; charset=utf-8", b)
		} else {
			c.JSON(http.StatusOK, resp)
		}
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
