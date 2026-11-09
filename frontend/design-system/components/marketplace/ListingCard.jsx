import React from 'react';
import { Icon } from '../core/Icon.jsx';
import { IconButton } from '../core/IconButton.jsx';
import { Badge } from '../core/Badge.jsx';
import { Price } from './Price.jsx';
const COND_TONE = { 'Mới': 'primary', 'Như mới': 'primary' };
export function ListingCard({ title, price, original, condition, location, postedAt, image, placeholderIcon = 'image', href = '#', saved, onSave, sold, onClick, style }) {
  return (
    <article className="masm-card masm-card--interactive masm-lc" style={style} onClick={onClick}>
      <div className="masm-lc-media">
        {image ? <img src={image} alt="" /> : <Icon name={placeholderIcon} size={40} />}
        {sold && <div style={{ position: 'absolute', inset: 0, background: 'rgba(20,19,16,.45)', display: 'grid', placeItems: 'center' }}><Badge tone="solid">Đã bán</Badge></div>}
      </div>
      {onSave && <IconButton icon="heart" label={saved ? 'Bỏ lưu tin' : 'Lưu tin'} float pressed={!!saved} onClick={e => { e.stopPropagation(); onSave(); }} />}
      <div className="masm-lc-body">
        <Price value={price} original={original} size="sm" tone="primary" style={{ fontSize: 16 }} />
        <a className="masm-lc-title" href={href} onClick={e => { if (onClick) e.preventDefault(); }}>{title}</a>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 2, minWidth: 0 }}>
          {condition && <Badge size="sm" tone={COND_TONE[condition] || 'neutral'}>{condition}</Badge>}
          <span className="masm-lc-meta">{location && <><Icon name="map-pin" size={12} />{location}</>}{postedAt && <> · {postedAt}</>}</span>
        </div>
      </div>
    </article>
  );
}
