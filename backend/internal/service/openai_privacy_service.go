package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

// PrivacyClientFactory creates an HTTP client for privacy API calls.
// Injected from repository layer to avoid import cycles.
type PrivacyClientFactory func(proxyURL string) (*req.Client, error)

const (
	openAISettingsURL = "https://chatgpt.com/backend-api/settings/account_user_setting"

	PrivacyModeTrainingOff = "training_off"
	PrivacyModeFailed      = "training_set_failed"
	PrivacyModeCFBlocked   = "training_set_cf_blocked"
)

func shouldSkipOpenAIPrivacyEnsure(extra map[string]any) bool {
	if extra == nil {
		return false
	}
	raw, ok := extra["privacy_mode"]
	if !ok {
		return false
	}
	mode, _ := raw.(string)
	mode = strings.TrimSpace(mode)
	return mode != PrivacyModeFailed && mode != PrivacyModeCFBlocked
}

// disableOpenAITraining calls ChatGPT settings API to turn off "Improve the model for everyone".
// Returns privacy_mode value: "training_off" on success, "cf_blocked" / "failed" on failure.
func disableOpenAITraining(ctx context.Context, clientFactory PrivacyClientFactory, accessToken, proxyURL string) string {
	if accessToken == "" || clientFactory == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client, err := clientFactory(proxyURL)
	if err != nil {
		slog.Warn("openai_privacy_client_error", "error", err.Error())
		return PrivacyModeFailed
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetHeader("Origin", "https://chatgpt.com").
		SetHeader("Referer", "https://chatgpt.com/").
		SetHeader("Accept", "application/json").
		SetHeader("sec-fetch-mode", "cors").
		SetHeader("sec-fetch-site", "same-origin").
		SetHeader("sec-fetch-dest", "empty").
		SetQueryParam("feature", "training_allowed").
		SetQueryParam("value", "false").
		Patch(openAISettingsURL)

	if err != nil {
		slog.Warn("openai_privacy_request_error", "error", err.Error())
		return PrivacyModeFailed
	}

	if resp.StatusCode == 403 || resp.StatusCode == 503 {
		body := resp.String()
		if strings.Contains(body, "cloudflare") || strings.Contains(body, "cf-") || strings.Contains(body, "Just a moment") {
			slog.Warn("openai_privacy_cf_blocked", "status", resp.StatusCode)
			return PrivacyModeCFBlocked
		}
	}

	if !resp.IsSuccessState() {
		slog.Warn("openai_privacy_failed", "status", resp.StatusCode, "body", truncate(resp.String(), 200))
		return PrivacyModeFailed
	}

	slog.Info("openai_privacy_training_disabled")
	return PrivacyModeTrainingOff
}

// ChatGPTAccountInfo 从 chatgpt.com/backend-api/accounts/check 获取的账号信息
type ChatGPTAccountInfo struct {
	PlanType string
	Email    string
	// AccountID 是本条信息所属账号的标识（优先取 account.account_id，否则取 accounts
	// 的 map key）。accounts/check 是多账号/工作区端点，调用方需要据此判断拿到的
	// plan_type / expires_at 到底属于个人账号还是某个 workspace。
	AccountID             string
	SubscriptionExpiresAt string // entitlement.expires_at (RFC3339)
}

var (
	chatGPTAccountsCheckURL = "https://chatgpt.com/backend-api/accounts/check/v4-2023-04-27"
	chatGPTSubscriptionsURL = "https://chatgpt.com/backend-api/subscriptions"
)

// fetchChatGPTAccountInfo calls ChatGPT backend-api to get account info (plan_type, etc.).
// Used as fallback when id_token doesn't contain these fields (e.g., Mobile RT).
// orgID is used to match the correct account when multiple accounts exist (e.g., personal + team).
// Returns nil on any failure (best-effort, non-blocking).
func fetchChatGPTAccountInfo(ctx context.Context, clientFactory PrivacyClientFactory, accessToken, proxyURL, orgID string) *ChatGPTAccountInfo {
	if accessToken == "" || clientFactory == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client, err := clientFactory(proxyURL)
	if err != nil {
		slog.Debug("chatgpt_account_check_client_error", "error", err.Error())
		return nil
	}

	var result map[string]any
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetHeader("Origin", "https://chatgpt.com").
		SetHeader("Referer", "https://chatgpt.com/").
		SetHeader("Accept", "application/json").
		SetSuccessResult(&result).
		Get(chatGPTAccountsCheckURL)

	if err != nil {
		slog.Debug("chatgpt_account_check_request_error", "error", err.Error())
		return nil
	}

	if !resp.IsSuccessState() {
		slog.Debug("chatgpt_account_check_failed", "status", resp.StatusCode, "body", truncate(resp.String(), 200))
		return nil
	}

	info := &ChatGPTAccountInfo{}

	accounts, ok := result["accounts"].(map[string]any)
	if !ok {
		slog.Debug("chatgpt_account_check_no_accounts", "body", truncate(resp.String(), 300))
		return nil
	}

	// 在多账号里挑最能代表用户订阅的计划：个人订阅(pro/plus/team)优先于
	// business 工作区，避免"曾加入过 business 计划"污染计划展示。
	// 选号时跳过已停用/过期的账号（isUsableChatGPTAccountCandidate）。
	// AccountID 随选中的那条一起回填（上游新增）：调用方据此判断拿到的
	// plan_type / expires_at 属于 token 自己的个人账号还是某个 workspace。
	info.PlanType, info.SubscriptionExpiresAt, info.AccountID = selectChatGPTAccount(accounts, orgID)

	if info.PlanType == "" {
		slog.Debug("chatgpt_account_check_no_plan_type", "body", truncate(resp.String(), 300))
		return nil
	}

	slog.Info("chatgpt_account_check_success", "plan_type", info.PlanType, "subscription_expires_at", info.SubscriptionExpiresAt, "org_id", orgID)
	return info
}

// fetchChatGPTSubscriptionExpiresAt reads the lightweight subscription endpoint used by
// ChatGPT/Codex clients. Some Plus accounts no longer expose entitlement.expires_at in
// accounts/check, but this endpoint still returns active_until.
func fetchChatGPTSubscriptionExpiresAt(ctx context.Context, clientFactory PrivacyClientFactory, accessToken, proxyURL, accountID string) string {
	accountID = strings.TrimSpace(accountID)
	if accessToken == "" || accountID == "" || clientFactory == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client, err := clientFactory(proxyURL)
	if err != nil {
		slog.Debug("chatgpt_subscription_client_error", "error", err.Error())
		return ""
	}

	var result struct {
		PlanType    string `json:"plan_type"`
		ActiveUntil string `json:"active_until"`
		WillRenew   bool   `json:"will_renew"`
		ID          string `json:"id"`
	}
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetHeader("Origin", "https://chatgpt.com").
		SetHeader("Referer", "https://chatgpt.com/").
		SetHeader("Accept", "application/json").
		SetSuccessResult(&result).
		SetQueryParam("account_id", accountID).
		Get(chatGPTSubscriptionsURL)
	if err != nil {
		slog.Debug("chatgpt_subscription_request_error", "error", err.Error())
		return ""
	}
	if !resp.IsSuccessState() {
		slog.Debug("chatgpt_subscription_failed", "status", resp.StatusCode, "body", truncate(resp.String(), 200))
		return ""
	}

	activeUntil := strings.TrimSpace(result.ActiveUntil)
	if activeUntil == "" {
		slog.Debug("chatgpt_subscription_no_active_until", "plan_type", result.PlanType, "has_subscription_id", strings.TrimSpace(result.ID) != "", "will_renew", result.WillRenew)
		return ""
	}
	if _, err := time.Parse(time.RFC3339, activeUntil); err != nil {
		slog.Debug("chatgpt_subscription_bad_active_until", "active_until", activeUntil, "error", err.Error())
		return ""
	}

	slog.Info("chatgpt_subscription_success", "plan_type", result.PlanType, "subscription_expires_at", activeUntil, "account_id", accountID)
	return activeUntil
}

// selectChatGPTAccount 从 accounts/check 的多账号里挑最能代表用户订阅的计划，
// 并回传选中那条的账号标识（accountID，供调用方判断个人账号还是 workspace）。
// 选号优先级：计划类型分值(个人订阅 > business/未知 > free) > orgID(poid)命中 > is_default。
// 个人订阅(pro/plus/team)优先于 business 工作区，避免"曾加入 business"污染计划展示；
// 个人订阅之间则按 orgID 命中(token 实际作用的 org)决定，保持原有语义。
func selectChatGPTAccount(accounts map[string]any, orgID string) (planType, expiresAt, accountID string) {
	type candidate struct {
		id        string
		accountID string
		planType  string
		expiresAt string
		priority  int
		orgMatch  bool
		isDefault bool
	}
	var best *candidate
	now := time.Now()
	for id, acctRaw := range accounts {
		acct, ok := acctRaw.(map[string]any)
		if !ok {
			continue
		}
		// 跳过已停用/过期的账号，避免"曾加入过"污染计划展示。
		if !isUsableChatGPTAccountCandidate(acct, now) {
			continue
		}
		pt := extractPlanType(acct)
		if pt == "" {
			continue
		}
		cur := candidate{
			id: id,
			// orgMatch 仍按 accounts 的 map key(poid) 判定；accountID 另取
			// account.account_id（缺失时回落 map key），两者语义不同不可互换。
			accountID: chatGPTAccountObjectID(acct, id),
			planType:  pt,
			expiresAt: extractEntitlementExpiresAt(acct),
			priority:  planTypePriority(pt),
			orgMatch:  orgID != "" && id == orgID,
		}
		if account, ok := acct["account"].(map[string]any); ok {
			cur.isDefault, _ = account["is_default"].(bool)
		}
		better := best == nil
		if best != nil {
			switch {
			case cur.priority != best.priority:
				better = cur.priority > best.priority
			case cur.orgMatch != best.orgMatch:
				better = cur.orgMatch
			case cur.isDefault != best.isDefault:
				better = cur.isDefault
			default:
				// 全部相等时用账号 id 兜底，保证选号在 Go map 随机遍历下确定。
				better = cur.id < best.id
			}
		}
		if better {
			c := cur
			best = &c
		}
	}
	if best == nil {
		return "", "", ""
	}
	return best.planType, best.expiresAt, best.accountID
}

// planTypePriority 给计划类型打分用于多账号选号：
// 个人订阅(pro/plus/team)=3 > business/企业/未知=2 > free=1 > 空=0。
func planTypePriority(planType string) int {
	switch {
	case strings.TrimSpace(planType) == "":
		return 0
	case isConsumerPlanType(planType):
		return 3
	case strings.EqualFold(strings.TrimSpace(planType), "free"):
		return 1
	default: // business / enterprise / 未知
		return 2
	}
}

// isBusinessPlanType 判断是否为企业/工作区类计划(非个人订阅)，
// 例如 self_serve_business_usage_based、business、enterprise。
func isBusinessPlanType(planType string) bool {
	p := strings.ToLower(strings.TrimSpace(planType))
	return strings.Contains(p, "business") || strings.Contains(p, "enterprise")
}

// isConsumerPlanType 判断是否为个人订阅计划(pro/plus/team)。
func isConsumerPlanType(planType string) bool {
	if isBusinessPlanType(planType) {
		return false
	}
	p := strings.ToLower(strings.TrimSpace(planType))
	return strings.Contains(p, "pro") || strings.Contains(p, "plus") || strings.Contains(p, "team")
}

// chatGPTAccountObjectID 取单个 account 对象的账号标识。
// accounts 的 map key 有时是 "default" 这类别名，所以优先读 account.account_id。
func chatGPTAccountObjectID(acct map[string]any, fallbackID string) string {
	if account, ok := acct["account"].(map[string]any); ok {
		if id, ok := account["account_id"].(string); ok && strings.TrimSpace(id) != "" {
			return strings.TrimSpace(id)
		}
	}
	return strings.TrimSpace(fallbackID)
}

// extractPlanType 从单个 account 对象中提取 plan_type
func extractPlanType(acct map[string]any) string {
	if account, ok := acct["account"].(map[string]any); ok {
		if planType, ok := account["plan_type"].(string); ok && planType != "" {
			return planType
		}
	}
	if entitlement, ok := acct["entitlement"].(map[string]any); ok {
		if subPlan, ok := entitlement["subscription_plan"].(string); ok && subPlan != "" {
			return subPlan
		}
	}
	return ""
}

func isUsableChatGPTAccountCandidate(acct map[string]any, now time.Time) bool {
	if acct == nil || hasChatGPTAccountDeactivatedMarker(acct) {
		return false
	}
	if account, ok := acct["account"].(map[string]any); ok && hasChatGPTAccountDeactivatedMarker(account) {
		return false
	}

	expiresAt := extractEntitlementExpiresAt(acct)
	if expiresAt == "" {
		return true
	}
	expiry, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return true
	}
	return expiry.After(now)
}

func hasChatGPTAccountDeactivatedMarker(obj map[string]any) bool {
	for _, key := range []string{"deactivated", "is_deactivated", "disabled", "is_disabled"} {
		if value, ok := obj[key].(bool); ok && value {
			return true
		}
	}
	for _, key := range []string{"deactivated_at", "disabled_at", "deleted_at"} {
		if value, ok := obj[key].(string); ok && strings.TrimSpace(value) != "" {
			return true
		}
	}
	for _, key := range []string{"status", "state"} {
		value, _ := obj[key].(string)
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "deactivated", "disabled", "deleted", "inactive", "suspended":
			return true
		}
	}
	return false
}

// extractEntitlementExpiresAt 从 entitlement 中提取 expires_at。
// 预期为 RFC3339 字符串格式，如 "2026-05-02T20:32:12+00:00"。
func extractEntitlementExpiresAt(acct map[string]any) string {
	entitlement, ok := acct["entitlement"].(map[string]any)
	if !ok {
		return ""
	}
	ea, _ := entitlement["expires_at"].(string)
	return ea
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + fmt.Sprintf("...(%d more)", len(s)-n)
}
