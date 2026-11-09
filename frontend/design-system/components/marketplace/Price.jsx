import React from 'react';
export const formatVND = n => new Intl.NumberFormat('vi-VN').format(Math.round(n)) + ' ₫';
const SIZES = { sm: 'var(--type-label)', md: 'var(--type-price)', lg: 'var(--type-price-lg)' };
export function Price({ value, original, size = 'md', tone = 'strong', style }) {
  const color = tone === 'primary' ? 'var(--primary-text)' : tone === 'ai' ? 'var(--ai-text)' : 'var(--text-strong)';
  return (
    <span className="masm-num" style={{ display: 'inline-flex', alignItems: 'baseline', gap: 8, flexWrap: 'wrap', ...style }}>
      <span style={{ font: SIZES[size], color }}>{formatVND(value)}</span>
      {original != null && <s style={{ font: '400 13px/1 var(--font-sans)', color: 'var(--text-subtle)' }}><span className="masm-sr">Giá gốc </span>{formatVND(original)}</s>}
    </span>
  );
}
