package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"masm/backend/internal/ai"
	"masm/backend/internal/config"
	"masm/backend/internal/httpapi"
	"masm/backend/internal/notify"
	"masm/backend/internal/storage"
	"masm/backend/internal/store"
	"masm/backend/migrations"
)

func main() {
	cfg := config.Load()
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	if os.Getenv("JWT_SECRET") == "" {
		log.Println("CẢNH BÁO: đang dùng JWT_SECRET mặc định, chỉ dành cho dev.")
	}

	var repo store.Repo
	if cfg.DatabaseURL != "" {
		pg, err := store.NewPostgres(context.Background(), cfg.DatabaseURL, migrations.FS)
		if err != nil {
			log.Fatalf("Postgres: %v", err)
		}
		defer pg.Close()
		repo = pg
		log.Println("dùng Postgres")
	} else {
		repo = store.NewMemory()
		log.Println("CẢNH BÁO: dùng kho RAM, dữ liệu mất khi khởi động lại (đặt DATABASE_URL để dùng Postgres).")
	}
	store.Seed(repo, cfg.AdminEmail, cfg.AdminPassword, cfg.SeedDemo)

	var client ai.Client = ai.Stub{}
	if cfg.AIServiceURL != "" {
		client = ai.NewHTTPClient(cfg.AIServiceURL)
	}
	var up storage.Presigner = storage.Fake{BaseURL: "http://localhost:9000/fake-r2"}
	if cfg.R2AccountID != "" && cfg.R2Bucket != "" {
		r2, err := storage.NewR2(storage.R2Config{AccountID: cfg.R2AccountID, AccessKeyID: cfg.R2AccessKeyID, SecretAccessKey: cfg.R2SecretKey, Bucket: cfg.R2Bucket})
		if err != nil {
			log.Fatalf("R2: %v", err)
		}
		up = r2
	} else {
		log.Println("CẢNH BÁO: chưa cấu hình R2, upload ảnh dùng URL giả.")
	}
	appCtx, stopApp := context.WithCancel(context.Background())
	defer stopApp()
	var hub notify.Hub = notify.NewMemHub()
	var queue ai.Queue = ai.NewMemQueue()
	var lim httpapi.Limiter // nil = trong RAM
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			log.Fatalf("REDIS_URL: %v", err)
		}
		rdb := redis.NewClient(opt)
		if err := rdb.Ping(appCtx).Err(); err != nil {
			log.Fatalf("Redis: %v", err)
		}
		defer rdb.Close()
		queue, lim, hub = ai.NewRedisQueue(rdb), httpapi.NewRedisLimiter(rdb, 20, time.Minute), notify.NewRedisHub(rdb)
		log.Println("dùng Redis cho hàng đợi AI và rate limit")
	} else {
		log.Println("CẢNH BÁO: chưa đặt REDIS_URL, hàng đợi AI và rate limit chạy trong RAM (chỉ đúng với một instance).")
	}
	repo = notify.Repo{Repo: repo, Hub: hub} // thông báo mới được phát realtime
	runner := ai.NewRunner(appCtx, repo, client, queue, cfg.AIStepDelay, 4)

	api := httpapi.New(cfg, repo, runner, up, lim)
	api.SetHub(hub)
	if cfg.SMTPAddr != "" {
		api.SetMailer(httpapi.SMTPMailer{Addr: cfg.SMTPAddr, User: cfg.SMTPUser, Pass: cfg.SMTPPass, From: cfg.SMTPFrom})
	} else {
		log.Println("CẢNH BÁO: chưa đặt SMTP_ADDR, email (đặt lại mật khẩu) chỉ được ghi vào log.")
	}
	srv := &http.Server{
		Addr: ":" + cfg.Port, Handler: api.Router(),
		ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second,
	}
	go func() {
		log.Printf("MASM backend lắng nghe :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
