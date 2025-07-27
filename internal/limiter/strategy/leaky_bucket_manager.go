package strategy

import (
	"fmt"
	"gate-limiter/internal/limiter"
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
	bucket  map[string]*LeakyBucket
	mutex   sync.Mutex
	handler limiter.ProxyHandler
}

func NewLeakyBucketManager(handler limiter.ProxyHandler) *LeakyBucketManager {
	return &LeakyBucketManager{
		bucket:  make(map[string]*LeakyBucket),
		mutex:   sync.Mutex{},
		handler: handler,
	}
}

func (m *LeakyBucketManager) AddRequest(key string, req QueuedRequest) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	bucket, ok := m.bucket[key]
	if !ok {
		bucket = &LeakyBucket{queue: make(chan QueuedRequest), bucketSize: 100}
		m.bucket[key] = bucket
	}

	// 버킷이 가득 차있으면 false를 반환하고, 여유공간이 있으면 req를 큐에 넣고 true를 반환한다.
	select {
	case bucket.queue <- req:
		return true
	default:
		return false
	}
}

func (m *LeakyBucketManager) CountBucketFreeCapacity(key string) (int, error) {
	bucket, ok := m.bucket[key]
	if !ok {
		return 0, fmt.Errorf("No Bucket Found: key=%s\n", key)
	}
	// 채널의 용량과 현재길이를 빼면 여유공간을 알 수 있다.
	return cap(bucket.queue) - len(bucket.queue), nil
}

func (m *LeakyBucketManager) StartScheduling(api ApiMatchResult) {
	go func() {
		ticker := time.NewTicker(time.Duration(api.WindowSeconds) * time.Second)
		for range ticker.C {
			m.mutex.Lock()
			for _, bucket := range m.bucket {
				select {
				case req := <-bucket.queue:
					m.handler.ToOrigin(req.Writer, req.Request, api.Target)
				default:
					// 큐에 요청이 없는 경우는 무시
					// TODO: 사용자에게 요청이 버려졌음을 알려야한다.
				}
			}
			m.mutex.Unlock()
		}
	}()
}
