import React from 'react';
export const confidenceLevel = v => v >= 75 ? 'Cao' : v >= 50 ? 'Trung bình' : 'Thấp';
export function ConfidenceBar({ value = 0, label = 'Độ tin cậy', note }) {
  const v = Math.max(0, Math.min(100, value));
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8, font: 'var(--type-body-sm)' }}>
        <span style={{ color: 'var(--text-muted)' }}>{label}</span>
        <span className="masm-num" style={{ fontWeight: 600, color: 'var(--ai-text)' }}>{v}% · {confidenceLevel(v)}</span>
      </div>
      <div className="masm-meter" role="meter" aria-label={label} aria-valuemin={0} aria-valuemax={100} aria-valuenow={v} aria-valuetext={v + '%, ' + confidenceLevel(v)}><i style={{ width: v + '%' }} /></div>
      {note && <div style={{ font: '400 12px/1.4 var(--font-sans)', color: 'var(--text-muted)' }}>{note}</div>}
    </div>
  );
}
