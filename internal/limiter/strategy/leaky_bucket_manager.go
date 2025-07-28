package strategy

import (
	"fmt"
	config_ratelimiter "gate-limiter/config/limiterconfig"
	"gate-limiter/internal/limiter"
	"log"
	"net/http"
	"sync"
	"time"
)

type LeakyBucket struct {
	queue      chan QueuedRequest
	bucketSize int
}

type QueuedRequest struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

type LeakyBucketManager struct {
	buckets map[string]map[string]*LeakyBucket // api_id -> ip_address -> bucket
	mu      sync.Mutex
	handler limiter.ProxyHandler
	config  config_ratelimiter.Api
}

func NewLeakyBucketManager(handler limiter.ProxyHandler, apis []config_ratelimiter.Api) *LeakyBucketManager {
	m := &LeakyBucketManager{
		buckets: make(map[string]map[string]*LeakyBucket),
		handler: handler,
	}
	// 맵 초기화
	for _, api := range apis {
		m.buckets[api.Identifier] = make(map[string]*LeakyBucket)
		go m.startScheduling(api) // 스케줄링 시작
	}

	return m
}

func (m *LeakyBucketManager) AddRequest(
	apiIdentifier string,
	key string,
	req QueuedRequest,
	api ApiMatchResult,
) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	bucket, ok := m.buckets[apiIdentifier][key]
	if !ok {
		bucket = &LeakyBucket{queue: make(chan QueuedRequest), bucketSize: api.BucketSize}
		m.buckets[apiIdentifier][key] = bucket
	}
	// 버킷이 가득 차있으면 false를 반환하고, 여유공간이 있으면 req를 큐에 넣고 true를 반환한다.
	select {
	case bucket.queue <- req:
		return true
	default:
		return false
	}
}

func (m *LeakyBucketManager) CountBucketFreeCapacity(apiIdentifier string, key string) (int, error) {
	bucket, ok := m.buckets[apiIdentifier][key]
	if !ok {
		return 0, fmt.Errorf("No Bucket Found: key=%s\n", key)
	}
	// 채널의 용량과 현재길이를 빼면 여유공간을 알 수 있다.
	return cap(bucket.queue) - len(bucket.queue), nil
}

func (m *LeakyBucketManager) startScheduling(api config_ratelimiter.Api) {
	ticker := time.NewTicker(time.Duration(api.WindowSeconds) * time.Second)
	log.Printf("%s Ticker Start\n", api.Identifier)
	defer ticker.Stop() // for range ticker.C가 끝나지 않는 이상 함수가 리턴되지 않으니 Stop은 프로그램종료전까지는 절대 호출되지 않는다.

	// TODO 중첩 for 문으로부터 해방될 방법은 없는가
	for range ticker.C {
		log.Printf("%s_bucket 검사\n", api.Identifier)
		// 락 적용위치 최소화를 위해
		m.mu.Lock()
		buckets := make([]*LeakyBucket, 0, len(m.buckets[api.Identifier]))
		for _, b := range m.buckets[api.Identifier] {
			buckets = append(buckets, b)
		}
		m.mu.Unlock()

		// 락을 풀고 실제 요청 처리
		for _, bucket := range buckets {
		drain:
			for {
				select {
				case req := <-bucket.queue:
					m.handler.ToOrigin(req.Writer, req.Request, api.Target)
				default:
					break drain
				}
			}
		}

	}
}
