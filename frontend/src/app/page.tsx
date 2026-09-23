'use client';
import { useQuery } from '@tanstack/react-query';
import { feedApi } from '@/api/feed';
import PostCard from '@/components/PostCard';
import EmptyState from '@/components/EmptyState';
import { usePagination } from '@/hooks/usePagination';

export default function DiscoverPage() {
  const { page, pageSize, setPage } = usePagination(10);
  const { data, isLoading } = useQuery({
    queryKey: ['feed', 'discover', page, pageSize],
    queryFn: () => feedApi.discover({ page, page_size: pageSize }),
  });

  return (
    <div className="mx-auto max-w-2xl">
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-bold">🐾 发现</h1>
        <span className="text-sm text-gray-400">共 {data?.total ?? 0} 条动态</span>
      </div>
      {isLoading ? (
        <div className="py-20 text-center text-gray-400">加载中...</div>
      ) : !data || data.items.length === 0 ? (
        <EmptyState title="还没有动态" description="成为第一个晒娃的铲屎官吧" />
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
