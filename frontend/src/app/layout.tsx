import type { Metadata } from 'next';
import './globals.css';
import Providers from '@/components/Providers';
import Navbar from '@/components/Navbar';

export const metadata: Metadata = {
  title: '萌宠社区 - 宠物图文短视频社交社区',
  description: '专属铲屎官的图文+短视频社交社区',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN">
      <body className="min-h-screen bg-orange-50/40 text-gray-800">
        <Providers>
          <Navbar />
          <main className="mx-auto max-w-6xl px-4 py-6">{children}</main>
          <footer className="border-t border-orange-100 py-6 text-center text-xs text-gray-400">
            🐾 萌宠社区 · 宠物图文短视频社交社区
          </footer>
        </Providers>
      </body>
    </html>
  );
}
