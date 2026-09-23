import { http } from '@/utils/request';
import type { AuditLog, PageData } from '@/types';

export const auditApi = {
  list: (params?: { action?: string; username?: string; page?: number; page_size?: number }) =>
    http.get<PageData<AuditLog>>('/admin/audit-logs', params),
};
