export interface TooltipProps {
  /** Short text, one line. */
  content: React.ReactNode;
  /** A single focusable element. */
  children: React.ReactElement;
}
export declare function Tooltip(props: TooltipProps): JSX.Element;
