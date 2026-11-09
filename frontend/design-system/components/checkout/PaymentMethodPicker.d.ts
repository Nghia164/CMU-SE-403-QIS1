export interface PaymentMethod { value: string; name: string; desc?: string; /** Text mark shown in the colored tile (no partner logo assets supplied) */ logo: string; color: string; disabled?: boolean }
/**
 * Radio-group tiles for VNPay / MoMo at checkout.
 * @startingPoint section="Checkout" subtitle="VNPay / MoMo payment method picker" viewport="700x260"
 */
export interface PaymentMethodPickerProps {
  value?: string;
  onChange?: (value: string) => void;
  /** Defaults to VNPay + MoMo */
  methods?: PaymentMethod[];
  name?: string;
  legend?: string;
}
export declare function PaymentMethodPicker(props: PaymentMethodPickerProps): JSX.Element;
