package types

import "time"

type TokenBucket struct {
	Token          int       `json:"token" redisclient:"token"`
	LastRefillTime time.Time `json:"last_refill_time" redisclient:"last_refill_time"`
}

func NewTokenBucket(token int) *TokenBucket {
	return &TokenBucket{Token: token, LastRefillTime: time.Now()}
}
