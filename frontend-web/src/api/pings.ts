// src/api/pings.ts — Toutes les requêtes liées aux pings.
import { apiClient } from './client'

// ─── Types ────────────────────────────────────────────────────────────────────

export interface Ping {
  id: string
  type: 'animal' | 'food'
  lat: number
  lon: number
  created_by: string
  is_active: boolean
  fed_at: string | null
  animal_type: 'cat' | 'dog' | 'other' | null
  animal_count: number
  created_at: string
  updated_at: string
}

export interface CreatePingRequest {
  type: 'animal' | 'food'
  lat: number
  lon: number
  animal_type?: 'cat' | 'dog' | 'other'
  animal_count?: number
}

export interface PingMedia {
  id: string
  ping_id: string
  url: string
  uploaded_by: string
  created_at: string
}

export interface FeedingEvent {
  id: string
  ping_id: string
  fed_by: string
  username: string
  note: string | null
  animal_count_seen: number
  created_at: string
}

// ─── API functions ────────────────────────────────────────────────────────────

export async function getPingsNearby(
  lat: number,
  lon: number,
  radius = 2000,
): Promise<Ping[]> {
  const res = await apiClient.get<Ping[]>(
    `/pings?lat=${lat}&lon=${lon}&radius=${radius}`,
  )
  return res.data
}

export async function createPing(data: CreatePingRequest): Promise<Ping> {
  const res = await apiClient.post<Ping>('/pings', data)
  return res.data
}

export async function markFed(
  pingId: string,
  note?: string,
  animalCountSeen?: number,
): Promise<void> {
  await apiClient.post(`/pings/${pingId}/feedings`, {
    note: note ?? null,
    animal_count_seen: animalCountSeen ?? 1,
  })
}

export async function confirmPing(pingId: string): Promise<void> {
  await apiClient.patch(`/pings/${pingId}/confirm`)
}

export async function getPingMedia(pingId: string): Promise<PingMedia[]> {
  const res = await apiClient.get<PingMedia[]>(`/pings/${pingId}/media`)
  return res.data
}

export async function uploadPingMedia(
  pingId: string,
  file: File,
): Promise<PingMedia> {
  const form = new FormData()
  form.append('photo', file)
  const res = await apiClient.post<PingMedia>(`/pings/${pingId}/media`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return res.data
}

export async function getPingFeedings(pingId: string): Promise<FeedingEvent[]> {
  const res = await apiClient.get<FeedingEvent[]>(`/pings/${pingId}/feedings`)
  return res.data
}

export async function deactivatePing(pingId: string): Promise<void> {
  await apiClient.delete(`/pings/${pingId}`)
}
