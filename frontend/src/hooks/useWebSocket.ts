'use client';
import { useEffect, useRef, useState } from 'react';
import { getToken } from '@/utils/request';

export type WSStatus = 'connecting' | 'open' | 'closed';

// 实时私信 websocket hook
export function useWebSocket(onMessage?: (data: unknown) => void) {
  const wsRef = useRef<WebSocket | null>(null);
  const [status, setStatus] = useState<WSStatus>('closed');
  const cbRef = useRef(onMessage);
  cbRef.current = onMessage;

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const ws = new WebSocket(`${proto}://${window.location.host}/api/v1/chat/ws?token=${token}`);
    wsRef.current = ws;
    setStatus('connecting');
    ws.onopen = () => setStatus('open');
    ws.onclose = () => setStatus('closed');
    ws.onerror = () => setStatus('closed');
    ws.onmessage = (ev) => {
      try {
        const payload = JSON.parse(ev.data as string);
        cbRef.current?.(payload);
      } catch {
        // ignore malformed frames
      }
    };
    return () => {
      ws.close();
    };
  }, []);

  const send = (msg: { to_id: string; msg_type: string; content: string }) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    }
  };

  return { status, send };
}
