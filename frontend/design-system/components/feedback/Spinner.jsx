import React from 'react';
export function Spinner({ size = 20, label = 'Đang tải', style }) {
  return <span role="status" aria-label={label} className="masm-spinner" style={{ width: size, height: size, ...style }} />;
}
