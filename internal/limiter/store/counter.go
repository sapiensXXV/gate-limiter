package store

import "context"

type CounterResult struct {
	Count int64
}

type CounterStore interface {
	IncrementAndGet(ctx context.Context, key string, windowSeconds int) (CounterResult, error)
}
