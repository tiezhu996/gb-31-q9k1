// 日期、数量、状态格式化工具
export function formatDate(s?: string): string {
  if (!s) return '-';
  return s.split(' ')[0];
}

export function formatDateTime(s?: string): string {
  return s || '-';
}

export function formatCount(n?: number): string {
  const v = n || 0;
  if (v >= 10000) return `${(v / 10000).toFixed(1)}w`;
  if (v >= 1000) return `${(v / 1000).toFixed(1)}k`;
  return String(v);
}

export function formatGender(g?: string): string {
  return g === 'male' ? '男孩' : g === 'female' ? '女孩' : '-';
}

export function truncate(s?: string, len = 40): string {
  if (!s) return '';
  return s.length > len ? `${s.slice(0, len)}...` : s;
}
