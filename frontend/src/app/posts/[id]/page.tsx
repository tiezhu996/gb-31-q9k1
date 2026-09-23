'use client';
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { postApi } from '@/api/post';
import StatusBadge from '@/components/StatusBadge';
import EmptyState from '@/components/EmptyState';
import { useAuth } from '@/hooks/useAuth';
import { usePagination } from '@/hooks/usePagination';

export default function PostDetailPage() {
  const params = useParams<{ id: string }>();
  const qc = useQueryClient();
  const { user, isLogin } = useAuth();
  const [comment, setComment] = useState('');
  const [error, setError] = useState('');
  const { page, pageSize } = usePagination(20);

  const { data: post, isLoading } = useQuery({
    queryKey: ['post', params.id],
    queryFn: () => postApi.get(params.id),
  });
  const { data: comments } = useQuery({
    queryKey: ['comments', params.id, page, pageSize],
    queryFn: () => postApi.listComments(params.id, { page, page_size: pageSize }),
  });

  const commentMut = useMutation({
    mutationFn: () => postApi.addComment(params.id, comment),
    onSuccess: () => {
      setComment('');
      qc.invalidateQueries({ queryKey: ['comments', params.id] });
      qc.invalidateQueries({ queryKey: ['post', params.id] });
    },
    onError: (e) => setError((e as Error).message),
  });

  if (isLoading) return <div className="py-20 text-center text-gray-400">加载中...</div>;
  if (!post) return <EmptyState title="动态不存在" />;

  return (
    <div className="mx-auto max-w-2xl">
      <Link href="/" className="text-sm text-brand-600 hover:underline">← 返回发现页</Link>
      <div className="mt-4 rounded-2xl border border-orange-100 bg-white p-6 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-orange-100 text-lg">
            {post.author_avatar ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={post.author_avatar} alt={post.author_name} className="h-10 w-10 rounded-full object-cover" />
            ) : '🐾'}
          </div>
          <div>
            <div className="text-sm font-semibold">{post.author_name}</div>
            <div className="text-xs text-gray-400">{post.created_at} · {post.city}</div>
          </div>
          <div className="ml-auto flex gap-2">
            <StatusBadge kind="post" value={post.status} />
            <StatusBadge kind="type" value={post.type} />
          </div>
        </div>
        <p className="mt-4 whitespace-pre-wrap text-gray-800">{post.content}</p>
        {post.media?.map((m, i) => (
          <div key={i} className="mt-3 overflow-hidden rounded-xl bg-gray-100">
            {m.type === 'video' ? (
              <video src={m.url} className="w-full" controls />
            ) : (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={m.url} alt="media" className="w-full object-contain" />
            )}
          </div>
        ))}
        <div className="mt-3 flex gap-2 text-xs text-gray-500">
          {post.topics?.map((t) => (
            <Link key={t} href={`/topics?name=${encodeURIComponent(t)}`} className="text-brand-600">#{t}</Link>
          ))}
        </div>
        <div className="mt-4 flex gap-6 border-t border-gray-100 pt-3 text-sm text-gray-500">
          <span>👍 {post.like_count}</span>
          <span>⭐ {post.favorite_count}</span>
          <span>💬 {post.comment_count}</span>
          <span>🔁 {post.forward_count}</span>
        </div>
      </div>

      <div className="mt-4 rounded-2xl border border-orange-100 bg-white p-6 shadow-sm">
        <h2 className="mb-3 font-semibold">💬 评论 ({comments?.total ?? 0})</h2>
        {isLogin ? (
          <div className="flex gap-2">
            <input
              className="flex-1 rounded-lg border border-gray-200 px-3 py-2"
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              placeholder="友善评论..."
            />
            <button className="rounded-lg bg-brand-500 px-4 text-white" onClick={() => commentMut.mutate()} disabled={!comment.trim()}>
              发送
            </button>
          </div>
        ) : (
          <div className="text-sm text-gray-400">
            请 <Link href="/login" className="text-brand-600">登录</Link> 后评论
          </div>
        )}
        {error && <div className="mt-2 text-xs text-red-500">{error}</div>}
        <div className="mt-4 space-y-3">
          {!comments || !comments.items || comments.items.length === 0 ? (
            <div className="py-4 text-center text-sm text-gray-400">还没有评论</div>
          ) : (
            comments.items.map((c) => (
              <div key={c.id} className="rounded-xl bg-orange-50/50 p-3">
                <div className="flex items-center gap-2 text-sm">
                  <span className="font-medium">{c.author_name}</span>
                  <span className="text-xs text-gray-400">{c.created_at}</span>
                  {(user?.id === c.author_id || user?.role === 'admin') && (
                    <button
                      className="ml-auto text-xs text-red-400"
                      onClick={() => postApi.deleteComment(post.id, c.id).then(() => qc.invalidateQueries({ queryKey: ['comments', post.id] })).catch(() => undefined)}
                    >
                      删除
                    </button>
                  )}
                </div>
                <p className="mt-1 text-sm text-gray-700">{c.content}</p>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
