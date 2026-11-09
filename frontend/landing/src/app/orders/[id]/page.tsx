import type { Metadata } from 'next';
import { Suspense } from 'react';
import { OrderDetail } from '@/components/OrderViews';

export const metadata: Metadata = { title: 'Chi tiết đơn hàng' };

export default function OrderPage({ params }: { params: { id: string } }) {
  return (
    <Suspense fallback={null}>
      <OrderDetail id={params.id} />
    </Suspense>
  );
}
