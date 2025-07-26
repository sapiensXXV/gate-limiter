package strategy

import (
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
	handler limiter.DefaultProxyHandler
}

func (m *LeakyBucketManager) AddRequest(key string, req QueuedRequest) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	bucket, ok := m.bucket[key]
	if !ok {
		bucket = &LeakyBucket{queue: make(chan QueuedRequest), bucketSize: 100}
		m.bucket[key] = bucket
	}

	select {
	case bucket.queue <- req:
		return true
	default:
		return false
	}
}

func (m *LeakyBucketManager) StartScheduling(api ApiMatchResult) {
	go func() {
		ticker := time.NewTicker(time.Duration(api.WindowSeconds) * time.Second)
		for range ticker.C {
			m.mutex.Lock()
			for key, bucket := range m.bucket {
				select {
				case req := <-bucket.queue:
					m.handler.ToOrigin(req.Writer, req.Request)
				default:
					// 큐에 요청이 없는 경우는 무시
				}
			}
			m.mutex.Unlock()
		}
	}()
}

//func NewLeakyBucketManager(process func(bucket.LeakyBucketRequest)) *LeakyBucketManager {
//	return &LeakyBucketManager{
//		bucket:  make(map[string]*bucket.LeakyBucket),
//		process: process,
//	}
//}
//func (m *LeakyBucketManager) StartScheduling(rate time.Duration) {
//	go func() {
//		ticker := time.NewTicker(rate) // rate 간격으로 ticker
//		defer ticker.Stop()
//		for range ticker.C {
//			m.mutex.Lock()
//			for _, bucket := range m.bucket {
//				select {
//				case _ := <-bucket.Queue:
//					go m.process()
//				default:
//					//비어있으면 패스
//				}
//			}
//			m.mutex.Unlock()
//		}
//	}()
//}
//
//func (m *LeakyBucketManager) AddRequest(key string, request bucket.LeakyBucketRequest, bucketSize int) bool {
//	m.mutex.Lock()
//	defer m.mutex.Unlock()
//
//	bc, exists := m.bucket[key]
//	if !exists {
//		bc = &bucket.LeakyBucket{
//			Queue: make(chan bucket.LeakyBucketRequest, bucketSize),
//		}
//		m.bucket[key] = bc
//	}
//
//	select {
//	case bc.Queue <- request:
//		return true
//	default:
//		return false
//	}
//}
