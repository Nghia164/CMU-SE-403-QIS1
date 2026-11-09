import React from 'react';
import { Icon } from '../core/Icon.jsx';
export const PAYMENT_METHODS = [
  { value: 'vnpay', name: 'VNPay', desc: 'Thẻ ATM nội địa, Visa/Mastercard, QR ngân hàng', logo: 'VNPAY', color: 'var(--vnpay)' },
  { value: 'momo', name: 'Ví MoMo', desc: 'Thanh toán nhanh qua ứng dụng MoMo', logo: 'MoMo', color: 'var(--momo)' },
];
export function PaymentMethodPicker({ value, onChange, methods = PAYMENT_METHODS, name = 'payment', legend = 'Phương thức thanh toán' }) {
  return (
    <fieldset style={{ border: 0, margin: 0, padding: 0, display: 'flex', flexDirection: 'column', gap: 8, minWidth: 0 }}>
      <legend style={{ font: 'var(--type-label)', color: 'var(--text-strong)', marginBottom: 8, padding: 0 }}>{legend}</legend>
      {methods.map(m => (
        <label key={m.value} className="masm-pay">
          <input type="radio" className="masm-sr" name={name} value={m.value} checked={value === m.value} disabled={m.disabled} onChange={() => onChange && onChange(m.value)} />
          <span className="masm-pay-logo" style={{ background: m.color }} aria-hidden="true">{m.logo}</span>
          <span style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 2 }}>
            <span style={{ font: 'var(--type-title)', color: 'var(--text-strong)' }}>{m.name}</span>
            <span style={{ font: '400 13px/1.4 var(--font-sans)', color: 'var(--text-muted)' }}>{m.desc}</span>
          </span>
          <span aria-hidden="true" style={{ width: 22, height: 22, borderRadius: '50%', flex: 'none', border: value === m.value ? '7px solid var(--primary)' : '1.5px solid var(--border-strong)', background: 'var(--surface-card)' }} />
        </label>
      ))}
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, font: '400 12px/1.4 var(--font-sans)', color: 'var(--text-muted)', marginTop: 4 }}><Icon name="shield-check" size={14} style={{ color: 'var(--primary)' }} />Tiền được giữ an toàn đến khi bạn xác nhận đã nhận hàng.</div>
    </fieldset>
  );
}
