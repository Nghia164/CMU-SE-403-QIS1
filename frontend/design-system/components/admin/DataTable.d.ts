export interface DataColumn<R = any> {
  key: string;
  header: React.ReactNode;
  width?: number | string;
  align?: 'left' | 'right' | 'center';
  render?: (row: R) => React.ReactNode;
}
/**
 * Admin table: sticky header, row hover, optional selection for bulk actions.
 * @startingPoint section="Admin" subtitle="Data table — accounts, moderation, reports" viewport="700x320"
 */
export interface DataTableProps<R = any> {
  columns: DataColumn<R>[];
  rows: R[];
  /** Field used as React key + selection id (default "id") */
  rowKey?: string;
  selectable?: boolean;
  selected?: (string | number)[];
  onSelectedChange?: (ids: (string | number)[]) => void;
  dense?: boolean;
  /** Screen-reader caption */
  caption?: string;
  empty?: React.ReactNode;
}
export declare function DataTable(props: DataTableProps): JSX.Element;
