'use client';
/**
 * Cổng duy nhất để dùng MASM Design System (frontend/design-system).
 * - 'use client': các component của design system dùng hook nên chạy ở client; re-export tại đây
 *   để Server Component vẫn render được (chỉ truyền props có thể serialize).
 * - Ép kiểu "lỏng" để chấp nhận style, aria-*, data-*, as... mà design system truyền xuống DOM.
 */
import type { ComponentType, CSSProperties, ElementType } from 'react';
import { Button as DSButton, type ButtonProps } from '@ds/components/core/Button';
import { IconButton as DSIconButton, type IconButtonProps } from '@ds/components/core/IconButton';
import { Icon as DSIcon, type IconProps } from '@ds/components/core/Icon';
import { Badge as DSBadge, type BadgeProps } from '@ds/components/core/Badge';
import { Card as DSCard, type CardProps } from '@ds/components/core/Card';
import { Input as DSInput, type InputProps } from '@ds/components/forms/Input';
import { Select as DSSelect, type SelectProps } from '@ds/components/forms/Select';
import { Checkbox as DSCheckbox, type CheckboxProps } from '@ds/components/forms/Checkbox';
import { Radio as DSRadio, type RadioProps } from '@ds/components/forms/Radio';
import { Toast as DSToast, type ToastProps } from '@ds/components/feedback/Toast';
import { Stepper as DSStepper, type StepperProps } from '@ds/components/navigation/Stepper';
import { ListingCard as DSListingCard, type ListingCardProps } from '@ds/components/marketplace/ListingCard';
import { Price as DSPrice, type PriceProps } from '@ds/components/marketplace/Price';
import { PaymentMethodPicker as DSPaymentMethodPicker, type PaymentMethodPickerProps } from '@ds/components/checkout/PaymentMethodPicker';
import { OrderStatus as DSOrderStatus, type OrderStatusProps } from '@ds/components/checkout/OrderStatus';
import { AiSuggestion as DSAiSuggestion, type AiSuggestionProps } from '@ds/components/ai/AiSuggestion';
import { AiProgress as DSAiProgress, type AiProgressProps } from '@ds/components/ai/AiProgress';
import { AiBadge as DSAiBadge } from '@ds/components/ai/AiBadge';
import { ListingWizard as DSListingWizard } from '@ds/components/listing/ListingWizard';
import { PhotoUploader as DSPhotoUploader } from '@ds/components/listing/PhotoUploader';
import { PublishConfirm as DSPublishConfirm } from '@ds/components/listing/PublishConfirm';
import { AiPriceRange as DSAiPriceRange } from '@ds/components/ai/AiPriceRange';

type Extra = {
  style?: CSSProperties;
  as?: ElementType;
  id?: string;
  title?: string;
  tabIndex?: number;
  autoFocus?: boolean;
  autoComplete?: string;
  name?: string;
  role?: string;
  [key: `aria-${string}`]: unknown;
  [key: `data-${string}`]: unknown;
};
type L<P> = ComponentType<P & Extra>;

export const Button = DSButton as unknown as L<ButtonProps>;
export const IconButton = DSIconButton as unknown as L<IconButtonProps>;
export const Icon = DSIcon as unknown as L<IconProps>;
export const Badge = DSBadge as unknown as L<BadgeProps>;
export const Card = DSCard as unknown as L<CardProps>;
export const Input = DSInput as unknown as L<InputProps>;
export const Select = DSSelect as unknown as L<SelectProps>;
export const Checkbox = DSCheckbox as unknown as L<CheckboxProps>;
export const Radio = DSRadio as unknown as L<RadioProps>;
export const Toast = DSToast as unknown as L<ToastProps>;
export const Stepper = DSStepper as unknown as L<StepperProps>;
export const ListingCard = DSListingCard as unknown as L<ListingCardProps>;
export const Price = DSPrice as unknown as L<PriceProps>;
export const PaymentMethodPicker = DSPaymentMethodPicker as unknown as L<PaymentMethodPickerProps>;
export const OrderStatus = DSOrderStatus as unknown as L<OrderStatusProps>;
export const AiSuggestion = DSAiSuggestion as unknown as L<AiSuggestionProps>;
export const AiProgress = DSAiProgress as unknown as L<AiProgressProps>;
// Các component dưới đây có props nhiều hoặc d.ts chưa đủ; dùng kiểu lỏng
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Any = ComponentType<any>;
export const AiBadge = DSAiBadge as unknown as Any;
export const ListingWizard = DSListingWizard as unknown as Any;
export const PhotoUploader = DSPhotoUploader as unknown as Any;
export const PublishConfirm = DSPublishConfirm as unknown as Any;
export const AiPriceRange = DSAiPriceRange as unknown as Any;
