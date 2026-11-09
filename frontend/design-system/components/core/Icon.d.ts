export interface IconProps {
  /** Lucide icon name in kebab-case, e.g. "camera", "sparkles", "shopping-bag". */
  name: string;
  /** Pixel size. 16 inline, 20 default, 24 nav. */
  size?: number;
  /** Accessible label. Omit for decorative icons (aria-hidden). */
  label?: string;
  className?: string;
  style?: React.CSSProperties;
}
export declare function Icon(props: IconProps): JSX.Element;
