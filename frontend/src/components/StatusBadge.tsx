'use client';
import {
  POST_STATUS_LABEL,
  MEETUP_STATUS_LABEL,
  USER_STATUS_LABEL,
  POST_TYPE_LABEL,
  PET_SPECIES_LABEL,
} from '@/constants/enums';

const COLOR_MAP: Record<string, string> = {
  approved: 'bg-green-100 text-green-700',
  pending: 'bg-amber-100 text-amber-700',
  rejected: 'bg-red-100 text-red-700',
  open: 'bg-green-100 text-green-700',
  full: 'bg-blue-100 text-blue-700',
  cancelled: 'bg-gray-200 text-gray-600',
  completed: 'bg-purple-100 text-purple-700',
  active: 'bg-green-100 text-green-700',
  banned: 'bg-red-100 text-red-700',
  image: 'bg-orange-100 text-orange-700',
  video: 'bg-indigo-100 text-indigo-700',
  dog: 'bg-amber-100 text-amber-700',
  cat: 'bg-pink-100 text-pink-700',
};

export function statusLabel(kind: 'post' | 'meetup' | 'user' | 'type' | 'species', value: string): string {
  switch (kind) {
    case 'post':
      return POST_STATUS_LABEL[value] || value;
    case 'meetup':
      return MEETUP_STATUS_LABEL[value] || value;
    case 'user':
      return USER_STATUS_LABEL[value] || value;
    case 'type':
      return POST_TYPE_LABEL[value] || value;
    case 'species':
      return PET_SPECIES_LABEL[value] || value;
    default:
      return value;
  }
}

export default function StatusBadge({ kind, value }: { kind: 'post' | 'meetup' | 'user' | 'type' | 'species'; value: string }) {
  return (
    <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${COLOR_MAP[value] || 'bg-gray-100 text-gray-600'}`}>
      {statusLabel(kind, value)}
    </span>
  );
}
