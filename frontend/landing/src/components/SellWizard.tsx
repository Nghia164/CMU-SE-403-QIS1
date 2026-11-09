'use client';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useCallback, useEffect, useRef, useState } from 'react';
import { createAIJob, getAIJob } from '@/lib/ai-mock';
import { fmtNum } from '@/lib/format';
import type { AIJob, AISuggestionResult } from '@/lib/types';
import {
  AiBadge,
  AiPriceRange,
  AiProgress,
  AiSuggestion,
  Badge,
  Button,
  Card,
  Icon,
  Input,
  ListingWizard,
  PhotoUploader,
  Price,
  PublishConfirm,
  Select,
} from './ds';

interface Photo {
  id: number;
  src: string;
  alt: string;
}

/* ---------- Trường do AI gợi ý: amber khi còn là gợi ý, mất amber khi người bán sửa ---------- */
function AiField({
  label,
  value,
  onChange,
  variants,
  multiline,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  variants: string[];
  multiline?: boolean;
}) {
  const [edit, setEdit] = useState(false);
  const [regen, setRegen] = useState(false);
  const [idx, setIdx] = useState(0);

  if (edit) {
    return (
      <div className="stack stack--xs">
        <Input label={label} value={value} onChange={(e) => onChange(e.target.value)} multiline={multiline} rows={4} autoFocus />
        <Button
          variant="quiet"
          size="sm"
          icon="sparkles"
          className="self-start"
          onClick={() => {
            onChange(variants[idx]);
            setEdit(false);
          }}
        >
          Khôi phục gợi ý AI
        </Button>
      </div>
    );
  }
  return (
    <AiSuggestion
      label={label}
      regenerating={regen}
      onEdit={() => setEdit(true)}
      onRegenerate={() => {
        setRegen(true);
        setTimeout(() => {
          const n = (idx + 1) % variants.length;
          setIdx(n);
          onChange(variants[n]);
          setRegen(false);
        }, 900);
      }}
    >
      <span className={multiline ? 'ai-text ai-text--sm' : 'ai-text'}>{value}</span>
    </AiSuggestion>
  );
}

export function SellWizard() {
  const router = useRouter();
  const [step, setStep] = useState(0);
  const [photos, setPhotos] = useState<Photo[]>([]);
  const [job, setJob] = useState<AIJob | null>(null);
  const [manual, setManual] = useState(false);
  const [result, setResult] = useState<AISuggestionResult | null>(null);

  const [title, setTitle] = useState('');
  const [desc, setDesc] = useState('');
  const [tags, setTags] = useState('');
  const [price, setPrice] = useState(0);
  const [pick, setPick] = useState<'quick' | 'optimal'>('optimal');
  const [priceRegen, setPriceRegen] = useState(false);
  const [category, setCategory] = useState('Xe đạp');
  const [cond, setCond] = useState('Tốt');
  const [ok, setOk] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [done, setDone] = useState(false);
  const photoUrls = useRef<string[]>([]);

  useEffect(() => () => photoUrls.current.forEach((u) => URL.revokeObjectURL(u)), []);

  const addPhotos = useCallback((files: File[]) => {
    setPhotos((p) => {
      const room = 8 - p.length;
      const next = files
        .filter((f) => f.type.startsWith('image/'))
        .slice(0, room)
        .map((f, i) => {
          const src = URL.createObjectURL(f);
          photoUrls.current.push(src);
          return { id: Date.now() + i, src, alt: f.name };
        });
      return [...p, ...next];
    });
  }, []);

  /* Bước 1: tạo job AI (nhận jobId ngay) rồi poll trạng thái */
  const startAnalysis = async () => {
    setStep(1);
    setManual(false);
    setJob(null);
    const { jobId } = await createAIJob(photos.length);
    setJob(await getAIJob(jobId));
    jobRef.current = jobId;
  };
  const jobRef = useRef<string | null>(null);

  useEffect(() => {
    if (step !== 1 || manual) return;
    const t = setInterval(async () => {
      if (!jobRef.current) return;
      const j = await getAIJob(jobRef.current);
      setJob(j);
      if (j.status === 'SUCCEEDED' || j.status === 'FAILED') clearInterval(t);
    }, 500);
    return () => clearInterval(t);
  }, [step, manual]);

  /* Khi AI xong lần đầu: điền sẵn các trường bằng gợi ý (vẫn là đề xuất, người bán sửa được) */
  useEffect(() => {
    if (job?.status === 'SUCCEEDED' && job.result && !result) {
      setResult(job.result);
      setTitle(job.result.titles[0]);
      setDesc(job.result.descriptions[0]);
      setTags(job.result.hashtags[0]);
      setPrice(job.result.price.optimal);
      setPick('optimal');
    }
  }, [job, result]);

  const aiDone = job?.status === 'SUCCEEDED';
  const customPrice = !!result && price !== result.price.quick && price !== result.price.optimal;
  const usingAI = !manual && !!result;

  const goManual = () => {
    setManual(true);
    setResult(null);
    setTitle('');
    setDesc('');
    setTags('');
    setPrice(0);
    setStep(2);
  };

  if (done) {
    return (
      <div className="sell-done">
        <span className="done-icon">
          <Icon name="circle-check" size={40} />
        </span>
        <h2>Đã đăng tin!</h2>
        <p className="muted">Tin của bạn đang chờ kiểm duyệt, thường mất dưới 15 phút. Chúng tôi sẽ báo khi tin được hiển thị.</p>
        <div className="stack stack--sm done-actions">
          <Button block as={Link} href="/">
            Về trang chủ
          </Button>
          <Button
            block
            variant="secondary"
            onClick={() => {
              setDone(false);
              setStep(0);
              setOk(false);
              setPhotos([]);
              setResult(null);
              setJob(null);
              jobRef.current = null;
            }}
          >
            Đăng tin khác
          </Button>
        </div>
      </div>
    );
  }

  const footer =
    step === 0
      ? { onNext: startAnalysis, nextLabel: 'Phân tích bằng AI', nextDisabled: photos.length === 0 }
      : step === 1
        ? { onNext: () => setStep(2), nextLabel: aiDone ? 'Xem gợi ý' : 'AI đang phân tích…', nextDisabled: !aiDone }
        : step === 2
          ? { onNext: () => setStep(3), nextLabel: 'Tiếp tục', nextDisabled: !title.trim() || !price }
          : { hideFooter: true };

  return (
    <ListingWizard current={step} onBack={() => setStep(step - 1)} onClose={() => router.push('/')} {...footer}>
      {step === 0 && (
        <>
          <div>
            <h2 className="step-title">Thêm ảnh sản phẩm</h2>
            <p className="muted small">AI sẽ nhận diện, đề xuất giá và soạn nội dung. Bạn luôn có thể sửa lại.</p>
          </div>
          <PhotoUploader photos={photos} onAdd={addPhotos} onRemove={(id: number) => setPhotos((p) => p.filter((x) => x.id !== id))} />
          <button type="button" className="link-btn" onClick={goManual}>
            Tự nhập thông tin, không dùng AI
          </button>
        </>
      )}

      {step === 1 && (
        <>
          <div className="row">
            {photos.slice(0, 4).map((p) => (
              <div key={p.id} className="masm-thumb" style={{ width: 56 }}>
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src={p.src} alt="" />
              </div>
            ))}
          </div>
          <AiProgress steps={job?.steps ?? []} onCancel={goManual} />
        </>
      )}

      {step === 2 && (
        <>
          {usingAI && result ? (
            <AiField label="Tiêu đề" value={title} onChange={setTitle} variants={result.titles} />
          ) : (
            <Input label="Tiêu đề" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Ví dụ: Xe đạp Asama 26 inch, còn mới 90%" />
          )}
          <div className="form-2">
            <Select label="Danh mục" options={['Xe đạp', 'Xe máy', 'Phụ tùng', 'Điện tử', 'Nhà cửa', 'Thời trang', 'Sách', 'Khác']} value={category} onChange={(e) => setCategory(e.target.value)} />
            <Select label="Tình trạng" options={['Mới', 'Như mới', 'Tốt', 'Khá', 'Cũ']} value={cond} onChange={(e) => setCond(e.target.value)} />
          </div>

          {usingAI && result && (
            <AiPriceRange
              quick={result.price.quick}
              optimal={result.price.optimal}
              confidence={result.price.confidence}
              basis={result.price.basis}
              sources={result.sources}
              maxSources={3}
              selected={customPrice ? undefined : pick}
              onSelect={(w: 'quick' | 'optimal') => {
                setPick(w);
                setPrice(result.price[w]);
              }}
              regenerating={priceRegen}
              onRegenerate={() => {
                setPriceRegen(true);
                setTimeout(() => setPriceRegen(false), 900);
              }}
              onEdit={() => document.getElementById('sell-price')?.focus()}
            />
          )}
          <Input
            id="sell-price"
            label="Giá bán của bạn"
            suffix="₫"
            inputMode="numeric"
            value={fmtNum(price)}
            onChange={(e) => setPrice(+e.target.value.replace(/\D/g, ''))}
            hint={usingAI ? (customPrice ? 'Bạn đã tự đặt giá.' : 'Đã điền theo gợi ý AI, bạn có thể sửa.') : 'Nhập giá bạn muốn bán.'}
          />

          {usingAI && result ? (
            <>
              <AiField label="Mô tả" value={desc} onChange={setDesc} variants={result.descriptions} multiline />
              <AiField label="Hashtag" value={tags} onChange={setTags} variants={result.hashtags} />
            </>
          ) : (
            <>
              <Input label="Mô tả" value={desc} onChange={(e) => setDesc(e.target.value)} multiline rows={4} placeholder="Tình trạng, thời gian sử dụng, phụ kiện đi kèm…" />
              <Input label="Hashtag" value={tags} onChange={(e) => setTags(e.target.value)} placeholder="#xedap #asama" />
            </>
          )}
          <p className="muted small">AI chỉ đề xuất. Bạn là người quyết định giá bán cuối cùng.</p>
        </>
      )}

      {step === 3 && (
        <>
          <h2 className="step-title">Xác nhận đăng tin</h2>
          <PublishConfirm
            confirmed={ok}
            onConfirmedChange={setOk}
            publishing={publishing}
            onPublish={() => {
              setPublishing(true);
              setTimeout(() => {
                setPublishing(false);
                setDone(true);
              }, 900);
            }}
          >
            <Card className="row row--top">
              <div className="masm-thumb" style={{ width: 80, flex: 'none' }}>
                {photos[0] ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={photos[0].src} alt="" />
                ) : (
                  <Icon name="image" size={28} />
                )}
              </div>
              <div className="stack stack--xs minw0">
                <div className="card-title">{title}</div>
                <Price value={price} tone="primary" />
                <div className="row row--wrap">
                  <Badge size="sm">{cond}</Badge>
                  <Badge size="sm">{photos.length} ảnh</Badge>
                  {usingAI && !customPrice && <AiBadge size="sm">Giá AI</AiBadge>}
                </div>
              </div>
            </Card>
            <Card className="small">
              {desc}
              <div className="tags">{tags}</div>
            </Card>
            <Button variant="ghost" icon="pencil" size="sm" className="self-start" onClick={() => setStep(2)}>
              Quay lại chỉnh sửa
            </Button>
          </PublishConfirm>
        </>
      )}
    </ListingWizard>
  );
}
