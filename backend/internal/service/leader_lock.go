package service

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// LeaderLocker provides distributed leader election for periodic background tasks.
// It uses Redis SetNX as the primary lock mechanism and falls back to PostgreSQL
// advisory locks when Redis is unavailable.
//
// In "simple" run mode (single instance), locking is skipped entirely.
type LeaderLocker struct {
	db          *sql.DB
	redisClient *redis.Client
	cfg         *config.Config
	instanceID  string

	warnOnce sync.Once
}

// NewLeaderLocker creates a LeaderLocker. Both db and redisClient may be nil;
// when both are nil, TryAcquire always succeeds (single-instance assumption).
func NewLeaderLocker(db *sql.DB, redisClient *redis.Client, cfg *config.Config) *LeaderLocker {
	return &LeaderLocker{
		db:          db,
		redisClient: redisClient,
		cfg:         cfg,
		instanceID:  uuid.NewString(),
	}
}

var leaderLockReleaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`)

// TryAcquire attempts to acquire a distributed lock for the given key.
// Returns (release, true) on success, or (nil, false) if another instance holds the lock.
// The caller MUST call release() (if non-nil) when done.
func (l *LeaderLocker) TryAcquire(ctx context.Context, key string, ttl time.Duration) (release func(), ok bool) {
	if l == nil {
		return nil, true
	}
	if l.cfg != nil && l.cfg.RunMode == config.RunModeSimple {
		return nil, true
	}

	// Primary: Redis SetNX
	if l.redisClient != nil {
		acquired, err := l.redisClient.SetNX(ctx, key, l.instanceID, ttl).Result()
		if err == nil {
			if !acquired {
				return nil, false
			}
			return func() {
				_, _ = leaderLockReleaseScript.Run(ctx, l.redisClient, []string{key}, l.instanceID).Result()
			}, true
		}
		// Redis error: fall through to DB advisory lock
		l.warnOnce.Do(func() {
			logger.LegacyPrintf("service.leader_lock", "[LeaderLocker] Redis SetNX failed; falling back to DB advisory lock: %v", err)
		})
	}

	// Fallback: PostgreSQL advisory lock
	return tryAcquireDBAdvisoryLock(ctx, l.db, hashAdvisoryLockID(key))
}
