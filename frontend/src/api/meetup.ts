import { http } from '@/utils/request';
import type { Meetup, MeetupStatus, PageData } from '@/types';

export const meetupApi = {
  list: (params?: { city?: string; status?: string; page?: number; page_size?: number }) =>
    http.get<PageData<Meetup>>('/meetups', params),
  get: (id: string) => http.get<Meetup>(`/meetups/${id}`),
  mine: (params?: { page?: number; page_size?: number }) =>
    http.get<PageData<Meetup>>('/meetups/me', params),
  create: (data: {
    title: string;
    description?: string;
    city: string;
    location: string;
    meet_time: string;
    duration_minutes?: number;
    max_people: number;
  }) => http.post<Meetup>('/meetups', data),
  join: (id: string) => http.post<{ message: string }>(`/meetups/${id}/join`),
  cancelJoin: (id: string) => http.delete<{ message: string }>(`/meetups/${id}/join`),
  updateStatus: (id: string, status: MeetupStatus) =>
    http.patch<{ message: string }>(`/meetups/${id}/status`, { status }),
};
