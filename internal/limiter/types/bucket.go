package types

import "time"

// TokenBucket 토큰버킷을 표현하는 구조체
type TokenBucket struct {
	Token          int       `json:"token"`
	LastRefillTime time.Time `json:"last_refill_time"`
}

func NewTokenBucket(token int) *TokenBucket {
	return &TokenBucket{Token: token, LastRefillTime: time.Now()}
}

// LeakyQueueItem leaky-bucket 큐에 들어가는 항목
// 응답을 비동기로 쓰지 않기 위해서 Writer는 저장하지 않는다
type LeakyQueueItem struct {
	Done       chan struct{}
	EnqueuedAt time.Time
}

// LeakyBucket 누출버킷을 표현하는 구조체
type LeakyBucket struct {
	Queue           chan *LeakyQueueItem
	BucketSize      int
	LastProcessTime time.Time
}
