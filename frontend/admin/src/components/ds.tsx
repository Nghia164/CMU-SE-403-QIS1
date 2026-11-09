'use client';
/** Cổng duy nhất để dùng MASM Design System trong Admin. Kiểu "lỏng" vì d.ts của DS chưa nhận style/aria. */
import type { ComponentType } from 'react';
import { Button as B } from '@ds/components/core/Button';
import { IconButton as IB } from '@ds/components/core/IconButton';
import { Icon as I } from '@ds/components/core/Icon';
import { Badge as Bd } from '@ds/components/core/Badge';
import { Card as C } from '@ds/components/core/Card';
import { Input as In } from '@ds/components/forms/Input';
import { Radio as R } from '@ds/components/forms/Radio';
import { Dialog as D } from '@ds/components/feedback/Dialog';
import { Toast as T } from '@ds/components/feedback/Toast';
import { Tabs as Tb } from '@ds/components/navigation/Tabs';
import { DataTable as DT } from '@ds/components/admin/DataTable';
import { StatCard as SC } from '@ds/components/admin/StatCard';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Any = ComponentType<any>;
export const Button = B as unknown as Any;
export const IconButton = IB as unknown as Any;
export const Icon = I as unknown as Any;
export const Badge = Bd as unknown as Any;
export const Card = C as unknown as Any;
export const Input = In as unknown as Any;
export const Radio = R as unknown as Any;
export const Dialog = D as unknown as Any;
export const Toast = T as unknown as Any;
export const Tabs = Tb as unknown as Any;
export const DataTable = DT as unknown as Any;
export const StatCard = SC as unknown as Any;
