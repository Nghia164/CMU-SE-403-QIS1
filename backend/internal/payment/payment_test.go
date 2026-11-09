package payment

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestVNPayHashSortsAndEncodes(t *testing.T) {
	data, h := VNPayHash("s3cret", map[string]string{
		"vnp_OrderInfo": "Thanh toan don hang", "vnp_Amount": "100", "vnp_Empty": "", "other": "x", "vnp_SecureHash": "zzz",
	})
	if data != "vnp_Amount=100&vnp_OrderInfo=Thanh+toan+don+hang" {
		t.Fatalf("hashData = %q", data)
	}
	if len(h) != 128 { // SHA-512 d?ng hex
		t.Fatalf("hash d?i %d", len(h))
	}
}

func TestVNPayURLRoundTrip(t *testing.T) {
	g := New(Config{VNPayTmnCode: "T", VNPaySecret: "k", VNPayURL: "https://x/pay", ReturnURL: "https://web"})
	g.Now = func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }
	u, err := g.vnpayURL("DH-1", 1450000, "1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "vnp_CreateDate=20260102100405") { // ??i sang GMT+7
		t.Fatalf("gi? VN sai: %s", u)
	}
	parsed, _ := url.Parse(u)
	res, ok := g.VerifyVNPay(parsed.Query()) // URL do ta k? ph?i t? x?c minh ???c
	if !ok || res.OrderID != "DH-1" || res.Amount != 1450000 {
		t.Fatalf("verify = %+v %v", res, ok)
	}
	q := parsed.Query()
	q.Set("vnp_Amount", "1")
	if _, ok := g.VerifyVNPay(q); ok {
		t.Fatal("??i s? ti?n ph?i l?m h?ng ch? k?")
	}
}

func TestMoMoVerify(t *testing.T) {
	g := New(Config{MoMoPartnerCode: "P", MoMoAccessKey: "a", MoMoSecret: "s"})
	n := MoMoIPN{PartnerCode: "P", OrderID: "DH-1", Amount: 5000, TransID: 77, ResultCode: 0}
	n.Signature = g.SignMoMoIPN(n)
	if res, ok := g.VerifyMoMo(n); !ok || !res.Success || res.TxnID != "77" {
		t.Fatalf("verify = %+v %v", res, ok)
	}
	n.Amount = 1
	if _, ok := g.VerifyMoMo(n); ok {
		t.Fatal("s?a d? li?u ph?i l?m h?ng ch? k?")
	}
}
