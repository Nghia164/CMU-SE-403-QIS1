'use client';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { useState, type FormEvent } from 'react';
import { Button, Card, Input } from './ds';
import { ApiError, forgotPassword, hasApi, resetPassword } from '@/lib/session';

const msg = (e: unknown) => (e instanceof ApiError ? e.message : 'Không kết nối được máy chủ. Bạn thử lại sau.');

export function ForgotPasswordForm() {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [sent, setSent] = useState(false);
  const [busy, setBusy] = useState(false);

  const submit = async (ev: FormEvent) => {
    ev.preventDefault();
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) return setError('Email chưa đúng định dạng, ví dụ ban@email.com');
    setError('');
    setBusy(true);
    try {
      if (hasApi) await forgotPassword(email.trim());
      setSent(true);
    } catch (e) {
      setError(msg(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="container container--auth stack">
      <h1 className="page-title">Quên mật khẩu</h1>
      <Card>
        {sent ? (
          <div className="stack stack--sm" role="status">
            <p>Nếu email này đã đăng ký, chúng tôi đã gửi hướng dẫn đặt lại mật khẩu. Liên kết có hiệu lực trong 30 phút.</p>
            <p className="muted small">Bạn chưa thấy thư? Kiểm tra mục thư rác hoặc thử lại sau vài phút.</p>
          </div>
        ) : (
          <form className="stack stack--sm" onSubmit={submit} noValidate>
            <p className="muted small">Nhập email đã đăng ký, chúng tôi sẽ gửi liên kết để bạn đặt mật khẩu mới.</p>
            <Input label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} error={error} autoComplete="email" />
            <Button type="submit" size="lg" block disabled={busy}>
              {busy ? 'Đang gửi…' : 'Gửi liên kết'}
            </Button>
          </form>
        )}
      </Card>
      <p className="muted small center">
        <Link href="/login">Quay lại đăng nhập</Link>
      </p>
    </div>
  );
}

export function ResetPasswordForm() {
  const token = useSearchParams().get('token') ?? '';
  const [pw, setPw] = useState('');
  const [again, setAgain] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [done, setDone] = useState(false);
  const [busy, setBusy] = useState(false);

  const submit = async (ev: FormEvent) => {
    ev.preventDefault();
    const e: Record<string, string> = {};
    if (pw.length < 8) e.pw = 'Mật khẩu cần ít nhất 8 ký tự';
    else if (pw.length > 72) e.pw = 'Mật khẩu tối đa 72 ký tự';
    if (again !== pw) e.again = 'Hai mật khẩu chưa giống nhau';
    setErrors(e);
    if (Object.keys(e).length) return;
    setBusy(true);
    try {
      if (hasApi) await resetPassword(token, pw);
      setDone(true);
    } catch (err) {
      setErrors({ form: msg(err) });
    } finally {
      setBusy(false);
    }
  };

  if (!token) {
    return (
      <div className="container container--auth stack">
        <h1 className="page-title">Đặt lại mật khẩu</h1>
        <Card>
          <p>Liên kết không hợp lệ. Bạn hãy <Link href="/quen-mat-khau">yêu cầu liên kết mới</Link>.</p>
        </Card>
      </div>
    );
  }
  return (
    <div className="container container--auth stack">
      <h1 className="page-title">Đặt lại mật khẩu</h1>
      <Card>
        {done ? (
          <div className="stack stack--sm" role="status">
            <p>Mật khẩu đã được đổi. Bạn hãy đăng nhập lại bằng mật khẩu mới.</p>
            <Button as={Link} href="/login" block>
              Đăng nhập
            </Button>
          </div>
        ) : (
          <form className="stack stack--sm" onSubmit={submit} noValidate>
            <Input label="Mật khẩu mới" type="password" value={pw} onChange={(e) => setPw(e.target.value)} error={errors.pw} hint="Ít nhất 8 ký tự." autoComplete="new-password" />
            <Input label="Nhập lại mật khẩu mới" type="password" value={again} onChange={(e) => setAgain(e.target.value)} error={errors.again} autoComplete="new-password" />
            {errors.form && (
              <p className="field-error" role="alert">
                {errors.form} <Link href="/quen-mat-khau">Yêu cầu liên kết mới</Link>
              </p>
            )}
            <Button type="submit" size="lg" block disabled={busy}>
              {busy ? 'Đang lưu…' : 'Đặt lại mật khẩu'}
            </Button>
          </form>
        )}
      </Card>
    </div>
  );
}