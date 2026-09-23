'use client';

export interface Column<T> {
  key: string;
  title: string;
  render?: (row: T) => React.ReactNode;
}

// 通用数据表格：审计页/话题页等复用
export default function DataTable<T extends { id: string }>({
  columns,
  rows,
  empty = '暂无数据',
}: {
  columns: Column<T>[];
  rows: T[];
  empty?: string;
}) {
  if (!rows || rows.length === 0) {
    return <div className="rounded-xl border border-gray-200 bg-white p-8 text-center text-sm text-gray-500">{empty}</div>;
  }
  return (
    <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white">
      <table className="min-w-full divide-y divide-gray-200 text-sm">
        <thead className="bg-gray-50">
          <tr>
            {columns.map((col) => (
              <th key={col.key} className="px-4 py-3 text-left font-medium text-gray-600">
                {col.title}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {rows.map((row) => (
            <tr key={row.id} className="hover:bg-orange-50/40">
              {columns.map((col) => (
                <td key={col.key} className="px-4 py-3">
                  {col.render ? col.render(row) : String((row as Record<string, unknown>)[col.key] ?? '-')}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
