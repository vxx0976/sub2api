package service

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/robfig/cron/v3"
)

const (
	HealthStatusAvailable   = "available"
	HealthStatusUnavailable = "unavailable"
	groupHealthMaxSample    = 3 // max accounts to test per group
	groupHealthMaxWorkers   = 5 // max concurrent group checks
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

		_, err := c.AddFunc("*/30 * * * *", func() { s.runHealthCheck() })
		if err != nil {
			logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] not started (invalid schedule): %v", err)
			return
		}
		s.cron = c
		s.cron.Start()
		logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] started (tick=every 30 minutes)")
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

	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] ListActive error: %v", err)
		return
	}
	if len(groups) == 0 {
		return
	}

	logger.LegacyPrintf("service.group_health_check", "[GroupHealthCheck] checking %d active groups", len(groups))

	sem := make(chan struct{}, groupHealthMaxWorkers)
	var wg sync.WaitGroup

	for i := range groups {
		sem <- struct{}{}
		wg.Add(1)
		go func(g *Group) {
			defer wg.Done()
			defer func() { <-sem }()
			s.checkOneGroup(ctx, g)
		}(&groups[i])
	}

	wg.Wait()
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

	// Randomly sample up to groupHealthMaxSample accounts
	sample := accounts
	if len(sample) > groupHealthMaxSample {
		rand.Shuffle(len(sample), func(i, j int) { sample[i], sample[j] = sample[j], sample[i] })
		sample = sample[:groupHealthMaxSample]
	}

	// Determine test model based on platform
	modelID := getDefaultTestModel(group.Platform)

	healthy := 0
	for _, acc := range sample {
		result, err := s.accountTestSvc.RunTestBackground(ctx, acc.ID, modelID)
		if err != nil {
			continue
		}
		if result.Status == "success" {
			healthy++
			// Auto-recover on success
			if s.rateLimitSvc != nil {
				_, _ = s.rateLimitSvc.RecoverAccountAfterSuccessfulTest(ctx, acc.ID)
			}
		}
	}

	status := HealthStatusUnavailable
	if healthy > 0 {
		status = HealthStatusAvailable
	}

	now := time.Now()
	if err := s.groupRepo.UpdateHealthStatus(ctx, group.ID, status, healthy, len(sample), now); err != nil {
		slog.Error("GroupHealthCheck: update status", "group_id", group.ID, "error", err)
	}
}

// getDefaultTestModel returns a lightweight model for health checks per platform.
func getDefaultTestModel(platform string) string {
	switch platform {
	case "anthropic":
		return "claude-haiku-4-5-20251001"
	case "openai":
		return "gpt-4o-mini"
	case "gemini":
		return "gemini-2.0-flash"
	case "antigravity":
		return "claude-haiku-4-5-20251001"
	default:
		return "claude-haiku-4-5-20251001"
	}
}
