package httpapi

import (
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"masm/backend/internal/ai"
	"masm/backend/internal/store"
)

var conditions = map[string]bool{"Mới": true, "Như mới": true, "Tốt": true, "Khá": true, "Cũ": true}

func (s *Server) categories(c *gin.Context) { c.JSON(200, s.repo.Categories()) }

func (s *Server) listListings(c *gin.Context) {
	out := s.repo.ListListings(store.ListFilter{Query: c.Query("q"), Category: c.Query("cat"), Sort: c.Query("sort")})
	c.JSON(200, gin.H{"items": out, "total": len(out)})
}

func (s *Server) getListing(c *gin.Context) {
	l, err := s.repo.Listing(c.Param("id"))
	if err != nil || (l.Status != "active" && l.Status != "sold") {
		fail(c, 404, "not_found", "Không tìm thấy tin đăng.")
		return
	}
	c.JSON(200, l)
}

type createListingReq struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Price         int64    `json:"price"`
	OriginalPrice int64    `json:"original_price"`
	Condition     string   `json:"condition"`
	Location      string   `json:"location"`
	Category      string   `json:"category"`
	Hashtags      []string `json:"hashtags"`
	AIGenerated   bool     `json:"ai_generated"`
	// Người bán phải xác nhận đã xem lại nội dung (kể cả phần AI gợi ý) trước khi đăng.
	Confirmed bool `json:"confirmed"`
}

func (s *Server) createListing(c *gin.Context) {
	var r createListingReq
	if c.ShouldBindJSON(&r) != nil {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	if !r.Confirmed {
		fail(c, 422, "confirm_required", "Bạn cần xác nhận đã xem lại nội dung trước khi đăng.")
		return
	}
	r.Title, r.Description = strings.TrimSpace(r.Title), strings.TrimSpace(r.Description)
	if n := utf8.RuneCountInString(r.Title); n < 5 || n > 120 {
		fail(c, 422, "invalid_title", "Tiêu đề cần từ 5 đến 120 ký tự.")
		return
	}
	if utf8.RuneCountInString(r.Description) > 2000 {
		fail(c, 422, "invalid_description", "Mô tả tối đa 2000 ký tự.")
		return
	}
	if r.Price <= 0 || r.Price > 1_000_000_000 {
		fail(c, 422, "invalid_price", "Giá cần lớn hơn 0 và nhỏ hơn 1.000.000.000 ₫.")
		return
	}
	if !conditions[r.Condition] {
		fail(c, 422, "invalid_condition", "Tình trạng không hợp lệ.")
		return
	}
	icon := ""
	for _, ct := range s.repo.Categories() {
		if ct.Value == r.Category {
			icon = ct.Icon
		}
	}
	if icon == "" {
		fail(c, 422, "invalid_category", "Danh mục không hợp lệ.")
		return
	}
	if len(r.Hashtags) > 10 {
		r.Hashtags = r.Hashtags[:10]
	}
	u := cur(c)
	l := s.repo.CreateListing(store.Listing{
		Title: r.Title, Description: r.Description, Price: r.Price, OriginalPrice: r.OriginalPrice,
		Condition: r.Condition, Location: r.Location, Category: r.Category, Icon: icon, Hashtags: r.Hashtags,
		SellerID: u.ID, SellerName: u.Name, Status: "pending_review", AIGenerated: r.AIGenerated,
	})
	c.JSON(201, l)
}

/* ---------- AI job: tạo job rồi trả job_id, client poll ---------- */

func (s *Server) createAIJob(c *gin.Context) {
	var in ai.Input
	if c.ShouldBindJSON(&in) != nil || len(in.ImageKeys) == 0 || len(in.ImageKeys) > 10 {
		fail(c, 422, "invalid_images", "Cần từ 1 đến 10 ảnh để AI phân tích.")
		return
	}
	if !ownsKeys(cur(c).ID, in.ImageKeys) {
		fail(c, 403, "invalid_image_key", "Ảnh không thuộc về tài khoản của bạn.")
		return
	}
	j, err := s.runner.Enqueue(cur(c).ID, in)
	if err != nil {
		fail(c, 503, "ai_busy", "Hệ thống AI đang bận. Bạn có thể đăng tin thủ công.")
		return
	}
	c.JSON(202, gin.H{"job_id": j.ID, "status": j.Status})
}

func (s *Server) getAIJob(c *gin.Context) {
	j, err := s.repo.Job(c.Param("id"))
	if err != nil || j.SellerID != cur(c).ID {
		fail(c, 404, "not_found", "Không tìm thấy job.")
		return
	}
	c.JSON(200, j)
}
