import type { Metadata } from 'next';
import { OrderList } from '@/components/OrderViews';

export const metadata: Metadata = { title: 'Đơn hàng' };

export default function OrdersPage() {
  return <OrderList />;
}
