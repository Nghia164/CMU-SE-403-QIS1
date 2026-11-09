package httpapi

import (
	"fmt"
	"net/http"
	"time"

	"encoding/json"
	"github.com/gin-gonic/gin"
)

// streamNotifications là luồng Server-Sent Events: mỗi thông báo mới được đẩy ngay tới client.
// Client dùng fetch streaming (có header Authorization), không dùng EventSource vì nó không gửi được header.
func (s *Server) streamNotifications(c *gin.Context) {
	fl, ok := c.Writer.(http.Flusher)
	if !ok || s.hub == nil {
		fail(c, 501, "not_supported", "Máy chủ chưa hỗ trợ thông báo realtime.")
		return
	}
	// Server có WriteTimeout 30 giây; luồng này phải sống lâu hơn.
	_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Time{})
	h := c.Writer.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("X-Accel-Buffering", "no") // Nginx không được gom đệm
	c.Status(200)
	fmt.Fprint(c.Writer, ": connected\n\n")
	fl.Flush()

	ch := s.hub.Subscribe(c.Request.Context(), cur(c).ID)
	tick := time.NewTicker(25 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case n, ok := <-ch:
			if !ok {
				return
			}
			b, _ := json.Marshal(n)
			fmt.Fprintf(c.Writer, "event: notification\ndata: %s\n\n", b)
			fl.Flush()
		case <-tick.C:
			fmt.Fprint(c.Writer, ": ping\n\n")
			fl.Flush()
		}
	}
}
