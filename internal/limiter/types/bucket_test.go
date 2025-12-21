package types

import (
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	initialTokens := 10
	tb := NewTokenBucket(initialTokens)

	// 1. 초기화 검증
	if tb.Token != initialTokens {
		t.Errorf("Expected %d tokens, got %d", initialTokens, tb.Token)
	}

	// 2. 토큰 소모 로직 테스트 (Take 메서드가 있다고 가정)
	// tb.Take(5)
	// if tb.Token != 5 { ... }

	// 3. 시간 경과에 따른 리필 테스트
	// 실제 time.Sleep을 쓰기보다, LastRefillTime을 과거로 조작하여 테스트하는 것이 빠릅니다.
	tb.LastRefillTime = time.Now().Add(-time.Second * 2)
	// 이후 Refill 기능을 호출하여 토큰이 찼는지 확인
}
