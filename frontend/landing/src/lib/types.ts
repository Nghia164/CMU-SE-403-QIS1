export type Condition = 'Mới' | 'Như mới' | 'Tốt' | 'Khá' | 'Cũ';

export interface Listing {
  id: string;
  title: string;
  price: number;
  original?: number;
  condition: Condition;
  location: string;
  postedAt: string;
  /** Tên icon Lucide dùng làm ảnh giữ chỗ theo danh mục */
  icon: string;
  cat: string;
  seller: string;
  rating: number;
  sales: number;
  desc: string;
  sold?: boolean;
}

export interface Category {
  value: string;
  label: string;
  icon: string;
}

export interface PriceSource {
  source: string;
  title: string;
  price: number;
  condition: Condition;
}

export type OrderStatusKey =
  | 'pending_payment'
  | 'placed'
  | 'paid'
  | 'shipping'
  | 'delivered'
  | 'cancelled'
  | 'refunded';

export type PaymentProvider = 'vnpay' | 'momo';

export interface Order {
  id: string;
  listingId: string;
  status: OrderStatusKey;
  method: 'VNPay' | 'MoMo';
  events: Partial<Record<'placed' | 'paid' | 'shipping' | 'delivered' | 'cancelled', string>>;
}

export interface User {
  name: string;
  email: string;
}

/** Hợp đồng job gợi ý AI (khớp docs: Backend trả job_id, client poll trạng thái) */
export type AIJobStatus = 'QUEUED' | 'RUNNING' | 'SUCCEEDED' | 'FAILED';

export interface AIStep {
  label: string;
  status: 'pending' | 'running' | 'done' | 'error';
  detail?: string;
}

export interface AISuggestionResult {
  titles: string[];
  descriptions: string[];
  hashtags: string[];
  price: { quick: number; optimal: number; confidence: number; basis: string };
  sources: PriceSource[];
}

export interface AIJob {
  jobId: string;
  status: AIJobStatus;
  progress: number;
  steps: AIStep[];
  result?: AISuggestionResult;
}
