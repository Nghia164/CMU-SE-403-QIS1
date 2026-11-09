// Package payment tích hợp VNPay (pay 2.1.0, HMAC-SHA512) và MoMo (v2 captureWallet, HMAC-SHA256).
// Tạo link thanh toán và xác minh chữ ký IPN. Chưa được thử với sandbox thật của hai cổng:
// trước khi chạy thật cần đối chiếu từng trường với tài liệu hiện hành của VNPay và MoMo.
package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var vnTZ = time.FixedZone("ICT", 7*3600)

var ErrNotConfigured = errors.New("cổng thanh toán chưa được cấu hình")

type Config struct {
	VNPayTmnCode string
	VNPaySecret  string
	VNPayURL     string // https://sandbox.vnpayment.vn/paymentv2/vpcpay.html

	MoMoPartnerCode string
	MoMoAccessKey   string
	MoMoSecret      string
	MoMoEndpoint    string // https://test-payment.momo.vn/v2/gateway/api/create

	ReturnURL  string // URL công khai của web; người mua quay về {ReturnURL}/orders/{id}
	IPNBaseURL string // URL công khai của API; MoMo gọi {IPNBaseURL}/api/v1/payments/momo/ipn
}

type Gateway struct {
	Config
	HTTP *http.Client
	Now  func() time.Time
}

func New(c Config) Gateway {
	return Gateway{Config: c, HTTP: &http.Client{Timeout: 15 * time.Second}, Now: time.Now}
}

// Result là kết quả IPN đã được xác minh chữ ký, dùng chung cho hai cổng.
type Result struct {
	OrderID string
	Amount  int64 // đồng
	Success bool
	TxnID   string
}

func equalHex(a, b string) bool {
	x, err1 := hex.DecodeString(strings.ToLower(a))
	y, err2 := hex.DecodeString(strings.ToLower(b))
	return err1 == nil && err2 == nil && hmac.Equal(x, y)
}

/* ---------------- VNPay ---------------- */

// VNPayHash tính vnp_SecureHash: sắp xếp khoá tăng dần, nối key=urlencode(value) bằng &, HMAC-SHA512.
func VNPayHash(secret string, p map[string]string) (hashData, hash string) {
	keys := make([]string, 0, len(p))
	for k, v := range p {
		if v != "" && strings.HasPrefix(k, "vnp_") && k != "vnp_SecureHash" && k != "vnp_SecureHashType" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + url.QueryEscape(p[k])
	}
	hashData = strings.Join(parts, "&")
	m := hmac.New(sha512.New, []byte(secret))
	m.Write([]byte(hashData))
	return hashData, hex.EncodeToString(m.Sum(nil))
}

func (g Gateway) vnpayURL(orderID string, amount int64, clientIP string) (string, error) {
	if g.VNPayTmnCode == "" || g.VNPaySecret == "" {
		return "", ErrNotConfigured
	}
	now := g.Now().In(vnTZ)
	p := map[string]string{
		"vnp_Version": "2.1.0", "vnp_Command": "pay", "vnp_TmnCode": g.VNPayTmnCode,
		"vnp_Amount":     strconv.FormatInt(amount*100, 10), // VNPay tính theo đơn vị x100
		"vnp_CurrCode":   "VND",
		"vnp_TxnRef":     orderID,
		"vnp_OrderInfo":  "Thanh toan don hang " + orderID,
		"vnp_OrderType":  "other",
		"vnp_Locale":     "vn",
		"vnp_ReturnUrl":  g.ReturnURL + "/orders/" + orderID,
		"vnp_IpAddr":     clientIP,
		"vnp_CreateDate": now.Format("20060102150405"),
		"vnp_ExpireDate": now.Add(15 * time.Minute).Format("20060102150405"),
	}
	data, h := VNPayHash(g.VNPaySecret, p)
	return g.VNPayURL + "?" + data + "&vnp_SecureHash=" + h, nil
}

// VerifyVNPay xác minh IPN (GET query) của VNPay.
func (g Gateway) VerifyVNPay(q url.Values) (Result, bool) {
	p := map[string]string{}
	for k := range q {
		p[k] = q.Get(k)
	}
	_, want := VNPayHash(g.VNPaySecret, p)
	if g.VNPaySecret == "" || !equalHex(q.Get("vnp_SecureHash"), want) {
		return Result{}, false
	}
	amt, err := strconv.ParseInt(q.Get("vnp_Amount"), 10, 64)
	if err != nil || amt%100 != 0 {
		return Result{}, false
	}
	return Result{
		OrderID: q.Get("vnp_TxnRef"), Amount: amt / 100, TxnID: q.Get("vnp_TransactionNo"),
		Success: q.Get("vnp_ResponseCode") == "00" && q.Get("vnp_TransactionStatus") == "00",
	}, true
}

/* ---------------- MoMo ---------------- */

type MoMoIPN struct {
	PartnerCode  string `json:"partnerCode"`
	OrderID      string `json:"orderId"`
	RequestID    string `json:"requestId"`
	Amount       int64  `json:"amount"`
	OrderInfo    string `json:"orderInfo"`
	OrderType    string `json:"orderType"`
	TransID      int64  `json:"transId"`
	ResultCode   int    `json:"resultCode"`
	Message      string `json:"message"`
	PayType      string `json:"payType"`
	ResponseTime int64  `json:"responseTime"`
	ExtraData    string `json:"extraData"`
	Signature    string `json:"signature"`
}

func (g Gateway) momoIPNSignature(n MoMoIPN) string {
	raw := fmt.Sprintf("accessKey=%s&amount=%d&extraData=%s&message=%s&orderId=%s&orderInfo=%s&orderType=%s&partnerCode=%s&payType=%s&requestId=%s&responseTime=%d&resultCode=%d&transId=%d",
		g.MoMoAccessKey, n.Amount, n.ExtraData, n.Message, n.OrderID, n.OrderInfo, n.OrderType, n.PartnerCode, n.PayType, n.RequestID, n.ResponseTime, n.ResultCode, n.TransID)
	m := hmac.New(sha256.New, []byte(g.MoMoSecret))
	m.Write([]byte(raw))
	return hex.EncodeToString(m.Sum(nil))
}

// SignMoMoIPN dùng cho test và công cụ giả lập IPN.
func (g Gateway) SignMoMoIPN(n MoMoIPN) string { return g.momoIPNSignature(n) }

func (g Gateway) VerifyMoMo(n MoMoIPN) (Result, bool) {
	if g.MoMoSecret == "" || n.PartnerCode != g.MoMoPartnerCode || !equalHex(n.Signature, g.momoIPNSignature(n)) {
		return Result{}, false
	}
	return Result{OrderID: n.OrderID, Amount: n.Amount, Success: n.ResultCode == 0, TxnID: strconv.FormatInt(n.TransID, 10)}, true
}

func (g Gateway) momoURL(ctx context.Context, orderID string, amount int64) (string, error) {
	if g.MoMoPartnerCode == "" || g.MoMoSecret == "" {
		return "", ErrNotConfigured
	}
	redirect := g.ReturnURL + "/orders/" + orderID
	ipn := g.IPNBaseURL + "/api/v1/payments/momo/ipn"
	info := "Thanh toan don hang " + orderID
	reqID := orderID
	raw := fmt.Sprintf("accessKey=%s&amount=%d&extraData=&ipnUrl=%s&orderId=%s&orderInfo=%s&partnerCode=%s&redirectUrl=%s&requestId=%s&requestType=captureWallet",
		g.MoMoAccessKey, amount, ipn, orderID, info, g.MoMoPartnerCode, redirect, reqID)
	m := hmac.New(sha256.New, []byte(g.MoMoSecret))
	m.Write([]byte(raw))
	body, _ := json.Marshal(map[string]any{
		"partnerCode": g.MoMoPartnerCode, "accessKey": g.MoMoAccessKey, "requestId": reqID, "amount": amount,
		"orderId": orderID, "orderInfo": info, "redirectUrl": redirect, "ipnUrl": ipn,
		"requestType": "captureWallet", "extraData": "", "lang": "vi", "signature": hex.EncodeToString(m.Sum(nil)),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.MoMoEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := g.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	var out struct {
		ResultCode int    `json:"resultCode"`
		Message    string `json:"message"`
		PayURL     string `json:"payUrl"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", fmt.Errorf("momo: phản hồi không hợp lệ (HTTP %d)", res.StatusCode)
	}
	if out.ResultCode != 0 || out.PayURL == "" {
		return "", fmt.Errorf("momo: %d %s", out.ResultCode, out.Message)
	}
	return out.PayURL, nil
}

// PayURL trả về địa chỉ chuyển người mua sang cổng thanh toán.
func (g Gateway) PayURL(ctx context.Context, provider, orderID string, amount int64, clientIP string) (string, error) {
	switch provider {
	case "vnpay":
		return g.vnpayURL(orderID, amount, clientIP)
	case "momo":
		return g.momoURL(ctx, orderID, amount)
	}
	return "", errors.New("cổng thanh toán không hợp lệ")
}
