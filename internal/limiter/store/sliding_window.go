package store

import "context"

type SlidingWindowResult struct {
	Allowed       bool
	Remaining     int
	RetryAfterSec int
}

type SlidingWindowLogStore interface {
	Allow(ctx context.Context, key string, limit, windowSec int, nowUnix int64, member string) (SlidingWindowResult, error)
}

type SlidingWindowCounterStore interface {
	Allow(ctx context.Context, key string, limit, windowSec int, nowUnix int64, member string) (SlidingWindowResult, error)
}
