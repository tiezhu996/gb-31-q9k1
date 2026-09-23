import { http } from '@/utils/request';
import type { TokenResponse, User, PageData } from '@/types';

export const userApi = {
  register: (data: { username: string; password: string; nickname: string; city?: string }) =>
    http.post<TokenResponse>('/auth/register', data),
  login: (data: { username: string; password: string }) => http.post<TokenResponse>('/auth/login', data),
  me: () => http.get<User>('/users/me'),
  getUser: (id: string) => http.get<User>(`/users/${id}`),
  updateProfile: (data: Partial<User>) => http.put<User>('/users/me', data),
  follow: (followee_id: string) => http.post<{ message: string }>('/users/me/follow', { followee_id }),
  unfollow: (id: string) => http.delete<{ message: string }>(`/users/me/follow/${id}`),
  following: (params?: { page?: number; page_size?: number }) =>
    http.get<PageData<User>>('/users/me/following', params),
  followers: (params?: { page?: number; page_size?: number }) =>
    http.get<PageData<User>>('/users/me/followers', params),
};
