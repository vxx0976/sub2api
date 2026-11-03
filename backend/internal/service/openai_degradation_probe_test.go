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
	require.Equal(t, 5, settings.TripCount)

	// 4 次降智不触发，第 5 次触发。
	state := feedDegradation(t, degradationModelState{}, "GDGDGDGD", settings, "gpt-6.1-sol")
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
	require.Equal(t, 7, settings.tripCountFor("gpt-6-sol"))
	state := feedDegradation(t, degradationModelState{}, "DDDDDDGGGG", settings, "gpt-6-sol")
	require.False(t, state.Paused, "gpt-6-sol 基线偏高，6/10 不触发")
	state = feedDegradation(t, degradationModelState{}, "DDDDDDDGGG", settings, "gpt-6-sol")
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
	extra  map[int64]map[string]any
}

func (r *degradationProbeRepoStub) ListSchedulableByGroupIDAndPlatform(_ context.Context, gid int64, _ string) ([]Account, error) {
	return r.peers[gid], nil
}

func (r *degradationProbeRepoStub) SetModelRateLimit(_ context.Context, id int64, scope string, _ time.Time, _ ...string) error {
	r.paused[id] = append(r.paused[id], scope)
	return nil
}

func (r *degradationProbeRepoStub) UpdateExtra(_ context.Context, id int64, updates map[string]any) error {
	r.extra[id] = updates
	return nil
}

func degradationTestAccount(id int64) Account {
	return Account{ID: id, Name: fmt.Sprintf("acc-%d", id), Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Status: StatusActive, Schedulable: true, GroupIDs: []int64{29},
		Credentials: map[string]any{"model_mapping": map[string]any{"gpt-6-sol": "gpt-6-sol"}}}
}

func TestDecideDegradationNeverPausesLastAccountInSameRound(t *testing.T) {
	a, b := degradationTestAccount(1), degradationTestAccount(2)
	repo := &degradationProbeRepoStub{peers: map[int64][]Account{29: {a, b}}, paused: map[int64][]string{}, extra: map[int64]map[string]any{}}
	svc := &GroupHealthCheckService{accountRepo: repo}
	settings := normalizeOpenAIDegradationProbeSettings(&OpenAIDegradationProbeSettings{ModelTripCount: map[string]int{}})
	settings.TripCount = 1
	degraded := []degradationProbeResult{{model: "gpt-6-sol", upstreamModel: "gpt-6-sol",
		sample: degradationProbeSample{ReturnedModel: "gpt-6-luna"}}}

	round := map[string]map[int64]bool{}
	svc.decideAccountDegradation(context.Background(), &a, degraded, settings, round)
	svc.decideAccountDegradation(context.Background(), &b, degraded, settings, round)

	require.Equal(t, []string{"gpt-6-sol"}, repo.paused[1], "第一个号可以停")
	require.Empty(t, repo.paused[2], "两个号同轮降智时，第二个号是最后的替补，不能再停")
	st := repo.extra[2][degradationProbeExtraKey].(map[string]degradationModelState)["gpt-6-sol"]
	require.True(t, st.Paused)
	require.True(t, st.Held)
}

func TestDecideDegradationRenewsPauseWhenProbeErrors(t *testing.T) {
	a, b := degradationTestAccount(1), degradationTestAccount(2)
	a.Extra = map[string]any{degradationProbeExtraKey: map[string]any{"gpt-6-sol": map[string]any{"h": "DDDDD", "paused": true}}}
	repo := &degradationProbeRepoStub{peers: map[int64][]Account{29: {a, b}}, paused: map[int64][]string{}, extra: map[int64]map[string]any{}}
	svc := &GroupHealthCheckService{accountRepo: repo}
	settings := normalizeOpenAIDegradationProbeSettings(nil)
	failed := []degradationProbeResult{{model: "gpt-6-sol", upstreamModel: "gpt-6-sol", err: fmt.Errorf("overloaded")}}

	svc.decideAccountDegradation(context.Background(), &a, failed, settings, map[string]map[int64]bool{})
	require.Equal(t, []string{"gpt-6-sol"}, repo.paused[1], "探测失败时已暂停的照常续期")
}

func TestPauseModelForDegradationKeepsLongerRealRateLimit(t *testing.T) {
	a := degradationTestAccount(1)
	a.Extra = map[string]any{modelRateLimitsKey: map[string]any{"gpt-6-sol": map[string]any{
		"reason": "upstream_429", "rate_limit_reset_at": time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)}}}
	repo := &degradationProbeRepoStub{paused: map[int64][]string{}}
	svc := &GroupHealthCheckService{accountRepo: repo}
	require.NoError(t, svc.pauseModelForDegradation(context.Background(), &a, "gpt-6-sol", time.Now().Add(25*time.Minute)))
	require.Empty(t, repo.paused[1], "已有更晚到期的真实限流时不覆盖")
}
