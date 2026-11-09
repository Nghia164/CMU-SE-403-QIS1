import React from 'react';
import { Icon } from '../core/Icon.jsx';
import { Button } from '../core/Button.jsx';
import { Badge } from '../core/Badge.jsx';
import { formatVND } from '../marketplace/Price.jsx';
import { ConfidenceBar } from './ConfidenceBar.jsx';
function Tile({ title, price, eta, active, onClick }) {
  return (
    <button type="button" onClick={onClick} aria-pressed={active}
      style={{ textAlign: 'left', padding: 12, borderRadius: 'var(--radius-md)', border: active ? '2px solid var(--ai-accent)' : '1px solid var(--ai-border)', background: active ? 'var(--surface-card)' : 'rgba(255,255,255,.5)', cursor: onClick ? 'pointer' : 'default', display: 'flex', flexDirection: 'column', gap: 4, font: 'inherit', color: 'inherit', minWidth: 0 }}>
      <span style={{ display: 'flex', alignItems: 'center', gap: 6, font: '600 12px/1.2 var(--font-sans)', color: 'var(--ai-text)' }}>{title}{active && <Icon name="circle-check" size={14} style={{ marginLeft: 'auto' }} />}</span>
      <span className="masm-num" style={{ font: 'var(--type-price)', color: 'var(--text-strong)' }}>{formatVND(price)}</span>
      {eta && <span style={{ font: '400 12px/1.3 var(--font-sans)', color: 'var(--text-muted)' }}>{eta}</span>}
    </button>
  );
}
export function AiPriceRange({ quick, optimal, quickEta = 'Dự kiến bán trong ~3 ngày', optimalEta = 'Dự kiến bán trong ~10 ngày', confidence, basis, sources = [], selected, onSelect, onEdit, onRegenerate, regenerating, maxSources = 4 }) {
  const [all, setAll] = React.useState(false);
  const shown = all ? sources : sources.slice(0, maxSources);
  const prices = sources.map(s => s.price).concat([quick, optimal]);
  const lo = Math.min(...prices), hi = Math.max(...prices), pos = p => hi === lo ? 50 : ((p - lo) / (hi - lo)) * 100;
  return (
    <section className="masm-ai" aria-label="Khoảng giá AI đề xuất" aria-busy={regenerating || undefined}>
      <div className="masm-ai-head">
        <span className="masm-ai-kicker"><Icon name="sparkles" size={16} />AI đề xuất giá</span>
        <span className="masm-ai-actions">
          {onEdit && <Button variant="quiet" size="sm" icon="pencil" onClick={onEdit}>Tự nhập giá</Button>}
          {onRegenerate && <Button variant="quiet" size="sm" icon="refresh-cw" onClick={onRegenerate} loading={regenerating}>Tạo lại</Button>}
        </span>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0,1fr) minmax(0,1fr)', gap: 8, opacity: regenerating ? .5 : 1 }}>
        <Tile title="Bán nhanh" price={quick} eta={quickEta} active={selected === 'quick'} onClick={onSelect && (() => onSelect('quick'))} />
        <Tile title="Giá tối ưu" price={optimal} eta={optimalEta} active={selected === 'optimal'} onClick={onSelect && (() => onSelect('optimal'))} />
      </div>
      {sources.length > 1 && (
        <div aria-hidden="true" style={{ padding: '6px 4px 0' }}>
          <div style={{ position: 'relative', height: 8, borderRadius: 999, background: 'var(--ai-bg-strong)' }}>
            <div style={{ position: 'absolute', top: 0, bottom: 0, left: pos(Math.min(quick, optimal)) + '%', width: Math.abs(pos(optimal) - pos(quick)) + '%', background: 'var(--ai-accent)', borderRadius: 999 }} />
            {sources.map((s, i) => <span key={i} style={{ position: 'absolute', top: 1, left: `calc(${pos(s.price)}% - 3px)`, width: 6, height: 6, borderRadius: 3, background: 'var(--surface-card)', border: '1px solid var(--ai-text)' }} />)}
          </div>
          <div className="masm-num" style={{ display: 'flex', justifyContent: 'space-between', font: '400 11px/1.6 var(--font-sans)', color: 'var(--text-muted)' }}><span>{formatVND(lo)}</span><span>{formatVND(hi)}</span></div>
        </div>
      )}
      {confidence != null && <ConfidenceBar value={confidence} note={basis} />}
      {sources.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 0, borderTop: '1px solid var(--ai-border)', paddingTop: 8 }}>
          <div style={{ font: '600 13px/1.3 var(--font-sans)', color: 'var(--text-strong)', padding: '2px 0 6px' }}>Nguồn so sánh ({sources.length})</div>
          <ul style={{ listStyle: 'none', margin: 0, padding: 0 }}>
            {shown.map((s, i) => (
              <li key={i} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '7px 0', borderTop: i ? '1px dashed var(--ai-border)' : 'none', font: 'var(--type-body-sm)' }}>
                <Badge size="sm">{s.source}</Badge>
                <span style={{ flex: 1, minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', color: 'var(--text-body)' }}>{s.url ? <a href={s.url} target="_blank" rel="noreferrer" style={{ color: 'inherit' }}>{s.title}</a> : s.title}{s.condition && <span style={{ color: 'var(--text-subtle)' }}> · {s.condition}</span>}</span>
                <span className="masm-num" style={{ fontWeight: 600, color: 'var(--text-strong)' }}>{formatVND(s.price)}</span>
              </li>
            ))}
          </ul>
          {sources.length > maxSources && <Button variant="quiet" size="sm" onClick={() => setAll(!all)} iconRight={all ? 'chevron-up' : 'chevron-down'} style={{ alignSelf: 'flex-start', color: 'var(--ai-text)', paddingLeft: 0 }}>{all ? 'Thu gọn' : `Xem thêm ${sources.length - maxSources} nguồn`}</Button>}
        </div>
      )}
      <div style={{ display: 'flex', gap: 6, font: '400 12px/1.4 var(--font-sans)', color: 'var(--ai-text)' }}><Icon name="info" size={14} style={{ marginTop: 1 }} />AI chỉ đề xuất. Bạn là người quyết định giá bán cuối cùng.</div>
    </section>
  );
}
