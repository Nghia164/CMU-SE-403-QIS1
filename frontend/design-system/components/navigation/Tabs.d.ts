export interface TabItem { value: string; label: string; icon?: string; count?: number }
export interface TabsProps {
  items: TabItem[];
  value: string;
  onChange?: (value: string) => void;
  /** line = underline (page sections, admin) · pill = segmented control (filters, mobile) */
  variant?: 'line' | 'pill';
  /** aria-label for the tablist */
  label?: string;
}
export declare function Tabs(props: TabsProps): JSX.Element;
