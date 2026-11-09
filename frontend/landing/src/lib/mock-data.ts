import type { Category, Listing, Order, PriceSource } from './types';

export const LISTINGS: Listing[] = [
  { id: 'l1', title: 'Xe đạp Asama 26 inch, mới thay lốp và xích', price: 1450000, condition: 'Tốt', location: 'Q.3, TP.HCM', postedAt: '2 giờ trước', icon: 'bike', cat: 'xe', seller: 'Minh Anh', rating: 4.8, sales: 23,
    desc: 'Xe dùng đi học 2 năm, vừa thay lốp và xích tháng trước. Khung chắc, không gỉ, phanh ăn. Có trầy nhẹ ở khung sau (xem ảnh 3). Ưu tiên giao dịch quanh Q.3.' },
  { id: 'l2', title: 'Máy ảnh Fujifilm X-T20 kèm lens 15-45mm, còn hộp', price: 8900000, original: 12500000, condition: 'Như mới', location: 'Cầu Giấy, Hà Nội', postedAt: 'Hôm qua', icon: 'camera', cat: 'dien-tu', seller: 'Thu Hà', rating: 4.9, sales: 41,
    desc: 'Shot count khoảng 4.000. Còn hộp, sạc, dây đeo zin. Lens không mốc, không rơ.' },
  { id: 'l3', title: 'Ghế gỗ sồi phong cách Bắc Âu', price: 650000, condition: 'Khá', location: 'Hải Châu, Đà Nẵng', postedAt: '3 ngày trước', icon: 'armchair', cat: 'nha', seller: 'Quốc Bảo', rating: 4.6, sales: 7,
    desc: 'Ghế chắc chắn, mặt ngồi có vài vết xước nhỏ.' },
  { id: 'l4', title: 'iPhone 12 128GB, pin 86%, máy zin', price: 6200000, condition: 'Tốt', location: 'Q.7, TP.HCM', postedAt: '5 giờ trước', icon: 'smartphone', cat: 'dien-tu', seller: 'Hoàng Long', rating: 4.7, sales: 12,
    desc: 'Máy zin chưa sửa chữa, Face ID nhạy.' },
  { id: 'l5', title: 'Áo khoác denim Levi’s size M', price: 320000, condition: 'Như mới', location: 'Ninh Kiều, Cần Thơ', postedAt: '1 ngày trước', icon: 'shirt', cat: 'thoi-trang', seller: 'Ngọc Trâm', rating: 5.0, sales: 3,
    desc: 'Mặc 2 lần, không phai màu.' },
  { id: 'l6', title: 'Bộ sách Harry Potter bản tiếng Việt (7 cuốn)', price: 480000, condition: 'Tốt', location: 'Thủ Đức, TP.HCM', postedAt: '6 giờ trước', icon: 'book-open', cat: 'sach', seller: 'Gia Huy', rating: 4.9, sales: 18,
    desc: 'Đủ 7 tập, gáy sách còn đẹp.' },
  { id: 'l7', title: 'Nồi chiên không dầu Philips 4.1L', price: 890000, condition: 'Tốt', location: 'Long Biên, Hà Nội', postedAt: '4 ngày trước', icon: 'cooking-pot', cat: 'nha', seller: 'Thanh Tâm', rating: 4.5, sales: 9,
    desc: 'Hoạt động tốt, lòng nồi chống dính còn mới.', sold: true },
  { id: 'l8', title: 'Đàn guitar acoustic Yamaha F310', price: 1750000, condition: 'Khá', location: 'Biên Hoà, Đồng Nai', postedAt: '1 tuần trước', icon: 'guitar', cat: 'khac', seller: 'Đức Minh', rating: 4.8, sales: 5,
    desc: 'Đàn tiếng ấm, có bao đàn.' },
];

export const CATEGORIES: Category[] = [
  { value: 'all', label: 'Tất cả', icon: 'layout-grid' },
  { value: 'dien-tu', label: 'Điện tử', icon: 'smartphone' },
  { value: 'xe', label: 'Xe cộ', icon: 'bike' },
  { value: 'nha', label: 'Nhà cửa', icon: 'sofa' },
  { value: 'thoi-trang', label: 'Thời trang', icon: 'shirt' },
  { value: 'sach', label: 'Sách', icon: 'book-open' },
  { value: 'khac', label: 'Khác', icon: 'package' },
];

export const PRICE_SOURCES: PriceSource[] = [
  { source: 'Chợ Tốt', title: 'Asama 26" màu xanh, đi ít', price: 1300000, condition: 'Tốt' },
  { source: 'Facebook', title: 'Xe đạp Asama đi học', price: 1150000, condition: 'Khá' },
  { source: 'Chợ Tốt', title: 'Asama AMT 26 inch', price: 1600000, condition: 'Như mới' },
  { source: 'Shopee', title: 'Xe Asama cũ thanh lý', price: 1050000, condition: 'Cũ' },
  { source: 'Facebook', title: 'Asama mới 90%, có giỏ', price: 1700000, condition: 'Như mới' },
  { source: 'Chợ Tốt', title: 'Xe Asama nữ 26"', price: 1400000, condition: 'Tốt' },
];

export const SEED_ORDERS: Order[] = [
  { id: 'DH-2611-08427', listingId: 'l2', status: 'shipping', method: 'VNPay',
    events: { placed: '09:15 · 08/11/2026', paid: '09:16 · 08/11 · qua VNPay', shipping: 'GHN · Mã vận đơn GHN82734 · dự kiến 10/11' } },
  { id: 'DH-2611-07911', listingId: 'l6', status: 'delivered', method: 'MoMo',
    events: { placed: '14:02 · 02/11/2026', paid: '14:03 · qua MoMo', shipping: '03/11 · GHTK', delivered: '05/11 · Đã xác nhận nhận hàng' } },
  { id: 'DH-2610-05530', listingId: 'l3', status: 'cancelled', method: 'VNPay',
    events: { placed: '20:40 · 28/10/2026', cancelled: 'Người bán huỷ · Đã hoàn tiền' } },
];
