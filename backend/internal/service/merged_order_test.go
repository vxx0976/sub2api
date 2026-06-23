package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMapManualAdjustment_PositiveAdd(t *testing.T) {
	uid := int64(42)
	used := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	created := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)

	m := mapManualAdjustment(&RedeemCode{
		ID:        7,
		Code:      "abc123",
		Type:      AdjustmentTypeAdminBalance,
		Value:     100,
		Status:    StatusUsed,
		UsedBy:    &uid,
		UsedAt:    &used,
		CreatedAt: created,
		Notes:     "VIP gift",
	})

	require.Equal(t, MergedChannelManual, m.Channel)
	require.Equal(t, MergedChannelManual, m.PayType)
	require.Equal(t, mergedStatusCompleted, m.Status)
	require.Equal(t, int64(7), m.ID)
	require.Equal(t, "ADJ-7", m.OrderNo) // 合成订单号，不暴露随机 code
	require.Equal(t, uid, m.UserID)
	require.Equal(t, "100", m.CreditAmount)
	require.Equal(t, "", m.Amount) // 手工调整无"支付金额"
	require.NotNil(t, m.Note)
	require.Equal(t, "VIP gift", *m.Note)
	require.True(t, m.createdAt.Equal(used)) // 排序键以 used_at 为准
	require.NotNil(t, m.PaidAt)
}

func TestMapManualAdjustment_NegativeValueKeepsSign(t *testing.T) {
	uid := int64(1)
	used := time.Now()
	m := mapManualAdjustment(&RedeemCode{
		ID:     9,
		Value:  -50,
		Status: StatusUsed,
		UsedBy: &uid,
		UsedAt: &used,
	})
	require.Equal(t, "-50", m.CreditAmount) // 扣减/退款保留负号
	require.Nil(t, m.Note)                  // 空 notes → nil
}

func TestMapManualAdjustment_NilUsedByAndFallbackTime(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	m := mapManualAdjustment(&RedeemCode{
		ID:        3,
		Value:     10,
		Status:    StatusUsed,
		UsedBy:    nil, // 防御：缺 used_by 时 UserID=0，不 panic
		UsedAt:    nil, // used_at 缺失时回退 created_at
		CreatedAt: created,
	})
	require.Equal(t, int64(0), m.UserID)
	require.True(t, m.createdAt.Equal(created))
}
