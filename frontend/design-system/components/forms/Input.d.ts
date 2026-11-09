export interface InputProps {
  label?: string;
  hint?: string;
  /** Error message; sets aria-invalid and replaces hint. */
  error?: string;
  /** Leading Lucide icon (not for multiline). */
  icon?: string;
  /** Trailing unit text, e.g. "₫". */
  suffix?: string;
  multiline?: boolean;
  rows?: number;
  /** Field currently holds an AI-written value (amber fill) — remove once the seller edits. */
  ai?: boolean;
  /** Extra node beside the label, e.g. <Badge tone="ai">. */
  labelExtra?: React.ReactNode;
  id?: string;
  value?: string;
  defaultValue?: string;
  placeholder?: string;
  type?: string;
  inputMode?: string;
  required?: boolean;
  disabled?: boolean;
  onChange?: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
  className?: string;
}
export declare function Input(props: InputProps): JSX.Element;
