import Link from 'next/link';
import { Button } from '@/components/ds';

export default function NotFound() {
  return (
    <div className="container container--narrow stack">
      <h1 className="page-title">Không tìm thấy trang</h1>
      <p className="muted">Tin đăng hoặc trang bạn tìm có thể đã bị gỡ hoặc đường dẫn chưa đúng.</p>
      <Button as={Link} href="/" className="self-start">
        Về trang chủ
      </Button>
    </div>
  );
}
