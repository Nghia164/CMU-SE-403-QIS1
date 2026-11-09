package httpapi

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter quyết định một khoá (IP) còn được phép gọi hay không.
type Limiter interface{ Allow(key string) bool }

// NewRedisLimiter đếm theo cửa sổ cố định trong Redis, dùng chung cho mọi instance.
// Nếu Redis lỗi thì cho qua (đăng nhập vẫn bị chặn bởi khoá tài khoản sau 5 lần sai).
func NewRedisLimiter(rdb *redis.Client, max int, window time.Duration) Limiter {
	return &redisLimiter{rdb: rdb, max: max, window: window}
}

type redisLimiter struct {
	rdb    *redis.Client
	max    int
	window time.Duration
}

func (l *redisLimiter) Allow(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	k := fmt.Sprintf("masm:rl:%s:%d", key, time.Now().UnixNano()/int64(l.window))
	pipe := l.rdb.TxPipeline()
	n := pipe.Incr(ctx, k)
	pipe.Expire(ctx, k, l.window)
	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("redis rate limit: %v", err)
		return true
	}
	return n.Val() <= int64(l.max)
}
