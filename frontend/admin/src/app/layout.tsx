import type { Metadata, Viewport } from 'next';
import type { ReactNode } from 'react';
import '@ds/styles.css';
import './admin.css';

export const metadata: Metadata = {
  title: { default: 'MASM Admin', template: '%s · MASM Admin' },
  description: 'Quản trị MASM: tài khoản, kiểm duyệt tin, báo cáo vi phạm, đơn hàng.',
};
export const viewport: Viewport = { width: 'device-width', initialScale: 1, themeColor: '#157260' };

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="vi">
      <body>{children}</body>
    </html>
  );
}
