'use client';
import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { petApi } from '@/api/pet';
import StatusBadge from '@/components/StatusBadge';
import EmptyState from '@/components/EmptyState';
import { formatDate, formatGender } from '@/utils/format';
import { useAuth } from '@/hooks/useAuth';

export default function PetDetailPage() {
  const params = useParams<{ id: string }>();
  const { user } = useAuth();
  const { data: pet, isLoading } = useQuery({
    queryKey: ['pet', params.id],
    queryFn: () => petApi.get(params.id),
  });

  if (isLoading) return <div className="py-20 text-center text-gray-400">加载中...</div>;
  if (!pet) return <EmptyState title="宠物不存在" />;

  const isOwner = user?.id === pet.owner_id;

  return (
    <div className="mx-auto max-w-2xl">
      <Link href="/pets" className="text-sm text-brand-600 hover:underline">← 返回宠物广场</Link>
      <div className="mt-4 rounded-2xl border border-orange-100 bg-white p-6 shadow-sm">
        <div className="flex items-center gap-4">
          {pet.avatar ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={pet.avatar} alt={pet.name} className="h-24 w-24 rounded-full object-cover" />
          ) : (
            <div className="flex h-24 w-24 items-center justify-center rounded-full bg-orange-100 text-4xl">🐾</div>
          )}
          <div>
            <div className="flex items-center gap-2 text-xl font-bold">
              {pet.name}
              <StatusBadge kind="species" value={pet.species} />
            </div>
            <div className="mt-1 text-sm text-gray-500">
              {pet.breed || '未知品种'} · {formatGender(pet.gender)} · 生日 {formatDate(pet.birthday)}
            </div>
            <div className="mt-1 text-sm text-gray-500">📍 {pet.city} · 粉丝 {pet.follower_count}</div>
          </div>
          {isOwner && <div className="ml-auto text-xs text-gray-400">我的宠物</div>}
        </div>
        <div className="mt-4 grid grid-cols-2 gap-4 text-sm">
          <div className="rounded-xl bg-orange-50 p-3">
            <div className="text-gray-500">性格</div>
            <div className="mt-1">{pet.personality || '待补充'}</div>
          </div>
          <div className="rounded-xl bg-orange-50 p-3">
            <div className="text-gray-500">简介</div>
            <div className="mt-1">{pet.bio || '待补充'}</div>
          </div>
        </div>
        {pet.album && pet.album.length > 0 && (
          <div className="mt-4">
            <div className="mb-2 text-sm font-medium text-gray-600">📷 相册</div>
            <div className="grid grid-cols-3 gap-2">
              {pet.album.map((url, i) => (
                // eslint-disable-next-line @next/next/no-img-element
                <img key={i} src={url} alt={`album-${i}`} className="aspect-square w-full rounded-xl object-cover" />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
