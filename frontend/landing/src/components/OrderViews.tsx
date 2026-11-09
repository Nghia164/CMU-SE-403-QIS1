'use client';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import { LISTINGS } from '@/lib/mock-data';
import { confirmReceived, getOrder, listOrders } from '@/lib/orders-store';
import type { Listing, Order } from '@/lib/types';
import { Button, Card, Icon, OrderStatus, Price, Toast } from './ds';

const listingOf = (id: string): Listing => LISTINGS.find((l) => l.id === id) ?? LISTINGS[0];

export function OrderList() {
  const [orders, setOrders] = useState<Order[] | null>(null);
  useEffect(() => setOrders(listOrders()), []);

  return (
    <div className="container container--narrow stack">
      <h1 className="page-title">Đơn hàng của tôi</h1>
      {orders === null ? (
        <p className="muted">Đang tải…</p>
      ) : (
        <ul className="plain-list stack stack--sm">
          {orders.map((o) => {
            const l = listingOf(o.listingId);
            return (
              <li key={o.id}>
                <Card interactive className="order-row">
                  <div className="masm-thumb summary-thumb">
                    <Icon name={l.icon} size={24} />
                  </div>
                  <div className="order-main">
                    <Link href={`/orders/${o.id}`} className="order-link">
                      {l.title}
                    </Link>
                    <div className="muted small mono">{o.id}</div>
                  </div>
                  <Price value={l.price} size="sm" />
                  <OrderStatus variant="badge" status={o.status} />
                </Card>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}

export function OrderDetail({ id }: { id: string }) {
  const params = useSearchParams();
  const [order, setOrder] = useState<Order | null | undefined>(undefined);
  const [toast, setToast] = useState(false);

  useEffect(() => {
    setOrder(getOrder(id) ?? null);
    setToast(params.get('new') === '1');
  }, [id, params]);

  if (order === undefined) return <div className="container"><p className="muted">Đang tải…</p></div>;
  if (order === null) {
    return (
      <div className="container container--narrow stack">
        <h1 className="page-title">Không tìm thấy đơn hàng</h1>
        <p className="muted">Mã đơn {id} không tồn tại hoặc không thuộc tài khoản của bạn.</p>
        <Button as={Link} href="/orders" variant="secondary">
          Về danh sách đơn
        </Button>
      </div>
    );
  }
  const l = listingOf(order.listingId);

  return (
    <div className="container container--narrow stack">
      <div className="row row--wrap">
        <h1 className="page-title">Đơn hàng</h1>
        <span className="muted mono">{order.id}</span>
        <OrderStatus variant="badge" status={order.status} />
      </div>
      <div className="two-col">
        <Card>
          <h2 className="card-title">Trạng thái</h2>
          <OrderStatus status={order.status} events={order.events} />
        </Card>
        <Card className="stack stack--sm">
          <div className="row">
            <div className="masm-thumb summary-thumb">
              <Icon name={l.icon} size={26} />
            </div>
            <div>
              <Link href={`/listing/${l.id}`} className="order-link">
                {l.title}
              </Link>
              <Price value={l.price} size="sm" />
            </div>
          </div>
          <div className="row">
            <Button variant="secondary" icon="message-circle" block>
              Chat với người bán
            </Button>
            <Button
              block
              icon="package-check"
              disabled={order.status !== 'shipping'}
              onClick={() => {
                confirmReceived(order.id);
                setOrder(getOrder(order.id) ?? order);
              }}
            >
              Đã nhận hàng
            </Button>
          </div>
          <p className="muted small">Chỉ bấm “Đã nhận hàng” sau khi kiểm tra sản phẩm. Tiền sẽ được chuyển cho người bán.</p>
        </Card>
      </div>
      {toast && (
        <div className="toast-slot" role="status">
          <Toast tone="success" title="Thanh toán thành công" message="Người bán đã được thông báo để chuẩn bị hàng." onClose={() => setToast(false)} />
        </div>
      )}
    </div>
  );
}
