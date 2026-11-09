package storage

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestR2PresignPut(t *testing.T) {
	r, err := NewR2(R2Config{AccountID: "acc123", AccessKeyID: "AK", SecretAccessKey: "SK", Bucket: "masm"})
	if err != nil {
		t.Fatal(err)
	}
	u, err := r.PresignPut(context.Background(), "uploads/u_1/a.jpg", "image/jpeg", 1234, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"https://acc123.r2.cloudflarestorage.com/masm/uploads/u_1/a.jpg", "X-Amz-Signature=", "X-Amz-Expires=60"} {
		if !strings.Contains(u, want) {
			t.Fatalf("thi?u %q trong %s", want, u)
		}
	}
}
