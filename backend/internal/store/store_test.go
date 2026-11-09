package store_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"masm/backend/internal/store"
	"masm/backend/migrations"
)

// Bộ kiểm thử hợp đồng: cùng một bộ test chạy cho kho RAM và Postgres (khi có TEST_DATABASE_URL),
// để hai bản cài đặt store.Repo hành xử giống nhau.
func TestMemoryContract(t *testing.T) { contract(t, store.NewMemory()) }

func TestPostgresContract(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("đặt TEST_DATABASE_URL để chạy trên Postgres")
	}
	pg, err := store.NewPostgres(context.Background(), url, migrations.FS)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pg.Close)
	if err := pg.Reset(context.Background()); err != nil {
		t.Fatal(err)
	}
	contract(t, pg)
}

func contract(t *testing.T, r store.Repo) {
	t.Helper()
	seller, err := r.CreateUser(store.User{Name: "Người bán", Email: "Seller@X.vn", PasswordHash: "h", Role: store.RoleSeller, Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	buyer, _ := r.CreateUser(store.User{Name: "Người mua", Email: "buyer@x.vn", PasswordHash: "h", Role: store.RoleBuyer, Status: "active"})

	t.Run("users", func(t *testing.T) {
		if _, err := r.CreateUser(store.User{Name: "Trùng", Email: "SELLER@x.vn", PasswordHash: "h", Role: store.RoleBuyer, Status: "active"}); !errors.Is(err, store.ErrConflict) {
			t.Fatalf("email trùng (không phân biệt hoa thường) phải là ErrConflict, được %v", err)
		}
		got, err := r.UserByEmail("seller@x.vn")
		if err != nil || got.ID != seller.ID {
			t.Fatalf("UserByEmail: %v %v", got, err)
		}
		if _, err := r.UserByID("khong-co"); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("UserByID: %v", err)
		}
		got.PasswordHash, got.FailedLogins, got.LockedUntil = "h2", 3, time.Now().Add(time.Hour)
		if err := r.UpdateUser(got); err != nil {
			t.Fatal(err)
		}
		again, _ := r.UserByID(seller.ID)
		if again.PasswordHash != "h2" || again.FailedLogins != 3 || again.LockedUntil.IsZero() {
			t.Fatalf("UpdateUser không lưu đủ: %+v", again)
		}
		if err := r.UpdateUser(store.User{ID: "khong-co"}); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("UpdateUser không có: %v", err)
		}
	})

	mk := func(title string, price int64, cat, status string, ago time.Duration) store.Listing {
		return r.CreateListing(store.Listing{Title: title, Description: "Mô tả " + title, Price: price, Condition: "Tốt", Location: "HN",
			Category: cat, Icon: "package", SellerID: seller.ID, SellerName: seller.Name, Status: status, CreatedAt: time.Now().Add(-ago)})
	}
	a := mk("Xe đạp Asama", 1450000, "xe", "active", 3*time.Hour)
	b := mk("Máy ảnh Fuji", 8900000, "dien-tu", "active", 2*time.Hour)
	c := mk("Ghế Bắc Âu", 650000, "nha", "pending_review", time.Hour)
	d := mk("Nồi chiên", 890000, "nha", "sold", 5*time.Hour)

	t.Run("listings", func(t *testing.T) {
		if n := len(r.Categories()); n != 6 {
			t.Fatalf("categories = %d", n)
		}
		ids := func(f store.ListFilter) []string {
			var o []string
			for _, l := range r.ListListings(f) {
				o = append(o, l.ID)
			}
			return o
		}
		if got := ids(store.ListFilter{}); len(got) != 3 || got[0] != b.ID { // active + sold, mới nhất trước
			t.Fatalf("mặc định: %v", got)
		}
		if got := ids(store.ListFilter{Query: "xe dap"}); len(got) != 1 || got[0] != a.ID { // tìm không dấu
			t.Fatalf("tìm không dấu: %v", got)
		}
		if got := ids(store.ListFilter{Category: "dien-tu"}); len(got) != 1 || got[0] != b.ID {
			t.Fatalf("danh mục: %v", got)
		}
		if got := ids(store.ListFilter{Sort: "price_asc"}); got[0] != d.ID {
			t.Fatalf("giá tăng: %v", got)
		}
		if got := ids(store.ListFilter{Sort: "price_desc"}); got[0] != b.ID {
			t.Fatalf("giá giảm: %v", got)
		}
		if got := ids(store.ListFilter{Status: "pending_review"}); len(got) != 1 || got[0] != c.ID {
			t.Fatalf("chờ duyệt: %v", got)
		}
		if got := ids(store.ListFilter{Status: "all"}); len(got) != 4 {
			t.Fatalf("all: %v", got)
		}
		c.Status = "active"
		if err := r.UpdateListing(c); err != nil {
			t.Fatal(err)
		}
		if got, _ := r.Listing(c.ID); got.Status != "active" {
			t.Fatalf("UpdateListing: %+v", got)
		}
		if _, err := r.Listing("khong-co"); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("Listing: %v", err)
		}
	})

	t.Run("orders", func(t *testing.T) {
		mkOrder := func(l store.Listing) store.Order {
			return r.CreateOrder(store.Order{ListingID: l.ID, ListingTitle: l.Title, BuyerID: buyer.ID, SellerID: seller.ID, Amount: l.Price,
				Method: "momo", Status: "pending_payment", ReceiverName: "T", Phone: "0912345678", Address: "x", Events: map[string]time.Time{"placed": time.Now()}})
		}
		o1, o2 := mkOrder(a), mkOrder(a)
		o1.Status = "paid"
		o1.PaymentRef = "T1"
		o1.Events["paid"] = time.Now()
		if err := r.UpdateOrder(o1); err != nil {
			t.Fatal(err)
		}
		got, err := r.Order(o1.ID)
		if err != nil || got.Status != "paid" || got.PaymentRef != "T1" || got.Events["paid"].IsZero() {
			t.Fatalf("Order: %+v %v", got, err)
		}
		// hai đơn cùng thanh toán một tin: lần thứ hai phải bị từ chối (chống bán trùng)
		o2.Status = "paid"
		if err := r.UpdateOrder(o2); err == nil {
			t.Fatal("đơn paid thứ hai của cùng một tin phải bị từ chối")
		}
		if n := len(r.OrdersOfUser(buyer.ID)); n != 2 {
			t.Fatalf("OrdersOfUser(buyer) = %d", n)
		}
		if n := len(r.OrdersOfUser(seller.ID)); n != 2 {
			t.Fatalf("OrdersOfUser(seller) = %d", n)
		}
		if n := len(r.AllOrders()); n != 2 {
			t.Fatalf("AllOrders = %d", n)
		}
	})

	t.Run("tokens", func(t *testing.T) {
		exp := time.Now().Add(time.Hour)
		if err := r.SaveToken(store.Token{UserID: buyer.ID, Kind: "refresh", Hash: "h1", ExpiresAt: exp}); err != nil {
			t.Fatal(err)
		}
		if tk, err := r.ConsumeToken("refresh", "h1"); err != nil || tk.UserID != buyer.ID {
			t.Fatalf("consume lần 1: %+v %v", tk, err)
		}
		if tk, err := r.ConsumeToken("refresh", "h1"); !errors.Is(err, store.ErrConflict) || tk.UserID != buyer.ID {
			t.Fatalf("dùng lại phải là ErrConflict kèm user: %+v %v", tk, err)
		}
		if _, err := r.ConsumeToken("reset", "h1"); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("sai loại token: %v", err)
		}
		_ = r.SaveToken(store.Token{UserID: buyer.ID, Kind: "refresh", Hash: "old", ExpiresAt: time.Now().Add(-time.Minute)})
		if _, err := r.ConsumeToken("refresh", "old"); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("hết hạn: %v", err)
		}
		_ = r.SaveToken(store.Token{UserID: buyer.ID, Kind: "refresh", Hash: "h2", ExpiresAt: exp})
		_ = r.SaveToken(store.Token{UserID: buyer.ID, Kind: "reset", Hash: "h3", ExpiresAt: exp})
		r.RevokeTokens(buyer.ID, "refresh")
		if _, err := r.ConsumeToken("refresh", "h2"); !errors.Is(err, store.ErrConflict) {
			t.Fatalf("sau RevokeTokens: %v", err)
		}
		if _, err := r.ConsumeToken("reset", "h3"); err != nil { // loại khác không bị ảnh hưởng
			t.Fatalf("reset token: %v", err)
		}
		// hai lần consume đồng thời: đúng một bên thắng
		_ = r.SaveToken(store.Token{UserID: buyer.ID, Kind: "refresh", Hash: "race", ExpiresAt: exp})
		res := make(chan error, 8)
		for i := 0; i < 8; i++ {
			go func() { _, err := r.ConsumeToken("refresh", "race"); res <- err }()
		}
		wins := 0
		for i := 0; i < 8; i++ {
			if <-res == nil {
				wins++
			}
		}
		if wins != 1 {
			t.Fatalf("có %d bên consume thành công, phải đúng 1", wins)
		}
	})

	t.Run("notifications", func(t *testing.T) {
		for _, title := range []string{"một", "hai", "ba"} {
			r.AddNotification(store.Notification{UserID: buyer.ID, Kind: "order_paid", Title: title})
			time.Sleep(2 * time.Millisecond)
		}
		items, unread := r.NotificationsOf(buyer.ID, 2)
		if len(items) != 2 || unread != 3 || items[0].Title != "ba" {
			t.Fatalf("NotificationsOf: %+v unread=%d", items, unread)
		}
		if err := r.MarkNotificationRead(seller.ID, items[0].ID); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("đọc của người khác: %v", err)
		}
		if err := r.MarkNotificationRead(buyer.ID, items[0].ID); err != nil {
			t.Fatal(err)
		}
		if _, unread := r.NotificationsOf(buyer.ID, 10); unread != 2 {
			t.Fatalf("unread sau đọc 1 = %d", unread)
		}
		_ = r.MarkNotificationRead(buyer.ID, "")
		if _, unread := r.NotificationsOf(buyer.ID, 10); unread != 0 {
			t.Fatalf("unread sau đọc hết = %d", unread)
		}
	})

	t.Run("jobs", func(t *testing.T) {
		j := store.AIJob{ID: "job_1", SellerID: seller.ID, Status: "QUEUED", CreatedAt: time.Now(), Steps: []store.AIStep{{Label: "A", Status: "pending"}}}
		r.SaveJob(j)
		j.Status, j.Progress = "SUCCEEDED", 100
		j.Result = &store.AIResult{Titles: []string{"T"}, Optimal: 1450000, Sources: []store.PriceSource{{Source: "chotot", Price: 1}}}
		r.SaveJob(j) // upsert
		got, err := r.Job("job_1")
		if err != nil || got.Status != "SUCCEEDED" || got.Result == nil || got.Result.Optimal != 1450000 || len(got.Steps) != 1 {
			t.Fatalf("Job: %+v %v", got, err)
		}
		if _, err := r.Job("khong-co"); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("Job không có: %v", err)
		}
	})

	t.Run("reports-audit-stats", func(t *testing.T) {
		r1 := r.CreateReport(store.Report{ReporterID: buyer.ID, TargetType: "listing", TargetID: a.ID, TargetTitle: a.Title, Reason: "Nghi giả"})
		r2 := r.CreateReport(store.Report{ReporterID: seller.ID, TargetType: "listing", TargetID: a.ID, TargetTitle: a.Title, Reason: "Sai giá"})
		if r2.Count != 2 {
			t.Fatalf("Count gộp theo đối tượng = %d", r2.Count)
		}
		r1.Status, r1.Action = "done", "ignore"
		if err := r.UpdateReport(r1); err != nil {
			t.Fatal(err)
		}
		if got, _ := r.Report(r1.ID); got.Status != "done" || got.Action != "ignore" {
			t.Fatalf("UpdateReport: %+v", got)
		}
		if n := len(r.ListReports()); n != 2 {
			t.Fatalf("ListReports = %d", n)
		}
		r.AddAudit(store.AuditLog{ActorID: seller.ID, Action: "x", Target: "y"})
		s := r.Stats()
		if s.Users != 2 || s.OpenReports != 1 || s.Revenue != 1450000 || s.PendingListing != 0 {
			t.Fatalf("Stats = %+v", s)
		}
	})
}

// Trên Supabase, anon key đọc được mọi bảng public trừ khi bật RLS. Mô phỏng role anon của Supabase.
func TestPostgresRLSBlocksAnon(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("đặt TEST_DATABASE_URL để chạy trên Postgres")
	}
	ctx := context.Background()
	pg, err := store.NewPostgres(ctx, url, migrations.FS)
	if err != nil {
		t.Fatal(err)
	}
	defer pg.Close()
	if err := pg.Reset(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := pg.CreateUser(store.User{Name: "A", Email: "rls@x.vn", PasswordHash: "h", Role: store.RoleBuyer, Status: "active"}); err != nil {
		t.Fatal(err)
	}
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	for _, q := range []string{
		`do $$ begin if not exists (select from pg_roles where rolname='anon') then create role anon nologin; end if; end $$`,
		`grant usage on schema public to anon`,
		`grant select, insert, update, delete on all tables in schema public to anon`, // Supabase cấp sẵn quyền này
		`set role anon`,
	} {
		if _, err := conn.Exec(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	var n int
	if err := conn.QueryRow(ctx, `select count(*) from users`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("anon thấy %d dòng users (err=%v), RLS phải chặn", n, err)
	}
	if _, err := conn.Exec(ctx, `insert into users (id,name,email,password_hash,role) values ('u_x','X','x@x.vn','h','admin')`); err == nil {
		t.Fatal("anon ghi được vào users")
	}
	conn.Exec(ctx, `reset role`)
	var off []string
	rows, _ := conn.Query(ctx, `select tablename from pg_tables where schemaname='public' and not rowsecurity`)
	for rows.Next() {
		var s string
		rows.Scan(&s)
		off = append(off, s)
	}
	if len(off) > 0 {
		t.Fatalf("bảng chưa bật RLS: %v", off)
	}
}
