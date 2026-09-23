'use client';
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { feedApi } from '@/api/feed';
import PostCard from '@/components/PostCard';
import EmptyState from '@/components/EmptyState';
import { usePagination } from '@/hooks/usePagination';

const CITIES = ['上海', '北京', '广州', '深圳', '杭州'];

export default function NearbyPage() {
  const [city, setCity] = useState('上海');
  const { page, pageSize, setPage } = usePagination(10);
  const { data, isLoading } = useQuery({
    queryKey: ['feed', 'nearby', city, page, pageSize],
    queryFn: () => feedApi.nearby({ city, page, page_size: pageSize }),
  });

  return (
    <div className="mx-auto max-w-2xl">
      <h1 className="mb-4 text-xl font-bold">📍 同城</h1>
      <div className="mb-4 flex flex-wrap gap-2">
        {CITIES.map((c) => (
          <button
            key={c}
            className={`rounded-full px-3 py-1 text-sm ${city === c ? 'bg-brand-500 text-white' : 'bg-white text-gray-600 border border-gray-200'}`}
            onClick={() => { setCity(c); setPage(1); }}
          >
            {c}
          </button>
        ))}
      </div>
      {isLoading ? (
        <div className="py-20 text-center text-gray-400">加载中...</div>
      ) : !data || data.items.length === 0 ? (
        <EmptyState title={`${city} 还没有动态`} description="成为第一个发同城动态的人吧" />
      ) : (
        <div className="space-y-4">
          {data.items.map((post) => (
            <PostCard key={post.id} post={post} />
          ))}
          {data.total > page * pageSize && (
            <button className="w-full rounded-xl border border-orange-200 py-2 text-sm text-brand-600" onClick={() => setPage(page + 1)}>
              加载更多
            </button>
          )}
        </div>
      )}
    </div>
  );
}
