'use client';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import Link from 'next/link';
import { userApi } from '@/api/user';
import { useAuth } from '@/hooks/useAuth';
import { setToken } from '@/utils/request';

export default function LoginPage() {
  const router = useRouter();
  const { setUser } = useAuth();
  const [form, setForm] = useState({ username: 'demo', password: 'demo123' });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const submit = async () => {
    setError('');
    setLoading(true);
    try {
      const data = await userApi.login(form);
      setToken(data.token);
      setUser(data.user);
      router.push('/');
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-md">
      <div className="rounded-2xl border border-orange-100 bg-white p-8 shadow-sm">
        <h1 className="text-center text-2xl font-bold text-brand-600">🐾 登录萌宠社区</h1>
        <p className="mt-2 text-center text-sm text-gray-500">演示账号 demo / demo123，管理员 admin / admin123</p>
        <div className="mt-6 space-y-4">
          <div>
            <label className="text-sm text-gray-600">用户名</label>
            <input
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 focus:border-brand-500 focus:outline-none"
              value={form.username}
              onChange={(e) => setForm({ ...form, username: e.target.value })}
            />
          </div>
          <div>
            <label className="text-sm text-gray-600">密码</label>
            <input
              type="password"
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 focus:border-brand-500 focus:outline-none"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
            />
          </div>
          {error && <div className="text-sm text-red-500">{error}</div>}
          <button
            className="w-full rounded-lg bg-brand-500 py-2 font-medium text-white hover:bg-brand-600 disabled:opacity-50"
            onClick={submit}
            disabled={loading}
          >
            {loading ? '登录中...' : '登录'}
          </button>
          <div className="text-center text-sm text-gray-500">
            还没有账号？<Link href="/register" className="text-brand-600">立即注册</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
