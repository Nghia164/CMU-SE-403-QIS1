import React from 'react';
import { Icon } from '../core/Icon.jsx';
import { Button } from '../core/Button.jsx';
export function AiSuggestion({ label, children, onEdit, onRegenerate, regenerating, editLabel = 'Sửa', regenerateLabel = 'Tạo lại', footer }) {
  return (
    <section className="masm-ai" aria-busy={regenerating || undefined}>
      <div className="masm-ai-head">
        <span className="masm-ai-kicker"><Icon name="sparkles" size={16} />AI gợi ý{label && <span style={{ color: 'var(--text-muted)', fontWeight: 600 }}>· {label}</span>}</span>
        <span className="masm-ai-actions">
          {onEdit && <Button variant="quiet" size="sm" icon="pencil" onClick={onEdit}>{editLabel}</Button>}
          {onRegenerate && <Button variant="quiet" size="sm" icon="refresh-cw" onClick={onRegenerate} loading={regenerating}>{regenerateLabel}</Button>}
        </span>
      </div>
      {regenerating
        ? <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}><div className="masm-shimmer" style={{ height: 14, width: '90%' }} /><div className="masm-shimmer" style={{ height: 14, width: '70%' }} /></div>
        : <div style={{ color: 'var(--text-strong)' }}>{children}</div>}
      {footer}
    </section>
  );
}
