import React from 'react';
import { Icon } from '../core/Icon.jsx';
export function StatCard({ label, value, delta, deltaTone = 'success', icon, hint }) {
  const col = deltaTone === 'danger' ? 'var(--danger-text)' : deltaTone === 'neutral' ? 'var(--text-muted)' : 'var(--success-text)';
  return (
    <div className="masm-card" style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 6, minWidth: 0 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, font: '500 13px/1.3 var(--font-sans)', color: 'var(--text-muted)' }}>{icon && <Icon name={icon} size={16} />}{label}</div>
      <div className="masm-num" style={{ font: 'var(--type-h2)', color: 'var(--text-strong)' }}>{value}</div>
      {(delta || hint) && <div style={{ font: '400 12px/1.3 var(--font-sans)', color: 'var(--text-muted)' }}>{delta && <span className="masm-num" style={{ color: col, fontWeight: 600 }}>{delta}</span>} {hint}</div>}
    </div>
  );
}
