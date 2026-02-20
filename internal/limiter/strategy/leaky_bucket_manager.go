package strategy

import (
	"context"
	"fmt"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"log/slog"
	"math"
	"sync"
	"time"
)

type LeakyBucketManager struct {
	buckets map[string]map[string]*types.LeakyBucket // api_id -> client_id -> bucket
	mu      sync.Mutex
	handler types.ProxyHandler
	config  settings.Api
}

func NewLeakyBucketManager(
	ctx context.Context,
	apis []settings.Api,
) *LeakyBucketManager {
	m := &LeakyBucketManager{
		buckets: make(map[string]map[string]*types.LeakyBucket),
	}
	for _, api := range apis {
		m.buckets[api.Identifier] = make(map[string]*types.LeakyBucket)
		go m.startScheduling(ctx, api)
	}

	return m
}

func (m *LeakyBucketManager) Enqueue(
	apiIdentifier string,
	key string,
	api types.ApiMatchResult,
) (*types.LeakyQueueItem, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	buckets, ok := m.buckets[apiIdentifier]
	if !ok {
		buckets = make(map[string]*types.LeakyBucket)
		m.buckets[apiIdentifier] = buckets
	}

	bucket, ok := buckets[key]
	if !ok {
		bucket = &types.LeakyBucket{
			Queue:           make(chan *types.LeakyQueueItem, api.Limit),
			BucketSize:      api.Limit,
			LastProcessTime: time.Now(),
		}
		buckets[key] = bucket
	}

	item := &types.LeakyQueueItem{
		Done:       make(chan struct{}),
		EnqueuedAt: time.Now(),
	}

	select {
	case bucket.Queue <- item:
		return item, true
	default:
		return nil, false
	}

}

func (m *LeakyBucketManager) CountBucketFreeCapacity(apiIdentifier string, key string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bucket, ok := m.buckets[apiIdentifier][key]
	if !ok {
		return 0, fmt.Errorf("No Bucket Found: api=%s key=%s\n", apiIdentifier, key)
	}
	return cap(bucket.Queue) - len(bucket.Queue), nil
}

func (m *LeakyBucketManager) CalcRetryTimeAfter(
	apiIdentifier string,
	key string,
	api types.ApiMatchResult,
) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bucket, ok := m.buckets[apiIdentifier][key]
	if !ok {
		return 0, fmt.Errorf("No Bucket Found: api=%s, key=%s\n", apiIdentifier, key)
	}

	interval := time.Duration(api.RefillSeconds) * time.Second
	if interval <= 0 {
		return 0, nil
	}

	nextLeakAt := bucket.LastProcessTime.Add(interval)
	wait := nextLeakAt.Sub(time.Now())
	if wait <= 0 {
		return 0, nil
	}

	return int(math.Ceil(wait.Seconds())), nil
}

func (m *LeakyBucketManager) startScheduling(ctx context.Context, api settings.Api) {
	interval := time.Duration(api.RefillSeconds) * time.Second
	if interval <= 0 {
		interval = time.Second
	}

	ticker := time.NewTicker(interval)
	slog.Info("leaky-bucket scheduler started", "api", api.Identifier, "interval", interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("leaky-bucket scheduler stopped", "api", api.Identifier)
			return
		case <-ticker.C:
			m.processBuckets(api.Identifier)
		}
	}
}

func (m *LeakyBucketManager) processBuckets(identifier string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for key, bucket := range m.buckets[identifier] {
		if now.Sub(bucket.LastProcessTime) > 5*time.Minute {
			delete(m.buckets[identifier], key)
			continue
		}

		select {
		case item := <-bucket.Queue:
			bucket.LastProcessTime = time.Now()
			close(item.Done)
		default:
		}
	}
}
