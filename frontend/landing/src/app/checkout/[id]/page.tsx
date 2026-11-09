import type { Metadata } from 'next';
import { notFound, redirect } from 'next/navigation';
import { getListing } from '@/lib/api';
import { CheckoutForm } from '@/components/CheckoutForm';

export const metadata: Metadata = { title: 'Thanh toán' };

export default async function CheckoutPage({ params }: { params: { id: string } }) {
  const l = await getListing(params.id);
  if (!l) notFound();
  if (l.sold) redirect(`/listing/${l.id}`);
  return <CheckoutForm listing={l} />;
}
