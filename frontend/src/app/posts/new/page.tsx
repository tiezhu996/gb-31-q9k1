'use client';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { postApi } from '@/api/post';
import { uploadMedia } from '@/api/media';
import { useAuth } from '@/hooks/useAuth';
import { POST_TYPE } from '@/constants/enums';
import type { Post, PostType, MediaItem } from '@/types';

export default function NewPostPage() {
  const router = useRouter();
  const { isLogin, hydrated } = useAuth();
  const [form, setForm] = useState<{
    content: string;
    type: PostType;
    media: MediaItem[];
    topics: string[];
    location: string;
    city: string;
    pet_ids: string[];
  }>({
    content: '',
    type: 'image',
    media: [],
    topics: [] as string[],
    location: '',
    city: '上海',
    pet_ids: [] as string[],
  });
  const [topicInput, setTopicInput] = useState('');
  const [petIdInput, setPetIdInput] = useState('');
  const [error, setError] = useState('');
  const [uploading, setUploading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (hydrated && !isLogin) {
      router.push('/login');
    }
  }, [hydrated, isLogin, router]);

  if (!hydrated) {
    return <div className="py-20 text-center text-gray-400">加载中...</div>;
  }

  const handleFile = async (file: File) => {
    setUploading(true);
    setError('');
    try {
      const res = await uploadMedia(file);
      const mediaType = res.type === 'video' ? POST_TYPE.VIDEO : POST_TYPE.IMAGE;
      setForm((f) => ({ ...f, type: mediaType, media: [...f.media, { type: mediaType, url: res.url }] }));
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setUploading(false);
    }
  };

  const addTopic = () => {
    const t = topicInput.trim().replace(/^#/, '');
    if (t && !form.topics.includes(t)) {
      setForm((f) => ({ ...f, topics: [...f.topics, t] }));
    }
    setTopicInput('');
  };

  const addPetId = () => {
    const id = petIdInput.trim();
    if (id && !form.pet_ids.includes(id)) {
      setForm((f) => ({ ...f, pet_ids: [...f.pet_ids, id] }));
    }
    setPetIdInput('');
  };

  const submit = async () => {
    setError('');
    if (!form.content.trim()) {
      setError('请填写动态内容');
      return;
    }
    setSubmitting(true);
    try {
      await postApi.create(form as Partial<Post>);
      router.push('/');
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="mx-auto max-w-2xl">
      <h1 className="mb-4 text-xl font-bold">📝 发布动态</h1>
      <div className="space-y-4 rounded-2xl border border-orange-100 bg-white p-6 shadow-sm">
        <div>
          <label className="text-sm text-gray-600">内容 *（支持话题 #话题名）</label>
          <textarea
            className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2"
            rows={5}
            value={form.content}
            onChange={(e) => setForm({ ...form, content: e.target.value })}
            placeholder="晒晒你的毛孩子吧～"
          />
        </div>
        <div>
          <label className="text-sm text-gray-600">上传图片/视频（≤50MB）</label>
          <input type="file" accept="image/*,video/mp4" className="mt-1 w-full text-sm" onChange={(e) => e.target.files?.[0] && handleFile(e.target.files[0])} />
          {uploading && <div className="mt-1 text-xs text-gray-400">上传中...</div>}
          {form.media.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-2">
              {form.media.map((m, i) => (
                <div key={i} className="relative">
                  {m.type === 'video' ? (
                    <video src={m.url} className="h-20 w-20 rounded-lg object-cover" />
                  ) : (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={m.url} alt="preview" className="h-20 w-20 rounded-lg object-cover" />
                  )}
                  <button className="absolute -right-1 -top-1 rounded-full bg-red-500 px-1 text-xs text-white" onClick={() => setForm((f) => ({ ...f, media: f.media.filter((_, j) => j !== i) }))}>✕</button>
                </div>
              ))}
            </div>
          )}
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="text-sm text-gray-600">地点</label>
            <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.location} onChange={(e) => setForm({ ...form, location: e.target.value })} placeholder="世纪公园" />
          </div>
          <div>
            <label className="text-sm text-gray-600">城市</label>
            <input className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2" value={form.city} onChange={(e) => setForm({ ...form, city: e.target.value })} />
          </div>
        </div>
        <div>
          <label className="text-sm text-gray-600">话题</label>
          <div className="flex gap-2">
            <input className="mt-1 flex-1 rounded-lg border border-gray-200 px-3 py-2" value={topicInput} onChange={(e) => setTopicInput(e.target.value)} placeholder="柴犬日常" />
            <button className="mt-1 rounded-lg bg-orange-100 px-3 text-sm text-brand-700" onClick={addTopic}>添加</button>
          </div>
          {form.topics.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-2">
              {form.topics.map((t) => (
                <span key={t} className="rounded-full bg-orange-100 px-2 py-0.5 text-xs text-brand-700">
                  #{t}
                  <button className="ml-1" onClick={() => setForm((f) => ({ ...f, topics: f.topics.filter((x) => x !== t) }))}>✕</button>
                </span>
              ))}
            </div>
          )}
        </div>
        <div>
          <label className="text-sm text-gray-600">宠物标签（宠物 ID）</label>
          <div className="flex gap-2">
            <input className="mt-1 flex-1 rounded-lg border border-gray-200 px-3 py-2" value={petIdInput} onChange={(e) => setPetIdInput(e.target.value)} placeholder="宠物 ID（可选）" />
            <button className="mt-1 rounded-lg bg-orange-100 px-3 text-sm text-brand-700" onClick={addPetId}>添加</button>
          </div>
        </div>
        {error && <div className="text-sm text-red-500">{error}</div>}
        <button className="w-full rounded-lg bg-brand-500 py-2 text-white disabled:opacity-50" onClick={submit} disabled={submitting}>
          {submitting ? '发布中...' : '发布动态'}
        </button>
      </div>
    </div>
  );
}
