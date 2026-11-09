import React from 'react';
import { Badge } from '../core/Badge.jsx';
export function AiBadge({ children = 'AI gợi ý', size = 'md' }) {
  return <Badge tone="ai" icon="sparkles" size={size}>{children}</Badge>;
}
