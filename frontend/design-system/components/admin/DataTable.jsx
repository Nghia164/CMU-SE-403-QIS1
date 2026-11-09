import React from 'react';
import { Checkbox } from '../forms/Checkbox.jsx';
export function DataTable({ columns = [], rows = [], rowKey = 'id', selectable, selected = [], onSelectedChange, dense, caption, empty = 'Không có dữ liệu' }) {
  const key = r => r[rowKey];
  const all = rows.length > 0 && rows.every(r => selected.includes(key(r)));
  const toggle = k => onSelectedChange && onSelectedChange(selected.includes(k) ? selected.filter(x => x !== k) : [...selected, k]);
  return (
    <div className="masm-table-wrap">
      <table className={'masm-table' + (dense ? ' masm-table--dense' : '')}>
        {caption && <caption className="masm-sr">{caption}</caption>}
        <thead><tr>
          {selectable && <th style={{ width: 44 }}><Checkbox aria-label="Chọn tất cả" checked={all} onChange={() => onSelectedChange && onSelectedChange(all ? [] : rows.map(key))} /></th>}
          {columns.map(c => <th key={c.key} style={{ width: c.width, textAlign: c.align || 'left' }} scope="col">{c.header}</th>)}
        </tr></thead>
        <tbody>
          {rows.length === 0 && <tr><td colSpan={columns.length + (selectable ? 1 : 0)} style={{ textAlign: 'center', padding: 32, color: 'var(--text-muted)' }}>{empty}</td></tr>}
          {rows.map(r => (
            <tr key={key(r)} aria-selected={selectable ? selected.includes(key(r)) : undefined}>
              {selectable && <td><Checkbox aria-label="Chọn dòng" checked={selected.includes(key(r))} onChange={() => toggle(key(r))} /></td>}
              {columns.map(c => <td key={c.key} style={{ textAlign: c.align || 'left' }}>{c.render ? c.render(r) : r[c.key]}</td>)}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
