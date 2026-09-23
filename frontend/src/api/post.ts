import { http } from '@/utils/request';
import type { Comment, PageData, Post, Topic } from '@/types';

export const postApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get<PageData<Post>>('/posts', params),
  get: (id: string) => http.get<Post>(`/posts/${id}`),
  create: (data: Partial<Post>) => http.post<Post>('/posts', data),
  listByAuthor: (userId: string, params?: { page?: number; page_size?: number }) =>
    http.get<PageData<Post>>(`/users/${userId}/posts`, params),
  listByTopic: (name: string, params?: { page?: number; page_size?: number }) =>
    http.get<PageData<Post>>(`/topics/${encodeURIComponent(name)}/posts`, params),
  listComments: (postId: string, params?: { page?: number; page_size?: number }) =>
    http.get<PageData<Comment>>(`/posts/${postId}/comments`, params),
  addComment: (postId: string, content: string) =>
    http.post<Comment>(`/posts/${postId}/comments`, { content }),
  deleteComment: (postId: string, commentId: string) =>
    http.delete<{ message: string }>(`/posts/${postId}/comments/${commentId}`),
  interact: (postId: string, type: 'like' | 'favorite' | 'forward') =>
    http.post<{ active: boolean; type: string }>(`/posts/${postId}/interact`, { type }),
  topics: (params?: { page?: number; page_size?: number }) =>
    http.get<PageData<Topic>>('/topics', params),
};
