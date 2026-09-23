'use client';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import type { Meetup } from '@/types';
import { meetupApi } from '@/api/meetup';
import StatusBadge from './StatusBadge';

// 卡片内展示活动时长；后端缺省 120 分钟，兼容历史数据。
function formatDuration(minutes: number): string {
  const m = minutes > 0 ? minutes : 120;
  if (m % 60 === 0) return `${m / 60} 小时`;
  const hours = Math.floor(m / 60);
  const rest = m % 60;
  return hours > 0 ? `${hours} 小时 ${rest} 分钟` : `${rest} 分钟`;
}

export default function MeetupCard({ meetup }: { meetup: Meetup }) {
  const qc = useQueryClient();
  const [error, setError] = useState('');
  const invalidate = () => qc.invalidateQueries({ queryKey: ['meetups'] });

  const joinMut = useMutation({
    mutationFn: () => meetupApi.join(meetup.id),
    onSuccess: invalidate,
    onError: (e) => {
      setError((e as Error).message);
      invalidate();
    },
  });
  const cancelMut = useMutation({
    mutationFn: () => meetupApi.cancelJoin(meetup.id),
    onSuccess: invalidate,
    onError: (e) => setError((e as Error).message),
  });

  const isFull = meetup.status === 'full';

  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
      <div className="flex items-center justify-between">
        <span className="font-semibold">{meetup.title}</span>
        <StatusBadge kind="meetup" value={meetup.status} />
      </div>
      <div className="mt-2 text-sm text-gray-500">
        📍 {meetup.city} · {meetup.location}
      </div>
      <div className="mt-1 text-sm text-gray-500">
        🕒 {meetup.meet_time} ~ {meetup.end_time}
        <span className="ml-2 rounded-full bg-orange-50 px-2 py-0.5 text-xs text-orange-600">
          时长 {formatDuration(meetup.duration_minutes)}
        </span>
      </div>
      <p className="mt-2 text-sm text-gray-700">{meetup.description}</p>
      <div className="mt-3 flex items-center justify-between text-sm">
        <span className="text-gray-500">
          {meetup.joined_count}/{meetup.max_people} 人已报名 · 发起人 {meetup.creator_name}
          {isFull && <span className="ml-1 text-red-500">（名额已满）</span>}
        </span>
        {meetup.status === 'open' &&
          (meetup.joined ? (
            <button className="rounded-lg border border-gray-300 px-3 py-1 text-xs" onClick={() => cancelMut.mutate()}>
              取消报名
            </button>
          ) : (
            <button className="rounded-lg bg-brand-500 px-3 py-1 text-xs text-white" onClick={() => joinMut.mutate()}>
              报名参加
            </button>
          ))}
      </div>
      {error && <div className="mt-2 rounded-lg bg-red-50 px-2 py-1 text-xs text-red-600">{error}</div>}
    </div>
  );
}
