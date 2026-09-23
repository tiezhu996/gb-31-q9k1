'use client';
import { useEffect, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { chatApi } from '@/api/chat';
import EmptyState from '@/components/EmptyState';
import { useAuth } from '@/hooks/useAuth';
import { useWebSocket } from '@/hooks/useWebSocket';
import type { ChatMessage } from '@/types';

export default function ChatPage() {
  const { user, isLogin, hydrated } = useAuth();
  const qc = useQueryClient();
  const [peerId, setPeerId] = useState('');
  const [text, setText] = useState('');
  const [messages, setMessages] = useState<ChatMessage[]>([]);

  const { data: conversations, refetch: refetchConvs } = useQuery({
    queryKey: ['conversations'],
    queryFn: () => chatApi.conversations(),
    enabled: isLogin,
  });

  const { data: history, refetch: refetchHistory } = useQuery({
    queryKey: ['conversation', peerId],
    queryFn: () => chatApi.conversation(peerId, { page: 1, page_size: 50 }),
    enabled: Boolean(peerId),
  });

  useEffect(() => {
    if (history) setMessages(history.items);
  }, [history]);

  const { status, send } = useWebSocket((payload) => {
    const data = payload as { event?: string; data?: ChatMessage };
    if (data.event === 'message' && data.data) {
      const msg = data.data;
      if (peerId && (msg.from_id === peerId || msg.to_id === peerId)) {
        setMessages((prev) => [...prev, msg]);
      }
      refetchConvs();
    }
  });

  const sendMsg = async () => {
    const content = text.trim();
    if (!content || !peerId) return;
    try {
      await chatApi.send({ to_id: peerId, type: 'text', content });
      setText('');
      refetchHistory();
      refetchConvs();
    } catch {
      // 通过 websocket 重试发送
      send({ to_id: peerId, msg_type: 'text', content });
      setText('');
    }
  };

  if (hydrated && !isLogin) {
    return <EmptyState title="请先登录" description="登录后查看私信" />;
  }

  return (
    <div className="mx-auto grid max-w-4xl gap-4 md:grid-cols-3">
      <div className="rounded-2xl border border-orange-100 bg-white p-4 shadow-sm md:col-span-1">
        <div className="mb-3 flex items-center justify-between">
          <h1 className="font-bold">💬 私信</h1>
          <span className={`h-2 w-2 rounded-full ${status === 'open' ? 'bg-green-500' : 'bg-gray-300'}`} title={`连接状态: ${status}`} />
        </div>
        {!conversations || conversations.length === 0 ? (
          <div className="py-8 text-center text-sm text-gray-400">暂无会话</div>
        ) : (
          <div className="space-y-2">
            {conversations.map((c) => (
              <button
                key={c.peer_id}
                className={`w-full rounded-xl p-3 text-left ${peerId === c.peer_id ? 'bg-orange-100' : 'hover:bg-orange-50'}`}
                onClick={() => setPeerId(c.peer_id)}
              >
                <div className="flex items-center justify-between">
                  <span className="font-medium">{c.peer_name}</span>
                  {c.unread_count > 0 && <span className="rounded-full bg-red-500 px-1.5 text-xs text-white">{c.unread_count}</span>}
                </div>
                <div className="mt-1 truncate text-xs text-gray-500">{c.last_message}</div>
                <div className="mt-0.5 text-xs text-gray-400">{c.last_time}</div>
              </button>
            ))}
          </div>
        )}
      </div>
      <div className="flex min-h-[480px] flex-col rounded-2xl border border-orange-100 bg-white p-4 shadow-sm md:col-span-2">
        {!peerId ? (
          <div className="flex flex-1 items-center justify-center text-sm text-gray-400">
            选择左侧会话开始聊天（ws 状态：{status}）
          </div>
        ) : (
          <>
            <div className="mb-3 border-b border-gray-100 pb-2 text-sm font-medium">与 {peerId} 的会话</div>
            <div className="flex-1 space-y-2 overflow-y-auto">
              {messages.map((m) => (
                <div key={m.id} className={`flex ${m.from_id === user?.id ? 'justify-end' : 'justify-start'}`}>
                  <div className={`max-w-[75%] rounded-2xl px-3 py-2 text-sm ${m.from_id === user?.id ? 'bg-brand-500 text-white' : 'bg-gray-100 text-gray-800'}`}>
                    {m.type === 'image' ? (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img src={m.content} alt="msg" className="max-h-40 rounded-lg" />
                    ) : (
                      m.content
                    )}
                    <div className={`mt-0.5 text-[10px] ${m.from_id === user?.id ? 'text-white/70' : 'text-gray-400'}`}>{m.created_at}</div>
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-3 flex gap-2 border-t border-gray-100 pt-3">
              <input
                className="flex-1 rounded-lg border border-gray-200 px-3 py-2"
                value={text}
                onChange={(e) => setText(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && sendMsg()}
                placeholder="输入消息..."
              />
              <button className="rounded-lg bg-brand-500 px-4 text-white" onClick={sendMsg}>发送</button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
