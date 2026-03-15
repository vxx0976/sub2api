package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

// 分组状态计数器缓存常量定义
//
// 设计说明：
// 使用 Redis 简单计数器跟踪每个分组在 5 分钟窗口内的请求成功/总量：
// - Key: gs:{groupID}:ok:{5minBucket} / gs:{groupID}:total:{5minBucket}
// - Value: 计数
// - TTL: 10 分钟（覆盖当前窗口 + 冗余）
//
// 使用 TxPipeline（MULTI/EXEC）执行 INCR + EXPIRE，保证原子性且兼容 Redis Cluster。
// 通过 rdb.Time() 获取服务端时间，避免多实例时钟不同步。
const (
	gsKeyPrefix = "gs:"
	gsKeyTTL    = 10 * time.Minute
	// 5 分钟窗口 = unix / 300
	gsBucketSeconds = 300
)

// GroupStatusCacheImpl 分组状态计数器 Redis 实现
type GroupStatusCacheImpl struct {
	rdb *redis.Client
}

// NewGroupStatusCache 创建分组状态计数器缓存
func NewGroupStatusCache(rdb *redis.Client) service.GroupStatusCache {
	return &GroupStatusCacheImpl{rdb: rdb}
}

// currentBucket 获取当前 5 分钟窗口的 bucket 值
func (c *GroupStatusCacheImpl) currentBucket(ctx context.Context) (int64, error) {
	serverTime, err := c.rdb.Time(ctx).Result()
	if err != nil {
		return 0, fmt.Errorf("redis TIME: %w", err)
	}
	return serverTime.Unix() / gsBucketSeconds, nil
}

// IncrSuccess 递增分组成功计数
func (c *GroupStatusCacheImpl) IncrSuccess(ctx context.Context, groupID int64) error {
	bucket, err := c.currentBucket(ctx)
	if err != nil {
		return fmt.Errorf("group status incr success: %w", err)
	}

	key := fmt.Sprintf("%s%d:ok:%d", gsKeyPrefix, groupID, bucket)
	pipe := c.rdb.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, gsKeyTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("group status incr success: %w", err)
	}
	return nil
}

// IncrTotal 递增分组总请求计数
func (c *GroupStatusCacheImpl) IncrTotal(ctx context.Context, groupID int64) error {
	bucket, err := c.currentBucket(ctx)
	if err != nil {
		return fmt.Errorf("group status incr total: %w", err)
	}

	key := fmt.Sprintf("%s%d:total:%d", gsKeyPrefix, groupID, bucket)
	pipe := c.rdb.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, gsKeyTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("group status incr total: %w", err)
	}
	return nil
}

// GetGroupStatuses 批量获取多个分组的状态计数
func (c *GroupStatusCacheImpl) GetGroupStatuses(ctx context.Context, groupIDs []int64) (map[int64]service.GroupStatusData, error) {
	if len(groupIDs) == 0 {
		return map[int64]service.GroupStatusData{}, nil
	}

	bucket, err := c.currentBucket(ctx)
	if err != nil {
		return nil, fmt.Errorf("group status batch get: %w", err)
	}
	bucketStr := strconv.FormatInt(bucket, 10)

	// 使用 Pipeline 批量 GET
	pipe := c.rdb.Pipeline()
	okCmds := make(map[int64]*redis.StringCmd, len(groupIDs))
	totalCmds := make(map[int64]*redis.StringCmd, len(groupIDs))
	for _, id := range groupIDs {
		okKey := fmt.Sprintf("%s%d:ok:%s", gsKeyPrefix, id, bucketStr)
		totalKey := fmt.Sprintf("%s%d:total:%s", gsKeyPrefix, id, bucketStr)
		okCmds[id] = pipe.Get(ctx, okKey)
		totalCmds[id] = pipe.Get(ctx, totalKey)
	}

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("group status batch get: %w", err)
	}

	result := make(map[int64]service.GroupStatusData, len(groupIDs))
	for _, id := range groupIDs {
		var data service.GroupStatusData
		if val, err := okCmds[id].Int64(); err == nil {
			data.Success = val
		}
		if val, err := totalCmds[id].Int64(); err == nil {
			data.Total = val
		}
		result[id] = data
	}
	return result, nil
}
