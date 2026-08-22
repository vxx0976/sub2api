package service

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

const (
	HealthStatusAvailable   = "available"
	HealthStatusUnavailable = "unavailable"
	groupHealthMaxWorkers   = 5 // max concurrent group checks
	// 健康检查历史日志保留天数：状态页最长展示 30 天，多留 1 天兼容 UTC 边界。
	groupHealthLogRetentionDays = 31
)

type GroupHealthCheckService struct {
	groupRepo      GroupRepository
	accountRepo    AccountRepository
	accountTestSvc *AccountTestService
	rateLimitSvc   *RateLimitService
	cfg            *config.Config

	cron      *cron.Cron
	startOnce sync.Once
	stopOnce  sync.Once

	lockCache  LeaderLockCache
	db         *sql.DB
	instanceID string
}

func NewGroupHealthCheckService(
	groupRepo GroupRepository,
	accountRepo AccountRepository,
	accountTestSvc *AccountTestService,
	rateLimitSvc *RateLimitService,
	cfg *config.Config,
) *GroupHealthCheckService {
	return &GroupHealthCheckService{
		groupRepo:      groupRepo,
		accountRepo:    accountRepo,
		accountTestSvc: accountTestSvc,
		rateLimitSvc:   rateLimitSvc,
		cfg:            cfg,
		instanceID:     uuid.NewString(),
	}
}

// SetLeaderLock injects the leader-lock cache and DB used to elect a single
// instance for the periodic health checks. When both are nil they run ungated.
func (s *GroupHealthCheckService) SetLeaderLock(lockCache LeaderLockCache, db *sql.DB) {
	if s == nil {
		return
	}
	s.lockCache = lockCache
	s.db = db
}

func (s *GroupHealthCheckService) Start() {
	if s == nil {
		return
	}
	s.startOnce.Do(func() {
		loc := time.Local
		if s.cfg != nil {
			if parsed, err := time.LoadLocation(s.cfg.Timezone); err == nil && parsed != nil {
				loc = parsed
			}
		}

		c := cron.New(cron.WithParser(cron.NewParser(
			cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow,
		)), cron.WithLocation(loc))

		_, err := c.AddFunc("*/5 * * * *", func() { s.runHealthCheck() })
		if err != nil {
			logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] not started (invalid schedule): %v", err)
			return
		}
		// 每天 03:17（按配置时区）清理过期历史日志，错峰避开整点流量高峰。
		if _, err := c.AddFunc("17 3 * * *", func() { s.runLogRetention() }); err != nil {
			logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] log retention not scheduled: %v", err)
		}
		s.cron = c
		s.cron.Start()
		logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] started (tick=every 5 minutes, per-group interval)")
	})
}

func (s *GroupHealthCheckService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.cron != nil {
			ctx := s.cron.Stop()
			select {
			case <-ctx.Done():
			case <-time.After(3 * time.Second):
			}
		}
	})
}

func (s *GroupHealthCheckService) runHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	release, ok := tryAcquireSingletonLeaderLock(ctx, s.lockCache, s.db, "leader:group_health_check", s.instanceID, 15*time.Minute)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}

	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] ListActive error: %v", err)
		return
	}
	if len(groups) == 0 {
		return
	}

	// 根据每个分组的检查间隔过滤需要检查的分组
	now := time.Now()
	var dueGroups []*Group
	for i := range groups {
		g := &groups[i]
		interval := g.HealthCheckIntervalMin
		if interval <= 0 {
			interval = 30 // 默认 30 分钟
		}
		if g.LastHealthCheckAt != nil && now.Sub(*g.LastHealthCheckAt) < time.Duration(interval)*time.Minute {
			continue // 未到检查时间，跳过
		}
		dueGroups = append(dueGroups, g)
	}

	if len(dueGroups) == 0 {
		return
	}

	logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] checking %d/%d active groups", len(dueGroups), len(groups))

	sem := make(chan struct{}, groupHealthMaxWorkers)
	var wg sync.WaitGroup

	for _, g := range dueGroups {
		sem <- struct{}{}
		wg.Add(1)
		go func(g *Group) {
			defer wg.Done()
			defer func() { <-sem }()
			s.checkOneGroup(ctx, g)
		}(g)
	}

	wg.Wait()
}

// CheckGroupByID 手动触发单个分组的健康检查
func (s *GroupHealthCheckService) CheckGroupByID(ctx context.Context, groupID int64) error {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	s.checkOneGroup(ctx, group)
	return nil
}

func (s *GroupHealthCheckService) checkOneGroup(ctx context.Context, group *Group) {
	// Get schedulable accounts for this group
	accounts, err := s.accountRepo.ListSchedulableByGroupID(ctx, group.ID)
	if err != nil {
		slog.Error("GroupHealthCheck: list accounts", "group_id", group.ID, "error", err)
		return
	}

	if len(accounts) == 0 {
		// No accounts = unavailable
		now := time.Now()
		_ = s.groupRepo.UpdateHealthStatus(ctx, group.ID, HealthStatusUnavailable, 0, 0, now)
		if err := s.groupRepo.InsertHealthCheckLog(ctx, group.ID, HealthStatusUnavailable, now); err != nil {
			slog.Error("GroupHealthCheck: insert log", "group_id", group.ID, "error", err)
		}
		return
	}

	// Determine test model: group-level override wins over platform default
	modelID := strings.TrimSpace(group.HealthCheckTestModel)
	if modelID == "" {
		modelID = getDefaultTestModel(group.Platform)
	}

	// 逐个测试所有账号，第一个成功即绿灯并停止
	healthy := 0
	tested := 0
	for _, acc := range accounts {
		tested++
		// 单个账号测试超时 60 秒，避免长时间挂起
		testCtx, testCancel := context.WithTimeout(ctx, 60*time.Second)
		result, err := s.accountTestSvc.RunTestBackground(testCtx, acc.ID, modelID)
		testCancel()
		if err != nil {
			continue
		}
		if result.Status == "success" {
			healthy++
			// Auto-recover on success
			if s.rateLimitSvc != nil {
				_, _ = s.rateLimitSvc.RecoverAccountAfterSuccessfulTest(ctx, acc.ID)
			}
			break // 有一个可用即达标，跳过剩余账号
		}
	}

	status := HealthStatusUnavailable
	if healthy > 0 {
		status = HealthStatusAvailable
	}

	now := time.Now()
	if err := s.groupRepo.UpdateHealthStatus(ctx, group.ID, status, healthy, tested, now); err != nil {
		slog.Error("GroupHealthCheck: update status", "group_id", group.ID, "error", err)
	}
	if err := s.groupRepo.InsertHealthCheckLog(ctx, group.ID, status, now); err != nil {
		slog.Error("GroupHealthCheck: insert log", "group_id", group.ID, "error", err)
	}
}

// runLogRetention 在 leader 节点删除 retention 之外的探测历史，避免明细无限增长。
func (s *GroupHealthCheckService) runLogRetention() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	release, ok := tryAcquireSingletonLeaderLock(ctx, s.lockCache, s.db, "leader:group_health_check_retention", s.instanceID, 15*time.Minute)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}

	cutoff := time.Now().AddDate(0, 0, -groupHealthLogRetentionDays)
	deleted, err := s.groupRepo.DeleteHealthCheckLogsBefore(ctx, cutoff)
	if err != nil {
		logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] log retention failed: %v", err)
		return
	}
	if deleted > 0 {
		logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] log retention pruned %d rows (cutoff=%s)", deleted, cutoff.Format(time.RFC3339))
	}
}

// getDefaultTestModel returns a lightweight model for health checks per platform.
func getDefaultTestModel(platform string) string {
	switch platform {
	case "anthropic":
		return "claude-haiku-4-5-20251001"
	case "openai":
		return "gpt-5.4"
	case "gemini":
		return "gemini-2.0-flash"
	case "antigravity":
		return "claude-haiku-4-5-20251001"
	case "deepseek":
		return "deepseek-chat"
	case "kimi":
		return "kimi-k2"
	case "zhipu":
		return "GLM-5.1"
	case "grok":
		return "grok-4.3"
	default:
		return "claude-haiku-4-5-20251001"
	}
}
