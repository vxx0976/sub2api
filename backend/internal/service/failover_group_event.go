package service

import (
	"context"
	"time"
)

// FailoverEventReason 枚举智能路由状态机触发的事件类型。
const (
	FailoverEventReasonSoftDemote         = "soft_demote"
	FailoverEventReasonProbeRecovered     = "probe_recovered"
	FailoverEventReasonHealthReconcile    = "health_reconcile"
	FailoverEventReasonManualPin          = "manual_pin"
	FailoverEventReasonManualUnpin        = "manual_unpin"
	FailoverEventReasonManualUnpinExpired = "manual_unpin_expired"
	FailoverEventReasonBootstrap          = "bootstrap"
	FailoverEventReasonNoMembersAvailable = "no_members_available"
	FailoverEventReasonConfigChange       = "config_change"
)

// FailoverGroupEvent 记录虚拟分组 active 成员的一次变更事件。
type FailoverGroupEvent struct {
	ID             int64     `json:"id"`
	VirtualGroupID int64     `json:"virtual_group_id"`
	FromMemberID   *int64    `json:"from_member_id,omitempty"`
	ToMemberID     *int64    `json:"to_member_id,omitempty"`
	Reason         string    `json:"reason"`
	TriggeredBy    *int64    `json:"triggered_by,omitempty"`
	Note           *string   `json:"note,omitempty"`
	OccurredAt     time.Time `json:"occurred_at"`
}

// FailoverEventRepository 仅追加 / 读取的接口。
type FailoverEventRepository interface {
	Create(ctx context.Context, event *FailoverGroupEvent) error
	ListByGroupID(ctx context.Context, virtualGroupID int64, limit int) ([]*FailoverGroupEvent, error)
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

// FailoverMemberSnapshot 智能路由成员的一次快照，用于详情页渲染。
type FailoverMemberSnapshot struct {
	GroupID              int64      `json:"group_id"`
	Name                 string     `json:"name"`
	Platform             string     `json:"platform"`
	HealthStatus         string     `json:"health_status"`
	SchedulableAccounts  int        `json:"schedulable_accounts"`
	LastHealthCheckAt    *time.Time `json:"last_health_check_at,omitempty"`
	HealthyAccounts      int        `json:"healthy_accounts"`
	TotalCheckedAccounts int        `json:"total_checked_accounts"`
}

// FailoverStatus 智能路由详情响应。
type FailoverStatus struct {
	VirtualGroupID       int64                    `json:"virtual_group_id"`
	Name                 string                   `json:"name"`
	Platform             string                   `json:"platform"`
	ActiveMemberID       *int64                   `json:"active_member_id,omitempty"`
	PinMemberID          *int64                   `json:"pin_member_id,omitempty"`
	PinExpiresAt         *time.Time               `json:"pin_expires_at,omitempty"`
	Members              []FailoverMemberSnapshot `json:"members"`
	RecentEvents         []*FailoverGroupEvent    `json:"recent_events"`
}

// FailoverMemberUsage 展示单个成员在某时间窗内的实际承接用量。
type FailoverMemberUsage struct {
	GroupID  int64   `json:"group_id"`
	Name     string  `json:"name"`
	Requests int64   `json:"requests"`
	Tokens   int64   `json:"tokens"`
	Cost     float64 `json:"cost"`
}
