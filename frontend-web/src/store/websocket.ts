// src/store/websocket.ts — Zustand store pour la connexion WebSocket.
//
// Gère le cycle de vie d'une seule connexion WS :
//   connect()    → ouvre la connexion + écoute les messages entrants
//   disconnect() → ferme proprement la connexion
//   subscribe()  → envoie la bounding box au serveur (abonnement à une zone)
//   sendPosition() → envoie la position GPS de l'utilisateur
//
// Reconnexion automatique : si la connexion se coupe, tentative après 3s (max 10×).
//
// Les messages entrants ("ping_created", "ping_updated", "feeder_position")
// sont dispatchés directement dans useMapStore.
import { create } from 'zustand'
import { useMapStore } from './map'
import type { Ping } from '../api/pings'

const WS_URL = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/api/ws`
const MAX_RECONNECT = 10
const RECONNECT_DELAY = 3000

export interface BoundingBox {
  min_lat: number
  max_lat: number
  min_lon: number
  max_lon: number
}

interface WSStore {
  socket: WebSocket | null
  connected: boolean
  connect: () => void
  disconnect: () => void
  subscribe: (bbox: BoundingBox) => void
  sendPosition: (lat: number, lon: number) => void
}

let reconnectCount = 0
let reconnectTimer: ReturnType<typeof setTimeout> | null = null

function scheduleReconnect(get: () => WSStore, set: (p: Partial<WSStore>) => void) {
  if (reconnectTimer) return
  if (reconnectCount >= MAX_RECONNECT) {
    reconnectCount = 0
    return
  }
  reconnectCount++
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    get().connect()
  }, RECONNECT_DELAY * reconnectCount) // exponential: 3s, 6s, 9s...
}

export const useWebSocketStore = create<WSStore>((set, get) => ({
  socket: null,
  connected: false,

  connect() {
    if (get().socket) return

    const ws = new WebSocket(WS_URL)

    ws.onopen = () => {
      reconnectCount = 0
      set({ connected: true })
    }

    ws.onclose = () => {
      set({ socket: null, connected: false })
      scheduleReconnect(get, set)
    }

    ws.onerror = () => {
      ws.close()
    }

    ws.onmessage = evt => {
      let msg: { type: string; ping?: Ping; feeder_id?: string; lat?: number; lon?: number }
      try {
        msg = JSON.parse(evt.data as string)
      } catch {
        return
      }

      const mapStore = useMapStore.getState()

      if (msg.type === 'ping_created' && msg.ping) {
        mapStore.addPing(msg.ping)
      } else if (msg.type === 'ping_updated' && msg.ping) {
        if (!msg.ping.is_active) {
          mapStore.removePing(msg.ping.id)
        } else {
          mapStore.updatePing(msg.ping)
        }
      } else if (
        msg.type === 'feeder_position' &&
        msg.feeder_id &&
        msg.lat !== undefined &&
        msg.lon !== undefined
      ) {
        mapStore.updateFeeder({
          feeder_id: msg.feeder_id,
          lat: msg.lat!,
          lon: msg.lon!,
        })
      }
    }

    set({ socket: ws })
  },

  disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    reconnectCount = 0
    get().socket?.close()
    set({ socket: null, connected: false })
  },

  subscribe(bbox: BoundingBox) {
    const { socket } = get()
    if (!socket || socket.readyState !== WebSocket.OPEN) return
    socket.send(JSON.stringify({ type: 'subscribe', bounding_box: bbox }))
  },

  sendPosition(lat: number, lon: number) {
    const { socket } = get()
    if (!socket || socket.readyState !== WebSocket.OPEN) return
    socket.send(JSON.stringify({ type: 'position', lat, lon }))
  },
}))
