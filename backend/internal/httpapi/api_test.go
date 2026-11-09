package httpapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"masm/backend/internal/ai"
	"masm/backend/internal/config"
	"masm/backend/internal/notify"
	"masm/backend/internal/payment"
	"masm/backend/internal/storage"
	"masm/backend/internal/store"
	"masm/backend/migrations"
)

type env struct {
	t    *testing.T
	h    http.Handler
	cfg  config.Config
	repo store.Repo
	srv  *Server
}

func newEnv(t *testing.T) *env {
	cfg := config.Load()
	cfg.AIStepDelay = time.Millisecond
	momo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"resultCode":0,"message":"ok","payUrl":"https://pay.momo.test/abc"}`))
	}))
	t.Cleanup(momo.Close)
	cfg.VNPayTmnCode, cfg.VNPaySecret = "TMNTEST", "vnpay-test-secret"
	cfg.MoMoPartner, cfg.MoMoAccessKey, cfg.MoMoSecret, cfg.MoMoEndpoint = "MOMOTEST", "ak", "momo-test-secret", momo.URL
	var repo store.Repo = store.NewMemory()
	// Có TEST_DATABASE_URL thì chạy cả bộ test trên Postgres thật (xoá sạch dữ liệu trước mỗi test).
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		pg, err := store.NewPostgres(context.Background(), url, migrations.FS)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(pg.Close)
		if err := pg.Reset(context.Background()); err != nil {
			t.Fatal(err)
		}
		repo = pg
	}
	store.Seed(repo, cfg.AdminEmail, cfg.AdminPassword, true)
	ctx, stop := context.WithCancel(context.Background())
	t.Cleanup(stop)
	hub := notify.NewMemHub()
	repo = notify.Repo{Repo: repo, Hub: hub}
	srv := New(cfg, repo, ai.NewRunner(ctx, repo, ai.Stub{}, ai.NewMemQueue(), cfg.AIStepDelay, 2), storage.Fake{BaseURL: "http://fake"}, nil)
	srv.SetHub(hub)
	return &env{t, srv.Router(), cfg, repo, srv}
}

func (e *env) do(method, path, token string, body any) (int, map[string]any) {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	e.h.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func (e *env) login(email, pw string) string {
	code, r := e.do("POST", "/api/v1/auth/login", "", map[string]string{"email": email, "password": pw})
	if code != 200 {
		e.t.Fatalf("login %s: %d %v", email, code, r)
	}
	return r["token"].(string)
}

func TestListingsSearchAccentInsensitive(t *testing.T) {
	e := newEnv(t)
	_, r := e.do("GET", "/api/v1/listings?q=may+anh", "", nil)
	if r["total"].(float64) != 1 {
		t.Fatalf("want 1 got %v", r["total"])
	}
	// tin chờ duyệt không hiện công khai
	_, r = e.do("GET", "/api/v1/listings?q=hang+hieu", "", nil)
	if r["total"].(float64) != 0 {
		t.Fatalf("pending listing leaked: %v", r)
	}
}

func TestAuthRBACAndLockout(t *testing.T) {
	e := newEnv(t)
	if c, _ := e.do("GET", "/api/v1/me", "", nil); c != 401 {
		t.Fatalf("me without token: %d", c)
	}
	buyer := e.login("buyer@masm.vn", "Demo@12345")
	if c, _ := e.do("GET", "/api/v1/admin/stats", buyer, nil); c != 403 {
		t.Fatalf("buyer on admin: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/ai/jobs", buyer, map[string]any{"image_keys": []string{"a"}}); c != 403 {
		t.Fatalf("buyer creating AI job: %d", c)
	}
	// không cho tự đăng ký admin
	_, r := e.do("POST", "/api/v1/auth/register", "", map[string]string{"name": "X", "email": "x@x.vn", "password": "12345678", "role": "admin"})
	if r["user"].(map[string]any)["role"] != "buyer" {
		t.Fatalf("self-registered admin: %v", r)
	}
	for i := 0; i < 5; i++ {
		e.do("POST", "/api/v1/auth/login", "", map[string]string{"email": "buyer@masm.vn", "password": "sai"})
	}
	if c, _ := e.do("POST", "/api/v1/auth/login", "", map[string]string{"email": "buyer@masm.vn", "password": "Demo@12345"}); c != 423 {
		t.Fatalf("want 423 lockout, got %d", c)
	}
}

func TestSellerPublishNeedsConfirmAndModeration(t *testing.T) {
	e := newEnv(t)
	seller := e.login("seller@masm.vn", "Demo@12345")
	body := map[string]any{"title": "Xe đạp thể thao", "description": "Còn tốt", "price": 1000000, "condition": "Tốt", "category": "xe", "ai_generated": true}
	if c, r := e.do("POST", "/api/v1/listings", seller, body); c != 422 || r["error"].(map[string]any)["code"] != "confirm_required" {
		t.Fatalf("publish without confirm: %d %v", c, r)
	}
	body["confirmed"] = true
	c, l := e.do("POST", "/api/v1/listings", seller, body)
	if c != 201 || l["status"] != "pending_review" {
		t.Fatalf("publish: %d %v", c, l)
	}
	admin := e.login(e.cfg.AdminEmail, e.cfg.AdminPassword)
	if c, _ := e.do("POST", "/api/v1/admin/moderation/"+l["id"].(string)+"/approve", admin, nil); c != 200 {
		t.Fatalf("approve: %d", c)
	}
	if c, _ := e.do("GET", "/api/v1/listings/"+l["id"].(string), "", nil); c != 200 {
		t.Fatalf("approved listing not public: %d", c)
	}
}

func TestAIJobAsync(t *testing.T) {
	e := newEnv(t)
	seller := e.login("seller@masm.vn", "Demo@12345")
	c, r := e.do("POST", "/api/v1/ai/jobs", seller, map[string]any{"image_keys": []string{"uploads/" + e.userID("seller@masm.vn") + "/k1.jpg"}, "hint": "Xe đạp"})
	if c != 202 || r["job_id"] == nil {
		t.Fatalf("create job: %d %v", c, r)
	}
	id := r["job_id"].(string)
	buyer := e.login("buyer@masm.vn", "Demo@12345")
	if c, _ := e.do("GET", "/api/v1/ai/jobs/"+id, buyer, nil); c != 403 {
		t.Fatalf("buyer reading job: %d", c)
	}
	for i := 0; i < 200; i++ {
		_, j := e.do("GET", "/api/v1/ai/jobs/"+id, seller, nil)
		if j["status"] == "SUCCEEDED" {
			if j["result"] == nil {
				t.Fatal("no result")
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("job did not finish")
}

func (e *env) momoIPN(oid string, amount int64, code int) payment.MoMoIPN {
	g := payment.New(payment.Config{MoMoPartnerCode: e.cfg.MoMoPartner, MoMoAccessKey: e.cfg.MoMoAccessKey, MoMoSecret: e.cfg.MoMoSecret})
	n := payment.MoMoIPN{PartnerCode: e.cfg.MoMoPartner, OrderID: oid, RequestID: oid, Amount: amount, OrderInfo: "x", OrderType: "momo_wallet",
		TransID: 4001, ResultCode: code, Message: "ok", PayType: "qr", ResponseTime: 1700000000000}
	n.Signature = g.SignMoMoIPN(n)
	return n
}

func (e *env) vnpayIPN(oid string, amount100 string, code string, secret string) (int, map[string]any) {
	p := map[string]string{"vnp_TxnRef": oid, "vnp_Amount": amount100, "vnp_ResponseCode": code, "vnp_TransactionStatus": code, "vnp_TransactionNo": "9001", "vnp_TmnCode": e.cfg.VNPayTmnCode}
	_, h := payment.VNPayHash(secret, p)
	q := url.Values{}
	for k, v := range p {
		q.Set(k, v)
	}
	q.Set("vnp_SecureHash", h)
	w := httptest.NewRecorder()
	e.h.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/payments/vnpay/ipn?"+q.Encode(), nil))
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func (e *env) firstListing(q string) string {
	_, ls := e.do("GET", "/api/v1/listings?q="+q, "", nil)
	return ls["items"].([]any)[0].(map[string]any)["id"].(string)
}

func TestOrderPaymentFlowMoMo(t *testing.T) {
	e := newEnv(t)
	buyer := e.login("buyer@masm.vn", "Demo@12345")
	seller := e.login("seller@masm.vn", "Demo@12345")
	lid := e.firstListing("asama")

	req := map[string]any{"listing_id": lid, "method": "momo", "receiver_name": "Trâm", "phone": "0912345678", "address": "Q.3"}
	if c, _ := e.do("POST", "/api/v1/orders", seller, req); c != 422 {
		t.Fatalf("buying own listing: %d", c)
	}
	req["phone"] = "123"
	if c, _ := e.do("POST", "/api/v1/orders", buyer, req); c != 422 {
		t.Fatalf("bad phone: %d", c)
	}
	req["phone"] = "0912345678"
	c, r := e.do("POST", "/api/v1/orders", buyer, req)
	if c != 201 || r["pay_url"] != "https://pay.momo.test/abc" {
		t.Fatalf("create order: %d %v", c, r)
	}
	oid := r["order"].(map[string]any)["id"].(string)

	bad := e.momoIPN(oid, 1450000, 0)
	bad.Signature = "00"
	if c, _ := e.do("POST", "/api/v1/payments/momo/ipn", "", bad); c != 401 {
		t.Fatalf("bad signature: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/payments/momo/ipn", "", e.momoIPN(oid, 1, 0)); c != 422 {
		t.Fatalf("amount mismatch: %d", c)
	}
	ok := e.momoIPN(oid, 1450000, 0)
	if c, _ := e.do("POST", "/api/v1/payments/momo/ipn", "", ok); c != 204 {
		t.Fatalf("ipn: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/payments/momo/ipn", "", ok); c != 204 { // gọi lại
		t.Fatalf("ipn replay: %d", c)
	}
	if _, o := e.do("GET", "/api/v1/orders/"+oid, buyer, nil); o["status"] != "paid" {
		t.Fatalf("order sau IPN: %v", o)
	}
	if c, o := e.do("POST", "/api/v1/orders/"+oid+"/ship", seller, nil); c != 200 || o["status"] != "shipping" {
		t.Fatalf("ship: %d %v", c, o)
	}
	if c, o := e.do("POST", "/api/v1/orders/"+oid+"/confirm-received", buyer, nil); c != 200 || o["status"] != "delivered" {
		t.Fatalf("confirm: %d %v", c, o)
	}
	if c, _ := e.do("POST", "/api/v1/orders", buyer, req); c != 409 {
		t.Fatalf("buy sold listing: %d", c)
	}
	e.do("POST", "/api/v1/auth/register", "", map[string]string{"name": "Lạ", "email": "la@x.vn", "password": "12345678"})
	other := e.login("la@x.vn", "12345678")
	if c, _ := e.do("GET", "/api/v1/orders/"+oid, other, nil); c != 404 {
		t.Fatalf("stranger reads order: %d", c)
	}
}

func TestOrderPaymentFlowVNPay(t *testing.T) {
	e := newEnv(t)
	buyer := e.login("buyer@masm.vn", "Demo@12345")
	req := map[string]any{"listing_id": e.firstListing("asama"), "method": "vnpay", "receiver_name": "Trâm", "phone": "0912345678", "address": "Q.3"}
	c, r := e.do("POST", "/api/v1/orders", buyer, req)
	if c != 201 {
		t.Fatalf("create order: %d %v", c, r)
	}
	pu := r["pay_url"].(string)
	if !strings.HasPrefix(pu, e.cfg.VNPayURL+"?") || !strings.Contains(pu, "vnp_Amount=145000000") || !strings.Contains(pu, "vnp_SecureHash=") {
		t.Fatalf("pay_url: %s", pu)
	}
	oid := r["order"].(map[string]any)["id"].(string)

	if _, out := e.vnpayIPN(oid, "145000000", "00", "sai-khoa"); out["RspCode"] != "97" {
		t.Fatalf("chữ ký sai: %v", out)
	}
	if _, out := e.vnpayIPN("DH-khong-co", "145000000", "00", e.cfg.VNPaySecret); out["RspCode"] != "01" {
		t.Fatalf("không có đơn: %v", out)
	}
	if _, out := e.vnpayIPN(oid, "100", "00", e.cfg.VNPaySecret); out["RspCode"] != "04" {
		t.Fatalf("sai tiền: %v", out)
	}
	if _, out := e.vnpayIPN(oid, "145000000", "00", e.cfg.VNPaySecret); out["RspCode"] != "00" {
		t.Fatalf("thành công: %v", out)
	}
	if _, out := e.vnpayIPN(oid, "145000000", "00", e.cfg.VNPaySecret); out["RspCode"] != "02" {
		t.Fatalf("gọi lại: %v", out)
	}
	if _, o := e.do("GET", "/api/v1/orders/"+oid, buyer, nil); o["status"] != "paid" {
		t.Fatalf("order sau IPN: %v", o)
	}
}

func TestPaymentFailureCancelsOrder(t *testing.T) {
	e := newEnv(t)
	buyer := e.login("buyer@masm.vn", "Demo@12345")
	// IPN thất bại từ cổng: đơn bị huỷ, tin vẫn bán được
	req := map[string]any{"listing_id": e.firstListing("asama"), "method": "momo", "receiver_name": "Trâm", "phone": "0912345678", "address": "Q.3"}
	_, r := e.do("POST", "/api/v1/orders", buyer, req)
	oid := r["order"].(map[string]any)["id"].(string)
	e.do("POST", "/api/v1/payments/momo/ipn", "", e.momoIPN(oid, 1450000, 1006))
	if _, o := e.do("GET", "/api/v1/orders/"+oid, buyer, nil); o["status"] != "cancelled" {
		t.Fatalf("order: %v", o)
	}
	// cổng chưa cấu hình: 503 và đơn bị huỷ
	e.srv.pay = payment.New(payment.Config{})
	if c, _ := e.do("POST", "/api/v1/orders", buyer, req); c != 503 {
		t.Fatalf("chưa cấu hình cổng: %d", c)
	}
}
func TestReportAndAdminResolve(t *testing.T) {
	e := newEnv(t)
	buyer := e.login("buyer@masm.vn", "Demo@12345")
	admin := e.login(e.cfg.AdminEmail, e.cfg.AdminPassword)
	_, ls := e.do("GET", "/api/v1/listings?q=guitar", "", nil)
	lid := ls["items"].([]any)[0].(map[string]any)["id"].(string)
	c, rep := e.do("POST", "/api/v1/reports", buyer, map[string]any{"target_type": "listing", "target_id": lid, "reason": "Mô tả sai tình trạng"})
	if c != 201 {
		t.Fatalf("report: %d %v", c, rep)
	}
	if c, _ := e.do("POST", "/api/v1/admin/reports/"+rep["id"].(string)+"/resolve", admin, map[string]string{"action": "remove"}); c != 200 {
		t.Fatalf("resolve: %d", c)
	}
	if c, _ := e.do("GET", "/api/v1/listings/"+lid, "", nil); c != 404 {
		t.Fatalf("removed listing still public: %d", c)
	}
}

func TestLockedUserBlocked(t *testing.T) {
	e := newEnv(t)
	admin := e.login(e.cfg.AdminEmail, e.cfg.AdminPassword)
	buyer := e.login("buyer@masm.vn", "Demo@12345")
	_, u := e.do("GET", "/api/v1/me", buyer, nil)
	e.do("POST", "/api/v1/admin/users/"+u["id"].(string)+"/lock", admin, nil)
	if c, _ := e.do("GET", "/api/v1/me", buyer, nil); c != 403 {
		t.Fatalf("locked user still allowed: %d", c)
	}
}

func (e *env) userID(email string) string {
	u, err := e.repo.UserByEmail(email)
	if err != nil {
		e.t.Fatal(err)
	}
	return u.ID
}

func TestPresignUploadAndKeyOwnership(t *testing.T) {
	e := newEnv(t)
	seller := e.login("seller@masm.vn", "Demo@12345")
	buyer := e.login("buyer@masm.vn", "Demo@12345")

	if c, _ := e.do("POST", "/api/v1/uploads/presign", buyer, map[string]any{"content_type": "image/png", "size": 1000}); c != 403 {
		t.Fatalf("buyer presign: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/uploads/presign", seller, map[string]any{"content_type": "application/pdf", "size": 1000}); c != 422 {
		t.Fatalf("bad type: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/uploads/presign", seller, map[string]any{"content_type": "image/png", "size": 6 << 20}); c != 422 {
		t.Fatalf("too big: %d", c)
	}
	c, r := e.do("POST", "/api/v1/uploads/presign", seller, map[string]any{"content_type": "image/jpeg", "size": 2000})
	if c != 200 || r["upload_url"] == nil {
		t.Fatalf("presign: %d %v", c, r)
	}
	key := r["key"].(string)
	if want := "uploads/" + e.userID("seller@masm.vn") + "/"; len(key) < len(want) || key[:len(want)] != want {
		t.Fatalf("key không nằm trong thư mục người dùng: %s", key)
	}
	if c, _ := e.do("POST", "/api/v1/ai/jobs", seller, map[string]any{"image_keys": []string{key}}); c != 202 {
		t.Fatalf("job with own key: %d", c)
	}
	other := "uploads/" + e.userID("buyer@masm.vn") + "/x.jpg"
	if c, _ := e.do("POST", "/api/v1/ai/jobs", seller, map[string]any{"image_keys": []string{other}}); c != 403 {
		t.Fatalf("job with foreign key: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/ai/jobs", seller, map[string]any{"image_keys": []string{"uploads/" + e.userID("seller@masm.vn") + "/../x.jpg"}}); c != 403 {
		t.Fatalf("path traversal: %d", c)
	}
}

type captureMailer struct{ last string }

func (m *captureMailer) Send(_, _, body string) error { m.last = body; return nil }

func tokenFrom(body string) string {
	i := strings.Index(body, "token=")
	rest := body[i+len("token="):]
	if j := strings.IndexAny(rest, " \n"); j >= 0 {
		rest = rest[:j]
	}
	v, _ := url.QueryUnescape(rest)
	return v
}

func TestRefreshRotationAndReuseDetection(t *testing.T) {
	e := newEnv(t)
	c, r := e.do("POST", "/api/v1/auth/login", "", map[string]string{"email": "buyer@masm.vn", "password": "Demo@12345"})
	if c != 200 || r["refresh_token"] == nil {
		t.Fatalf("login: %d %v", c, r)
	}
	rt1 := r["refresh_token"].(string)
	c, r = e.do("POST", "/api/v1/auth/refresh", "", map[string]string{"refresh_token": rt1})
	if c != 200 || r["token"] == nil {
		t.Fatalf("refresh: %d %v", c, r)
	}
	rt2 := r["refresh_token"].(string)
	if rt2 == rt1 {
		t.Fatal("refresh token phải được xoay vòng")
	}
	// trình lại token cũ: bị từ chối và thu hồi luôn token mới
	if c, _ := e.do("POST", "/api/v1/auth/refresh", "", map[string]string{"refresh_token": rt1}); c != 401 {
		t.Fatalf("reuse: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/auth/refresh", "", map[string]string{"refresh_token": rt2}); c != 401 {
		t.Fatalf("token mới phải bị thu hồi sau khi phát hiện dùng lại: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/auth/refresh", "", map[string]string{"refresh_token": "bogus"}); c != 401 {
		t.Fatalf("bogus: %d", c)
	}
}

func TestLogoutRevokesRefresh(t *testing.T) {
	e := newEnv(t)
	_, r := e.do("POST", "/api/v1/auth/login", "", map[string]string{"email": "buyer@masm.vn", "password": "Demo@12345"})
	rt := r["refresh_token"].(string)
	if c, _ := e.do("POST", "/api/v1/auth/logout", "", map[string]string{"refresh_token": rt}); c != 204 {
		t.Fatalf("logout: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/auth/refresh", "", map[string]string{"refresh_token": rt}); c != 401 {
		t.Fatalf("sau logout: %d", c)
	}
}

func TestChangePassword(t *testing.T) {
	e := newEnv(t)
	_, r := e.do("POST", "/api/v1/auth/login", "", map[string]string{"email": "buyer@masm.vn", "password": "Demo@12345"})
	access, oldRT := r["token"].(string), r["refresh_token"].(string)
	body := func(o, n string) map[string]string { return map[string]string{"old_password": o, "new_password": n} }
	if c, _ := e.do("POST", "/api/v1/auth/change-password", access, body("sai-mat-khau", "Moi@123456")); c != 403 {
		t.Fatalf("wrong old: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/auth/change-password", access, body("Demo@12345", "ngan")); c != 422 {
		t.Fatalf("weak: %d", c)
	}
	c, r := e.do("POST", "/api/v1/auth/change-password", access, body("Demo@12345", "Moi@123456"))
	if c != 200 || r["refresh_token"] == nil {
		t.Fatalf("change: %d %v", c, r)
	}
	if c, _ := e.do("POST", "/api/v1/auth/refresh", "", map[string]string{"refresh_token": oldRT}); c != 401 {
		t.Fatalf("refresh cũ phải bị thu hồi: %d", c)
	}
	if c, _ := e.do("GET", "/api/v1/me", access, nil); c != 401 {
		t.Fatalf("access token cũ phải bị từ chối ngay: %d", c)
	}
	if c, _ := e.do("GET", "/api/v1/me", r["token"].(string), nil); c != 200 {
		t.Fatalf("access token mới phải dùng được: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/auth/login", "", map[string]string{"email": "buyer@masm.vn", "password": "Demo@12345"}); c != 401 {
		t.Fatalf("mật khẩu cũ: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/auth/login", "", map[string]string{"email": "buyer@masm.vn", "password": "Moi@123456"}); c != 200 {
		t.Fatalf("mật khẩu mới: %d", c)
	}
}

func TestForgotAndResetPassword(t *testing.T) {
	e := newEnv(t)
	mail := &captureMailer{}
	e.srv.SetMailer(mail)

	// email không tồn tại: vẫn 202, không gửi gì
	if c, _ := e.do("POST", "/api/v1/auth/forgot-password", "", map[string]string{"email": "khong-co@masm.vn"}); c != 202 || mail.last != "" {
		t.Fatalf("unknown email: %d %q", c, mail.last)
	}
	if c, _ := e.do("POST", "/api/v1/auth/forgot-password", "", map[string]string{"email": "buyer@masm.vn"}); c != 202 || mail.last == "" {
		t.Fatalf("forgot: %d %q", c, mail.last)
	}
	tok := tokenFrom(mail.last)
	if c, _ := e.do("POST", "/api/v1/auth/reset-password", "", map[string]string{"token": tok, "new_password": "ngan"}); c != 422 {
		t.Fatalf("weak reset: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/auth/reset-password", "", map[string]string{"token": "sai", "new_password": "Moi@123456"}); c != 400 {
		t.Fatalf("bad token: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/auth/reset-password", "", map[string]string{"token": tok, "new_password": "Moi@123456"}); c != 204 {
		t.Fatalf("reset: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/auth/reset-password", "", map[string]string{"token": tok, "new_password": "Khac@123456"}); c != 400 {
		t.Fatalf("token chỉ dùng một lần: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/auth/login", "", map[string]string{"email": "buyer@masm.vn", "password": "Moi@123456"}); c != 200 {
		t.Fatalf("login sau reset: %d", c)
	}
}

func TestNotifications(t *testing.T) {
	e := newEnv(t)
	buyer := e.login("buyer@masm.vn", "Demo@12345")
	seller := e.login("seller@masm.vn", "Demo@12345")
	admin := e.login(e.cfg.AdminEmail, e.cfg.AdminPassword)

	if c, r := e.do("GET", "/api/v1/notifications", seller, nil); c != 200 || r["unread"].(float64) != 0 {
		t.Fatalf("ban đầu: %d %v", c, r)
	}
	// admin duyệt tin đang chờ -> người bán nhận thông báo
	_, q := e.do("GET", "/api/v1/admin/moderation", admin, nil)
	id := q["items"].([]any)[0].(map[string]any)["id"].(string)
	if c, _ := e.do("POST", "/api/v1/admin/moderation/"+id+"/approve", admin, nil); c != 200 {
		t.Fatalf("approve: %d", c)
	}
	// mua hàng và thanh toán -> cả hai bên nhận thông báo
	req := map[string]any{"listing_id": e.firstListing("asama"), "method": "momo", "receiver_name": "Trâm", "phone": "0912345678", "address": "Q.3"}
	_, r := e.do("POST", "/api/v1/orders", buyer, req)
	oid := r["order"].(map[string]any)["id"].(string)
	e.do("POST", "/api/v1/payments/momo/ipn", "", e.momoIPN(oid, 1450000, 0))

	_, r = e.do("GET", "/api/v1/notifications", seller, nil)
	if r["unread"].(float64) != 2 { // duyệt tin + có đơn mới
		t.Fatalf("seller: %v", r)
	}
	items := r["items"].([]any)
	if items[0].(map[string]any)["kind"] != "order_paid" { // mới nhất trước
		t.Fatalf("thứ tự: %v", items)
	}
	nid := items[0].(map[string]any)["id"].(string)
	if c, _ := e.do("POST", "/api/v1/notifications/"+nid+"/read", buyer, nil); c != 404 {
		t.Fatalf("đọc thông báo của người khác: %d", c)
	}
	if c, _ := e.do("POST", "/api/v1/notifications/"+nid+"/read", seller, nil); c != 204 {
		t.Fatalf("read: %d", c)
	}
	if _, r = e.do("GET", "/api/v1/notifications", seller, nil); r["unread"].(float64) != 1 {
		t.Fatalf("sau read: %v", r)
	}
	e.do("POST", "/api/v1/notifications/read-all", seller, nil)
	if _, r = e.do("GET", "/api/v1/notifications", seller, nil); r["unread"].(float64) != 0 {
		t.Fatalf("sau read-all: %v", r)
	}
	if _, r = e.do("GET", "/api/v1/notifications", buyer, nil); r["unread"].(float64) != 1 {
		t.Fatalf("buyer: %v", r)
	}
	if c, _ := e.do("GET", "/api/v1/notifications", "", nil); c != 401 {
		t.Fatalf("không đăng nhập: %d", c)
	}
}

func TestNotificationStreamSSE(t *testing.T) {
	e := newEnv(t)
	buyer := e.login("buyer@masm.vn", "Demo@12345")
	ts := httptest.NewServer(e.h)
	defer ts.Close()

	if resp, err := http.Get(ts.URL + "/api/v1/notifications/stream"); err != nil || resp.StatusCode != 401 {
		t.Fatalf("không đăng nhập: %v %v", resp, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", ts.URL+"/api/v1/notifications/stream", nil)
	req.Header.Set("Authorization", "Bearer "+buyer)
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("mở luồng: %v %v", resp, err)
	}
	defer resp.Body.Close()
	rd := bufio.NewReader(resp.Body)
	if line, _ := rd.ReadString('\n'); !strings.HasPrefix(line, ": connected") {
		t.Fatalf("dòng đầu: %q", line)
	}
	rd.ReadString('\n')

	// tạo thông báo cho buyer -> luồng nhận được ngay
	e.repo.AddNotification(store.Notification{UserID: e.userID("buyer@masm.vn"), Kind: "order_paid", Title: "Thanh toán thành công"})
	var event, data string
	for data == "" {
		line, err := rd.ReadString('\n')
		if err != nil {
			t.Fatalf("đọc luồng: %v", err)
		}
		switch {
		case strings.HasPrefix(line, "event: "):
			event = strings.TrimSpace(strings.TrimPrefix(line, "event: "))
		case strings.HasPrefix(line, "data: "):
			data = strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		}
	}
	if event != "notification" || !strings.Contains(data, "Thanh toán thành công") {
		t.Fatalf("event=%q data=%q", event, data)
	}
}
