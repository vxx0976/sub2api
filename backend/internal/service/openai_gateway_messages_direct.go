package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
//   - DeepSeek: https://api.deepseek.com  →  https://api.deepseek.com/anthropic/v1/messages
//   - Moonshot: https://api.kimi.com/coding/v1  →  https://api.kimi.com/coding/v1/messages
func buildAnthropicDirectMessagesURL(account *Account) string {
	switch account.Platform {
	case PlatformDeepSeek:
		baseURL := account.GetDeepSeekBaseURL()
		// Strip /v1 suffix — the Anthropic-compatible path lives at /anthropic/v1/messages
		baseURL = strings.TrimSuffix(strings.TrimRight(baseURL, "/"), "/v1")
		return baseURL + "/anthropic/v1/messages"
	case PlatformMoonshot:
		baseURL := account.GetMoonshotBaseURL()
		return strings.TrimRight(baseURL, "/") + "/messages"
	default:
		return ""
	}
}

// forwardAnthropicDirect forwards an Anthropic Messages request directly to
// upstream platforms that expose a native Anthropic-compatible endpoint
// (DeepSeek /anthropic, Kimi /coding). Unlike the normal ForwardAsAnthropic
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
	req.Header.Set("x-api-key", token)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("accept", "application/json")

	// Kimi Code API requires User-Agent containing "claude-code".
	if account.Platform == PlatformMoonshot {
		baseURL := account.GetMoonshotBaseURL()
		if strings.Contains(baseURL, "api.kimi.com") {
			req.Header.Set("user-agent", "claude-code/1.0")
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

		upstreamMsg := strings.TrimSpace(string(respBody))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)

		if s.shouldFailoverOpenAIUpstreamResponse(resp.StatusCode, upstreamMsg, respBody) {
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
		return s.handleAnthropicDirectStreamingResponse(resp, c, originalModel, billingModel, upstreamModel, startTime)
	}
	return s.handleAnthropicDirectBufferedResponse(resp, c, originalModel, billingModel, upstreamModel, startTime)
}

// handleAnthropicDirectStreamingResponse pipes an upstream Anthropic SSE stream
// directly to the client while extracting usage information for billing.
func (s *OpenAIGatewayService) handleAnthropicDirectStreamingResponse(
	resp *http.Response,
	c *gin.Context,
	originalModel, billingModel, upstreamModel string,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	var usage OpenAIUsage
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
		c.Writer.(http.Flusher).Flush()

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
			usage.InputTokens = int(msg.Get("usage.input_tokens").Int())
			usage.CacheCreationInputTokens = int(msg.Get("usage.cache_creation_input_tokens").Int())
			usage.CacheReadInputTokens = int(msg.Get("usage.cache_read_input_tokens").Int())

			// DeepSeek/Kimi Anthropic compat bug: when cache hits exist,
			// input_tokens reports only cache-miss count instead of the total.
			// Standard Anthropic: input_tokens = total (including cached).
			// Detect and normalize so downstream billing computes correct
			// actualInput = input_tokens - cache_read = cache_miss.
			if usage.CacheReadInputTokens > 0 && usage.InputTokens < usage.CacheReadInputTokens {
				usage.InputTokens += usage.CacheReadInputTokens
			}

		case "content_block_start", "content_block_delta":
			if firstTokenMs == nil {
				ms := int(time.Since(startTime).Milliseconds())
				firstTokenMs = &ms
			}

		case "message_delta":
			// Extract output tokens.
			usage.OutputTokens = int(gjson.Get(data, "usage.output_tokens").Int())
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
	originalModel, billingModel, upstreamModel string,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	// Even when stream=false, some Anthropic-compatible upstreams may still
	// return SSE. Detect by Content-Type and delegate to the streaming handler.
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "text/event-stream") {
		return s.handleAnthropicDirectBufferedSSE(resp, c, originalModel, billingModel, upstreamModel, startTime)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writeAnthropicError(c, http.StatusBadGateway, "api_error", "Failed to read upstream response")
		return nil, fmt.Errorf("read upstream response: %w", err)
	}

	// Extract usage from the JSON response.
	var usage OpenAIUsage
	usage.InputTokens = int(gjson.GetBytes(respBody, "usage.input_tokens").Int())
	usage.OutputTokens = int(gjson.GetBytes(respBody, "usage.output_tokens").Int())
	usage.CacheCreationInputTokens = int(gjson.GetBytes(respBody, "usage.cache_creation_input_tokens").Int())
	usage.CacheReadInputTokens = int(gjson.GetBytes(respBody, "usage.cache_read_input_tokens").Int())

	// DeepSeek/Kimi Anthropic compat bug: see streaming handler comment.
	if usage.CacheReadInputTokens > 0 && usage.InputTokens < usage.CacheReadInputTokens {
		usage.InputTokens += usage.CacheReadInputTokens
	}

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
	originalModel, billingModel, upstreamModel string,
	startTime time.Time,
) (*OpenAIForwardResult, error) {
	var usage OpenAIUsage
	var lastMessageData []byte
	var responseID, requestID string

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
			usage.InputTokens = int(msg.Get("usage.input_tokens").Int())
			usage.CacheCreationInputTokens = int(msg.Get("usage.cache_creation_input_tokens").Int())
			usage.CacheReadInputTokens = int(msg.Get("usage.cache_read_input_tokens").Int())
			// DeepSeek/Kimi Anthropic compat bug: see streaming handler comment.
			if usage.CacheReadInputTokens > 0 && usage.InputTokens < usage.CacheReadInputTokens {
				usage.InputTokens += usage.CacheReadInputTokens
			}
			// Store the initial message object as the base for the final response.
			lastMessageData = []byte(msg.Raw)

		case "content_block_delta":
			// Content is assembled by the client from the streamed message;
			// for buffered mode we need to build it ourselves. For simplicity
			// we accumulate text from deltas.
			// (Full assembly would require tracking content blocks, but most
			// coding-agent flows only have a single text block.)

		case "message_delta":
			usage.OutputTokens = int(gjson.Get(data, "usage.output_tokens").Int())
		}
	}

	requestID = resp.Header.Get("x-request-id")

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
			msgResp["usage"] = anthropicUsage{
				InputTokens:              usage.InputTokens,
				OutputTokens:             usage.OutputTokens,
				CacheCreationInputTokens: usage.CacheCreationInputTokens,
				CacheReadInputTokens:     usage.CacheReadInputTokens,
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
