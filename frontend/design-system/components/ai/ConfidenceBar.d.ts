export interface ConfidenceBarProps {
  /** 0–100. Levels: ≥75 Cao · ≥50 Trung bình · else Thấp */
  value: number;
  label?: string;
  note?: React.ReactNode;
}
export declare function ConfidenceBar(props: ConfidenceBarProps): JSX.Element;
