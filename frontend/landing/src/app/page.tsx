import Link from 'next/link';
import { Suspense } from 'react';
import { listCategories, listListings, type SortKey } from '@/lib/api';
import { Button } from '@/components/ds';
import { ListingGrid } from '@/components/ListingGrid';
import { SortSelect } from '@/components/SortSelect';

interface HomeProps {
  searchParams: { q?: string; cat?: string; sort?: string };
}

export default async function HomePage({ searchParams }: HomeProps) {
  const cat = searchParams.cat ?? 'all';
  const q = searchParams.q ?? '';
  const sort = (['price-asc', 'price-desc'].includes(searchParams.sort ?? '') ? searchParams.sort : 'newest') as SortKey;
  const [categories, listings] = await Promise.all([listCategories(), listListings({ q, cat, sort })]);

  const hrefFor = (value: string) => {
    const p = new URLSearchParams();
    if (q) p.set('q', q);
    if (value !== 'all') p.set('cat', value);
    if (sort !== 'newest') p.set('sort', sort);
    return '/' + (p.toString() ? `?${p}` : '');
  };

  return (
    <div className="container stack">
      <section className="hero" aria-labelledby="hero-title">
        <div className="hero-text">
          <h1 id="hero-title">Đồ cũ, giá đúng, bán nhanh</h1>
          <p>Chụp vài tấm ảnh. AI nhận diện sản phẩm, gợi ý giá và soạn sẵn tin đăng. Bạn chỉ cần duyệt và bấm đăng.</p>
        </div>
        <Button size="lg" icon="camera" as={Link} href="/sell">
          Đăng tin ngay
        </Button>
      </section>

      <nav className="chips" aria-label="Danh mục">
        {categories.map((c) => (
          <Button
            key={c.value}
            size="sm"
            icon={c.icon}
            variant={cat === c.value ? 'primary' : 'secondary'}
            as={Link}
            href={hrefFor(c.value)}
            aria-current={cat === c.value ? 'true' : undefined}
            className="chip"
          >
            {c.label}
          </Button>
        ))}
      </nav>

      <div className="list-head">
        <h2>
          {q ? `Kết quả cho “${q}”` : 'Tin mới gần bạn'} <span className="muted small">· {listings.length} tin</span>
        </h2>
        <div className="sort">
          <Suspense fallback={null}>
            <SortSelect value={sort} />
          </Suspense>
        </div>
      </div>

      <ListingGrid items={listings} />
    </div>
  );
}
