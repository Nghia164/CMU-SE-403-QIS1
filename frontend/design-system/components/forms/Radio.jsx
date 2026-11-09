import React from 'react';
export function Radio({ label, description, className, ...rest }) {
  return (
    <label className={'masm-choice' + (className ? ' ' + className : '')}>
      <input type="radio" className="masm-radio" {...rest} />
      {(label || description) && <span className="masm-choice-text">{label && <b>{label}</b>}{description && <span style={{ color: 'var(--text-muted)' }}>{description}</span>}</span>}
    </label>
  );
}
