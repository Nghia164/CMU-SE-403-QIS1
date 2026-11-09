export interface AiAgentStep {
  /** e.g. "Nhận diện sản phẩm", "Đánh giá tình trạng", "So sánh giá thị trường", "Soạn tiêu đề & mô tả" */
  label: string;
  status: 'pending' | 'running' | 'done' | 'error';
  /** Result or sub-status line, e.g. "Fujifilm X-T20 · độ tin cậy 91%" */
  detail?: string;
}
export interface AiProgressProps {
  title?: string;
  steps: AiAgentStep[];
  /** 0–100; derived from done steps when omitted */
  progress?: number;
  /** Escape hatch to manual entry */
  onCancel?: () => void;
}
export declare function AiProgress(props: AiProgressProps): JSX.Element;
