package service

import (
	"context"
	"fmt"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrGroupNotFound = infraerrors.NotFound("GROUP_NOT_FOUND", "group not found")
	ErrGroupExists   = infraerrors.Conflict("GROUP_EXISTS", "group name already exists")
)

type GroupRepository interface {
	Create(ctx context.Context, group *Group) error
	GetByID(ctx context.Context, id int64) (*Group, error)
	GetByIDLite(ctx context.Context, id int64) (*Group, error)
	Update(ctx context.Context, group *Group) error
	Delete(ctx context.Context, id int64) error
	DeleteCascade(ctx context.Context, id int64, migrateToGroupID *int64) ([]int64, error)

	List(ctx context.Context, params pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error)
	ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, status, search string, isExclusive *bool, isPurchasable *bool) ([]Group, *pagination.PaginationResult, error)
	ListActive(ctx context.Context) ([]Group, error)
	ListActiveByPlatform(ctx context.Context, platform string) ([]Group, error)
	ListPurchasable(ctx context.Context) ([]Group, error)
	// ListPurchasableByOwnerID returns purchasable subscription groups owned by the given reseller.
	ListPurchasableByOwnerID(ctx context.Context, ownerID int64) ([]Group, error)

	ExistsByName(ctx context.Context, name string) (bool, error)
	GetAccountCount(ctx context.Context, groupID int64) (total int64, active int64, err error)
	DeleteAccountGroupsByGroupID(ctx context.Context, groupID int64) (int64, error)
	// GetAccountIDsByGroupIDs 获取多个分组的所有账号 ID（去重）
	GetAccountIDsByGroupIDs(ctx context.Context, groupIDs []int64) ([]int64, error)
	// BindAccountsToGroup 将多个账号绑定到指定分组
	BindAccountsToGroup(ctx context.Context, groupID int64, accountIDs []int64) error
	// UpdateSortOrders 批量更新分组排序
	UpdateSortOrders(ctx context.Context, updates []GroupSortOrderUpdate) error
	// UpdateHealthStatus 更新分组健康检查状态
	UpdateHealthStatus(ctx context.Context, groupID int64, status string, healthy int, total int, checkedAt time.Time) error
	// CountByOwnerID returns the number of groups owned by the given user.
	CountByOwnerID(ctx context.Context, ownerID int64) (int64, error)

	// ListFailoverGroups 返回所有启用状态的智能路由（虚拟故障转移分组），包括成员 ID 列表和激活成员。
	ListFailoverGroups(ctx context.Context) ([]*Group, error)
	// ListFailoverGroupsReferencing 返回将 memberID 包含在 failover_member_ids 内的虚拟分组，用于删除校验。
	ListFailoverGroupsReferencing(ctx context.Context, memberID int64) ([]*Group, error)
	// UpdateFailoverActive 以乐观锁方式更新激活成员 id，成功返回 true；版本号不匹配返回 false 且 err=nil。
	UpdateFailoverActive(ctx context.Context, groupID int64, newMemberID int64, expectedVersion int64) (bool, error)
	// SetFailoverPin 设置或刷新手动锁定；expiresAt 为 nil 表示永久 pin。
	SetFailoverPin(ctx context.Context, groupID int64, memberID int64, expiresAt *time.Time) error
	// ClearFailoverPin 清除手动锁定。
	ClearFailoverPin(ctx context.Context, groupID int64) error
	// UpdateFailoverConfig 更新智能路由的成员列表、启用位以及初始激活成员（用于 admin 创建/更新）。
	UpdateFailoverConfig(ctx context.Context, groupID int64, isFailover bool, memberIDs []int64, activeMemberID *int64) error
	// CountSchedulableAccountsByGroup 统计给定分组当前 available 状态下可调度的账号数量。
	CountSchedulableAccountsByGroup(ctx context.Context, groupID int64) (int, error)
}

// GroupSortOrderUpdate 分组排序更新
type GroupSortOrderUpdate struct {
	ID        int64 `json:"id"`
	SortOrder int   `json:"sort_order"`
}

// CreateGroupRequest 创建分组请求
type CreateGroupRequest struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	RateMultiplier float64 `json:"rate_multiplier"`
	IsExclusive    bool    `json:"is_exclusive"`
}

// UpdateGroupRequest 更新分组请求
type UpdateGroupRequest struct {
	Name           *string  `json:"name"`
	Description    *string  `json:"description"`
	RateMultiplier *float64 `json:"rate_multiplier"`
	IsExclusive    *bool    `json:"is_exclusive"`
	Status         *string  `json:"status"`
}

// GroupService 分组管理服务
type GroupService struct {
	groupRepo            GroupRepository
	authCacheInvalidator APIKeyAuthCacheInvalidator
}

// NewGroupService 创建分组服务实例
func NewGroupService(groupRepo GroupRepository, authCacheInvalidator APIKeyAuthCacheInvalidator) *GroupService {
	return &GroupService{
		groupRepo:            groupRepo,
		authCacheInvalidator: authCacheInvalidator,
	}
}

// Create 创建分组
func (s *GroupService) Create(ctx context.Context, req CreateGroupRequest) (*Group, error) {
	// 检查名称是否已存在
	exists, err := s.groupRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("check group exists: %w", err)
	}
	if exists {
		return nil, ErrGroupExists
	}

	// 创建分组
	group := &Group{
		Name:             req.Name,
		Description:      req.Description,
		Platform:         PlatformAnthropic,
		RateMultiplier:   req.RateMultiplier,
		IsExclusive:      req.IsExclusive,
		Status:           StatusActive,
		SubscriptionType: SubscriptionTypeStandard,
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}

	return group, nil
}

// GetByID 根据ID获取分组
func (s *GroupService) GetByID(ctx context.Context, id int64) (*Group, error) {
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}
	return group, nil
}

// List 获取分组列表
func (s *GroupService) List(ctx context.Context, params pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error) {
	groups, pagination, err := s.groupRepo.List(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("list groups: %w", err)
	}
	return groups, pagination, nil
}

// ListActive 获取活跃分组列表
func (s *GroupService) ListActive(ctx context.Context) ([]Group, error) {
	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}
	return groups, nil
}

// Update 更新分组
func (s *GroupService) Update(ctx context.Context, id int64, req UpdateGroupRequest) (*Group, error) {
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}

	// 更新字段
	if req.Name != nil && *req.Name != group.Name {
		// 检查新名称是否已存在
		exists, err := s.groupRepo.ExistsByName(ctx, *req.Name)
		if err != nil {
			return nil, fmt.Errorf("check group exists: %w", err)
		}
		if exists {
			return nil, ErrGroupExists
		}
		group.Name = *req.Name
	}

	if req.Description != nil {
		group.Description = *req.Description
	}

	if req.RateMultiplier != nil {
		group.RateMultiplier = *req.RateMultiplier
	}

	if req.IsExclusive != nil {
		group.IsExclusive = *req.IsExclusive
	}

	if req.Status != nil {
		group.Status = *req.Status
	}

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, fmt.Errorf("update group: %w", err)
	}
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByGroupID(ctx, id)
	}

	return group, nil
}

// Delete 删除分组
func (s *GroupService) Delete(ctx context.Context, id int64) error {
	// 检查分组是否存在
	_, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get group: %w", err)
	}

	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByGroupID(ctx, id)
	}
	if err := s.groupRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete group: %w", err)
	}

	return nil
}

// GetStats 获取分组统计信息
func (s *GroupService) GetStats(ctx context.Context, id int64) (map[string]any, error) {
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}

	// 获取账号数量
	accountCount, _, err := s.groupRepo.GetAccountCount(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get account count: %w", err)
	}

	stats := map[string]any{
		"id":              group.ID,
		"name":            group.Name,
		"rate_multiplier": group.RateMultiplier,
		"is_exclusive":    group.IsExclusive,
		"status":          group.Status,
		"account_count":   accountCount,
	}

	return stats, nil
}
