package strategy

import (
	"fmt"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"log"
	"sync"
	"time"
)

type LeakyBucketManager struct {
	buckets map[string]map[string]*types.LeakyBucket // api_id -> ip_address -> bucket
	mu      sync.Mutex
	handler types.ProxyHandler
	config  settings.Api
}

func NewLeakyBucketManager(handler types.ProxyHandler, apis []settings.Api) *LeakyBucketManager {
	m := &LeakyBucketManager{
		buckets: make(map[string]map[string]*types.LeakyBucket),
		handler: handler,
	}
	// 맵 초기화
	for _, api := range apis {
		m.buckets[api.Identifier] = make(map[string]*types.LeakyBucket)
		go m.startScheduling(api) // 스케줄링 시작
	}

	return m
}

func (m *LeakyBucketManager) AddRequest(
	apiIdentifier string,
	key string,
	req types.QueuedRequest,
	api types.ApiMatchResult,
) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	bucket, ok := m.buckets[apiIdentifier][key]
	if !ok {
		bucket = &types.LeakyBucket{
			Queue:           make(chan types.QueuedRequest),
			BucketSize:      api.BucketSize,
			LastProcessTime: time.Now(),
		}
		m.buckets[apiIdentifier][key] = bucket
	}
	// 버킷이 가득 차있으면 false를 반환하고, 여유공간이 있으면 req를 큐에 넣고 true를 반환한다.
	select {
	case bucket.Queue <- req:
		return true
	default:
		return false
	}
}

func (m *LeakyBucketManager) CountBucketFreeCapacity(apiIdentifier string, key string) (int, error) {
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
	bucket, ok := m.buckets[apiIdentifier][key]
	if !ok {
		return 0, fmt.Errorf("No Bucket Found: api=%s, key=%s\n", apiIdentifier, key)
	}

	// 현재 시간에 요청이 불가능하다는 것은 요청을 처리한 후 아직 리필타임이 찾아오지 않았다는 것을 의미한다.
	// 마지막 작업시간 + 리필타임 - 현재시간 으로 계산하면 처리되기까지 남은 시간을 알 수 있다. (= 새로운 요청을 삽입할 수 있는 시간)
	seconds := bucket.LastProcessTime.Add(time.Duration(api.RefillSeconds)).Sub(time.Now()).Seconds()
	if seconds <= 0 {
		return 0, nil
	}
	return int(seconds), nil
}

func (m *LeakyBucketManager) startScheduling(api settings.Api) {
	ticker := time.NewTicker(time.Duration(api.RefillSeconds) * time.Second)
	log.Printf("%s Ticker Start\n", api.Identifier)
	defer ticker.Stop() // for range ticker.C가 끝나지 않는 이상 함수가 리턴되지 않으니 Stop은 프로그램종료전까지는 절대 호출되지 않는다.

	// TODO 중첩 for 문으로부터 해방될 방법은 없는가
	for range ticker.C {
		log.Printf("%s_bucket 검사\n", api.Identifier)
		// 락 적용위치 최소화를 위해
		m.mu.Lock()
		buckets := make([]*types.LeakyBucket, 0, len(m.buckets[api.Identifier]))
		for _, b := range m.buckets[api.Identifier] {
			buckets = append(buckets, b)
		}
		m.mu.Unlock()

		// 락을 풀고 실제 요청 처리
		for _, bucket := range buckets {
		drain:
			for {
				select {
				case req := <-bucket.Queue:
					m.handler.ToOrigin(req.Writer, req.Request, api.Target)
				default:
					break drain
				}
			}
		}

	}
}
