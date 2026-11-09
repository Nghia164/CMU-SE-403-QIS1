/**
 * Client xác thực cho Backend (Go). Chỉ hoạt động khi đặt NEXT_PUBLIC_API_URL;
 * không đặt thì web chạy bản demo với dữ liệu giả như trước.
 *
 * Token lưu trong localStorage: đơn giản nhưng đọc được nếu có XSS. Khi triển khai thật nên chuyển
 * refresh token sang cookie HttpOnly. Access token hết hạn (401) thì tự đổi bằng refresh token, một lần
 * cho mọi request đang chờ.
 */
export const API_URL = (process.env.NEXT_PUBLIC_API_URL ?? '').replace(/\/$/, '');
export const hasApi = API_URL !== '';

const KEY = 'masm-tokens';

interface Tokens {
  access: string;
  refresh: string;
}

export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(message);
  }
}

export interface ApiUser {
  id: string;
  name: string;
  email: string;
  role: 'admin' | 'seller' | 'buyer';
}

function readTokens(): Tokens | null {
  try {
    const raw = localStorage.getItem(KEY);
    return raw ? (JSON.parse(raw) as Tokens) : null;
  } catch {
    return null;
  }
}
function writeTokens(t: Tokens | null) {
  try {
    if (t) localStorage.setItem(KEY, JSON.stringify(t));
    else localStorage.removeItem(KEY);
  } catch {
    /* bỏ qua: phiên chỉ sống tới khi tải lại trang */
  }
}

export const isSignedIn = () => readTokens() !== null;

async function parse<T>(res: Response): Promise<T> {
  if (res.status === 204) return undefined as T;
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    const e = (body as { error?: { code?: string; message?: string } }).error;
    throw new ApiError(res.status, e?.code ?? 'error', e?.message ?? 'Có lỗi xảy ra, bạn thử lại sau.');
  }
  return body as T;
}

function post<T>(path: string, body?: unknown, token?: string): Promise<T> {
  return fetch(API_URL + '/api/v1' + path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: body === undefined ? undefined : JSON.stringify(body),
  }).then((r) => parse<T>(r));
}

interface AuthResp {
  token: string;
  refresh_token: string;
  user: ApiUser;
}
function accept(r: AuthResp): ApiUser {
  writeTokens({ access: r.token, refresh: r.refresh_token });
  return r.user;
}

export const apiLogin = (email: string, password: string) => post<AuthResp>('/auth/login', { email, password }).then(accept);
export const apiRegister = (name: string, email: string, password: string, role: string) =>
  post<AuthResp>('/auth/register', { name, email, password, role }).then(accept);
export const forgotPassword = (email: string) => post<{ message: string }>('/auth/forgot-password', { email });
export const resetPassword = (token: string, newPassword: string) =>
  post<void>('/auth/reset-password', { token, new_password: newPassword });

export async function apiLogout() {
  const t = readTokens();
  writeTokens(null);
  if (t) await post('/auth/logout', { refresh_token: t.refresh }).catch(() => {});
}

/** Một lần refresh dùng chung: nhiều request cùng gặp 401 không được đổi token song song
 *  (refresh token dùng một lần; dùng lại sẽ bị coi là bị đánh cắp và đăng xuất mọi thiết bị). */
let refreshing: Promise<string | null> | null = null;
function refreshAccess(): Promise<string | null> {
  refreshing ??= (async () => {
    const t = readTokens();
    if (!t) return null;
    try {
      const r = await post<AuthResp>('/auth/refresh', { refresh_token: t.refresh });
      accept(r);
      return r.token;
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) writeTokens(null); // phiên đã hết hạn thật
      return null;
    } finally {
      setTimeout(() => (refreshing = null), 0);
    }
  })();
  return refreshing;
}

/** fetch có Authorization; gặp 401 thì refresh rồi thử lại đúng một lần. */
export async function authFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const call = (token: string) =>
    fetch(API_URL + '/api/v1' + path, { ...init, headers: { ...(init.headers as Record<string, string>), Authorization: `Bearer ${token}` } });
  const t = readTokens();
  if (!t) throw new ApiError(401, 'unauthorized', 'Bạn cần đăng nhập.');
  let res = await call(t.access);
  if (res.status === 401) {
    const fresh = await refreshAccess();
    if (!fresh) throw new ApiError(401, 'unauthorized', 'Phiên đăng nhập đã hết hạn. Bạn vui lòng đăng nhập lại.');
    res = await call(fresh);
  }
  return res;
}

export const authJSON = async <T,>(path: string, init?: RequestInit) => parse<T>(await authFetch(path, init));