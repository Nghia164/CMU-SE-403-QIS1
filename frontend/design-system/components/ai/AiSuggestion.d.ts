export interface AiSuggestionProps {
  /** What was suggested: "Tiêu đề", "Mô tả", "Hashtag", "Danh mục" */
  label?: string;
  children?: React.ReactNode;
  /** Always provide — AI only proposes, seller can edit. */
  onEdit?: () => void;
  /** Always provide — seller can ask for another proposal. */
  onRegenerate?: () => void;
  regenerating?: boolean;
  editLabel?: string;
  regenerateLabel?: string;
  footer?: React.ReactNode;
}
export declare function AiSuggestion(props: AiSuggestionProps): JSX.Element;
