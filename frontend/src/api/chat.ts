import { http } from '@/utils/request';
import type { ChatMessage, Conversation, PageData } from '@/types';

export const chatApi = {
  conversations: () => http.get<Conversation[]>('/chat/conversations'),
  conversation: (peerId: string, params?: { page?: number; page_size?: number }) =>
    http.get<PageData<ChatMessage>>(`/chat/conversations/${peerId}`, params),
  send: (data: { to_id: string; type: string; content: string }) =>
    http.post<ChatMessage>('/chat/messages', data),
};
