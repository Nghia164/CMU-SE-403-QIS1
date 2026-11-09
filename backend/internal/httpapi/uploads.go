package httpapi

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"masm/backend/internal/store"
)

const (
	maxImageBytes = 5 << 20
	presignTTL    = 10 * time.Minute
)

var imageExt = map[string]string{"image/jpeg": "jpg", "image/png": "png", "image/webp": "webp"}

// uploadPrefix là thư mục riêng của từng người dùng; key ảnh phải nằm trong đó.
func uploadPrefix(userID string) string { return "uploads/" + userID + "/" }

type presignReq struct {
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// presignUpload cấp URL PUT để client tải ảnh thẳng lên R2, rồi gửi `key` vào /ai/jobs.
func (s *Server) presignUpload(c *gin.Context) {
	var r presignReq
	if c.ShouldBindJSON(&r) != nil {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	ext, ok := imageExt[r.ContentType]
	if !ok {
		fail(c, 422, "invalid_type", "Chỉ nhận ảnh JPG, PNG hoặc WebP.")
		return
	}
	if r.Size <= 0 || r.Size > maxImageBytes {
		fail(c, 422, "invalid_size", "Mỗi ảnh tối đa 5 MB.")
		return
	}
	key := fmt.Sprintf("%s%s.%s", uploadPrefix(cur(c).ID), store.NewID("img_"), ext)
	url, err := s.storage.PresignPut(c.Request.Context(), key, r.ContentType, r.Size, presignTTL)
	if err != nil {
		fail(c, 503, "storage_unavailable", "Chưa thể tải ảnh lên lúc này. Bạn thử lại sau.")
		return
	}
	c.JSON(200, gin.H{
		"key": key, "upload_url": url, "method": "PUT",
		"headers":    gin.H{"Content-Type": r.ContentType},
		"expires_in": int(presignTTL.Seconds()),
	})
}

// ownsKeys kiểm tra mọi key ảnh thuộc về người dùng, tránh dùng ảnh của người khác.
func ownsKeys(userID string, keys []string) bool {
	p := uploadPrefix(userID)
	for _, k := range keys {
		if !strings.HasPrefix(k, p) || strings.Contains(k, "..") {
			return false
		}
	}
	return true
}
