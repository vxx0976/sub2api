//go:build unit

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// TestSchedulerCachePreservesActiveWindow 钉死调度快照 metadata payload 必须保留定时上线窗口。
//
// 选号热路径读的是 metadata 快照，窗口由 Account.IsSchedulable 在选号时按当前时刻判定；
// 字段缺失时 IsWithinActiveWindow 视为「全天可用」，窗口外的账号照常被调度，而且没有任何报错。
func TestSchedulerCachePreservesActiveWindow(t *testing.T) {
	start, end := "06:00", "00:00"
	account := service.Account{
		ID:              9101,
		Name:            "active-window",
		Platform:        service.PlatformOpenAI,
		Type:            service.AccountTypeOAuth,
		Status:          service.StatusActive,
		Schedulable:     true,
		Concurrency:     1,
		ActiveStartTime: &start,
		ActiveEndTime:   &end,
	}
	outOfWindow := time.Date(2026, 9, 14, 3, 0, 0, 0, time.Local)
	inWindow := time.Date(2026, 9, 14, 12, 0, 0, 0, time.Local)

	t.Run("metadata payload keeps the window", func(t *testing.T) {
		_, meta, err := marshalSchedulerCacheAccount(account)
		require.NoError(t, err)
		decoded, err := decodeCachedAccount(meta)
		require.NoError(t, err)
		require.NotNil(t, decoded.ActiveStartTime)
		require.NotNil(t, decoded.ActiveEndTime)
		require.False(t, decoded.IsWithinActiveWindow(outOfWindow), "跨午夜窗口 06:00-00:00 在 03:00 应处于下线状态")
		require.True(t, decoded.IsWithinActiveWindow(inWindow))
	})

	t.Run("snapshot read path keeps the window", func(t *testing.T) {
		cache := newSchedulerCacheUnit(t)
		ctx := context.Background()
		bucket := service.SchedulerBucket{GroupID: 1, Platform: service.PlatformOpenAI, Mode: service.SchedulerModeSingle}
		token, err := cache.CaptureBucketWriteToken(ctx, bucket)
		require.NoError(t, err)
		require.NoError(t, cache.SetSnapshot(ctx, bucket, token, []service.Account{account}))

		got, hit, err := cache.GetSnapshot(ctx, bucket)
		require.NoError(t, err)
		require.True(t, hit)
		require.Len(t, got, 1)
		require.NotNil(t, got[0].ActiveStartTime, "选号读取的快照账号不得丢失定时上线窗口")
		require.False(t, got[0].IsWithinActiveWindow(outOfWindow))
	})
}
