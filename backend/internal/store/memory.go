package store

import (
	"crypto/rand"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Memory là Repo lưu trong RAM: dùng cho dev và test. Mất dữ liệu khi khởi động lại.
type Memory struct {
	mu      sync.RWMutex
	users   map[string]User
	listing map[string]Listing
	orders  map[string]Order
	reports map[string]Report
	jobs    map[string]AIJob
	tokens  map[string]Token // key: kind + ":" + hash
	notifs  []Notification
	audits  []AuditLog
	cats    []Category
}

func NewMemory() *Memory {
	return &Memory{
		users: map[string]User{}, listing: map[string]Listing{}, orders: map[string]Order{},
		reports: map[string]Report{}, jobs: map[string]AIJob{}, tokens: map[string]Token{},
		cats: []Category{
			{"dien-tu", "Điện tử", "smartphone"}, {"xe", "Xe cộ", "bike"}, {"nha", "Nhà cửa", "sofa"},
			{"thoi-trang", "Thời trang", "shirt"}, {"sach", "Sách", "book-open"}, {"khac", "Khác", "package"},
		},
	}
}

// NewID tạo mã ngẫu nhiên không đoán được.
func NewID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

// Fold bỏ dấu tiếng Việt và hạ chữ thường để tìm kiếm không dấu.
func Fold(s string) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "đ", "d"), "Đ", "d")
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return strings.ToLower(b.String())
}

func (m *Memory) CreateUser(u User) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.users {
		if strings.EqualFold(x.Email, u.Email) {
			return User{}, ErrConflict
		}
	}
	u.ID, u.CreatedAt = NewID("u_"), time.Now()
	m.users[u.ID] = u
	return u, nil
}

func (m *Memory) UserByEmail(email string) (User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return User{}, ErrNotFound
}

func (m *Memory) UserByID(id string) (User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return User{}, ErrNotFound
}

func (m *Memory) UpdateUser(u User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[u.ID]; !ok {
		return ErrNotFound
	}
	m.users[u.ID] = u
	return nil
}

func (m *Memory) ListUsers() []User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	counts := map[string]int{}
	for _, l := range m.listing {
		counts[l.SellerID]++
	}
	out := make([]User, 0, len(m.users))
	for _, u := range m.users {
		u.ListingCount = counts[u.ID]
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (m *Memory) Categories() []Category { return m.cats }

func (m *Memory) CreateListing(l Listing) Listing {
	m.mu.Lock()
	defer m.mu.Unlock()
	if l.ID == "" {
		l.ID = NewID("l_")
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now()
	}
	m.listing[l.ID] = l
	return l
}

func (m *Memory) Listing(id string) (Listing, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if l, ok := m.listing[id]; ok {
		return l, nil
	}
	return Listing{}, ErrNotFound
}

func (m *Memory) ListListings(f ListFilter) []Listing {
	m.mu.RLock()
	defer m.mu.RUnlock()
	status := f.Status
	if status == "" {
		status = "active"
	}
	q := Fold(strings.TrimSpace(f.Query))
	var out []Listing
	for _, l := range m.listing {
		visible := l.Status == status || (status == "active" && l.Status == "sold")
		if status == "all" {
			visible = true
		}
		if !visible {
			continue
		}
		if f.Category != "" && f.Category != "all" && l.Category != f.Category {
			continue
		}
		if q != "" && !strings.Contains(Fold(l.Title+" "+l.Description), q) {
			continue
		}
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool {
		switch f.Sort {
		case "price_asc":
			return out[i].Price < out[j].Price
		case "price_desc":
			return out[i].Price > out[j].Price
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (m *Memory) UpdateListing(l Listing) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.listing[l.ID]; !ok {
		return ErrNotFound
	}
	m.listing[l.ID] = l
	return nil
}

func (m *Memory) CreateOrder(o Order) Order {
	m.mu.Lock()
	defer m.mu.Unlock()
	o.ID, o.CreatedAt = NewID("DH-"), time.Now()
	m.orders[o.ID] = o
	return o
}

func (m *Memory) Order(id string) (Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if o, ok := m.orders[id]; ok {
		return o, nil
	}
	return Order{}, ErrNotFound
}

func (m *Memory) UpdateOrder(o Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.orders[o.ID]; !ok {
		return ErrNotFound
	}
	// Giống ràng buộc duy nhất của Postgres: mỗi tin chỉ có một đơn đã thanh toán.
	if paidLike(o.Status) {
		for id, x := range m.orders {
			if id != o.ID && x.ListingID == o.ListingID && paidLike(x.Status) {
				return ErrConflict
			}
		}
	}
	m.orders[o.ID] = o
	return nil
}

func paidLike(s string) bool { return s == "paid" || s == "shipping" || s == "delivered" }

func (m *Memory) sortedOrders(keep func(Order) bool) []Order {
	var out []Order
	for _, o := range m.orders {
		if keep(o) {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (m *Memory) OrdersOfUser(id string) []Order {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sortedOrders(func(o Order) bool { return o.BuyerID == id || o.SellerID == id })
}

func (m *Memory) AllOrders() []Order {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sortedOrders(func(Order) bool { return true })
}

func (m *Memory) CreateReport(r Report) Report {
	m.mu.Lock()
	defer m.mu.Unlock()
	r.ID, r.CreatedAt, r.Status = NewID("BC-"), time.Now(), "new"
	r.Count = 1
	for id, x := range m.reports {
		if x.TargetType == r.TargetType && x.TargetID == r.TargetID && x.Status != "done" {
			x.Count++
			m.reports[id] = x
			r.Count = x.Count
		}
	}
	m.reports[r.ID] = r
	return r
}

func (m *Memory) Report(id string) (Report, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if r, ok := m.reports[id]; ok {
		return r, nil
	}
	return Report{}, ErrNotFound
}

func (m *Memory) UpdateReport(r Report) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.reports[r.ID]; !ok {
		return ErrNotFound
	}
	m.reports[r.ID] = r
	return nil
}

func (m *Memory) ListReports() []Report {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Report, 0, len(m.reports))
	for _, r := range m.reports {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (m *Memory) SaveJob(j AIJob) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j.Steps = append([]AIStep(nil), j.Steps...)
	m.jobs[j.ID] = j
}

func (m *Memory) Job(id string) (AIJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, ok := m.jobs[id]
	if !ok {
		return AIJob{}, ErrNotFound
	}
	j.Steps = append([]AIStep(nil), j.Steps...)
	return j, nil
}

func (m *Memory) AddAudit(a AuditLog) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a.ID, a.CreatedAt = NewID("a_"), time.Now()
	m.audits = append(m.audits, a)
}

func (m *Memory) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := Stats{Users: len(m.users)}
	for _, l := range m.listing {
		if l.Status == "pending_review" {
			s.PendingListing++
		}
	}
	for _, r := range m.reports {
		if r.Status != "done" {
			s.OpenReports++
		}
	}
	for _, o := range m.orders {
		if o.Status == "paid" || o.Status == "shipping" || o.Status == "delivered" {
			s.Revenue += o.Amount
		}
	}
	return s
}

func (m *Memory) SaveToken(t Token) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[t.Kind+":"+t.Hash] = t
	return nil
}

func (m *Memory) ConsumeToken(kind, hash string) (Token, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tokens[kind+":"+hash]
	if !ok || time.Now().After(t.ExpiresAt) {
		return Token{}, ErrNotFound
	}
	if !t.RevokedAt.IsZero() {
		return t, ErrConflict
	}
	t.RevokedAt = time.Now()
	m.tokens[kind+":"+hash] = t
	return t, nil
}

func (m *Memory) RevokeTokens(userID, kind string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, t := range m.tokens {
		if t.UserID == userID && t.Kind == kind && t.RevokedAt.IsZero() {
			t.RevokedAt = time.Now()
			m.tokens[k] = t
		}
	}
}

func (m *Memory) AddNotification(n Notification) Notification {
	m.mu.Lock()
	defer m.mu.Unlock()
	n.ID, n.CreatedAt, n.Read = NewID("n_"), time.Now(), false
	m.notifs = append(m.notifs, n)
	return n
}

func (m *Memory) NotificationsOf(userID string, limit int) ([]Notification, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Notification
	unread := 0
	for i := len(m.notifs) - 1; i >= 0; i-- { // mới nhất trước
		n := m.notifs[i]
		if n.UserID != userID {
			continue
		}
		if !n.Read {
			unread++
		}
		if len(out) < limit {
			out = append(out, n)
		}
	}
	return out, unread
}

func (m *Memory) MarkNotificationRead(userID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	found := id == ""
	for i := range m.notifs {
		if m.notifs[i].UserID == userID && (id == "" || m.notifs[i].ID == id) {
			m.notifs[i].Read, found = true, true
		}
	}
	if !found {
		return ErrNotFound
	}
	return nil
}