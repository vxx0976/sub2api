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

	log.Printf("[TelegramBot] reseller=%d chat=%d cmd=/%s args=%q", inst.resellerID, chatID, cmd, args)

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
		m.cmdAdminKeys(inst, chatID, ctx)
	case "key":
		m.cmdAdminKeyDetail(inst, chatID, ctx, args)
	case "create":
		m.cmdAdminCreateKey(inst, chatID, ctx, args)
	case "delete":
		m.cmdAdminDeleteKey(inst, chatID, ctx, args)
	case "resetquota":
		m.cmdAdminResetQuota(inst, chatID, ctx, args)
	case "stats":
		m.cmdAdminStats(inst, chatID, ctx)
	case "help":
		m.cmdAdminHelp(inst, chatID)
	default:
		sendReply(inst.bot, chatID, "❓ 未知命令，输入 /admin 查看可用命令")
	}
}

func (m *TelegramBotManager) cmdAdminHelp(inst *botInstance, chatID int64) {
	text := "🔧 *管理员命令*\n\n" +
		"/unbind - 解除管理员绑定\n" +
		"/keys - 列出所有密钥\n" +
		"/key <ID> - 查看密钥详情\n" +
		"/create <名称> - 创建密钥\n" +
		"/delete <ID> - 删除密钥\n" +
		"/resetquota <ID> - 重置配额\n" +
		"/stats - 仪表盘统计\n" +
		"/help - 帮助菜单"
	sendReply(inst.bot, chatID, text)
}

func (m *TelegramBotManager) cmdAdminUnbind(inst *botInstance, chatID int64) {
	ctx := context.Background()
	if err := m.settingRepo.Set(ctx, inst.resellerID, ResellerSettingTgChatID, ""); err != nil {
		sendReply(inst.bot, chatID, "❌ 解绑失败: "+err.Error())
		return
	}
	sendReply(inst.bot, chatID, "✅ 已解除管理员绑定")
}

const keysPerPage = 8

func (m *TelegramBotManager) cmdAdminKeys(inst *botInstance, chatID int64, ctx context.Context) {
	keys, _, err := m.resellerSvc.ListKeys(ctx, inst.resellerID, 1, 500)
	if err != nil {
		sendReply(inst.bot, chatID, "❌ 获取密钥列表失败: "+err.Error())
		return
	}

	if len(keys) == 0 {
		sendReply(inst.bot, chatID, "📭 暂无密钥\n\n使用 /create <名称> 创建密钥")
		return
	}

	text, markup := buildKeysListMessage(keys, 1)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = markup
	if _, err := inst.bot.Send(msg); err != nil {
		log.Printf("[TelegramBot] Failed to send keys list: %v", err)
	}
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

	text, markup := buildKeyDetailMessage(key, 1)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = markup
	if _, err := inst.bot.Send(msg); err != nil {
		log.Printf("[TelegramBot] Failed to send key detail: %v", err)
	}
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
		m.cmdUserUnbindKey(inst, chatID, key)
	case "mykey":
		sendReply(inst.bot, chatID, formatKeyStatus(key))
	case "usage":
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
	text := "📋 *用户命令*\n\n" +
		"/mykey - 查看密钥状态\n" +
		"/usage - 查看用量详情\n" +
		"/unbindkey - 解除绑定\n" +
		"/help - 帮助菜单"
	sendReply(inst.bot, chatID, text)
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
			"如果您是商户管理员，请使用 /bind <绑定码> 绑定\n"+
			"如果您是用户，请使用 /bindkey <API密钥> 绑定您的密钥")
	case "bind":
		m.cmdBind(inst, chatID, args)
	case "bindkey":
		m.cmdBindKey(inst, chatID, args)
	case "help":
		sendReply(inst.bot, chatID, "📋 *可用命令*\n\n"+
			"/bind <绑定码> - 商户管理员绑定\n"+
			"/bindkey <API密钥> - 用户绑定密钥\n"+
			"/help - 帮助菜单")
	default:
		sendReply(inst.bot, chatID, "❓ 请先绑定。商户使用 /bind，用户使用 /bindkey")
	}
}

// cmdBind handles reseller admin binding via a bind code.
func (m *TelegramBotManager) cmdBind(inst *botInstance, chatID int64, args string) {
	code := strings.TrimSpace(args)
	if code == "" {
		sendReply(inst.bot, chatID, "用法: /bind <绑定码>\n\n请在商户设置页面生成绑定码")
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
		sendReply(inst.bot, chatID, "❌ 此密钥不属于当前商户")
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

// --- Inline Keyboard Builders ---

func buildKeysListMessage(keys []APIKey, page int) (string, tgbotapi.InlineKeyboardMarkup) {
	total := len(keys)
	totalPages := (total + keysPerPage - 1) / keysPerPage
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * keysPerPage
	end := start + keysPerPage
	if end > total {
		end = total
	}
	pageKeys := keys[start:end]

	text := fmt.Sprintf("🔑 *密钥列表* (共 %d 个)", total)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, k := range pageKeys {
		status := "✅"
		if k.Status != StatusActive {
			status = "⛔"
		}
		quota := "不限"
		if k.Quota > 0 {
			quota = fmt.Sprintf("$%.1f/$%.0f", k.QuotaUsed, k.Quota)
		}
		label := fmt.Sprintf("%s %s — %s", status, k.Name, quota)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("key:%d:%d", k.ID, page)),
		))
	}

	// Pagination row
	if totalPages > 1 {
		var navButtons []tgbotapi.InlineKeyboardButton
		if page > 1 {
			navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData("« 上一页", fmt.Sprintf("keys:%d", page-1)))
		}
		navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d/%d", page, totalPages), "noop"))
		if page < totalPages {
			navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData("下一页 »", fmt.Sprintf("keys:%d", page+1)))
		}
		rows = append(rows, navButtons)
	}

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return text, markup
}

func buildKeyDetailMessage(k *APIKey, page int) (string, tgbotapi.InlineKeyboardMarkup) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔑 *密钥详情* #%d\n\n", k.ID))
	sb.WriteString(fmt.Sprintf("名称: %s\n", escapeMarkdown(k.Name)))
	sb.WriteString(fmt.Sprintf("密钥: `%s`\n", maskKey(k.Key)))
	sb.WriteString(fmt.Sprintf("状态: %s\n", statusEmoji(k.Status)))
	if k.Notes != "" {
		sb.WriteString(fmt.Sprintf("备注: %s\n", escapeMarkdown(k.Notes)))
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
		sb.WriteString(fmt.Sprintf("分组: %s\n", escapeMarkdown(k.Group.Name)))
	}
	if k.TgChatID != nil {
		sb.WriteString(fmt.Sprintf("TG绑定: Chat %d\n", *k.TgChatID))
	}
	sb.WriteString(fmt.Sprintf("创建: %s", k.CreatedAt.Format("2006-01-02 15:04")))

	id := k.ID
	// Action buttons
	row1 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✏️ 名称", fmt.Sprintf("ke:%d:name", id)),
		tgbotapi.NewInlineKeyboardButtonData("📝 备注", fmt.Sprintf("ke:%d:notes", id)),
	)
	row2 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("💰 额度", fmt.Sprintf("ke:%d:quota", id)),
		tgbotapi.NewInlineKeyboardButtonData("📅 有效期", fmt.Sprintf("ke:%d:exp", id)),
	)

	toggleLabel := "⏸ 禁用"
	if k.Status != StatusActive {
		toggleLabel = "▶️ 启用"
	}
	row3 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔄 重置配额", fmt.Sprintf("kr:%d", id)),
		tgbotapi.NewInlineKeyboardButtonData(toggleLabel, fmt.Sprintf("kt:%d", id)),
	)
	row4 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🗑 删除", fmt.Sprintf("kd:%d", id)),
	)
	row5 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« 返回列表", fmt.Sprintf("back:keys:%d", page)),
	)

	markup := tgbotapi.NewInlineKeyboardMarkup(row1, row2, row3, row4, row5)
	return sb.String(), markup
}

// --- Callback Query Handler ---

func (m *TelegramBotManager) handleCallback(inst *botInstance, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	msgID := cq.Message.MessageID
	data := cq.Data

	// Answer callback to remove loading indicator
	callback := tgbotapi.NewCallback(cq.ID, "")
	if _, err := inst.bot.Request(callback); err != nil {
		log.Printf("[TelegramBot] Failed to answer callback: %v", err)
	}

	// Only admin can use key management callbacks
	resellerChatID := m.getResellerChatID(inst.resellerID)
	if resellerChatID == 0 || chatID != resellerChatID {
		return
	}

	// Clear any pending edit when a new button is pressed (unless it's noop)
	if data != "noop" {
		inst.editsMu.Lock()
		delete(inst.edits, chatID)
		inst.editsMu.Unlock()
	}

	ctx := context.Background()

	switch {
	case data == "noop":
		// Do nothing
	case strings.HasPrefix(data, "keys:"):
		// Paginated key list
		page, _ := strconv.Atoi(strings.TrimPrefix(data, "keys:"))
		m.cbKeysList(inst, chatID, msgID, ctx, page)

	case strings.HasPrefix(data, "key:"):
		// Key detail: "key:<id>:<page>"
		parts := strings.SplitN(strings.TrimPrefix(data, "key:"), ":", 2)
		keyID, _ := strconv.ParseInt(parts[0], 10, 64)
		page := 1
		if len(parts) > 1 {
			page, _ = strconv.Atoi(parts[1])
		}
		m.cbKeyDetail(inst, chatID, msgID, ctx, keyID, page)

	case strings.HasPrefix(data, "ke:"):
		// Edit key field: "ke:<id>:<field>"
		parts := strings.SplitN(strings.TrimPrefix(data, "ke:"), ":", 2)
		keyID, _ := strconv.ParseInt(parts[0], 10, 64)
		field := ""
		if len(parts) > 1 {
			field = parts[1]
		}
		m.cbKeyEdit(inst, chatID, msgID, ctx, keyID, field)

	case strings.HasPrefix(data, "kt:"):
		// Toggle key status
		keyID, _ := strconv.ParseInt(strings.TrimPrefix(data, "kt:"), 10, 64)
		m.cbKeyToggle(inst, chatID, msgID, ctx, keyID)

	case strings.HasPrefix(data, "kr:"):
		// Reset quota
		keyID, _ := strconv.ParseInt(strings.TrimPrefix(data, "kr:"), 10, 64)
		m.cbKeyResetQuota(inst, chatID, msgID, ctx, keyID)

	case strings.HasPrefix(data, "kd:"):
		// Delete: "kd:<id>" for confirm prompt, "kd:<id>:y" for actual delete
		rest := strings.TrimPrefix(data, "kd:")
		if strings.HasSuffix(rest, ":y") {
			keyID, _ := strconv.ParseInt(strings.TrimSuffix(rest, ":y"), 10, 64)
			m.cbKeyDeleteConfirm(inst, chatID, msgID, ctx, keyID)
		} else {
			keyID, _ := strconv.ParseInt(rest, 10, 64)
			m.cbKeyDeletePrompt(inst, chatID, msgID, ctx, keyID)
		}

	case strings.HasPrefix(data, "back:keys:"):
		page, _ := strconv.Atoi(strings.TrimPrefix(data, "back:keys:"))
		m.cbKeysList(inst, chatID, msgID, ctx, page)
	}
}

// cbKeysList edits the message to show the paginated key list.
func (m *TelegramBotManager) cbKeysList(inst *botInstance, chatID int64, msgID int, ctx context.Context, page int) {
	keys, _, err := m.resellerSvc.ListKeys(ctx, inst.resellerID, 1, 500)
	if err != nil {
		log.Printf("[TelegramBot] cbKeysList error: %v", err)
		return
	}

	if len(keys) == 0 {
		edit := tgbotapi.NewEditMessageText(chatID, msgID, "📭 暂无密钥")
		if _, err := inst.bot.Send(edit); err != nil {
			log.Printf("[TelegramBot] Failed to edit message: %v", err)
		}
		return
	}

	text, markup := buildKeysListMessage(keys, page)
	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, text, markup)
	edit.ParseMode = "Markdown"
	if _, err := inst.bot.Send(edit); err != nil {
		log.Printf("[TelegramBot] Failed to edit keys list: %v", err)
	}
}

// cbKeyDetail edits the message to show key detail with action buttons.
func (m *TelegramBotManager) cbKeyDetail(inst *botInstance, chatID int64, msgID int, ctx context.Context, keyID int64, page int) {
	key, err := m.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil || key.UserID != inst.resellerID {
		edit := tgbotapi.NewEditMessageText(chatID, msgID, "❌ 密钥未找到")
		inst.bot.Send(edit)
		return
	}

	text, markup := buildKeyDetailMessage(key, page)
	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, text, markup)
	edit.ParseMode = "Markdown"
	if _, err := inst.bot.Send(edit); err != nil {
		log.Printf("[TelegramBot] Failed to edit key detail: %v", err)
	}
}

// cbKeyEdit starts the edit flow for a key field.
func (m *TelegramBotManager) cbKeyEdit(inst *botInstance, chatID int64, msgID int, ctx context.Context, keyID int64, field string) {
	key, err := m.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil || key.UserID != inst.resellerID {
		return
	}

	var prompt string
	switch field {
	case "name":
		prompt = fmt.Sprintf("请输入密钥 #%d 的新名称（当前: %s）\n\n发送 /cancel 取消", keyID, key.Name)
	case "notes":
		current := key.Notes
		if current == "" {
			current = "无"
		}
		prompt = fmt.Sprintf("请输入密钥 #%d 的新备注（当前: %s）\n发送空格清除备注，/cancel 取消", keyID, current)
	case "quota":
		current := "不限"
		if key.Quota > 0 {
			current = fmt.Sprintf("$%.2f", key.Quota)
		}
		prompt = fmt.Sprintf("请输入密钥 #%d 的额度限制，单位美元（当前: %s）\n输入 0 表示不限，/cancel 取消", keyID, current)
	case "exp":
		current := "永不"
		if key.ExpiresAt != nil {
			current = key.ExpiresAt.Format("2006-01-02")
		}
		prompt = fmt.Sprintf("请输入密钥 #%d 的有效期（当前: %s）\n格式: 2025-12-31 或 30d（30天后）或 0（清除）\n/cancel 取消", keyID, current)
	default:
		return
	}

	// Send prompt message
	pmsg := tgbotapi.NewMessage(chatID, prompt)
	sent, err := inst.bot.Send(pmsg)
	if err != nil {
		log.Printf("[TelegramBot] Failed to send edit prompt: %v", err)
		return
	}

	// Set pending edit state
	inst.editsMu.Lock()
	inst.edits[chatID] = &pendingEdit{
		keyID:   keyID,
		field:   field,
		msgID:   sent.MessageID,
		detailM: msgID,
		listPg:  1,
	}
	inst.editsMu.Unlock()
}

// cbKeyToggle toggles key status between active and disabled.
func (m *TelegramBotManager) cbKeyToggle(inst *botInstance, chatID int64, msgID int, ctx context.Context, keyID int64) {
	key, err := m.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil || key.UserID != inst.resellerID {
		return
	}

	newStatus := StatusDisabled
	if key.Status != StatusActive {
		newStatus = StatusActive
	}

	_, err = m.resellerSvc.UpdateKey(ctx, inst.resellerID, keyID, &UpdateResellerKeyInput{
		Status: &newStatus,
	})
	if err != nil {
		answerText := "❌ 切换状态失败: " + err.Error()
		sendReply(inst.bot, chatID, answerText)
		return
	}

	// Refresh detail view
	m.cbKeyDetail(inst, chatID, msgID, ctx, keyID, 1)
}

// cbKeyResetQuota resets the key's quota usage.
func (m *TelegramBotManager) cbKeyResetQuota(inst *botInstance, chatID int64, msgID int, ctx context.Context, keyID int64) {
	_, err := m.resellerSvc.ResetKeyQuota(ctx, inst.resellerID, keyID)
	if err != nil {
		sendReply(inst.bot, chatID, "❌ 重置失败: "+err.Error())
		return
	}

	// Refresh detail view
	m.cbKeyDetail(inst, chatID, msgID, ctx, keyID, 1)
}

// cbKeyDeletePrompt shows a confirmation dialog for deleting a key.
func (m *TelegramBotManager) cbKeyDeletePrompt(inst *botInstance, chatID int64, msgID int, ctx context.Context, keyID int64) {
	key, err := m.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil || key.UserID != inst.resellerID {
		return
	}

	text := fmt.Sprintf("⚠️ *确认删除*\n\n确定要删除密钥 #%d (%s) 吗？\n此操作不可撤销！", keyID, escapeMarkdown(key.Name))
	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ 确认删除", fmt.Sprintf("kd:%d:y", keyID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", fmt.Sprintf("key:%d:1", keyID)),
		),
	)

	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, text, markup)
	edit.ParseMode = "Markdown"
	if _, err := inst.bot.Send(edit); err != nil {
		log.Printf("[TelegramBot] Failed to edit delete prompt: %v", err)
	}
}

// cbKeyDeleteConfirm performs the actual deletion.
func (m *TelegramBotManager) cbKeyDeleteConfirm(inst *botInstance, chatID int64, msgID int, ctx context.Context, keyID int64) {
	if err := m.resellerSvc.DeleteKey(ctx, inst.resellerID, keyID); err != nil {
		sendReply(inst.bot, chatID, "❌ 删除失败: "+err.Error())
		return
	}

	// Go back to key list
	m.cbKeysList(inst, chatID, msgID, ctx, 1)
}

// --- Text Input Handler (Edit Flow) ---

func (m *TelegramBotManager) handleTextInput(inst *botInstance, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)

	inst.editsMu.Lock()
	pe := inst.edits[chatID]
	inst.editsMu.Unlock()

	if pe == nil {
		// No pending edit — ignore
		return
	}

	// Handle /cancel
	if text == "/cancel" {
		inst.editsMu.Lock()
		delete(inst.edits, chatID)
		inst.editsMu.Unlock()
		// Delete prompt message
		del := tgbotapi.NewDeleteMessage(chatID, pe.msgID)
		inst.bot.Request(del)
		// Delete user's cancel message
		del2 := tgbotapi.NewDeleteMessage(chatID, msg.MessageID)
		inst.bot.Request(del2)
		return
	}

	ctx := context.Background()
	input := &UpdateResellerKeyInput{}

	switch pe.field {
	case "name":
		if text == "" || len(text) > 100 {
			sendReply(inst.bot, chatID, "❌ 名称不能为空，且长度不超过 100 字符")
			return
		}
		input.Name = &text

	case "notes":
		if len(text) > 500 {
			sendReply(inst.bot, chatID, "❌ 备注长度不超过 500 字符")
			return
		}
		// A single space clears notes
		if text == " " || text == "" {
			empty := ""
			input.Notes = &empty
		} else {
			input.Notes = &text
		}

	case "quota":
		val, err := strconv.ParseFloat(text, 64)
		if err != nil || val < 0 {
			sendReply(inst.bot, chatID, "❌ 请输入有效的数字（>= 0），0 表示不限")
			return
		}
		input.Quota = &val

	case "exp":
		if text == "0" {
			// Clear expiration — use a far future date then clear
			// We need to handle this through the service
			input.ExpiresAt = nil
			// Special: update directly via apiKeyRepo to clear expiration
			key, err := m.apiKeyRepo.GetByID(ctx, pe.keyID)
			if err != nil || key.UserID != inst.resellerID {
				sendReply(inst.bot, chatID, "❌ 密钥未找到")
				return
			}
			key.ExpiresAt = nil
			if key.Status == "expired" {
				key.Status = StatusActive
			}
			if err := m.apiKeyRepo.Update(ctx, key); err != nil {
				sendReply(inst.bot, chatID, "❌ 更新失败: "+err.Error())
				return
			}
			m.finishEdit(inst, chatID, pe, msg.MessageID)
			return
		}

		// Try parsing as "Nd" (days)
		if strings.HasSuffix(text, "d") || strings.HasSuffix(text, "D") {
			days, err := strconv.Atoi(strings.TrimRight(text, "dD"))
			if err != nil || days <= 0 {
				sendReply(inst.bot, chatID, "❌ 天数必须为正整数，如 30d")
				return
			}
			t := time.Now().AddDate(0, 0, days)
			input.ExpiresAt = &t
		} else {
			// Try parsing as date
			t, err := time.Parse("2006-01-02", text)
			if err != nil {
				sendReply(inst.bot, chatID, "❌ 格式无效，请使用 2025-12-31 或 30d")
				return
			}
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			input.ExpiresAt = &t
		}

	default:
		return
	}

	// Perform update
	if _, err := m.resellerSvc.UpdateKey(ctx, inst.resellerID, pe.keyID, input); err != nil {
		sendReply(inst.bot, chatID, "❌ 更新失败: "+err.Error())
		return
	}

	m.finishEdit(inst, chatID, pe, msg.MessageID)
}

// finishEdit cleans up after a successful edit and refreshes the detail view.
func (m *TelegramBotManager) finishEdit(inst *botInstance, chatID int64, pe *pendingEdit, userMsgID int) {
	// Clear pending edit
	inst.editsMu.Lock()
	delete(inst.edits, chatID)
	inst.editsMu.Unlock()

	// Delete the prompt message and user's input message
	del := tgbotapi.NewDeleteMessage(chatID, pe.msgID)
	inst.bot.Request(del)
	del2 := tgbotapi.NewDeleteMessage(chatID, userMsgID)
	inst.bot.Request(del2)

	// Refresh the detail message
	ctx := context.Background()
	m.cbKeyDetail(inst, chatID, pe.detailM, ctx, pe.keyID, pe.listPg)
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

// escapeMarkdown escapes special Markdown characters for Telegram.
func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"`", "\\`",
		"[", "\\[",
	)
	return replacer.Replace(s)
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
