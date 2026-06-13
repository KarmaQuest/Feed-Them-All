// src/store/auth.ts — Zustand auth store.
//
// Stocke l'état de connexion en mémoire (jamais localStorage).
// Le token JWT est géré via api/client.ts (setAccessToken/getAccessToken).
import { create } from 'zustand'

export interface AuthUser {
  id: string
  username: string
  role: string
}

interface AuthStore {
  user: AuthUser | null
  isLogged: boolean
  login: (user: AuthUser) => void
  logout: () => void
}

export const useAuthStore = create<AuthStore>((set) => ({
  user: null,
  isLogged: false,
  login: (user) => set({ user, isLogged: true }),
  logout: () => set({ user: null, isLogged: false }),
}))
