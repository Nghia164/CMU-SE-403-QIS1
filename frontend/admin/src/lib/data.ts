/** Dữ liệu giả cho Admin. Khi có Backend, thay bằng gọi API (cùng hình dạng). */
export type Tone = 'success' | 'danger' | 'warning' | 'info' | 'neutral';

export interface AUser { id: string; name: string; email: string; role: string; joined: string; listings: number; status: 'active' | 'locked' | 'pending' }
export interface AMod { id: string; title: string; icon: string; seller: string; price: number; ai: [string, number]; at: string }
export interface AReport { id: string; target: string; kind: string; reason: string; reporter: string; count: number; status: 'new' | 'progress' | 'done' }
export interface AOrder { id: string; item: string; buyer: string; seller: string; amount: number; method: 'VNPay' | 'MoMo'; status: 'paid' | 'shipping' | 'done' | 'refund' }

export const USERS: AUser[] = [
  { id: 'u1', name: 'Nguyễn Minh Anh', email: 'minhanh@gmail.com', role: 'Người bán', joined: '12/03/2026', listings: 23, status: 'active' },
  { id: 'u2', name: 'Trần Quốc Bảo', email: 'bao.tran@gmail.com', role: 'Người mua', joined: '02/07/2026', listings: 0, status: 'locked' },
  { id: 'u3', name: 'Lê Thu Hà', email: 'thuha.le@gmail.com', role: 'Người bán', joined: '28/10/2026', listings: 4, status: 'pending' },
  { id: 'u4', name: 'Phạm Gia Huy', email: 'giahuy.p@gmail.com', role: 'Người bán', joined: '15/01/2026', listings: 18, status: 'active' },
  { id: 'u5', name: 'Võ Ngọc Trâm', email: 'ngoctram.vo@gmail.com', role: 'Người mua', joined: '09/11/2026', listings: 0, status: 'active' },
];
export const MOD: AMod[] = [
  { id: 'm1', title: 'Xe đạp Asama 26 inch, mới thay lốp', icon: 'bike', seller: 'Nguyễn Minh Anh', price: 1450000, ai: ['Hợp lệ', 96], at: '5 phút trước' },
  { id: 'm2', title: 'Túi xách hàng hiệu like auth giá rẻ', icon: 'shopping-bag', seller: 'Shop Thanh Lý 247', price: 350000, ai: ['Nghi hàng giả', 71], at: '12 phút trước' },
  { id: 'm3', title: 'Máy ảnh Fujifilm X-T20 kèm lens', icon: 'camera', seller: 'Lê Thu Hà', price: 8900000, ai: ['Hợp lệ', 93], at: '18 phút trước' },
  { id: 'm4', title: 'Thuốc giảm cân nhập khẩu', icon: 'pill', seller: 'Bảo An Store', price: 220000, ai: ['Hàng cấm', 88], at: '25 phút trước' },
  { id: 'm5', title: 'Ghế gỗ sồi phong cách Bắc Âu', icon: 'armchair', seller: 'Phạm Gia Huy', price: 650000, ai: ['Giá bất thường', 58], at: '40 phút trước' },
];
export const REPORTS: AReport[] = [
  { id: 'BC-1093', target: 'Túi xách hàng hiệu like auth', kind: 'Tin đăng', reason: 'Hàng giả, nhái thương hiệu', reporter: 'Võ Ngọc Trâm', count: 4, status: 'new' },
  { id: 'BC-1092', target: 'Shop Thanh Lý 247', kind: 'Tài khoản', reason: 'Lừa đảo, không giao hàng', reporter: 'Trần Quốc Bảo', count: 2, status: 'progress' },
  { id: 'BC-1088', target: 'iPhone 12 128GB pin 86%', kind: 'Tin đăng', reason: 'Mô tả sai tình trạng', reporter: 'Phạm Gia Huy', count: 1, status: 'new' },
  { id: 'BC-1081', target: 'Đàn guitar Yamaha F310', kind: 'Tin đăng', reason: 'Spam, đăng trùng lặp', reporter: 'Lê Thu Hà', count: 1, status: 'done' },
];
export const ORDERS: AOrder[] = [
  { id: 'DH-240981', item: 'Xe đạp Asama 26 inch', buyer: 'Võ Ngọc Trâm', seller: 'Nguyễn Minh Anh', amount: 1450000, method: 'MoMo', status: 'paid' },
  { id: 'DH-240977', item: 'Máy ảnh Fujifilm X-T20', buyer: 'Phạm Gia Huy', seller: 'Lê Thu Hà', amount: 8900000, method: 'VNPay', status: 'shipping' },
  { id: 'DH-240960', item: 'Ghế gỗ sồi Bắc Âu', buyer: 'Trần Quốc Bảo', seller: 'Phạm Gia Huy', amount: 650000, method: 'VNPay', status: 'done' },
  { id: 'DH-240951', item: 'Đàn guitar Yamaha F310', buyer: 'Võ Ngọc Trâm', seller: 'Nguyễn Minh Anh', amount: 1900000, method: 'MoMo', status: 'refund' },
];

export const USER_STATUS: Record<AUser['status'], [string, Tone]> = { active: ['Hoạt động', 'success'], locked: ['Đã khoá', 'danger'], pending: ['Chờ xác minh', 'warning'] };
export const REPORT_STATUS: Record<AReport['status'], [string, Tone]> = { new: ['Mới', 'danger'], progress: ['Đang xử lý', 'info'], done: ['Đã xử lý', 'neutral'] };
export const ORDER_STATUS: Record<AOrder['status'], [string, Tone]> = { paid: ['Đã thanh toán', 'info'], shipping: ['Đang giao', 'warning'], done: ['Hoàn tất', 'success'], refund: ['Hoàn tiền', 'danger'] };

export const vnd = (n: number) => new Intl.NumberFormat('vi-VN').format(n) + ' ₫';
