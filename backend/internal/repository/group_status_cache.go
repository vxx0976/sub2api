package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

// 分组状态计数器缓存
//
// 双维度计数：
//   - 5 分钟窗口：用于实时状态判断（Key: gs:{gid}:ok:{bucket} / gs:{gid}:total:{bucket}，TTL 10min）
//   - 日维度：用于 30 天 uptime 条形图（Key: gs:{gid}:ok:d:{YYYYMMDD} / gs:{gid}:total:d:{YYYYMMDD}，TTL 31d）
const (
	gsKeyPrefix     = "gs:"
	gsKeyTTL        = 10 * time.Minute
	gsBucketSeconds = 300
	gsDailyTTL      = 31 * 24 * time.Hour
)

// GroupStatusCacheImpl 分组状态计数器 Redis 实现
type GroupStatusCacheImpl struct {
	rdb *redis.Client
}

// NewGroupStatusCache 创建分组状态计数器缓存
func NewGroupStatusCache(rdb *redis.Client) service.GroupStatusCache {
	return &GroupStatusCacheImpl{rdb: rdb}
}

// getServerTime 获取 Redis 服务端时间
func (c *GroupStatusCacheImpl) getServerTime(ctx context.Context) (time.Time, error) {
	t, err := c.rdb.Time(ctx).Result()
	if err != nil {
		return time.Time{}, fmt.Errorf("redis TIME: %w", err)
	}
	return t, nil
}

// Incr 递增分组计数（总量 +1，success 为 true 时成功 +1，latency 累加用于计算平均）
// 单次 TIME 调用 + 单个 TxPipeline，避免 bucket 边界分裂
func (c *GroupStatusCacheImpl) Incr(ctx context.Context, groupID int64, success bool, latencyMs int64) error {
	serverTime, err := c.getServerTime(ctx)
	if err != nil {
		return fmt.Errorf("group status incr: %w", err)
	}

	bucket := serverTime.Unix() / gsBucketSeconds
	dateStr := serverTime.UTC().Format("20060102")

	pipe := c.rdb.TxPipeline()

	// 5 分钟窗口 - total
	totalBucketKey := fmt.Sprintf("%s%d:total:%d", gsKeyPrefix, groupID, bucket)
	pipe.Incr(ctx, totalBucketKey)
	pipe.Expire(ctx, totalBucketKey, gsKeyTTL)

	// 日维度 - total
	totalDailyKey := fmt.Sprintf("%s%d:total:d:%s", gsKeyPrefix, groupID, dateStr)
	pipe.Incr(ctx, totalDailyKey)
	pipe.Expire(ctx, totalDailyKey, gsDailyTTL)

	// 5 分钟窗口 - latency 累加
	if latencyMs > 0 {
		latencyBucketKey := fmt.Sprintf("%s%d:latency:%d", gsKeyPrefix, groupID, bucket)
		pipe.IncrBy(ctx, latencyBucketKey, latencyMs)
		pipe.Expire(ctx, latencyBucketKey, gsKeyTTL)
	}

	if success {
		// 5 分钟窗口 - ok
		okBucketKey := fmt.Sprintf("%s%d:ok:%d", gsKeyPrefix, groupID, bucket)
		pipe.Incr(ctx, okBucketKey)
		pipe.Expire(ctx, okBucketKey, gsKeyTTL)

		// 日维度 - ok
		okDailyKey := fmt.Sprintf("%s%d:ok:d:%s", gsKeyPrefix, groupID, dateStr)
		pipe.Incr(ctx, okDailyKey)
		pipe.Expire(ctx, okDailyKey, gsDailyTTL)

		// 日维度 - latency 累加
		if latencyMs > 0 {
			latencyDailyKey := fmt.Sprintf("%s%d:latency:d:%s", gsKeyPrefix, groupID, dateStr)
			pipe.IncrBy(ctx, latencyDailyKey, latencyMs)
			pipe.Expire(ctx, latencyDailyKey, gsDailyTTL)
		}
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("group status incr: %w", err)
	}
	return nil
}

// GetGroupStatuses 批量获取多个分组的实时状态计数（当前 + 上一个 5 分钟窗口，消除桶边界冷启动）
func (c *GroupStatusCacheImpl) GetGroupStatuses(ctx context.Context, groupIDs []int64) (map[int64]service.GroupStatusData, error) {
	if len(groupIDs) == 0 {
		return map[int64]service.GroupStatusData{}, nil
	}

	serverTime, err := c.getServerTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("group status batch get: %w", err)
	}
	currentBucket := serverTime.Unix() / gsBucketSeconds
	prevBucket := currentBucket - 1

	pipe := c.rdb.Pipeline()
	type bucketCmds struct {
		okCur     *redis.StringCmd
		totalCur  *redis.StringCmd
		okPrev    *redis.StringCmd
		totalPrev *redis.StringCmd
		latencyCur  *redis.StringCmd
		latencyPrev *redis.StringCmd
	}
	cmds := make(map[int64]bucketCmds, len(groupIDs))
	for _, id := range groupIDs {
		cmds[id] = bucketCmds{
			okCur:       pipe.Get(ctx, fmt.Sprintf("%s%d:ok:%d", gsKeyPrefix, id, currentBucket)),
			totalCur:    pipe.Get(ctx, fmt.Sprintf("%s%d:total:%d", gsKeyPrefix, id, currentBucket)),
			okPrev:      pipe.Get(ctx, fmt.Sprintf("%s%d:ok:%d", gsKeyPrefix, id, prevBucket)),
			totalPrev:   pipe.Get(ctx, fmt.Sprintf("%s%d:total:%d", gsKeyPrefix, id, prevBucket)),
			latencyCur:  pipe.Get(ctx, fmt.Sprintf("%s%d:latency:%d", gsKeyPrefix, id, currentBucket)),
			latencyPrev: pipe.Get(ctx, fmt.Sprintf("%s%d:latency:%d", gsKeyPrefix, id, prevBucket)),
		}
	}

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("group status batch get: %w", err)
	}

	result := make(map[int64]service.GroupStatusData, len(groupIDs))
	for _, id := range groupIDs {
		bc := cmds[id]
		var data service.GroupStatusData
		if val, err := bc.okCur.Int64(); err == nil {
			data.Success += val
		}
		if val, err := bc.okPrev.Int64(); err == nil {
			data.Success += val
		}
		if val, err := bc.totalCur.Int64(); err == nil {
			data.Total += val
		}
		if val, err := bc.totalPrev.Int64(); err == nil {
			data.Total += val
		}
		// 计算平均延迟
		var totalLatency int64
		if val, err := bc.latencyCur.Int64(); err == nil {
			totalLatency += val
		}
		if val, err := bc.latencyPrev.Int64(); err == nil {
			totalLatency += val
		}
		if data.Total > 0 && totalLatency > 0 {
			data.AvgLatency = totalLatency / data.Total
		}
		result[id] = data
	}
	return result, nil
}

// GetDailyHistory 获取多个分组最近 N 天的每日状态数据
func (c *GroupStatusCacheImpl) GetDailyHistory(ctx context.Context, groupIDs []int64, days int) (map[int64][]service.DailyStatusData, error) {
	if len(groupIDs) == 0 || days <= 0 {
		return map[int64][]service.DailyStatusData{}, nil
	}

	serverTime, err := c.getServerTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("group status daily history: %w", err)
	}
	today := serverTime.UTC()

	// 生成日期列表（从 days-1 天前到今天）
	dates := make([]string, days)
	for i := 0; i < days; i++ {
		d := today.AddDate(0, 0, -(days - 1 - i))
		dates[i] = d.Format("20060102")
	}

	// Pipeline 批量读取
	pipe := c.rdb.Pipeline()
	type cmdTriple struct {
		ok      *redis.StringCmd
		total   *redis.StringCmd
		latency *redis.StringCmd
	}
	cmds := make(map[int64][]cmdTriple, len(groupIDs))
	for _, gid := range groupIDs {
		triples := make([]cmdTriple, days)
		for i, dateStr := range dates {
			triples[i] = cmdTriple{
				ok:      pipe.Get(ctx, fmt.Sprintf("%s%d:ok:d:%s", gsKeyPrefix, gid, dateStr)),
				total:   pipe.Get(ctx, fmt.Sprintf("%s%d:total:d:%s", gsKeyPrefix, gid, dateStr)),
				latency: pipe.Get(ctx, fmt.Sprintf("%s%d:latency:d:%s", gsKeyPrefix, gid, dateStr)),
			}
		}
		cmds[gid] = triples
	}

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("group status daily history: %w", err)
	}

	result := make(map[int64][]service.DailyStatusData, len(groupIDs))
	for _, gid := range groupIDs {
		history := make([]service.DailyStatusData, days)
		for i, triple := range cmds[gid] {
			history[i].Date = dates[i]
			if val, err := triple.ok.Int64(); err == nil {
				history[i].Success = val
			}
			if val, err := triple.total.Int64(); err == nil {
				history[i].Total = val
			}
			// 计算平均延迟
			if val, err := triple.latency.Int64(); err == nil && history[i].Total > 0 {
				history[i].AvgLatency = val / history[i].Total
			}
		}
		result[gid] = history
	}
	return result, nil
}
