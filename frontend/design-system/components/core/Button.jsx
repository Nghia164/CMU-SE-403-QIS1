import React from 'react';
import { Icon } from './Icon.jsx';
const cx = (...a) => a.filter(Boolean).join(' ');
export function Button({ variant = 'primary', size = 'md', block, icon, iconRight, loading, disabled, children, className, as, type = 'button', ...rest }) {
  const Tag = as || (rest.href ? 'a' : 'button');
  const isz = size === 'lg' ? 20 : 18;
  return (
    <Tag type={Tag === 'button' ? type : undefined} disabled={Tag === 'button' ? (disabled || loading) : undefined} aria-busy={loading || undefined}
      className={cx('masm-btn', 'masm-v-' + variant, size !== 'md' && 'masm-btn--' + size, block && 'masm-btn--block', className)} {...rest}>
      {loading ? <span className="masm-spinner" style={{ width: 16, height: 16 }} aria-hidden="true" /> : icon && <Icon name={icon} size={isz} />}
      {children}
      {iconRight && !loading && <Icon name={iconRight} size={isz} />}
    </Tag>
  );
}
