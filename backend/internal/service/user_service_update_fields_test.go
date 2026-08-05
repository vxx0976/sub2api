//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// 这些用例锁死"每个入口只声明自己真正要改的列"：
// 任何退回整行回写的改动都会让并发写入被陈旧快照覆盖，并在这里变红。

func TestUpdateProfile_OnlyDeclaresRequestedColumns(t *testing.T) {
	username := "renamed"
	tests := []struct {
		name string
		req  UpdateProfileRequest
		want UserUpdateFields
	}{
		{
			name: "username only",
			req:  UpdateProfileRequest{Username: &username},
			want: UserUpdateFields{Username: true},
		},
		{
			name: "notify settings only",
			req:  UpdateProfileRequest{BalanceNotifyEnabled: boolPtr(true)},
			want: UserUpdateFields{BalanceNotifySettings: true},
		},
		{
			name: "username and notify threshold",
			req:  UpdateProfileRequest{Username: &username, BalanceNotifyThreshold: float64Ptr(1.5)},
			want: UserUpdateFields{Username: true, BalanceNotifySettings: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{getByIDUser: &User{ID: 7, Balance: 0.30, Status: StatusActive}}
			svc := NewUserService(repo, nil, nil, nil)

			_, err := svc.UpdateProfile(context.Background(), 7, tt.req)
			require.NoError(t, err)
			require.Equal(t, []UserUpdateFields{tt.want}, repo.updateFields)
		})
	}
}

// 只改头像时用户行没有任何列要写，不应产生一次整行更新。
func TestUpdateProfile_AvatarOnlySkipsUserRowWrite(t *testing.T) {
	repo := &mockUserRepo{getByIDUser: &User{ID: 7, Balance: 0.30}}
	svc := NewUserService(repo, nil, nil, nil)

	avatar := "https://cdn.example.com/a.png"
	_, err := svc.UpdateProfile(context.Background(), 7, UpdateProfileRequest{AvatarURL: &avatar})
	require.NoError(t, err)
	require.Len(t, repo.upsertAvatarArgs, 1, "avatar must still be stored")
	require.Equal(t, []UserUpdateFields{{}}, repo.updateFields, "no user column should be declared")
}

// 改密在 dev 上不走整行 Update，也不走 Update(..., UserUpdateFields{PasswordHash: true})：
// dev 的 users 表有真实 token_version 列，必须用 UpdatePasswordAndBumpTokenVersion
// 只写 password_hash 并原子自增版本号（上游那条路径版本号不落库，旧令牌不会失效）。
// 所以这里断言的「最小写入面」比上游更严：一次专用写入，且完全不碰 Update。
func TestChangePassword_OnlyWritesPasswordHashAndBumpsTokenVersion(t *testing.T) {
	user := &User{ID: 7, Balance: 0.30}
	require.NoError(t, user.SetPassword("old-password"))
	repo := &mockUserRepo{getByIDUser: user}
	svc := NewUserService(repo, nil, nil, nil)

	err := svc.ChangePassword(context.Background(), 7, ChangePasswordRequest{
		CurrentPassword: "old-password",
		NewPassword:     "new-password",
	})
	require.NoError(t, err)

	require.Empty(t, repo.updateFields, "改密不得退回整行/按列 Update，否则会用旧快照覆盖并发写入的 balance 等字段")
	require.Len(t, repo.passwordBumps, 1, "必须且只能走一次专用改密写入")
	require.Equal(t, int64(7), repo.passwordBumps[0].userID)

	// GetByID 返回的是副本，所以校验落库哈希本身，而不是比对测试持有的 user 对象。
	stored := &User{PasswordHash: repo.passwordBumps[0].passwordHash}
	require.True(t, stored.CheckPassword("new-password"), "落库的必须是新密码哈希")
	require.False(t, stored.CheckPassword("old-password"))
}

// 改状态在 dev 上走只写 status 列的专用语句，比上游的
// Update(..., UserUpdateFields{Status: true}) 还少一次 GetByID —— 少读一次就少一个
// 用旧快照覆盖 balance / token_version 的窗口。断言同样落在「一次专用写入 + 完全不碰
// Update」上。
func TestUpdateStatus_OnlyWritesStatusColumn(t *testing.T) {
	repo := &mockUserRepo{getByIDUser: &User{ID: 7, Balance: 0.30, Status: StatusActive}}
	svc := NewUserService(repo, nil, nil, nil)

	require.NoError(t, svc.UpdateStatus(context.Background(), 7, StatusDisabled))

	require.Empty(t, repo.updateFields, "改状态不得退回整行/按列 Update")
	require.Equal(t, []mockStatusWrite{{userID: 7, status: StatusDisabled}}, repo.statusWrites)
}
