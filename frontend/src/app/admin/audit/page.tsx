'use client';
import { useQuery } from '@tanstack/react-query';
import { auditApi } from '@/api/audit';
import DataTable from '@/components/DataTable';
import EmptyState from '@/components/EmptyState';
import { useAuth } from '@/hooks/useAuth';
import { usePagination } from '@/hooks/usePagination';
import type { AuditLog } from '@/types';

export default function AuditPage() {
  const { isAdmin, isLogin } = useAuth();
  const { page, pageSize } = usePagination(20);
  const { data, isLoading } = useQuery({
    queryKey: ['audit', page, pageSize],
    queryFn: () => auditApi.list({ page, page_size: pageSize }),
    enabled: isAdmin,
  });

  if (!isLogin) return <EmptyState title="请先登录" description="管理员可查看审计日志" />;
  if (!isAdmin) return <EmptyState title="无权限" description="仅管理员可访问审计页面" />;

  return (
    <div className="mx-auto max-w-5xl">
      <h1 className="mb-4 text-xl font-bold">📋 操作审计日志</h1>
      {isLoading ? (
        <div className="py-20 text-center text-gray-400">加载中...</div>
      ) : !data || data.items.length === 0 ? (
        <EmptyState title="暂无审计日志" />
      ) : (
        <DataTable<AuditLog>
          columns={[
            { key: 'created_at', title: '时间', render: (r) => r.created_at },
            { key: 'username', title: '用户' },
            { key: 'action', title: '动作' },
            { key: 'resource', title: '资源' },
            { key: 'resource_id', title: '资源ID', render: (r) => <span className="font-mono text-xs">{r.resource_id}</span> },
            { key: 'detail', title: '详情' },
            { key: 'ip', title: 'IP' },
          ]}
          rows={data.items}
        />
      )}
    </div>
  );
}
