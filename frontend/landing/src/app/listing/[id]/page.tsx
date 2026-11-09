import type { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { categoryLabel, getListing } from '@/lib/api';
import { ListingDetail } from '@/components/ListingDetail';

interface Props {
  params: { id: string };
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const l = await getListing(params.id);
  return l ? { title: l.title, description: l.desc } : { title: 'Không tìm thấy tin' };
}

export default async function ListingPage({ params }: Props) {
  const l = await getListing(params.id);
  if (!l) notFound();
  return <ListingDetail listing={l} categoryLabel={categoryLabel(l.cat)} />;
}
