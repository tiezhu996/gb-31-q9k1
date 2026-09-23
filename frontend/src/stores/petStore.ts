'use client';
import { create } from 'zustand';
import type { Pet } from '@/types';

interface PetState {
  pets: Pet[];
  setPets: (pets: Pet[]) => void;
  addPet: (pet: Pet) => void;
}

export const usePetStore = create<PetState>((set) => ({
  pets: [],
  setPets: (pets) => set({ pets }),
  addPet: (pet) => set((s) => ({ pets: [pet, ...s.pets] })),
}));
