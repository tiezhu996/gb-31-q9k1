'use client';
import Link from 'next/link';
import type { Pet } from '@/types';
import StatusBadge from './StatusBadge';
import { formatGender, formatDate } from '@/utils/format';

export default function PetCard({ pet }: { pet: Pet }) {
  return (
    <Link href={`/pets/${pet.id}`} className="block rounded-2xl border border-gray-100 bg-white p-4 shadow-sm transition hover:shadow-md">
      <div className="flex items-center gap-3">
        {pet.avatar ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={pet.avatar} alt={pet.name} className="h-14 w-14 rounded-full object-cover" />
        ) : (
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-orange-100 text-2xl">🐾</div>
        )}
        <div>
          <div className="flex items-center gap-2">
            <span className="font-semibold">{pet.name}</span>
            <StatusBadge kind="species" value={pet.species} />
          </div>
          <div className="text-xs text-gray-500">
            {pet.breed || '未知品种'} · {formatGender(pet.gender)} · {formatDate(pet.birthday)}
          </div>
        </div>
        <div className="ml-auto text-xs text-gray-400">粉丝 {pet.follower_count}</div>
      </div>
      <p className="mt-2 line-clamp-2 text-sm text-gray-600">{pet.bio || '这个毛孩子还没有自我介绍~'}</p>
    </Link>
  );
}
