/**
 * Shell for the 4-step sell flow: header (back/close, title, step count), Stepper, scrolling body, sticky CTA.
 * @startingPoint section="Listing" subtitle="Sell wizard shell — 4 steps, sticky CTA" viewport="390x700"
 */
export interface ListingWizardProps {
  /** Default: ['Ảnh','AI phân tích','Duyệt & sửa','Đăng'] */
  steps?: string[];
  current: number;
  title?: string;
  onBack?: () => void;
  onClose?: () => void;
  onNext?: () => void;
  nextLabel?: string;
  nextDisabled?: boolean;
  /** Hide the sticky footer (e.g. step 2 while AI runs, step 4 which has its own Đăng button) */
  hideFooter?: boolean;
  children?: React.ReactNode;
  style?: React.CSSProperties;
}
export declare function ListingWizard(props: ListingWizardProps): JSX.Element;
