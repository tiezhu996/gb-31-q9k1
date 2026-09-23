import { http } from '@/utils/request';
import type { Pet, PageData } from '@/types';

export const petApi = {
  list: (params?: { species?: string; city?: string; page?: number; page_size?: number }) =>
    http.get<PageData<Pet>>('/pets', params),
  get: (id: string) => http.get<Pet>(`/pets/${id}`),
  mine: (params?: { page?: number; page_size?: number }) =>
    http.get<PageData<Pet>>('/pets/me', params),
  create: (data: Partial<Pet>) => http.post<Pet>('/pets', data),
  update: (id: string, data: Partial<Pet>) => http.put<Pet>(`/pets/${id}`, data),
  remove: (id: string) => http.delete<{ message: string }>(`/pets/${id}`),
  follow: (id: string) => http.post<{ message: string }>(`/pets/${id}/follow`),
  unfollow: (id: string) => http.delete<{ message: string }>(`/pets/${id}/follow`),
};
