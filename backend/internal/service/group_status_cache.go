package service

import "context"

// GroupStatusData holds the success/total counters for a group
type GroupStatusData struct {
	Success int64
	Total   int64
}

// GroupStatusCache 分组实时状态计数器缓存
// 使用 Redis 计数器跟踪每个分组在 5 分钟窗口内的请求成功/总量
type GroupStatusCache interface {
	// IncrSuccess 递增分组成功计数
	IncrSuccess(ctx context.Context, groupID int64) error
	// IncrTotal 递增分组总请求计数
	IncrTotal(ctx context.Context, groupID int64) error
	// GetGroupStatuses 批量获取多个分组的状态计数
	GetGroupStatuses(ctx context.Context, groupIDs []int64) (map[int64]GroupStatusData, error)
}
