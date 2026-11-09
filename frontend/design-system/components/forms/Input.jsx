import React from 'react';
import { Icon } from '../core/Icon.jsx';
const cx = (...a) => a.filter(Boolean).join(' ');
export function Input({ label, hint, error, icon, suffix, multiline, ai, id, className, required, labelExtra, ...rest }) {
  const auto = React.useId();
  const fid = id || auto;
  const desc = error ? fid + '-err' : hint ? fid + '-hint' : undefined;
  const Tag = multiline ? 'textarea' : 'input';
  return (
    <div className={cx('masm-field', className)}>
      {label && <label className="masm-label" htmlFor={fid}>{label}{required && <span aria-hidden="true" style={{ color: 'var(--danger)' }}>*</span>}{labelExtra}</label>}
      <div className="masm-control">
        {icon && !multiline && <Icon name={icon} size={20} />}
        <Tag id={fid} required={required} aria-invalid={error ? true : undefined} aria-describedby={desc}
          className={cx('masm-input', icon && !multiline && 'masm-input--lead', suffix && 'masm-input--trail', ai && 'masm-input--ai')} {...rest} />
        {suffix && <span className="masm-trail">{suffix}</span>}
      </div>
      {error ? <div className="masm-error" id={fid + '-err'}><Icon name="circle-alert" size={14} />{error}</div>
        : hint && <div className="masm-hint" id={fid + '-hint'}>{hint}</div>}
    </div>
  );
}
