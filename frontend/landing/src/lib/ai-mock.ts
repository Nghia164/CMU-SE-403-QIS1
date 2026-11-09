/**
 * Mô phỏng job gợi ý AI bất đồng bộ.
 * Hợp đồng giống Backend thật: createAIJob trả jobId ngay; getAIJob được poll cho đến khi SUCCEEDED/FAILED.
 * Thay phần cài đặt bằng gọi API (POST /api/ai-jobs, GET /api/ai-jobs/:id) khi Backend sẵn sàng.
 */
import { PRICE_SOURCES } from './mock-data';
import type { AIJob, AIStep, AISuggestionResult } from './types';

const AGENTS = ['Nhận diện sản phẩm', 'Đánh giá tình trạng', 'So sánh giá thị trường', 'Soạn tiêu đề, mô tả và hashtag'];
const DETAIL = [
  'Xe đạp Asama 26 inch · độ chính xác 91%',
  'Tốt — trầy nhẹ ở khung sau',
  '12 tin tương tự trong 30 ngày',
  'Đã tạo 1 tiêu đề, 1 mô tả, 5 hashtag',
];
const STEP_MS = 900;

const RESULT: AISuggestionResult = {
  titles: [
    'Xe đạp Asama 26 inch, mới thay lốp, khung chắc',
    'Xe đạp Asama 26" đi học, đã bảo dưỡng',
    'Bán xe đạp Asama 26 inch còn tốt, giá mềm',
  ],
  descriptions: [
    'Xe đạp Asama 26 inch, tình trạng tốt. Mới thay lốp và xích, phanh ăn, khung không gỉ. Có vài vết trầy nhẹ ở khung sau. Phù hợp đi học, đi làm quãng ngắn.',
    'Asama 26 inch đã qua sử dụng, vận hành êm. Lốp và xích mới thay, yên êm. Trầy nhẹ ngoại hình, không ảnh hưởng sử dụng.',
  ],
  hashtags: ['#xedap #asama #xedapcu #26inch #hcm', '#xedapdihoc #asama #dodung #thanhly'],
  price: { quick: 1200000, optimal: 1450000, confidence: 82, basis: 'Dựa trên 12 tin tương tự trong 30 ngày' },
  sources: PRICE_SOURCES,
};

const started = new Map<string, number>();

export async function createAIJob(_photoCount: number): Promise<{ jobId: string }> {
  const jobId = 'job-' + Math.random().toString(36).slice(2, 10);
  started.set(jobId, Date.now());
  return { jobId };
}

export async function getAIJob(jobId: string): Promise<AIJob> {
  const t0 = started.get(jobId) ?? Date.now();
  const elapsed = Date.now() - t0;
  const at = Math.min(AGENTS.length, Math.floor(elapsed / STEP_MS));
  const done = at >= AGENTS.length;
  const steps: AIStep[] = AGENTS.map((label, i) => ({
    label,
    status: i < at ? 'done' : i === at ? 'running' : 'pending',
    detail: i < at ? DETAIL[i] : i === at ? 'Đang xử lý…' : undefined,
  }));
  return {
    jobId,
    status: done ? 'SUCCEEDED' : elapsed > 0 ? 'RUNNING' : 'QUEUED',
    progress: Math.round((at / AGENTS.length) * 100),
    steps,
    result: done ? RESULT : undefined,
  };
}
