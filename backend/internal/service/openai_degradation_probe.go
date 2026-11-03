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
	"math"
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
	// RelativeMargin 相对判定：账号的降智比例须比同模型表现最好的号高出至少该值才暂停。
	// 降智也可能是官方行为——同模型所有号都降智时差距拉不开，一律不停（已停的放回）。
	// nil 取默认 0.3。
	RelativeMargin *float64 `json:"relative_margin,omitempty"`

	// 首字异常慢：最近 SlowTTFTWindowMinutes 分钟真实流量里，某号某模型的首字中位数
	// ≥ 同模型其他号中位数的 SlowTTFTRatio 倍且多出 SlowTTFTMinGapSeconds 秒，暂停该号该模型。
	SlowTTFTDisabled      bool    `json:"slow_ttft_disabled,omitempty"`
	SlowTTFTWindowMinutes int     `json:"slow_ttft_window_minutes,omitempty"`
	SlowTTFTMinRequests   int     `json:"slow_ttft_min_requests,omitempty"`
	SlowTTFTRatio         float64 `json:"slow_ttft_ratio,omitempty"`
	SlowTTFTMinGapSeconds int     `json:"slow_ttft_min_gap_seconds,omitempty"`
}

func DefaultOpenAIDegradationProbeSettings() *OpenAIDegradationProbeSettings {
	return &OpenAIDegradationProbeSettings{
		Enabled:            false,
		IntervalMinutes:    10,
		Models:             []string{"gpt-6.1-sol", "gpt-6-sol", "gpt-6-astra", "gpt-5.6-terra"},
		Window:             10,
		TripCount:          6,
		ModelTripCount:     map[string]int{"gpt-6-sol": 8},
		RecoverWindow:      6,
		RecoverMaxDegraded: new(1),
		LowReasoningTokens: 60,
		RelativeMargin:     new(0.3),

		SlowTTFTWindowMinutes: 30,
		SlowTTFTMinRequests:   8,
		SlowTTFTRatio:         2,
		SlowTTFTMinGapSeconds: 10,
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
	margin := *def.RelativeMargin
	if result.RelativeMargin != nil {
		margin = *result.RelativeMargin
	}
	result.RelativeMargin = new(math.Max(0, math.Min(margin, 1)))
	if result.SlowTTFTWindowMinutes <= 0 {
		result.SlowTTFTWindowMinutes = def.SlowTTFTWindowMinutes
	}
	result.SlowTTFTWindowMinutes = clampInt(result.SlowTTFTWindowMinutes, 10, 120)
	if result.SlowTTFTMinRequests <= 0 {
		result.SlowTTFTMinRequests = def.SlowTTFTMinRequests
	}
	if result.SlowTTFTRatio < 1.2 {
		result.SlowTTFTRatio = def.SlowTTFTRatio
	}
	if result.SlowTTFTMinGapSeconds <= 0 {
		result.SlowTTFTMinGapSeconds = def.SlowTTFTMinGapSeconds
	}
	return &result
}

func (s *OpenAIDegradationProbeSettings) relativeMargin() float64 {
	if s.RelativeMargin == nil {
		return 0.3
	}
	return *s.RelativeMargin
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
	// Held 判定为需暂停但因分组内没有替补而未暂停。
	Held bool `json:"held,omitempty"`
	// Effective 本轮是否实际处于暂停（Paused 只是自身迟滞状态，还要叠加相对判定/首字慢/保底）。
	Effective bool `json:"effective,omitempty"`
	SlowTTFT  bool `json:"slow_ttft,omitempty"`
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
	return reason == degradationProbeReason || strings.HasPrefix(reason, degradationProbeReason+":")
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

// degradationDecision 判定阶段某账号 × 某模型的中间结果。
type degradationDecision struct {
	result degradationProbeResult
	prev   degradationModelState
	next   degradationModelState
}

const (
	degradationPauseReasonSlowTTFT = degradationProbeReason + ":slow_first_token"
	// degradationSlowTTFTMinKeys 首字比较要求的最少不同 API Key 数：首字主要取决于上下文长度与推理档，
	// 粘性会话会把单个重度长上下文用户固定在一个号上，只看中位数会把「负载重」误判成「号慢」。
	degradationSlowTTFTMinKeys = 3
	// degradationRelativeMinSamples 参与「同模型最好号」比较的最少样本数。
	degradationRelativeMinSamples = 5
)

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
	slow := map[degradationTTFTKey]bool{}
	if !settings.SlowTTFTDisabled {
		slow = s.loadSlowFirstTokenAccounts(decideCtx, settings)
	}
	s.decideDegradationRound(decideCtx, targets, results, settings, slow)
}

// decideDegradationRound 串行判定一轮：先更新每个「账号 × 模型」的历史，再按同模型最好号
// 做相对判定（全部降智视为官方行为、不停），叠加首字异常慢，最后按上游模型合并结论、
// 逐个写入/解除暂停。
func (s *GroupHealthCheckService) decideDegradationRound(ctx context.Context, targets []*Account, results [][]degradationProbeResult, settings *OpenAIDegradationProbeSettings, slow map[degradationTTFTKey]bool) {
	decisions := make([][]degradationDecision, len(targets))
	bestRate := map[string]float64{}
	for idx, account := range targets {
		states := readDegradationProbeStates(account)
		for _, r := range results[idx] {
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
				// 参照线只取本轮探测成功的号：长期失败的号带着旧的干净历史会把参照线压低，
				// 让「全部降智（官方行为）」时所有号都显得突出、被一路停到只剩保底。
				if rate, ok := degradationRecentRate(next.History, settings.Window); ok {
					if best, seen := bestRate[r.model]; !seen || rate < best {
						bestRate[r.model] = rate
					}
				}
			}
			decisions[idx] = append(decisions[idx], degradationDecision{result: r, prev: prev, next: next})
		}
	}

	pausedThisRound := map[string]map[int64]bool{}
	for idx, account := range targets {
		if len(decisions[idx]) == 0 {
			continue
		}
		// 探测阶段可能持续数分钟，期间线上流量可能已对同一模型写入真实限流；暂停/解除前
		// 重读账号，避免用旧快照把真实限流当成探针条目覆盖或解除。
		fresh := account
		if latest, err := s.accountRepo.GetByID(ctx, account.ID); err == nil && latest != nil {
			fresh = latest
		}
		states := readDegradationProbeStates(account)
		alertOnly := slices.Contains(settings.AlertOnlyAccountIDs, account.ID)

		// 多个请求模型可能映射到同一上游模型，暂停/解除都作用在上游 key 上：先合并结论，
		// 否则一个要停、一个要解，会每轮来回翻转。
		type upstreamVerdict struct {
			reason string
			model  string
		}
		verdicts := map[string]*upstreamVerdict{}
		order := []string{}
		for i := range decisions[idx] {
			d := &decisions[idx][i]
			margin := settings.relativeMargin()
			if d.prev.Effective && !d.prev.SlowTTFT {
				// 已在停的放宽留存门槛，避免参照线小幅波动导致停/放来回切。
				margin = math.Max(0, margin-0.1)
			}
			degraded := d.next.Paused && degradationStandsOut(d.next.History, settings.Window, margin, bestRate, d.result.model)
			slowTTFT := slow[degradationTTFTKey{accountID: account.ID, upstreamModel: d.result.upstreamModel}]
			d.next.SlowTTFT = slowTTFT
			d.next.Effective = degraded // 暂存，下方按上游合并后改写为实际结果
			v := verdicts[d.result.upstreamModel]
			if v == nil {
				v = &upstreamVerdict{model: d.result.model}
				verdicts[d.result.upstreamModel] = v
				order = append(order, d.result.upstreamModel)
			}
			switch {
			case degraded:
				v.reason, v.model = degradationProbeReason, d.result.model
			case slowTTFT && v.reason == "":
				v.reason, v.model = degradationPauseReasonSlowTTFT, d.result.model
			}
		}

		applied := map[string]struct{ effective, held bool }{}
		for _, upstream := range order {
			v := verdicts[upstream]
			held := false
			if v.reason != "" && !alertOnly {
				// 每轮都重新检查替补：其他号之后出问题时，不能继续停着把模型整个停掉。
				if !s.degradationPauseLeavesAlternative(ctx, account, v.model, pausedThisRound) {
					held = true
				} else if err := s.pauseModelForDegradation(ctx, fresh, upstream, s.degradationPauseUntil(settings, v.reason), v.reason); err != nil {
					logger.LegacyPrintf("service.degradation_probe", "[DegradationProbe] account=%d model=%s pause failed: %v", account.ID, upstream, err)
				} else {
					if pausedThisRound[v.model] == nil {
						pausedThisRound[v.model] = map[int64]bool{}
					}
					pausedThisRound[v.model][account.ID] = true
				}
			}
			effective := v.reason != "" && !held && !alertOnly
			if !effective {
				s.liftDegradationPause(ctx, fresh, upstream)
			}
			applied[upstream] = struct{ effective, held bool }{effective, held}
		}

		for _, d := range decisions[idx] {
			next := d.next
			degradedVsBest := next.Effective
			ap := applied[d.result.upstreamModel]
			next.Effective, next.Held = ap.effective, ap.held
			if next.Effective != d.prev.Effective || next.Held != d.prev.Held || next.Paused != d.prev.Paused {
				slog.Warn("openai_degradation_probe_state_changed",
					"account_id", account.ID, "account_name", account.Name, "model", d.result.model,
					"tripped", next.Paused, "degraded_vs_best", degradedVsBest, "slow_first_token", next.SlowTTFT,
					"effective_pause", next.Effective, "held_last_account", next.Held, "alert_only", alertOnly,
					"history", next.History, "last", next.Last)
			}
			states[d.result.model] = next
		}
		if err := s.accountRepo.UpdateExtra(ctx, account.ID, map[string]any{degradationProbeExtraKey: states}); err != nil {
			logger.LegacyPrintf("service.degradation_probe", "[DegradationProbe] account=%d persist state failed: %v", account.ID, err)
		}
	}
}

// degradationRecentRate 最近 window 个样本的降智比例；样本不足时 ok=false。
func degradationRecentRate(history string, window int) (float64, bool) {
	if len(history) > window {
		history = history[len(history)-window:]
	}
	if len(history) < degradationRelativeMinSamples {
		return 0, false
	}
	return float64(strings.Count(history, string(degradationSampleDegraded))) / float64(len(history)), true
}

// degradationStandsOut 相对判定：降智比例须比同模型表现最好的号高出 RelativeMargin。
// 同模型只有自己有足够样本时无从比较，视同「所有号都降智」，不停。
func degradationStandsOut(history string, window int, margin float64, bestRate map[string]float64, model string) bool {
	rate, ok := degradationRecentRate(history, window)
	if !ok {
		return false
	}
	best, seen := bestRate[model]
	if !seen {
		return false
	}
	return rate-best >= margin-1e-9
}

func (s *GroupHealthCheckService) degradationPauseUntil(settings *OpenAIDegradationProbeSettings, reason string) time.Time {
	d := time.Duration(2*settings.IntervalMinutes+5) * time.Minute
	if reason == degradationPauseReasonSlowTTFT {
		// 首字慢按真实流量判定：暂停期间该号没有流量，窗口内旧数据老化后自然解除并重新评估。
		if w := time.Duration(settings.SlowTTFTWindowMinutes) * time.Minute; w > d {
			d = w
		}
	}
	return time.Now().Add(d)
}

// pauseModelForDegradation 写入/续期探针暂停。已有更晚到期的非探针限流（如上游真实 429）时不覆盖：
// 那样会把真实限流提前结束，还会给它打上探针标记、让通用清理不再清它。
func (s *GroupHealthCheckService) pauseModelForDegradation(ctx context.Context, account *Account, upstreamModel string, until time.Time, reason string) error {
	if limits, ok := account.Extra[modelRateLimitsKey].(map[string]any); ok {
		if entry, ok := limits[upstreamModel]; ok && !isDegradationProbeRateLimitEntry(entry) {
			if resetAt := account.modelRateLimitResetAt(upstreamModel); resetAt != nil && resetAt.After(until) {
				return nil
			}
		}
	}
	return s.accountRepo.SetModelRateLimit(ctx, account.ID, upstreamModel, until, reason)
}

// liftDegradationPause 立即解除仍在生效的探针暂停（恢复、被判为官方行为、保底、改为只告警时）。
// 只动探针自己写的条目；写一个已到期的时间即可让调度立刻放行。
func (s *GroupHealthCheckService) liftDegradationPause(ctx context.Context, account *Account, upstreamModel string) {
	limits, ok := account.Extra[modelRateLimitsKey].(map[string]any)
	if !ok {
		return
	}
	entry, ok := limits[upstreamModel]
	if !ok || !isDegradationProbeRateLimitEntry(entry) || !account.isRateLimitActiveForKey(upstreamModel) {
		return
	}
	if err := s.accountRepo.SetModelRateLimit(ctx, account.ID, upstreamModel, time.Now(), degradationProbeReason); err != nil {
		logger.LegacyPrintf("service.degradation_probe", "[DegradationProbe] account=%d model=%s lift failed: %v", account.ID, upstreamModel, err)
	}
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

type degradationTTFTKey struct {
	accountID     int64
	upstreamModel string
}

type degradationTTFTStat struct {
	key      degradationTTFTKey
	medianMs float64
}

// loadSlowFirstTokenAccounts 用最近一段时间的真实流量找出首字明显异常慢的「账号 × 模型」。
func (s *GroupHealthCheckService) loadSlowFirstTokenAccounts(ctx context.Context, settings *OpenAIDegradationProbeSettings) map[degradationTTFTKey]bool {
	if s.db == nil {
		return map[degradationTTFTKey]bool{}
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT account_id, upstream_model, percentile_cont(0.5) WITHIN GROUP (ORDER BY first_token_ms)
		FROM usage_logs
		WHERE created_at > NOW() - make_interval(mins => $1)
		  AND stream AND first_token_ms > 0 AND upstream_model IS NOT NULL AND upstream_model <> ''
		  AND account_id IN (SELECT id FROM accounts WHERE platform = $2 AND type = $3 AND deleted_at IS NULL)
		GROUP BY account_id, upstream_model
		HAVING count(*) >= $4 AND count(DISTINCT api_key_id) >= $5`,
		settings.SlowTTFTWindowMinutes, PlatformOpenAI, AccountTypeAPIKey, settings.SlowTTFTMinRequests, degradationSlowTTFTMinKeys)
	if err != nil {
		logger.LegacyPrintf("service.degradation_probe", "[DegradationProbe] load first-token stats failed: %v", err)
		return map[degradationTTFTKey]bool{}
	}
	defer func() { _ = rows.Close() }()
	var stats []degradationTTFTStat
	for rows.Next() {
		var st degradationTTFTStat
		if err := rows.Scan(&st.key.accountID, &st.key.upstreamModel, &st.medianMs); err != nil {
			continue
		}
		stats = append(stats, st)
	}
	return detectSlowFirstToken(stats, settings.SlowTTFTRatio, float64(settings.SlowTTFTMinGapSeconds)*1000)
}

// detectSlowFirstToken 同一上游模型下，某号首字中位数 ≥ 其他号中位数的中位数 × ratio 且多出 minGapMs 即判慢。
// 同模型只有一个号有数据时无从比较，不判。所有号一起慢时参照系也慢，不会误停。
func detectSlowFirstToken(stats []degradationTTFTStat, ratio, minGapMs float64) map[degradationTTFTKey]bool {
	byModel := map[string][]degradationTTFTStat{}
	for _, st := range stats {
		byModel[st.key.upstreamModel] = append(byModel[st.key.upstreamModel], st)
	}
	slow := map[degradationTTFTKey]bool{}
	for _, group := range byModel {
		if len(group) < 2 {
			continue
		}
		for _, st := range group {
			others := make([]float64, 0, len(group)-1)
			for _, o := range group {
				if o.key != st.key {
					others = append(others, o.medianMs)
				}
			}
			slices.Sort(others)
			ref := others[len(others)/2]
			if len(others)%2 == 0 {
				ref = (others[len(others)/2-1] + others[len(others)/2]) / 2
			}
			if st.medianMs >= ref*ratio && st.medianMs-ref >= minGapMs {
				slow[st.key] = true
			}
		}
	}
	return slow
}
