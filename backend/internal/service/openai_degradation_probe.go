package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/tidwall/gjson"

	cryptorand "crypto/rand"
)

// 降智探针：定期给 OpenAI API Key 中转号的主力模型出一道随机 12 位×12 位乘法，
// 依据「上游回报的模型名」与「推理 token 数」判断该号上的该模型是否被偷偷降级
// （换成弱模型/低推理档却报正确模型名）。判定降智后只暂停「该账号 × 该模型」
// （写 model_rate_limits，reason=degradation_probe），恢复后不再续期、自然到期放回。
//
// 判据只认「推理 token 过低」与「模型名不符」：真模型偶尔也会把 12 位乘法算错
// （推理充分但答案错），答错本身噪声太大，只记日志不计入判定。

const (
	degradationProbeReason     = "degradation_probe"
	degradationProbeExtraKey   = "degradation_probe"
	degradationProbeLeaderKey  = "leader:openai_degradation_probe"
	degradationProbeHistoryMax = 20
	degradationProbeMaxBody    = 4 << 20
	degradationProbeTimeout    = 3 * time.Minute
	degradationProbeWorkers    = 8

	degradationSampleGood     = 'G'
	degradationSampleDegraded = 'D'

	openAIDegradationProbeSettingsCacheTTL = 30 * time.Second
)

// OpenAIDegradationProbeSettings 降智探针配置（settings 表 openai_degradation_probe_settings，JSON）。
type OpenAIDegradationProbeSettings struct {
	Enabled         bool     `json:"enabled"`
	IntervalMinutes int      `json:"interval_minutes"`
	Models          []string `json:"models"`
	// Window 最近多少个有效样本参与「是否暂停」的判定；TripCount 为其中降智样本的触发数。
	Window    int `json:"window"`
	TripCount int `json:"trip_count"`
	// ModelTripCount 按模型覆盖 TripCount（各模型自身的低推理基线不同，如 gpt-6-sol 正常号也有 15~25%）。
	ModelTripCount map[string]int `json:"model_trip_count,omitempty"`
	// 已暂停时，最近 RecoverWindow 个样本中降智数 ≤ RecoverMaxDegraded 才恢复（迟滞，防抖动）。
	RecoverWindow int `json:"recover_window"`
	// RecoverMaxDegraded 为 nil 表示未配置（取默认 1）；显式 0 表示必须全部正常才恢复。
	RecoverMaxDegraded *int `json:"recover_max_degraded,omitempty"`
	LowReasoningTokens int  `json:"low_reasoning_tokens"`
	// AlertOnlyAccountIDs 中的账号只判定、记录、打日志，不暂停调度。
	AlertOnlyAccountIDs []int64 `json:"alert_only_account_ids,omitempty"`
}

func DefaultOpenAIDegradationProbeSettings() *OpenAIDegradationProbeSettings {
	return &OpenAIDegradationProbeSettings{
		Enabled:            false,
		IntervalMinutes:    10,
		Models:             []string{"gpt-6.1-sol", "gpt-6-sol", "gpt-6-astra", "gpt-5.6-terra"},
		Window:             10,
		TripCount:          5,
		ModelTripCount:     map[string]int{"gpt-6-sol": 7},
		RecoverWindow:      6,
		RecoverMaxDegraded: new(1),
		LowReasoningTokens: 60,
	}
}

type cachedOpenAIDegradationProbeSettings struct {
	settings  OpenAIDegradationProbeSettings
	expiresAt time.Time
}

func clampInt(v, lo, hi int) int {
	return max(lo, min(v, hi))
}

func normalizeOpenAIDegradationProbeSettings(settings *OpenAIDegradationProbeSettings) *OpenAIDegradationProbeSettings {
	def := DefaultOpenAIDegradationProbeSettings()
	if settings == nil {
		return def
	}
	result := *settings
	// 只允许整除 60 的间隔，cron 按分钟取模对齐。
	switch result.IntervalMinutes {
	case 5, 10, 15, 20, 30, 60:
	default:
		result.IntervalMinutes = def.IntervalMinutes
	}
	models := make([]string, 0, len(result.Models))
	for _, m := range result.Models {
		if m = strings.TrimSpace(m); m != "" && !slices.Contains(models, m) {
			models = append(models, m)
		}
	}
	if len(models) == 0 {
		models = def.Models
	}
	result.Models = models
	if result.Window <= 0 {
		result.Window = def.Window
	}
	result.Window = clampInt(result.Window, 3, degradationProbeHistoryMax)
	if result.TripCount <= 0 {
		result.TripCount = def.TripCount
	}
	result.TripCount = clampInt(result.TripCount, 1, result.Window)
	if result.ModelTripCount == nil {
		result.ModelTripCount = def.ModelTripCount
	}
	trips := make(map[string]int, len(result.ModelTripCount))
	for m, n := range result.ModelTripCount {
		if n > 0 {
			trips[strings.TrimSpace(m)] = clampInt(n, 1, result.Window)
		}
	}
	result.ModelTripCount = trips
	if result.RecoverWindow <= 0 {
		result.RecoverWindow = def.RecoverWindow
	}
	result.RecoverWindow = clampInt(result.RecoverWindow, 1, degradationProbeHistoryMax)
	recoverMax := *def.RecoverMaxDegraded
	if result.RecoverMaxDegraded != nil {
		recoverMax = *result.RecoverMaxDegraded
	}
	result.RecoverMaxDegraded = new(clampInt(recoverMax, 0, result.RecoverWindow-1))
	if result.LowReasoningTokens <= 0 {
		result.LowReasoningTokens = def.LowReasoningTokens
	}
	return &result
}

func (s *OpenAIDegradationProbeSettings) recoverMaxDegraded() int {
	if s.RecoverMaxDegraded == nil {
		return 1
	}
	return *s.RecoverMaxDegraded
}

func (s *OpenAIDegradationProbeSettings) tripCountFor(model string) int {
	if n, ok := s.ModelTripCount[model]; ok && n > 0 {
		return n
	}
	return s.TripCount
}

func (s *SettingService) GetOpenAIDegradationProbeSettings(ctx context.Context) (*OpenAIDegradationProbeSettings, error) {
	if s == nil || s.settingRepo == nil {
		return DefaultOpenAIDegradationProbeSettings(), nil
	}
	if cached, ok := s.openAIDegradationProbeCache.Load().(*cachedOpenAIDegradationProbeSettings); ok && cached != nil && time.Now().Before(cached.expiresAt) {
		result := cached.settings
		return &result, nil
	}

	settings := DefaultOpenAIDegradationProbeSettings()
	value, err := s.settingRepo.GetValue(ctx, SettingKeyOpenAIDegradationProbeSettings)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return nil, fmt.Errorf("get OpenAI degradation probe settings: %w", err)
	}
	if err == nil && strings.TrimSpace(value) != "" {
		var stored OpenAIDegradationProbeSettings
		if json.Unmarshal([]byte(value), &stored) == nil {
			settings = normalizeOpenAIDegradationProbeSettings(&stored)
		}
	}
	s.openAIDegradationProbeCache.Store(&cachedOpenAIDegradationProbeSettings{
		settings:  *settings,
		expiresAt: time.Now().Add(openAIDegradationProbeSettingsCacheTTL),
	})
	result := *settings
	return &result, nil
}

// degradationProbeSample 一次探测的观测结果。
type degradationProbeSample struct {
	ReturnedModel   string
	ReasoningTokens int64
	HasReasoning    bool
	Correct         bool
}

// classifyDegradationSample 判定单个样本：模型名不符或推理 token 过低即为降智。
// 上游没回 reasoning_tokens 时只凭模型名判断。
func classifyDegradationSample(sample degradationProbeSample, upstreamModel string, lowReasoning int) byte {
	if !degradationModelMatches(sample.ReturnedModel, upstreamModel) {
		return degradationSampleDegraded
	}
	// 推理 token 为 0 视同未回报：部分中转会把 usage 明细清零，按 0 判会把整个号误判降智。
	if sample.HasReasoning && sample.ReasoningTokens > 0 && sample.ReasoningTokens < int64(lowReasoning) {
		return degradationSampleDegraded
	}
	return degradationSampleGood
}

// degradationModelMatches 上游回报的模型名允许带日期/版本后缀（gpt-6-sol-2026-09-01），
// 但不能是别的模型（gpt-6-luna）。未回报模型名时不据此判降智。
func degradationModelMatches(returned, upstream string) bool {
	returned = strings.ToLower(strings.TrimSpace(returned))
	upstream = strings.ToLower(strings.TrimSpace(upstream))
	if returned == "" || upstream == "" || returned == upstream {
		return true
	}
	if rest, ok := strings.CutPrefix(returned, upstream+"-"); ok {
		// 只放行纯日期/数字版本后缀，避免 gpt-6-sol 匹配到 gpt-6-sol-mini 之类。
		return strings.Trim(rest, "0123456789-") == ""
	}
	return false
}

// degradationModelState 单个「账号 × 模型」的探针状态，持久化在 accounts.extra.degradation_probe[model]。
// 探针每轮可能由不同节点执行，状态不能只放内存。
type degradationModelState struct {
	History string `json:"h"`
	Paused  bool   `json:"paused"`
	At      string `json:"at"`
	Last    string `json:"last,omitempty"`
	// Held 判定为降智但因分组内没有替补而未暂停。
	Held bool `json:"held,omitempty"`
}

// nextDegradationState 追加样本并按迟滞规则决定是否暂停。
func nextDegradationState(prev degradationModelState, sample byte, settings *OpenAIDegradationProbeSettings, model string) degradationModelState {
	next := prev
	next.History += string(sample)
	if len(next.History) > degradationProbeHistoryMax {
		next.History = next.History[len(next.History)-degradationProbeHistoryMax:]
	}
	countDegraded := func(n int) (int, int) {
		h := next.History
		if len(h) > n {
			h = h[len(h)-n:]
		}
		return strings.Count(h, string(degradationSampleDegraded)), len(h)
	}
	if prev.Paused {
		degraded, total := countDegraded(settings.RecoverWindow)
		if total >= settings.RecoverWindow && degraded <= settings.recoverMaxDegraded() {
			next.Paused = false
		}
		return next
	}
	degraded, _ := countDegraded(settings.Window)
	if degraded >= settings.tripCountFor(model) {
		next.Paused = true
	}
	return next
}

func readDegradationProbeStates(account *Account) map[string]degradationModelState {
	states := map[string]degradationModelState{}
	if account == nil || account.Extra == nil {
		return states
	}
	raw, ok := account.Extra[degradationProbeExtraKey]
	if !ok || raw == nil {
		return states
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return states
	}
	_ = json.Unmarshal(data, &states)
	return states
}

// isDegradationProbeRateLimitEntry 报告 model_rate_limits 中的条目是否由降智探针写入。
// 这类条目由探针自己续期/到期，不能被「测试成功即恢复」等通用清理顺手抹掉。
func isDegradationProbeRateLimitEntry(entry any) bool {
	m, ok := entry.(map[string]any)
	if !ok {
		return false
	}
	reason, _ := m["reason"].(string)
	return reason == degradationProbeReason
}

func degradationProbeRandomOperand() *big.Int {
	lo := big.NewInt(100_000_000_000)
	span := big.NewInt(900_000_000_000)
	n, err := cryptorand.Int(cryptorand.Reader, span)
	if err != nil {
		n = big.NewInt(time.Now().UnixNano() % 900_000_000_000)
	}
	return n.Add(n, lo)
}

// runDegradationProbeSample 直连账号上游 /v1/responses（流式），出一道随机乘法题。
func (s *AccountTestService) runDegradationProbeSample(ctx context.Context, account *Account, upstreamModel string) (degradationProbeSample, error) {
	var sample degradationProbeSample
	credentialAccount := account
	if account.IsCredentialShadow() {
		resolved, err := resolveCredentialAccount(ctx, s.accountRepo, account)
		if err != nil {
			return sample, err
		}
		credentialAccount = resolved
	}
	apiKey := credentialAccount.GetOpenAIProtocolAPIKey()
	if apiKey == "" {
		return sample, errors.New("no api key")
	}
	baseURL := credentialAccount.GetOpenAIBaseURL()
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return sample, err
	}

	a, b := degradationProbeRandomOperand(), degradationProbeRandomOperand()
	want := new(big.Int).Mul(a, b).String()
	payload, _ := json.Marshal(map[string]any{
		"model":             upstreamModel,
		"input":             fmt.Sprintf("计算 %s × %s，只输出最终整数结果，不要任何其他文字。", a, b),
		"reasoning":         map[string]any{"effort": "medium"},
		"max_output_tokens": 8000,
		"stream":            true,
		"store":             false,
	})

	probeCtx, cancel := context.WithTimeout(ctx, degradationProbeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(probeCtx, http.MethodPost, buildOpenAIResponsesURL(normalizedBaseURL), bytes.NewReader(payload))
	if err != nil {
		return sample, err
	}
	req = req.WithContext(WithHTTPUpstreamProfile(req.Context(), HTTPUpstreamProfileOpenAI))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "text/event-stream")
	applyOpenAICodexProbeHeaders(req.Header)
	account.ApplyHeaderOverrides(req.Header)

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	resp, err := s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(account))
	if err != nil {
		return sample, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return sample, fmt.Errorf("upstream status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	final, err := readDegradationProbeFinalResponse(resp.Body)
	if err != nil {
		return sample, err
	}
	sample.ReturnedModel = gjson.GetBytes(final, "model").String()
	if rt := gjson.GetBytes(final, "usage.output_tokens_details.reasoning_tokens"); rt.Exists() {
		sample.HasReasoning = true
		sample.ReasoningTokens = rt.Int()
	}
	var text strings.Builder
	gjson.GetBytes(final, "output").ForEach(func(_, item gjson.Result) bool {
		if item.Get("type").String() == "message" {
			item.Get("content").ForEach(func(_, c gjson.Result) bool {
				text.WriteString(c.Get("text").String())
				return true
			})
		}
		return true
	})
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, text.String())
	sample.Correct = digits == want
	return sample, nil
}

// readDegradationProbeFinalResponse 读 Responses SSE，返回 response.completed/incomplete 携带的 response 对象。
// 也兼容上游忽略 stream 直接回整包 JSON 的情况。
func readDegradationProbeFinalResponse(body io.Reader) ([]byte, error) {
	reader := bufio.NewReaderSize(io.LimitReader(body, degradationProbeMaxBody), 64<<10)
	var plain bytes.Buffer
	sawEvent := false
	for {
		line, err := reader.ReadBytes('\n')
		trimmed := bytes.TrimSpace(line)
		if data, ok := bytes.CutPrefix(trimmed, []byte("data:")); ok {
			sawEvent = true
			data = bytes.TrimSpace(data)
			switch gjson.GetBytes(data, "type").String() {
			case "response.completed", "response.incomplete":
				return []byte(gjson.GetBytes(data, "response").Raw), nil
			case "response.failed", "error":
				msg := gjson.GetBytes(data, "response.error.message").String()
				if msg == "" {
					msg = gjson.GetBytes(data, "error.message").String()
				}
				if msg == "" {
					msg = gjson.GetBytes(data, "message").String()
				}
				return nil, fmt.Errorf("stream failed: %s", msg)
			}
		} else if !sawEvent {
			plain.Write(line)
		}
		if err != nil {
			if !sawEvent && gjson.ValidBytes(bytes.TrimSpace(plain.Bytes())) && gjson.GetBytes(plain.Bytes(), "output").Exists() {
				return bytes.TrimSpace(plain.Bytes()), nil
			}
			if errors.Is(err, io.EOF) {
				return nil, errors.New("stream ended without completed event")
			}
			return nil, err
		}
	}
}

// degradationProbeResult 某账号 × 某模型本轮的探测结果。
type degradationProbeResult struct {
	model         string
	upstreamModel string
	sample        degradationProbeSample
	err           error
}

// runDegradationProbe 一轮探测：先并行探测（每个账号内串行，部分中转按 key 限并发），
// 再串行逐个判定与暂停——判定必须串行，且要计入本轮已决定暂停的号，否则两个号同轮
// 同时降智时会互相以为对方还在、一起被停，把模型整个停掉。
func (s *GroupHealthCheckService) runDegradationProbe() {
	if s == nil || s.rateLimitSvc == nil || s.accountTestSvc == nil {
		return
	}
	settings, err := s.rateLimitSvc.settingService.GetOpenAIDegradationProbeSettings(context.Background())
	if err != nil || settings == nil || !settings.Enabled {
		return
	}
	if time.Now().Minute()%settings.IntervalMinutes != 0 {
		return
	}
	// 本轮必须在下一轮开始前结束：否则下一 tick 拿不到锁被跳过，暂停续期出现空档。
	roundBudget := time.Duration(settings.IntervalMinutes)*time.Minute - time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), roundBudget)
	defer cancel()
	release, ok := tryAcquireSingletonLeaderLock(ctx, s.lockCache, s.db, degradationProbeLeaderKey, s.instanceID, roundBudget)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}

	accounts, err := s.accountRepo.ListByPlatform(ctx, PlatformOpenAI)
	if err != nil {
		logger.LegacyPrintf("service.degradation_probe", "[DegradationProbe] list accounts failed: %v", err)
		return
	}
	targets := make([]*Account, 0, len(accounts))
	for i := range accounts {
		a := &accounts[i]
		if a.Type != AccountTypeAPIKey || !a.IsSchedulable() || !openai_compat.ShouldUseResponsesAPI(a.Extra) {
			continue
		}
		targets = append(targets, a)
	}

	results := make([][]degradationProbeResult, len(targets))
	sem := make(chan struct{}, degradationProbeWorkers)
	var wg sync.WaitGroup
	for idx, account := range targets {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, account *Account) {
			defer wg.Done()
			defer func() { <-sem }()
			for _, model := range settings.Models {
				if !account.IsModelSupported(model) {
					continue
				}
				upstreamModel := account.GetMappedModel(model)
				sample, err := s.accountTestSvc.runDegradationProbeSample(ctx, account, upstreamModel)
				results[idx] = append(results[idx], degradationProbeResult{model: model, upstreamModel: upstreamModel, sample: sample, err: err})
			}
		}(idx, account)
	}
	wg.Wait()

	// 判定阶段用独立 context：探测阶段可能已耗尽本轮预算，但续期/持久化不能因此丢失。
	decideCtx, decideCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer decideCancel()
	pausedThisRound := map[string]map[int64]bool{}
	for idx, account := range targets {
		s.decideAccountDegradation(decideCtx, account, results[idx], settings, pausedThisRound)
	}
}

func (s *GroupHealthCheckService) decideAccountDegradation(ctx context.Context, account *Account, results []degradationProbeResult, settings *OpenAIDegradationProbeSettings, pausedThisRound map[string]map[int64]bool) {
	if len(results) == 0 {
		return
	}
	states := readDegradationProbeStates(account)
	alertOnly := slices.Contains(settings.AlertOnlyAccountIDs, account.ID)
	pauseUntil := time.Now().Add(time.Duration(2*settings.IntervalMinutes+5) * time.Minute)

	for _, r := range results {
		prev := states[r.model]
		next := prev
		if r.err != nil {
			// 请求失败不计入判定（过载/限流由熔断与冷却处理），但已暂停的照常续期——
			// 中转往往在降智的同时也在过载，失败时不续期会让暂停过期、降智号重新接流量。
			logger.LegacyPrintf("service.degradation_probe", "[DegradationProbe] account=%d(%s) model=%s probe failed: %v", account.ID, account.Name, r.upstreamModel, r.err)
		} else {
			verdict := classifyDegradationSample(r.sample, r.upstreamModel, settings.LowReasoningTokens)
			next = nextDegradationState(prev, verdict, settings, r.model)
			next.At = time.Now().UTC().Format(time.RFC3339)
			next.Last = fmt.Sprintf("model=%s reasoning=%d correct=%v", r.sample.ReturnedModel, r.sample.ReasoningTokens, r.sample.Correct)
		}

		held := false
		if next.Paused && !alertOnly {
			// 每轮都重新检查替补：其他号之后出问题时，不能继续停着把模型整个停掉。
			if !s.degradationPauseLeavesAlternative(ctx, account, r.model, pausedThisRound) {
				held = true
			} else if err := s.pauseModelForDegradation(ctx, account, r.upstreamModel, pauseUntil); err != nil {
				logger.LegacyPrintf("service.degradation_probe", "[DegradationProbe] account=%d model=%s pause failed: %v", account.ID, r.upstreamModel, err)
			} else {
				if pausedThisRound[r.model] == nil {
					pausedThisRound[r.model] = map[int64]bool{}
				}
				pausedThisRound[r.model][account.ID] = true
			}
		}
		if next.Paused != prev.Paused || held != prev.Held {
			slog.Warn("openai_degradation_probe_state_changed",
				"account_id", account.ID, "account_name", account.Name, "model", r.model,
				"paused", next.Paused, "held_last_account", held, "alert_only", alertOnly,
				"history", next.History, "last", next.Last)
		}
		next.Held = held
		states[r.model] = next
	}
	if err := s.accountRepo.UpdateExtra(ctx, account.ID, map[string]any{degradationProbeExtraKey: states}); err != nil {
		logger.LegacyPrintf("service.degradation_probe", "[DegradationProbe] account=%d persist state failed: %v", account.ID, err)
	}
}

// pauseModelForDegradation 写入/续期探针暂停。已有更晚到期的非探针限流（如上游真实 429）时不覆盖：
// 那样会把真实限流提前结束，还会给它打上探针标记、让通用清理不再清它。
func (s *GroupHealthCheckService) pauseModelForDegradation(ctx context.Context, account *Account, upstreamModel string, until time.Time) error {
	if limits, ok := account.Extra[modelRateLimitsKey].(map[string]any); ok {
		if entry, ok := limits[upstreamModel]; ok && !isDegradationProbeRateLimitEntry(entry) {
			if resetAt := account.modelRateLimitResetAt(upstreamModel); resetAt != nil && resetAt.After(until) {
				return nil
			}
		}
	}
	return s.accountRepo.SetModelRateLimit(ctx, account.ID, upstreamModel, until, degradationProbeReason)
}

// degradationPauseLeavesAlternative 报告账号所在的每个分组里，是否都还有其他能服务该模型的
// 可调度账号（本轮已决定暂停的号不算）。
func (s *GroupHealthCheckService) degradationPauseLeavesAlternative(ctx context.Context, account *Account, model string, pausedThisRound map[string]map[int64]bool) bool {
	if len(account.GroupIDs) == 0 {
		return true
	}
	for _, gid := range account.GroupIDs {
		peers, err := s.accountRepo.ListSchedulableByGroupIDAndPlatform(ctx, gid, account.Platform)
		if err != nil {
			return false
		}
		found := false
		for i := range peers {
			p := &peers[i]
			if p.ID == account.ID || pausedThisRound[model][p.ID] {
				continue
			}
			if p.IsModelSupported(model) && p.IsSchedulableForModelWithContext(ctx, model) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
