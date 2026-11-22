package repository

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

const dashboardAdminPredicate = `NOT EXISTS (SELECT 1 FROM users dashboard_user WHERE dashboard_user.id = usage_logs.user_id AND dashboard_user.role = 'admin')`

func TestBusinessUsageTrendReadsFilteredLogsNotOperationalAggregates(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}
	start := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	ctx := usagestats.WithBusinessScope(context.Background())

	mock.ExpectQuery(regexp.QuoteMeta(dashboardAdminPredicate)).
		WithArgs(start, end).
		WillReturnRows(sqlmock.NewRows([]string{"date", "requests", "input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_tokens", "cost", "actual_cost"}))

	trend, err := repo.GetUsageTrendWithFilters(ctx, start, end, "day", 0, 0, 0, 0, "", nil, nil, nil)
	require.NoError(t, err)
	require.Empty(t, trend)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBusinessScopedAggregateQueriesExcludeAdminUnlessUserSelected(t *testing.T) {
	start := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	adminRows := sqlmock.NewRows([]string{"date", "requests", "input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_tokens", "cost", "actual_cost"})
	modelRows := sqlmock.NewRows([]string{"model", "requests", "input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_tokens", "cost", "actual_cost", "account_cost"})

	tests := []struct {
		name       string
		business   bool
		userID     int64
		column     string
		rows       func() *sqlmock.Rows
		invoke     func(context.Context, *sql.DB) error
		wantFilter bool
	}{
		{
			name: "model business aggregate", business: true, column: "usage_logs.user_id", wantFilter: true,
			rows: func() *sqlmock.Rows { return modelRows },
			invoke: func(ctx context.Context, db *sql.DB) error {
				_, err := NewUsageLogRepository(nil, db).GetModelStatsWithFilters(ctx, start, end, 0, 0, 0, 0, nil, nil, nil)
				return err
			},
		},
		{
			name: "model explicit admin personal view", business: true, userID: 42, column: "usage_logs.user_id",
			rows: func() *sqlmock.Rows {
				return sqlmock.NewRows([]string{"model", "requests", "input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_tokens", "cost", "actual_cost", "account_cost"})
			},
			invoke: func(ctx context.Context, db *sql.DB) error {
				_, err := NewUsageLogRepository(nil, db).GetModelStatsWithFilters(ctx, start, end, 42, 0, 0, 0, nil, nil, nil)
				return err
			},
		},
		{
			name: "group business aggregate", business: true, column: "ul.user_id", wantFilter: true,
			rows: func() *sqlmock.Rows {
				return sqlmock.NewRows([]string{"group_id", "group_name", "requests", "total_tokens", "cost", "actual_cost", "account_cost"})
			},
			invoke: func(ctx context.Context, db *sql.DB) error {
				_, err := NewUsageLogRepository(nil, db).GetGroupStatsWithFilters(ctx, start, end, 0, 0, 0, 0, nil, nil, nil)
				return err
			},
		},
		{
			name: "user breakdown business aggregate", business: true, column: "ul.user_id", wantFilter: true,
			rows: func() *sqlmock.Rows {
				return sqlmock.NewRows([]string{"user_id", "email", "requests", "input_tokens", "output_tokens", "cache_tokens", "total_tokens", "cost", "actual_cost", "account_cost"})
			},
			invoke: func(ctx context.Context, db *sql.DB) error {
				_, err := NewUsageLogRepository(nil, db).GetUserBreakdownStats(ctx, start, end, usagestats.UserBreakdownDimension{}, 0)
				return err
			},
		},
		{
			name: "user breakdown explicit personal view", business: true, userID: 42, column: "ul.user_id",
			rows: func() *sqlmock.Rows {
				return sqlmock.NewRows([]string{"user_id", "email", "requests", "input_tokens", "output_tokens", "cache_tokens", "total_tokens", "cost", "actual_cost", "account_cost"})
			},
			invoke: func(ctx context.Context, db *sql.DB) error {
				_, err := NewUsageLogRepository(nil, db).GetUserBreakdownStats(ctx, start, end, usagestats.UserBreakdownDimension{UserID: 42}, 0)
				return err
			},
		},
		{
			name: "user spending business ranking", business: true, column: "u.user_id", wantFilter: true,
			rows: func() *sqlmock.Rows {
				return sqlmock.NewRows([]string{"user_id", "email", "username", "actual_cost", "requests", "tokens", "total_actual_cost", "total_requests", "total_tokens"})
			},
			invoke: func(ctx context.Context, db *sql.DB) error {
				_, err := NewUsageLogRepository(nil, db).GetUserSpendingRanking(ctx, start, end, 12)
				return err
			},
		},
		{
			name: "user trend business top users", business: true, column: "usage_logs.user_id", wantFilter: true,
			rows: func() *sqlmock.Rows { return adminRows },
			invoke: func(ctx context.Context, db *sql.DB) error {
				_, err := NewUsageLogRepository(nil, db).GetUserUsageTrend(ctx, start, end, "day", 12, "actual_cost")
				return err
			},
		},
		{
			name: "API key trend business top keys", business: true, column: "usage_logs.user_id", wantFilter: true,
			rows: func() *sqlmock.Rows {
				return sqlmock.NewRows([]string{"date", "api_key_id", "key_name", "requests", "tokens"})
			},
			invoke: func(ctx context.Context, db *sql.DB) error {
				_, err := NewUsageLogRepository(nil, db).GetAPIKeyUsageTrend(ctx, start, end, "day", 12)
				return err
			},
		},
		{
			name: "operational account query without scope", userID: 0, column: "usage_logs.user_id",
			rows: func() *sqlmock.Rows {
				return sqlmock.NewRows([]string{"model", "requests", "input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_tokens", "cost", "actual_cost", "account_cost"})
			},
			invoke: func(ctx context.Context, db *sql.DB) error {
				_, err := NewUsageLogRepository(nil, db).GetModelStatsWithFilters(ctx, start, end, 0, 0, 7, 0, nil, nil, nil)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.business {
				ctx = usagestats.WithBusinessScope(ctx)
			}
			var query string
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(_, actualSQL string) error {
				query = actualSQL
				return nil
			})))
			require.NoError(t, err)
			t.Cleanup(func() { _ = db.Close() })
			mock.ExpectQuery(".+").WillReturnRows(tt.rows())
			require.NoError(t, tt.invoke(ctx, db))
			require.NoError(t, mock.ExpectationsWereMet())
			predicate := "dashboard_user.id = " + tt.column + " AND dashboard_user.role = 'admin'"
			if tt.wantFilter {
				require.Contains(t, query, predicate)
			} else {
				require.NotContains(t, query, "dashboard_user.role = 'admin'")
			}
		})
	}
}

func TestDashboardBusinessUsageConditionScope(t *testing.T) {
	ctx := usagestats.WithBusinessScope(context.Background())
	filter := dashboardBusinessUsageCondition(ctx, 0, "ul.user_id")
	require.Contains(t, filter, "role = 'admin'")
	require.Contains(t, filter, "ul.user_id")
	require.Empty(t, dashboardBusinessUsageCondition(ctx, 99, "ul.user_id"))
	require.Empty(t, dashboardBusinessUsageCondition(context.Background(), 0, "ul.user_id"))
	require.True(t, usagestats.IsBusinessScope(ctx))
	require.False(t, usagestats.IsBusinessScope(context.Background()))
	require.False(t, strings.Contains(dashboardBusinessUsageCondition(ctx, 1, "u.user_id"), "admin"))
}
