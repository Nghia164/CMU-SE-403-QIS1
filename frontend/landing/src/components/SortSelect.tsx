'use client';
import { useRouter, useSearchParams } from 'next/navigation';
import { Select } from './ds';

const OPTIONS = [
  { value: 'newest', label: 'Mới nhất' },
  { value: 'price-asc', label: 'Giá thấp đến cao' },
  { value: 'price-desc', label: 'Giá cao đến thấp' },
];

export function SortSelect({ value }: { value: string }) {
  const router = useRouter();
  const params = useSearchParams();
  return (
    <Select
      aria-label="Sắp xếp"
      options={OPTIONS}
      value={value}
      onChange={(e) => {
        const next = new URLSearchParams(params.toString());
        if (e.target.value === 'newest') next.delete('sort');
        else next.set('sort', e.target.value);
        router.push('/' + (next.toString() ? `?${next}` : ''));
      }}
    />
  );
}
