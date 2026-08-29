// src/api/client.ts — Axios instance with JWT auth and auto-refresh.
//
// Toutes les requêtes API passent par ce client.
// Le token JWT est stocké en mémoire (non localStorage pour la sécurité XSS).
// Le refresh token est un cookie HttpOnly géré automatiquement par le navigateur.
import axios from 'axios'

const BASE_URL = '/api'

export const apiClient = axios.create({
  baseURL: BASE_URL,
  withCredentials: true, // sends HttpOnly refresh cookie automatically
})

// In-memory token store (never localStorage — XSS safe)
let accessToken: string | null = null

export function setAccessToken(token: string | null) {
  accessToken = token
}

export function getAccessToken(): string | null {
  return accessToken
}

// Attach access token to every request
apiClient.interceptors.request.use(config => {
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }
  return config
})

// Auto-refresh on 401
apiClient.interceptors.response.use(
  res => res,
  async error => {
    const original = error.config
    if (error.response?.status === 401 && !original._retry) {
      original._retry = true
      try {
        const res = await axios.post(`${BASE_URL}/auth/refresh`, {}, { withCredentials: true })
        const newToken = res.data.access_token
        setAccessToken(newToken)
        original.headers.Authorization = `Bearer ${newToken}`
        return apiClient(original)
      } catch {
        setAccessToken(null)
        // Soft logout — dispatche un event, App.tsx navigue via React Router
        // sans rechargement complet (évite la boucle deco/reco Firefox)
        window.dispatchEvent(new CustomEvent('auth:session-expired'))
      }
    }
    return Promise.reject(error)
  }
)
