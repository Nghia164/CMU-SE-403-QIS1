'use client';
import { useEffect, useMemo, useState, type ReactNode } from 'react';
import { Badge, Button, Dialog, DataTable, Icon, IconButton, Input, Radio, StatCard, Toast } from './ds';
import {
  MOD, ORDERS, ORDER_STATUS, REPORTS, REPORT_STATUS, USERS, USER_STATUS, vnd,
  type AMod, type AOrder, type AReport, type AUser,
} from '@/lib/data';

type Section = 'overview' | 'acc' | 'mod' | 'rep' | 'orders' | 'settings';
interface Note { tone: 'success' | 'danger' | 'info'; title: string; message?: string }

const norm = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').replace(/đ/g, 'd').toLowerCase();

export function AdminApp() {
  const [section, setSection] = useState<Section>('overview');
  const [theme, setTheme] = useState<'light' | 'dark'>('light');
  const [menu, setMenu] = useState(false);
  const [query, setQuery] = useState('');
  const [users, setUsers] = useState<AUser[]>(USERS);
  const [mod, setMod] = useState<AMod[]>(MOD);
  const [reports, setReports] = useState<AReport[]>(REPORTS);
  const [note, setNote] = useState<Note | null>(null);

  useEffect(() => {
    try {
      const s = localStorage.getItem('masm-admin-theme');
      setTheme(s === 'dark' || s === 'light' ? s : window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
    } catch { /* bỏ qua */ }
  }, []);
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    try { localStorage.setItem('masm-admin-theme', theme); } catch { /* bỏ qua */ }
  }, [theme]);
  useEffect(() => {
    if (!note) return;
    const t = setTimeout(() => setNote(null), 4000);
    return () => clearTimeout(t);
  }, [note]);

  const openReports = reports.filter((r) => r.status !== 'done').length;
  const go = (s: Section) => { setSection(s); setMenu(false); };
  const items: [Section, string, string, number?][] = [
    ['overview', 'Tổng quan', 'layout-dashboard'],
    ['acc', 'Tài khoản', 'users'],
    ['mod', 'Kiểm duyệt tin', 'shield-check', mod.length],
    ['rep', 'Báo cáo vi phạm', 'flag', openReports],
    ['orders', 'Đơn hàng', 'receipt'],
    ['settings', 'Cài đặt', 'settings'],
  ];
  const titles: Record<Section, string> = { overview: 'Tổng quan', acc: 'Tài khoản', mod: 'Kiểm duyệt tin', rep: 'Báo cáo vi phạm', orders: 'Đơn hàng', settings: 'Cài đặt' };

  return (
    <div className="adm">
      <a href="#main" className="skip-link">Bỏ qua điều hướng</a>
      {menu && <div className="scrim" onClick={() => setMenu(false)} aria-hidden="true" />}
      <aside className={'side' + (menu ? ' open' : '')} aria-label="Điều hướng quản trị">
        <div className="side-brand"><b>MASM</b><span>Admin</span></div>
        <nav style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          {items.map(([k, label, icon, n]) => (
            <button key={k} type="button" className="nav-item" aria-current={section === k ? 'page' : undefined} onClick={() => go(k)}>
              <Icon name={icon} size={18} />
              <span className="grow">{label}</span>
              {n ? <span className="nav-count">{n}</span> : null}
            </button>
          ))}
        </nav>
        <div className="side-foot">Đồ án CMU-SE 403</div>
      </aside>

      <div className="main-col">
        <header className="top">
          <IconButton className="menu-btn" icon="menu" label="Mở menu" onClick={() => setMenu(true)} />
          <div className="top-search">
            <Input icon="search" placeholder="Tìm người dùng, tin đăng, mã báo cáo…" aria-label="Tìm kiếm" value={query} onChange={(e: { target: { value: string } }) => setQuery(e.target.value)} />
          </div>
          <div className="top-right">
            <IconButton icon={theme === 'dark' ? 'sun' : 'moon'} label={theme === 'dark' ? 'Chế độ sáng' : 'Chế độ tối'} onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')} />
            <IconButton icon="bell" label="Thông báo" />
            <div className="me"><i>QT</i><span>Quản trị viên</span></div>
          </div>
        </header>

        <main id="main" className="content">
          <h1>{titles[section]}</h1>
          {section === 'overview' && <Overview users={users} mod={mod} reports={reports} go={go} />}
          {section === 'acc' && <Accounts users={users} setUsers={setUsers} query={query} notify={setNote} />}
          {section === 'mod' && <Moderation rows={mod} setRows={setMod} query={query} notify={setNote} />}
          {section === 'rep' && <Reports rows={reports} setRows={setReports} query={query} notify={setNote} />}
          {section === 'orders' && <Orders query={query} />}
          {section === 'settings' && <Settings notify={setNote} />}
        </main>
      </div>

      {note && (
        <div className="toast-slot">
          <Toast tone={note.tone} title={note.title} message={note.message} onClose={() => setNote(null)} />
        </div>
      )}
    </div>
  );
}

function Wrap({ children }: { children: ReactNode }) {
  return <div className="table-wrap">{children}</div>;
}

/* ---------- Tổng quan ---------- */
function Overview({ users, mod, reports, go }: { users: AUser[]; mod: AMod[]; reports: AReport[]; go: (s: Section) => void }) {
  const flagged = mod.filter((m) => m.ai[0] !== 'Hợp lệ').length;
  const open = reports.filter((r) => r.status !== 'done').length;
  return (
    <>
      <div className="kpis">
        <StatCard icon="users" label="Tài khoản" value={String(users.length)} delta="+2" deltaTone="success" hint="trong 7 ngày" />
        <StatCard icon="shield-check" label="Tin chờ duyệt" value={String(mod.length)} delta={`${flagged} AI cảnh báo`} deltaTone={flagged ? 'danger' : 'neutral'} />
        <StatCard icon="flag" label="Báo cáo chờ xử lý" value={String(open)} delta="+1" deltaTone="danger" hint="so với hôm qua" />
        <StatCard icon="receipt" label="Doanh thu giao dịch" value={vnd(ORDERS.reduce((s, o) => s + o.amount, 0))} delta="+12%" deltaTone="success" hint="tuần này" />
      </div>
      <div className="two">
        <div className="stack">
          <div className="bar"><b style={{ color: 'var(--text-strong)' }}>Cần duyệt gấp</b><Button size="sm" variant="quiet" onClick={() => go('mod')}>Xem tất cả</Button></div>
          <Wrap>
            <DataTable caption="Tin AI cảnh báo" rows={mod.filter((m) => m.ai[0] !== 'Hợp lệ')} empty="Không có tin nào bị AI cảnh báo." columns={[
              { key: 'title', header: 'Tin đăng', render: (r: AMod) => <span className="cell-main">{r.title}</span> },
              { key: 'ai', header: 'AI', render: (r: AMod) => <Badge size="sm" tone="ai" icon="sparkles">{r.ai[0]} · {r.ai[1]}%</Badge> },
            ]} />
          </Wrap>
        </div>
        <div className="stack">
          <div className="bar"><b style={{ color: 'var(--text-strong)' }}>Báo cáo mới</b><Button size="sm" variant="quiet" onClick={() => go('rep')}>Xem tất cả</Button></div>
          <Wrap>
            <DataTable caption="Báo cáo mới" rows={reports.filter((r) => r.status === 'new')} empty="Không có báo cáo mới." columns={[
              { key: 'target', header: 'Đối tượng', render: (r: AReport) => <span className="cell-main">{r.target}</span> },
              { key: 'reason', header: 'Lý do' },
            ]} />
          </Wrap>
        </div>
      </div>
    </>
  );
}

/* ---------- Tài khoản ---------- */
function Accounts({ users, setUsers, query, notify }: { users: AUser[]; setUsers: (f: (u: AUser[]) => AUser[]) => void; query: string; notify: (n: Note) => void }) {
  const [confirm, setConfirm] = useState<AUser | null>(null);
  const rows = useMemo(() => users.filter((u) => !query || norm(u.name + ' ' + u.email).includes(norm(query))), [users, query]);
  const flip = (u: AUser) => {
    const to = u.status === 'locked' ? 'active' : 'locked';
    setUsers((us) => us.map((x) => (x.id === u.id ? { ...x, status: to } : x)));
    notify({ tone: 'success', title: to === 'locked' ? 'Đã khoá tài khoản' : 'Đã mở khoá tài khoản', message: u.name });
  };
  return (
    <>
      <Wrap>
        <DataTable caption="Danh sách tài khoản" rows={rows} empty="Không tìm thấy tài khoản phù hợp." columns={[
          { key: 'name', header: 'Người dùng', render: (r: AUser) => <div><div className="cell-main">{r.name}</div><div className="cell-sub">{r.email}</div></div> },
          { key: 'role', header: 'Vai trò' },
          { key: 'joined', header: 'Ngày tham gia' },
          { key: 'listings', header: 'Tin đăng', align: 'right' },
          { key: 'status', header: 'Trạng thái', render: (r: AUser) => <Badge size="sm" tone={USER_STATUS[r.status][1]}>{USER_STATUS[r.status][0]}</Badge> },
          { key: 'a', header: '', align: 'right', render: (r: AUser) => (
            <div className="cell-row">
              {r.status === 'locked'
                ? <Button size="sm" variant="secondary" icon="lock-open" onClick={() => flip(r)}>Mở khoá</Button>
                : <Button size="sm" variant="secondary" icon="lock" onClick={() => setConfirm(r)}>Khoá</Button>}
            </div>
          ) },
        ]} />
      </Wrap>
      <Dialog open={!!confirm} onClose={() => setConfirm(null)} title="Khoá tài khoản?" footer={<>
        <Button variant="secondary" onClick={() => setConfirm(null)}>Huỷ</Button>
        <Button variant="danger" icon="lock" onClick={() => { if (confirm) flip(confirm); setConfirm(null); }}>Khoá tài khoản</Button>
      </>}>
        {confirm && <>Tài khoản <b>{confirm.name}</b> sẽ không thể đăng nhập, đăng tin hay đặt mua cho đến khi được mở khoá. Các tin đang hiển thị sẽ bị ẩn.</>}
      </Dialog>
    </>
  );
}

/* ---------- Kiểm duyệt ---------- */
function Moderation({ rows, setRows, query, notify }: { rows: AMod[]; setRows: (f: (r: AMod[]) => AMod[]) => void; query: string; notify: (n: Note) => void }) {
  const [sel, setSel] = useState<(string | number)[]>([]);
  const shown = useMemo(() => rows.filter((r) => !query || norm(r.title + ' ' + r.seller).includes(norm(query))), [rows, query]);
  const resolve = (ids: (string | number)[], ok: boolean) => {
    setRows((rs) => rs.filter((r) => !ids.includes(r.id)));
    setSel([]);
    notify({ tone: ok ? 'success' : 'info', title: ok ? `Đã duyệt ${ids.length} tin` : `Đã từ chối ${ids.length} tin`, message: ok ? 'Tin được hiển thị cho người mua.' : 'Người bán sẽ nhận thông báo kèm lý do.' });
  };
  return (
    <div className="stack">
      <div className="bar">
        <span>{sel.length ? `Đã chọn ${sel.length} tin` : `${rows.length} tin chờ duyệt · AI sàng lọc trước, quản trị viên quyết định cuối cùng`}</span>
        {sel.length > 0 && (
          <>
            <Button size="sm" icon="check" onClick={() => resolve(sel, true)}>Duyệt đã chọn</Button>
            <Button size="sm" variant="secondary" icon="x" onClick={() => resolve(sel, false)}>Từ chối</Button>
          </>
        )}
      </div>
      <Wrap>
        <DataTable caption="Hàng đợi kiểm duyệt" selectable selected={sel} onSelectedChange={setSel} rows={shown} empty="Không còn tin nào chờ duyệt." columns={[
          { key: 'title', header: 'Tin đăng', render: (r: AMod) => (
            <div className="thumb-cell">
              <div className="masm-thumb" style={{ width: 44, flex: 'none' }}><Icon name={r.icon} size={20} /></div>
              <div><div className="cell-main">{r.title}</div><div className="cell-sub">{r.seller} · {r.at}</div></div>
            </div>
          ) },
          { key: 'price', header: 'Giá', align: 'right', render: (r: AMod) => <span className="masm-num" style={{ fontWeight: 600 }}>{vnd(r.price)}</span> },
          { key: 'ai', header: 'AI sàng lọc', render: (r: AMod) => <Badge size="sm" tone="ai" icon="sparkles">{r.ai[0]} · {r.ai[1]}%</Badge> },
          { key: 'a', header: '', align: 'right', render: (r: AMod) => (
            <div className="cell-row">
              <Button size="sm" variant="secondary" onClick={() => resolve([r.id], false)}>Từ chối</Button>
              <Button size="sm" onClick={() => resolve([r.id], true)}>Duyệt</Button>
            </div>
          ) },
        ]} />
      </Wrap>
    </div>
  );
}

/* ---------- Báo cáo ---------- */
function Reports({ rows, setRows, query, notify }: { rows: AReport[]; setRows: (f: (r: AReport[]) => AReport[]) => void; query: string; notify: (n: Note) => void }) {
  const [open, setOpen] = useState<AReport | null>(null);
  const [act, setAct] = useState('remove');
  const [memo, setMemo] = useState('');
  const shown = useMemo(() => rows.filter((r) => !query || norm(r.id + ' ' + r.target + ' ' + r.reason).includes(norm(query))), [rows, query]);
  const actions: [string, string, string][] = [
    ['remove', 'Gỡ tin đăng', 'Ẩn tin và thông báo cho người bán'],
    ['warn', 'Cảnh cáo người bán', 'Ghi nhận vi phạm, chưa gỡ tin'],
    ['lock', 'Khoá tài khoản', 'Áp dụng cho vi phạm nghiêm trọng hoặc tái phạm'],
    ['ignore', 'Bỏ qua', 'Báo cáo không hợp lệ'],
  ];
  const confirm = () => {
    if (!open) return;
    setRows((rs) => rs.map((x) => (x.id === open.id ? { ...x, status: 'done' } : x)));
    notify({ tone: 'success', title: `Đã xử lý ${open.id}`, message: actions.find((a) => a[0] === act)?.[1] });
    setOpen(null);
    setMemo('');
  };
  return (
    <>
      <Wrap>
        <DataTable caption="Báo cáo vi phạm" rows={shown} empty="Không có báo cáo phù hợp." columns={[
          { key: 'id', header: 'Mã', render: (r: AReport) => <span style={{ font: 'var(--type-mono)', fontSize: 12 }}>{r.id}</span> },
          { key: 'target', header: 'Đối tượng', render: (r: AReport) => <div><div className="cell-main">{r.target}</div><div className="cell-sub">{r.kind}</div></div> },
          { key: 'reason', header: 'Lý do' },
          { key: 'reporter', header: 'Người báo cáo' },
          { key: 'count', header: 'Số báo cáo', align: 'right', render: (r: AReport) => <span className="masm-num" style={{ fontWeight: r.count > 2 ? 700 : 400, color: r.count > 2 ? 'var(--danger-text)' : 'inherit' }}>{r.count}</span> },
          { key: 'status', header: 'Trạng thái', render: (r: AReport) => <Badge size="sm" tone={REPORT_STATUS[r.status][1]}>{REPORT_STATUS[r.status][0]}</Badge> },
          { key: 'a', header: '', align: 'right', render: (r: AReport) => r.status === 'done'
            ? <span className="cell-sub">Đã đóng</span>
            : <Button size="sm" onClick={() => { setOpen(r); setAct('remove'); }}>Xử lý</Button> },
        ]} />
      </Wrap>
      <Dialog open={!!open} onClose={() => setOpen(null)} title={open ? `Xử lý báo cáo ${open.id}` : ''} footer={<>
        <Button variant="secondary" onClick={() => setOpen(null)}>Huỷ</Button>
        <Button variant={act === 'ignore' ? 'primary' : 'danger'} onClick={confirm}>Xác nhận xử lý</Button>
      </>}>
        {open && (
          <div className="stack">
            <div><b style={{ color: 'var(--text-strong)' }}>{open.target}</b> — {open.reason} ({open.count} báo cáo)</div>
            <div role="radiogroup" aria-label="Hình thức xử lý" className="stack">
              {actions.map(([v, l, d]) => <Radio key={v} name="act" value={v} label={l} description={d} checked={act === v} onChange={() => setAct(v)} />)}
            </div>
            <Input label="Ghi chú nội bộ" multiline rows={2} placeholder="Lý do quyết định (chỉ quản trị viên thấy)" value={memo} onChange={(e: { target: { value: string } }) => setMemo(e.target.value)} />
          </div>
        )}
      </Dialog>
    </>
  );
}

/* ---------- Đơn hàng ---------- */
function Orders({ query }: { query: string }) {
  const rows = ORDERS.filter((o) => !query || norm(o.id + ' ' + o.item + ' ' + o.buyer + ' ' + o.seller).includes(norm(query)));
  return (
    <Wrap>
      <DataTable caption="Đơn hàng" rows={rows} empty="Không có đơn hàng phù hợp." columns={[
        { key: 'id', header: 'Mã đơn', render: (r: AOrder) => <span style={{ font: 'var(--type-mono)', fontSize: 12 }}>{r.id}</span> },
        { key: 'item', header: 'Sản phẩm', render: (r: AOrder) => <div><div className="cell-main">{r.item}</div><div className="cell-sub">{r.seller} → {r.buyer}</div></div> },
        { key: 'method', header: 'Thanh toán' },
        { key: 'amount', header: 'Số tiền', align: 'right', render: (r: AOrder) => <span className="masm-num" style={{ fontWeight: 600 }}>{vnd(r.amount)}</span> },
        { key: 'status', header: 'Trạng thái', render: (r: AOrder) => <Badge size="sm" tone={ORDER_STATUS[r.status][1]}>{ORDER_STATUS[r.status][0]}</Badge> },
      ]} />
    </Wrap>
  );
}

/* ---------- Cài đặt ---------- */
function Settings({ notify }: { notify: (n: Note) => void }) {
  const [thr, setThr] = useState('70');
  const [err, setErr] = useState('');
  const save = () => {
    const n = Number(thr);
    if (!Number.isInteger(n) || n < 0 || n > 100) { setErr('Nhập số nguyên từ 0 đến 100'); return; }
    setErr('');
    notify({ tone: 'success', title: 'Đã lưu cài đặt', message: `Ngưỡng cảnh báo AI: ${n}%` });
  };
  return (
    <div className="stack" style={{ maxWidth: 480 }}>
      <Input label="Ngưỡng AI tự chuyển tin sang kiểm duyệt thủ công (%)" value={thr} onChange={(e: { target: { value: string } }) => setThr(e.target.value)} error={err} hint="Tin có độ tin cậy dưới ngưỡng sẽ vào hàng đợi duyệt." inputMode="numeric" />
      <Button className="self-start" onClick={save}>Lưu cài đặt</Button>
      <p className="cell-sub">Bản demo: cài đặt chưa được lưu lên máy chủ.</p>
    </div>
  );
}
