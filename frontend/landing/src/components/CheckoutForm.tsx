'use client';
import { useRouter } from 'next/navigation';
import { useState, type FormEvent } from 'react';
import { vnd } from '@/lib/format';
import { createPaidOrder } from '@/lib/orders-store';
import type { Listing, PaymentProvider } from '@/lib/types';
import { Button, Card, Icon, Input, PaymentMethodPicker, Price, Radio } from './ds';

const SHIP_FEE = 30000;

export function CheckoutForm({ listing: l }: { listing: Listing }) {
  const router = useRouter();
  const [method, setMethod] = useState<PaymentProvider>('vnpay');
  const [paying, setPaying] = useState(false);
  const [name, setName] = useState('Nguyễn Văn An');
  const [phone, setPhone] = useState('0912 345 678');
  const [address, setAddress] = useState('12 Lê Lợi, P. Bến Nghé, Q.1, TP.HCM');
  const [errors, setErrors] = useState<{ name?: string; phone?: string; address?: string }>({});
  const total = l.price + SHIP_FEE;
  const gateway = method === 'vnpay' ? 'VNPay' : 'MoMo';

  const validate = () => {
    const e: typeof errors = {};
    if (!name.trim()) e.name = 'Nhập họ và tên người nhận';
    if (!/^(0|\+84)\d{9,10}$/.test(phone.replace(/\s/g, ''))) e.phone = 'Số điện thoại chưa đúng định dạng, ví dụ 0912 345 678';
    if (address.trim().length < 8) e.address = 'Nhập địa chỉ đầy đủ gồm số nhà, phường, quận';
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const submit = (ev: FormEvent) => {
    ev.preventDefault();
    if (!validate()) return;
    setPaying(true);
    // Mô phỏng chuyển hướng cổng thanh toán và nhận callback; Backend thật sẽ xác minh chữ ký trước khi ghi nhận
    setTimeout(() => {
      const order = createPaidOrder(l.id, method);
      router.push(`/orders/${order.id}?new=1`);
    }, 1100);
  };

  return (
    <div className="container container--narrow stack">
      <h1 className="page-title">Thanh toán</h1>
      <form className="checkout-grid" onSubmit={submit} noValidate>
        <div className="stack">
          <Card className="stack stack--sm">
            <h2 className="card-title">Thông tin nhận hàng</h2>
            <div className="row row--gap-lg">
              <Radio name="ship" label="Giao tận nơi" defaultChecked />
              <Radio name="ship" label="Hẹn gặp trực tiếp" />
            </div>
            <div className="form-2">
              <Input label="Họ và tên" value={name} onChange={(e) => setName(e.target.value)} error={errors.name} autoComplete="name" />
              <Input label="Số điện thoại" value={phone} onChange={(e) => setPhone(e.target.value)} inputMode="tel" error={errors.phone} autoComplete="tel" />
            </div>
            <Input label="Địa chỉ" value={address} onChange={(e) => setAddress(e.target.value)} error={errors.address} autoComplete="street-address" />
          </Card>
          <Card>
            <PaymentMethodPicker value={method} onChange={(v: string) => setMethod(v as PaymentProvider)} />
          </Card>
        </div>

        <Card className="summary stack stack--sm">
          <div className="row">
            <div className="masm-thumb summary-thumb">
              <Icon name={l.icon} size={24} />
            </div>
            <div>
              <div className="summary-title">{l.title}</div>
              <div className="muted small">Người bán: {l.seller}</div>
            </div>
          </div>
          <dl className="sum-lines masm-num">
            <div>
              <dt>Giá sản phẩm</dt>
              <dd>{vnd(l.price)}</dd>
            </div>
            <div>
              <dt>Phí vận chuyển</dt>
              <dd>{vnd(SHIP_FEE)}</dd>
            </div>
          </dl>
          <div className="sum-total">
            <span className="label">Tổng cộng</span>
            <Price value={total} />
          </div>
          <Button size="lg" block type="submit" loading={paying}>
            {paying ? `Đang chuyển đến ${gateway}…` : `Thanh toán ${vnd(total)}`}
          </Button>
          <p className="muted small center">Bạn sẽ được chuyển sang cổng {gateway} để hoàn tất.</p>
        </Card>
      </form>
    </div>
  );
}
