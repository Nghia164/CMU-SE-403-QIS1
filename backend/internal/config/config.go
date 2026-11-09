// Package config đọc cấu hình từ biến môi trường.
package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string // rỗng = kho RAM (dev)
	JWTSecret      string
	TokenTTL       time.Duration
	AllowedOrigins []string
	AIServiceURL   string // rỗng = dùng AI giả lập trong tiến trình
	AIStepDelay    time.Duration
	VNPayTmnCode   string
	VNPaySecret    string
	VNPayURL       string
	MoMoPartner    string
	MoMoAccessKey  string
	MoMoSecret     string
	MoMoEndpoint   string
	APIPublicURL   string // URL công khai của API (MoMo gọi IPN về đây)
	PublicURL      string // URL công khai của web, dùng làm return URL thanh toán
	AdminEmail     string
	AdminPassword  string
	SeedDemo       bool
	SMTPAddr       string // host:port; rỗng = chỉ ghi email vào log
	SMTPUser       string
	SMTPPass       string
	SMTPFrom       string
	RedisURL       string // rỗng = hàng đợi và rate limit trong RAM
	R2AccountID    string
	R2AccessKeyID  string
	R2SecretKey    string
	R2Bucket       string
}

func Load() Config {
	return Config{
		Port:           get("PORT", "8080"),
		DatabaseURL:    get("DATABASE_URL", ""),
		JWTSecret:      get("JWT_SECRET", "dev-only-change-me-please-32-bytes!!"),
		TokenTTL:       12 * time.Hour,
		AllowedOrigins: strings.Split(get("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001"), ","),
		AIServiceURL:   get("AI_SERVICE_URL", ""),
		AIStepDelay:    dur(get("AI_STEP_DELAY", "900ms")),
		VNPayTmnCode:   get("VNPAY_TMN_CODE", ""),
		VNPaySecret:    get("VNPAY_HASH_SECRET", ""),
		VNPayURL:       get("VNPAY_URL", "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"),
		MoMoPartner:    get("MOMO_PARTNER_CODE", ""),
		MoMoAccessKey:  get("MOMO_ACCESS_KEY", ""),
		MoMoSecret:     get("MOMO_SECRET_KEY", ""),
		MoMoEndpoint:   get("MOMO_ENDPOINT", "https://test-payment.momo.vn/v2/gateway/api/create"),
		APIPublicURL:   get("API_PUBLIC_URL", "http://localhost:8080"),
		PublicURL:      get("PUBLIC_URL", "http://localhost:3000"),
		AdminEmail:     get("ADMIN_EMAIL", "admin@masm.vn"),
		AdminPassword:  get("ADMIN_PASSWORD", "Admin@12345"),
		SeedDemo:       get("SEED_DEMO", "true") == "true",
		SMTPAddr:       get("SMTP_ADDR", ""),
		SMTPUser:       get("SMTP_USER", ""),
		SMTPPass:       get("SMTP_PASS", ""),
		SMTPFrom:       get("SMTP_FROM", "MASM <no-reply@masm.vn>"),
		RedisURL:       get("REDIS_URL", ""),
		R2AccountID:    get("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:  get("R2_ACCESS_KEY_ID", ""),
		R2SecretKey:    get("R2_SECRET_ACCESS_KEY", ""),
		R2Bucket:       get("R2_BUCKET", ""),
	}
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func dur(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 900 * time.Millisecond
	}
	return d
}
