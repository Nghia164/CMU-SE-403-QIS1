package httpapi

import (
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"masm/backend/internal/store"
)

const (
	maxFailedLogins = 5
	lockDuration    = 15 * time.Minute
)

type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // buyer | seller
}

type authResp struct {
	Token        string     `json:"token"`
	RefreshToken string     `json:"refresh_token"`
	User         store.User `json:"user"`
}

func (s *Server) register(c *gin.Context) {
	var r registerReq
	if c.ShouldBindJSON(&r) != nil {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	r.Name, r.Email = strings.TrimSpace(r.Name), strings.TrimSpace(r.Email)
	if r.Name == "" || utf8.RuneCountInString(r.Name) > 60 {
		fail(c, 422, "invalid_name", "Nhập tên hiển thị (tối đa 60 ký tự).")
		return
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		fail(c, 422, "invalid_email", "Email chưa đúng định dạng, ví dụ ban@email.com.")
		return
	}
	if !validPassword(r.Password) {
		fail(c, 422, "weak_password", weakPasswordMsg)
		return
	}
	role := store.RoleBuyer
	if r.Role == "seller" { // không cho tự đăng ký admin
		role = store.RoleSeller
	}
	h, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(c, 500, "internal", "Có lỗi xảy ra, vui lòng thử lại.")
		return
	}
	u, err := s.repo.CreateUser(store.User{Name: r.Name, Email: r.Email, PasswordHash: string(h), Role: role, Status: "active"})
	if errors.Is(err, store.ErrConflict) {
		fail(c, 409, "email_taken", "Email này đã được đăng ký.")
		return
	}
	resp, err := s.session(u)
	if err != nil {
		fail(c, 500, "internal", "Có lỗi xảy ra, vui lòng thử lại.")
		return
	}
	c.JSON(201, resp)
}

func (s *Server) login(c *gin.Context) {
	var r struct{ Email, Password string }
	if c.ShouldBindJSON(&r) != nil {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	const generic = "Email hoặc mật khẩu chưa đúng."
	u, err := s.repo.UserByEmail(strings.TrimSpace(r.Email))
	if err != nil {
		// vẫn so khớp giả để thời gian phản hồi không lộ email có tồn tại hay không
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$7EqJtq98hPqEX7fNZaFWoOa5Z5Zf5p5uQk1m4Zk0y6Qp6v2m9p5Wy"), []byte(r.Password))
		fail(c, 401, "bad_credentials", generic)
		return
	}
	if time.Now().Before(u.LockedUntil) {
		fail(c, 423, "login_locked", "Tài khoản tạm khoá do đăng nhập sai nhiều lần. Thử lại sau 15 phút.")
		return
	}
	if u.Status == "locked" {
		fail(c, 403, "account_locked", "Tài khoản của bạn đã bị khoá.")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(r.Password)) != nil {
		u.FailedLogins++
		if u.FailedLogins >= maxFailedLogins {
			u.LockedUntil, u.FailedLogins = time.Now().Add(lockDuration), 0
		}
		_ = s.repo.UpdateUser(u)
		fail(c, 401, "bad_credentials", generic)
		return
	}
	u.FailedLogins, u.LockedUntil = 0, time.Time{}
	_ = s.repo.UpdateUser(u)
	resp, err := s.session(u)
	if err != nil {
		fail(c, 500, "internal", "Có lỗi xảy ra, vui lòng thử lại.")
		return
	}
	c.JSON(200, resp)
}

func (s *Server) me(c *gin.Context) { c.JSON(200, cur(c)) }
