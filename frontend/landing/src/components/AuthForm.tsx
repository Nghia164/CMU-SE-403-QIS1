'use client';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useState, type FormEvent } from 'react';
import { Button, Card, Input, Select } from './ds';
import { useAuth } from './Providers';
import { ApiError, apiLogin, apiRegister, hasApi } from '@/lib/session';

type Mode = 'login' | 'register';

/** Có NEXT_PUBLIC_API_URL thì đăng nhập/đăng ký qua Backend; không thì chạy bản demo giả lập. */
export function AuthForm({ mode }: { mode: Mode }) {
  const router = useRouter();
  const { login } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState('buyer');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const isRegister = mode === 'register';

  const [busy, setBusy] = useState(false);

  const submit = async (ev: FormEvent) => {
    ev.preventDefault();
    const e: Record<string, string> = {};
    if (isRegister && !name.trim()) e.name = 'Nhập tên hiển thị của bạn';
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) e.email = 'Email chưa đúng định dạng, ví dụ ban@email.com';
    if (password.length < 8) e.password = 'Mật khẩu cần ít nhất 8 ký tự';
    setErrors(e);
    if (Object.keys(e).length) return;
    if (!hasApi) {
      login({ name: isRegister ? name.trim() : email.split('@')[0], email });
      router.push('/');
      return;
    }
    setBusy(true);
    try {
      const u = isRegister ? await apiRegister(name.trim(), email.trim(), password, role) : await apiLogin(email.trim(), password);
      login({ name: u.name, email: u.email });
      router.push('/');
    } catch (err) {
      setErrors({ form: err instanceof ApiError ? err.message : 'Không kết nối được máy chủ. Bạn thử lại sau.' });
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="container container--auth stack">
      <h1 className="page-title">{isRegister ? 'Tạo tài khoản' : 'Đăng nhập'}</h1>
      <Card>
        <form className="stack stack--sm" onSubmit={submit} noValidate>
          {isRegister && <Input label="Tên hiển thị" value={name} onChange={(e) => setName(e.target.value)} error={errors.name} autoComplete="name" />}
          <Input label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} error={errors.email} autoComplete="email" />
          <Input
            label="Mật khẩu"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={errors.password}
            hint={isRegister ? 'Ít nhất 8 ký tự.' : undefined}
            autoComplete={isRegister ? 'new-password' : 'current-password'}
          />
          {isRegister && (
            <Select
              label="Bạn muốn"
              options={[
                { value: 'buyer', label: 'Mua đồ cũ' },
                { value: 'seller', label: 'Bán đồ cũ' },
              ]}
              value={role}
              onChange={(e) => setRole(e.target.value)}
            />
          )}
          {errors.form && (
            <p className="field-error" role="alert">
              {errors.form}
            </p>
          )}
          {!isRegister && (
            <p className="small">
              <Link href="/quen-mat-khau">Quên mật khẩu?</Link>
            </p>
          )}
          <Button type="submit" size="lg" block disabled={busy}>
            {busy ? 'Đang xử lý…' : isRegister ? 'Tạo tài khoản' : 'Đăng nhập'}
          </Button>
        </form>
      </Card>
      <p className="muted small center">
        {isRegister ? (
          <>
            Đã có tài khoản? <Link href="/login">Đăng nhập</Link>
          </>
        ) : (
          <>
            Chưa có tài khoản? <Link href="/register">Tạo tài khoản</Link>
          </>
        )}
      </p>
      {!hasApi && <p className="muted small center">Bản demo: chưa kết nối máy chủ, thông tin chỉ lưu trong trình duyệt này.</p>}
    </div>
  );
}
