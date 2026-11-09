package httpapi

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisLimiter(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	l := NewRedisLimiter(rdb, 3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("lần %d phải được phép", i+1)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Fatal("lần thứ 4 phải bị chặn")
	}
	if !l.Allow("5.6.7.8") {
		t.Fatal("IP khác không bị ảnh hưởng")
	}
	mr.Close()
	if !l.Allow("1.2.3.4") {
		t.Fatal("Redis lỗi thì phải cho qua")
	}
}
