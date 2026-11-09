// Package store chứa mô hình dữ liệu và kho lưu trữ (Repo).
package store

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleSeller Role = "seller"
	RoleBuyer  Role = "buyer"
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	Status       string    `json:"status"` // active | locked | pending
	CreatedAt    time.Time `json:"created_at"`
	FailedLogins int       `json:"-"`
	LockedUntil  time.Time `json:"-"`
	// Tăng khi đổi mật khẩu; access token mang phiên bản khác bị từ chối.
	TokenVersion int `json:"-"`
	ListingCount int       `json:"listing_count"`
}

type Category struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
}

// Trạng thái tin: pending_review -> active | rejected; active -> sold | hidden.
type Listing struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Price         int64     `json:"price"`
	OriginalPrice int64     `json:"original_price,omitempty"`
	Condition     string    `json:"condition"`
	Location      string    `json:"location"`
	Category      string    `json:"category"`
	Icon          string    `json:"icon"`
	Hashtags      []string  `json:"hashtags,omitempty"`
	SellerID      string    `json:"seller_id"`
	SellerName    string    `json:"seller_name"`
	Status        string    `json:"status"`
	AIVerdict     string    `json:"ai_verdict,omitempty"`
	AIScore       int       `json:"ai_score,omitempty"`
	AIGenerated   bool      `json:"ai_generated"`
	CreatedAt     time.Time `json:"created_at"`
}

type Order struct {
	ID           string               `json:"id"`
	ListingID    string               `json:"listing_id"`
	ListingTitle string               `json:"listing_title"`
	BuyerID      string               `json:"buyer_id"`
	SellerID     string               `json:"seller_id"`
	Amount       int64                `json:"amount"`
	Method       string               `json:"method"` // vnpay | momo
	Status       string               `json:"status"` // pending_payment | paid | shipping | delivered | cancelled | refunded
	ReceiverName string               `json:"receiver_name"`
	Phone        string               `json:"phone"`
	Address      string               `json:"address"`
	Events       map[string]time.Time `json:"events"`
	PaymentRef   string               `json:"payment_ref,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
}

type Report struct {
	ID          string    `json:"id"`
	TargetType  string    `json:"target_type"` // listing | user
	TargetID    string    `json:"target_id"`
	TargetTitle string    `json:"target_title"`
	Reason      string    `json:"reason"`
	ReporterID  string    `json:"reporter_id"`
	Status      string    `json:"status"` // new | progress | done
	Action      string    `json:"action,omitempty"`
	Note        string    `json:"note,omitempty"`
	Count       int       `json:"count"`
	CreatedAt   time.Time `json:"created_at"`
}

type AIStep struct {
	Label  string `json:"label"`
	Status string `json:"status"` // pending | running | done | error
	Detail string `json:"detail,omitempty"`
}

type PriceSource struct {
	Source    string `json:"source"`
	Title     string `json:"title"`
	Price     int64  `json:"price"`
	Condition string `json:"condition"`
}

type AIResult struct {
	Titles       []string      `json:"titles"`
	Descriptions []string      `json:"descriptions"`
	Hashtags     []string      `json:"hashtags"`
	Quick        int64         `json:"price_quick"`
	Optimal      int64         `json:"price_optimal"`
	Confidence   int           `json:"confidence"`
	Basis        string        `json:"basis"`
	Sources      []PriceSource `json:"sources"`
}

type AIJob struct {
	ID        string    `json:"job_id"`
	SellerID  string    `json:"-"`
	Status    string    `json:"status"` // QUEUED | RUNNING | SUCCEEDED | FAILED
	Progress  int       `json:"progress"`
	Steps     []AIStep  `json:"steps"`
	Result    *AIResult `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Token là refresh token hoặc token đặt lại mật khẩu. Chỉ lưu hash SHA-256, dùng một lần.
type Token struct {
	ID        string
	UserID    string
	Kind      string // refresh | reset
	Hash      string
	ExpiresAt time.Time
	RevokedAt time.Time // zero = còn hiệu lực
}

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Kind      string    `json:"kind"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Link      string    `json:"link"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

type AuditLog struct {
	ID        string    `json:"id"`
	ActorID   string    `json:"actor_id"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	CreatedAt time.Time `json:"created_at"`
}

type ListFilter struct {
	Query    string
	Category string
	Sort     string // newest | price_asc | price_desc
	Status   string // mặc định: active
}

type Stats struct {
	Users          int   `json:"users"`
	PendingListing int   `json:"pending_listings"`
	OpenReports    int   `json:"open_reports"`
	Revenue        int64 `json:"revenue"`
}

// Repo là cổng truy cập dữ liệu. Bản Postgres (Supabase) cài cùng giao diện này.
type Repo interface {
	CreateUser(u User) (User, error)
	UserByEmail(email string) (User, error)
	UserByID(id string) (User, error)
	UpdateUser(u User) error
	ListUsers() []User

	Categories() []Category
	CreateListing(l Listing) Listing
	Listing(id string) (Listing, error)
	ListListings(f ListFilter) []Listing
	UpdateListing(l Listing) error

	CreateOrder(o Order) Order
	Order(id string) (Order, error)
	UpdateOrder(o Order) error
	OrdersOfUser(userID string) []Order
	AllOrders() []Order

	CreateReport(r Report) Report
	Report(id string) (Report, error)
	UpdateReport(r Report) error
	ListReports() []Report

	SaveJob(j AIJob)
	Job(id string) (AIJob, error)

	SaveToken(t Token) error
	// ConsumeToken đánh dấu token đã dùng, một cách nguyên tử. Không có/hết hạn: ErrNotFound.
	// Token đã dùng rồi: trả token cùng ErrConflict (dấu hiệu bị đánh cắp/dùng lại).
	ConsumeToken(kind, hash string) (Token, error)
	RevokeTokens(userID, kind string)

	AddNotification(n Notification) Notification
	NotificationsOf(userID string, limit int) (items []Notification, unread int)
	MarkNotificationRead(userID, id string) error // id rỗng = đánh dấu tất cả

	AddAudit(a AuditLog)
	Stats() Stats
}
