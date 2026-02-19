package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/pkg/redisclient"
	"net"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

// newTestRedisClient 는 miniredis 기반의 테스트용 RedisClient를 생성한다.
// 테스트 종료 시 자동으로 정리된다.
func newTestRedisClient(t *testing.T) (*miniredis.Miniredis, types.RedisClient) {
	t.Helper()
	mr := miniredis.RunT(t)
	host, portStr, _ := net.SplitHostPort(mr.Addr())
	port, _ := strconv.Atoi(portStr)
	return mr, redisclient.NewDefaultRedisClient(&settings.RedisClientConfig{
		Host: host,
		Port: port,
	})
}
