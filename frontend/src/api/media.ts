import axios from 'axios';
import { getToken } from '@/utils/request';

export interface UploadResult {
  url: string;
  type: 'image' | 'video';
  message: string;
}

export async function uploadMedia(file: File): Promise<UploadResult> {
  const form = new FormData();
  form.append('file', file);
  const res = await axios.post<{ code: number; message: string; data: UploadResult }>(
    '/api/v1/media/upload',
    form,
    {
      headers: { Authorization: `Bearer ${getToken()}`, 'Content-Type': 'multipart/form-data' },
      timeout: 60000,
    }
  );
  if (res.data.code !== 0) {
    throw new Error(res.data.message || '上传失败');
  }
  return res.data.data;
}
