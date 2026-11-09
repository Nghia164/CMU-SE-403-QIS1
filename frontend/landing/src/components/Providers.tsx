'use client';
import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import type { User } from '@/lib/types';
import { apiLogout, hasApi } from '@/lib/session';

/* ---------- Tin đã lưu ---------- */
interface SavedCtx {
  saved: string[];
  toggle: (id: string) => void;
}
const SavedContext = createContext<SavedCtx>({ saved: [], toggle: () => {} });
export const useSaved = () => useContext(SavedContext);

/* ---------- Tài khoản (mock) ---------- */
interface AuthCtx {
  user: User | null;
  login: (u: User) => void;
  logout: () => void;
}
const AuthContext = createContext<AuthCtx>({ user: null, login: () => {}, logout: () => {} });
export const useAuth = () => useContext(AuthContext);

/* ---------- Giao diện sáng/tối ---------- */
type Theme = 'light' | 'dark';
interface ThemeCtx {
  theme: Theme;
  toggle: () => void;
}
const ThemeContext = createContext<ThemeCtx>({ theme: 'light', toggle: () => {} });
export const useTheme = () => useContext(ThemeContext);

function readJSON<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key);
    return raw ? (JSON.parse(raw) as T) : fallback;
  } catch {
    return fallback;
  }
}
function writeJSON(key: string, value: unknown) {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {
    /* localStorage không khả dụng: bỏ qua, giao diện vẫn chạy */
  }
}

export function Providers({ children }: { children: ReactNode }) {
  const [saved, setSaved] = useState<string[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [theme, setTheme] = useState<Theme>('light');

  useEffect(() => {
    setSaved(readJSON<string[]>('masm-saved', []));
    setUser(readJSON<User | null>('masm-user', null));
    const stored = readJSON<Theme | null>('masm-theme', null);
    const initial: Theme = stored ?? (window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
    setTheme(initial);
  }, []);

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  const toggleSaved = useCallback((id: string) => {
    setSaved((s) => {
      const next = s.includes(id) ? s.filter((x) => x !== id) : [...s, id];
      writeJSON('masm-saved', next);
      return next;
    });
  }, []);

  const authValue = useMemo<AuthCtx>(
    () => ({
      user,
      login: (u) => {
        setUser(u);
        writeJSON('masm-user', u);
      },
      logout: () => {
        if (hasApi) void apiLogout();
        setUser(null);
        try {
          localStorage.removeItem('masm-user');
        } catch {
          /* bỏ qua */
        }
      },
    }),
    [user],
  );

  const themeValue = useMemo<ThemeCtx>(
    () => ({
      theme,
      toggle: () =>
        setTheme((t) => {
          const next: Theme = t === 'dark' ? 'light' : 'dark';
          writeJSON('masm-theme', next);
          return next;
        }),
    }),
    [theme],
  );

  return (
    <ThemeContext.Provider value={themeValue}>
      <AuthContext.Provider value={authValue}>
        <SavedContext.Provider value={{ saved, toggle: toggleSaved }}>{children}</SavedContext.Provider>
      </AuthContext.Provider>
    </ThemeContext.Provider>
  );
}
