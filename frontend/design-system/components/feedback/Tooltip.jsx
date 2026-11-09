import React from 'react';
export function Tooltip({ content, children }) {
  const id = React.useId();
  return (
    <span className="masm-tip">
      {React.isValidElement(children) ? React.cloneElement(children, { 'aria-describedby': id }) : children}
      <span role="tooltip" id={id} className="masm-tip-bubble">{content}</span>
    </span>
  );
}
