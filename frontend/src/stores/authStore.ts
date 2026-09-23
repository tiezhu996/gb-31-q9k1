'use client';
import { create } from 'zustand';
import type { User } from '@/types';
import { loadUser, saveUser, clearUser } from '@/utils/auth';
import { clearToken } from '@/utils/request';

interface AuthState {
  user: User | null;
  hydrated: boolean;
  hydrate: () => void;
  setUser: (user: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  hydrated: false,
  // 客户端挂载后再读取 localStorage，避免 SSR 水合不一致
  hydrate: () => {
    if (typeof window === 'undefined') return;
    set({ user: loadUser(), hydrated: true });
  },
  setUser: (user) => {
    saveUser(user);
    set({ user, hydrated: true });
  },
  logout: () => {
    clearUser();
    clearToken();
    set({ user: null });
  },
}));
