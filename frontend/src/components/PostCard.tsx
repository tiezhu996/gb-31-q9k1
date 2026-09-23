'use client';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { Post } from '@/types';
import { postApi } from '@/api/post';
import StatusBadge from './StatusBadge';
import { formatCount, truncate } from '@/utils/format';

export default function PostCard({ post }: { post: Post }) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [error, setError] = useState('');

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['posts'] });
    queryClient.invalidateQueries({ queryKey: ['feed'] });
  };

  const likeMut = useMutation({
    mutationFn: () => postApi.interact(post.id, 'like'),
    onSuccess: invalidate,
    onError: (e) => setError((e as Error).message),
  });
  const favMut = useMutation({
    mutationFn: () => postApi.interact(post.id, 'favorite'),
    onSuccess: invalidate,
    onError: (e) => setError((e as Error).message),
  });

  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
      <div className="flex items-center gap-3">
        <Link href={`/users/${post.author_id}`} className="flex items-center gap-3">
          {post.author_avatar ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={post.author_avatar} alt={post.author_name} className="h-10 w-10 rounded-full object-cover" />
          ) : (
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-orange-100 text-lg">🐾</div>
          )}
          <div>
            <div className="text-sm font-semibold">{post.author_name}</div>
            <div className="text-xs text-gray-400">{post.created_at}</div>
          </div>
        </Link>
        <div className="ml-auto">
          <StatusBadge kind="post" value={post.status} />
          <StatusBadge kind="type" value={post.type} />
        </div>
      </div>
      <p className="mt-3 whitespace-pre-wrap text-sm leading-relaxed text-gray-800">{post.content}</p>
      {post.media && post.media.length > 0 && (
        <div className="mt-3 grid grid-cols-2 gap-2">
          {post.media.slice(0, 4).map((m, i) => (
            <div key={i} className="aspect-square overflow-hidden rounded-xl bg-gray-100">
              {m.type === 'video' ? (
                <video src={m.url} className="h-full w-full object-cover" controls />
              ) : (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={m.url} alt="media" className="h-full w-full object-cover" />
              )}
            </div>
          ))}
        </div>
      )}
      <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-gray-500">
        <span>📍 {post.location || '未填写地点'}</span>
        {post.topics?.map((t) => (
          <Link key={t} href={`/topics?name=${encodeURIComponent(t)}`} className="text-brand-600 hover:underline">
            #{t}
          </Link>
        ))}
      </div>
      <div className="mt-3 flex items-center gap-4 border-t border-gray-50 pt-3 text-sm">
        <button
          className={`flex items-center gap-1 ${post.liked ? 'text-brand-600' : 'text-gray-500 hover:text-brand-600'}`}
          onClick={() => likeMut.mutate()}
        >
          👍 {formatCount(post.like_count)}
        </button>
        <button
          className={`flex items-center gap-1 ${post.favorited ? 'text-amber-500' : 'text-gray-500 hover:text-amber-500'}`}
          onClick={() => favMut.mutate()}
        >
          ⭐ {formatCount(post.favorite_count)}
        </button>
        <button className="flex items-center gap-1 text-gray-500 hover:text-brand-600" onClick={() => router.push(`/posts/${post.id}`)}>
          💬 {formatCount(post.comment_count)}
        </button>
        <button
          className="ml-auto text-gray-500 hover:text-brand-600"
          onClick={() => {
            navigator.clipboard?.writeText(`${window.location.origin}/posts/${post.id}`).catch(() => undefined);
            postApi.interact(post.id, 'forward').then(invalidate).catch(() => undefined);
          }}
        >
          🔁 {formatCount(post.forward_count)}
        </button>
      </div>
      {error && <div className="mt-2 text-xs text-red-500">{error}</div>}
      <div className="mt-2 text-xs text-gray-400">
        内容摘要：{truncate(post.content, 24)}
      </div>
    </div>
  );
}
