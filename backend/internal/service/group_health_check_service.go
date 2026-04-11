package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/robfig/cron/v3"
)

const (
	HealthStatusAvailable   = "available"
	HealthStatusUnavailable = "unavailable"
	groupHealthMaxWorkers   = 5 // max concurrent group checks
)

type GroupHealthCheckService struct {
	groupRepo        GroupRepository
	accountRepo      AccountRepository
	accountTestSvc   *AccountTestService
	rateLimitSvc     *RateLimitService
	cfg              *config.Config
	locker           *LeaderLocker
	failoverGroupSvc *FailoverGroupService

	cron      *cron.Cron
	startOnce sync.Once
	stopOnce  sync.Once
}

// SetFailoverGroupService 由 wiring 阶段注入，避免构造时的循环依赖。
// 当设置后，每轮健康检查结束会触发一次智能路由 reconcile。
func (s *GroupHealthCheckService) SetFailoverGroupService(svc *FailoverGroupService) {
	if s == nil {
		return
	}
	s.failoverGroupSvc = svc
}

func NewGroupHealthCheckService(
	groupRepo GroupRepository,
	accountRepo AccountRepository,
	accountTestSvc *AccountTestService,
	rateLimitSvc *RateLimitService,
	cfg *config.Config,
	locker *LeaderLocker,
) *GroupHealthCheckService {
	return &GroupHealthCheckService{
		groupRepo:      groupRepo,
		accountRepo:    accountRepo,
		accountTestSvc: accountTestSvc,
		rateLimitSvc:   rateLimitSvc,
		cfg:            cfg,
		locker:         locker,
	}
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

	release, ok := s.locker.TryAcquire(ctx, "leader:group_health_check", 15*time.Minute)
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
		// 智能路由（虚拟故障转移分组）没有自己的账号，跳过硬健康检查
		if g.IsFailoverGroup {
			continue
		}
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
		// 即使没有成员分组需要检查，仍然触发一次虚拟路由的 reconcile（它依赖 DB 快照）
		s.runFailoverReconcile(ctx)
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

	// 健康检查写入的最新 health_status 是故障转移 reconcile 的关键辅助信号
	s.runFailoverReconcile(ctx)
}

// runFailoverReconcile 在仍持有 leader 锁的前提下触发智能路由 reconcile。
// 失败不影响健康检查本身的正确性，仅打一条日志。
func (s *GroupHealthCheckService) runFailoverReconcile(ctx context.Context) {
	if s.failoverGroupSvc == nil {
		return
	}
	if err := s.failoverGroupSvc.ReconcileAll(ctx, 0); err != nil {
		logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] failover reconcile failed: %v", err)
	}
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
		return
	}

	// Determine test model based on platform
	modelID := getDefaultTestModel(group.Platform)

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
	default:
		return "claude-haiku-4-5-20251001"
	}
}
