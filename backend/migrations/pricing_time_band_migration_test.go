package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// 迁移文件一旦发布就被 checksum 锁死不可修改，写错只能追加新迁移来修，
// 因此这里在 CI 里先把关键子句钉住。
func TestMigration223AddsPricingTimeBandColumns(t *testing.T) {
	content, err := FS.ReadFile("223_usage_log_pricing_time_band.sql")
	require.NoError(t, err)
	sql := string(content)

	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS pricing_time_band VARCHAR(16)")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS priced_at TIMESTAMPTZ")

	// CHECK 必须允许 NULL（未走内置分档价的行）并只接受两个合法档位。
	require.Contains(t, sql, "usage_logs_pricing_time_band_check")
	require.Contains(t, sql, "pricing_time_band IS NULL")
	require.Contains(t, sql, "'peak'")
	require.Contains(t, sql, "'offpeak'")

	// NOT VALID：避免对 usage_logs 这张大表做全表校验扫描（照 204 的写法）。
	require.Contains(t, sql, "NOT VALID")

	// 重跑安全：DROP CONSTRAINT IF EXISTS 在前。
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS usage_logs_pricing_time_band_check")
	require.Less(t,
		indexOf(sql, "DROP CONSTRAINT IF EXISTS usage_logs_pricing_time_band_check"),
		indexOf(sql, "ADD CONSTRAINT usage_logs_pricing_time_band_check"),
		"DROP 必须在 ADD 之前，否则重跑会因约束已存在而失败")
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
