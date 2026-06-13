// src/api/admin.ts — All admin dashboard API calls.
//
// Chaque fonction correspond à une route /admin/*.
// Aucun fetch direct dans les composants — tout passe par ce fichier.
import { apiClient } from './client'

// ─── Types ────────────────────────────────────────────────────────────────────

export interface AdminUser {
  id: string
  username: string
  email: string
  role: string
  xp: number
  is_banned: boolean
  created_at: string
}

export interface LevelThreshold {
  level: number
  min_xp: number
}

export interface AdminXPAction {
  action: string
  xp_value: number
  daily_limit: number
}

export interface AdminBadge {
  id: string
  slug: string
  label: string
  description: string
  condition: unknown
}

export interface AdminShopItem {
  id: string
  slug: string
  name: string
  category: string
  price_cents: number
  currency: string
  unlock_condition: unknown
  is_active: boolean
}

export interface AdminPing {
  id: string
  type: string
  created_by: string
  is_active: boolean
  report_count: number
  created_at: string
}

// ─── Users ────────────────────────────────────────────────────────────────────

export const listUsers = (page = 1, search = '') =>
  apiClient
    .get<AdminUser[]>(`/admin/users?page=${page}&search=${encodeURIComponent(search)}`)
    .then((r) => r.data)

export const updateUser = (id: string, body: { role?: string; is_banned?: boolean }) =>
  apiClient.patch(`/admin/users/${id}`, body)

export const createUser = (body: { email: string; username: string; password: string; role: string }) =>
  apiClient.post<{ id: string }>('/admin/users', body).then((r) => r.data)

export const deleteUser = (id: string) =>
  apiClient.delete(`/admin/users/${id}`)

// ─── XP Actions ───────────────────────────────────────────────────────────────

export const listXPActions = () =>
  apiClient.get<AdminXPAction[]>('/admin/xp-actions').then((r) => r.data)

export const updateXPAction = (action: string, body: { xp_value?: number; daily_limit?: number }) =>
  apiClient.put(`/admin/xp-actions/${action}`, body)

export const createXPAction = (body: { action: string; xp_value: number; daily_limit: number }) =>
  apiClient.post('/admin/xp-actions', body)

// ─── Level Thresholds ─────────────────────────────────────────────────────────

export const listThresholds = () =>
  apiClient.get<LevelThreshold[]>('/admin/level-thresholds').then((r) => r.data)

export const replaceThresholds = (thresholds: LevelThreshold[]) =>
  apiClient.put('/admin/level-thresholds', { thresholds })

// ─── Badges ───────────────────────────────────────────────────────────────────

export const listBadges = () =>
  apiClient.get<AdminBadge[]>('/admin/badges').then((r) => r.data)

export const createBadge = (body: Omit<AdminBadge, 'id'>) =>
  apiClient.post<{ id: string }>('/admin/badges', body).then((r) => r.data)

export const updateBadge = (id: string, body: Omit<AdminBadge, 'id'>) =>
  apiClient.put(`/admin/badges/${id}`, body)

export const deleteBadge = (id: string) =>
  apiClient.delete(`/admin/badges/${id}`)

// ─── Shop Items ───────────────────────────────────────────────────────────────

export const listShopItems = () =>
  apiClient.get<AdminShopItem[]>('/admin/shop-items').then((r) => r.data)

export const createShopItem = (body: Omit<AdminShopItem, 'id'>) =>
  apiClient.post<{ id: string }>('/admin/shop-items', body).then((r) => r.data)

export const updateShopItem = (id: string, body: Omit<AdminShopItem, 'id'>) =>
  apiClient.put(`/admin/shop-items/${id}`, body)

export const deleteShopItem = (id: string) =>
  apiClient.delete(`/admin/shop-items/${id}`)

// ─── Pings ────────────────────────────────────────────────────────────────────

export const listPingsAdmin = (params: { active?: boolean; flagged?: boolean } = {}) => {
  const q = new URLSearchParams()
  if (params.active !== undefined) q.set('active', String(params.active))
  if (params.flagged !== undefined) q.set('flagged', String(params.flagged))
  return apiClient.get<AdminPing[]>(`/admin/pings?${q}`).then((r) => r.data)
}

export const forceDeactivatePing = (id: string) =>
  apiClient.delete(`/admin/pings/${id}`)

export const createPingAdmin = (body: { user_id: string; type: string; lat: number; lon: number }) =>
  apiClient.post<{ id: string }>('/admin/pings', body).then((r) => r.data)
