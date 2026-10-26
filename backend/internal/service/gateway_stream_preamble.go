package service

import (
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
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
