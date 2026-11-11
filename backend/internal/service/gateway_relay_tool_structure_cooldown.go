package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

const (
	// relayToolStructureRejectCooldown 中转拒绝请求的 Claude Code 工具结构时对账号的冷却时长。
	// 这类拒绝来自中转自己的入参校验（"incomplete or unsupported Claude Code tool structure"），
	// 同一请求换到同组其他中转通常能过；不冷却的话，粘性会话会把同一会话反复钉在这家中转上，
	// 每次都先失败一次再换号（2026-11-11 早上 Max-开心 两小时里被同一会话重试了 60 次）。
	relayToolStructureRejectCooldown = 30 * time.Minute
	// relayToolStructureRejectMinAlternatives 冷却要求分组内至少还剩多少个可调度替补。
	// 取 2 而不是 1：请求本身真的残缺时每个中转都会拒，这样一个坏请求最多只能冷却掉一个号，
	// 不会顺着换号把整组逐个冷却到无号可用。
	relayToolStructureRejectMinAlternatives = 2
	// relayToolStructureRejectPausedKey 标记本次请求已经因工具结构拒绝冷却过账号。
	relayToolStructureRejectPausedKey = "gateway_relay_tool_structure_paused"
	// relayToolStructureRejectWriteTimeout 冷却写入的独立超时。写入发生在请求路径上，
	// 且刻意不受客户端断开取消，必须自带上限，否则快照刷新卡住会把换号一起拖死。
	relayToolStructureRejectWriteTimeout = 2 * time.Second
)

// relayToolStructurePauseMu 串行化"数替补 → 写冷却"这一步。
// 替补数是写之前读的，不串行化的话，同组多个号在同一瞬间被拒会各自读到足够的替补、
// 各自写冷却，把整组冷却到无号可用（读-改-写竞态）。同一进程内的并发由此收敛，
// 跨副本的极端并发仍靠 relayToolStructureRejectMinAlternatives 兜底。
var relayToolStructurePauseMu sync.Mutex

// isRelayToolStructureRejectedError 识别中转对 Claude Code 工具结构的拒绝，例如
// {"error":{"message":"request contains an incomplete or unsupported Claude Code tool structure",
// "type":"invalid_request_error"}}。官方 API 不会返回这种文案，它只来自中转的入参校验。
func isRelayToolStructureRejectedError(statusCode int, respBody []byte) bool {
	if statusCode != http.StatusUnprocessableEntity && statusCode != http.StatusBadRequest {
		return false
	}
	msg := strings.ToLower(extractUpstreamErrorMessage(respBody))
	if !strings.Contains(msg, "tool structure") {
		return false
	}
	return strings.Contains(msg, "incomplete") || strings.Contains(msg, "unsupported")
}

func relayToolStructureRejectPaused(c *gin.Context) bool {
	if c == nil {
		return false
	}
	v, ok := c.Get(relayToolStructureRejectPausedKey)
	if !ok {
		return false
	}
	paused, _ := v.(bool)
	return paused
}

func setRelayToolStructureRejectPaused(c *gin.Context) {
	if c != nil {
		c.Set(relayToolStructureRejectPausedKey, true)
	}
}

// tempUnscheduleRelayToolStructureReject 对中转的工具结构拒绝做账号级冷却，
// 让后续请求（含同一会话的重试）先走其他号，而不是继续粘在这家中转上。
func (s *GatewayService) tempUnscheduleRelayToolStructureReject(ctx context.Context, c *gin.Context, account *Account, body []byte, model string) {
	if s == nil || s.accountRepo == nil || account == nil || c == nil {
		return
	}
	// 只对中转号生效：官方 OAuth 号的入参校验由网关自身控制，冷却账号解决不了问题。
	if account.Type != AccountTypeAPIKey || account.Platform != PlatformAnthropic {
		return
	}
	if relayToolStructureRejectPaused(c) {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now := time.Now()
	if account.TempUnschedulableUntil != nil && now.Before(*account.TempUnschedulableUntil) {
		return
	}
	apiKey := getAPIKeyFromContext(c)
	if apiKey == nil || apiKey.GroupID == nil || *apiKey.GroupID <= 0 {
		return
	}
	groupIDs := []int64{*apiKey.GroupID}
	for _, gid := range account.GroupIDs {
		if gid > 0 && gid != *apiKey.GroupID {
			groupIDs = append(groupIDs, gid)
		}
	}
	relayToolStructurePauseMu.Lock()
	defer relayToolStructurePauseMu.Unlock()
	// 冷却作用于账号本身（所有分组）：任一分组替补不够就放弃，避免把整组打到无号可用。
	// 替补数用请求上下文查（组合分组的路由决策等信息在 ctx 上），写入用不受取消影响的 ctx。
	for _, gid := range groupIDs {
		if s.countRelayAlternatives(ctx, gid, account, model) < relayToolStructureRejectMinAlternatives {
			return
		}
	}
	until := now.Add(relayToolStructureRejectCooldown).Truncate(time.Second).Add(time.Second)
	reason := fmt.Sprintf("relay rejected claude code tool structure (auto temp-unschedule %s): %s",
		relayToolStructureRejectCooldown, truncateString(extractUpstreamErrorMessage(body), 120))
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), relayToolStructureRejectWriteTimeout)
	defer cancel()
	if err := s.accountRepo.SetTempUnschedulable(writeCtx, account.ID, until, reason); err != nil {
		logger.LegacyPrintf("service.gateway",
			"[Forward] relay tool structure reject temp_unschedule_failed account=%d error=%v", account.ID, err)
		return
	}
	setRelayToolStructureRejectPaused(c)
	logger.LegacyPrintf("service.gateway",
		"[Forward] relay tool structure reject cooldown account=%d(%s) until=%s",
		account.ID, account.Name, until.Format("15:04:05"))
}

// countRelayAlternatives 返回分组内除 account 外、可调度且支持 model 的账号数。
func (s *GatewayService) countRelayAlternatives(ctx context.Context, groupID int64, account *Account, model string) int {
	if s == nil || s.accountRepo == nil || account == nil {
		return 0
	}
	peers, err := s.accountRepo.ListSchedulableByGroupIDAndPlatform(ctx, groupID, account.Platform)
	if err != nil {
		return 0
	}
	count := 0
	for i := range peers {
		if peers[i].ID == account.ID || !peers[i].IsSchedulable() {
			continue
		}
		if model != "" && !s.isModelSupportedByAccountWithContext(ctx, &peers[i], model) {
			continue
		}
		count++
	}
	return count
}
