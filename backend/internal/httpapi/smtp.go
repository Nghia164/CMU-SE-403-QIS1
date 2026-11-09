package httpapi

import (
	"fmt"
	"mime"
	"net/mail"
	"net/smtp"
	"strings"
)

// SMTPMailer gửi email qua SMTP (STARTTLS tự bật nếu máy chủ hỗ trợ, do net/smtp xử lý).
type SMTPMailer struct {
	Addr, User, Pass, From string // Addr dạng host:port
}

func (m SMTPMailer) Send(to, subject, body string) error {
	var auth smtp.Auth
	if m.User != "" {
		auth = smtp.PlainAuth("", m.User, m.Pass, strings.Split(m.Addr, ":")[0])
	}
	// Lệnh MAIL FROM chỉ nhận địa chỉ trần; tên hiển thị chỉ nằm trong header From.
	envelope := m.From
	if a, err := mail.ParseAddress(m.From); err == nil {
		envelope = a.Address
	}
	return smtp.SendMail(m.Addr, auth, envelope, []string{to}, buildMessage(m.From, to, subject, body))
}

// buildMessage dựng thư UTF-8; chặn xuống dòng trong header để không bị chèn header.
func buildMessage(from, to, subject, body string) []byte {
	clean := func(s string) string { return strings.NewReplacer("\r", "", "\n", "").Replace(s) }
	return []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		clean(from), clean(to), mime.QEncoding.Encode("utf-8", clean(subject)), body))
}
