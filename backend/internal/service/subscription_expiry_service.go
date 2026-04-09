package service

import (
	"context"
	"log"
	"sync"
	"time"
)

// SubscriptionExpiryService periodically updates expired subscription status.
type SubscriptionExpiryService struct {
	userSubRepo UserSubscriptionRepository
	locker      *LeaderLocker
	interval    time.Duration
	stopCh      chan struct{}
	stopOnce    sync.Once
	wg          sync.WaitGroup
}

func NewSubscriptionExpiryService(userSubRepo UserSubscriptionRepository, interval time.Duration, locker *LeaderLocker) *SubscriptionExpiryService {
	return &SubscriptionExpiryService{
		userSubRepo: userSubRepo,
		locker:      locker,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

func (s *SubscriptionExpiryService) Start() {
	if s == nil || s.userSubRepo == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
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

func (s *SubscriptionExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *SubscriptionExpiryService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	release, ok := s.locker.TryAcquire(ctx, "leader:subscription_expiry", 2*time.Minute)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}

	updated, err := s.userSubRepo.BatchUpdateExpiredStatus(ctx)
	if err != nil {
		log.Printf("[SubscriptionExpiry] Update expired subscriptions failed: %v", err)
		return
	}
	if updated > 0 {
		log.Printf("[SubscriptionExpiry] Updated %d expired subscriptions", updated)
	}
}
