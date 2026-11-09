export interface ToastProps {
  tone?: 'success' | 'danger' | 'info' | 'ai';
  title?: string;
  message?: React.ReactNode;
  actionLabel?: string;
  onAction?: () => void;
  onClose?: () => void;
  style?: React.CSSProperties;
}
export declare function Toast(props: ToastProps): JSX.Element;
