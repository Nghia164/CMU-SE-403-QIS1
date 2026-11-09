export interface PublishConfirmProps {
  confirmed: boolean;
  onConfirmedChange?: (v: boolean) => void;
  onPublish?: () => void;
  publishing?: boolean;
  /** Summary of the listing shown above the confirmation */
  children?: React.ReactNode;
  confirmLabel?: string;
  confirmDescription?: string;
  publishLabel?: string;
}
export declare function PublishConfirm(props: PublishConfirmProps): JSX.Element;
