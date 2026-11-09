import type { Metadata, Viewport } from 'next';
import type { ReactNode } from 'react';
import '@ds/styles.css';
import './app.css';
import { Providers } from '@/components/Providers';
import { Footer, Header } from '@/components/Header';

export const metadata: Metadata = {
  title: { default: 'MASM · Đồ cũ, giá đúng, bán nhanh', template: '%s · MASM' },
  description:
    'Chợ đồ cũ có AI hỗ trợ: chụp ảnh, AI gợi ý giá và soạn tin đăng, bạn duyệt rồi đăng. Mua bán và thanh toán VNPay, MoMo ngay trên nền tảng.',
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  themeColor: '#157260',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="vi">
      <body>
        <Providers>
          <a href="#main" className="skip-link">
            Bỏ qua điều hướng
          </a>
          <div className="shell">
            <Header />
            <main id="main" className="shell-main">
              {children}
            </main>
            <Footer />
          </div>
        </Providers>
      </body>
    </html>
  );
}
