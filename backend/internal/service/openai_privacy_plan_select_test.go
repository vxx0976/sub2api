//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// selectChatGPTPlanType 是 selectChatGPTAccount 的双返回值形态，仅本测试使用。
// 放在测试文件而非生产代码里：生产侧无人调用它，留在那儿会被 golangci-lint 的
// unused 判为死码（lint 不带 unit tag，看不到这里的调用）。
func selectChatGPTPlanType(accounts map[string]any, orgID string) (planType, expiresAt string) {
	planType, expiresAt, _ = selectChatGPTAccount(accounts, orgID)
	return planType, expiresAt
}

func acct(planType, expiresAt string, isDefault bool) map[string]any {
	account := map[string]any{}
	if planType != "" {
		account["plan_type"] = planType
	}
	if isDefault {
		account["is_default"] = true
	}
	entitlement := map[string]any{}
	if expiresAt != "" {
		entitlement["expires_at"] = expiresAt
	}
	return map[string]any{"account": account, "entitlement": entitlement}
}

func TestSelectChatGPTPlanType_ConsumerWinsOverBusiness(t *testing.T) {
	accounts := map[string]any{
		// poid 指向 business 工作区，但用户个人是 Pro：应取 Pro。
		// 到期日均取远期，避免触发 selectChatGPTPlanType 内的“跳过已过期账号”过滤。
		"org-biz": acct("self_serve_business_usage_based", "2099-01-01T00:00:00+00:00", true),
		"org-pro": acct("pro", "2099-12-31T00:00:00+00:00", false),
	}
	plan, expiresAt := selectChatGPTPlanType(accounts, "org-biz")
	require.Equal(t, "pro", plan)
	require.Equal(t, "2099-12-31T00:00:00+00:00", expiresAt)
}

func TestSelectChatGPTPlanType_BusinessOnly(t *testing.T) {
	accounts := map[string]any{
		"org-biz": acct("self_serve_business_usage_based", "2099-01-01T00:00:00+00:00", true),
	}
	plan, _ := selectChatGPTPlanType(accounts, "org-biz")
	require.Equal(t, "self_serve_business_usage_based", plan)
}

func TestSelectChatGPTPlanType_OrgMatchBreaksConsumerTie(t *testing.T) {
	// 个人订阅之间(plus vs team)按 orgID 命中决定，保持原有"token 实际作用 org"语义。
	accounts := map[string]any{
		"org-team": acct("team", "2099-06-01T00:00:00+00:00", false),
		"org-plus": acct("plus", "2099-07-01T00:00:00+00:00", false),
	}
	plan, expiresAt := selectChatGPTPlanType(accounts, "org-team")
	require.Equal(t, "team", plan)
	require.Equal(t, "2099-06-01T00:00:00+00:00", expiresAt)
}

func TestSelectChatGPTPlanType_BusinessBeatsFree(t *testing.T) {
	accounts := map[string]any{
		"org-free": acct("free", "", true),
		"org-biz":  acct("self_serve_business_usage_based", "2099-01-01T00:00:00+00:00", false),
	}
	plan, _ := selectChatGPTPlanType(accounts, "")
	require.Equal(t, "self_serve_business_usage_based", plan)
}

func TestSelectChatGPTPlanType_DeterministicTieBreak(t *testing.T) {
	// 两个同优先级、均无 orgID 命中、均非 default：必须在 Go map 随机遍历下结果稳定。
	accounts := map[string]any{
		"org-a": acct("plus", "2099-05-01T00:00:00+00:00", false),
		"org-b": acct("team", "2099-06-01T00:00:00+00:00", false),
	}
	first, _ := selectChatGPTPlanType(accounts, "")
	for i := 0; i < 50; i++ {
		got, _ := selectChatGPTPlanType(accounts, "")
		require.Equal(t, first, got, "selection must be deterministic across runs")
	}
	// id "org-a" < "org-b"，兜底取 org-a 的 plus。
	require.Equal(t, "plus", first)
}

func TestSelectChatGPTPlanType_Empty(t *testing.T) {
	plan, expiresAt := selectChatGPTPlanType(map[string]any{}, "")
	require.Equal(t, "", plan)
	require.Equal(t, "", expiresAt)
}

func TestSelectChatGPTPlanType_SkipsExpiredAndDeactivated(t *testing.T) {
	// 上游并入的行为：已过期/已停用账号不参与选号，避免“曾加入过”污染计划展示。
	expiredBiz := acct("self_serve_business_usage_based", "2000-01-01T00:00:00+00:00", true)
	activePlus := acct("plus", "2099-07-01T00:00:00+00:00", false)
	accounts := map[string]any{
		"org-expired": expiredBiz,
		"org-active":  activePlus,
	}
	plan, _ := selectChatGPTPlanType(accounts, "org-expired")
	require.Equal(t, "plus", plan, "过期的 business 账号应被跳过，取仍有效的个人订阅")

	deactivated := acct("pro", "2099-01-01T00:00:00+00:00", true)
	deactivated["deactivated"] = true
	plan2, _ := selectChatGPTPlanType(map[string]any{"org-off": deactivated}, "org-off")
	require.Equal(t, "", plan2, "已停用账号应被跳过")
}

func TestPlanTypeClassifiers(t *testing.T) {
	require.True(t, isBusinessPlanType("self_serve_business_usage_based"))
	require.True(t, isBusinessPlanType("Enterprise"))
	require.False(t, isBusinessPlanType("pro"))

	require.True(t, isConsumerPlanType("pro"))
	require.True(t, isConsumerPlanType("chatgptpro"))
	require.True(t, isConsumerPlanType("plus"))
	require.True(t, isConsumerPlanType("team"))
	require.False(t, isConsumerPlanType("free"))
	require.False(t, isConsumerPlanType("self_serve_business_usage_based"))
}
