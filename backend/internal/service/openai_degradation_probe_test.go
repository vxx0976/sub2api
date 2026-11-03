//go:build unit

package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDegradationModelMatches(t *testing.T) {
	require.True(t, degradationModelMatches("gpt-6-sol", "gpt-6-sol"))
	require.True(t, degradationModelMatches("GPT-6-Sol", "gpt-6-sol"))
	require.True(t, degradationModelMatches("gpt-6-sol-2026-09-01", "gpt-6-sol"))
	require.True(t, degradationModelMatches("", "gpt-6-sol"), "未回报模型名时不据此判降智")
	require.False(t, degradationModelMatches("gpt-6-luna", "gpt-6-sol"))
	require.False(t, degradationModelMatches("gpt-6-sol-mini", "gpt-6-sol"))
	require.False(t, degradationModelMatches("gpt-6", "gpt-6-sol"))
}

func TestClassifyDegradationSample(t *testing.T) {
	good := degradationProbeSample{ReturnedModel: "gpt-6.1-sol", ReasoningTokens: 250, HasReasoning: true, Correct: true}
	require.Equal(t, byte(degradationSampleGood), classifyDegradationSample(good, "gpt-6.1-sol", 60))

	wrongButReasoned := good
	wrongButReasoned.Correct = false
	require.Equal(t, byte(degradationSampleGood), classifyDegradationSample(wrongButReasoned, "gpt-6.1-sol", 60), "答错但推理充分不计降智")

	low := good
	low.ReasoningTokens = 18
	require.Equal(t, byte(degradationSampleDegraded), classifyDegradationSample(low, "gpt-6.1-sol", 60))

	swapped := good
	swapped.ReturnedModel = "gpt-6-luna"
	require.Equal(t, byte(degradationSampleDegraded), classifyDegradationSample(swapped, "gpt-6.1-sol", 60))

	zeroed := degradationProbeSample{ReturnedModel: "gpt-6.1-sol", HasReasoning: true, ReasoningTokens: 0}
	require.Equal(t, byte(degradationSampleGood), classifyDegradationSample(zeroed, "gpt-6.1-sol", 60), "推理 token 为 0 视同中转未回报")

	noUsage := degradationProbeSample{ReturnedModel: "gpt-6.1-sol"}
	require.Equal(t, byte(degradationSampleGood), classifyDegradationSample(noUsage, "gpt-6.1-sol", 60))
}

func feedDegradation(t *testing.T, state degradationModelState, seq string, settings *OpenAIDegradationProbeSettings, model string) degradationModelState {
	t.Helper()
	for i := 0; i < len(seq); i++ {
		state = nextDegradationState(state, seq[i], settings, model)
	}
	return state
}

func TestNextDegradationStateTripAndRecover(t *testing.T) {
	settings := normalizeOpenAIDegradationProbeSettings(&OpenAIDegradationProbeSettings{})
	require.Equal(t, 10, settings.Window)
	require.Equal(t, 6, settings.TripCount)

	// 5 次降智不触发，第 6 次触发。
	state := feedDegradation(t, degradationModelState{}, "GDGDGDGDGD", settings, "gpt-6.1-sol")
	require.False(t, state.Paused)
	state = feedDegradation(t, state, "D", settings, "gpt-6.1-sol")
	require.True(t, state.Paused)

	// 暂停后：最近 6 个样本里降智 ≤1 才恢复。
	state = feedDegradation(t, state, "GGGG", settings, "gpt-6.1-sol")
	require.True(t, state.Paused, "最近 6 个里还有 2 个降智")
	state = feedDegradation(t, state, "G", settings, "gpt-6.1-sol")
	require.False(t, state.Paused)

	// 恢复后不会因窗口里残留的旧降智立即再次暂停（需要重新累计到阈值）。
	state = feedDegradation(t, state, "G", settings, "gpt-6.1-sol")
	require.False(t, state.Paused)
}

func TestNextDegradationStatePerModelTripCount(t *testing.T) {
	settings := normalizeOpenAIDegradationProbeSettings(&OpenAIDegradationProbeSettings{})
	require.Equal(t, 8, settings.tripCountFor("gpt-6-sol"))
	state := feedDegradation(t, degradationModelState{}, "DDDDDDDGGG", settings, "gpt-6-sol")
	require.False(t, state.Paused, "gpt-6-sol 基线偏高，7/10 不触发")
	state = feedDegradation(t, degradationModelState{}, "DDDDDDDDGG", settings, "gpt-6-sol")
	require.True(t, state.Paused)
}

func TestNextDegradationStateHistoryBounded(t *testing.T) {
	settings := normalizeOpenAIDegradationProbeSettings(&OpenAIDegradationProbeSettings{})
	state := feedDegradation(t, degradationModelState{}, strings.Repeat("G", 50), settings, "gpt-6.1-sol")
	require.Len(t, state.History, degradationProbeHistoryMax)
}

func TestNormalizeOpenAIDegradationProbeSettings(t *testing.T) {
	s := normalizeOpenAIDegradationProbeSettings(&OpenAIDegradationProbeSettings{
		IntervalMinutes:    7,
		Models:             []string{" gpt-6-sol ", "", "gpt-6-sol"},
		Window:             100,
		TripCount:          99,
		RecoverWindow:      3,
		RecoverMaxDegraded: new(9),
	})
	require.Equal(t, 10, s.IntervalMinutes)
	require.Equal(t, []string{"gpt-6-sol"}, s.Models)
	require.Equal(t, degradationProbeHistoryMax, s.Window)
	require.Equal(t, s.Window, s.TripCount)
	require.Equal(t, 2, s.recoverMaxDegraded())
	require.Equal(t, 0, normalizeOpenAIDegradationProbeSettings(&OpenAIDegradationProbeSettings{RecoverMaxDegraded: new(0)}).recoverMaxDegraded(), "显式 0 不被默认值覆盖")
	require.Equal(t, 1, normalizeOpenAIDegradationProbeSettings(&OpenAIDegradationProbeSettings{}).recoverMaxDegraded())
	require.Equal(t, 60, s.LowReasoningTokens)
}

func TestReadDegradationProbeFinalResponse(t *testing.T) {
	sse := "event: response.created\ndata: {\"type\":\"response.created\"}\n\n" +
		": keepalive\n\n" +
		"data: {\"type\":\"response.completed\",\"response\":{\"model\":\"gpt-6-sol\",\"usage\":{\"output_tokens_details\":{\"reasoning_tokens\":321}}}}\n\n"
	final, err := readDegradationProbeFinalResponse(strings.NewReader(sse))
	require.NoError(t, err)
	require.Contains(t, string(final), "\"reasoning_tokens\":321")

	_, err = readDegradationProbeFinalResponse(strings.NewReader("data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"message\":\"overloaded\"}}}\n\n"))
	require.ErrorContains(t, err, "overloaded")

	_, err = readDegradationProbeFinalResponse(strings.NewReader("data: {\"type\":\"response.created\"}\n\n"))
	require.Error(t, err)

	final, err = readDegradationProbeFinalResponse(strings.NewReader(`{"model":"gpt-6-sol","output":[]}`))
	require.NoError(t, err, "上游忽略 stream 回整包 JSON")
	require.Contains(t, string(final), "gpt-6-sol")
}

func TestHasRecoverableRuntimeStateIgnoresDegradationProbeEntries(t *testing.T) {
	probeOnly := &Account{Extra: map[string]any{modelRateLimitsKey: map[string]any{
		"gpt-6-sol": map[string]any{"reason": degradationProbeReason, "rate_limit_reset_at": "2099-01-01T00:00:00Z"},
	}}}
	require.False(t, hasRecoverableRuntimeState(probeOnly))

	mixed := &Account{Extra: map[string]any{modelRateLimitsKey: map[string]any{
		"gpt-6-sol":    map[string]any{"reason": degradationProbeReason},
		"gpt-5.6-luna": map[string]any{"reason": "upstream_404_model_not_found"},
	}}}
	require.True(t, hasRecoverableRuntimeState(mixed))
}

func TestReadDegradationProbeStatesRoundTrip(t *testing.T) {
	account := &Account{Extra: map[string]any{degradationProbeExtraKey: map[string]any{
		"gpt-6-sol": map[string]any{"h": "GGD", "paused": true, "at": "2026-11-03T00:00:00Z"},
	}}}
	states := readDegradationProbeStates(account)
	require.Equal(t, "GGD", states["gpt-6-sol"].History)
	require.True(t, states["gpt-6-sol"].Paused)
	require.Empty(t, readDegradationProbeStates(&Account{}))
}

type degradationProbeRepoStub struct {
	AccountRepository
	peers  map[int64][]Account
	paused map[int64][]string
	lifted map[int64][]string
	extra  map[int64]map[string]any
	fresh  map[int64]*Account
}

func newDegradationProbeRepoStub(peers map[int64][]Account) *degradationProbeRepoStub {
	return &degradationProbeRepoStub{peers: peers, paused: map[int64][]string{}, lifted: map[int64][]string{}, extra: map[int64]map[string]any{}}
}

func (r *degradationProbeRepoStub) ListSchedulableByGroupIDAndPlatform(_ context.Context, gid int64, _ string) ([]Account, error) {
	return r.peers[gid], nil
}

func (r *degradationProbeRepoStub) SetModelRateLimit(_ context.Context, id int64, scope string, resetAt time.Time, _ ...string) error {
	if resetAt.After(time.Now().Add(time.Minute)) {
		r.paused[id] = append(r.paused[id], scope)
	} else {
		r.lifted[id] = append(r.lifted[id], scope)
	}
	return nil
}

func (r *degradationProbeRepoStub) GetByID(_ context.Context, id int64) (*Account, error) {
	if a, ok := r.fresh[id]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("not found")
}

func (r *degradationProbeRepoStub) UpdateExtra(_ context.Context, id int64, updates map[string]any) error {
	r.extra[id] = updates
	return nil
}

func degradationTestAccount(id int64, gid int64, history string) *Account {
	a := &Account{ID: id, Name: fmt.Sprintf("acc-%d", id), Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Status: StatusActive, Schedulable: true, GroupIDs: []int64{gid},
		Credentials: map[string]any{"model_mapping": map[string]any{"gpt-6.1-sol": "gpt-6.1-sol"}}}
	if history != "" {
		a.Extra = map[string]any{degradationProbeExtraKey: map[string]any{"gpt-6.1-sol": map[string]any{"h": history}}}
	}
	return a
}

func degradationResult(degraded bool) []degradationProbeResult {
	sample := degradationProbeSample{ReturnedModel: "gpt-6.1-sol", HasReasoning: true, ReasoningTokens: 300, Correct: true}
	if degraded {
		sample.ReasoningTokens = 15
	}
	return []degradationProbeResult{{model: "gpt-6.1-sol", upstreamModel: "gpt-6.1-sol", sample: sample}}
}

func degradationPeers(accounts ...*Account) []Account {
	out := make([]Account, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, *a)
	}
	return out
}

func runDegradationRound(svc *GroupHealthCheckService, targets []*Account, degraded []bool, slow map[degradationTTFTKey]bool) {
	results := make([][]degradationProbeResult, len(targets))
	for i := range targets {
		results[i] = degradationResult(degraded[i])
	}
	if slow == nil {
		slow = map[degradationTTFTKey]bool{}
	}
	svc.decideDegradationRound(context.Background(), targets, results, normalizeOpenAIDegradationProbeSettings(nil), slow)
}

func TestDecideDegradationPausesOnlyOutliers(t *testing.T) {
	a := degradationTestAccount(1, 29, "DDDDDDDDD")
	b := degradationTestAccount(2, 29, "GGGGGGGGG")
	c := degradationTestAccount(3, 29, "GGGGGGGGG")
	repo := newDegradationProbeRepoStub(map[int64][]Account{29: degradationPeers(a, b, c)})
	svc := &GroupHealthCheckService{accountRepo: repo}
	runDegradationRound(svc, []*Account{a, b, c}, []bool{true, false, false}, nil)
	require.Equal(t, []string{"gpt-6.1-sol"}, repo.paused[1])
	require.Empty(t, repo.paused[2])
	require.Empty(t, repo.paused[3])
}

func TestDecideDegradationAllDegradedTreatedAsOfficial(t *testing.T) {
	a := degradationTestAccount(1, 29, "DDDDDDDDD")
	b := degradationTestAccount(2, 29, "DDDDDDDGD")
	// a 上一轮已被探针暂停：全部降智被判为官方行为后应立即解除。
	a.Extra[modelRateLimitsKey] = map[string]any{"gpt-6.1-sol": map[string]any{
		"reason": degradationProbeReason, "rate_limit_reset_at": time.Now().Add(20 * time.Minute).UTC().Format(time.RFC3339)}}
	repo := newDegradationProbeRepoStub(map[int64][]Account{29: degradationPeers(a, b)})
	svc := &GroupHealthCheckService{accountRepo: repo}
	runDegradationRound(svc, []*Account{a, b}, []bool{true, true}, nil)
	require.Empty(t, repo.paused[1])
	require.Empty(t, repo.paused[2])
	require.Equal(t, []string{"gpt-6.1-sol"}, repo.lifted[1], "已有的探针暂停被解除")
}

func TestDecideDegradationNeverPausesLastAccountInSameRound(t *testing.T) {
	// a、b 在分组 29 里只有彼此；c 在别的分组里提供「表现最好」的参照。
	a := degradationTestAccount(1, 29, "DDDDDDDDD")
	b := degradationTestAccount(2, 29, "DDDDDDDDD")
	c := degradationTestAccount(3, 30, "GGGGGGGGG")
	repo := newDegradationProbeRepoStub(map[int64][]Account{29: degradationPeers(a, b), 30: degradationPeers(c)})
	svc := &GroupHealthCheckService{accountRepo: repo}
	runDegradationRound(svc, []*Account{a, b, c}, []bool{true, true, false}, nil)
	require.Equal(t, []string{"gpt-6.1-sol"}, repo.paused[1], "第一个号可以停")
	require.Empty(t, repo.paused[2], "第二个号是分组里最后的替补，不能再停")
	st := repo.extra[2][degradationProbeExtraKey].(map[string]degradationModelState)["gpt-6.1-sol"]
	require.True(t, st.Held)
	require.False(t, st.Effective)
}

func TestDecideDegradationRenewsPauseWhenProbeErrors(t *testing.T) {
	a := degradationTestAccount(1, 29, "DDDDDDDDDD")
	a.Extra[degradationProbeExtraKey] = map[string]any{"gpt-6.1-sol": map[string]any{"h": "DDDDDDDDDD", "paused": true}}
	b := degradationTestAccount(2, 29, "GGGGGGGGGG")
	repo := newDegradationProbeRepoStub(map[int64][]Account{29: degradationPeers(a, b)})
	svc := &GroupHealthCheckService{accountRepo: repo}
	results := [][]degradationProbeResult{
		{{model: "gpt-6.1-sol", upstreamModel: "gpt-6.1-sol", err: fmt.Errorf("overloaded")}},
		degradationResult(false),
	}
	svc.decideDegradationRound(context.Background(), []*Account{a, b}, results, normalizeOpenAIDegradationProbeSettings(nil), map[degradationTTFTKey]bool{})
	require.Equal(t, []string{"gpt-6.1-sol"}, repo.paused[1], "探测失败时已暂停的照常续期")
}

func TestDecideDegradationPausesSlowFirstToken(t *testing.T) {
	a := degradationTestAccount(1, 29, "GGGGGGGGG")
	b := degradationTestAccount(2, 29, "GGGGGGGGG")
	repo := newDegradationProbeRepoStub(map[int64][]Account{29: degradationPeers(a, b)})
	svc := &GroupHealthCheckService{accountRepo: repo}
	runDegradationRound(svc, []*Account{a, b}, []bool{false, false}, map[degradationTTFTKey]bool{{accountID: 1, upstreamModel: "gpt-6.1-sol"}: true})
	require.Equal(t, []string{"gpt-6.1-sol"}, repo.paused[1])
	require.Empty(t, repo.paused[2])
}

func TestDecideDegradationAlertOnlyNeverPauses(t *testing.T) {
	a := degradationTestAccount(1, 29, "DDDDDDDDD")
	b := degradationTestAccount(2, 29, "GGGGGGGGG")
	repo := newDegradationProbeRepoStub(map[int64][]Account{29: degradationPeers(a, b)})
	svc := &GroupHealthCheckService{accountRepo: repo}
	settings := normalizeOpenAIDegradationProbeSettings(&OpenAIDegradationProbeSettings{AlertOnlyAccountIDs: []int64{1}})
	svc.decideDegradationRound(context.Background(), []*Account{a, b}, [][]degradationProbeResult{degradationResult(true), degradationResult(false)}, settings, map[degradationTTFTKey]bool{})
	require.Empty(t, repo.paused[1])
}

func TestPauseModelForDegradationKeepsLongerRealRateLimit(t *testing.T) {
	a := degradationTestAccount(1, 29, "")
	a.Extra = map[string]any{modelRateLimitsKey: map[string]any{"gpt-6.1-sol": map[string]any{
		"reason": "upstream_429", "rate_limit_reset_at": time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)}}}
	repo := newDegradationProbeRepoStub(nil)
	svc := &GroupHealthCheckService{accountRepo: repo}
	require.NoError(t, svc.pauseModelForDegradation(context.Background(), a, "gpt-6.1-sol", time.Now().Add(25*time.Minute), degradationProbeReason))
	require.Empty(t, repo.paused[1], "已有更晚到期的真实限流时不覆盖")
}

func TestDetectSlowFirstToken(t *testing.T) {
	k := func(id int64) degradationTTFTKey {
		return degradationTTFTKey{accountID: id, upstreamModel: "gpt-6.1-sol"}
	}
	stats := []degradationTTFTStat{{k(1), 25000}, {k(2), 6000}, {k(3), 7000}, {k(4), 5000}}
	slow := detectSlowFirstToken(stats, 2, 10000)
	require.True(t, slow[k(1)])
	require.Len(t, slow, 1)

	require.Empty(t, detectSlowFirstToken([]degradationTTFTStat{{k(1), 15000}, {k(2), 6000}}, 2, 10000), "差距不到 10 秒不判")
	require.Empty(t, detectSlowFirstToken([]degradationTTFTStat{{k(1), 30000}, {k(2), 28000}, {k(3), 32000}}, 2, 10000), "所有号一起慢不判")
	require.Empty(t, detectSlowFirstToken([]degradationTTFTStat{{k(1), 60000}}, 2, 10000), "只有一个号无从比较")
}

func TestIsDegradationProbeRateLimitEntryCoversSlowTTFT(t *testing.T) {
	require.True(t, isDegradationProbeRateLimitEntry(map[string]any{"reason": degradationPauseReasonSlowTTFT}))
	require.False(t, isDegradationProbeRateLimitEntry(map[string]any{"reason": "degradation_probe_other"}))
}

func TestDecideDegradationUsesFreshAccountForLift(t *testing.T) {
	a := degradationTestAccount(1, 29, "GGGGGGGGG")
	a.Extra[modelRateLimitsKey] = map[string]any{"gpt-6.1-sol": map[string]any{
		"reason": degradationProbeReason, "rate_limit_reset_at": time.Now().Add(20 * time.Minute).UTC().Format(time.RFC3339)}}
	b := degradationTestAccount(2, 29, "GGGGGGGGG")
	// 探测期间线上流量写入了真实限流：重读后的账号上该条目已不是探针的，不能解除。
	latest := degradationTestAccount(1, 29, "GGGGGGGGG")
	latest.Extra[modelRateLimitsKey] = map[string]any{"gpt-6.1-sol": map[string]any{
		"reason": "temp_unschedulable", "rate_limit_reset_at": time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339)}}
	repo := newDegradationProbeRepoStub(map[int64][]Account{29: degradationPeers(a, b)})
	repo.fresh = map[int64]*Account{1: latest}
	svc := &GroupHealthCheckService{accountRepo: repo}
	runDegradationRound(svc, []*Account{a, b}, []bool{false, false}, nil)
	require.Empty(t, repo.lifted[1])
}

func TestDecideDegradationMergesModelsSharingUpstream(t *testing.T) {
	a := degradationTestAccount(1, 29, "")
	a.Credentials = map[string]any{"model_mapping": map[string]any{"gpt-6.1-sol": "X", "gpt-6-sol": "X"}}
	a.Extra = map[string]any{
		degradationProbeExtraKey: map[string]any{"gpt-6.1-sol": map[string]any{"h": "DDDDDDDDD"}, "gpt-6-sol": map[string]any{"h": "GGGGGGGGG"}},
		modelRateLimitsKey: map[string]any{"X": map[string]any{
			"reason": degradationProbeReason, "rate_limit_reset_at": time.Now().Add(20 * time.Minute).UTC().Format(time.RFC3339)}},
	}
	b := degradationTestAccount(2, 29, "GGGGGGGGG")
	repo := newDegradationProbeRepoStub(map[int64][]Account{29: degradationPeers(a, b)})
	svc := &GroupHealthCheckService{accountRepo: repo}
	good := degradationProbeSample{ReturnedModel: "X", HasReasoning: true, ReasoningTokens: 300}
	bad := good
	bad.ReasoningTokens = 10
	results := [][]degradationProbeResult{
		{{model: "gpt-6.1-sol", upstreamModel: "X", sample: bad}, {model: "gpt-6-sol", upstreamModel: "X", sample: good}},
		degradationResult(false),
	}
	svc.decideDegradationRound(context.Background(), []*Account{a, b}, results, normalizeOpenAIDegradationProbeSettings(nil), map[degradationTTFTKey]bool{})
	require.Equal(t, []string{"X"}, repo.paused[1])
	require.Empty(t, repo.lifted[1], "同一上游 key 不能同轮又停又解")
}

func TestDegradationStandsOutHysteresis(t *testing.T) {
	best := map[string]float64{"m": 0.3}
	require.False(t, degradationStandsOut("DDDDDGGGGG", 10, 0.3, best, "m"), "0.5-0.3 < 0.3 新停不触发")
	require.True(t, degradationStandsOut("DDDDDGGGGG", 10, 0.2, best, "m"), "已停的门槛放宽到 0.2 继续停")
}
