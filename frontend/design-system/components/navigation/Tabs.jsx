import React from 'react';
import { Icon } from '../core/Icon.jsx';
export function Tabs({ items = [], value, onChange, variant = 'line', label }) {
  const refs = React.useRef([]);
  const idx = items.findIndex(i => i.value === value);
  const onKey = e => {
    if (e.key !== 'ArrowRight' && e.key !== 'ArrowLeft') return;
    e.preventDefault();
    const n = (idx + (e.key === 'ArrowRight' ? 1 : -1) + items.length) % items.length;
    onChange && onChange(items[n].value); refs.current[n] && refs.current[n].focus();
  };
  return (
    <div role="tablist" aria-label={label} className={'masm-tabs' + (variant === 'pill' ? ' masm-tabs--pill' : '')} onKeyDown={onKey}>
      {items.map((it, i) => (
        <button key={it.value} ref={el => refs.current[i] = el} role="tab" type="button" aria-selected={it.value === value} tabIndex={it.value === value ? 0 : -1}
          className="masm-tab" onClick={() => onChange && onChange(it.value)}>
          {it.icon && <Icon name={it.icon} size={18} />}{it.label}{it.count != null && <span className="masm-tab-count">{it.count}</span>}
        </button>
      ))}
    </div>
  );
}
