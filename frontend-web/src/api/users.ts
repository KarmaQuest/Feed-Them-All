// src/api/users.ts — Requêtes API liées aux profils utilisateurs.
import { apiClient } from './client'

export interface BadgeSummary {
  slug: string
  label: string
}

export interface UserProfile {
  id: string
  username: string
  role: string
  roles: string[]
  xp: number
  level: number
  badges: BadgeSummary[]
  avatar_config: Record<string, unknown>
  nb_pings: number
  nb_feedings: number
  is_private: boolean
}

/** Profil réduit retourné quand le profil est privé et qu'on n'est pas le propriétaire */
export interface PrivateProfile {
  id: string
  username: string
  level: number
  is_private: true
}

export async function getUserProfile(userId: string): Promise<UserProfile | PrivateProfile> {
  const res = await apiClient.get<UserProfile | PrivateProfile>(`/users/${userId}/profile`)
  return res.data
}

export async function updatePrivacy(isPrivate: boolean): Promise<void> {
  await apiClient.patch('/users/me/privacy', { is_private: isPrivate })
}

export async function updateAvatar(config: Record<string, unknown>): Promise<void> {
  await apiClient.patch('/users/me/avatar', { config })
}
