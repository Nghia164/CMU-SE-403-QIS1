// Package httpapi: REST API (Gin).
package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"masm/backend/internal/ai"
	"masm/backend/internal/config"
	"masm/backend/internal/notify"
	"masm/backend/internal/payment"
	"masm/backend/internal/storage"
	"masm/backend/internal/store"
)

type Server struct {
	cfg    config.Config
	repo   store.Repo
	runner *ai.Runner
	pay    payment.Gateway
	m      *metrics
	logins  Limiter
	storage storage.Presigner
	mailer  Mailer
	hub     notify.Hub
}

// New nhận limiter (nil = giới hạn trong RAM, chỉ đúng khi chạy một instance).
func New(cfg config.Config, repo store.Repo, runner *ai.Runner, up storage.Presigner, lim Limiter) *Server {
	if lim == nil {
		lim = newMemLimiter(20, time.Minute)
	}
	return &Server{
		storage: up, mailer: LogMailer{},
		cfg: cfg, repo: repo, runner: runner, m: newMetrics(),
		pay: payment.New(payment.Config{
			VNPayTmnCode: cfg.VNPayTmnCode, VNPaySecret: cfg.VNPaySecret, VNPayURL: cfg.VNPayURL,
			MoMoPartnerCode: cfg.MoMoPartner, MoMoAccessKey: cfg.MoMoAccessKey, MoMoSecret: cfg.MoMoSecret, MoMoEndpoint: cfg.MoMoEndpoint,
			ReturnURL: cfg.PublicURL, IPNBaseURL: cfg.APIPublicURL,
		}),
		logins: lim,
	}
}

func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), s.cors(), s.m.middleware())

	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/metrics", s.m.handler)

	v1 := r.Group("/api/v1")
	v1.POST("/auth/register", s.rateLimit(), s.register)
	v1.POST("/auth/login", s.rateLimit(), s.login)
	v1.POST("/auth/refresh", s.rateLimit(), s.refresh)
	v1.POST("/auth/logout", s.logout)
	v1.POST("/auth/forgot-password", s.rateLimit(), s.forgotPassword)
	v1.POST("/auth/reset-password", s.rateLimit(), s.resetPassword)
	v1.GET("/categories", s.categories)
	v1.GET("/listings", s.listListings)
	v1.GET("/listings/:id", s.getListing)
	// IPN do cổng thanh toán gọi, xác thực bằng chữ ký, không dùng JWT.
	v1.GET("/payments/vnpay/ipn", s.vnpayIPN)
	v1.POST("/payments/momo/ipn", s.momoIPN)

	a := v1.Group("", s.auth())
	a.GET("/me", s.me)
	a.POST("/auth/change-password", s.changePassword)
	a.GET("/notifications", s.listNotifications)
	a.GET("/notifications/stream", s.streamNotifications)
	a.POST("/notifications/read-all", s.readAllNotifications)
	a.POST("/notifications/:id/read", s.readNotification)
	a.POST("/reports", s.createReport)
	a.POST("/orders", s.createOrder)
	a.GET("/orders", s.listOrders)
	a.GET("/orders/:id", s.getOrder)
	a.POST("/orders/:id/confirm-received", s.confirmReceived)

	seller := a.Group("", s.role(store.RoleSeller, store.RoleAdmin))
	seller.POST("/listings", s.createListing)
	seller.POST("/uploads/presign", s.presignUpload)
	seller.POST("/ai/jobs", s.createAIJob)
	seller.GET("/ai/jobs/:id", s.getAIJob)
	seller.POST("/orders/:id/ship", s.shipOrder)

	adm := a.Group("/admin", s.role(store.RoleAdmin))
	adm.GET("/stats", s.adminStats)
	adm.GET("/users", s.adminUsers)
	adm.POST("/users/:id/lock", s.adminSetLock(true))
	adm.POST("/users/:id/unlock", s.adminSetLock(false))
	adm.GET("/moderation", s.adminModeration)
	adm.POST("/moderation/:id/approve", s.adminModerate(true))
	adm.POST("/moderation/:id/reject", s.adminModerate(false))
	adm.GET("/reports", s.adminReports)
	adm.POST("/reports/:id/resolve", s.adminResolveReport)
	adm.GET("/orders", s.adminOrders)
	return r
}

type apiErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func fail(c *gin.Context, status int, code, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"error": apiErr{code, msg}})
}

func (s *Server) cors() gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, o := range s.cfg.AllowedOrigins {
		allowed[strings.TrimSpace(o)] = true
	}
	return func(c *gin.Context) {
		if o := c.GetHeader("Origin"); o != "" && allowed[o] {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", o)
			h.Set("Vary", "Origin")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// SetHub bật thông báo realtime (SSE).
func (s *Server) SetHub(h notify.Hub) { s.hub = h }

// SetMailer thay bộ gửi email (mặc định chỉ ghi log).
func (s *Server) SetMailer(m Mailer) { s.mailer = m }
