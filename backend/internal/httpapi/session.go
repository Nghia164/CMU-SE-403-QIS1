package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"masm/backend/internal/store"
)

const (
	refreshTTL = 30 * 24 * time.Hour
	resetTTL   = 30 * time.Minute

	kindRefresh = "refresh"
	kindReset   = "reset"
)

// Mailer gửi email. Chưa có nhà cung cấp thì dùng LogMailer (ghi nội dung vào log).
type Mailer interface {
	Send(to, subject, body string) error
}

type LogMailer struct{}

func (LogMailer) Send(to, subject, body string) error {
	log.Printf("[email giả lập] tới=%s | %s | %s", to, subject, body)
	return nil
}

func newSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func hashSecret(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// newToken tạo token ngẫu nhiên, chỉ lưu hash; trả giá trị gốc để gửi cho client.
func (s *Server) newToken(userID, kind string, ttl time.Duration) (string, error) {
	raw := newSecret()
	err := s.repo.SaveToken(store.Token{UserID: userID, Kind: kind, Hash: hashSecret(raw), ExpiresAt: time.Now().Add(ttl)})
	return raw, err
}

// session cấp access token (JWT) và refresh token mới cho người dùng.
func (s *Server) session(u store.User) (authResp, error) {
	tok, err := s.issueToken(u)
	if err != nil {
		return authResp{}, err
	}
	rt, err := s.newToken(u.ID, kindRefresh, refreshTTL)
	if err != nil {
		return authResp{}, err
	}
	return authResp{Token: tok, RefreshToken: rt, User: u}, nil
}

func validPassword(p string) bool { return len(p) >= 8 && len(p) <= 72 }

const weakPasswordMsg = "Mật khẩu cần từ 8 đến 72 ký tự."

// refresh đổi refresh token lấy cặp token mới (xoay vòng). Dùng lại token cũ bị coi là
// dấu hiệu bị đánh cắp: thu hồi mọi refresh token của người dùng đó.
func (s *Server) refresh(c *gin.Context) {
	var r struct {
		RefreshToken string `json:"refresh_token"`
	}
	if c.ShouldBindJSON(&r) != nil || r.RefreshToken == "" {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	const expired = "Phiên đăng nhập đã hết hạn. Bạn vui lòng đăng nhập lại."
	t, err := s.repo.ConsumeToken(kindRefresh, hashSecret(r.RefreshToken))
	if err != nil {
		if t.UserID != "" { // token đã dùng rồi mà bị trình lại
			s.repo.RevokeTokens(t.UserID, kindRefresh)
		}
		fail(c, 401, "invalid_refresh", expired)
		return
	}
	u, err := s.repo.UserByID(t.UserID)
	if err != nil || u.Status == "locked" {
		fail(c, 401, "invalid_refresh", expired)
		return
	}
	resp, err := s.session(u)
	if err != nil {
		fail(c, 500, "internal", "Có lỗi xảy ra, vui lòng thử lại.")
		return
	}
	c.JSON(200, resp)
}

func (s *Server) logout(c *gin.Context) {
	var r struct {
		RefreshToken string `json:"refresh_token"`
	}
	if c.ShouldBindJSON(&r) == nil && r.RefreshToken != "" {
		_, _ = s.repo.ConsumeToken(kindRefresh, hashSecret(r.RefreshToken))
	}
	c.Status(204)
}

func (s *Server) changePassword(c *gin.Context) {
	var r struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	u := cur(c)
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(r.OldPassword)) != nil {
		fail(c, 403, "wrong_password", "Mật khẩu hiện tại chưa đúng.")
		return
	}
	if !validPassword(r.NewPassword) {
		fail(c, 422, "weak_password", weakPasswordMsg)
		return
	}
	if r.NewPassword == r.OldPassword {
		fail(c, 422, "same_password", "Mật khẩu mới cần khác mật khẩu hiện tại.")
		return
	}
	if !s.setPassword(c, &u, r.NewPassword) {
		return
	}
	resp, err := s.session(u) // đăng xuất mọi thiết bị khác, giữ thiết bị này
	if err != nil {
		fail(c, 500, "internal", "Có lỗi xảy ra, vui lòng thử lại.")
		return
	}
	c.JSON(200, resp)
}

func (s *Server) setPassword(c *gin.Context, u *store.User, pw string) bool {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		fail(c, 500, "internal", "Có lỗi xảy ra, vui lòng thử lại.")
		return false
	}
	u.PasswordHash, u.FailedLogins, u.LockedUntil = string(h), 0, time.Time{}
	u.TokenVersion++ // vô hiệu mọi access token đã phát hành
	if err := s.repo.UpdateUser(*u); err != nil {
		fail(c, 500, "internal", "Có lỗi xảy ra, vui lòng thử lại.")
		return false
	}
	s.repo.RevokeTokens(u.ID, kindRefresh)
	s.repo.AddAudit(store.AuditLog{ActorID: u.ID, Action: "password_changed", Target: u.ID})
	return true
}

// forgotPassword luôn trả 202 dù email có tồn tại hay không, để không lộ tài khoản.
func (s *Server) forgotPassword(c *gin.Context) {
	var r struct {
		Email string `json:"email"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	if u, err := s.repo.UserByEmail(r.Email); err == nil && u.Status != "locked" {
		if raw, err := s.newToken(u.ID, kindReset, resetTTL); err == nil {
			link := s.cfg.PublicURL + "/dat-lai-mat-khau?token=" + url.QueryEscape(raw)
			body := "Bạn vừa yêu cầu đặt lại mật khẩu MASM. Mở liên kết sau trong 30 phút: " + link + "\nNếu không phải bạn, hãy bỏ qua email này."
			if err := s.mailer.Send(u.Email, "Đặt lại mật khẩu MASM", body); err != nil {
				log.Printf("gửi email đặt lại mật khẩu: %v", err)
			}
		}
	}
	c.JSON(202, gin.H{"message": "Nếu email đã đăng ký, chúng tôi đã gửi hướng dẫn đặt lại mật khẩu."})
}

func (s *Server) resetPassword(c *gin.Context) {
	var r struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if c.ShouldBindJSON(&r) != nil || r.Token == "" {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	if !validPassword(r.NewPassword) {
		fail(c, 422, "weak_password", weakPasswordMsg)
		return
	}
	t, err := s.repo.ConsumeToken(kindReset, hashSecret(r.Token))
	if err != nil {
		fail(c, 400, "invalid_reset", "Liên kết đặt lại mật khẩu không hợp lệ hoặc đã hết hạn.")
		return
	}
	u, err := s.repo.UserByID(t.UserID)
	if err != nil {
		fail(c, 400, "invalid_reset", "Liên kết đặt lại mật khẩu không hợp lệ hoặc đã hết hạn.")
		return
	}
	if !s.setPassword(c, &u, r.NewPassword) {
		return
	}
	c.Status(204)
}
