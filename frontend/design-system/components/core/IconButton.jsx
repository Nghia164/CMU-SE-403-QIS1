import React from 'react';
import { Icon } from './Icon.jsx';
const cx = (...a) => a.filter(Boolean).join(' ');
export function IconButton({ icon, label, variant = 'quiet', size = 'md', round, float, pressed, className, type = 'button', ...rest }) {
  return (
    <button type={type} aria-label={label} title={label} aria-pressed={pressed}
      className={cx('masm-iconbtn', !float && 'masm-v-' + variant, size === 'sm' && 'masm-iconbtn--sm', (round || float) && 'masm-iconbtn--round', float && 'masm-iconbtn--float', className)} {...rest}>
      <Icon name={icon} size={size === 'sm' ? 18 : 20} />
    </button>
  );
}
