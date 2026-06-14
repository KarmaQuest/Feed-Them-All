// src/store/auth.ts — Zustand auth store.
//
// Persistance de session : le refresh token est un cookie HttpOnly stocké en DB.
// Au démarrage, initialize() appelle POST /auth/refresh → le backend valide le
// cookie, retourne un nouvel access_token + l'objet user complet depuis la DB.
// Rien n'est écrit en localStorage — toute la persistance est côté serveur.
import { create } from 'zustand'
import axios from 'axios'
import { setAccessToken } from '../api/client'

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
  initialize: () => Promise<void>
}

export const useAuthStore = create<AuthStore>((set) => ({
  user: null,
  isLogged: false,

  login: (user) => set({ user, isLogged: true }),

  logout: () => {
    setAccessToken(null)
    set({ user: null, isLogged: false })
  },

  // Appelé au démarrage de l'app : restaure la session via le refresh token
  // cookie HttpOnly (géré par le navigateur, stocké en DB côté serveur).
  // Le backend retourne access_token + user → aucun localStorage nécessaire.
  initialize: async () => {
    try {
      const res = await axios.post(
        '/api/auth/refresh',
        {},
        { withCredentials: true },
      )
      setAccessToken(res.data.access_token)
      const u = res.data.user
      set({
        user: { id: u.id, username: u.username, role: u.role },
        isLogged: true,
      })
    } catch {
      // Pas de refresh token valide — session expirée ou première visite
      setAccessToken(null)
      set({ user: null, isLogged: false })
    }
  },
}))
