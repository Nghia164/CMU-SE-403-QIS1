import React from 'react';
import { Icon } from '../core/Icon.jsx';
import { Badge } from '../core/Badge.jsx';
export const ORDER_STEPS = [
  { key: 'placed', label: 'Đã đặt hàng', icon: 'receipt' },
  { key: 'paid', label: 'Đã thanh toán', icon: 'wallet' },
  { key: 'shipping', label: 'Đang giao hàng', icon: 'truck' },
  { key: 'delivered', label: 'Đã nhận hàng', icon: 'package-check' },
];
export const ORDER_BADGE = {
  pending_payment: ['Chờ thanh toán', 'warning'], placed: ['Chờ thanh toán', 'warning'], paid: ['Đã thanh toán', 'info'],
  shipping: ['Đang giao', 'primary'], delivered: ['Hoàn tất', 'success'], cancelled: ['Đã huỷ', 'danger'], refunded: ['Đã hoàn tiền', 'neutral'],
};
export function OrderStatus({ status = 'placed', variant = 'timeline', events = {} }) {
  if (variant === 'badge') { const [l, t] = ORDER_BADGE[status] || [status, 'neutral']; return <Badge tone={t}>{l}</Badge>; }
  const idx = status === 'pending_payment' ? 0 : ORDER_STEPS.findIndex(s => s.key === status);
  const steps = status === 'cancelled' ? [ORDER_STEPS[0], { key: 'cancelled', label: 'Đơn đã huỷ', icon: 'circle-x' }] : ORDER_STEPS;
  return (
    <ol className="masm-otl" aria-label="Trạng thái đơn hàng">
      {steps.map((s, i) => {
        const state = status === 'cancelled' ? (i === 0 ? 'done' : 'cancelled') : i < idx || (i === idx && status === 'delivered') ? 'done' : i === idx ? 'current' : 'upcoming';
        return (
          <li key={s.key} data-state={state} aria-current={state === 'current' ? 'step' : undefined}>
            <span className="masm-otl-dot"><Icon name={state === 'done' ? 'check' : s.icon} size={16} /></span>
            <span><div className="masm-otl-t">{s.label}</div>{events[s.key] && <div className="masm-otl-s">{events[s.key]}</div>}</span>
          </li>
        );
      })}
    </ol>
  );
}
