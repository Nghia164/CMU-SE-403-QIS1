/**
 * Primary action control. One primary per view; AI variant only for AI-originated actions.
 * @startingPoint section="Core" subtitle="Buttons — primary, secondary, ghost, danger, AI" viewport="700x260"
 */
export interface ButtonProps {
  /** primary = teal fill · secondary = outlined · ghost = teal text · quiet = muted text · danger = destructive · ai = amber (AI actions: "Tạo lại", "Dùng gợi ý") */
  variant?: 'primary' | 'secondary' | 'ghost' | 'quiet' | 'danger' | 'ai';
  /** sm 36px · md 44px (default, mobile tap target) · lg 52px (sticky CTA) */
  size?: 'sm' | 'md' | 'lg';
  block?: boolean;
  /** Lucide icon name, leading */
  icon?: string;
  iconRight?: string;
  loading?: boolean;
  disabled?: boolean;
  href?: string;
  type?: 'button' | 'submit' | 'reset';
  onClick?: (e: React.MouseEvent) => void;
  children?: React.ReactNode;
  className?: string;
}
export declare function Button(props: ButtonProps): JSX.Element;
