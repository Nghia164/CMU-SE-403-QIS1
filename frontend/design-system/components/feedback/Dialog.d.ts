export interface DialogProps {
  open: boolean;
  onClose?: () => void;
  title: string;
  children?: React.ReactNode;
  /** Buttons, right-aligned. Put the primary action last. */
  footer?: React.ReactNode;
  width?: number;
}
export declare function Dialog(props: DialogProps): JSX.Element | null;
