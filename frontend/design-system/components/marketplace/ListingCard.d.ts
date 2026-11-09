/**
 * Listing tile for feeds and search results: square photo, price, 2-line title, condition, location.
 * @startingPoint section="Marketplace" subtitle="Listing card — photo, price, condition" viewport="700x340"
 */
export interface ListingCardProps {
  title: string;
  /** VND */
  price: number;
  original?: number;
  /** "Mới" | "Như mới" | "Tốt" | "Khá" | "Cũ" */
  condition?: 'Mới' | 'Như mới' | 'Tốt' | 'Khá' | 'Cũ' | string;
  location?: string;
  /** Relative time, e.g. "2 giờ trước" */
  postedAt?: string;
  image?: string;
  /** Lucide icon for the empty-photo placeholder (category glyph) */
  placeholderIcon?: string;
  href?: string;
  saved?: boolean;
  /** Shows the heart toggle when provided */
  onSave?: () => void;
  sold?: boolean;
  onClick?: () => void;
  style?: React.CSSProperties;
}
export declare function ListingCard(props: ListingCardProps): JSX.Element;
