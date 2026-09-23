'use client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { useAuthStore } from '@/stores/authStore';

export default function Providers({ children }: { children: React.ReactNode }) {
  const [client] = useState(() => new QueryClient({ defaultOptions: { queries: { retry: 1, refetchOnWindowFocus: false } } }));
  useEffect(() => {
    useAuthStore.getState().hydrate();
  }, []);
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}
