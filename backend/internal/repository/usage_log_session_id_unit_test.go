//go:build unit

package repository

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func newSessionIDUsageLog(sessionID *string) *service.UsageLog {
	return &service.UsageLog{
		UserID:       1,
		APIKeyID:     2,
		AccountID:    3,
		RequestID:    "req-session-id",
		Model:        "claude-3",
		InputTokens:  10,
		OutputTokens: 5,
		TotalCost:    1.0,
		ActualCost:   1.0,
		SessionID:    sessionID,
		CreatedAt:    time.Now().UTC(),
	}
}

// TestPrepareUsageLogInsert_SessionIDArgWiring pins the session_id column to the
// arg slice / arg-type table so the five INSERT column lists stay in sync. session_id
// is the penultimate arg (created_at is always last).
func TestPrepareUsageLogInsert_SessionIDArgWiring(t *testing.T) {
	// 不写死总列数：dev 比上游多若干 fork 列（cache_ttl_overridden / merchant_rate_snapshot /
	// platform_cost_snapshot / long_context_billing_applied / image_input_* / country_code /
	// pricing_time_band / priced_at 等），上游的 60 在 fork 上恒不成立。真正要钉的是
	// 「session_id 在倒数第二位且参数表与实参对齐」，这一点由下面的关系式断言覆盖。
	sessionID := "sess-persisted-123"
	prepared := prepareUsageLogInsert(newSessionIDUsageLog(&sessionID))

	require.Len(t, prepared.args, len(usageLogInsertArgTypes),
		"prepared args must match the arg-type table length")

	// created_at is last; session_id is the arg immediately before it.
	sessionArg := prepared.args[len(prepared.args)-2]
	ns, ok := sessionArg.(sql.NullString)
	require.True(t, ok, "session_id arg should be a sql.NullString, got %T", sessionArg)
	require.True(t, ns.Valid)
	require.Equal(t, sessionID, ns.String)

	require.Equal(t, "text", usageLogInsertArgTypes[len(usageLogInsertArgTypes)-2],
		"session_id arg type must be text")
}

// TestPrepareUsageLogInsert_SessionIDNullWhenAbsent proves an absent session id is
// persisted as SQL NULL rather than an empty string.
func TestPrepareUsageLogInsert_SessionIDNullWhenAbsent(t *testing.T) {
	prepared := prepareUsageLogInsert(newSessionIDUsageLog(nil))
	sessionArg := prepared.args[len(prepared.args)-2]
	ns, ok := sessionArg.(sql.NullString)
	require.True(t, ok, "session_id arg should be a sql.NullString, got %T", sessionArg)
	require.False(t, ns.Valid, "absent session id must be NULL, not empty string")

	empty := ""
	preparedEmpty := prepareUsageLogInsert(newSessionIDUsageLog(&empty))
	nsEmpty := preparedEmpty.args[len(preparedEmpty.args)-2].(sql.NullString)
	require.False(t, nsEmpty.Valid, "empty session id must also be NULL")
}

func TestPrepareUsageLogInsert_RequestedReasoningEffortArgWiring(t *testing.T) {
	requested := "max"
	forwarded := "xhigh"
	prepared := prepareUsageLogInsert(&service.UsageLog{
		UserID:                   1,
		APIKeyID:                 2,
		AccountID:                3,
		RequestID:                "req-requested-effort",
		Model:                    "gpt-5.4",
		ReasoningEffort:          &forwarded,
		RequestedReasoningEffort: &requested,
		CreatedAt:                time.Now().UTC(),
	})

	require.Len(t, prepared.args, len(usageLogInsertArgTypes))

	// 不写死下标：fork 在 reasoning_effort 之前多了 country_code / pricing_time_band /
	// priced_at 等列，上游的 47/48 在 fork 上是 50/51。改为从 usageLogSelectColumns
	// 反推（它比 INSERT 列表只多一个前导 id），这样以后两边再加列都不会假红/假绿。
	effortIdx := usageLogInsertArgIndex(t, "reasoning_effort")
	requestedIdx := usageLogInsertArgIndex(t, "requested_reasoning_effort")
	require.Equal(t, effortIdx+1, requestedIdx,
		"requested_reasoning_effort must immediately follow reasoning_effort")
	require.Equal(t, "text", usageLogInsertArgTypes[requestedIdx], "requested_reasoning_effort must follow reasoning_effort")
	require.Equal(t, "text", usageLogInsertArgTypes[effortIdx], "reasoning_effort arg type must stay text")

	forwardedArg, ok := prepared.args[effortIdx].(sql.NullString)
	require.True(t, ok)
	require.True(t, forwardedArg.Valid)
	require.Equal(t, forwarded, forwardedArg.String)

	requestedArg, ok := prepared.args[requestedIdx].(sql.NullString)
	require.True(t, ok)
	require.True(t, requestedArg.Valid)
	require.Equal(t, requested, requestedArg.String)
}

// usageLogInsertArgIndex 把列名换算成 prepareUsageLogInsert 实参下标。
// usageLogSelectColumns 与 INSERT 列表同序，仅多一个前导 id。
func usageLogInsertArgIndex(t *testing.T, column string) int {
	t.Helper()
	cols := strings.Split(usageLogSelectColumns, ", ")
	require.Equal(t, "id", cols[0], "select list must still start with id")
	require.Len(t, cols, len(usageLogInsertArgTypes)+1,
		"select list must be the insert list plus the leading id column")
	for i, c := range cols[1:] {
		if c == column {
			return i
		}
	}
	t.Fatalf("column %q not found in usageLogSelectColumns", column)
	return -1
}

// TestUsageLogInsertQueries_IncludeSessionID guards that every generated INSERT path
// and the SELECT column list reference session_id.
func TestUsageLogInsertQueries_IncludeSessionID(t *testing.T) {
	require.Contains(t, usageLogSelectColumns, "requested_reasoning_effort",
		"SELECT column list must include requested_reasoning_effort")
	require.Contains(t, usageLogSelectColumns, "session_id",
		"SELECT column list must include session_id")

	sessionID := "sess-in-query"
	log := newSessionIDUsageLog(&sessionID)
	prepared := prepareUsageLogInsert(log)
	key := usageLogBatchKey(log.RequestID, log.APIKeyID)

	batchQuery, batchArgs := buildUsageLogBatchInsertQuery([]string{key},
		map[string]usageLogInsertPrepared{key: prepared})
	require.Contains(t, batchQuery, "session_id")
	require.Contains(t, batchQuery, "requested_reasoning_effort")
	// Two column references (INSERT column list + SELECT ... FROM input) plus the CTE def.
	require.GreaterOrEqual(t, strings.Count(batchQuery, "session_id"), 3)
	require.Len(t, batchArgs, len(prepared.args)+1,
		"batch args include the synthetic input_index before usage-log values")

	bestEffortQuery, bestEffortArgs := buildUsageLogBestEffortInsertQuery([]usageLogInsertPrepared{prepared})
	require.Contains(t, bestEffortQuery, "session_id")
	require.Len(t, bestEffortArgs, len(prepared.args))
}
