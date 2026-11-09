import React from 'react';
import { Checkbox } from '../forms/Checkbox.jsx';
import { Button } from '../core/Button.jsx';
export function PublishConfirm({ confirmed, onConfirmedChange, onPublish, publishing, children, confirmLabel = 'Tôi đã kiểm tra và xác nhận thông tin tin đăng là chính xác', confirmDescription = 'Nội dung do AI gợi ý chỉ mang tính tham khảo. Tôi chịu trách nhiệm về tin đăng này.', publishLabel = 'Đăng tin' }) {
  const hid = React.useId();
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {children}
      <div style={{ padding: 14, borderRadius: 'var(--radius-lg)', border: '1.5px solid ' + (confirmed ? 'var(--primary)' : 'var(--border-strong)'), background: confirmed ? 'var(--primary-soft)' : 'var(--surface-card)', transition: 'all var(--dur-fast)' }}>
        <Checkbox checked={!!confirmed} onChange={e => onConfirmedChange && onConfirmedChange(e.target.checked)} label={confirmLabel} description={confirmDescription} />
      </div>
      <Button size="lg" block icon="send" disabled={!confirmed} loading={publishing} onClick={onPublish} aria-describedby={!confirmed ? hid : undefined}>{publishLabel}</Button>
      {!confirmed && <div id={hid} style={{ font: '400 12px/1.4 var(--font-sans)', color: 'var(--text-muted)', textAlign: 'center' }}>Tick xác nhận ở trên để bật nút Đăng</div>}
    </div>
  );
}
