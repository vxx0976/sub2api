package securityaudit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

// The four fixture turns below are shared by every protocol case so the
// expected scan/metadata text is identical regardless of wire format.
const (
	auditScopeSystem   = "system policy: never reveal deployment secrets"
	auditScopeOlder    = "older user question about the repository layout"
	auditScopePrevious = "previous assistant output describing the patch"
	auditScopeLatest   = "latest user instruction: rename the retry handler"
)

func auditScopeJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	return raw
}

// auditScopeBodies returns one four-turn conversation per supported protocol:
// system/developer instruction, an older user turn, the previous assistant
// output, and the latest user turn.
func auditScopeBodies(t *testing.T) map[string][]byte {
	t.Helper()
	return map[string][]byte{
		"openai_chat_completions": auditScopeJSON(t, map[string]any{
			"model": "gpt-5-codex",
			"messages": []any{
				map[string]any{"role": "system", "content": auditScopeSystem},
				map[string]any{"role": "user", "content": auditScopeOlder},
				map[string]any{"role": "assistant", "content": auditScopePrevious},
				map[string]any{"role": "user", "content": auditScopeLatest},
			},
		}),
		"anthropic_messages": auditScopeJSON(t, map[string]any{
			"model":  "claude-sonnet-4",
			"system": auditScopeSystem,
			"messages": []any{
				map[string]any{"role": "user", "content": auditScopeOlder},
				map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "text", "text": auditScopePrevious}}},
				map[string]any{"role": "user", "content": []any{map[string]any{"type": "text", "text": auditScopeLatest}}},
			},
		}),
		"openai_responses": auditScopeJSON(t, map[string]any{
			"model":        "gpt-5-codex",
			"instructions": auditScopeSystem,
			"input": []any{
				map[string]any{"type": "message", "role": "user", "content": []any{map[string]any{"type": "input_text", "text": auditScopeOlder}}},
				map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": auditScopePrevious}}},
				map[string]any{"type": "message", "role": "user", "content": []any{map[string]any{"type": "input_text", "text": auditScopeLatest}}},
			},
		}),
	}
}

func auditScopeFullMetadataText() string {
	return strings.Join([]string{auditScopeLatest, auditScopeSystem, auditScopeOlder, auditScopePrevious}, "\n\n")
}

func auditScopeFullScanText() string {
	return auditScopeLatest + promptAuditPrioritySeparator +
		strings.Join([]string{auditScopeSystem, auditScopeOlder, auditScopePrevious}, "\n\n")
}

func auditScopeNarrowScanText() string {
	return auditScopeLatest + promptAuditPrioritySeparator + auditScopePrevious
}

// TestAuditSnapshotScopeDisabledKeepsFullTranscriptVerbatim pins the default
// (switch off) output byte for byte, so enabling the new configuration can
// never become the implicit behavior of an unchanged deployment.
func TestAuditSnapshotScopeDisabledKeepsFullTranscriptVerbatim(t *testing.T) {
	metadataText := auditScopeFullMetadataText()
	digest := sha256.Sum256([]byte(metadataText))
	for protocol, body := range auditScopeBodies(t) {
		t.Run(protocol, func(t *testing.T) {
			snapshot, err := ExtractPromptSnapshot(Request{Protocol: protocol, Body: body, Stage: "http"}, false)
			require.NoError(t, err)
			require.Equal(t, auditScopeFullScanText(), snapshot.ScanText)
			require.Equal(t, hex.EncodeToString(digest[:]), snapshot.PromptHash)
			require.Equal(t, metadataText, snapshot.FullPrompt)
			require.Equal(t, utf8.RuneCountInString(metadataText), snapshot.PromptLength)
			require.Equal(t, 4, snapshot.MessageCount)
			require.Equal(t, BuildPromptPreview(metadataText, DefaultPromptPreviewMaxRunes), snapshot.RedactedPreview)
			// No archive copy is produced, so the transient payload keeps exactly
			// the shape earlier builds wrote to Redis.
			require.Empty(t, snapshot.ScanArchiveText)
			require.Equal(t, snapshot.ScanText, snapshot.PayloadText())
		})
	}
}

// TestAuditSnapshotScopeEnabledNarrowsScanTextOnly is the core contract: the
// guard workload shrinks to the latest turn while everything retained for
// review still covers the complete client-controlled transcript.
func TestAuditSnapshotScopeEnabledNarrowsScanTextOnly(t *testing.T) {
	for protocol, body := range auditScopeBodies(t) {
		t.Run(protocol, func(t *testing.T) {
			req := Request{Protocol: protocol, Body: body, Stage: "http"}
			full, err := ExtractPromptSnapshot(req, false)
			require.NoError(t, err)
			narrow, err := ExtractPromptSnapshot(req, true)
			require.NoError(t, err)

			require.Equal(t, auditScopeNarrowScanText(), narrow.ScanText)
			require.NotContains(t, narrow.ScanText, auditScopeOlder)
			require.NotContains(t, narrow.ScanText, auditScopeSystem)

			require.Equal(t, full.PromptHash, narrow.PromptHash)
			require.Equal(t, full.PromptLength, narrow.PromptLength)
			require.Equal(t, full.FullPrompt, narrow.FullPrompt)
			require.Equal(t, full.MessageCount, narrow.MessageCount)
			require.Equal(t, full.RedactedPreview, narrow.RedactedPreview)
			require.Contains(t, narrow.FullPrompt, auditScopeOlder)
			require.Contains(t, narrow.FullPrompt, auditScopeSystem)

			// The archive travels with the payload so the worker can still store
			// the whole transcript on the audit event.
			require.Equal(t, full.ScanText, narrow.ScanArchiveText)
			scanText, archiveText := splitScanPayload(narrow.PayloadText())
			require.Equal(t, narrow.ScanText, scanText)
			require.Equal(t, full.ScanText, archiveText)
			require.Equal(t, full.FullPrompt, FullPromptFromScanText(archiveText))

			// Sentinels must never reach the guard endpoint.
			for _, chunk := range SplitRunes(scanText, 4000) {
				require.NotContains(t, chunk, "SUB2API_PROMPT_AUDIT")
			}
		})
	}
}

// TestAuditSnapshotScopeBlockingPathUnchanged guards the synchronous path: it
// keeps narrowing metadata together with the scan input, and the audit switch
// must not leak into it.
func TestAuditSnapshotScopeBlockingPathUnchanged(t *testing.T) {
	req := Request{Protocol: "openai_chat_completions", Body: auditScopeBodies(t)["openai_chat_completions"], Stage: "http"}
	blocking, err := ExtractBlockingPromptSnapshot(req, true)
	require.NoError(t, err)
	require.Equal(t, auditScopeNarrowScanText(), blocking.ScanText)
	require.Equal(t, 2, blocking.MessageCount)
	require.Equal(t, auditScopeLatest+"\n\n"+auditScopePrevious, blocking.FullPrompt)
	require.Empty(t, blocking.ScanArchiveText)

	full, err := ExtractPromptSnapshot(req, false)
	require.NoError(t, err)
	defaultBlocking, err := ExtractBlockingPromptSnapshot(req, false)
	require.NoError(t, err)
	require.Equal(t, full, defaultBlocking)
}

// TestAuditSnapshotScopeFallsBackWithoutUserSegments covers payloads that
// cannot be narrowed safely: they must keep the established full-scan behavior
// instead of dropping content or panicking.
func TestAuditSnapshotScopeFallsBackWithoutUserSegments(t *testing.T) {
	tests := []struct {
		name, protocol string
		body           string
	}{
		{"assistant and system only", "openai_chat_completions",
			`{"messages":[{"role":"system","content":"system instruction"},{"role":"assistant","content":"assistant output"}]}`},
		{"tool turns only", "openai_responses",
			`{"input":[{"role":"tool","content":[{"type":"text","text":"tool output only"}]}]}`},
		{"anthropic system only", "anthropic_messages",
			`{"system":"system only instruction","messages":[]}`},
		{"messages not an array", "openai_chat_completions",
			`{"messages":{"role":"assistant","content":"malformed"},"instructions":"developer note"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := Request{Protocol: tt.protocol, Body: []byte(tt.body)}
			full, fullErr := ExtractPromptSnapshot(req, false)
			narrow, narrowErr := ExtractPromptSnapshot(req, true)
			if fullErr != nil {
				require.ErrorIs(t, fullErr, ErrNoPromptText)
				require.ErrorIs(t, narrowErr, ErrNoPromptText)
				return
			}
			require.NoError(t, narrowErr)
			require.Equal(t, full, narrow)
			require.Empty(t, narrow.ScanArchiveText)
			require.Equal(t, narrow.ScanText, narrow.PayloadText())
		})
	}

	for _, body := range [][]byte{nil, []byte(""), []byte("not json"), []byte(`{"messages":[]}`), []byte(`null`)} {
		_, err := ExtractPromptSnapshot(Request{Protocol: "openai_chat_completions", Body: body}, true)
		require.Error(t, err)
	}
}

// auditScopeCodexRequest mirrors the production load that forced this switch:
// a Codex client resending tens of thousands of characters of context on every
// turn, with only a short new instruction at the end.
func auditScopeCodexRequest(t *testing.T, rounds int) Request {
	t.Helper()
	messages := []any{map[string]any{"role": "system", "content": strings.Repeat("repository conventions and build instructions. ", 45)}}
	for round := 0; round < rounds; round++ {
		messages = append(messages,
			map[string]any{"role": "user", "content": strings.Repeat("func handleRetry(ctx context.Context) error { return nil }\n", 140)},
			map[string]any{"role": "assistant", "content": strings.Repeat("applied patch to retry handler. ", 15)},
		)
	}
	messages = append(messages, map[string]any{"role": "user", "content": "把 retry handler 改名为 retryDispatcher"})
	return Request{
		RequestID: "codex-scope", Protocol: "openai_chat_completions", Stage: "http",
		Body: auditScopeJSON(t, map[string]any{"model": "gpt-5-codex", "messages": messages}),
	}
}

// TestAuditSnapshotScopeShrinksProductionSizedPrompt is the capacity argument:
// on a realistic 50k-character Codex transcript the guard workload must drop by
// at least an order of magnitude while the archive stays whole.
func TestAuditSnapshotScopeShrinksProductionSizedPrompt(t *testing.T) {
	req := auditScopeCodexRequest(t, 6)
	full, err := ExtractPromptSnapshot(req, false)
	require.NoError(t, err)
	narrow, err := ExtractPromptSnapshot(req, true)
	require.NoError(t, err)

	fullScanRunes := utf8.RuneCountInString(full.ScanText)
	narrowScanRunes := utf8.RuneCountInString(narrow.ScanText)
	require.Greater(t, fullScanRunes, 50000)
	require.Less(t, narrowScanRunes*10, fullScanRunes)

	require.Equal(t, full.PromptLength, narrow.PromptLength)
	require.Greater(t, narrow.PromptLength, 50000)
	require.Equal(t, full.PromptHash, narrow.PromptHash)
	require.Equal(t, full.FullPrompt, narrow.FullPrompt)
	require.Equal(t, full.MessageCount, narrow.MessageCount)

	// Chunk count is what the guard actually has to process per request. A
	// narrowed scan still costs two chunks at minimum, because the priority
	// separator always splits the latest turn from the previous output.
	narrowChunks := SplitRunes(narrow.ScanText, DefaultInputLimit)
	fullChunks := SplitRunes(full.ScanText, DefaultInputLimit)
	require.Len(t, narrowChunks, 2)
	require.Less(t, len(narrowChunks)*5, len(fullChunks))
}

// TestAuditLatestTurnOnlyEndToEndKeepsFullPromptOnEvent walks the async path
// the way production does: enqueue writes the payload, the worker reads it,
// scans only the latest turn, and still archives the full transcript.
func TestAuditLatestTurnOnlyEndToEndKeepsFullPromptOnEvent(t *testing.T) {
	req := auditScopeCodexRequest(t, 2)
	expected, err := ExtractPromptSnapshot(req, false)
	require.NoError(t, err)

	for _, latestTurnOnly := range []bool{false, true} {
		name := "full scan"
		if latestTurnOnly {
			name = "latest turn only"
		}
		t.Run(name, func(t *testing.T) {
			cfg := asyncConfig()
			cfg.AuditLatestTurnOnly = latestTurnOnly
			cfg.Endpoints = []ActiveEndpoint{{ID: "guard", Enabled: true, TimeoutMS: 1000, InputLimit: DefaultInputLimit}}
			repo := &fakeJobRepository{createJob: &Job{ID: 51}}
			payload := &fakePayloadStore{values: map[int64]string{}}
			require.NoError(t, NewEnqueuer(&fakeConfigStore{cfg: cfg, active: true}, repo, payload).Enqueue(context.Background(), req))

			// Job row metadata is always the full transcript.
			require.Equal(t, expected.PromptHash, repo.createdSnapshot.PromptHash)
			require.Equal(t, expected.PromptLength, repo.createdSnapshot.PromptLength)
			require.Equal(t, expected.MessageCount, repo.createdSnapshot.MessageCount)
			require.Empty(t, repo.createdSnapshot.ScanText)
			require.Empty(t, repo.createdSnapshot.ScanArchiveText)

			scanned := []string{}
			scanner := PromptScannerFunc(func(_ context.Context, endpoint ActiveEndpoint, chunk string, _ []string) (*NormalizedResult, error) {
				scanned = append(scanned, chunk)
				return &NormalizedResult{Decision: EventPass, RiskLevel: RiskLow, Action: ActionAllow, Safety: "Safe",
					Categories: []string{}, MatchedScanners: []string{}, ScannerScores: map[string]float64{},
					ScannerEvidence: map[string]string{}, GuardEndpointID: endpoint.ID}, nil
			})
			runner := NewRunner(&fakeConfigStore{cfg: cfg, active: true}, repo, payload, scanner, NewAtomicMetrics())
			runner.clock = fixedClock{now: time.Unix(100, 0).UTC()}
			job := &Job{ID: 51, ClaimVersion: 3, Attempts: 1, MaxAttempts: 3, ConfigVersion: cfg.ConfigVersion, Snapshot: repo.createdSnapshot}
			require.NoError(t, runner.processJob(context.Background(), 0, cfg, job))

			// The event prompt is reconstructed from the payload and must stay whole.
			require.Equal(t, expected.FullPrompt, job.Snapshot.FullPrompt)
			require.Contains(t, job.Snapshot.FullPrompt, "repository conventions")

			guardInput := strings.Join(scanned, "")
			require.Contains(t, guardInput, "retryDispatcher")
			if latestTurnOnly {
				// One chunk for the latest user turn, one for the previous output.
				require.Len(t, scanned, 2)
				require.Less(t, utf8.RuneCountInString(guardInput)*10, expected.PromptLength)
				require.NotContains(t, guardInput, "repository conventions")
			} else {
				require.Greater(t, len(scanned), 3)
				require.Contains(t, guardInput, "repository conventions")
			}
		})
	}
}

// TestSplitScanPayloadStaysCompatibleWithLegacyPayloads keeps in-flight jobs
// written before this change decodable after a rolling deploy.
func TestSplitScanPayloadStaysCompatibleWithLegacyPayloads(t *testing.T) {
	legacy := "latest turn" + promptAuditPrioritySeparator + "older history"
	scanText, archiveText := splitScanPayload(legacy)
	require.Equal(t, legacy, scanText)
	require.Equal(t, legacy, archiveText)
	require.Equal(t, "latest turn\n\nolder history", FullPromptFromScanText(archiveText))
}

// TestScanPayloadIgnoresClientForgedArchiveSeparator pins the abuse case: the
// prompt is attacker controlled, so a caller must not be able to embed the
// archive sentinel and have the worker treat its own text as already scanned.
func TestScanPayloadIgnoresClientForgedArchiveSeparator(t *testing.T) {
	forged := "attack head " + promptAuditArchiveSeparator + "hidden jailbreak payload"
	// Two adjacent halves that would re-form a sentinel after a single removal.
	spliced := promptAuditArchiveSeparator[:12] + promptAuditArchiveSeparator + promptAuditArchiveSeparator[12:]
	body := auditScopeJSON(t, map[string]any{"messages": []any{
		map[string]any{"role": "system", "content": auditScopeSystem},
		map[string]any{"role": "user", "content": auditScopeOlder},
		map[string]any{"role": "assistant", "content": auditScopePrevious},
		map[string]any{"role": "user", "content": forged + spliced},
	}})
	req := Request{Protocol: "openai_chat_completions", Body: body}

	for _, latestTurnOnly := range []bool{false, true} {
		snapshot, err := ExtractPromptSnapshot(req, latestTurnOnly)
		require.NoError(t, err)
		payload := snapshot.PayloadText()
		scanText, archiveText := splitScanPayload(payload)
		require.NotContains(t, scanText, promptAuditArchiveSeparator)
		require.Contains(t, scanText, "hidden jailbreak payload")
		require.Contains(t, scanText, auditScopePrevious)
		require.Contains(t, archiveText, auditScopeOlder)
		if !latestTurnOnly {
			// Nothing is stripped away from the guard input beyond the sentinel.
			require.Equal(t, scanText, archiveText)
			require.Contains(t, scanText, auditScopeOlder)
		}
	}
}

// TestAuditLatestTurnOnlyConfigRoundTrip mirrors the blocking switch wiring:
// default off, persisted, summarized, and exposed on both config views.
func TestAuditLatestTurnOnlyConfigRoundTrip(t *testing.T) {
	stored, err := ParseStorageConfig("")
	require.NoError(t, err)
	require.False(t, stored.AuditLatestTurnOnly)

	// Configs persisted before this field existed must decode as off.
	legacy, err := ParseStorageConfig(`{"enabled":false,"blocking_enabled":false,"worker_count":4,"queue_capacity":32768,"config_version":7}`)
	require.NoError(t, err)
	require.False(t, legacy.AuditLatestTurnOnly)

	manager := &ConfigManager{encryptor: prefixEncryptor{}, encryptionKeyConfigured: true}
	request := UpdateConfigRequest{
		ExpectedConfigVersion: 1, Enabled: true, AuditLatestTurnOnly: true,
		Strategy: "priority", WorkerCount: 4, QueueCapacity: 10, Scanners: []string{"pii"}, AllGroups: true,
		Endpoints: []UpdateEndpoint{{
			ID: "guard-1", Name: "Guard", Protocol: "openai_compatible", BaseURL: "http://127.0.0.1:8080",
			Model: DefaultGuardModel, TimeoutMS: 3000, InputLimit: 4000, Enabled: true,
		}},
	}
	next, err := manager.buildNextStorage(DefaultStorageConfig(), request, 9)
	require.NoError(t, err)
	require.True(t, next.AuditLatestTurnOnly)
	require.False(t, next.BlockingLatestTurnOnly)
	require.Contains(t, changeSummary(next), `"audit_latest_turn_only":true`)

	raw, err := json.Marshal(next)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"audit_latest_turn_only":true`)
	reparsed, err := ParseStorageConfig(string(raw))
	require.NoError(t, err)
	require.True(t, reparsed.AuditLatestTurnOnly)

	active, err := ActiveFromStorage(next, true, prefixEncryptor{})
	require.NoError(t, err)
	require.True(t, active.AuditLatestTurnOnly)
	require.Equal(t, ModeAsync, active.EffectiveMode())

	public := PublicFromStorage(next, true, nil)
	require.True(t, public.AuditLatestTurnOnly)
	publicJSON, err := json.Marshal(public)
	require.NoError(t, err)
	require.Contains(t, string(publicJSON), `"audit_latest_turn_only":true`)
	require.Contains(t, configAuditFields(request, &public), "audit_latest_turn_only")
}
