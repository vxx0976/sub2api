package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// buildAnthropicDirectMessagesURL returns the upstream Anthropic Messages API
// endpoint for platforms that natively support the Anthropic protocol.
//
//   - Deepseek: https://api.deepseek.com  →  https://api.deepseek.com/anthropic/v1/messages
//   - Kimi: https://api.kimi.com/coding/v1  →  https://api.kimi.com/coding/v1/messages
//   - Zhipu 官方: https://open.bigmodel.cn  →  https://open.bigmodel.cn/api/anthropic/v1/messages
//   - Zhipu 中转: https://relay.orbitai.cc  →  https://relay.orbitai.cc/v1/messages
func buildAnthropicDirectMessagesURL(account *Account) string {
	switch account.Platform {
	case PlatformDeepseek:
		baseURL := account.GetDeepseekBaseURL()
		// Strip /v1 suffix — the Anthropic-compatible path lives at /anthropic/v1/messages
		baseURL = strings.TrimSuffix(strings.TrimRight(baseURL, "/"), "/v1")
		return baseURL + "/anthropic/v1/messages"
	case PlatformKimi:
		baseURL := account.GetKimiBaseURL()
		return strings.TrimRight(baseURL, "/") + "/messages"
	case PlatformZhipu:
		// Zhipu (GLM) 原生 Anthropic 端点的根因上游而异：
		//   - 智谱官方 open.bigmodel.cn / api.z.ai：端点在 /api/anthropic 下
		//     （官方 ANTHROPIC_BASE_URL=https://open.bigmodel.cn/api/anthropic）。
		//   - NewAPI 类中转（如 relay.orbitai.cc）：直接在根暴露 /v1/messages。
		baseURL := strings.TrimRight(account.GetZhipuBaseURL(), "/")
		if u, err := url.Parse(baseURL); err == nil &&
			(u.Host == "open.bigmodel.cn" || u.Host == "api.z.ai") &&
			!strings.Contains(u.Path, "/api/anthropic") {
			return u.Scheme + "://" + u.Host + "/api/anthropic/v1/messages"
		}
		baseURL = strings.TrimSuffix(baseURL, "/v1")
		return baseURL + "/v1/messages"
	default:
		return ""
	}
}

// normalizeAnthropicDirectInputUsage 把直连上游的 usage 归一为主路径计费口径
// （input_tokens 含 cache_read，下游 actualInput = input_tokens - cache_read）。
//
// 各平台 Anthropic 端点的 input_tokens 语义：
//   - Deepseek（已实测 api.deepseek.com：input=71、cache_read=2944 对应 3015 token
//     prompt）：input_tokens 永远只报缓存未命中数，必须无条件加回缓存桶；
//     条件判断会在新增内容超过缓存前缀时漏计。
//   - Kimi（2026-09-14 实测 api.kimi.com/coding/v1/messages：2482 token 前缀已缓存、
//     再追加约 6250 token 新内容 → input=6430、cache_read=2304，两者之和才是完整
//     prompt）：同样是未命中口径，且新增内容远超缓存前缀是常态，必须无条件加回。
//     旧的条件加回会在 input >= cache_read 时漏计整段缓存前缀的「全价 − 缓存价」差额。
//   - 其他平台（Zhipu 等）：usage 语义未实测，仅在 input_tokens < cache_read+cache_creation
//     （明显为未命中口径，总量口径下 input 恒 >= 两缓存桶之和）时加回，避免总量口径上游
//     把缓存前缀按全价+缓存价双重计费。
//
// cache_read 与 cache_creation 必须同款条件一起加回：下游计费按互斥三桶拆分
// （actualInput = InputTokens - cache_read - cache_creation，见 openai_gateway_service.go
// RecordUsage），因此归一后的 InputTokens 必须是含全部桶的总量。只加回 cache_read
// 会让下游多减一次 cache_creation，creation>input 时把真实新输入夹成 0（漏计新输入费）。
func normalizeAnthropicDirectInputUsage(platform string, usage *OpenAIUsage) {
	if platform == PlatformDeepseek || platform == PlatformKimi ||
		usage.InputTokens < usage.CacheReadInputTokens+usage.CacheCreationInputTokens {
		usage.InputTokens += usage.CacheReadInputTokens + usage.CacheCreationInputTokens
	}
}

// anthropicDirectInputUsage 是归一化后的输入三桶（total 含全部桶，与 OpenAIUsage.InputTokens 同口径）。
type anthropicDirectInputUsage struct {
	total, cacheRead, cacheCreation int
}

func (u anthropicDirectInputUsage) apply(usage *OpenAIUsage) {
	usage.InputTokens = u.total
	usage.CacheReadInputTokens = u.cacheRead
	usage.CacheCreationInputTokens = u.cacheCreation
}

// resolveAnthropicDirectInputUsage 从一个 usage 节点（message.usage / message_delta.usage /
// 非流式 usage）解析并归一化输入三桶：
//   - 上游带 prompt_tokens / prompt_cache_hit_tokens / prompt_cache_miss_tokens 时，按显式总量
//     拆分（normalizeAnthropicCompatiblePromptUsage，与 Anthropic 原生路径同一口径）——这类
//     上游的 input_tokens 可能是总量（历史 Kimi 样例 input=prompt=173306、cache_read=173056），
//     按平台一律加回会双计；
//   - 否则按 normalizeAnthropicDirectInputUsage 的逐平台口径加回缓存桶。
func resolveAnthropicDirectInputUsage(platform string, node gjson.Result) anthropicDirectInputUsage {
	claude := ClaudeUsage{
		InputTokens:              int(node.Get("input_tokens").Int()),
		CacheReadInputTokens:     int(node.Get("cache_read_input_tokens").Int()),
		CacheCreationInputTokens: int(node.Get("cache_creation_input_tokens").Int()),
	}
	// 只有显式总量（prompt_tokens>0）或显式未命中桶（prompt_cache_miss_tokens）时才走拆分：
	// 只带 prompt_cache_hit_tokens 的节点在共享 helper 里会把未命中输入算成 0（少计）。
	explicitPromptTotal := node.Get("prompt_tokens").Int() > 0 || node.Get("prompt_cache_miss_tokens").Exists()
	if explicitPromptTotal && normalizeAnthropicCompatiblePromptUsage(node, &claude) {
		return anthropicDirectInputUsage{
			total:         claude.InputTokens + claude.CacheReadInputTokens + claude.CacheCreationInputTokens,
			cacheRead:     claude.CacheReadInputTokens,
			cacheCreation: claude.CacheCreationInputTokens,
		}
	}
	u := OpenAIUsage{
		InputTokens:              claude.InputTokens,
		CacheReadInputTokens:     claude.CacheReadInputTokens,
		CacheCreationInputTokens: claude.CacheCreationInputTokens,
	}
	normalizeAnthropicDirectInputUsage(platform, &u)
	return anthropicDirectInputUsage{total: u.InputTokens, cacheRead: u.CacheReadInputTokens, cacheCreation: u.CacheCreationInputTokens}
}

// hasAnthropicDirectInputFields 判断 usage 节点是否显式携带了任一输入字段。
func hasAnthropicDirectInputFields(node gjson.Result) bool {
	for _, key := range []string{
		"input_tokens", "cache_read_input_tokens", "cache_creation_input_tokens",
		"prompt_tokens", "prompt_cache_hit_tokens", "prompt_cache_miss_tokens",
	} {
		if node.Get(key).Exists() {
			return true
		}
	}
	return false
}

// mergeAnthropicDirectDeltaInputUsage 用 message_delta 携带的最终输入拆分覆盖
// message_start 的值。Anthropic 流式协议里 message_delta.usage 是累计终值；Kimi 编程
// 端点在 message_start 只报「输入=完整 prompt、缓存=0」，真实的未命中/命中拆分只在
// message_delta 给出（实测 start 8736/0 → delta 32/8704），只取 start 会把缓存命中
// 按全价计费。
//
// 护栏（均按**归一化后**的总量比较，避免各平台加回口径不同导致误判）：
//   - delta 未显式携带输入字段（DeepSeek 等只报 output_tokens）→ 保持 start；
//   - delta 归一后总量为 0 或小于 start 总量 → 保持 start。宁可少享受缓存折扣（多收、
//     可发现），也绝不让残缺/口径不明的 delta 把已记录的 prompt 量调小（少收）。
//
// 返回是否发生了覆盖。
func mergeAnthropicDirectDeltaInputUsage(platform string, deltaUsage gjson.Result, start anthropicDirectInputUsage, usage *OpenAIUsage) bool {
	if !hasAnthropicDirectInputFields(deltaUsage) {
		return false
	}
	delta := resolveAnthropicDirectInputUsage(platform, deltaUsage)
	if delta.total == 0 || delta.total < start.total || delta == start {
		return false
	}
	delta.apply(usage)
	return true
}

// anthropicDirectReasoningEffort 计算存量 CN 直通路径的计费推理档。
//
// fork: 与显式协议的原生路径（forwardAnthropicViaNativeAnthropicEndpoint）用同一组 helper：
// 优先 output_config.effort，缺失且 thinking 已启用时按国产 passback-required 模型兜底 high。
// 唯一差异是 GLM-5.3：原生路径先做 NormalizeGLM53AnthropicThinking 再取档，本路径不改写
// 上游 body，因此按客户端原样发送的档位记账。
// forwardAnthropicDirect 的各个 return 都不带 ReasoningEffort，由调用方统一补上——
// 否则分组/渠道价卡的 reasoning_effort_multipliers 在这条路径上恒按 1x 计，
// usage_logs.reasoning_effort 也恒为 NULL。无任何 effort 信号时返回 nil（倍率保持 1x）。
// 必须用改写前的客户端 body 计算。
func anthropicDirectReasoningEffort(body []byte, billingModel string) *string {
	requested := NormalizeClaudeOutputEffort(gjson.GetBytes(body, "output_config.effort").String())
	return ApplyThinkingEnabledFallback(requested, body, billingModel)
}

// forwardAnthropicDirect forwards an Anthropic Messages request directly to
// upstream platforms that expose a native Anthropic-compatible endpoint
// (Deepseek /anthropic, Kimi /coding). Unlike the normal ForwardAsAnthropic
// path, this skips the Anthropic→Responses format conversion and pipes the
// upstream Anthropic SSE/JSON response through unchanged.
func (s *OpenAIGatewayService) forwardAnthropicDirect(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	originalModel, billingModel, upstreamModel string,
	clientStream bool,
	startTime time.Time,
) (*OpenAIForwardResult, error) {

	// 1. Replace model in the Anthropic request body.
	body = ReplaceModelInBody(body, upstreamModel)

	// 2. Build upstream URL.
	targetURL := buildAnthropicDirectMessagesURL(account)
	if targetURL == "" {
		return nil, fmt.Errorf("unsupported platform for direct Anthropic forwarding: %s", account.Platform)
	}

	// 3. Get access token.
	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	// 4. Build HTTP request.
	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build anthropic direct request: %w", err)
	}
	// Deepseek / Kimi / Zhipu 用标准 Anthropic x-api-key。
	req.Header.Set("x-api-key", token)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("accept", "application/json")

	// Kimi For Coding 对客户端做白名单校验，需为 Coding Agent UA（前缀 claude-cli/）。
	if account.Platform == PlatformKimi {
		baseURL := account.GetKimiBaseURL()
		if strings.Contains(baseURL, "api.kimi.com") {
			req.Header.Set("user-agent", kimiCodingUserAgent)
		}
	}

	// Passthrough anthropic-beta header from client if present.
	if beta := c.GetHeader("anthropic-beta"); beta != "" {
		req.Header.Set("anthropic-beta", beta)
	}

	logger.L().Debug("anthropic_direct: forwarding request",
		zap.Int64("account_id", account.ID),
		zap.String("platform", string(account.Platform)),
		zap.String("target_url", targetURL),
		zap.String("original_model", originalModel),
		zap.String("upstream_model", upstreamModel),
		zap.Bool("stream", clientStream),
	)

	// 5. Send request via httpUpstream (respects proxy settings).
	proxyURL := ""
	if account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		safeErr := sanitizeUpstreamErrorMessage(err.Error())
		setOpsUpstreamError(c, 0, safeErr, "")
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: 0,
			Kind:               "request_error",
			Message:            safeErr,
		})
		writeAnthropicError(c, http.StatusBadGateway, "api_error", "Upstream request failed")
		return nil, fmt.Errorf("upstream request failed: %s", safeErr)
	}
	defer func() { _ = resp.Body.Close() }()

	// 6. Handle error responses — support failover.
	if resp.StatusCode >= 400 {
		respBody := s.readUpstreamErrorBody(resp)
		_ = resp.Body.Close()
		// 回卷 Body：非 failover 分支下方的 handleAnthropicErrorResponse 会再次
		// readUpstreamErrorBody，若不回卷则读到空 body，applyErrorPassthroughRule /
		// cyber 检测 / extractUpstreamErrorMessage 全部拿不到真实上游错误。
		// 对齐 readOpenAIUpstreamError 的读-关-回卷模式。
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		upstreamMsg := strings.TrimSpace(string(respBody))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)

		if s.shouldFailoverOpenAIUpstreamResponse(account, resp.StatusCode, upstreamMsg, respBody) {
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				Kind:               "failover",
				Message:            upstreamMsg,
			})
			s.handleOpenAIAccountUpstreamError(ctx, account, resp.StatusCode, resp.Header, respBody, upstreamModel)
			return nil, &UpstreamFailoverError{
				StatusCode:             resp.StatusCode,
				ResponseBody:           respBody,
				RetryableOnSameAccount: account.IsPoolMode() && account.IsPoolModeRetryableStatus(resp.StatusCode),
			}
		}
		// Non-failover error: pass through the upstream Anthropic error to client.
		return s.handleAnthropicErrorResponse(resp, c, account, billingModel)
	}

	// 7. Handle successful response.
	if clientStream {
		return s.handleAnthropicDirectStreamingResponse(resp, c, account.Platform, originalModel, billingModel, upstreamModel, startTime)
	}
	return s.handleAnthropicDirectBufferedResponse(resp, c, account, originalModel, billingModel, upstreamModel, startTime)
}

// handleAnthropicDirectStreamingResponse pipes an upstream Anthropic SSE stream
// directly to the client while extracting usage information for billing.
func (s *OpenAIGatewayService) handleAnthropicDirectStreamingResponse(
	resp *http.Response,
	c *gin.Context,
	platform string,
	originalModel, billingModel, upstreamModel string,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	var usage OpenAIUsage
	var startInput anthropicDirectInputUsage
	var requestID, responseID string
	var firstTokenMs *int

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// Write every line to client immediately.
		_, writeErr := c.Writer.WriteString(line + "\n")
		if writeErr != nil {
			return &OpenAIForwardResult{
				Model:            originalModel,
				BillingModel:     billingModel,
				UpstreamModel:    upstreamModel,
				Usage:            usage,
				Stream:           true,
				Duration:         time.Since(startTime),
				ClientDisconnect: true,
			}, fmt.Errorf("client write error: %w", writeErr)
		}
		c.Writer.Flush()

		// Parse "data: {...}" or "data:{...}" lines to extract usage.
		// Standard Anthropic uses "data: " (with space), Kimi uses "data:" (no space).
		var data string
		if strings.HasPrefix(line, "data: ") {
			data = line[6:]
		} else if strings.HasPrefix(line, "data:") {
			data = line[5:]
		} else {
			continue
		}
		if data == "[DONE]" {
			continue
		}

		eventType := gjson.Get(data, "type").String()

		switch eventType {
		case "message_start":
			// First content event — record TTFT.
			if firstTokenMs == nil {
				ms := int(time.Since(startTime).Milliseconds())
				firstTokenMs = &ms
			}
			// Extract input tokens and request/response IDs.
			msg := gjson.Get(data, "message")
			responseID = msg.Get("id").String()
			// 归一化口径见 resolveAnthropicDirectInputUsage。
			startInput = resolveAnthropicDirectInputUsage(platform, msg.Get("usage"))
			startInput.apply(&usage)

		case "content_block_start", "content_block_delta":
			if firstTokenMs == nil {
				ms := int(time.Since(startTime).Milliseconds())
				firstTokenMs = &ms
			}

		case "message_delta":
			deltaUsage := gjson.Get(data, "usage")
			usage.OutputTokens = int(deltaUsage.Get("output_tokens").Int())
			mergeAnthropicDirectDeltaInputUsage(platform, deltaUsage, startInput, &usage)
		}
	}

	if err := scanner.Err(); err != nil {
		return &OpenAIForwardResult{
			Model:            originalModel,
			BillingModel:     billingModel,
			UpstreamModel:    upstreamModel,
			Usage:            usage,
			Stream:           true,
			Duration:         time.Since(startTime),
			ClientDisconnect: true,
		}, fmt.Errorf("upstream read error: %w", err)
	}

	// Extract request ID from response headers.
	if rid := resp.Header.Get("request-id"); rid != "" {
		requestID = rid
	} else if rid := resp.Header.Get("x-request-id"); rid != "" {
		requestID = rid
	}

	return &OpenAIForwardResult{
		RequestID:     requestID,
		ResponseID:    responseID,
		Usage:         usage,
		Model:         originalModel,
		BillingModel:  billingModel,
		UpstreamModel: upstreamModel,
		Stream:        true,
		Duration:      time.Since(startTime),
		FirstTokenMs:  firstTokenMs,
	}, nil
}

// handleAnthropicDirectBufferedResponse reads the full upstream Anthropic JSON
// response, writes it to the client, and extracts usage for billing.
func (s *OpenAIGatewayService) handleAnthropicDirectBufferedResponse(
	resp *http.Response,
	c *gin.Context,
	account *Account,
	originalModel, billingModel, upstreamModel string,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	platform := account.Platform
	// Even when stream=false, some Anthropic-compatible upstreams may still
	// return SSE. Detect by Content-Type and delegate to the streaming handler.
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "text/event-stream") {
		return s.handleAnthropicDirectBufferedSSE(resp, c, account, originalModel, billingModel, upstreamModel, startTime)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writeAnthropicError(c, http.StatusBadGateway, "api_error", "Failed to read upstream response")
		return nil, fmt.Errorf("read upstream response: %w", err)
	}

	// Extract usage from the JSON response.
	var usage OpenAIUsage
	usageNode := gjson.GetBytes(respBody, "usage")
	usage.OutputTokens = int(usageNode.Get("output_tokens").Int())
	// 归一化口径见 resolveAnthropicDirectInputUsage。
	resolveAnthropicDirectInputUsage(platform, usageNode).apply(&usage)

	responseID := gjson.GetBytes(respBody, "id").String()
	requestID := resp.Header.Get("x-request-id")

	// Write the JSON response to client.
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusOK)
	_, _ = c.Writer.Write(respBody)

	return &OpenAIForwardResult{
		RequestID:     requestID,
		ResponseID:    responseID,
		Usage:         usage,
		Model:         originalModel,
		BillingModel:  billingModel,
		UpstreamModel: upstreamModel,
		Stream:        false,
		Duration:      time.Since(startTime),
	}, nil
}

// handleAnthropicDirectBufferedSSE handles the case where the upstream returns
// SSE even though the client requested stream=false. It buffers all events,
// assembles the final message, and returns it as a single JSON response.
func (s *OpenAIGatewayService) handleAnthropicDirectBufferedSSE(
	resp *http.Response,
	c *gin.Context,
	account *Account,
	originalModel, billingModel, upstreamModel string,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	platform := account.Platform
	var usage OpenAIUsage
	var startInput anthropicDirectInputUsage
	var lastMessageData []byte
	var responseID, requestID string
	var stopReason string
	var errorEventData string

	// 累积各 content block 的文本/思考增量，用于在非流式响应里重建 content 数组。
	type sseContentBlock struct {
		typ string
		sb  strings.Builder
	}
	contentBlocks := map[int]*sseContentBlock{}
	var blockOrder []int
	blockAt := func(idx int) *sseContentBlock {
		b := contentBlocks[idx]
		if b == nil {
			b = &sseContentBlock{}
			contentBlocks[idx] = b
			blockOrder = append(blockOrder, idx)
		}
		return b
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		var data string
		if strings.HasPrefix(line, "data: ") {
			data = line[6:]
		} else if strings.HasPrefix(line, "data:") {
			data = line[5:]
		} else {
			continue
		}
		if data == "[DONE]" {
			continue
		}

		eventType := gjson.Get(data, "type").String()
		switch eventType {
		case "message_start":
			msg := gjson.Get(data, "message")
			responseID = msg.Get("id").String()
			// 归一化口径见 resolveAnthropicDirectInputUsage。
			startInput = resolveAnthropicDirectInputUsage(platform, msg.Get("usage"))
			startInput.apply(&usage)
			// Store the initial message object as the base for the final response.
			lastMessageData = []byte(msg.Raw)

		case "content_block_start":
			idx := int(gjson.Get(data, "index").Int())
			blockAt(idx).typ = gjson.Get(data, "content_block.type").String()

		case "content_block_delta":
			// 重建 content：累积 text / thinking 增量（按 block index 归位）。
			idx := int(gjson.Get(data, "index").Int())
			b := blockAt(idx)
			switch gjson.Get(data, "delta.type").String() {
			case "thinking_delta":
				if b.typ == "" {
					b.typ = "thinking"
				}
				_, _ = b.sb.WriteString(gjson.Get(data, "delta.thinking").String())
			case "text_delta", "":
				if b.typ == "" {
					b.typ = "text"
				}
				_, _ = b.sb.WriteString(gjson.Get(data, "delta.text").String())
			}

		case "message_delta":
			deltaUsage := gjson.Get(data, "usage")
			usage.OutputTokens = int(deltaUsage.Get("output_tokens").Int())
			mergeAnthropicDirectDeltaInputUsage(platform, deltaUsage, startInput, &usage)
			if sr := gjson.Get(data, "delta.stop_reason").String(); sr != "" {
				stopReason = sr
			}

		case "error":
			// 上游在 SSE 中下发 error 事件（overloaded_error 等）。buffered 路径此时
			// 尚未向客户端写入任何字节，记录后在循环外按错误处理（failover / 透传），
			// 不能当成功返回。保留最后一个 error 事件。
			errorEventData = data
		}
	}

	requestID = resp.Header.Get("x-request-id")

	// 断流 / 错误检测：buffered 路径的所有客户端写入都发生在下方（此刻尚未写出任何
	// 字节），因此返回 error 可安全触发上层 failover（对照流式路径 296 的 scanner.Err
	// 检查）。否则截断内容会被当成功返回、message_start 前断连会返回空 200。
	if errorEventData != "" {
		message := sanitizeUpstreamErrorMessage(strings.TrimSpace(gjson.Get(errorEventData, "error.message").String()))
		if openAIStreamFailedEventShouldFailover([]byte(errorEventData), message) {
			return nil, s.newOpenAIStreamFailoverError(c, account, false, requestID, []byte(errorEventData), message)
		}
		// 非可 failover 错误（invalid_request / policy 等）：以 Anthropic 错误格式回写客户端。
		message = s.recordOpenAIStreamUpstreamError(c, account, false, requestID, "http_error", []byte(errorEventData), message)
		if message == "" {
			message = "Upstream returned an error event"
		}
		status, errType := openAIStreamFailureClientStatus([]byte(errorEventData), message, "api_error")
		writeAnthropicError(c, status, errType, message)
		return nil, fmt.Errorf("anthropic direct sse error event: %s", message)
	}
	if err := scanner.Err(); err != nil {
		// 上游 SSE 读取中断（连接被截断）：截断内容不能当成功，触发 failover。
		return nil, s.newOpenAIStreamFailoverError(c, account, false, requestID, nil,
			"anthropic direct stream read error: "+sanitizeUpstreamErrorMessage(err.Error()))
	}
	if lastMessageData == nil {
		// 一字节 message_start 都没收到（message_start 前断连）：空 200 不能当成功，
		// 触发 failover。
		return nil, s.newOpenAIStreamFailoverError(c, account, false, requestID, nil,
			"anthropic direct stream closed before message_start")
	}

	// Build a minimal but correct JSON response.
	// Reconstruct from the message_start base + accumulated content.
	// For simplicity, re-read the SSE stream result. Since we already have
	// lastMessageData from message_start, inject final usage and write it.
	if lastMessageData != nil {
		// Update usage in the response.
		type anthropicUsage struct {
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
		}
		var msgResp map[string]any
		if err := json.Unmarshal(lastMessageData, &msgResp); err == nil {
			// 回给客户端的是 Anthropic 口径：input_tokens 只含未命中部分。usage.InputTokens
			// 是计费用的含全部桶总量，原样写回会让客户端把缓存再算一遍（上下文占用虚高）。
			msgResp["usage"] = anthropicUsage{
				InputTokens:              max(usage.InputTokens-usage.CacheReadInputTokens-usage.CacheCreationInputTokens, 0),
				OutputTokens:             usage.OutputTokens,
				CacheCreationInputTokens: usage.CacheCreationInputTokens,
				CacheReadInputTokens:     usage.CacheReadInputTokens,
			}
			// 用累积的增量重建 content 数组（text / thinking 块）。message_start 的
			// content 为空，必须在此回填，否则非流式客户端拿到空响应。
			// tool_use 等块本路径不重建（流式路径已正确处理）。
			if len(blockOrder) > 0 {
				blocks := make([]map[string]any, 0, len(blockOrder))
				for _, idx := range blockOrder {
					b := contentBlocks[idx]
					switch b.typ {
					case "thinking":
						blocks = append(blocks, map[string]any{"type": "thinking", "thinking": b.sb.String()})
					case "text", "":
						blocks = append(blocks, map[string]any{"type": "text", "text": b.sb.String()})
					}
				}
				if len(blocks) > 0 {
					msgResp["content"] = blocks
				}
			}
			if stopReason != "" {
				msgResp["stop_reason"] = stopReason
			}
			finalBody, _ := json.Marshal(msgResp)
			c.Writer.Header().Set("Content-Type", "application/json")
			c.Writer.WriteHeader(http.StatusOK)
			_, _ = c.Writer.Write(finalBody)
		}
	}

	return &OpenAIForwardResult{
		RequestID:     requestID,
		ResponseID:    responseID,
		Usage:         usage,
		Model:         originalModel,
		BillingModel:  billingModel,
		UpstreamModel: upstreamModel,
		Stream:        false,
		Duration:      time.Since(startTime),
	}, nil
}
