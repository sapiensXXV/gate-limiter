package strategy

import (
	"context"
	"fmt"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"log"
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
	// 맵 초기화 + API별 스케줄러 시작
	for _, api := range apis {
		m.buckets[api.Identifier] = make(map[string]*types.LeakyBucket)
		go m.startScheduling(ctx, api) // 스케줄링 시작
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
		m.buckets[apiIdentifier] = buckets // 만든 맵을 원본에 연결
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
	// 채널의 용량과 현재길이를 빼면 여유공간을 알 수 있다.
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
	log.Printf("%s leaky-bucket scheduler started (interval=%s)\n", api.Identifier, interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("%s leaky-bucket scheduler stopped\n", api.Identifier)
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
		// 5분 이상 큐에 들어온 요청이 없으면 버킷 삭제(클린업)
		if now.Sub(bucket.LastProcessTime) > 5*time.Minute {
			delete(m.buckets[identifier], key)
			continue
		}

		select {
		case item := <-bucket.Queue:
			bucket.LastProcessTime = time.Now()
			close(item.Done)
		default:
			// 큐가 비어있으면 즉시 패스
		}
	}
}
