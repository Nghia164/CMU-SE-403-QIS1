'use client';
import Link from 'next/link';
import { useCallback, useEffect, useRef, useState } from 'react';
import { IconButton } from './ds';
import { fetchNotices, markAllRead, markRead, streamNotices, type Notice } from '@/lib/notifications';

const ago = (iso: string) => {
  const s = Math.max(0, (Date.now() - new Date(iso).getTime()) / 1000);
  if (s < 60) return 'Vừa xong';
  if (s < 3600) return `${Math.floor(s / 60)} phút trước`;
  if (s < 86400) return `${Math.floor(s / 3600)} giờ trước`;
  return `${Math.floor(s / 86400)} ngày trước`;
};

/** Chuông thông báo: tải danh sách một lần rồi nhận thông báo mới qua SSE. */
export function NotificationBell() {
  const [items, setItems] = useState<Notice[]>([]);
  const [unread, setUnread] = useState(0);
  const [open, setOpen] = useState(false);
  const box = useRef<HTMLDivElement>(null);

  const load = useCallback(() => {
    fetchNotices()
      .then((r) => {
        setItems(r.items);
        setUnread(r.unread);
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    load();
    return streamNotices((n) => {
      setItems((cur) => (cur.some((x) => x.id === n.id) ? cur : [n, ...cur].slice(0, 30)));
      setUnread((u) => u + 1);
    });
  }, [load]);

  useEffect(() => {
    if (!open) return;
    const away = (e: MouseEvent) => !box.current?.contains(e.target as Node) && setOpen(false);
    const esc = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false);
    document.addEventListener('mousedown', away);
    document.addEventListener('keydown', esc);
    return () => {
      document.removeEventListener('mousedown', away);
      document.removeEventListener('keydown', esc);
    };
  }, [open]);

  const open1 = (n: Notice) => {
    setOpen(false);
    if (!n.read) {
      setItems((cur) => cur.map((x) => (x.id === n.id ? { ...x, read: true } : x)));
      setUnread((u) => Math.max(0, u - 1));
      markRead(n.id).catch(() => {});
    }
  };
  const readAll = () => {
    setItems((cur) => cur.map((x) => ({ ...x, read: true })));
    setUnread(0);
    markAllRead().catch(() => {});
  };

  return (
    <div className="notif" ref={box}>
      <IconButton icon="bell" label={unread ? `Thông báo, ${unread} chưa đọc` : 'Thông báo'} onClick={() => setOpen((o) => !o)} aria-expanded={open} aria-haspopup="true" />
      {unread > 0 && (
        <span className="notif-dot" aria-hidden="true">
          {unread > 9 ? '9+' : unread}
        </span>
      )}
      {open && (
        <div className="notif-panel" role="region" aria-label="Thông báo">
          <div className="notif-head">
            <strong>Thông báo</strong>
            {unread > 0 && (
              <button type="button" className="notif-link" onClick={readAll}>
                Đánh dấu đã đọc hết
              </button>
            )}
          </div>
          {items.length === 0 ? (
            <p className="muted small notif-empty">Bạn chưa có thông báo nào.</p>
          ) : (
            <ul className="notif-list">
              {items.map((n) => (
                <li key={n.id}>
                  <Link href={n.link || '/'} className={'notif-item' + (n.read ? '' : ' notif-item--new')} onClick={() => open1(n)}>
                    <span className="notif-title">{n.title}</span>
                    {n.body && <span className="muted small">{n.body}</span>}
                    <span className="muted small">{ago(n.created_at)}</span>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}