package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleMessage routes incoming commands based on the sender's identity.
func (m *TelegramBotManager) handleMessage(inst *botInstance, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	cmd := msg.Command()
	args := msg.CommandArguments()

	// 1. Check if sender is the reseller (admin)
	resellerChatID := m.getResellerChatID(inst.resellerID)
	if resellerChatID != 0 && chatID == resellerChatID {
		m.handleAdminCommand(inst, msg, cmd, args)
		return
	}

	// 2. Check if sender is a bound user (via api_keys.tg_chat_id)
	apiKey := m.findKeyByChatID(inst.resellerID, chatID)
	if apiKey != nil {
		m.handleUserCommand(inst, msg, cmd, args, apiKey)
		return
	}

	// 3. Unbound user — only allow /start, /bind, /bindkey, /help
	m.handleUnboundCommand(inst, msg, cmd, args)
}

// --- Admin (Reseller) Commands ---

func (m *TelegramBotManager) handleAdminCommand(inst *botInstance, msg *tgbotapi.Message, cmd, args string) {
	chatID := msg.Chat.ID
	ctx := context.Background()

	switch cmd {
	case "admin":
		m.cmdAdminHelp(inst, chatID)
	case "unbind":
		m.cmdAdminUnbind(inst, chatID)
	case "keys":
		if !featureEnabled(inst, "admin_keys") {
			sendReply(inst.bot, chatID, "⚠️ 此功能已关闭")
			return
		}
		m.cmdAdminKeys(inst, chatID, ctx)
	case "key":
		if !featureEnabled(inst, "admin_keys") {
			sendReply(inst.bot, chatID, "⚠️ 此功能已关闭")
			return
		}
		m.cmdAdminKeyDetail(inst, chatID, ctx, args)
	case "create":
		if !featureEnabled(inst, "admin_keys") {
			sendReply(inst.bot, chatID, "⚠️ 此功能已关闭")
			return
		}
		m.cmdAdminCreateKey(inst, chatID, ctx, args)
	case "delete":
		if !featureEnabled(inst, "admin_keys") {
			sendReply(inst.bot, chatID, "⚠️ 此功能已关闭")
			return
		}
		m.cmdAdminDeleteKey(inst, chatID, ctx, args)
	case "resetquota":
		if !featureEnabled(inst, "admin_keys") {
			sendReply(inst.bot, chatID, "⚠️ 此功能已关闭")
			return
		}
		m.cmdAdminResetQuota(inst, chatID, ctx, args)
	case "stats":
		if !featureEnabled(inst, "admin_stats") {
			sendReply(inst.bot, chatID, "⚠️ 此功能已关闭")
			return
		}
		m.cmdAdminStats(inst, chatID, ctx)
	case "help":
		m.cmdAdminHelp(inst, chatID)
	default:
		sendReply(inst.bot, chatID, "❓ 未知命令，输入 /admin 查看可用命令")
	}
}

func (m *TelegramBotManager) cmdAdminHelp(inst *botInstance, chatID int64) {
	var sb strings.Builder
	sb.WriteString("🔧 *管理员命令*\n\n")
	sb.WriteString("/unbind - 解除管理员绑定\n")
	if featureEnabled(inst, "admin_keys") {
		sb.WriteString("/keys - 列出所有密钥\n")
		sb.WriteString("/key <ID> - 查看密钥详情\n")
		sb.WriteString("/create <名称> - 创建密钥\n")
		sb.WriteString("/delete <ID> - 删除密钥\n")
		sb.WriteString("/resetquota <ID> - 重置配额\n")
	}
	if featureEnabled(inst, "admin_stats") {
		sb.WriteString("/stats - 仪表盘统计\n")
	}
	sb.WriteString("/help - 帮助菜单")
	sendReply(inst.bot, chatID, sb.String())
}

func (m *TelegramBotManager) cmdAdminUnbind(inst *botInstance, chatID int64) {
	ctx := context.Background()
	if err := m.settingRepo.Set(ctx, inst.resellerID, ResellerSettingTgChatID, ""); err != nil {
		sendReply(inst.bot, chatID, "❌ 解绑失败: "+err.Error())
		return
	}
	sendReply(inst.bot, chatID, "✅ 已解除管理员绑定")
}

func (m *TelegramBotManager) cmdAdminKeys(inst *botInstance, chatID int64, ctx context.Context) {
	keys, _, err := m.resellerSvc.ListKeys(ctx, inst.resellerID, 1, 50)
	if err != nil {
		sendReply(inst.bot, chatID, "❌ 获取密钥列表失败: "+err.Error())
		return
	}

	if len(keys) == 0 {
		sendReply(inst.bot, chatID, "📭 暂无密钥")
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔑 *密钥列表* (%d 个)\n\n", len(keys)))
	for _, k := range keys {
		status := "✅"
		if k.Status != StatusActive {
			status = "⛔"
		}
		quota := "不限"
		if k.Quota > 0 {
			quota = fmt.Sprintf("$%.2f/$%.2f", k.QuotaUsed, k.Quota)
		}
		sb.WriteString(fmt.Sprintf("%s `#%d` %s — %s\n", status, k.ID, k.Name, quota))
	}
	sb.WriteString("\n使用 /key <ID> 查看详情")
	sendReply(inst.bot, chatID, sb.String())
}

func (m *TelegramBotManager) cmdAdminKeyDetail(inst *botInstance, chatID int64, ctx context.Context, args string) {
	keyID, err := strconv.ParseInt(strings.TrimSpace(args), 10, 64)
	if err != nil || keyID <= 0 {
		sendReply(inst.bot, chatID, "用法: /key <ID>")
		return
	}

	key, err := m.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		sendReply(inst.bot, chatID, "❌ 密钥未找到")
		return
	}
	if key.UserID != inst.resellerID {
		sendReply(inst.bot, chatID, "❌ 此密钥不属于您")
		return
	}

	sendReply(inst.bot, chatID, formatKeyDetail(key))
}

func (m *TelegramBotManager) cmdAdminCreateKey(inst *botInstance, chatID int64, ctx context.Context, args string) {
	name := strings.TrimSpace(args)
	if name == "" {
		name = fmt.Sprintf("tg-%d", time.Now().Unix())
	}

	key, err := m.resellerSvc.CreateKey(ctx, inst.resellerID, &CreateResellerKeyInput{Name: name})
	if err != nil {
		sendReply(inst.bot, chatID, "❌ 创建失败: "+err.Error())
		return
	}

	sendReply(inst.bot, chatID, fmt.Sprintf("✅ 密钥已创建\n\n名称: %s\nID: `#%d`\n密钥: `%s`", key.Name, key.ID, key.Key))
}

func (m *TelegramBotManager) cmdAdminDeleteKey(inst *botInstance, chatID int64, ctx context.Context, args string) {
	keyID, err := strconv.ParseInt(strings.TrimSpace(args), 10, 64)
	if err != nil || keyID <= 0 {
		sendReply(inst.bot, chatID, "用法: /delete <ID>")
		return
	}

	if err := m.resellerSvc.DeleteKey(ctx, inst.resellerID, keyID); err != nil {
		sendReply(inst.bot, chatID, "❌ 删除失败: "+err.Error())
		return
	}

	sendReply(inst.bot, chatID, fmt.Sprintf("✅ 密钥 #%d 已删除", keyID))
}

func (m *TelegramBotManager) cmdAdminResetQuota(inst *botInstance, chatID int64, ctx context.Context, args string) {
	keyID, err := strconv.ParseInt(strings.TrimSpace(args), 10, 64)
	if err != nil || keyID <= 0 {
		sendReply(inst.bot, chatID, "用法: /resetquota <ID>")
		return
	}

	key, err := m.resellerSvc.ResetKeyQuota(ctx, inst.resellerID, keyID)
	if err != nil {
		sendReply(inst.bot, chatID, "❌ 重置失败: "+err.Error())
		return
	}

	sendReply(inst.bot, chatID, fmt.Sprintf("✅ 密钥 #%d 配额已重置\n当前使用: $%.2f", key.ID, key.QuotaUsed))
}

func (m *TelegramBotManager) cmdAdminStats(inst *botInstance, chatID int64, ctx context.Context) {
	stats, err := m.resellerSvc.GetDashboardStats(ctx, inst.resellerID)
	if err != nil {
		sendReply(inst.bot, chatID, "❌ 获取统计失败: "+err.Error())
		return
	}

	text := fmt.Sprintf(
		"📊 *仪表盘统计*\n\n"+
			"💰 余额: $%.2f\n"+
			"🌐 域名: %d（已验证 %d）\n"+
			"📦 套餐: %d\n"+
			"🔑 密钥: %d（活跃 %d）\n"+
			"📈 已用配额: $%.2f",
		stats.MyBalance,
		stats.DomainCount, stats.VerifiedDomains,
		stats.GroupCount,
		stats.KeyCount, stats.ActiveKeyCount,
		stats.TotalQuotaUsed,
	)
	sendReply(inst.bot, chatID, text)
}

// --- User (Key Holder) Commands ---

func (m *TelegramBotManager) handleUserCommand(inst *botInstance, msg *tgbotapi.Message, cmd, args string, key *APIKey) {
	chatID := msg.Chat.ID

	switch cmd {
	case "unbindkey":
		if !featureEnabled(inst, "user_query") {
			sendReply(inst.bot, chatID, "⚠️ 此功能已关闭")
			return
		}
		m.cmdUserUnbindKey(inst, chatID, key)
	case "mykey":
		if !featureEnabled(inst, "user_query") {
			sendReply(inst.bot, chatID, "⚠️ 此功能已关闭")
			return
		}
		sendReply(inst.bot, chatID, formatKeyStatus(key))
	case "usage":
		if !featureEnabled(inst, "user_query") {
			sendReply(inst.bot, chatID, "⚠️ 此功能已关闭")
			return
		}
		sendReply(inst.bot, chatID, formatKeyUsage(key))
	case "help":
		m.cmdUserHelp(inst, chatID)
	case "start":
		sendReply(inst.bot, chatID, fmt.Sprintf("👋 您已绑定密钥 `%s`\n\n使用 /mykey 查看状态，/help 查看帮助", maskKey(key.Key)))
	default:
		sendReply(inst.bot, chatID, "❓ 未知命令，输入 /help 查看可用命令")
	}
}

func (m *TelegramBotManager) cmdUserHelp(inst *botInstance, chatID int64) {
	var sb strings.Builder
	sb.WriteString("📋 *用户命令*\n\n")
	if featureEnabled(inst, "user_query") {
		sb.WriteString("/mykey - 查看密钥状态\n")
		sb.WriteString("/usage - 查看用量详情\n")
		sb.WriteString("/unbindkey - 解除绑定\n")
	}
	sb.WriteString("/help - 帮助菜单")
	sendReply(inst.bot, chatID, sb.String())
}

func (m *TelegramBotManager) cmdUserUnbindKey(inst *botInstance, chatID int64, key *APIKey) {
	ctx := context.Background()
	key.TgChatID = nil
	if err := m.apiKeyRepo.Update(ctx, key); err != nil {
		sendReply(inst.bot, chatID, "❌ 解绑失败: "+err.Error())
		return
	}
	sendReply(inst.bot, chatID, "✅ 已解除密钥绑定")
}

// --- Unbound Commands ---

func (m *TelegramBotManager) handleUnboundCommand(inst *botInstance, msg *tgbotapi.Message, cmd, args string) {
	chatID := msg.Chat.ID

	switch cmd {
	case "start":
		sendReply(inst.bot, chatID, "👋 欢迎！\n\n"+
			"如果您是分销商管理员，请使用 /bind <绑定码> 绑定\n"+
			"如果您是用户，请使用 /bindkey <API密钥> 绑定您的密钥")
	case "bind":
		m.cmdBind(inst, chatID, args)
	case "bindkey":
		if !featureEnabled(inst, "user_query") {
			sendReply(inst.bot, chatID, "⚠️ 此功能已关闭")
			return
		}
		m.cmdBindKey(inst, chatID, args)
	case "help":
		sendReply(inst.bot, chatID, "📋 *可用命令*\n\n"+
			"/bind <绑定码> - 分销商管理员绑定\n"+
			"/bindkey <API密钥> - 用户绑定密钥\n"+
			"/help - 帮助菜单")
	default:
		sendReply(inst.bot, chatID, "❓ 请先绑定。分销商使用 /bind，用户使用 /bindkey")
	}
}

// cmdBind handles reseller admin binding via a bind code.
func (m *TelegramBotManager) cmdBind(inst *botInstance, chatID int64, args string) {
	code := strings.TrimSpace(args)
	if code == "" {
		sendReply(inst.bot, chatID, "用法: /bind <绑定码>\n\n请在分销商设置页面生成绑定码")
		return
	}

	ctx := context.Background()
	storedCode, err := m.settingRepo.Get(ctx, inst.resellerID, ResellerSettingTgBindCode)
	if err != nil || storedCode == "" {
		sendReply(inst.bot, chatID, "❌ 绑定码无效或已过期，请重新生成")
		return
	}

	if storedCode != code {
		sendReply(inst.bot, chatID, "❌ 绑定码不匹配")
		return
	}

	// Bind: save chat ID, clear bind code
	if err := m.settingRepo.Set(ctx, inst.resellerID, ResellerSettingTgChatID, fmt.Sprintf("%d", chatID)); err != nil {
		sendReply(inst.bot, chatID, "❌ 绑定失败: "+err.Error())
		return
	}
	_ = m.settingRepo.Set(ctx, inst.resellerID, ResellerSettingTgBindCode, "")

	sendReply(inst.bot, chatID, "✅ 管理员绑定成功！\n\n使用 /admin 查看可用命令")
}

// cmdBindKey handles user binding via API key.
func (m *TelegramBotManager) cmdBindKey(inst *botInstance, chatID int64, args string) {
	keyStr := strings.TrimSpace(args)
	if keyStr == "" {
		sendReply(inst.bot, chatID, "用法: /bindkey <API密钥>")
		return
	}

	ctx := context.Background()

	// Look up the API key
	key, err := m.apiKeyRepo.GetByKey(ctx, keyStr)
	if err != nil {
		sendReply(inst.bot, chatID, "❌ 密钥无效")
		return
	}

	// Verify the key belongs to this reseller
	if key.UserID != inst.resellerID {
		sendReply(inst.bot, chatID, "❌ 此密钥不属于当前分销商")
		return
	}

	// Check if already bound to another chat
	if key.TgChatID != nil && *key.TgChatID != chatID {
		sendReply(inst.bot, chatID, "❌ 此密钥已被其他用户绑定")
		return
	}

	// Bind
	key.TgChatID = &chatID
	if err := m.apiKeyRepo.Update(ctx, key); err != nil {
		sendReply(inst.bot, chatID, "❌ 绑定失败: "+err.Error())
		return
	}

	sendReply(inst.bot, chatID, fmt.Sprintf("✅ 密钥绑定成功！\n\n密钥: `%s`\n\n使用 /mykey 查看状态", maskKey(key.Key)))
}

// findKeyByChatID looks up an API key bound to the given chat ID for this reseller.
func (m *TelegramBotManager) findKeyByChatID(resellerID int64, chatID int64) *APIKey {
	ctx := context.Background()
	key, err := m.apiKeyRepo.FindByTgChatID(ctx, resellerID, chatID)
	if err != nil {
		log.Printf("[TelegramBot] Error finding key by chat ID %d: %v", chatID, err)
		return nil
	}
	return key
}

// --- Formatting helpers ---

func formatKeyDetail(k *APIKey) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔑 *密钥详情* #%d\n\n", k.ID))
	sb.WriteString(fmt.Sprintf("名称: %s\n", k.Name))
	sb.WriteString(fmt.Sprintf("密钥: `%s`\n", maskKey(k.Key)))
	sb.WriteString(fmt.Sprintf("状态: %s\n", statusEmoji(k.Status)))
	if k.Notes != "" {
		sb.WriteString(fmt.Sprintf("备注: %s\n", k.Notes))
	}
	if k.Quota > 0 {
		sb.WriteString(fmt.Sprintf("配额: $%.2f / $%.2f (%.1f%%)\n", k.QuotaUsed, k.Quota, k.QuotaUsed/k.Quota*100))
	} else {
		sb.WriteString("配额: 不限\n")
	}
	if k.ExpiresAt != nil {
		sb.WriteString(fmt.Sprintf("过期: %s\n", k.ExpiresAt.Format("2006-01-02 15:04")))
	} else {
		sb.WriteString("过期: 永不\n")
	}
	if k.Group != nil {
		sb.WriteString(fmt.Sprintf("分组: %s\n", k.Group.Name))
	}
	if k.TgChatID != nil {
		sb.WriteString(fmt.Sprintf("TG绑定: Chat %d\n", *k.TgChatID))
	}
	sb.WriteString(fmt.Sprintf("创建: %s", k.CreatedAt.Format("2006-01-02 15:04")))
	return sb.String()
}

func formatKeyStatus(k *APIKey) string {
	var sb strings.Builder
	sb.WriteString("🔑 *我的密钥*\n\n")
	sb.WriteString(fmt.Sprintf("名称: %s\n", k.Name))
	sb.WriteString(fmt.Sprintf("状态: %s\n", statusEmoji(k.Status)))
	if k.Quota > 0 {
		remaining := k.Quota - k.QuotaUsed
		if remaining < 0 {
			remaining = 0
		}
		sb.WriteString(fmt.Sprintf("配额: $%.2f / $%.2f\n", k.QuotaUsed, k.Quota))
		sb.WriteString(fmt.Sprintf("剩余: $%.2f (%.1f%%)\n", remaining, remaining/k.Quota*100))
	} else {
		sb.WriteString("配额: 不限\n")
	}
	if k.ExpiresAt != nil {
		days := k.GetDaysUntilExpiry()
		if days <= 0 {
			sb.WriteString("过期: ⚠️ 已过期\n")
		} else {
			sb.WriteString(fmt.Sprintf("过期: %s（剩余 %d 天）\n", k.ExpiresAt.Format("2006-01-02"), days))
		}
	} else {
		sb.WriteString("过期: 永不\n")
	}
	if k.Group != nil {
		sb.WriteString(fmt.Sprintf("套餐: %s", k.Group.Name))
	}
	return sb.String()
}

func formatKeyUsage(k *APIKey) string {
	var sb strings.Builder
	sb.WriteString("📊 *用量详情*\n\n")
	sb.WriteString(fmt.Sprintf("密钥: %s\n", k.Name))
	if k.Quota > 0 {
		pct := k.QuotaUsed / k.Quota * 100
		sb.WriteString(fmt.Sprintf("已用: $%.4f / $%.2f (%.1f%%)\n", k.QuotaUsed, k.Quota, pct))
		remaining := k.Quota - k.QuotaUsed
		if remaining < 0 {
			remaining = 0
		}
		sb.WriteString(fmt.Sprintf("剩余: $%.4f\n", remaining))
	} else {
		sb.WriteString(fmt.Sprintf("已用: $%.4f\n", k.QuotaUsed))
		sb.WriteString("配额: 不限\n")
	}
	return sb.String()
}

func maskKey(key string) string {
	if len(key) <= 12 {
		return key
	}
	return key[:8] + "..." + key[len(key)-4:]
}

func statusEmoji(status string) string {
	switch status {
	case StatusActive:
		return "✅ 正常"
	case StatusDisabled:
		return "⛔ 已禁用"
	default:
		return "❓ " + status
	}
}

// ListKeysForNotification lists all API keys for a reseller, used by the notification service.
func (m *TelegramBotManager) ListKeysForNotification(ctx context.Context, resellerID int64) ([]APIKey, error) {
	keys, _, err := m.resellerSvc.ListKeys(ctx, resellerID, 1, 10000)
	return keys, err
}

// GetResellerBalance returns the reseller's current balance.
func (m *TelegramBotManager) GetResellerBalance(ctx context.Context, resellerID int64) (float64, error) {
	stats, err := m.resellerSvc.GetDashboardStats(ctx, resellerID)
	if err != nil {
		return 0, err
	}
	return stats.MyBalance, nil
}

// IsFeatureEnabled checks if a feature is enabled for a given reseller's bot.
func (m *TelegramBotManager) IsFeatureEnabled(resellerID int64, feature string) bool {
	m.mu.RLock()
	inst := m.bots[resellerID]
	m.mu.RUnlock()
	if inst == nil {
		return false
	}
	return featureEnabled(inst, feature)
}

// ListRunningBotResellerIDs returns the IDs of all resellers with running bots.
func (m *TelegramBotManager) ListRunningBotResellerIDs() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]int64, 0, len(m.bots))
	for id := range m.bots {
		ids = append(ids, id)
	}
	return ids
}
