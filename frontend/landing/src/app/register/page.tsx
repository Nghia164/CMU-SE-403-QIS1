import type { Metadata } from 'next';
import { AuthForm } from '@/components/AuthForm';

export const metadata: Metadata = { title: 'Tạo tài khoản' };

export default function RegisterPage() {
  return <AuthForm mode="register" />;
}
