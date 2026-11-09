import React from 'react';
const cx = (...a) => a.filter(Boolean).join(' ');
export function Card({ padding = 16, interactive, flat, as: Tag = 'div', className, style, children, ...rest }) {
  return <Tag className={cx('masm-card', interactive && 'masm-card--interactive', flat && 'masm-card--flat', className)} style={{ padding, ...style }} {...rest}>{children}</Tag>;
}
