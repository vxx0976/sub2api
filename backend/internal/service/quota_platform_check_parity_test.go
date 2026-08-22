package service

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

// 迁移 226 收紧后的 user_platform_quotas.platform CHECK 必须与 AllowedQuotaPlatforms 逐字一致。
//
// 为什么需要这条测试：migrations 包里的 TestMigration226FinalPlatformWhitelist 只能断言
// SQL 里出现了某几个平台名——它把同一份名单**硬编码了第二遍**，因此挡不住漂移：
// 谁往 AllowedQuotaPlatforms 加第 9 个平台，那边照样绿，直到线上「后台校验通过、
// INSERT 撞 DB CHECK」才炸。这类静默故障在 150 / 156 / 157 / 158 已经各修过一次
// （157 是上游误把白名单收回 5 个，158 才恢复）。
//
// 这里从两侧的**真源**取值比对：Go 常量 AllowedQuotaPlatforms vs 嵌入的迁移 SQL。
// 新增平台时这条会红，提示必须同时补一个收紧 CHECK 的新迁移（已发布迁移受 checksum
// 保护不能原地改）。
func TestQuotaPlatformCheckMatchesAllowedQuotaPlatforms(t *testing.T) {
	const migrationName = "226_rename_national_platforms_to_upstream_naming.sql"

	content, err := migrations.FS.ReadFile(migrationName)
	require.NoError(t, err, "读取迁移 %s 失败", migrationName)
	sql := string(content)

	// 定位 ADD CONSTRAINT ... CHECK (platform IN (...)) 里的平台列表。
	idx := strings.Index(sql, "ADD CONSTRAINT user_platform_quotas_platform_check")
	require.NotEqual(t, -1, idx, "迁移里找不到收紧后的 CHECK")

	tail := sql[idx:]
	open := strings.Index(tail, "CHECK (platform IN (")
	require.NotEqual(t, -1, open, "CHECK 子句形态与预期不符")
	rest := tail[open+len("CHECK (platform IN ("):]
	end := strings.Index(rest, ")")
	require.NotEqual(t, -1, end, "CHECK 平台列表没有闭合括号")

	inSQL := regexp.MustCompile(`'([a-z0-9_]+)'`).FindAllStringSubmatch(rest[:end], -1)
	require.NotEmpty(t, inSQL, "CHECK 里没解析出任何平台")

	sqlPlatforms := make([]string, 0, len(inSQL))
	for _, m := range inSQL {
		sqlPlatforms = append(sqlPlatforms, m[1])
	}

	goPlatforms := append([]string(nil), AllowedQuotaPlatforms...)
	sort.Strings(goPlatforms)
	sort.Strings(sqlPlatforms)

	require.Equal(t, goPlatforms, sqlPlatforms,
		"AllowedQuotaPlatforms 与迁移 %s 的 CHECK 名单不一致。\n"+
			"新增/删除平台时必须同时追加一个收紧 CHECK 的新迁移——"+
			"已发布迁移受 checksum 保护，不能原地修改。", migrationName)
}
