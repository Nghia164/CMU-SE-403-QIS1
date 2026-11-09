import type { Metadata } from 'next';
import { SellWizard } from '@/components/SellWizard';

export const metadata: Metadata = { title: 'Đăng tin bán' };

export default function SellPage() {
  return (
    <div className="sell-frame">
      <SellWizard />
    </div>
  );
}
