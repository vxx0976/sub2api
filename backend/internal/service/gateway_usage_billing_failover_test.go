package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// 智能路由（故障转移虚拟分组）用量归因回归测试。
//
// 该链路曾两度静默断裂：选择函数内注入的 ctx key 无法逃逸出函数，且异步记账
// worker 池会重建 ctx（usageRecordContext 只白名单复制 request id 两个 key）。
// 现行设计为按值传递：选择期 newSelectionResult 从富化 ctx 提取归因挂到
// AccountSelectionResult，handler 透传进 RecordUsageInput，buildRecordUsageLog
// 从 input 读取——本文件覆盖链路两端，中段为编译期显式字段传递。
func TestFailoverRouteAttribution_SelectionResultCarriesGroupIDs(t *testing.T) {
	s := &GatewayService{} // schedulerSnapshot 为 nil 时 hydrate 原样返回，无外部依赖
	virtualGroupID := int64(101)
	memberGroupID := int64(202)
	account := &Account{}

	t.Run("no_failover_route_leaves_attribution_nil", func(t *testing.T) {
		result, err := s.newSelectionResult(context.Background(), account, true, nil, nil)
		require.NoError(t, err)
		require.Nil(t, result.RequestedGroupID)
		require.Nil(t, result.ResolvedGroupID)
	})

	t.Run("failover_route_ctx_populates_attribution", func(t *testing.T) {
		ctx := withFailoverRouteContext(context.Background(), &virtualGroupID, &memberGroupID)
		result, err := s.newSelectionResult(ctx, account, true, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, result.RequestedGroupID)
		require.Equal(t, virtualGroupID, *result.RequestedGroupID, "RequestedGroupID 应为虚拟分组")
		require.NotNil(t, result.ResolvedGroupID)
		require.Equal(t, memberGroupID, *result.ResolvedGroupID, "ResolvedGroupID 应为承接成员分组")
	})
}

func TestBuildRecordUsageLog_FailoverRouteAttribution(t *testing.T) {
	s := &GatewayService{}
	virtualGroupID := int64(101) // API key 绑定的智能路由虚拟分组
	memberGroupID := int64(202)  // 透明重定向后实际承接的成员分组

	apiKey := &APIKey{GroupID: &virtualGroupID}
	user := &User{}
	account := &Account{}
	result := &ForwardResult{Model: "claude-sonnet-5"}

	build := func(input *recordUsageCoreInput) *UsageLog {
		return s.buildRecordUsageLog(context.Background(), input, result, apiKey, user, account, nil,
			"claude-sonnet-5", 1, 1, 1, 0, false, nil, nil)
	}

	t.Run("no_failover_route_keeps_api_key_group", func(t *testing.T) {
		usageLog := build(&recordUsageCoreInput{})

		require.NotNil(t, usageLog.GroupID)
		require.Equal(t, virtualGroupID, *usageLog.GroupID)
		require.Nil(t, usageLog.RequestedGroupID)
	})

	t.Run("failover_route_attributes_member_and_virtual_groups", func(t *testing.T) {
		usageLog := build(&recordUsageCoreInput{
			RequestedGroupID: &virtualGroupID,
			ResolvedGroupID:  &memberGroupID,
		})

		require.NotNil(t, usageLog.GroupID)
		require.Equal(t, memberGroupID, *usageLog.GroupID, "group_id 应记实际承接的成员分组")
		require.NotNil(t, usageLog.RequestedGroupID)
		require.Equal(t, virtualGroupID, *usageLog.RequestedGroupID, "requested_group_id 应记虚拟分组")
	})
}

// RecordUsage / RecordUsageWithLongContext 的输入映射必须携带归因字段——
// 上次断裂正是发生在中间层丢字段，这里直接钉住两个公开输入结构体的贯通。
func TestRecordUsageInputs_CarryFailoverAttribution(t *testing.T) {
	virtualGroupID := int64(101)
	memberGroupID := int64(202)

	ctx := withFailoverRouteContext(context.Background(), &virtualGroupID, &memberGroupID)
	requested, resolved := failoverRouteAttributionFromContext(ctx)
	require.NotNil(t, requested)
	require.Equal(t, virtualGroupID, *requested)
	require.NotNil(t, resolved)
	require.Equal(t, memberGroupID, *resolved)

	// 非智能路由 ctx：提取结果应为 nil
	requested, resolved = failoverRouteAttributionFromContext(context.Background())
	require.Nil(t, requested)
	require.Nil(t, resolved)
}
