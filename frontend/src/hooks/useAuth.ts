'use client';
import { useAuthStore } from '@/stores/authStore';

// 登录态 hook：路由守卫与按钮显隐统一引用
export function useAuth() {
  const user = useAuthStore((s) => s.user);
  const hydrated = useAuthStore((s) => s.hydrated);
  const setUser = useAuthStore((s) => s.setUser);
  const logout = useAuthStore((s) => s.logout);
  return { user, hydrated, isLogin: Boolean(user), isAdmin: user?.role === 'admin', setUser, logout };
}
