export interface PriceSource {
  /** Marketplace name, e.g. "Chợ Tốt" */
  source: string;
  title: string;
  price: number;
  condition?: string;
  url?: string;
}
/**
 * AI price block: quick-sale vs optimal price, range track, confidence, comparison sources.
 * @startingPoint section="AI" subtitle="AI price range — quick vs optimal, confidence, sources" viewport="700x560"
 */
export interface AiPriceRangeProps {
  /** Bán nhanh — VND */
  quick: number;
  /** Giá tối ưu — VND */
  optimal: number;
  quickEta?: string;
  optimalEta?: string;
  /** 0–100 */
  confidence?: number;
  /** One line explaining the basis, e.g. "Dựa trên 12 tin tương tự trong 30 ngày" */
  basis?: string;
  sources?: PriceSource[];
  /** Which tile the seller picked */
  selected?: 'quick' | 'optimal';
  onSelect?: (which: 'quick' | 'optimal') => void;
  /** Seller enters their own price */
  onEdit?: () => void;
  onRegenerate?: () => void;
  regenerating?: boolean;
  /** Sources shown before "Xem thêm" (default 4) */
  maxSources?: number;
}
export declare function AiPriceRange(props: AiPriceRangeProps): JSX.Element;
