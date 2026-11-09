import React from 'react';
export function Switch({ label, description, className, ...rest }) {
  return (
    <label className={'masm-choice' + (className ? ' ' + className : '')}>
      <input type="checkbox" className="masm-switch" role="switch" {...rest} />
      {(label || description) && <span className="masm-choice-text">{label && <b>{label}</b>}{description && <span style={{ color: 'var(--text-muted)' }}>{description}</span>}</span>}
    </label>
  );
}
