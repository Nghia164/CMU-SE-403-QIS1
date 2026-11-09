package ai

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisQueueRoundTrip(t *testing.T) {
	mr := miniredis.RunT(t)
	q := NewRedisQueue(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, id := range []string{"job_1", "job_2"} {
		if err := q.Push(ctx, task{JobID: id, In: Input{ImageKeys: []string{"k"}}}); err != nil {
			t.Fatal(err)
		}
	}
	for _, want := range []string{"job_1", "job_2"} { // FIFO
		got, err := q.Pop(ctx)
		if err != nil || got.JobID != want || got.In.ImageKeys[0] != "k" {
			t.Fatalf("pop = %+v, %v; muốn %s", got, err, want)
		}
	}
	cctx, ccancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer ccancel()
	if _, err := q.Pop(cctx); err == nil {
		t.Fatal("hàng đợi rỗng phải dừng khi ctx hết hạn")
	}
}
