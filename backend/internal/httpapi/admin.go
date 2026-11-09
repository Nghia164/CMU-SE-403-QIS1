package httpapi

import (
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"masm/backend/internal/store"
)

func (s *Server) audit(c *gin.Context, action, target string) {
	s.repo.AddAudit(store.AuditLog{ActorID: cur(c).ID, Action: action, Target: target})
}

/* ---------- Báo cáo vi phạm (mọi người dùng đã đăng nhập) ---------- */

func (s *Server) createReport(c *gin.Context) {
	var r struct {
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
		Reason     string `json:"reason"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	r.Reason = strings.TrimSpace(r.Reason)
	if n := utf8.RuneCountInString(r.Reason); n < 5 || n > 500 {
		fail(c, 422, "invalid_reason", "Lý do cần từ 5 đến 500 ký tự.")
		return
	}
	title := ""
	switch r.TargetType {
	case "listing":
		l, err := s.repo.Listing(r.TargetID)
		if err != nil {
			fail(c, 404, "not_found", "Không tìm thấy đối tượng bị báo cáo.")
			return
		}
		title = l.Title
	case "user":
		u, err := s.repo.UserByID(r.TargetID)
		if err != nil {
			fail(c, 404, "not_found", "Không tìm thấy đối tượng bị báo cáo.")
			return
		}
		title = u.Name
	default:
		fail(c, 422, "invalid_target", "Loại đối tượng không hợp lệ.")
		return
	}
	rep := s.repo.CreateReport(store.Report{TargetType: r.TargetType, TargetID: r.TargetID, TargetTitle: title, Reason: r.Reason, ReporterID: cur(c).ID})
	c.JSON(201, rep)
}

/* ---------- Admin ---------- */

func (s *Server) adminStats(c *gin.Context) { c.JSON(200, s.repo.Stats()) }
func (s *Server) adminUsers(c *gin.Context) { c.JSON(200, gin.H{"items": s.repo.ListUsers()}) }
func (s *Server) adminOrders(c *gin.Context) { c.JSON(200, gin.H{"items": s.repo.AllOrders()}) }
func (s *Server) adminReports(c *gin.Context) { c.JSON(200, gin.H{"items": s.repo.ListReports()}) }

func (s *Server) adminModeration(c *gin.Context) {
	c.JSON(200, gin.H{"items": s.repo.ListListings(store.ListFilter{Status: "pending_review", Sort: "newest"})})
}

func (s *Server) adminSetLock(lock bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, err := s.repo.UserByID(c.Param("id"))
		if err != nil {
			fail(c, 404, "not_found", "Không tìm thấy tài khoản.")
			return
		}
		if u.Role == store.RoleAdmin {
			fail(c, 422, "cannot_lock_admin", "Không thể khoá tài khoản quản trị.")
			return
		}
		u.Status = map[bool]string{true: "locked", false: "active"}[lock]
		_ = s.repo.UpdateUser(u)
		s.audit(c, "user."+u.Status, u.ID)
		c.JSON(200, u)
	}
}

func (s *Server) adminModerate(approve bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		l, err := s.repo.Listing(c.Param("id"))
		if err != nil || l.Status != "pending_review" {
			fail(c, 404, "not_found", "Tin không có trong hàng đợi duyệt.")
			return
		}
		l.Status = map[bool]string{true: "active", false: "rejected"}[approve]
		_ = s.repo.UpdateListing(l)
		s.audit(c, "listing."+l.Status, l.ID)
		if approve {
			s.notify(l.SellerID, "listing_approved", "Tin của bạn đã được duyệt", l.Title, "/listing/"+l.ID)
		} else {
			s.notify(l.SellerID, "listing_rejected", "Tin của bạn chưa được duyệt", l.Title, "/listing/"+l.ID)
		}
		c.JSON(200, l)
	}
}

func (s *Server) adminResolveReport(c *gin.Context) {
	var r struct{ Action, Note string }
	if c.ShouldBindJSON(&r) != nil {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	if r.Action != "remove" && r.Action != "warn" && r.Action != "lock" && r.Action != "ignore" {
		fail(c, 422, "invalid_action", "Hình thức xử lý không hợp lệ.")
		return
	}
	rep, err := s.repo.Report(c.Param("id"))
	if err != nil || rep.Status == "done" {
		fail(c, 404, "not_found", "Không tìm thấy báo cáo đang mở.")
		return
	}
	// Tìm người bán liên quan để khoá khi cần.
	sellerID := rep.TargetID
	if rep.TargetType == "listing" {
		if l, err := s.repo.Listing(rep.TargetID); err == nil {
			sellerID = l.SellerID
			if r.Action == "remove" {
				l.Status = "hidden"
				_ = s.repo.UpdateListing(l)
			}
		}
	}
	if r.Action == "lock" {
		if u, err := s.repo.UserByID(sellerID); err == nil && u.Role != store.RoleAdmin {
			u.Status = "locked"
			_ = s.repo.UpdateUser(u)
		}
	}
	rep.Status, rep.Action, rep.Note = "done", r.Action, r.Note
	_ = s.repo.UpdateReport(rep)
	s.audit(c, "report."+r.Action, rep.ID)
	c.JSON(200, rep)
}
