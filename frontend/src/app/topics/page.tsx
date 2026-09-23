'use client';
import { Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { postApi } from '@/api/post';
import PostCard from '@/components/PostCard';
import EmptyState from '@/components/EmptyState';
import { usePagination } from '@/hooks/usePagination';

export const dynamic = 'force-dynamic';

export default function TopicsPage() {
  return (
    <Suspense fallback={<div className="py-20 text-center text-gray-400">加载中...</div>}>
      <TopicsContent />
    </Suspense>
  );
}

function TopicsContent() {
  const searchParams = useSearchParams();
  const activeTopic = searchParams.get('name') || '';
  const { page, pageSize } = usePagination(10);
  const { data: topics } = useQuery({
    queryKey: ['topics'],
    queryFn: () => postApi.topics({ page: 1, page_size: 50 }),
  });
  const { data: posts } = useQuery({
    queryKey: ['topic-posts', activeTopic, page, pageSize],
    queryFn: () => postApi.listByTopic(activeTopic, { page, page_size: pageSize }),
    enabled: Boolean(activeTopic),
  });

  return (
    <div className="mx-auto max-w-3xl">
      <h1 className="mb-4 text-xl font-bold"># 话题广场</h1>
      <div className="mb-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
        {topics?.items.map((t) => (
          <Link
            key={t.id}
            href={`/topics?name=${encodeURIComponent(t.name)}`}
            className={`rounded-2xl border p-4 text-center ${activeTopic === t.name ? 'border-brand-500 bg-orange-50' : 'border-gray-100 bg-white hover:shadow-sm'}`}
          >
            <div className="font-semibold text-brand-600">#{t.name}</div>
            <div className="mt-1 text-xs text-gray-400">{t.post_count} 篇参与</div>
          </Link>
        ))}
      </div>
      {activeTopic ? (
        <div className="space-y-4">
          <div className="text-sm text-gray-500">话题 #{activeTopic} 的动态</div>
          {!posts ? (
            <div className="py-20 text-center text-gray-400">加载中...</div>
          ) : posts.items.length === 0 ? (
            <EmptyState title="该话题还没有动态" description="快来参与话题吧" />
          ) : (
            posts.items.map((post) => <PostCard key={post.id} post={post} />)
          )}
        </div>
      ) : (
        <EmptyState title="选择一个话题" description="点击上方话题查看参与动态" />
      )}
    </div>
  );
}
