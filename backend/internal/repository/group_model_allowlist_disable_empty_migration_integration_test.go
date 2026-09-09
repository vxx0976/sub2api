//go:build integration

package repository

import (
	"context"
	"testing"

	dbmigrations "github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

const groupModelAllowlistDisableEmptyMigration = "238_group_model_allowlist_disable_empty.sql"

// 238 关掉「enabled=true 但 models 为空」的历史脏数据。
//
// 这种组合在旧语义（只过滤 /v1/models 展示）下是彻底的 no-op，在 235 之后的新语义
// （同时做请求准入）下变成「白名单为空集」，该分组下每个网关请求都会被中间件拦成 404。
// 管理端现在已经把这种组合判成 400，库里残留的旧行只会造成整组静默停摆。
func TestMigration238DisablesEmptyModelAllowlist(t *testing.T) {
	tx := testTx(t)
	ctx := context.Background()

	insert := func(name, allowlist string) int64 {
		t.Helper()
		var id int64
		require.NoError(t, tx.QueryRowContext(ctx, `
INSERT INTO groups (name, platform, rate_multiplier, status, model_allowlist)
VALUES ($1, 'anthropic', 1, 'active', $2::jsonb)
RETURNING id
`, name, allowlist).Scan(&id))
		return id
	}

	emptyArray := insert("migration-238-empty-array", `{"enabled":true,"models":[]}`)
	missingKey := insert("migration-238-missing-models", `{"enabled":true}`)
	configured := insert("migration-238-configured", `{"enabled":true,"models":["claude-sonnet-5"]}`)
	disabled := insert("migration-238-disabled", `{"enabled":false,"models":[]}`)
	untouched := insert("migration-238-default", `{}`)

	migrationSQL, err := dbmigrations.FS.ReadFile(groupModelAllowlistDisableEmptyMigration)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err)

	read := func(id int64) string {
		t.Helper()
		var allowlist string
		require.NoError(t, tx.QueryRowContext(ctx,
			"SELECT model_allowlist::text FROM groups WHERE id = $1", id).Scan(&allowlist))
		return allowlist
	}

	// 空列表的两种写法都被关掉，models 字段原样保留（管理员重新勾选即可）。
	require.JSONEq(t, `{"enabled":false,"models":[]}`, read(emptyArray))
	require.JSONEq(t, `{"enabled":false}`, read(missingKey))

	// 已配置模型的白名单必须保持开启——那是管理员真实意图，本迁移不得误伤。
	require.JSONEq(t, `{"enabled":true,"models":["claude-sonnet-5"]}`, read(configured))
	require.JSONEq(t, `{"enabled":false,"models":[]}`, read(disabled))
	require.JSONEq(t, `{}`, read(untouched))

	// 重复执行安全（幂等）。
	_, err = tx.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err)
	require.JSONEq(t, `{"enabled":true,"models":["claude-sonnet-5"]}`, read(configured))
	require.JSONEq(t, `{"enabled":false,"models":[]}`, read(emptyArray))
}
