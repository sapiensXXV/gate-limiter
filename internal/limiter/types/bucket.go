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

// LeakyBucket 누출버킷을 표현하는 구조체
type LeakyBucket struct {
	Queue      chan QueuedRequest
	BucketSize int
}
