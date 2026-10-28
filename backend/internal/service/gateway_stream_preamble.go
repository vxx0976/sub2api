package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// streamPreambleOnlyKey 标记本次转发尝试写给客户端的只有前导帧（ping/注释/我方 keepalive），
// 尚未提交任何语义内容。中转在排队等上游并发时会先回 200 并持续发 ping，
// 等不到再在流内报错；此时客户端只收到 ping，换号重放不会造成流拼接腐化。
const streamPreambleOnlyKey = "gateway_stream_preamble_only"

// ResetStreamPreamble 在每次转发尝试开始前清除前导标记，避免上一次尝试的状态残留。
func ResetStreamPreamble(c *gin.Context) {
	if c != nil {
		c.Set(streamPreambleOnlyKey, false)
	}
}

func setStreamPreambleOnly(c *gin.Context, preambleOnly bool) {
	if c != nil {
		c.Set(streamPreambleOnlyKey, preambleOnly)
	}
}

func streamPreambleOnly(c *gin.Context) bool {
	if c == nil {
		return false
	}
	v, ok := c.Get(streamPreambleOnlyKey)
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

// StreamOutputCommitted 报告自 sizeBefore 以来是否已向客户端写出语义内容。
// 只写过前导帧时返回 false，调用方仍可换号重试。
func StreamOutputCommitted(c *gin.Context, sizeBefore int) bool {
	if c == nil || c.Writer == nil {
		return false
	}
	if c.Writer.Size() == sizeBefore {
		return false
	}
	return !streamPreambleOnly(c)
}

// anthropicStreamPreambleKind 对单个 SSE 事件分类：
//   - "ping"：可直接写给客户端、不提交流（保活）；
//   - "hold"：未携带计量 token 的 message_start，暂存到首个内容事件再写出；
//   - ""：语义事件，触发提交。
//
// 携带 token 的 message_start 视为上游已开始计量，立即提交：其后的流内错误按部分交付计费
// （shouldBillPartialStream），不换号。
func anthropicStreamPreambleKind(data string, usagePatch *sseUsagePatch) string {
	if data == "" {
		return "ping"
	}
	switch gjson.Get(data, "type").String() {
	case "ping":
		return "ping"
	case "message_start":
		if usagePatch != nil && (usagePatch.inputTokens > 0 || usagePatch.outputTokens > 0 ||
			usagePatch.cacheCreationInputTokens > 0 || usagePatch.cacheReadInputTokens > 0 ||
			usagePatch.cacheCreation5mTokens > 0 || usagePatch.cacheCreation1hTokens > 0) {
			return ""
		}
		return "hold"
	default:
		return ""
	}
}

// relayQueueRejectCooldown 中转排队后流内明确报"并发超限"时的冷却。
// 中转并发打满通常会持续一段时间，每次选回来都要先空等 30~120s 的 ping 才报错，
// 宁可让主号闲置一段时间，也不让用户反复空等。
const relayQueueRejectCooldown = 30 * time.Minute

// relayRateLimitRejectCooldown 其他流内 rate_limit_error（中转按 key/模型的 RPM、
// 级联中转透传的上游 429 等）的短冷却：这类限流往往只针对单个模型、几十秒内恢复，
// 不能按并发超限冷却 30 分钟。
const relayRateLimitRejectCooldown = 90 * time.Second

// relayQueueRejectStreamCooldown 返回首个语义事件之前的流内错误应冷却的时长，0 表示不冷却。
// "No available accounts" 可能只是单个模型无容量，不在此列（仅换号、不冷却整个账号）。
func relayQueueRejectStreamCooldown(body []byte) time.Duration {
	if strings.Contains(strings.ToLower(extractUpstreamErrorMessage(body)), "concurrency limit") {
		return relayQueueRejectCooldown
	}
	if gjson.GetBytes(body, "error.type").String() == "rate_limit_error" {
		return relayRateLimitRejectCooldown
	}
	return 0
}

// tempUnscheduleRelayQueueReject 对 API Key 中转的排队拒绝做冷却，让后续请求先走其他号。
// 冷却作用于账号本身（所有分组），因此要求账号所在的每个分组都还有能服务该模型的
// 其他可调度账号；任一分组只剩它时不冷却：冷却只会把"慢、偶尔成功"变成"全部失败"。
func (s *GatewayService) tempUnscheduleRelayQueueReject(ctx context.Context, c *gin.Context, account *Account, body []byte, model string) {
	if s == nil || s.accountRepo == nil || account == nil || account.Type != AccountTypeAPIKey {
		return
	}
	cooldown := relayQueueRejectStreamCooldown(body)
	if cooldown <= 0 {
		return
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
	for _, gid := range groupIDs {
		if !s.groupHasRelayAlternative(ctx, gid, account, model) {
			return
		}
	}
	until := now.Add(cooldown).Truncate(time.Second).Add(time.Second)
	reason := fmt.Sprintf("relay queue reject in stream (auto temp-unschedule %s): %s",
		cooldown, truncateString(extractUpstreamErrorMessage(body), 120))
	if err := s.accountRepo.SetTempUnschedulable(ctx, account.ID, until, reason); err != nil {
		logger.LegacyPrintf("service.gateway", "[Forward] relay queue reject temp_unschedule_failed account=%d error=%v", account.ID, err)
		return
	}
	logger.LegacyPrintf("service.gateway", "[Forward] relay queue reject cooldown account=%d(%s) until=%s", account.ID, account.Name, until.Format("15:04:05"))
}

// groupHasRelayAlternative 报告分组内是否还有除 account 外、可调度且支持 model 的账号。
func (s *GatewayService) groupHasRelayAlternative(ctx context.Context, groupID int64, account *Account, model string) bool {
	peers, err := s.accountRepo.ListSchedulableByGroupIDAndPlatform(ctx, groupID, account.Platform)
	if err != nil {
		return false
	}
	for i := range peers {
		if peers[i].ID == account.ID || !peers[i].IsSchedulable() {
			continue
		}
		if model != "" && !s.isModelSupportedByAccountWithContext(ctx, &peers[i], model) {
			continue
		}
		return true
	}
	return false
}

// writeClaudeErrorBody 写出 Anthropic 风格错误响应。若本次请求已以 200 SSE 开始
// （先前尝试只发过前导 ping 后换号），改写为 SSE error 帧，避免裸 JSON 混进事件流。
func writeClaudeErrorBody(c *gin.Context, status int, body []byte) {
	if c == nil {
		return
	}
	if !c.Writer.Written() || !strings.Contains(strings.ToLower(c.Writer.Header().Get("Content-Type")), "text/event-stream") {
		c.Data(status, "application/json; charset=utf-8", body)
		return
	}
	errType := gjson.GetBytes(body, "error.type").String()
	message := extractUpstreamErrorMessage(body)
	var compact bytes.Buffer
	if gjson.GetBytes(body, "type").String() != "error" || errType == "" || json.Compact(&compact, body) != nil {
		if errType == "" {
			errType = "api_error"
		}
		if message == "" {
			message = http.StatusText(status)
		}
		compact.Reset()
		data, _ := json.Marshal(gin.H{"type": "error", "error": gin.H{"type": errType, "message": message}})
		compact.Write(data)
	}
	MarkOpsStreamError(c, errType, message, status)
	_, _ = fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", compact.Bytes())
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// writeClaudeErrorJSON 同 writeClaudeErrorBody，接受 gin.H。
func writeClaudeErrorJSON(c *gin.Context, status int, obj gin.H) {
	body, err := json.Marshal(obj)
	if err != nil {
		c.JSON(status, obj)
		return
	}
	writeClaudeErrorBody(c, status, body)
}
