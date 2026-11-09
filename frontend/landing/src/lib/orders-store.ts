'use client';
/**
 * Kho đơn hàng phía trình duyệt (localStorage) cho bản mock.
 * Khi có Backend, thay bằng gọi API đơn hàng; chữ ký hàm giữ nguyên.
 */
import { SEED_ORDERS } from './mock-data';
import type { Order, PaymentProvider } from './types';

const KEY = 'masm-orders';

function readStored(): Order[] {
  try {
    return JSON.parse(localStorage.getItem(KEY) || '[]') as Order[];
  } catch {
    return [];
  }
}

export function listOrders(): Order[] {
  return [...readStored(), ...SEED_ORDERS];
}

export function getOrder(id: string): Order | undefined {
  return listOrders().find((o) => o.id === id);
}

function pad(n: number) {
  return String(n).padStart(2, '0');
}

export function createPaidOrder(listingId: string, provider: PaymentProvider): Order {
  const now = new Date();
  const code = `DH-${String(now.getFullYear()).slice(2)}${pad(now.getMonth() + 1)}-${String(Math.floor(Math.random() * 90000) + 10000)}`;
  const method = provider === 'vnpay' ? 'VNPay' : 'MoMo';
  const order: Order = {
    id: code,
    listingId,
    status: 'paid',
    method,
    events: { placed: 'Vừa xong', paid: `Vừa xong · qua ${method}` },
  };
  try {
    localStorage.setItem(KEY, JSON.stringify([order, ...readStored()]));
  } catch {
    /* localStorage không khả dụng: đơn chỉ tồn tại trong phiên này */
  }
  return order;
}

export function confirmReceived(id: string) {
  const stored = readStored().map((o) =>
    o.id === id ? { ...o, status: 'delivered' as const, events: { ...o.events, delivered: 'Vừa xong · Đã xác nhận nhận hàng' } } : o,
  );
  try {
    localStorage.setItem(KEY, JSON.stringify(stored));
  } catch {
    /* bỏ qua */
  }
}
