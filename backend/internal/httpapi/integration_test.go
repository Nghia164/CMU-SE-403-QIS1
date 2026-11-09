package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"masm/backend/internal/notify"
	"masm/backend/internal/store"
)

func realRedis(t *testing.T) *redis.Client {
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("đặt TEST_REDIS_URL để chạy với Redis thật")
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { rdb.Close() })
	return rdb
}

func TestRedisLimiterAgainstRealRedis(t *testing.T) {
	rdb := realRedis(t)
	key := "ip-" + store.NewID("")
	l := NewRedisLimiter(rdb, 3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow(key) {
			t.Fatalf("lần %d phải được phép", i+1)
		}
	}
	if l.Allow(key) {
		t.Fatal("lần 4 phải bị chặn")
	}
}

// Hai "instance" khác nhau (hai hub riêng) chia sẻ thông báo qua Redis pub/sub.
func TestRedisHubAcrossInstances(t *testing.T) {
	rdb := realRedis(t)
	a, b := notify.NewRedisHub(rdb), notify.NewRedisHub(rdb)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ch := a.Subscribe(ctx, "u_x")
	time.Sleep(300 * time.Millisecond) // chờ SUBSCRIBE có hiệu lực
	b.Publish("u_x", store.Notification{ID: "n1", Title: "Xin chào"})
	select {
	case n := <-ch:
		if n.ID != "n1" || n.Title != "Xin chào" {
			t.Fatalf("nhận %+v", n)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("instance A không nhận được thông báo từ instance B")
	}
}

// Chạy với Mailpit: TEST_SMTP_ADDR=host:1025, TEST_MAILPIT_API=http://host:8025.
func TestSMTPMailerAgainstMailpit(t *testing.T) {
	addr, api := os.Getenv("TEST_SMTP_ADDR"), os.Getenv("TEST_MAILPIT_API")
	if addr == "" || api == "" {
		t.Skip("đặt TEST_SMTP_ADDR và TEST_MAILPIT_API để chạy với Mailpit")
	}
	m := SMTPMailer{Addr: addr, From: "MASM <no-reply@masm.vn>"}
	if err := m.Send("nguoidung@example.com", "Đặt lại mật khẩu MASM", "Liên kết: https://masm.vn/dat-lai-mat-khau?token=abc"); err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(api + "/api/v1/messages")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out struct {
		Messages []struct {
			Subject string
			To      []struct{ Address string }
		}
	}
	if err := json.Unmarshal(raw, &out); err != nil || len(out.Messages) == 0 {
		t.Fatalf("Mailpit không có thư: %s (%v)", raw, err)
	}
	got := out.Messages[0]
	if got.Subject != "Đặt lại mật khẩu MASM" || len(got.To) == 0 || !strings.EqualFold(got.To[0].Address, "nguoidung@example.com") {
		t.Fatalf("thư nhận được: %+v", got)
	}
}