import { http } from '@/utils/request';
import type { PageData, Post } from '@/types';

export const feedApi = {
  discover: (params?: { page?: number; page_size?: number }) =>
    http.get<PageData<Post>>('/feed/discover', params),
  nearby: (params?: { city?: string; page?: number; page_size?: number }) =>
    http.get<PageData<Post>>('/feed/nearby', params),
  following: (params?: { page?: number; page_size?: number }) =>
    http.get<PageData<Post>>('/feed/following', params),
};
