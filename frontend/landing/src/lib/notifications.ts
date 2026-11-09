import { authFetch, authJSON } from './session';

export interface Notice {
  id: string;
  kind: string;
  title: string;
  body: string;
  link: string;
  read: boolean;
  created_at: string;
}

export const fetchNotices = () => authJSON<{ items: Notice[]; unread: number }>('/notifications');
export const markRead = (id: string) => authJSON<void>(`/notifications/${id}/read`, { method: 'POST' });
export const markAllRead = () => authJSON<void>('/notifications/read-all', { method: 'POST' });

/**
 * Nghe luồng SSE bằng fetch (EventSource không gửi được header Authorization).
 * Tự nối lại sau khi đứt, chờ tăng dần tới 30 giây. Trả về hàm dừng.
 */
export function streamNotices(onNotice: (n: Notice) => void): () => void {
  const ctrl = new AbortController();
  let delay = 1000;
  (async () => {
    while (!ctrl.signal.aborted) {
      try {
        const res = await authFetch('/notifications/stream', { signal: ctrl.signal });
        if (!res.ok || !res.body) throw new Error(String(res.status));
        delay = 1000;
        const reader = res.body.pipeThrough(new TextDecoderStream()).getReader();
        let buf = '';
        for (;;) {
          const { value, done } = await reader.read();
          if (done) break;
          buf += value;
          let i: number;
          while ((i = buf.indexOf('\n\n')) >= 0) {
            const block = buf.slice(0, i);
            buf = buf.slice(i + 2);
            const data = block.split('\n').find((l) => l.startsWith('data: '));
            if (block.includes('event: notification') && data) {
              try {
                onNotice(JSON.parse(data.slice(6)) as Notice);
              } catch {
                /* bỏ qua gói lỗi */
              }
            }
          }
        }
      } catch {
        if (ctrl.signal.aborted) return;
      }
      await new Promise((r) => setTimeout(r, delay));
      delay = Math.min(delay * 2, 30000);
    }
  })();
  return () => ctrl.abort();
}