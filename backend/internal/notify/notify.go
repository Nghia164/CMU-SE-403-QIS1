// Package notify phát thông báo realtime tới người dùng đang mở kết nối SSE.
// Bản Redis (pub/sub) cho phép nhiều instance backend: thông báo tạo ở instance này
// vẫn tới người dùng đang kết nối vào instance khác.
package notify

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/redis/go-redis/v9"

	"masm/backend/internal/store"
)

type Hub interface {
	Publish(userID string, n store.Notification)
	// Subscribe trả kênh nhận thông báo của userID tới khi ctx bị huỷ.
	Subscribe(ctx context.Context, userID string) <-chan store.Notification
}

/* ---- Trong tiến trình ---- */

type memHub struct {
	mu   sync.Mutex
	subs map[string]map[chan store.Notification]struct{}
}

func NewMemHub() Hub { return &memHub{subs: map[string]map[chan store.Notification]struct{}{}} }

func (h *memHub) Publish(userID string, n store.Notification) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs[userID] {
		select {
		case ch <- n:
		default: // client chậm: bỏ qua, nó vẫn lấy được qua GET /notifications
		}
	}
}

func (h *memHub) Subscribe(ctx context.Context, userID string) <-chan store.Notification {
	ch := make(chan store.Notification, 16)
	h.mu.Lock()
	if h.subs[userID] == nil {
		h.subs[userID] = map[chan store.Notification]struct{}{}
	}
	h.subs[userID][ch] = struct{}{}
	h.mu.Unlock()
	go func() {
		<-ctx.Done()
		h.mu.Lock()
		delete(h.subs[userID], ch)
		h.mu.Unlock()
	}()
	return ch
}

/* ---- Redis pub/sub ---- */

type redisHub struct{ rdb *redis.Client }

func NewRedisHub(rdb *redis.Client) Hub { return &redisHub{rdb: rdb} }

func channel(userID string) string { return "masm:notif:" + userID }

func (h *redisHub) Publish(userID string, n store.Notification) {
	b, _ := json.Marshal(n)
	_ = h.rdb.Publish(context.Background(), channel(userID), b).Err()
}

func (h *redisHub) Subscribe(ctx context.Context, userID string) <-chan store.Notification {
	ps := h.rdb.Subscribe(ctx, channel(userID))
	out := make(chan store.Notification, 16)
	go func() {
		defer close(out)
		defer ps.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-ps.Channel():
				if !ok {
					return
				}
				var n store.Notification
				if json.Unmarshal([]byte(m.Payload), &n) == nil {
					select {
					case out <- n:
					default:
					}
				}
			}
		}
	}()
	return out
}

/* ---- Repo bọc: mọi thông báo mới đều được phát realtime ---- */

type Repo struct {
	store.Repo
	Hub Hub
}

func (r Repo) AddNotification(n store.Notification) store.Notification {
	n = r.Repo.AddNotification(n)
	r.Hub.Publish(n.UserID, n)
	return n
}
