import type { Metadata } from 'next';
import { SavedView } from '@/components/SavedView';

export const metadata: Metadata = { title: 'Tin đã lưu' };

export default function SavedPage() {
  return <SavedView />;
}
