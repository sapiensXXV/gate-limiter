package bucket

import "time"

type TokenBucket struct {
	Token          int       `json:"token" redis:"token"`
	LastRefillTime time.Time `json:"last_refill_time" redis:"last_refill_time"`
}

func NewTokenBucket(token int) *TokenBucket {
	return &TokenBucket{Token: token, LastRefillTime: time.Now()}
}
