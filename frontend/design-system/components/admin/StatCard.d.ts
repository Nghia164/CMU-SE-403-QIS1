export interface StatCardProps {
  label: string;
  value: React.ReactNode;
  /** e.g. "+12%" */
  delta?: string;
  deltaTone?: 'success' | 'danger' | 'neutral';
  icon?: string;
  hint?: string;
}
export declare function StatCard(props: StatCardProps): JSX.Element;
