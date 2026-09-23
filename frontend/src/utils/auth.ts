// 登录态与角色工具：前端路由守卫与按钮显隐依据
import type { User } from '@/types';
import { getToken } from './request';

const USER_KEY = 'petsocial_user';

export function saveUser(user: User) {
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(USER_KEY, JSON.stringify(user));
  }
}

export function loadUser(): User | null {
  if (typeof window === 'undefined') return null;
  const raw = window.localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as User;
  } catch {
    return null;
  }
}

export function clearUser() {
  if (typeof window !== 'undefined') {
    window.localStorage.removeItem(USER_KEY);
  }
}

export function isLoggedIn(): boolean {
  return Boolean(getToken());
}

export function isAdmin(user: User | null): boolean {
  return user?.role === 'admin';
}
