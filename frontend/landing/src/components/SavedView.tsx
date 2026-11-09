'use client';
import Link from 'next/link';
import { LISTINGS } from '@/lib/mock-data';
import { Button } from './ds';
import { ListingGrid } from './ListingGrid';
import { useSaved } from './Providers';

export function SavedView() {
  const { saved } = useSaved();
  const items = LISTINGS.filter((l) => saved.includes(l.id));
  return (
    <div className="container stack">
      <h1 className="page-title">Tin đã lưu</h1>
      {items.length ? (
        <ListingGrid items={items} />
      ) : (
        <div className="empty stack stack--sm">
          <p className="muted">Bạn chưa lưu tin nào. Bấm biểu tượng trái tim trên tin đăng để lưu lại xem sau.</p>
          <Button as={Link} href="/" variant="secondary" className="self-start">
            Xem tin mới
          </Button>
        </div>
      )}
    </div>
  );
}
