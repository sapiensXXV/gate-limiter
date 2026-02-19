package types

import "time"

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
