'use client';
import { useState } from 'react';

// 分页 hook：所有列表页复用
export function usePagination(initialSize = 10) {
  const [page, setPage] = useState(1);
  const [pageSize] = useState(initialSize);
  return { page, pageSize, setPage };
}
