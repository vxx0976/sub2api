package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func newOrderEntRepo(t *testing.T) (service.OrderRepository, *dbent.Client, int64) {
	t.Helper()

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", t.Name()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(10)

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	user := client.User.Create().
		SetEmail("payer@example.com").
		SetPasswordHash("hash").
		SaveX(context.Background())

	return NewOrderRepo(client, nil), client, user.ID
}

func TestOrderRepoCreateWithUniquePaymentAmountSkipsActiveAndRecentExpired(t *testing.T) {
	repo, _, userID := newOrderEntRepo(t)
	ctx := context.Background()
	expiresAt := time.Now().Add(5 * time.Minute)

	first := &service.Order{
		OrderNo:      "A20260424000000000001",
		UserID:       userID,
		Amount:       10,
		CreditAmount: 1,
		Multiplier:   1,
		Status:       "pending",
		PayType:      "alipay",
		ExpiredAt:    &expiresAt,
	}
	require.NoError(t, repo.CreateWithUniquePaymentAmount(ctx, first, 10, 0.01, 2*time.Hour))
	require.Equal(t, 10.00, first.PaymentAmount)

	second := &service.Order{
		OrderNo:      "A20260424000000000002",
		UserID:       userID,
		Amount:       10,
		CreditAmount: 1,
		Multiplier:   1,
		Status:       "pending",
		PayType:      "alipay",
		ExpiredAt:    &expiresAt,
	}
	require.NoError(t, repo.CreateWithUniquePaymentAmount(ctx, second, 10, 0.01, 2*time.Hour))
	require.Equal(t, 10.01, second.PaymentAmount)

	recentExpiredAt := time.Now().Add(-30 * time.Minute)
	recentExpired := &service.Order{
		OrderNo:       "A20260424000000000003",
		UserID:        userID,
		Amount:        20,
		PaymentAmount: 20,
		CreditAmount:  1,
		Multiplier:    1,
		Status:        "expired",
		PayType:       "alipay",
		ExpiredAt:     &recentExpiredAt,
	}
	require.NoError(t, repo.Create(ctx, recentExpired))

	third := &service.Order{
		OrderNo:      "A20260424000000000004",
		UserID:       userID,
		Amount:       20,
		CreditAmount: 1,
		Multiplier:   1,
		Status:       "pending",
		PayType:      "alipay",
		ExpiredAt:    &expiresAt,
	}
	require.NoError(t, repo.CreateWithUniquePaymentAmount(ctx, third, 20, 0.01, 2*time.Hour))
	require.Equal(t, 20.01, third.PaymentAmount)
}

// 二次入账防护：reuseWindow 内刚入账（paid）的订单，其应付金额不可立即复用给新订单，
// 否则账单轮询的回溯窗口内同一笔账单可能被错配二次入账。
func TestOrderRepoCreateWithUniquePaymentAmountSkipsRecentPaid(t *testing.T) {
	repo, _, userID := newOrderEntRepo(t)
	ctx := context.Background()
	expiresAt := time.Now().Add(5 * time.Minute)

	paid := &service.Order{
		OrderNo:       "A20260424000000000101",
		UserID:        userID,
		Amount:        30,
		PaymentAmount: 30,
		CreditAmount:  1,
		Multiplier:    1,
		Status:        "pending",
		PayType:       "alipay",
		ExpiredAt:     &expiresAt,
	}
	require.NoError(t, repo.Create(ctx, paid))
	paidAt := time.Now().Add(-10 * time.Minute)
	tradeNo := "BILL-30-A"
	require.NoError(t, repo.UpdateStatus(ctx, paid.OrderNo, "pending", "paid", &tradeNo, &paidAt))

	next := &service.Order{
		OrderNo:      "A20260424000000000102",
		UserID:       userID,
		Amount:       30,
		CreditAmount: 1,
		Multiplier:   1,
		Status:       "pending",
		PayType:      "alipay",
		ExpiredAt:    &expiresAt,
	}
	require.NoError(t, repo.CreateWithUniquePaymentAmount(ctx, next, 30, 0.01, 2*time.Hour))
	require.Equal(t, 30.01, next.PaymentAmount)

	// TradeNoUsedByPaidOrder 兜底：已被 paid 订单占用的 trade_no 命中，自身与未占用的返回 false。
	used, err := repo.TradeNoUsedByPaidOrder(ctx, tradeNo, "A20260424000000000102")
	require.NoError(t, err)
	require.True(t, used)

	self, err := repo.TradeNoUsedByPaidOrder(ctx, tradeNo, paid.OrderNo)
	require.NoError(t, err)
	require.False(t, self)

	none, err := repo.TradeNoUsedByPaidOrder(ctx, "BILL-UNKNOWN", "A20260424000000000102")
	require.NoError(t, err)
	require.False(t, none)
}
