import React from 'react';
import { Icon } from '../core/Icon.jsx';
const TONE_ICON = { success: ['circle-check', 'var(--teal-300)'], danger: ['circle-x', '#F5A49C'], info: ['info', '#A9CBF0'], ai: ['sparkles', 'var(--amber-300)'] };
export function Toast({ tone = 'success', title, message, actionLabel, onAction, onClose, style }) {
  const [ic, col] = TONE_ICON[tone] || TONE_ICON.info;
  return (
    <div className="masm-toast" role="status" aria-live="polite" style={style}>
      <Icon name={ic} size={20} style={{ color: col, marginTop: 1 }} />
      <div style={{ flex: 1 }}>{title && <b>{title}</b>}{message}</div>
      {actionLabel && <button className="masm-btn masm-v-ghost" style={{ background: 'none' }} onClick={onAction}>{actionLabel}</button>}
      {onClose && <button className="masm-btn masm-v-ghost" aria-label="Đóng" onClick={onClose} style={{ background: 'none', color: 'inherit', padding: 4 }}><Icon name="x" size={16} /></button>}
    </div>
  );
}
