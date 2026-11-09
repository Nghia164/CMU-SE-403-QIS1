'use client';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { Suspense, useEffect, useState, type FormEvent } from 'react';
import { Button, Icon, IconButton, Input } from './ds';
import { NotificationBell } from './NotificationBell';
import { useAuth, useSaved, useTheme } from './Providers';
import { hasApi } from '@/lib/session';

function SearchBox() {
  const router = useRouter();
  const params = useSearchParams();
  const [q, setQ] = useState(params.get('q') ?? '');
  useEffect(() => setQ(params.get('q') ?? ''), [params]);
  const submit = (e: FormEvent) => {
    e.preventDefault();
    const next = new URLSearchParams();
    if (q.trim()) next.set('q', q.trim());
    router.push('/' + (next.toString() ? `?${next}` : ''));
  };
  return (
    <form onSubmit={submit} role="search" className="hdr-search">
      <Input icon="search" placeholder="Tìm xe đạp, máy ảnh, sách…" aria-label="Tìm kiếm" value={q} onChange={(e) => setQ(e.target.value)} />
    </form>
  );
}

export function Header() {
  const { user, logout } = useAuth();
  const { saved } = useSaved();
  const { theme, toggle } = useTheme();
  const initial = user?.name?.trim()?.[0]?.toUpperCase();

  return (
    <header className="hdr">
      <div className="container hdr-inner">
        <Link href="/" className="brand" aria-label="MASM, về trang chủ">
          MASM
        </Link>
        <Suspense fallback={<div className="hdr-search" />}>
          <SearchBox />
        </Suspense>
        <nav className="hdr-nav" aria-label="Điều hướng chính">
          <Button variant="quiet" icon="heart" as={Link} href="/saved" className="hdr-hide-sm">
            Đã lưu{saved.length ? ` (${saved.length})` : ''}
          </Button>
          <Button variant="quiet" icon="package" as={Link} href="/orders" className="hdr-hide-sm">
            Đơn hàng
          </Button>
          <IconButton icon={theme === 'dark' ? 'sun' : 'moon'} label={theme === 'dark' ? 'Chuyển sang giao diện sáng' : 'Chuyển sang giao diện tối'} onClick={toggle} />
          {user ? (
            <>
              {hasApi && <NotificationBell />}
              <span className="avatar" title={user.name} aria-label={`Tài khoản ${user.name}`}>
                {initial}
              </span>
              <IconButton icon="log-out" label="Đăng xuất" onClick={logout} />
            </>
          ) : (
            <Button variant="secondary" as={Link} href="/login">
              Đăng nhập
            </Button>
          )}
          <Button icon="camera" as={Link} href="/sell">
            <span className="hdr-hide-xs">Đăng tin</span>
            <span className="sr-only hdr-show-xs">Đăng tin</span>
          </Button>
        </nav>
      </div>
    </header>
  );
}

export function Footer() {
  return (
    <footer className="ftr">
      <div className="container ftr-inner">
        <span className="brand brand--sm">MASM</span>
        <span>Chợ đồ cũ có AI hỗ trợ · Đồ án CMU-SE 403</span>
        <span className="ftr-spacer" />
        <Link href="/">Điều khoản</Link>
        <Link href="/">Chính sách bảo mật</Link>
        <Link href="/">Trung tâm trợ giúp</Link>
        <Icon name="shield-check" size={16} label="Bảo vệ người mua" />
      </div>
    </footer>
  );
}
