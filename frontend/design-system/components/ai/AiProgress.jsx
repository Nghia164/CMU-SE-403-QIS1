import React from 'react';
import { Icon } from '../core/Icon.jsx';
export function AiProgress({ title = 'AI đang phân tích ảnh của bạn', steps = [], progress, onCancel }) {
  const pct = progress != null ? progress : Math.round(steps.filter(s => s.status === 'done').length / Math.max(1, steps.length) * 100);
  return (
    <section className="masm-ai" aria-label={title}>
      <div className="masm-ai-head">
        <span className="masm-ai-kicker"><Icon name="sparkles" size={16} />{title}</span>
        <span className="masm-num" style={{ marginLeft: 'auto', font: '700 14px/1 var(--font-sans)', color: 'var(--ai-text)' }}>{pct}%</span>
      </div>
      <div className="masm-meter" role="progressbar" aria-label={title} aria-valuemin={0} aria-valuemax={100} aria-valuenow={pct}><i style={{ width: pct + '%' }} /></div>
      <ol aria-live="polite" style={{ listStyle: 'none', margin: 0, padding: 0, display: 'flex', flexDirection: 'column', gap: 10 }}>
        {steps.map((s, i) => (
          <li key={i} style={{ display: 'flex', alignItems: 'flex-start', gap: 10, font: 'var(--type-body-sm)' }}>
            <span style={{ width: 20, height: 20, flex: 'none', display: 'grid', placeItems: 'center', color: s.status === 'done' ? 'var(--primary)' : 'var(--ai-icon)' }}>
              {s.status === 'done' ? <Icon name="circle-check" size={20} /> : s.status === 'running' ? <span className="masm-spinner" style={{ width: 16, height: 16 }} /> : s.status === 'error' ? <Icon name="circle-alert" size={20} style={{ color: 'var(--danger)' }} /> : <Icon name="circle" size={18} style={{ color: 'var(--ai-border)' }} />}
            </span>
            <span style={{ display: 'flex', flexDirection: 'column', gap: 2, minWidth: 0 }}>
              <span style={{ fontWeight: s.status === 'running' ? 600 : 500, color: s.status === 'pending' ? 'var(--text-subtle)' : 'var(--text-strong)' }}>{s.label}</span>
              {s.detail && <span style={{ font: '400 12px/1.4 var(--font-sans)', color: 'var(--text-muted)' }}>{s.detail}</span>}
            </span>
          </li>
        ))}
      </ol>
      {onCancel && <button type="button" className="masm-btn masm-v-quiet masm-btn--sm" style={{ alignSelf: 'flex-start', color: 'var(--ai-text)' }} onClick={onCancel}>Huỷ và nhập thủ công</button>}
    </section>
  );
}
