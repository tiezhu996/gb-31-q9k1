'use client';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';

const NAV = [
  { href: '/', label: '发现' },
  { href: '/following', label: '关注' },
  { href: '/nearby', label: '同城' },
  { href: '/pets', label: '宠物' },
  { href: '/meetups', label: '遛狗搭子' },
  { href: '/topics', label: '话题' },
  { href: '/chat', label: '私信' },
];

export default function Navbar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user, isLogin, isAdmin, hydrated, logout } = useAuth();

  return (
    <header className="sticky top-0 z-40 border-b border-orange-100 bg-white/90 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
        <Link href="/" className="flex items-center gap-2 text-lg font-bold text-brand-600">
          <span className="text-2xl">🐶</span> 萌宠社区
        </Link>
        <nav className="hidden items-center gap-1 md:flex">
          {NAV.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={`rounded-lg px-3 py-1.5 text-sm ${
                pathname === item.href ? 'bg-orange-100 font-medium text-brand-700' : 'text-gray-600 hover:bg-orange-50'
              }`}
            >
              {item.label}
            </Link>
          ))}
          {isAdmin && (
            <Link
              href="/admin/audit"
              className={`rounded-lg px-3 py-1.5 text-sm ${
                pathname.startsWith('/admin') ? 'bg-orange-100 font-medium text-brand-700' : 'text-gray-600 hover:bg-orange-50'
              }`}
            >
              审计
            </Link>
          )}
        </nav>
        <div className="flex items-center gap-2">
          {!hydrated ? null : isLogin ? (
            <>
              <Link href="/posts/new" className="rounded-lg bg-brand-500 px-3 py-1.5 text-sm text-white hover:bg-brand-600">
                发布
              </Link>
              <span className="hidden text-sm text-gray-600 sm:inline">{user?.nickname}</span>
              <button
                className="rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50"
                onClick={() => {
                  logout();
                  router.push('/login');
                }}
              >
                退出
              </button>
            </>
          ) : (
            <>
              <Link href="/login" className="rounded-lg px-3 py-1.5 text-sm text-gray-600 hover:bg-orange-50">
                登录
              </Link>
              <Link href="/register" className="rounded-lg bg-brand-500 px-3 py-1.5 text-sm text-white hover:bg-brand-600">
                注册
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
