// src/store/map.ts — Zustand store pour les pings affichés sur la carte.
//
// Contient :
//   - La liste des pings chargés depuis l'API
//   - Le ping sélectionné (popup ouverte)
//   - La position de l'utilisateur (géolocalisation browser)
//   - Les positions des autres feeders (reçues via WebSocket)
import { create } from 'zustand'
import type { Ping } from '../api/pings'

export interface FeederPosition {
  feeder_id: string
  lat: number
  lon: number
}

interface MapStore {
  // Pings
  pings: Ping[]
  setPings: (pings: Ping[]) => void
  addPing: (ping: Ping) => void
  updatePing: (ping: Ping) => void
  removePing: (id: string) => void

  // Ping sélectionné (popup)
  selectedPingId: string | null
  setSelectedPing: (id: string | null) => void

  // Position utilisateur
  userLat: number | null
  userLon: number | null
  setUserPosition: (lat: number, lon: number) => void

  // Feeders actifs (via WebSocket)
  feeders: Record<string, FeederPosition>
  updateFeeder: (pos: FeederPosition) => void
}

export const useMapStore = create<MapStore>((set) => ({
  pings: [],
  setPings: (pings) => set({ pings }),
  addPing: (ping) =>
    set((s) => ({
      pings: s.pings.some((p) => p.id === ping.id)
        ? s.pings
        : [ping, ...s.pings],
    })),
  updatePing: (ping) =>
    set((s) => ({
      pings: s.pings.map((p) => (p.id === ping.id ? ping : p)),
    })),
  removePing: (id) =>
    set((s) => ({ pings: s.pings.filter((p) => p.id !== id) })),

  selectedPingId: null,
  setSelectedPing: (id) => set({ selectedPingId: id }),

  userLat: null,
  userLon: null,
  setUserPosition: (lat, lon) => set({ userLat: lat, userLon: lon }),

  feeders: {},
  updateFeeder: (pos) =>
    set((s) => ({
      feeders: { ...s.feeders, [pos.feeder_id]: pos },
    })),
}))
