'use client';
import { create } from 'zustand';
import type { Meetup } from '@/types';

interface MeetupState {
  meetups: Meetup[];
  setMeetups: (meetups: Meetup[]) => void;
  updateMeetup: (id: string, patch: Partial<Meetup>) => void;
}

export const useMeetupStore = create<MeetupState>((set) => ({
  meetups: [],
  setMeetups: (meetups) => set({ meetups }),
  updateMeetup: (id, patch) =>
    set((s) => ({ meetups: s.meetups.map((m) => (m.id === id ? { ...m, ...patch } : m)) })),
}));
