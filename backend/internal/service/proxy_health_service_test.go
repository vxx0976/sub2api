//go:build unit

package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type healthDegradeCall struct {
	proxyID int64
	target  *int64
	change  bool
}

type healthRestoreCall struct {
	proxyID int64
}

type proxyHealthRepoFake struct {
	ProxyRepository
	proxies  map[int64]*Proxy
	fail     map[int64]int
	ok       map[int64]int
	degrades []healthDegradeCall
	restores []healthRestoreCall
}

func newProxyHealthRepoFake(proxies ...Proxy) *proxyHealthRepoFake {
	f := &proxyHealthRepoFake{proxies: map[int64]*Proxy{}, fail: map[int64]int{}, ok: map[int64]int{}}
	for i := range proxies {
		p := proxies[i]
		if p.HealthStatus == "" {
			p.HealthStatus = ProxyHealthHealthy
		}
		f.proxies[p.ID] = &p
	}
	return f
}

func (f *proxyHealthRepoFake) ListProxiesForHealthCheck(context.Context) ([]Proxy, error) {
	out := make([]Proxy, 0, len(f.proxies))
	for id := int64(1); id <= int64(len(f.proxies))+10; id++ {
		p, ok := f.proxies[id]
		if !ok || p.Status != StatusActive {
			continue
		}
		if p.FailureFallbackMode != FallbackModeNone || p.IsDegraded() {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (f *proxyHealthRepoFake) ListAllForFallback(context.Context) ([]Proxy, error) {
	out := make([]Proxy, 0, len(f.proxies))
	for _, p := range f.proxies {
		out = append(out, *p)
	}
	return out, nil
}

func (f *proxyHealthRepoFake) RecordProxyHealthResult(_ context.Context, id int64, success bool) (ProxyHealthStreak, error) {
	p, ok := f.proxies[id]
	if !ok {
		return ProxyHealthStreak{}, nil
	}
	if success {
		f.fail[id] = 0
		f.ok[id]++
	} else {
		f.ok[id] = 0
		f.fail[id]++
	}
	return ProxyHealthStreak{FailStreak: f.fail[id], OkStreak: f.ok[id], HealthStatus: p.HealthStatus, Found: true}, nil
}

func (f *proxyHealthRepoFake) MarkProxyDegraded(_ context.Context, id int64, target *int64, change bool, _ string) ([]int64, bool, error) {
	f.degrades = append(f.degrades, healthDegradeCall{proxyID: id, target: target, change: change})
	p := f.proxies[id]
	transitioned := p.HealthStatus != ProxyHealthDegraded
	p.HealthStatus = ProxyHealthDegraded
	return nil, transitioned, nil
}

func (f *proxyHealthRepoFake) MarkProxyHealthy(_ context.Context, id int64) ([]int64, bool, error) {
	p := f.proxies[id]
	if p.HealthStatus != ProxyHealthDegraded {
		return nil, false, nil
	}
	f.restores = append(f.restores, healthRestoreCall{proxyID: id})
	p.HealthStatus = ProxyHealthHealthy
	return nil, true, nil
}

type proxyHealthProberFake struct {
	mu        sync.Mutex
	errs      map[string]error
	directErr error
	calls     map[string]int
	// block 中的代理探测会一直阻塞到 ctx 结束，模拟整轮预算耗尽。
	block map[string]bool
}

func (p *proxyHealthProberFake) ProbeProxy(ctx context.Context, proxyURL string) (*ProxyExitInfo, int64, error) {
	p.mu.Lock()
	if p.calls == nil {
		p.calls = map[string]int{}
	}
	p.calls[proxyURL]++
	blocked := p.block[proxyURL]
	err := p.errs[proxyURL]
	directErr := p.directErr
	p.mu.Unlock()
	if blocked {
		<-ctx.Done()
		return nil, 0, ctx.Err()
	}
	if proxyURL == "" {
		return &ProxyExitInfo{}, 1, directErr
	}
	return &ProxyExitInfo{}, 1, err
}

func healthTestProxy(id int64, port int, mode string, backup *int64) Proxy {
	return Proxy{ID: id, Name: "p", Protocol: "socks5", Host: "10.0.0.1", Port: port, Status: StatusActive, FailureFallbackMode: mode, FailureBackupProxyID: backup}
}

func newTestProxyHealthService(repo ProxyRepository, prober ProxyExitInfoProber) *ProxyHealthService {
	svc := NewProxyHealthService(repo, prober, time.Minute)
	svc.now = func() time.Time { return time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC) }
	return svc
}

func TestProxyHealthServiceFallsBackAfterConsecutiveFailures(t *testing.T) {
	p := healthTestProxy(1, 1081, FallbackModeDirect, nil)
	repo := newProxyHealthRepoFake(p)
	prober := &proxyHealthProberFake{errs: map[string]error{p.URL(): errors.New("connection refused")}}
	svc := newTestProxyHealthService(repo, prober)

	for i := 1; i < proxyHealthFailThreshold; i++ {
		require.NoError(t, svc.checkOnce(context.Background()))
		require.Empty(t, repo.degrades, "round %d must not degrade yet", i)
	}
	require.NoError(t, svc.checkOnce(context.Background()))
	require.Len(t, repo.degrades, 1)
	require.Equal(t, int64(1), repo.degrades[0].proxyID)
	require.True(t, repo.degrades[0].change)
	require.Nil(t, repo.degrades[0].target)
	require.Equal(t, ProxyHealthDegraded, repo.proxies[1].HealthStatus)

	// 仍故障时每轮幂等重扫，改投故障期间新绑定到该代理的账号。
	require.NoError(t, svc.checkOnce(context.Background()))
	require.Len(t, repo.degrades, 2)
}

func TestProxyHealthServiceMovesToHealthyBackupProxy(t *testing.T) {
	p := healthTestProxy(1, 1081, FallbackModeProxy, i64(2))
	backup := healthTestProxy(2, 1082, FallbackModeNone, nil)
	repo := newProxyHealthRepoFake(p, backup)
	prober := &proxyHealthProberFake{errs: map[string]error{p.URL(): errors.New("timeout")}}
	svc := newTestProxyHealthService(repo, prober)

	for i := 0; i < proxyHealthFailThreshold; i++ {
		require.NoError(t, svc.checkOnce(context.Background()))
	}
	require.Len(t, repo.degrades, 1)
	require.True(t, repo.degrades[0].change)
	require.Equal(t, int64(2), *repo.degrades[0].target)
	require.Zero(t, prober.calls[backup.URL()], "backup without failure fallback is not a health check candidate")
}

func TestProxyHealthServiceSkipsRoundWhenDirectControlProbeFails(t *testing.T) {
	p := healthTestProxy(1, 1081, FallbackModeDirect, nil)
	repo := newProxyHealthRepoFake(p)
	prober := &proxyHealthProberFake{errs: map[string]error{p.URL(): errors.New("network down")}, directErr: errors.New("network down")}
	svc := newTestProxyHealthService(repo, prober)

	for i := 0; i < proxyHealthFailThreshold+2; i++ {
		require.NoError(t, svc.checkOnce(context.Background()))
	}
	require.Empty(t, repo.degrades)
	require.Zero(t, repo.fail[1], "rounds with a failing control probe must not count")
}

func TestProxyHealthServiceRestoresAfterConsecutiveSuccesses(t *testing.T) {
	p := healthTestProxy(1, 1081, FallbackModeProxy, i64(2))
	p.HealthStatus = ProxyHealthDegraded
	backup := healthTestProxy(2, 1082, FallbackModeProxy, i64(3))
	tail := healthTestProxy(3, 1083, FallbackModeNone, nil)
	repo := newProxyHealthRepoFake(p, backup, tail)
	prober := &proxyHealthProberFake{errs: map[string]error{}}
	svc := newTestProxyHealthService(repo, prober)

	for i := 1; i < proxyHealthRecoverThreshold; i++ {
		require.NoError(t, svc.checkOnce(context.Background()))
		require.Empty(t, repo.restores, "round %d must not restore yet", i)
	}
	require.NoError(t, svc.checkOnce(context.Background()))
	require.Len(t, repo.restores, 1)
	require.Equal(t, int64(1), repo.restores[0].proxyID)

	require.NoError(t, svc.checkOnce(context.Background()))
	require.Len(t, repo.restores, 1, "healthy proxy is not restored again")
}

func TestProxyHealthServiceFailureResetsRecoveryStreak(t *testing.T) {
	p := healthTestProxy(1, 1081, FallbackModeDirect, nil)
	p.HealthStatus = ProxyHealthDegraded
	repo := newProxyHealthRepoFake(p)
	prober := &proxyHealthProberFake{errs: map[string]error{}}
	svc := newTestProxyHealthService(repo, prober)

	for i := 1; i < proxyHealthRecoverThreshold; i++ {
		require.NoError(t, svc.checkOnce(context.Background()))
	}
	prober.errs[p.URL()] = errors.New("flap")
	require.NoError(t, svc.checkOnce(context.Background()))
	delete(prober.errs, p.URL())
	for i := 1; i < proxyHealthRecoverThreshold; i++ {
		require.NoError(t, svc.checkOnce(context.Background()))
	}
	require.Empty(t, repo.restores, "a failure in between must restart the recovery streak")
	require.NoError(t, svc.checkOnce(context.Background()))
	require.Len(t, repo.restores, 1)
}

func TestProxyHealthServiceKeepsProbingDegradedProxyWithFallbackDisabled(t *testing.T) {
	p := healthTestProxy(1, 1081, FallbackModeNone, nil)
	p.HealthStatus = ProxyHealthDegraded
	repo := newProxyHealthRepoFake(p)
	prober := &proxyHealthProberFake{errs: map[string]error{}}
	svc := newTestProxyHealthService(repo, prober)

	for i := 0; i < proxyHealthRecoverThreshold; i++ {
		require.NoError(t, svc.checkOnce(context.Background()))
	}
	require.Len(t, repo.restores, 1)
}

func TestProxyHealthServiceBudgetExhaustedProbesAreNotCounted(t *testing.T) {
	p := healthTestProxy(1, 1081, FallbackModeDirect, nil)
	repo := newProxyHealthRepoFake(p)
	prober := &proxyHealthProberFake{errs: map[string]error{}, block: map[string]bool{p.URL(): true}}
	svc := newTestProxyHealthService(repo, prober)

	phaseCtx, phaseCancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer phaseCancel()
	results := svc.probeAll(phaseCtx, []Proxy{p})
	require.Len(t, results, 1)
	require.True(t, results[0].skipped, "a failure caused by the phase budget is not a proxy failure")
	require.Zero(t, repo.fail[1])
}

func TestProxyHealthServiceResolvesAgainstThisRoundsDegrades(t *testing.T) {
	// backup(id=1) 与 primary(id=2) 同一轮达到阈值：primary 解析回退目标时必须看到 backup 已故障。
	backup := healthTestProxy(1, 1081, FallbackModeDirect, nil)
	primary := healthTestProxy(2, 1082, FallbackModeProxy, i64(1))
	repo := newProxyHealthRepoFake(backup, primary)
	prober := &proxyHealthProberFake{errs: map[string]error{backup.URL(): errors.New("down"), primary.URL(): errors.New("down")}}
	svc := newTestProxyHealthService(repo, prober)

	for i := 0; i < proxyHealthFailThreshold; i++ {
		require.NoError(t, svc.checkOnce(context.Background()))
	}
	var primaryCall *healthDegradeCall
	for i := range repo.degrades {
		if repo.degrades[i].proxyID == 2 {
			primaryCall = &repo.degrades[i]
		}
	}
	require.NotNil(t, primaryCall)
	require.True(t, primaryCall.change)
	require.Nil(t, primaryCall.target, "degraded backup must be skipped in favour of its direct tail")
}

func TestRedactProxyCredentials(t *testing.T) {
	p := Proxy{Password: "p@ss word"}
	msg := "dial socks5://u:p@ss word@h and socks5://u:p%40ss+word@h and p%40ss%20word"
	got := redactProxyCredentials(msg, p)
	require.NotContains(t, got, "p@ss word")
	require.NotContains(t, got, "p%40ss+word")
	require.NotContains(t, got, "p%40ss%20word")
	require.Equal(t, "plain", redactProxyCredentials("plain", Proxy{}))
}
