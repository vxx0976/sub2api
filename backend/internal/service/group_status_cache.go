package service

import "context"

// GroupStatusData holds the success/total counters for a group
type GroupStatusData struct {
	Success int64
	Total   int64
}

// DailyStatusData 每日状态数据
type DailyStatusData struct {
	Date    string `json:"date"`    // YYYYMMDD
	Success int64  `json:"success"`
	Total   int64  `json:"total"`
}

// GroupStatusCache 分组状态计数器缓存
type GroupStatusCache interface {
	// Incr 递增分组计数（总量 +1，success 为 true 时成功 +1）
	// 单次 TIME 调用 + 单个 Pipeline，避免 bucket 边界分裂
	Incr(ctx context.Context, groupID int64, success bool) error
	// GetGroupStatuses 批量获取多个分组的实时状态计数（当前 + 上一个 5 分钟窗口）
	GetGroupStatuses(ctx context.Context, groupIDs []int64) (map[int64]GroupStatusData, error)
	// GetDailyHistory 获取多个分组最近 N 天的每日状态数据
	GetDailyHistory(ctx context.Context, groupIDs []int64, days int) (map[int64][]DailyStatusData, error)
}
