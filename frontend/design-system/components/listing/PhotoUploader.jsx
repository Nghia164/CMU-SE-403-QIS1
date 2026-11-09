import React from 'react';
import { Icon } from '../core/Icon.jsx';
import { IconButton } from '../core/IconButton.jsx';
import { Badge } from '../core/Badge.jsx';
export function PhotoUploader({ photos = [], max = 8, onAdd, onRemove, hint = 'Tối đa 8 ảnh · JPG, PNG ≤ 10MB · Ảnh rõ, đủ sáng giúp AI nhận diện chính xác hơn' }) {
  const input = React.useRef(null);
  const [over, setOver] = React.useState(false);
  const full = photos.length >= max;
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {!full && (
        <button type="button" className="masm-drop" data-over={over} onClick={() => input.current && input.current.click()}
          onDragOver={e => { e.preventDefault(); setOver(true); }} onDragLeave={() => setOver(false)}
          onDrop={e => { e.preventDefault(); setOver(false); onAdd && onAdd(Array.from(e.dataTransfer.files)); }}>
          <span style={{ width: 48, height: 48, borderRadius: '50%', background: 'var(--primary-soft)', color: 'var(--primary-text)', display: 'grid', placeItems: 'center' }}><Icon name="camera" size={24} /></span>
          <span style={{ font: 'var(--type-title)', color: 'var(--text-strong)' }}>Chụp hoặc tải ảnh lên</span>
          <span style={{ font: '400 13px/1.45 var(--font-sans)', color: 'var(--text-muted)', maxWidth: 320 }}>{hint}</span>
        </button>
      )}
      <input ref={input} type="file" accept="image/*" multiple capture="environment" hidden onChange={e => { onAdd && onAdd(Array.from(e.target.files)); e.target.value = ''; }} />
      {photos.length > 0 && (
        <div>
          <div style={{ font: '500 13px/1.3 var(--font-sans)', color: 'var(--text-muted)', marginBottom: 8 }} className="masm-num">{photos.length}/{max} ảnh</div>
          <ul className="masm-thumbs" style={{ listStyle: 'none', margin: 0, padding: 0 }}>
            {photos.map((p, i) => (
              <li key={p.id || i} className="masm-thumb">
                {p.src ? <img src={p.src} alt={p.alt || 'Ảnh ' + (i + 1)} /> : <Icon name={p.icon || 'image'} size={28} />}
                {i === 0 && <span className="masm-thumb-cover"><Badge tone="solid" size="sm">Ảnh bìa</Badge></span>}
                {onRemove && <IconButton icon="x" label={'Xoá ảnh ' + (i + 1)} float onClick={() => onRemove(p.id || i)} />}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
