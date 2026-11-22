package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

func TestDashboardFinanceQueriesExcludeAdminUsers(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	tests := []struct {
		name   string
		column string
		invoke func(*sql.DB) error
	}{
		{
			name:   "usage actual and account cost",
			column: "user_id",
			invoke: func(db *sql.DB) error {
				_, err := NewUsageLogRepository(nil, db).SumUsageCostsByDay(ctx, start, end, "Asia/Shanghai")
				return err
			},
		},
		{
			name:   "alipay and manual orders",
			column: "user_id",
			invoke: func(db *sql.DB) error {
				_, err := NewOrderRepo(nil, db).SumPaidCreditByDay(ctx, start, end, "Asia/Shanghai")
				return err
			},
		},
		{
			name:   "legacy recharge orders",
			column: "user_id",
			invoke: func(db *sql.DB) error {
				_, err := NewRechargeOrderRepo(nil, db).SumPaidCreditByDay(ctx, start, end, "Asia/Shanghai")
				return err
			},
		},
		{
			name:   "usdt orders",
			column: "user_id",
			invoke: func(db *sql.DB) error {
				_, err := NewUsdtOrderRepo(nil, db).SumPaidCreditByDay(ctx, start, end, "Asia/Shanghai")
				return err
			},
		},
		{
			name:   "platform balance redeem codes",
			column: "used_by",
			invoke: func(db *sql.DB) error {
				_, err := NewRedeemCodeRepository(nil, db).SumPositiveValueByDayForTypes(ctx, start, end, "Asia/Shanghai", []string{"balance", "admin_balance"})
				return err
			},
		},
		{
			name:   "manual balance adjustments",
			column: "used_by",
			invoke: func(db *sql.DB) error {
				_, err := NewRedeemCodeRepository(nil, db).SumManualAdminBalanceByDay(ctx, start, end, "Asia/Shanghai")
				return err
			},
		},
	}

	tests = append(tests, struct {
		name   string
		column string
		invoke func(*sql.DB) error
	}{
		name:   "retained business dashboard cost totals",
		column: "user_id",
		invoke: func(db *sql.DB) error {
			return NewUsageLogRepository(nil, db).(*usageLogRepository).FillDashboardBusinessCostStats(ctx, &usagestats.DashboardStats{}, start, end)
		},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := sqlmock.QueryMatcherFunc(func(_, actual string) error {
				predicate := tt.column + " NOT IN (SELECT id FROM users WHERE role = 'admin')"
				if !strings.Contains(actual, predicate) {
					return fmt.Errorf("经营统计缺少管理员排除条件 %q: %s", predicate, actual)
				}
				return nil
			})
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
			require.NoError(t, err)
			defer db.Close()

			rows := sqlmock.NewRows([]string{"day", "total"})
			if tt.name == "retained business dashboard cost totals" {
				rows = sqlmock.NewRows([]string{"earliest", "total_cost", "total_actual_cost", "total_account_cost", "today_cost", "today_actual_cost", "today_account_cost"}).
					AddRow(nil, 0, 0, 0, 0, 0, 0)
			}
			mock.ExpectQuery("finance").WillReturnRows(rows)
			require.NoError(t, tt.invoke(db))
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
