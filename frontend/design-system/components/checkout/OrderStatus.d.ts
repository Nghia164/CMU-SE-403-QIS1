export type OrderStatusKey = 'pending_payment' | 'placed' | 'paid' | 'shipping' | 'delivered' | 'cancelled' | 'refunded';
export interface OrderStatusProps {
  status: OrderStatusKey;
  /** timeline = vertical stepper (order detail) · badge = compact pill (lists, admin) */
  variant?: 'timeline' | 'badge';
  /** Sub-line per step key, e.g. { paid: '14:02 · 08/11 · qua VNPay' } */
  events?: Partial<Record<'placed' | 'paid' | 'shipping' | 'delivered' | 'cancelled', string>>;
}
export declare function OrderStatus(props: OrderStatusProps): JSX.Element;
