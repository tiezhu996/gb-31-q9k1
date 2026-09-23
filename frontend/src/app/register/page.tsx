'use client';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import Link from 'next/link';
import { userApi } from '@/api/user';
import { useAuth } from '@/hooks/useAuth';
import { setToken } from '@/utils/request';

export default function RegisterPage() {
  const router = useRouter();
  const { setUser } = useAuth();
  const [form, setForm] = useState({ username: '', password: '', nickname: '', city: '上海' });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const submit = async () => {
    setError('');
    if (!form.username || !form.password || !form.nickname) {
      setError('请填写用户名、密码和昵称');
      return;
    }
    setLoading(true);
    try {
      const data = await userApi.register(form);
      setToken(data.token);
      setUser(data.user);
      router.push('/pets');
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-md">
      <div className="rounded-2xl border border-orange-100 bg-white p-8 shadow-sm">
        <h1 className="text-center text-2xl font-bold text-brand-600">🐾 注册萌宠社区</h1>
        <div className="mt-6 space-y-4">
          <div>
            <label className="text-sm text-gray-600">用户名（≥3 位）</label>
            <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} />
          </div>
          <div>
            <label className="text-sm text-gray-600">昵称</label>
            <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.nickname} onChange={(e) => setForm({ ...form, nickname: e.target.value })} />
          </div>
          <div>
            <label className="text-sm text-gray-600">密码（≥6 位）</label>
            <input type="password" className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} />
          </div>
          <div>
            <label className="text-sm text-gray-600">城市</label>
            <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.city} onChange={(e) => setForm({ ...form, city: e.target.value })} />
          </div>
          {error && <div className="text-sm text-red-500">{error}</div>}
          <button className="w-full rounded-lg bg-brand-500 py-2 font-medium text-white hover:bg-brand-600 disabled:opacity-50" onClick={submit} disabled={loading}>
            {loading ? '注册中...' : '注册'}
          </button>
          <div className="text-center text-sm text-gray-500">
            已有账号？<Link href="/login" className="text-brand-600">去登录</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
