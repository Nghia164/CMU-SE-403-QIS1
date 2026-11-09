export interface IconButtonProps {
  icon: string;
  /** Required accessible name (Vietnamese), also used as tooltip title. */
  label: string;
  variant?: 'primary' | 'secondary' | 'ghost' | 'quiet' | 'danger' | 'ai';
  size?: 'sm' | 'md';
  round?: boolean;
  /** White floating chip over photos (favorite, remove). */
  float?: boolean;
  /** Toggle state, e.g. favorite. Pressed float buttons turn red. */
  pressed?: boolean;
  onClick?: (e: React.MouseEvent) => void;
  disabled?: boolean;
  className?: string;
}
export declare function IconButton(props: IconButtonProps): JSX.Element;
