'use client';
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { petApi } from '@/api/pet';
import PetCard from '@/components/PetCard';
import EmptyState from '@/components/EmptyState';
import ConfirmDialog from '@/components/ConfirmDialog';
import { usePagination } from '@/hooks/usePagination';
import { PET_SPECIES, PET_SPECIES_LABEL, PET_GENDER } from '@/constants/enums';
import type { Pet, PetSpecies, PetGender } from '@/types';

export default function PetsPage() {
  const [showCreate, setShowCreate] = useState(false);
  const [confirmDel, setConfirmDel] = useState<string | null>(null);
  const { page, pageSize, setPage } = usePagination(12);
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['pets', page, pageSize],
    queryFn: () => petApi.list({ page, page_size: pageSize }),
  });
  const [form, setForm] = useState<{ name: string; species: PetSpecies; breed: string; birthday: string; gender: PetGender; personality: string; bio: string; city: string; avatar: string }>({
    name: '',
    species: 'dog',
    breed: '',
    birthday: '2022-06-01',
    gender: 'male',
    personality: '',
    bio: '',
    city: '上海',
    avatar: '',
  });
  const [error, setError] = useState('');
  const [creating, setCreating] = useState(false);

  const create = async () => {
    setError('');
    if (!form.name) {
      setError('请填写宠物名字');
      return;
    }
    setCreating(true);
    try {
      await petApi.create(form as Partial<Pet>);
      setShowCreate(false);
      setForm({ name: '', species: 'dog', breed: '', birthday: '2022-06-01', gender: 'male', personality: '', bio: '', city: '上海', avatar: '' });
      refetch();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setCreating(false);
    }
  };

  const remove = async () => {
    if (!confirmDel) return;
    try {
      await petApi.remove(confirmDel);
      setConfirmDel(null);
      refetch();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  return (
    <div className="mx-auto max-w-4xl">
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-bold">🐱 宠物广场</h1>
        <button className="rounded-lg bg-brand-500 px-4 py-2 text-sm text-white hover:bg-brand-600" onClick={() => setShowCreate(true)}>
          + 创建宠物档案
        </button>
      </div>
      {isLoading ? (
        <div className="py-20 text-center text-gray-400">加载中...</div>
      ) : !data || data.items.length === 0 ? (
        <EmptyState title="还没有宠物档案" description="为你的毛孩子创建专属主页吧" />
      ) : (
        <>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {data.items.map((pet) => (
              <div key={pet.id} className="relative">
                <PetCard pet={pet} />
              </div>
            ))}
          </div>
          {data.total > page * pageSize && (
            <button className="mt-4 w-full rounded-xl border border-orange-200 py-2 text-sm text-brand-600" onClick={() => setPage(page + 1)}>
              加载更多
            </button>
          )}
        </>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-6">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-lg font-semibold">创建宠物档案</h2>
              <button className="text-gray-400" onClick={() => setShowCreate(false)}>✕</button>
            </div>
            <div className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-sm text-gray-600">名字 *</label>
                  <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
                </div>
                <div>
                  <label className="text-sm text-gray-600">品种</label>
                  <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.breed} onChange={(e) => setForm({ ...form, breed: e.target.value })} />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-sm text-gray-600">物种</label>
                  <select className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.species} onChange={(e) => setForm({ ...form, species: e.target.value as PetSpecies })}>
                    {(Object.keys(PET_SPECIES) as Array<keyof typeof PET_SPECIES>).map((k) => (
                      <option key={k} value={PET_SPECIES[k]}>{PET_SPECIES_LABEL[PET_SPECIES[k]]}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="text-sm text-gray-600">性别</label>
                  <select className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.gender} onChange={(e) => setForm({ ...form, gender: e.target.value as PetGender })}>
                    <option value={PET_GENDER.MALE}>男孩</option>
                    <option value={PET_GENDER.FEMALE}>女孩</option>
                  </select>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-sm text-gray-600">生日</label>
                  <input type="date" className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.birthday} onChange={(e) => setForm({ ...form, birthday: e.target.value })} />
                </div>
                <div>
                  <label className="text-sm text-gray-600">城市</label>
                  <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.city} onChange={(e) => setForm({ ...form, city: e.target.value })} />
                </div>
              </div>
              <div>
                <label className="text-sm text-gray-600">性格</label>
                <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.personality} onChange={(e) => setForm({ ...form, personality: e.target.value })} />
              </div>
              <div>
                <label className="text-sm text-gray-600">简介</label>
                <textarea className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" rows={3} value={form.bio} onChange={(e) => setForm({ ...form, bio: e.target.value })} />
              </div>
              <div>
                <label className="text-sm text-gray-600">头像 URL</label>
                <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" placeholder="https://..." value={form.avatar} onChange={(e) => setForm({ ...form, avatar: e.target.value })} />
              </div>
              {error && <div className="text-sm text-red-500">{error}</div>}
              <button className="w-full rounded-lg bg-brand-500 py-2 text-white disabled:opacity-50" onClick={create} disabled={creating}>
                {creating ? '创建中...' : '创建档案'}
              </button>
            </div>
          </div>
        </div>
      )}
      <ConfirmDialog open={Boolean(confirmDel)} title="删除宠物档案" content="确定删除该宠物档案吗？该操作不可恢复。" onCancel={() => setConfirmDel(null)} onConfirm={remove} />
    </div>
  );
}
