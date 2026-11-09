package httpapi

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"masm/backend/internal/store"
)

const userKey = "user"

type accessClaims struct {
	jwt.RegisteredClaims
	Ver int `json:"ver"`
}

func (s *Server) issueToken(u store.User) (string, error) {
	claims := accessClaims{RegisteredClaims: jwt.RegisteredClaims{
		Subject: u.ID, IssuedAt: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.TokenTTL)),
	}, Ver: u.TokenVersion}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
}

// auth xác thực Bearer JWT, nạp người dùng và chặn tài khoản bị khoá.
func (s *Server) auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			fail(c, 401, "unauthorized", "Bạn cần đăng nhập.")
			return
		}
		tok, err := jwt.ParseWithClaims(strings.TrimPrefix(h, "Bearer "), &accessClaims{},
			func(*jwt.Token) (any, error) { return []byte(s.cfg.JWTSecret), nil },
			jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !tok.Valid {
			fail(c, 401, "unauthorized", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
			return
		}
		sub, _ := tok.Claims.GetSubject()
		u, err := s.repo.UserByID(sub)
		if err != nil {
			fail(c, 401, "unauthorized", "Tài khoản không tồn tại.")
			return
		}
		if u.Status == "locked" {
			fail(c, 403, "account_locked", "Tài khoản của bạn đã bị khoá.")
			return
		}
		if claims, ok := tok.Claims.(*accessClaims); !ok || claims.Ver != u.TokenVersion {
			fail(c, 401, "unauthorized", "Mật khẩu đã thay đổi, bạn vui lòng đăng nhập lại.")
			return
		}
		c.Set(userKey, u)
		c.Next()
	}
}

func (s *Server) role(roles ...store.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := cur(c)
		for _, r := range roles {
			if u.Role == r {
				return
			}
		}
		fail(c, 403, "forbidden", "Bạn không có quyền thực hiện thao tác này.")
	}
}

func cur(c *gin.Context) store.User { return c.MustGet(userKey).(store.User) }

/* ---------- Giới hạn tốc độ theo IP (chống dò mật khẩu) ---------- */

type memLimiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	hits   map[string][]time.Time
}

func newMemLimiter(max int, w time.Duration) *memLimiter {
	return &memLimiter{max: max, window: w, hits: map[string][]time.Time{}}
}

func (l *memLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	kept := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if now.Sub(t) < l.window {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.max {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	return true
}

func (s *Server) rateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.logins.Allow(c.ClientIP()) {
			fail(c, 429, "rate_limited", "Bạn thử quá nhiều lần. Vui lòng đợi một phút.")
			return
		}
	}
}

/* ---------- Metrics (định dạng Prometheus) ---------- */

type metrics struct {
	mu     sync.Mutex
	counts map[string]int64
	sumSec map[string]float64
}

func newMetrics() *metrics { return &metrics{counts: map[string]int64{}, sumSec: map[string]float64{}} }

func (m *metrics) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		key := fmt.Sprintf(`method="%s",path="%s",status="%d"`, c.Request.Method, path, c.Writer.Status())
		m.mu.Lock()
		m.counts[key]++
		m.sumSec[key] += time.Since(start).Seconds()
		m.mu.Unlock()
	}
}

func (m *metrics) handler(c *gin.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]string, 0, len(m.counts))
	for k := range m.counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("# TYPE masm_http_requests_total counter\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "masm_http_requests_total{%s} %d\n", k, m.counts[k])
	}
	b.WriteString("# TYPE masm_http_request_seconds_sum counter\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "masm_http_request_seconds_sum{%s} %f\n", k, m.sumSec[k])
	}
	c.Data(200, "text/plain; version=0.0.4", []byte(b.String()))
}
