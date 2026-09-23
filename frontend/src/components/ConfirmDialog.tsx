'use client';

// 通用确认弹窗
export default function ConfirmDialog({
  open,
  title,
  content,
  onCancel,
  onConfirm,
  loading,
}: {
  open: boolean;
  title: string;
  content: string;
  onCancel: () => void;
  onConfirm: () => void;
  loading?: boolean;
}) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="w-80 rounded-2xl bg-white p-6 shadow-xl">
        <div className="text-lg font-semibold">{title}</div>
        <div className="mt-2 text-sm text-gray-600">{content}</div>
        <div className="mt-5 flex justify-end gap-2">
          <button className="rounded-lg border border-gray-300 px-4 py-2 text-sm" onClick={onCancel}>
            取消
          </button>
          <button className="rounded-lg bg-brand-500 px-4 py-2 text-sm text-white disabled:opacity-50" onClick={onConfirm} disabled={loading}>
            {loading ? '处理中...' : '确认'}
          </button>
        </div>
      </div>
    </div>
  );
}
