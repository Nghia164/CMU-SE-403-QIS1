import type { Metadata } from 'next';
import { ForgotPasswordForm } from '@/components/PasswordForms';

export const metadata: Metadata = { title: 'Quên mật khẩu' };

export default function Page() {
  return <ForgotPasswordForm />;
}