export interface CardProps {
  /** px padding; 0 for media cards */
  padding?: number | string;
  /** Hover lift + focus ring (use when the whole card is a link) */
  interactive?: boolean;
  /** No shadow — for nested cards or dense admin views */
  flat?: boolean;
  as?: keyof JSX.IntrinsicElements;
  children?: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
}
export declare function Card(props: CardProps): JSX.Element;
