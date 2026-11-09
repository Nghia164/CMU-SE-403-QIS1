/** 1450000 -> "1.450.000 ₫" */
export const vnd = (n: number) => new Intl.NumberFormat('vi-VN').format(n) + ' ₫';

/** Chuỗi số có dấu chấm ngăn nghìn dùng cho ô nhập giá */
export const fmtNum = (n: number) => (n ? new Intl.NumberFormat('vi-VN').format(n) : '');
