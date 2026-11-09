export interface SelectOption { value: string; label: string }
export interface SelectProps {
  label?: string;
  hint?: string;
  error?: string;
  options: (string | SelectOption)[];
  placeholder?: string;
  value?: string;
  defaultValue?: string;
  onChange?: (e: React.ChangeEvent<HTMLSelectElement>) => void;
  disabled?: boolean;
  id?: string;
  className?: string;
}
export declare function Select(props: SelectProps): JSX.Element;
