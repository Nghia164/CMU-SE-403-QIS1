package httpapi

import (
	"github.com/gin-gonic/gin"

	"masm/backend/internal/store"
)

// notify tạo thông báo trong ứng dụng cho một người dùng.
func (s *Server) notify(userID, kind, title, body, link string) {
	s.repo.AddNotification(store.Notification{UserID: userID, Kind: kind, Title: title, Body: body, Link: link})
}

func (s *Server) listNotifications(c *gin.Context) {
	items, unread := s.repo.NotificationsOf(cur(c).ID, 30)
	if items == nil {
		items = []store.Notification{}
	}
	c.JSON(200, gin.H{"items": items, "unread": unread})
}

func (s *Server) readNotification(c *gin.Context) {
	if s.repo.MarkNotificationRead(cur(c).ID, c.Param("id")) != nil {
		fail(c, 404, "not_found", "Không tìm thấy thông báo.")
		return
	}
	c.Status(204)
}

func (s *Server) readAllNotifications(c *gin.Context) {
	_ = s.repo.MarkNotificationRead(cur(c).ID, "")
	c.Status(204)
}
