package ai

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type task struct {
	JobID string `json:"job_id"`
	In    Input  `json:"in"`
}

// Queue là hàng đợi job AI. Bản Redis cho phép nhiều instance backend cùng chia việc.
type Queue interface {
	Push(ctx context.Context, t task) error
	// Pop chờ tới khi có task; trả lỗi khi ctx bị huỷ.
	Pop(ctx context.Context) (task, error)
}

var ErrQueueFull = errors.New("hàng đợi AI đang đầy")

const maxQueued = 256

/* ---- Trong tiến trình (dev, một instance) ---- */

type memQueue struct{ ch chan task }

func NewMemQueue() Queue { return &memQueue{ch: make(chan task, maxQueued)} }

func (q *memQueue) Push(_ context.Context, t task) error {
	select {
	case q.ch <- t:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *memQueue) Pop(ctx context.Context) (task, error) {
	select {
	case t := <-q.ch:
		return t, nil
	case <-ctx.Done():
		return task{}, ctx.Err()
	}
}

/* ---- Redis ---- */

type redisQueue struct {
	rdb *redis.Client
	key string
}

func NewRedisQueue(rdb *redis.Client) Queue { return &redisQueue{rdb: rdb, key: "masm:ai:queue"} }

func (q *redisQueue) Push(ctx context.Context, t task) error {
	if n, err := q.rdb.LLen(ctx, q.key).Result(); err != nil {
		return err
	} else if n >= maxQueued {
		return ErrQueueFull
	}
	b, _ := json.Marshal(t)
	return q.rdb.LPush(ctx, q.key, b).Err()
}

func (q *redisQueue) Pop(ctx context.Context) (task, error) {
	for {
		// BRPOP chặn tối đa 2 giây rồi lặp lại để kịp thoát khi ctx bị huỷ.
		res, err := q.rdb.BRPop(ctx, 2*time.Second, q.key).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			if ctx.Err() != nil {
				return task{}, ctx.Err()
			}
			time.Sleep(time.Second) // Redis tạm lỗi: thử lại
			continue
		}
		var t task
		if json.Unmarshal([]byte(res[1]), &t) == nil {
			return t, nil
		}
	}
}
