import React from 'react';
import { Icon } from '../core/Icon.jsx';
const cx = (...a) => a.filter(Boolean).join(' ');
export function Select({ label, hint, error, options = [], placeholder, id, className, ...rest }) {
  const auto = React.useId();
  const fid = id || auto;
  return (
    <div className={cx('masm-field', className)}>
      {label && <label className="masm-label" htmlFor={fid}>{label}</label>}
      <div className="masm-control">
        <select id={fid} className="masm-input" aria-invalid={error ? true : undefined} {...rest}>
          {placeholder && <option value="" disabled>{placeholder}</option>}
          {options.map(o => typeof o === 'string' ? <option key={o} value={o}>{o}</option> : <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
        <Icon name="chevron-down" size={18} className="masm-chev" />
      </div>
      {error ? <div className="masm-error">{error}</div> : hint && <div className="masm-hint">{hint}</div>}
    </div>
  );
}
