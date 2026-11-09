import React from 'react';
import { IconButton } from '../core/IconButton.jsx';
export function Dialog({ open, onClose, title, children, footer, width = 480 }) {
  const ref = React.useRef(null);
  const tid = React.useId();
  React.useEffect(() => {
    if (!open) return;
    const prev = document.activeElement;
    const el = ref.current; el && el.focus();
    const onKey = e => { if (e.key === 'Escape') onClose && onClose(); };
    document.addEventListener('keydown', onKey);
    return () => { document.removeEventListener('keydown', onKey); prev && prev.focus && prev.focus(); };
  }, [open]);
  if (!open) return null;
  return (
    <div className="masm-overlay" onMouseDown={e => { if (e.target === e.currentTarget) onClose && onClose(); }}>
      <div ref={ref} tabIndex={-1} role="dialog" aria-modal="true" aria-labelledby={tid} className="masm-dialog" style={{ maxWidth: width, outline: 'none' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <h2 id={tid} className="masm-dialog-title" style={{ flex: 1 }}>{title}</h2>
          {onClose && <IconButton icon="x" label="Đóng" size="sm" onClick={onClose} />}
        </div>
        <div style={{ font: 'var(--type-body-sm)' }}>{children}</div>
        {footer && <div className="masm-dialog-foot">{footer}</div>}
      </div>
    </div>
  );
}
