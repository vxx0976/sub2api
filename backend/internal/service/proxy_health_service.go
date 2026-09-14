package service

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	proxyHealthCheckInterval = time.Minute
	// proxyHealthFailThreshold 连续失败多少轮判定故障（每轮一分钟）。
	proxyHealthFailThreshold = 3
	// proxyHealthRecoverThreshold 故障后连续成功多少轮判定恢复；比失败阈值大，避免抖动时账号来回切。
	proxyHealthRecoverThreshold = 5
	proxyHealthProbeTimeout     = 25 * time.Second
	proxyHealthProbeConcurrency = 32
	// proxyHealthProbePhaseTimeout 是一轮探测阶段的总预算；超出预算没测完的代理本轮不计数。
	proxyHealthProbePhaseTimeout = 70 * time.Second
	// proxyHealthEvalTimeout 是判定与落库阶段的独立预算，不受探测阶段超时影响。
	proxyHealthEvalTimeout   = 30 * time.Second
	proxyHealthLeaderLockKey = "leader:proxy_health"
	proxyHealthLeaderLockTTL = 3 * time.Minute
	// proxyHealthRoundGateKey 限制全集群每个检查周期只跑一轮：leader 锁跑完即释放，
	// 不加闸门的话多个实例各自的 ticker 会让每个代理一分钟被计数多次，阈值被变相缩短。
	proxyHealthRoundGateKey = "leader:proxy_health_round"
)

// ProxyHealthService 周期探测配置了故障回退的代理：连续失败达到阈值后把代理标记为故障，并按
// 故障回退配置把账号改投备用代理或直连；恢复后把仍停在回退目标上的账号改回原代理。
//
// 多实例下经 leader 锁单实例执行，连续失败/成功计数落库，leader 在实例间轮换不影响判定。
// 探测失败时会先直连探测一次作对照：本实例自身网络异常时本轮不计数，避免把所有代理误判为故障。
type ProxyHealthService struct {
	proxyRepo ProxyRepository
	prober    ProxyExitInfoProber

	interval         time.Duration
	failThreshold    int
	recoverThreshold int
	now              func() time.Time

	lockCache  LeaderLockCache
	db         *sql.DB
	instanceID string

	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

func NewProxyHealthService(proxyRepo ProxyRepository, prober ProxyExitInfoProber, interval time.Duration) *ProxyHealthService {
	return &ProxyHealthService{
		proxyRepo:        proxyRepo,
		prober:           prober,
		interval:         interval,
		failThreshold:    proxyHealthFailThreshold,
		recoverThreshold: proxyHealthRecoverThreshold,
		now:              time.Now,
		instanceID:       uuid.NewString(),
		stopCh:           make(chan struct{}),
	}
}

// SetLeaderLock injects the leader-lock cache and DB used to elect a single
// instance for the periodic job. When both are nil the job runs ungated.
func (s *ProxyHealthService) SetLeaderLock(lockCache LeaderLockCache, db *sql.DB) {
	if s == nil {
		return
	}
	s.lockCache = lockCache
	s.db = db
}

func (s *ProxyHealthService) Start() {
	if s == nil || s.proxyRepo == nil || s.prober == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *ProxyHealthService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
	s.wg.Wait()
}

func (s *ProxyHealthService) runOnce() {
	lockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	release, ok := tryAcquireSingletonLeaderLock(lockCtx, s.lockCache, s.db, proxyHealthLeaderLockKey, s.instanceID, proxyHealthLeaderLockTTL)
	if !ok {
		cancel()
		return
	}
	if release != nil {
		defer release()
	}
	// 闸门锁不释放、只靠 TTL 过期，且只走 Redis（DB advisory 锁随会话释放，做不了 TTL 闸门）。
	if s.lockCache != nil {
		gateTTL := s.interval - 5*time.Second
		if gateTTL <= 0 {
			gateTTL = s.interval
		}
		acquired, err := s.lockCache.TryAcquireLeaderLock(lockCtx, proxyHealthRoundGateKey, s.instanceID, gateTTL)
		if err != nil || !acquired {
			// 闸门不可用时宁可跳过：多实例各跑一轮会让计数翻倍、判定阈值被变相缩短。
			cancel()
			return
		}
	}
	cancel()

	if err := s.checkOnce(context.Background()); err != nil {
		log.Printf("[ProxyHealth] check failed: %v", err)
	}
}

type proxyHealthProbeResult struct {
	proxy Proxy
	err   error
	// skipped 表示本轮没测出结果（探测阶段预算耗尽），不计入连续失败/成功。
	skipped bool
}

// checkOnce 执行一轮探测与判定。
func (s *ProxyHealthService) checkOnce(ctx context.Context) error {
	listCtx, cancelList := context.WithTimeout(ctx, proxyHealthEvalTimeout)
	candidates, err := s.proxyRepo.ListProxiesForHealthCheck(listCtx)
	cancelList()
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return nil
	}

	probeCtx, cancelProbe := context.WithTimeout(ctx, proxyHealthProbePhaseTimeout)
	results := s.probeAll(probeCtx, candidates)
	cancelProbe()

	anyFailed := false
	measured := 0
	for _, r := range results {
		if r.skipped {
			continue
		}
		measured++
		if r.err != nil {
			anyFailed = true
		}
	}
	if measured < len(results) {
		log.Printf("[ProxyHealth] probe phase budget exhausted, %d/%d proxies not measured this round", len(results)-measured, len(results))
	}
	if anyFailed {
		controlCtx, cancelControl := context.WithTimeout(ctx, proxyHealthProbeTimeout)
		_, _, controlErr := s.prober.ProbeProxy(controlCtx, "")
		cancelControl()
		if controlErr != nil {
			log.Printf("[ProxyHealth] direct control probe failed, skip this round: %v", controlErr)
			return nil
		}
	}

	evalCtx, cancelEval := context.WithTimeout(ctx, proxyHealthEvalTimeout)
	defer cancelEval()

	var all map[int64]Proxy
	loadAll := func() (map[int64]Proxy, error) {
		if all != nil {
			return all, nil
		}
		list, err := s.proxyRepo.ListAllForFallback(evalCtx)
		if err != nil {
			return nil, err
		}
		all = make(map[int64]Proxy, len(list))
		for _, p := range list {
			all[p.ID] = p
		}
		return all, nil
	}
	invalidate := func() { all = nil }

	for _, r := range results {
		if r.skipped {
			continue
		}
		if err := s.evaluate(evalCtx, r, loadAll, invalidate); err != nil {
			log.Printf("[ProxyHealth] proxy %d evaluate failed: %v", r.proxy.ID, err)
		}
	}
	return nil
}

func (s *ProxyHealthService) probeAll(ctx context.Context, candidates []Proxy) []proxyHealthProbeResult {
	results := make([]proxyHealthProbeResult, len(candidates))
	slots := make(chan struct{}, proxyHealthProbeConcurrency)
	var wg sync.WaitGroup
	for i := range candidates {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := candidates[i]
			select {
			case slots <- struct{}{}:
			case <-ctx.Done():
				results[i] = proxyHealthProbeResult{proxy: p, skipped: true}
				return
			}
			defer func() { <-slots }()
			probeCtx, cancel := context.WithTimeout(ctx, proxyHealthProbeTimeout)
			defer cancel()
			_, _, err := s.prober.ProbeProxy(probeCtx, p.URL())
			// 探测阶段总预算耗尽导致的失败不是代理故障。
			if err != nil && ctx.Err() != nil {
				results[i] = proxyHealthProbeResult{proxy: p, skipped: true}
				return
			}
			results[i] = proxyHealthProbeResult{proxy: p, err: err}
		}(i)
	}
	wg.Wait()
	return results
}

func (s *ProxyHealthService) evaluate(ctx context.Context, r proxyHealthProbeResult, loadAll func() (map[int64]Proxy, error), invalidate func()) error {
	streak, err := s.proxyRepo.RecordProxyHealthResult(ctx, r.proxy.ID, r.err == nil)
	if err != nil || !streak.Found {
		return err
	}

	if r.err == nil {
		if streak.HealthStatus != ProxyHealthDegraded || streak.OkStreak < s.recoverThreshold {
			return nil
		}
		accountIDs, transitioned, err := s.proxyRepo.MarkProxyHealthy(ctx, r.proxy.ID)
		if err != nil {
			return err
		}
		if transitioned {
			invalidate()
			log.Printf("[ProxyHealth] proxy %d (%s) recovered after %d ok checks, restored %d accounts", r.proxy.ID, r.proxy.Name, streak.OkStreak, len(accountIDs))
		}
		return nil
	}

	if streak.FailStreak < s.failThreshold {
		return nil
	}
	byID, err := loadAll()
	if err != nil {
		return err
	}
	target, change := ResolveProxyFailureFallbackTarget(r.proxy, byID, s.now())
	errMsg := redactProxyCredentials(sanitizeUpstreamErrorMessage(r.err.Error()), r.proxy)
	accountIDs, transitioned, err := s.proxyRepo.MarkProxyDegraded(ctx, r.proxy.ID, target, change, errMsg)
	if err != nil {
		return err
	}
	if transitioned {
		// 本轮后续代理解析回退目标时要看到这次状态变化，避免把账号改投到刚判定故障的备用代理。
		invalidate()
		switch {
		case !change && r.proxy.FailureFallbackMode == FallbackModeProxy:
			log.Printf("[ProxyHealth] proxy %d (%s) degraded after %d failed checks; failure fallback chain unresolved, accounts kept: %s", r.proxy.ID, r.proxy.Name, streak.FailStreak, errMsg)
		case target == nil && change:
			log.Printf("[ProxyHealth] proxy %d (%s) degraded after %d failed checks, %d accounts fell back to direct: %s", r.proxy.ID, r.proxy.Name, streak.FailStreak, len(accountIDs), errMsg)
		case change:
			log.Printf("[ProxyHealth] proxy %d (%s) degraded after %d failed checks, %d accounts moved to proxy %d: %s", r.proxy.ID, r.proxy.Name, streak.FailStreak, len(accountIDs), *target, errMsg)
		default:
			log.Printf("[ProxyHealth] proxy %d (%s) degraded after %d failed checks: %s", r.proxy.ID, r.proxy.Name, streak.FailStreak, errMsg)
		}
	} else if len(accountIDs) > 0 {
		log.Printf("[ProxyHealth] proxy %d (%s) still degraded, moved %d accounts to the current fallback target", r.proxy.ID, r.proxy.Name, len(accountIDs))
	}
	return nil
}

// redactProxyCredentials 去掉错误文本里可能出现的代理口令（原文与 URL 编码两种形态），
// 该文本会落库并通过后台接口返回、写入日志。
func redactProxyCredentials(msg string, p Proxy) string {
	if p.Password == "" {
		return msg
	}
	userinfoEscaped := strings.TrimPrefix(url.UserPassword("x", p.Password).String(), "x:")
	for _, secret := range []string{p.Password, userinfoEscaped, url.QueryEscape(p.Password), url.PathEscape(p.Password)} {
		if secret != "" {
			msg = strings.ReplaceAll(msg, secret, "***")
		}
	}
	return msg
}
