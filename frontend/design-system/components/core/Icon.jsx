import React from 'react';
const BASE = 'https://unpkg.com/lucide-static@0.460.0/icons/';
/** Lucide icon rendered via CSS mask so it inherits currentColor. */
export function Icon({ name, size = 20, label, className, style, ...rest }) {
  const a11y = label ? { role: 'img', 'aria-label': label } : { 'aria-hidden': true };
  return <span className={'masm-icon' + (className ? ' ' + className : '')} style={{ width: size, height: size, '--masm-icon': `url(${BASE}${name}.svg)`, ...style }} {...a11y} {...rest} />;
}
