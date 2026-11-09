import React from 'react';
import { Icon } from './Icon.jsx';
const cx = (...a) => a.filter(Boolean).join(' ');
export function Badge({ tone = 'neutral', size = 'md', icon, children, className, ...rest }) {
  return (
    <span className={cx('masm-badge', 'masm-tone-' + tone, size === 'sm' && 'masm-badge--sm', className)} {...rest}>
      {icon && <Icon name={icon} size={size === 'sm' ? 12 : 14} />}{children}
    </span>
  );
}
