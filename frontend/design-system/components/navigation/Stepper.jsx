import React from 'react';
import { Icon } from '../core/Icon.jsx';
export function Stepper({ steps = [], current = 0, label = 'Tiến trình' }) {
  return (
    <ol className="masm-stepper" aria-label={label}>
      {steps.map((s, i) => {
        const state = i < current ? 'done' : i === current ? 'current' : 'upcoming';
        return (
          <li key={i} className="masm-step" data-state={state} aria-current={state === 'current' ? 'step' : undefined}>
            <span className="masm-step-bar" />
            <span className="masm-step-label"><span className="masm-step-num">{state === 'done' ? <Icon name="check" size={12} /> : i + 1}</span><span className="masm-step-text">{s}</span></span>
          </li>
        );
      })}
    </ol>
  );
}
