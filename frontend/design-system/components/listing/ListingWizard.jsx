import React from 'react';
import { IconButton } from '../core/IconButton.jsx';
import { Button } from '../core/Button.jsx';
import { Stepper } from '../navigation/Stepper.jsx';
export const SELL_STEPS = ['Ảnh', 'AI phân tích', 'Duyệt & sửa', 'Đăng'];
export function ListingWizard({ steps = SELL_STEPS, current = 0, title = 'Đăng tin bán', onBack, onClose, onNext, nextLabel = 'Tiếp tục', nextDisabled, hideFooter, children, style }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', minHeight: 0, height: '100%', background: 'var(--bg-app)', ...style }}>
      <header style={{ padding: '8px 8px 12px', background: 'var(--surface-card)', borderBottom: '1px solid var(--border-default)', display: 'flex', flexDirection: 'column', gap: 10 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          {current > 0 && onBack ? <IconButton icon="arrow-left" label="Quay lại" onClick={onBack} /> : <IconButton icon="x" label="Đóng" onClick={onClose} />}
          <h1 style={{ flex: 1, margin: 0, font: 'var(--type-title)', color: 'var(--text-strong)' }}>{title}</h1>
          <span className="masm-num" style={{ font: '500 13px/1 var(--font-sans)', color: 'var(--text-muted)', paddingRight: 8 }}>Bước {current + 1}/{steps.length}</span>
        </div>
        <div style={{ padding: '0 8px' }}><Stepper steps={steps} current={current} /></div>
      </header>
      <main style={{ flex: 1, overflow: 'auto', padding: 16, display: 'flex', flexDirection: 'column', gap: 16 }}>{children}</main>
      {!hideFooter && onNext && (
        <footer style={{ padding: '12px 16px', background: 'var(--surface-card)', borderTop: '1px solid var(--border-default)' }}>
          <Button size="lg" block onClick={onNext} disabled={nextDisabled} iconRight="arrow-right">{nextLabel}</Button>
        </footer>
      )}
    </div>
  );
}
