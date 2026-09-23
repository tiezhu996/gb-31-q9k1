'use client';
import { create } from 'zustand';
import type { Post } from '@/types';

interface PostState {
  posts: Post[];
  setPosts: (posts: Post[]) => void;
  updatePost: (id: string, patch: Partial<Post>) => void;
}

export const usePostStore = create<PostState>((set) => ({
  posts: [],
  setPosts: (posts) => set({ posts }),
  updatePost: (id, patch) =>
    set((s) => ({ posts: s.posts.map((p) => (p.id === id ? { ...p, ...patch } : p)) })),
}));
