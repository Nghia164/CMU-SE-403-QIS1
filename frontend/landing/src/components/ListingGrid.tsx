'use client';
import { useRouter } from 'next/navigation';
import type { Listing } from '@/lib/types';
import { ListingCard } from './ds';
import { useSaved } from './Providers';

export function ListingGrid({ items, empty }: { items: Listing[]; empty?: string }) {
  const router = useRouter();
  const { saved, toggle } = useSaved();
  if (!items.length) {
    return <p className="muted empty">{empty ?? 'Chưa có tin phù hợp. Thử đổi từ khoá hoặc danh mục khác.'}</p>;
  }
  return (
    <div className="grid-listings">
      {items.map((l) => (
        <ListingCard
          key={l.id}
          title={l.title}
          price={l.price}
          original={l.original}
          condition={l.condition}
          location={l.location}
          postedAt={l.postedAt}
          placeholderIcon={l.icon}
          sold={l.sold}
          href={`/listing/${l.id}`}
          saved={saved.includes(l.id)}
          onSave={() => toggle(l.id)}
          onClick={() => router.push(`/listing/${l.id}`)}
        />
      ))}
    </div>
  );
}
