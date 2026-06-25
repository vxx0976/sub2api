package service

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// channelBalanceRefreshLeaderLockKey gates the periodic all-channel balance
	// sweep so only one instance queries upstreams per cycle in an HA deployment.
	channelBalanceRefreshLeaderLockKey = "leader:channel_balance_refresh"
	// channelBalanceRefreshLeaderLockTTL is a crash-safety bound only; it must exceed
	// the worst-case sweep duration so the lock is not stolen mid-run. Slightly above
	// channelBalanceRefreshRunTimeout.
	channelBalanceRefreshLeaderLockTTL = 12 * time.Minute
	// channelBalanceRefreshRunTimeout caps a single sweep's wall-clock time.
	channelBalanceRefreshRunTimeout = 10 * time.Minute
	// channelBalanceRefreshConcurrency bounds parallel upstream balance queries per sweep.
	channelBalanceRefreshConcurrency = 8
)

// ChannelBalanceRefreshService periodically refreshes the cached balance of every
// channel that has a balance_url configured — the background equivalent of the admin
// "刷新所有渠道额度" button. The on/off switch and interval are read from settings each
// cycle (defaults: enabled, 10 minutes), so changes take effect without a restart.
type ChannelBalanceRefreshService struct {
	channelService *ChannelService
	settingService *SettingService

	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup

	lockCache  LeaderLockCache
	db         *sql.DB
	instanceID string
}

func NewChannelBalanceRefreshService(channelService *ChannelService, settingService *SettingService) *ChannelBalanceRefreshService {
	return &ChannelBalanceRefreshService{
		channelService: channelService,
		settingService: settingService,
		stopCh:         make(chan struct{}),
		instanceID:     uuid.NewString(),
	}
}

// SetLeaderLock injects the leader-lock cache and DB used to elect a single instance
// per sweep. When both are nil the sweep runs ungated (single-instance / test behavior).
func (s *ChannelBalanceRefreshService) SetLeaderLock(lockCache LeaderLockCache, db *sql.DB) {
	if s == nil {
		return
	}
	s.lockCache = lockCache
	s.db = db
}

func (s *ChannelBalanceRefreshService) Start() {
	if s == nil || s.channelService == nil || s.settingService == nil {
		return
	}
	s.wg.Add(1)
	go s.loop()
}

func (s *ChannelBalanceRefreshService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

// loop waits one interval before the first sweep (avoids a balance-query storm on
// every rolling restart), then re-reads the interval each cycle. A per-cycle timer
// is used instead of a fixed ticker so interval changes apply on the next cycle.
func (s *ChannelBalanceRefreshService) loop() {
	defer s.wg.Done()
	for {
		timer := time.NewTimer(s.currentInterval())
		select {
		case <-s.stopCh:
			timer.Stop()
			return
		case <-timer.C:
		}
		s.runOnce()
	}
}

// currentInterval returns the configured sweep interval. Fail-open default is 10
// minutes (via GetChannelBalanceRefreshRuntime), and the value is clamped to >= 1m
// so a misconfigured/zero value can never spin the loop.
func (s *ChannelBalanceRefreshService) currentInterval() time.Duration {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	minutes := s.settingService.GetChannelBalanceRefreshRuntime(ctx).IntervalMinutes
	if minutes < channelBalanceRefreshIntervalMin {
		minutes = channelBalanceRefreshIntervalFallback
	}
	return time.Duration(minutes) * time.Minute
}

func (s *ChannelBalanceRefreshService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), channelBalanceRefreshRunTimeout)
	defer cancel()

	// Re-check the switch at fire time: a disabled cycle is skipped without giving up
	// the schedule, so re-enabling resumes on the next tick.
	if !s.settingService.GetChannelBalanceRefreshRuntime(ctx).Enabled {
		return
	}

	release, ok := tryAcquireSingletonLeaderLock(ctx, s.lockCache, s.db, channelBalanceRefreshLeaderLockKey, s.instanceID, channelBalanceRefreshLeaderLockTTL)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}

	refreshed, total, err := s.channelService.RefreshAllBalances(ctx, channelBalanceRefreshConcurrency)
	if err != nil {
		slog.Warn("channel_balance_refresh: sweep failed", "error", err)
		return
	}
	if total > 0 {
		slog.Info("channel_balance_refresh: swept channel balances", "refreshed", refreshed, "total", total)
	}
}
