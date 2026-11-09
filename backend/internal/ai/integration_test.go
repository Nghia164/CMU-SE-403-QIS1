package ai

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// Chạy với Redis thật: đặt TEST_REDIS_URL.
func TestRedisQueueAgainstRealRedis(t *testing.T) {
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("đặt TEST_REDIS_URL để chạy với Redis thật")
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()
	rdb.Del(context.Background(), "masm:ai:queue")
	q := NewRedisQueue(rdb)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// BRPOP đang chờ thì Push phải đánh thức nó ngay
	got := make(chan task, 1)
	go func() { tk, _ := q.Pop(ctx); got <- tk }()
	time.Sleep(300 * time.Millisecond)
	if err := q.Push(ctx, task{JobID: "job_real", In: Input{ImageKeys: []string{"k"}}}); err != nil {
		t.Fatal(err)
	}
	select {
	case tk := <-got:
		if tk.JobID != "job_real" {
			t.Fatalf("pop = %+v", tk)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Pop không nhận được task")
	}
}