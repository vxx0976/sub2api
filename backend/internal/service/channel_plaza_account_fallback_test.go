//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// stubAccountRepoForPlaza 按分组返回账号，供广场的账号回落使用。
type stubAccountRepoForPlaza struct {
	AccountRepository
	byGroup map[int64][]Account
	err     error
	calls   int
}

func (s *stubAccountRepoForPlaza) ListByGroup(_ context.Context, groupID int64) ([]Account, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.byGroup[groupID], nil
}

func plazaAccount(id int64, platform, status string, models ...string) Account {
	mapping := map[string]any{}
	for _, m := range models {
		mapping[m] = m
	}
	acc := Account{ID: id, Platform: platform, Status: status}
	if len(mapping) > 0 {
		acc.Credentials = map[string]any{"model_mapping": mapping}
	}
	return acc
}

// accounts 取接口类型而非具体桩类型：传具体类型的 nil 指针会让接口非 nil
// （typed nil），测不到「未接线」这条路径。
func newPlazaServiceWithAccounts(groups []Group, accounts AccountRepository) *ChannelService {
	repo := &mockChannelRepository{
		listAllFn: func(ctx context.Context) ([]Channel, error) { return nil, nil },
	}
	return NewChannelService(repo, &stubGroupRepoForAvailable{activeGroups: groups}, nil, nil, accounts)
}

// 本 fork 的模型目录挂在账号上，不在渠道上（channel_groups / channel_model_pricing
// 生产库里是空表）。上游 ListPlazaGroups 只认渠道，会把每个分组算成 0 模型并整组丢弃，
// 广场页全空。这批用例钉住账号回落，避免以后有人把它当成"多余的分支"删掉。
func TestListPlazaGroups_AccountFallbackWhenNoChannelModels(t *testing.T) {
	groups := []Group{
		{ID: 1, Name: "GPT", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 1},
	}
	accounts := &stubAccountRepoForPlaza{byGroup: map[int64][]Account{
		1: {
			plazaAccount(11, PlatformOpenAI, StatusActive, "gpt-5", "gpt-4o"),
			plazaAccount(12, PlatformOpenAI, StatusActive, "gpt-4o", "o3"),
		},
	}}

	out, err := newPlazaServiceWithAccounts(groups, accounts).ListPlazaGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1, "渠道侧无模型时不能把分组整个丢掉")

	names := make([]string, 0, len(out[0].Models))
	for _, m := range out[0].Models {
		names = append(names, m.Name)
		require.Equal(t, PlatformOpenAI, m.Platform)
	}
	// 并集而非交集：两个账号映射不相交时取交集会得出空集
	//（历史上 GPT 分组显示 0 模型即此因）。
	require.Equal(t, []string{"gpt-4o", "gpt-5", "o3"}, names)
}

func TestListPlazaGroups_AccountFallbackSkipsIrrelevantAccounts(t *testing.T) {
	groups := []Group{
		{ID: 1, Name: "GPT", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 1},
	}
	accounts := &stubAccountRepoForPlaza{byGroup: map[int64][]Account{
		1: {
			plazaAccount(11, PlatformOpenAI, StatusActive, "gpt-5"),
			plazaAccount(12, PlatformOpenAI, "disabled", "should-not-appear"),
			plazaAccount(13, PlatformAnthropic, StatusActive, "claude-should-not-appear"),
		},
	}}

	out, err := newPlazaServiceWithAccounts(groups, accounts).ListPlazaGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Len(t, out[0].Models, 1)
	require.Equal(t, "gpt-5", out[0].Models[0].Name)
}

func TestListPlazaGroups_AccountFallbackIgnoresWildcardMappings(t *testing.T) {
	groups := []Group{
		{ID: 1, Name: "GPT", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 1},
	}
	accounts := &stubAccountRepoForPlaza{byGroup: map[int64][]Account{
		1: {plazaAccount(11, PlatformOpenAI, StatusActive, "gpt-5", "gpt-*")},
	}}

	out, err := newPlazaServiceWithAccounts(groups, accounts).ListPlazaGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, []string{"gpt-5"}, []string{out[0].Models[0].Name})
	require.Len(t, out[0].Models, 1, "通配符 from 不是具体模型，不能进目录")
}

// 全是透传账号（无 model_mapping）时回落到该平台 LiteLLM 全表；
// 没有 pricingService 就拿不到全表，此时分组照旧被丢弃（不硬造空分组）。
func TestListPlazaGroups_AccountFallbackPassthroughWithoutPricingDropsGroup(t *testing.T) {
	groups := []Group{
		{ID: 1, Name: "GPT", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 1},
	}
	accounts := &stubAccountRepoForPlaza{byGroup: map[int64][]Account{
		1: {plazaAccount(11, PlatformOpenAI, StatusActive)},
	}}

	out, err := newPlazaServiceWithAccounts(groups, accounts).ListPlazaGroups(context.Background())
	require.NoError(t, err)
	require.Empty(t, out)
}

// 账号仓储报错不能让整个广场 500：拿不到账号就当该分组没有模型。
func TestListPlazaGroups_AccountFallbackToleratesRepoError(t *testing.T) {
	groups := []Group{
		{ID: 1, Name: "GPT", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 1},
	}
	accounts := &stubAccountRepoForPlaza{err: errors.New("db down")}

	out, err := newPlazaServiceWithAccounts(groups, accounts).ListPlazaGroups(context.Background())
	require.NoError(t, err)
	require.Empty(t, out)
}

// accountRepo 未接线（测试/裁剪部署）时行为与上游完全一致，不 panic。
func TestListPlazaGroups_NilAccountRepoBehavesLikeUpstream(t *testing.T) {
	groups := []Group{
		{ID: 1, Name: "GPT", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 1},
	}
	out, err := newPlazaServiceWithAccounts(groups, nil).ListPlazaGroups(context.Background())
	require.NoError(t, err)
	require.Empty(t, out)
}

// 渠道侧已有模型时不得触发账号回落——配了渠道的部署行为必须与上游一致。
func TestListPlazaGroups_ChannelModelsSuppressAccountFallback(t *testing.T) {
	groups := []Group{
		{ID: 1, Name: "GPT", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 1},
	}
	channels := []Channel{plazaPricedChannel(9, "ch", []int64{1}, PlatformOpenAI, "from-channel")}
	accounts := &stubAccountRepoForPlaza{byGroup: map[int64][]Account{
		1: {plazaAccount(11, PlatformOpenAI, StatusActive, "from-account")},
	}}
	repo := &mockChannelRepository{
		listAllFn: func(ctx context.Context) ([]Channel, error) { return channels, nil },
	}
	svc := NewChannelService(repo, &stubGroupRepoForAvailable{activeGroups: groups}, nil, nil, accounts)

	out, err := svc.ListPlazaGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "from-channel", out[0].Models[0].Name)
	require.Len(t, out[0].Models, 1)
	require.Zero(t, accounts.calls, "渠道侧有模型就不该查账号")
}
