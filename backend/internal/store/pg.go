package store

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres cài store.Repo trên Postgres (Supabase). Giao diện Repo không có context/error ở vài hàm,
// nên lỗi CSDL ở các hàm đó được ghi log; hàm trả error thì trả về cho handler.
type Postgres struct{ pool *pgxpool.Pool }

const queryTimeout = 5 * time.Second

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), queryTimeout)
}

// NewPostgres kết nối và chạy các migration chưa áp dụng.
func NewPostgres(c context.Context, url string, migrations fs.FS) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL không hợp lệ: %w", err)
	}
	// Supabase transaction pooler (cổng 6543, Supavisor/PgBouncer) không giữ prepared statement giữa các
	// câu lệnh, nên pgx không được cache prepared statement. Người dùng vẫn ghi đè được bằng
	// ?default_query_exec_mode=... trong URL.
	if cfg.ConnConfig.Port == 6543 && !strings.Contains(url, "default_query_exec_mode") {
		cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	}
	if !strings.Contains(url, "pool_max_conns") {
		cfg.MaxConns = 10 // gói miễn phí của Supabase giới hạn số kết nối; 10 là đủ cho một instance
	}
	pool, err := pgxpool.NewWithConfig(c, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(c); err != nil {
		return nil, fmt.Errorf("không kết nối được CSDL: %w", err)
	}
	p := &Postgres{pool: pool}
	return p, p.migrate(c, migrations)
}

// Reset xoá toàn bộ dữ liệu nghiệp vụ (chỉ dùng cho test).
func (p *Postgres) Reset(c context.Context) error {
	_, err := p.pool.Exec(c, `truncate notifications, audit_logs, payment_transactions, suggestion_history, comparables, ai_jobs, media_assets, reports, orders, listings, refresh_tokens, users cascade`)
	return err
}

func (p *Postgres) Close() { p.pool.Close() }

func (p *Postgres) migrate(c context.Context, fsys fs.FS) error {
	if _, err := p.pool.Exec(c, `create table if not exists schema_migrations (name text primary key, applied_at timestamptz not null default now())`); err != nil {
		return err
	}
	names, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, n := range names {
		sql, err := fs.ReadFile(fsys, n)
		if err != nil {
			return err
		}
		tx, err := p.pool.Begin(c)
		if err != nil {
			return err
		}
		// Khoá theo transaction (dùng được qua transaction pooler): nhiều instance khởi động cùng lúc
		// sẽ chạy migration lần lượt, không giẫm lên nhau.
		if _, err := tx.Exec(c, `select pg_advisory_xact_lock(727274)`); err != nil {
			_ = tx.Rollback(c)
			return err
		}
		var done bool
		if err := tx.QueryRow(c, `select exists(select 1 from schema_migrations where name=$1)`, n).Scan(&done); err != nil {
			_ = tx.Rollback(c)
			return err
		}
		if done {
			_ = tx.Rollback(c)
			continue
		}
		if _, err := tx.Exec(c, string(sql)); err != nil {
			_ = tx.Rollback(c)
			return fmt.Errorf("migration %s: %w", n, err)
		}
		if _, err := tx.Exec(c, `insert into schema_migrations(name) values($1)`, n); err != nil {
			_ = tx.Rollback(c)
			return err
		}
		if err := tx.Commit(c); err != nil {
			return err
		}
		log.Printf("đã áp dụng migration %s", n)
	}
	return nil
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23505" {
		return ErrConflict
	}
	return err
}

func logErr(op string, err error) {
	if err != nil {
		log.Printf("postgres %s: %v", op, err)
	}
}

/* ---------- Users ---------- */

const userCols = `id,name,email,password_hash,role,status,failed_logins,locked_until,created_at,token_version`

func scanUser(row pgx.Row) (User, error) {
	var u User
	var locked *time.Time
	var role string
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &role, &u.Status, &u.FailedLogins, &locked, &u.CreatedAt, &u.TokenVersion)
	u.Role = Role(role)
	if locked != nil {
		u.LockedUntil = *locked
	}
	return u, mapErr(err)
}

func nullTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func (p *Postgres) CreateUser(u User) (User, error) {
	c, cancel := ctx()
	defer cancel()
	u.ID, u.CreatedAt = NewID("u_"), time.Now()
	_, err := p.pool.Exec(c, `insert into users (`+userCols+`) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		u.ID, u.Name, u.Email, u.PasswordHash, string(u.Role), u.Status, u.FailedLogins, nullTime(u.LockedUntil), u.CreatedAt, u.TokenVersion)
	if err != nil {
		return User{}, mapErr(err)
	}
	return u, nil
}

func (p *Postgres) UserByEmail(email string) (User, error) {
	c, cancel := ctx()
	defer cancel()
	return scanUser(p.pool.QueryRow(c, `select `+userCols+` from users where lower(email)=lower($1)`, email))
}

func (p *Postgres) UserByID(id string) (User, error) {
	c, cancel := ctx()
	defer cancel()
	return scanUser(p.pool.QueryRow(c, `select `+userCols+` from users where id=$1`, id))
}

func (p *Postgres) UpdateUser(u User) error {
	c, cancel := ctx()
	defer cancel()
	tag, err := p.pool.Exec(c, `update users set name=$2,status=$3,failed_logins=$4,locked_until=$5,password_hash=$6,token_version=$7 where id=$1`,
		u.ID, u.Name, u.Status, u.FailedLogins, nullTime(u.LockedUntil), u.PasswordHash, u.TokenVersion)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return mapErr(err)
}

func (p *Postgres) ListUsers() []User {
	c, cancel := ctx()
	defer cancel()
	rows, err := p.pool.Query(c, `select `+userCols+`, (select count(*) from listings l where l.seller_id=users.id) from users order by created_at`)
	if err != nil {
		logErr("ListUsers", err)
		return nil
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		var locked *time.Time
		var role string
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &role, &u.Status, &u.FailedLogins, &locked, &u.CreatedAt, &u.TokenVersion, &u.ListingCount); err != nil {
			logErr("ListUsers.scan", err)
			continue
		}
		u.Role = Role(role)
		out = append(out, u)
	}
	return out
}

/* ---------- Listings ---------- */

const listingCols = `id,title,description,price,original_price,condition,location,category,icon,hashtags,seller_id,seller_name,status,ai_verdict,ai_score,ai_generated,created_at`

func scanListing(row pgx.Row) (Listing, error) {
	var l Listing
	err := row.Scan(&l.ID, &l.Title, &l.Description, &l.Price, &l.OriginalPrice, &l.Condition, &l.Location, &l.Category,
		&l.Icon, &l.Hashtags, &l.SellerID, &l.SellerName, &l.Status, &l.AIVerdict, &l.AIScore, &l.AIGenerated, &l.CreatedAt)
	return l, mapErr(err)
}

func (p *Postgres) Categories() []Category {
	c, cancel := ctx()
	defer cancel()
	rows, err := p.pool.Query(c, `select value,label,icon from categories order by (value='khac'), label`)
	if err != nil {
		logErr("Categories", err)
		return nil
	}
	defer rows.Close()
	var out []Category
	for rows.Next() {
		var x Category
		if rows.Scan(&x.Value, &x.Label, &x.Icon) == nil {
			out = append(out, x)
		}
	}
	return out
}

func (p *Postgres) CreateListing(l Listing) Listing {
	c, cancel := ctx()
	defer cancel()
	if l.ID == "" {
		l.ID = NewID("l_")
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now()
	}
	if l.Hashtags == nil {
		l.Hashtags = []string{}
	}
	_, err := p.pool.Exec(c, `insert into listings (`+listingCols+`,search_text) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		l.ID, l.Title, l.Description, l.Price, l.OriginalPrice, l.Condition, l.Location, l.Category, l.Icon, l.Hashtags,
		l.SellerID, l.SellerName, l.Status, l.AIVerdict, l.AIScore, l.AIGenerated, l.CreatedAt, Fold(l.Title+" "+l.Description))
	logErr("CreateListing", err)
	return l
}

func (p *Postgres) Listing(id string) (Listing, error) {
	c, cancel := ctx()
	defer cancel()
	return scanListing(p.pool.QueryRow(c, `select `+listingCols+` from listings where id=$1`, id))
}

func (p *Postgres) ListListings(f ListFilter) []Listing {
	c, cancel := ctx()
	defer cancel()
	status := f.Status
	if status == "" {
		status = "active"
	}
	var where []string
	var args []any
	add := func(cond string, v any) {
		args = append(args, v)
		where = append(where, strings.ReplaceAll(cond, "?", fmt.Sprintf("$%d", len(args))))
	}
	switch status {
	case "all":
	case "active":
		where = append(where, `status in ('active','sold')`)
	default:
		add(`status = ?`, status)
	}
	if f.Category != "" && f.Category != "all" {
		add(`category = ?`, f.Category)
	}
	if q := Fold(strings.TrimSpace(f.Query)); q != "" {
		add(`strpos(search_text, ?) > 0`, q)
	}
	order := `created_at desc`
	switch f.Sort {
	case "price_asc":
		order = `price asc`
	case "price_desc":
		order = `price desc`
	}
	sql := `select ` + listingCols + ` from listings`
	if len(where) > 0 {
		sql += ` where ` + strings.Join(where, ` and `)
	}
	rows, err := p.pool.Query(c, sql+` order by `+order+` limit 500`, args...)
	if err != nil {
		logErr("ListListings", err)
		return nil
	}
	defer rows.Close()
	var out []Listing
	for rows.Next() {
		if l, err := scanListing(rows); err == nil {
			out = append(out, l)
		}
	}
	return out
}

func (p *Postgres) UpdateListing(l Listing) error {
	c, cancel := ctx()
	defer cancel()
	tag, err := p.pool.Exec(c, `update listings set title=$2,description=$3,price=$4,status=$5,ai_verdict=$6,ai_score=$7,search_text=$8 where id=$1`,
		l.ID, l.Title, l.Description, l.Price, l.Status, l.AIVerdict, l.AIScore, Fold(l.Title+" "+l.Description))
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return mapErr(err)
}

/* ---------- Orders ---------- */

const orderCols = `id,listing_id,listing_title,buyer_id,seller_id,amount,method,status,receiver_name,phone,address,events,payment_ref,created_at`

func scanOrder(row pgx.Row) (Order, error) {
	var o Order
	err := row.Scan(&o.ID, &o.ListingID, &o.ListingTitle, &o.BuyerID, &o.SellerID, &o.Amount, &o.Method, &o.Status,
		&o.ReceiverName, &o.Phone, &o.Address, &o.Events, &o.PaymentRef, &o.CreatedAt)
	if o.Events == nil {
		o.Events = map[string]time.Time{}
	}
	return o, mapErr(err)
}

func (p *Postgres) CreateOrder(o Order) Order {
	c, cancel := ctx()
	defer cancel()
	o.ID, o.CreatedAt = NewID("DH-"), time.Now()
	if o.Events == nil {
		o.Events = map[string]time.Time{}
	}
	_, err := p.pool.Exec(c, `insert into orders (`+orderCols+`) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		o.ID, o.ListingID, o.ListingTitle, o.BuyerID, o.SellerID, o.Amount, o.Method, o.Status,
		o.ReceiverName, o.Phone, o.Address, o.Events, o.PaymentRef, o.CreatedAt)
	logErr("CreateOrder", err)
	return o
}

func (p *Postgres) Order(id string) (Order, error) {
	c, cancel := ctx()
	defer cancel()
	return scanOrder(p.pool.QueryRow(c, `select `+orderCols+` from orders where id=$1`, id))
}

func (p *Postgres) UpdateOrder(o Order) error {
	c, cancel := ctx()
	defer cancel()
	tag, err := p.pool.Exec(c, `update orders set status=$2,events=$3,payment_ref=$4 where id=$1`, o.ID, o.Status, o.Events, o.PaymentRef)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return mapErr(err)
}

func (p *Postgres) queryOrders(op, where string, args ...any) []Order {
	c, cancel := ctx()
	defer cancel()
	rows, err := p.pool.Query(c, `select `+orderCols+` from orders `+where+` order by created_at desc limit 500`, args...)
	if err != nil {
		logErr(op, err)
		return nil
	}
	defer rows.Close()
	var out []Order
	for rows.Next() {
		if o, err := scanOrder(rows); err == nil {
			out = append(out, o)
		}
	}
	return out
}

func (p *Postgres) OrdersOfUser(id string) []Order {
	return p.queryOrders("OrdersOfUser", `where buyer_id=$1 or seller_id=$1`, id)
}
func (p *Postgres) AllOrders() []Order { return p.queryOrders("AllOrders", ``) }

/* ---------- Reports ---------- */

const reportSel = `select r.id,r.target_type,r.target_id,r.target_title,r.reason,r.reporter_id,r.status,r.action,r.note,r.created_at,
  (select count(*) from reports x where x.target_type=r.target_type and x.target_id=r.target_id) from reports r `

func scanReport(row pgx.Row) (Report, error) {
	var r Report
	err := row.Scan(&r.ID, &r.TargetType, &r.TargetID, &r.TargetTitle, &r.Reason, &r.ReporterID, &r.Status, &r.Action, &r.Note, &r.CreatedAt, &r.Count)
	return r, mapErr(err)
}

func (p *Postgres) CreateReport(r Report) Report {
	c, cancel := ctx()
	defer cancel()
	r.ID, r.CreatedAt, r.Status = NewID("BC-"), time.Now(), "new"
	_, err := p.pool.Exec(c, `insert into reports (id,reporter_id,target_type,target_id,target_title,reason,status,created_at) values ($1,$2,$3,$4,$5,$6,$7,$8)`,
		r.ID, r.ReporterID, r.TargetType, r.TargetID, r.TargetTitle, r.Reason, r.Status, r.CreatedAt)
	logErr("CreateReport", err)
	if got, err := p.Report(r.ID); err == nil {
		return got
	}
	return r
}

func (p *Postgres) Report(id string) (Report, error) {
	c, cancel := ctx()
	defer cancel()
	return scanReport(p.pool.QueryRow(c, reportSel+`where r.id=$1`, id))
}

func (p *Postgres) UpdateReport(r Report) error {
	c, cancel := ctx()
	defer cancel()
	tag, err := p.pool.Exec(c, `update reports set status=$2,action=$3,note=$4 where id=$1`, r.ID, r.Status, r.Action, r.Note)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return mapErr(err)
}

func (p *Postgres) ListReports() []Report {
	c, cancel := ctx()
	defer cancel()
	rows, err := p.pool.Query(c, reportSel+`order by r.created_at desc limit 500`)
	if err != nil {
		logErr("ListReports", err)
		return nil
	}
	defer rows.Close()
	var out []Report
	for rows.Next() {
		if r, err := scanReport(rows); err == nil {
			out = append(out, r)
		}
	}
	return out
}

/* ---------- Token (refresh, đặt lại mật khẩu) ---------- */

func (p *Postgres) SaveToken(t Token) error {
	c, cancel := ctx()
	defer cancel()
	if t.ID == "" {
		t.ID = NewID("tok_")
	}
	_, err := p.pool.Exec(c, `insert into refresh_tokens (id,user_id,kind,token_hash,expires_at) values ($1,$2,$3,$4,$5)`,
		t.ID, t.UserID, t.Kind, t.Hash, t.ExpiresAt)
	return mapErr(err)
}

func (p *Postgres) ConsumeToken(kind, hash string) (Token, error) {
	c, cancel := ctx()
	defer cancel()
	t := Token{Kind: kind, Hash: hash}
	// Một câu UPDATE nguyên tử: hai request đồng thời chỉ một cái thắng.
	err := p.pool.QueryRow(c, `update refresh_tokens set revoked_at=now()
	  where kind=$1 and token_hash=$2 and revoked_at is null and expires_at>now()
	  returning id,user_id,expires_at`, kind, hash).Scan(&t.ID, &t.UserID, &t.ExpiresAt)
	if err == nil {
		t.RevokedAt = time.Now()
		return t, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Token{}, err
	}
	var revoked *time.Time
	err = p.pool.QueryRow(c, `select id,user_id,expires_at,revoked_at from refresh_tokens where kind=$1 and token_hash=$2 and expires_at>now()`,
		kind, hash).Scan(&t.ID, &t.UserID, &t.ExpiresAt, &revoked)
	if err != nil {
		return Token{}, mapErr(err)
	}
	if revoked != nil {
		t.RevokedAt = *revoked
	}
	return t, ErrConflict
}

func (p *Postgres) RevokeTokens(userID, kind string) {
	c, cancel := ctx()
	defer cancel()
	_, err := p.pool.Exec(c, `update refresh_tokens set revoked_at=now() where user_id=$1 and kind=$2 and revoked_at is null`, userID, kind)
	logErr("RevokeTokens", err)
}

/* ---------- Thông báo ---------- */

func (p *Postgres) AddNotification(n Notification) Notification {
	c, cancel := ctx()
	defer cancel()
	n.ID, n.CreatedAt, n.Read = NewID("n_"), time.Now(), false
	_, err := p.pool.Exec(c, `insert into notifications (id,user_id,kind,title,body,link,created_at) values ($1,$2,$3,$4,$5,$6,$7)`,
		n.ID, n.UserID, n.Kind, n.Title, n.Body, n.Link, n.CreatedAt)
	logErr("AddNotification", err)
	return n
}

func (p *Postgres) NotificationsOf(userID string, limit int) ([]Notification, int) {
	c, cancel := ctx()
	defer cancel()
	var unread int
	logErr("NotificationsOf.count", p.pool.QueryRow(c, `select count(*) from notifications where user_id=$1 and read_at is null`, userID).Scan(&unread))
	rows, err := p.pool.Query(c, `select id,kind,title,body,link,read_at is not null,created_at from notifications where user_id=$1 order by created_at desc, id desc limit $2`, userID, limit)
	if err != nil {
		logErr("NotificationsOf", err)
		return nil, unread
	}
	defer rows.Close()
	var out []Notification
	for rows.Next() {
		var n Notification
		if rows.Scan(&n.ID, &n.Kind, &n.Title, &n.Body, &n.Link, &n.Read, &n.CreatedAt) == nil {
			out = append(out, n)
		}
	}
	return out, unread
}

func (p *Postgres) MarkNotificationRead(userID, id string) error {
	c, cancel := ctx()
	defer cancel()
	if id == "" {
		_, err := p.pool.Exec(c, `update notifications set read_at=now() where user_id=$1 and read_at is null`, userID)
		return err
	}
	tag, err := p.pool.Exec(c, `update notifications set read_at=coalesce(read_at,now()) where user_id=$1 and id=$2`, userID, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

/* ---------- AI jobs, audit, stats ---------- */

func (p *Postgres) SaveJob(j AIJob) {
	c, cancel := ctx()
	defer cancel()
	if j.Steps == nil {
		j.Steps = []AIStep{}
	}
	_, err := p.pool.Exec(c, `insert into ai_jobs (id,seller_id,status,progress,steps,result,error,created_at) values ($1,$2,$3,$4,$5,$6,$7,$8)
	  on conflict (id) do update set status=$3,progress=$4,steps=$5,result=$6,error=$7`,
		j.ID, j.SellerID, j.Status, j.Progress, j.Steps, j.Result, j.Error, j.CreatedAt)
	logErr("SaveJob", err)
}

func (p *Postgres) Job(id string) (AIJob, error) {
	c, cancel := ctx()
	defer cancel()
	var j AIJob
	err := p.pool.QueryRow(c, `select id,seller_id,status,progress,steps,result,error,created_at from ai_jobs where id=$1`, id).
		Scan(&j.ID, &j.SellerID, &j.Status, &j.Progress, &j.Steps, &j.Result, &j.Error, &j.CreatedAt)
	return j, mapErr(err)
}

func (p *Postgres) AddAudit(a AuditLog) {
	c, cancel := ctx()
	defer cancel()
	_, err := p.pool.Exec(c, `insert into audit_logs (id,actor_id,action,target) values ($1,$2,$3,$4)`, NewID("a_"), a.ActorID, a.Action, a.Target)
	logErr("AddAudit", err)
}

func (p *Postgres) Stats() Stats {
	c, cancel := ctx()
	defer cancel()
	var s Stats
	err := p.pool.QueryRow(c, `select
	  (select count(*) from users),
	  (select count(*) from listings where status='pending_review'),
	  (select count(*) from reports where status<>'done'),
	  (select coalesce(sum(amount),0) from orders where status in ('paid','shipping','delivered'))`).
		Scan(&s.Users, &s.PendingListing, &s.OpenReports, &s.Revenue)
	logErr("Stats", err)
	return s
}

var _ Repo = (*Postgres)(nil)
