package httpapi

import (
	"strings"
	"testing"
)

func TestBuildMessageBlocksHeaderInjection(t *testing.T) {
	msg := string(buildMessage("MASM <no-reply@masm.vn>", "a@b.vn\r\nBcc: evil@x.vn", "Đặt lại mật khẩu", "nội dung"))
	if strings.Contains(msg, "\r\nBcc:") {
		t.Fatalf("bị chèn header:\n%s", msg)
	}
	if !strings.Contains(msg, "Subject: =?utf-8?q?") || !strings.HasSuffix(msg, "nội dung\r\n") {
		t.Fatalf("thư sai định dạng:\n%s", msg)
	}
}
