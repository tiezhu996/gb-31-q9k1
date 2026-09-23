'use client';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import type { Meetup } from '@/types';
import { meetupApi } from '@/api/meetup';
import StatusBadge from './StatusBadge';

function formatDuration(minutes: number): string {
  if (minutes % 60 === 0) return `${minutes / 60} 小时`;
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return h > 0 ? `${h} 小时 ${m} 分钟` : `${m} 分钟`;
}

export default function MeetupCard({ meetup }: { meetup: Meetup }) {
  const qc = useQueryClient();
  const [error, setError] = useState('');
  const invalidate = () => qc.invalidateQueries({ queryKey: ['meetups'] });

  const joinMut = useMutation({
    mutationFn: () => meetupApi.join(meetup.id),
    onSuccess: invalidate,
    onError: (e) => setError((e as Error).message),
  });
  const cancelMut = useMutation({
    mutationFn: () => meetupApi.cancelJoin(meetup.id),
    onSuccess: invalidate,
    onError: (e) => setError((e as Error).message),
  });

  const isConflict = error.includes('冲突');
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
        🕒 {meetup.meet_time} ~ {meetup.end_time}（时长 {formatDuration(meetup.duration_minutes)}）
      </div>
      <p className="mt-2 text-sm text-gray-700">{meetup.description}</p>
      <div className="mt-3 flex items-center justify-between text-sm">
        <span className="text-gray-500">
          {meetup.joined_count}/{meetup.max_people} 人已报名 · 发起人 {meetup.creator_name}
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
        {isFull && <span className="text-xs text-blue-600">名额已满，可蹲候补</span>}
      </div>
      {error && (
        <div className={`mt-2 text-xs ${isConflict ? 'rounded-lg bg-amber-50 p-2 text-amber-700' : 'text-red-500'}`}>
          {isConflict && <span className="mr-1">⏰</span>}
          {error}
        </div>
      )}
    </div>
  );
}
