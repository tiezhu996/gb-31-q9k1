'use client';
import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { useState } from 'react';
import { useParams } from 'next/navigation';
import { userApi } from '@/api/user';
import { postApi } from '@/api/post';
import PostCard from '@/components/PostCard';
import EmptyState from '@/components/EmptyState';
import { useAuth } from '@/hooks/useAuth';
import { usePagination } from '@/hooks/usePagination';

export default function UserProfilePage() {
  const params = useParams<{ id: string }>();
  const { user } = useAuth();
  const { page, pageSize } = usePagination(10);
  const { data: profile, isLoading: loadingProfile } = useQuery({
    queryKey: ['user', params.id],
    queryFn: () => userApi.getUser(params.id),
  });
  const { data: posts } = useQuery({
    queryKey: ['user-posts', params.id, page, pageSize],
    queryFn: () => postApi.listByAuthor(params.id, { page, page_size: pageSize }),
  });

  if (loadingProfile) return <div className="py-20 text-center text-gray-400">加载中...</div>;
  if (!profile) return <EmptyState title="用户不存在" />;

  return (
    <div className="mx-auto max-w-2xl">
      <Link href="/" className="text-sm text-brand-600 hover:underline">← 返回</Link>
      <div className="mt-4 rounded-2xl border border-orange-100 bg-white p-6 shadow-sm">
        <div className="flex items-center gap-4">
          {profile.avatar ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={profile.avatar} alt={profile.nickname} className="h-20 w-20 rounded-full object-cover" />
          ) : (
            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-orange-100 text-3xl">🐾</div>
          )}
          <div>
            <div className="flex items-center gap-2 text-xl font-bold">
              {profile.nickname}
              {profile.role === 'admin' && <span className="rounded-full bg-purple-100 px-2 py-0.5 text-xs text-purple-700">管理员</span>}
            </div>
            <div className="mt-1 text-sm text-gray-500">@{profile.username} · 📍 {profile.city}</div>
            <div className="mt-1 text-sm text-gray-500">{profile.bio || '这个人很懒，什么都没写~'}</div>
          </div>
          <div className="ml-auto flex gap-6 text-center">
            <div>
              <div className="font-bold">{profile.follow_count}</div>
              <div className="text-xs text-gray-400">关注</div>
            </div>
            <div>
              <div className="font-bold">{profile.follower_count}</div>
              <div className="text-xs text-gray-400">粉丝</div>
            </div>
          </div>
        </div>
        {user && user.id !== profile.id && (
          <div className="mt-4 flex gap-2">
            <FollowButton targetId={profile.id} />
            <Link href={`/chat?peer=${profile.id}`} className="rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-600">
              发私信
            </Link>
          </div>
        )}
      </div>
      <h2 className="mb-3 mt-6 font-semibold">📝 Ta 的动态 ({posts?.total ?? 0})</h2>
      {!posts || posts.items.length === 0 ? (
        <EmptyState title="还没有动态" />
      ) : (
        <div className="space-y-4">
          {posts.items.map((post) => (
            <PostCard key={post.id} post={post} />
          ))}
        </div>
      )}
    </div>
  );
}

function FollowButton({ targetId }: { targetId: string }) {
  const [loading, setLoading] = useState(false);
  return (
    <button
      className="rounded-lg bg-brand-500 px-4 py-1.5 text-sm text-white hover:bg-brand-600 disabled:opacity-50"
      disabled={loading}
      onClick={async () => {
        setLoading(true);
        try {
          await userApi.follow(targetId);
          window.location.reload();
        } catch {
          window.alert('请先登录');
        } finally {
          setLoading(false);
        }
      }}
    >
      {loading ? '关注中...' : '关注'}
    </button>
  );
}
