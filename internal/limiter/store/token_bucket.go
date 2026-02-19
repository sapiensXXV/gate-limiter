package store

import "context"

type TokenBucketResult struct {
	Allowed       bool
	Remaining     int
	RetryAfterSec int
}

type TokenBucketStore interface {
	TryConsume(ctx context.Context, key string, maxTokens, refillSec, expireSec int) (TokenBucketResult, error)
}
