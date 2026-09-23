'use client';
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { meetupApi } from '@/api/meetup';
import MeetupCard from '@/components/MeetupCard';
import EmptyState from '@/components/EmptyState';
import { usePagination } from '@/hooks/usePagination';

const CITIES = ['全部', '上海', '北京', '广州', '深圳', '杭州'];

export default function MeetupsPage() {
  const [city, setCity] = useState('全部');
  const [showCreate, setShowCreate] = useState(false);
  const [error, setError] = useState('');
  const [creating, setCreating] = useState(false);
  const { page, pageSize, setPage } = usePagination(10);
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['meetups', city, page, pageSize],
    queryFn: () => meetupApi.list({ city: city === '全部' ? '' : city, page, page_size: pageSize }),
  });
  const [form, setForm] = useState({
    title: '',
    description: '',
    city: '上海',
    location: '',
    meet_time: '',
    max_people: 5,
  });

  const create = async () => {
    setError('');
    if (!form.title || !form.location || !form.meet_time) {
      setError('请填写标题、地点和约伴时间');
      return;
    }
    setCreating(true);
    try {
      await meetupApi.create(form);
      setShowCreate(false);
      setForm({ title: '', description: '', city: '上海', location: '', meet_time: '', max_people: 5 });
      refetch();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="mx-auto max-w-3xl">
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-bold">🐕 同城遛狗搭子</h1>
        <button className="rounded-lg bg-brand-500 px-4 py-2 text-sm text-white hover:bg-brand-600" onClick={() => setShowCreate(true)}>
          + 发布约伴
        </button>
      </div>
      <div className="mb-4 flex flex-wrap gap-2">
        {CITIES.map((c) => (
          <button
            key={c}
            className={`rounded-full px-3 py-1 text-sm ${city === c ? 'bg-brand-500 text-white' : 'border border-gray-200 bg-white text-gray-600'}`}
            onClick={() => { setCity(c); setPage(1); }}
          >
            {c}
          </button>
        ))}
      </div>
      {isLoading ? (
        <div className="py-20 text-center text-gray-400">加载中...</div>
      ) : !data || data.items.length === 0 ? (
        <EmptyState title="暂无约伴帖" description="发布一个约伴，找同城遛狗搭子" />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          {data.items.map((m) => (
            <MeetupCard key={m.id} meetup={m} />
          ))}
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-6">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-lg font-semibold">发布约伴帖</h2>
              <button className="text-gray-400" onClick={() => setShowCreate(false)}>✕</button>
            </div>
            <div className="space-y-3">
              <div>
                <label className="text-sm text-gray-600">标题 *</label>
                <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="周六世纪公园遛狗局" />
              </div>
              <div>
                <label className="text-sm text-gray-600">描述</label>
                <textarea className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" rows={3} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-sm text-gray-600">城市 *</label>
                  <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.city} onChange={(e) => setForm({ ...form, city: e.target.value })} />
                </div>
                <div>
                  <label className="text-sm text-gray-600">地点 *</label>
                  <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.location} onChange={(e) => setForm({ ...form, location: e.target.value })} />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-sm text-gray-600">约伴时间 *（2006-01-02 15:04）</label>
                  <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.meet_time} onChange={(e) => setForm({ ...form, meet_time: e.target.value })} placeholder="2026-08-20 10:00" />
                </div>
                <div>
                  <label className="text-sm text-gray-600">人数上限</label>
                  <input type="number" min={2} max={50} className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.max_people} onChange={(e) => setForm({ ...form, max_people: Number(e.target.value) })} />
                </div>
              </div>
              {error && <div className="text-sm text-red-500">{error}</div>}
              <button className="w-full rounded-lg bg-brand-500 py-2 text-white disabled:opacity-50" onClick={create} disabled={creating}>
                {creating ? '发布中...' : '发布约伴'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
