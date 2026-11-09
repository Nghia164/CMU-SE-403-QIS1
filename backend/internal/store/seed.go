package store

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

func hash(p string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(h)
}

// Seed tạo tài khoản quản trị và (tuỳ chọn) dữ liệu mẫu.
func Seed(r Repo, adminEmail, adminPassword string, demo bool) {
	if _, err := r.UserByEmail(adminEmail); err == nil {
		return // đã seed (CSDL bền vững): không nhân đôi dữ liệu
	}
	_, _ = r.CreateUser(User{Name: "Quản trị viên", Email: adminEmail, PasswordHash: hash(adminPassword), Role: RoleAdmin, Status: "active"})
	if !demo {
		return
	}
	seller, _ := r.CreateUser(User{Name: "Minh Anh", Email: "seller@masm.vn", PasswordHash: hash("Demo@12345"), Role: RoleSeller, Status: "active"})
	_, _ = r.CreateUser(User{Name: "Ngọc Trâm", Email: "buyer@masm.vn", PasswordHash: hash("Demo@12345"), Role: RoleBuyer, Status: "active"})

	now := time.Now()
	type s struct {
		title, desc, cond, loc, cat, icon string
		price, orig                        int64
		ago                                time.Duration
		status                             string
	}
	for _, x := range []s{
		{"Xe đạp Asama 26 inch, mới thay lốp và xích", "Xe dùng đi học 2 năm, vừa thay lốp và xích tháng trước. Khung chắc, không gỉ, phanh ăn.", "Tốt", "Q.3, TP.HCM", "xe", "bike", 1450000, 0, 2 * time.Hour, "active"},
		{"Máy ảnh Fujifilm X-T20 kèm lens 15-45mm, còn hộp", "Shot count khoảng 4.000. Còn hộp, sạc, dây đeo zin. Lens không mốc, không rơ.", "Như mới", "Cầu Giấy, Hà Nội", "dien-tu", "camera", 8900000, 12500000, 24 * time.Hour, "active"},
		{"Ghế gỗ sồi phong cách Bắc Âu", "Ghế chắc chắn, mặt ngồi có vài vết xước nhỏ.", "Khá", "Hải Châu, Đà Nẵng", "nha", "armchair", 650000, 0, 72 * time.Hour, "active"},
		{"iPhone 12 128GB, pin 86%, máy zin", "Máy zin chưa sửa chữa, Face ID nhạy.", "Tốt", "Q.7, TP.HCM", "dien-tu", "smartphone", 6200000, 0, 5 * time.Hour, "active"},
		{"Áo khoác denim Levi’s size M", "Mặc 2 lần, không phai màu.", "Như mới", "Ninh Kiều, Cần Thơ", "thoi-trang", "shirt", 320000, 0, 26 * time.Hour, "active"},
		{"Bộ sách Harry Potter bản tiếng Việt (7 cuốn)", "Đủ 7 tập, gáy sách còn đẹp.", "Tốt", "Thủ Đức, TP.HCM", "sach", "book-open", 480000, 0, 6 * time.Hour, "active"},
		{"Nồi chiên không dầu Philips 4.1L", "Hoạt động tốt, lòng nồi chống dính còn mới.", "Tốt", "Long Biên, Hà Nội", "nha", "cooking-pot", 890000, 0, 96 * time.Hour, "sold"},
		{"Đàn guitar acoustic Yamaha F310", "Đàn tiếng ấm, có bao đàn.", "Khá", "Biên Hoà, Đồng Nai", "khac", "guitar", 1750000, 0, 168 * time.Hour, "active"},
		{"Túi xách hàng hiệu like auth giá rẻ", "Túi nhập khẩu, giá thanh lý.", "Như mới", "Q.1, TP.HCM", "thoi-trang", "shopping-bag", 350000, 0, 12 * time.Minute, "pending_review"},
	} {
		l := Listing{
			Title: x.title, Description: x.desc, Condition: x.cond, Location: x.loc, Category: x.cat, Icon: x.icon,
			Price: x.price, OriginalPrice: x.orig, SellerID: seller.ID, SellerName: seller.Name,
			Status: x.status, CreatedAt: now.Add(-x.ago),
		}
		if x.status == "pending_review" {
			l.AIVerdict, l.AIScore = "Nghi hàng giả", 71
		}
		r.CreateListing(l)
	}
}
