'use client';

export default function EmptyState({ title = '暂无数据', description, action }: { title?: string; description?: string; action?: React.ReactNode }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-gray-300 bg-white py-16 text-center">
      <div className="text-5xl">🐾</div>
      <div className="mt-3 text-lg font-semibold text-gray-700">{title}</div>
      {description && <div className="mt-1 text-sm text-gray-500">{description}</div>}
      {action && <div className="mt-4">{action}</div>}
    </div>
  );
}
