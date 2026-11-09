export interface PriceProps {
  /** Amount in VND (integer). Rendered as "1.450.000 ₫". */
  value: number;
  /** Original/new price, struck through */
  original?: number;
  size?: 'sm' | 'md' | 'lg';
  /** primary = teal (listing cards) · ai = amber (AI-suggested amounts only) */
  tone?: 'strong' | 'primary' | 'ai';
  style?: React.CSSProperties;
}
export declare function Price(props: PriceProps): JSX.Element;
