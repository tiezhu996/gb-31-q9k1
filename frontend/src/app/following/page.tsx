'use client';
import { useQuery } from '@tanstack/react-query';
import { feedApi } from '@/api/feed';
import PostCard from '@/components/PostCard';
import EmptyState from '@/components/EmptyState';
import { usePagination } from '@/hooks/usePagination';

export default function FollowingPage() {
  const { page, pageSize, setPage } = usePagination(10);
  const { data, isLoading, isError } = useQuery({
    queryKey: ['feed', 'following', page, pageSize],
    queryFn: () => feedApi.following({ page, page_size: pageSize }),
    retry: 0,
  });

  return (
    <div className="mx-auto max-w-2xl">
      <h1 className="mb-4 text-xl font-bold">❤️ 关注</h1>
      {isLoading ? (
        <div className="py-20 text-center text-gray-400">加载中...</div>
      ) : isError ? (
        <EmptyState title="请先登录" description="登录后查看关注用户的动态" />
      ) : !data || data.items.length === 0 ? (
        <EmptyState title="关注列表为空" description="去发现页关注喜欢的铲屎官吧" />
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
