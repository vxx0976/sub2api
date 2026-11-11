//go:build unit

package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const relayToolStructureRejectBody = `{"error":{"message":"request contains an incomplete or unsupported Claude Code tool structure","type":"invalid_request_error"},"type":"error"}`

func TestIsRelayToolStructureRejectedError(t *testing.T) {
	require.True(t, isRelayToolStructureRejectedError(http.StatusUnprocessableEntity, []byte(relayToolStructureRejectBody)))
	// 同一文案也可能以 400 返回。
	require.True(t, isRelayToolStructureRejectedError(http.StatusBadRequest, []byte(relayToolStructureRejectBody)))
	// 其他 422 语义（普通入参校验、风控拦截）不能误判成工具结构拒绝。
	require.False(t, isRelayToolStructureRejectedError(http.StatusUnprocessableEntity, []byte(`{"error":{"message":"max_tokens: must be positive","type":"invalid_request_error"}}`)))
	require.False(t, isRelayToolStructureRejectedError(http.StatusUnprocessableEntity, []byte(`{"error":{"message":"request blocked: client identity could not be fully anonymized","type":"invalid_request_error"}}`)))
	// 只认 400/422：500 里的同文案不是入参校验。
	require.False(t, isRelayToolStructureRejectedError(http.StatusInternalServerError, []byte(relayToolStructureRejectBody)))
	require.False(t, isRelayToolStructureRejectedError(http.StatusUnprocessableEntity, nil))
	// 只有 "tool structure" 但缺修饰词，不足以判定。
	require.False(t, isRelayToolStructureRejectedError(http.StatusUnprocessableEntity, []byte(`{"error":{"message":"tool structure"}}`)))
}

func relayToolStructureTestContext(groupID int64) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	c.Set("api_key", &APIKey{ID: 1, GroupID: &groupID})
	return c
}

// /v1/messages 上遇到中转的工具结构拒绝：换号，并把这家号冷却 30 分钟，
// 否则粘性会话会一直把同一会话钉在它上面反复失败。
func TestRelayMessagesRejectedFailover_ToolStructureRejectCoolsAccount(t *testing.T) {
	relay := &Account{ID: 172, Name: "Max-开心", Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer1 := Account{ID: 232, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer2 := Account{ID: 281, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	repo := &relayQueueRejectRepoStub{peers: []Account{*relay, peer1, peer2}}
	s := &GatewayService{accountRepo: repo}
	c := relayToolStructureTestContext(86)

	resp := newRelayBlockedTestResp(http.StatusUnprocessableEntity, relayToolStructureRejectBody)
	failoverErr := s.relayMessagesRejectedFailover(context.Background(), c, resp, relay, false, "claude-opus-5-5")
	require.NotNil(t, failoverErr, "工具结构拒绝应换号")
	require.True(t, failoverErr.ShouldRetryNextAccount())
	require.False(t, failoverErr.RetryableOnSameAccount, "同一账号重试必然再被拒")
	require.Equal(t, relayToolStructureRejectBody, string(failoverErr.ResponseBody))

	require.Len(t, repo.tempUnschedCalls, 1)
	require.Equal(t, int64(172), repo.tempUnschedCalls[0].accountID)
	require.WithinDuration(t, time.Now().Add(relayToolStructureRejectCooldown), repo.tempUnschedCalls[0].until, 5*time.Second)
	require.Contains(t, repo.tempUnschedCalls[0].reason, "tool structure")

	v, _ := c.Get(OpsUpstreamErrorsKey)
	events, _ := v.([]*OpsUpstreamErrorEvent)
	require.Len(t, events, 1)
	require.Equal(t, "failover", events[0].Kind)
}

// 冷却要求分组内至少剩两个可调度替补：请求本身真的残缺时每个号都会拒，
// 不能顺着换号把整组逐个冷却到无号可用。
func TestTempUnscheduleRelayToolStructureReject_RequiresAlternatives(t *testing.T) {
	relay := &Account{ID: 172, Name: "Max-开心", Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer1 := Account{ID: 232, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer2 := Account{ID: 281, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}

	// 三号分组：剩两个替补，冷却。
	okRepo := &relayQueueRejectRepoStub{peers: []Account{*relay, peer1, peer2}}
	(&GatewayService{accountRepo: okRepo}).tempUnscheduleRelayToolStructureReject(context.Background(),
		relayToolStructureTestContext(86), relay, []byte(relayToolStructureRejectBody), "claude-opus-5-5")
	require.Len(t, okRepo.tempUnschedCalls, 1)

	// 两号分组：冷却后就只剩一个号，不冷却。
	twoRepo := &relayQueueRejectRepoStub{peers: []Account{*relay, peer1}}
	(&GatewayService{accountRepo: twoRepo}).tempUnscheduleRelayToolStructureReject(context.Background(),
		relayToolStructureTestContext(86), relay, []byte(relayToolStructureRejectBody), "claude-opus-5-5")
	require.Empty(t, twoRepo.tempUnschedCalls)

	// 替补不支持该模型：不算替补（若算，就会凑够两个而误冷却）。
	narrow := peer1
	narrow.Credentials = map[string]any{"model_mapping": map[string]any{"claude-sonnet-5": "claude-sonnet-5"}}
	narrowRepo := &relayQueueRejectRepoStub{peers: []Account{*relay, narrow, peer2}}
	(&GatewayService{accountRepo: narrowRepo}).tempUnscheduleRelayToolStructureReject(context.Background(),
		relayToolStructureTestContext(86), relay, []byte(relayToolStructureRejectBody), "claude-opus-5-5")
	require.Empty(t, narrowRepo.tempUnschedCalls)

	// 两个真正支持的替补：照常冷却。
	wideRepo := &relayQueueRejectRepoStub{peers: []Account{*relay, narrow, peer1, peer2}}
	(&GatewayService{accountRepo: wideRepo}).tempUnscheduleRelayToolStructureReject(context.Background(),
		relayToolStructureTestContext(86), relay, []byte(relayToolStructureRejectBody), "claude-opus-5-5")
	require.Len(t, wideRepo.tempUnschedCalls, 1)
}

func TestTempUnscheduleRelayToolStructureReject_Skips(t *testing.T) {
	relay := &Account{ID: 172, Name: "Max-开心", Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer1 := Account{ID: 232, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer2 := Account{ID: 281, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}

	// 同一请求最多冷却一个号。
	repo := &relayQueueRejectRepoStub{peers: []Account{*relay, peer1, peer2}}
	svc := &GatewayService{accountRepo: repo}
	c := relayToolStructureTestContext(86)
	svc.tempUnscheduleRelayToolStructureReject(context.Background(), c, relay, []byte(relayToolStructureRejectBody), "claude-opus-5-5")
	svc.tempUnscheduleRelayToolStructureReject(context.Background(), c, &peer1, []byte(relayToolStructureRejectBody), "claude-opus-5-5")
	require.Len(t, repo.tempUnschedCalls, 1)
	require.Equal(t, int64(172), repo.tempUnschedCalls[0].accountID)

	// 已在冷却中：不重复写。
	future := time.Now().Add(10 * time.Minute)
	cooling := *relay
	cooling.TempUnschedulableUntil = &future
	coolingRepo := &relayQueueRejectRepoStub{peers: []Account{*relay, peer1, peer2}}
	(&GatewayService{accountRepo: coolingRepo}).tempUnscheduleRelayToolStructureReject(context.Background(),
		relayToolStructureTestContext(86), &cooling, []byte(relayToolStructureRejectBody), "claude-opus-5-5")
	require.Empty(t, coolingRepo.tempUnschedCalls)

	// 官方 OAuth 号与非 anthropic 平台：冷却账号解决不了入参校验问题。
	oAuthRepo := &relayQueueRejectRepoStub{peers: []Account{*relay, peer1, peer2}}
	svcOAuth := &GatewayService{accountRepo: oAuthRepo}
	svcOAuth.tempUnscheduleRelayToolStructureReject(context.Background(), relayToolStructureTestContext(86),
		&Account{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true},
		[]byte(relayToolStructureRejectBody), "claude-opus-5-5")
	svcOAuth.tempUnscheduleRelayToolStructureReject(context.Background(), relayToolStructureTestContext(86),
		&Account{ID: 2, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true},
		[]byte(relayToolStructureRejectBody), "gpt-6-sol")
	require.Empty(t, oAuthRepo.tempUnschedCalls)

	// 无分组上下文：无法判断替补，不冷却。
	noGroupRepo := &relayQueueRejectRepoStub{peers: []Account{*relay, peer1, peer2}}
	gin.SetMode(gin.TestMode)
	bare, _ := gin.CreateTestContext(httptest.NewRecorder())
	bare.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	(&GatewayService{accountRepo: noGroupRepo}).tempUnscheduleRelayToolStructureReject(context.Background(),
		bare, relay, []byte(relayToolStructureRejectBody), "claude-opus-5-5")
	require.Empty(t, noGroupRepo.tempUnschedCalls)
}

// 中转号同时在另一个只有它自己的分组：冷却作用于所有分组，任一分组没替补就不能冷却。
func TestTempUnscheduleRelayToolStructureReject_AllGroupsNeedAlternatives(t *testing.T) {
	relay := &Account{ID: 172, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer1 := Account{ID: 232, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer2 := Account{ID: 281, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	multi := *relay
	multi.GroupIDs = []int64{86, 168}
	repo := &relayQueueRejectRepoStub{
		peers:      []Account{multi, peer1, peer2},
		groupPeers: map[int64][]Account{168: {multi}},
	}
	(&GatewayService{accountRepo: repo}).tempUnscheduleRelayToolStructureReject(context.Background(),
		relayToolStructureTestContext(86), &multi, []byte(relayToolStructureRejectBody), "claude-opus-5-5")
	require.Empty(t, repo.tempUnschedCalls)
}

// 未命中工具结构文案的 422 保持原处理：不换号、不冷却。
func TestRelayMessagesRejectedFailover_ToolStructureUnmatched(t *testing.T) {
	relay := &Account{ID: 172, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer1 := Account{ID: 232, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer2 := Account{ID: 281, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	repo := &relayQueueRejectRepoStub{peers: []Account{*relay, peer1, peer2}}
	s := &GatewayService{accountRepo: repo}
	c := relayToolStructureTestContext(86)

	resp := newRelayBlockedTestResp(http.StatusUnprocessableEntity, `{"error":{"message":"max_tokens: must be positive","type":"invalid_request_error"}}`)
	require.Nil(t, s.relayMessagesRejectedFailover(context.Background(), c, resp, relay, false, "claude-opus-5-5"))
	require.Empty(t, repo.tempUnschedCalls)

	// 响应体必须放回去给常规错误处理读取。
	body := make([]byte, 128)
	n, err := resp.Body.Read(body)
	require.NoError(t, err)
	require.Contains(t, string(body[:n]), "max_tokens")
}

// 官方 OAuth 号不会返回中转的工具结构文案，命中也不换号、不冷却，保持原处理。
func TestRelayMessagesRejectedFailover_OAuthToolStructureUnchanged(t *testing.T) {
	official := &Account{ID: 9, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true}
	peer := Account{ID: 232, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	peer2 := Account{ID: 281, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	repo := &relayQueueRejectRepoStub{peers: []Account{*official, peer, peer2}}
	s := &GatewayService{accountRepo: repo}
	c := relayToolStructureTestContext(86)

	resp := newRelayBlockedTestResp(http.StatusUnprocessableEntity, relayToolStructureRejectBody)
	require.Nil(t, s.relayMessagesRejectedFailover(context.Background(), c, resp, official, false, "claude-opus-5-5"))
	require.Empty(t, repo.tempUnschedCalls)
}
