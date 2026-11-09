'use client';
import Link from 'next/link';
import { useState } from 'react';
import type { Listing } from '@/lib/types';
import { Badge, Button, Card, Icon, IconButton, Price } from './ds';
import { useSaved } from './Providers';

export function ListingDetail({ listing: l, categoryLabel }: { listing: Listing; categoryLabel: string }) {
  const { saved, toggle } = useSaved();
  const [img, setImg] = useState(0);
  const [reported, setReported] = useState(false);

  return (
    <div className="container stack">
      <nav aria-label="Đường dẫn" className="crumbs">
        <Link href="/">Trang chủ</Link>
        <Icon name="chevron-right" size={14} />
        <Link href={`/?cat=${l.cat}`}>{categoryLabel}</Link>
        <Icon name="chevron-right" size={14} />
        <span aria-current="page">{l.title}</span>
      </nav>

      <div className="detail-grid">
        <div className="stack stack--sm">
          <div className="detail-photo">
            <Icon name={l.icon} size={96} />
            <span className="detail-count">
              <Badge tone="solid" size="sm">
                {img + 1}/5
              </Badge>
            </span>
          </div>
          <div className="thumb-row">
            {[0, 1, 2, 3, 4].map((i) => (
              <button
                key={i}
                type="button"
                className="masm-thumb thumb-btn"
                aria-label={`Ảnh ${i + 1}`}
                aria-pressed={img === i}
                onClick={() => setImg(i)}
              >
                <Icon name={i ? 'image' : l.icon} size={22} />
              </button>
            ))}
          </div>
          <Card>
            <h2 className="card-title">Mô tả</h2>
            <p className="desc">{l.desc}</p>
          </Card>
        </div>

        <aside className="detail-side stack">
          <div className="row row--wrap">
            <Badge tone="primary">{l.condition}</Badge>
            <span className="meta">
              <Icon name="map-pin" size={14} />
              {l.location} · {l.postedAt}
            </span>
          </div>
          <h1 className="detail-title">{l.title}</h1>
          <Price value={l.price} original={l.original} size="lg" />

          <div className="actions">
            <Button size="lg" block disabled={l.sold} as={l.sold ? undefined : Link} href={l.sold ? undefined : `/checkout/${l.id}`}>
              {l.sold ? 'Đã bán' : 'Mua ngay'}
            </Button>
            <Button size="lg" variant="secondary" icon="message-circle" className="hide-sm">
              Chat
            </Button>
            <IconButton
              icon="heart"
              label={saved.includes(l.id) ? 'Bỏ lưu tin' : 'Lưu tin'}
              variant="secondary"
              pressed={saved.includes(l.id)}
              onClick={() => toggle(l.id)}
            />
          </div>

          <div className="notice">
            <Icon name="shield-check" size={18} />
            Thanh toán qua MASM: tiền được giữ an toàn đến khi bạn xác nhận đã nhận hàng.
          </div>

          <Card className="seller">
            <span className="avatar avatar--lg" aria-hidden="true">
              {l.seller[0]}
            </span>
            <div className="seller-info">
              <div className="seller-name">{l.seller}</div>
              <div className="meta">
                <Icon name="star" size={14} style={{ color: 'var(--teal-500)' }} />
                {l.rating} · {l.sales} sản phẩm đã bán
              </div>
            </div>
            <Button variant="ghost" size="sm">
              Xem trang
            </Button>
          </Card>

          {reported ? (
            <p className="muted small" role="status">
              Cảm ơn bạn. Chúng tôi sẽ xem xét tin đăng này.
            </p>
          ) : (
            <Button variant="quiet" size="sm" icon="flag" onClick={() => setReported(true)} className="self-start">
              Báo cáo tin đăng
            </Button>
          )}
        </aside>
      </div>
    </div>
  );
}
