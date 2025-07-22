package redisclient

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

func AddToSortedSet(key string, member string, t time.Time) error {
	score := float64(t.Unix())
	z := &redis.Z{
		Score:  score,  // 정렬 기준
		Member: member, // 실제 저장될 값
	}
	return Rdb.ZAdd(ctx, key, *z).Err()
}

func RemoveOldEntries(key string, before time.Time) error {
	score := float64(before.Unix())
	return Rdb.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%f", score)).Err()

}
