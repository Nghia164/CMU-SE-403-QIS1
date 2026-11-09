/**
 * Lớp truy cập dữ liệu của web.
 * Hiện dùng dữ liệu mock. Khi Backend (Go, Gin) sẵn sàng, chỉ cần thay phần cài đặt
 * trong file này bằng gọi REST tới NEXT_PUBLIC_API_URL; các trang không phải sửa.
 */
import { CATEGORIES, LISTINGS } from './mock-data';
import type { Category, Listing } from './types';

export type SortKey = 'newest' | 'price-asc' | 'price-desc';

export interface ListingQuery {
  q?: string;
  cat?: string;
  sort?: SortKey;
}

const strip = (s: string) =>
  s.normalize('NFD').replace(/[̀-ͯ]/g, '').replace(/đ/g, 'd').toLowerCase();

export async function listListings({ q, cat, sort }: ListingQuery = {}): Promise<Listing[]> {
  let out = LISTINGS.filter((l) => !cat || cat === 'all' || l.cat === cat);
  if (q && q.trim()) {
    const needle = strip(q.trim());
    out = out.filter((l) => strip(l.title + ' ' + l.desc).includes(needle));
  }
  if (sort === 'price-asc') out = [...out].sort((a, b) => a.price - b.price);
  if (sort === 'price-desc') out = [...out].sort((a, b) => b.price - a.price);
  return out;
}

export async function getListing(id: string): Promise<Listing | undefined> {
  return LISTINGS.find((l) => l.id === id);
}

export async function listCategories(): Promise<Category[]> {
  return CATEGORIES;
}

export const categoryLabel = (value: string) => CATEGORIES.find((c) => c.value === value)?.label ?? '';
