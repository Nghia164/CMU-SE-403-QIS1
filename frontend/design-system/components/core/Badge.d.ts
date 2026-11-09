export interface BadgeProps {
  /** ai is reserved for AI-generated content. warning is clay (not amber). solid = dark chip for photo overlays. */
  tone?: 'neutral' | 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'ai' | 'solid';
  size?: 'sm' | 'md';
  icon?: string;
  children?: React.ReactNode;
  className?: string;
}
export declare function Badge(props: BadgeProps): JSX.Element;
