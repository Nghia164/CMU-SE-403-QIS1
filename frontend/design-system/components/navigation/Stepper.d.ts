export interface StepperProps {
  /** Short step labels, e.g. ["Ảnh","AI phân tích","Duyệt & sửa","Đăng"] */
  steps: string[];
  /** 0-based current step */
  current: number;
  label?: string;
}
export declare function Stepper(props: StepperProps): JSX.Element;
