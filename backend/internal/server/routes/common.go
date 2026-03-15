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

	// 分组实时状态接口（公开，无需认证，30 秒内存缓存）
	respCache := &groupStatusRespCache{ttl: 30 * time.Second}
	r.GET("/api/v1/group-status", func(c *gin.Context) {
		// 命中缓存直接返回
		if cached, ok := respCache.get(); ok {
			c.Data(http.StatusOK, "application/json; charset=utf-8", cached)
			return
		}

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

		const historyDays = 30
		dailyHistory, err := deps.GroupStatusCache.GetDailyHistory(ctx, groupIDs, historyDays)
		if err != nil {
			// 日维度数据获取失败不影响实时状态
			dailyHistory = nil
		}

		type dailyItem struct {
			Date    string  `json:"date"`
			Status  string  `json:"status"`
			Rate    float64 `json:"rate"`
			Total   int64   `json:"total"`
		}

		type groupStatusItem struct {
			ID            int64       `json:"id"`
			Name          string      `json:"name"`
			Status        string      `json:"status"`
			SuccessRate   float64     `json:"success_rate"`
			TotalRequests int64       `json:"total_requests"`
			Uptime30d     float64     `json:"uptime_30d"`
			DailyHistory  []dailyItem `json:"daily_history"`
		}

		items := make([]groupStatusItem, 0, len(groups))
		for _, g := range groups {
			data := statuses[g.ID]
			item := groupStatusItem{
				ID:            g.ID,
				Name:          g.Name,
				TotalRequests: data.Total,
			}

			// 实时状态
			if data.Total == 0 {
				item.Status = "operational" // 无请求时默认正常
				item.SuccessRate = 100
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

			// 30 天日维度历史
			if history, ok := dailyHistory[g.ID]; ok {
				var totalOK, totalAll int64
				item.DailyHistory = make([]dailyItem, len(history))
				for i, d := range history {
					di := dailyItem{Date: d.Date, Total: d.Total}
					if d.Total == 0 {
						di.Status = "operational" // 无数据的天默认正常
						di.Rate = 100
					} else {
						di.Rate = float64(d.Success) / float64(d.Total) * 100
						switch {
						case di.Rate >= 99:
							di.Status = "operational"
						case di.Rate >= 95:
							di.Status = "degraded"
						default:
							di.Status = "down"
						}
					}
					totalOK += d.Success
					totalAll += d.Total
					item.DailyHistory[i] = di
				}
				if totalAll > 0 {
					item.Uptime30d = float64(totalOK) / float64(totalAll) * 100
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
