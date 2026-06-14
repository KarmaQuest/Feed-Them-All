// src/components/map/MapView.tsx — Composant principal de la carte Leaflet.
//
// Responsabilités :
//   - Rendu de la carte OSM centrée sur la position utilisateur (ou Paris par défaut)
//   - Marqueurs pour chaque ping (animal/food/fed)
//   - Marqueur de position utilisateur
//   - Marqueurs feeders actifs (via WS)
//   - Abonnement WebSocket sur la bounding box visible
//   - Emit de position GPS au WS (si connecté)
import { useEffect, useCallback } from 'react'
import {
  MapContainer,
  TileLayer,
  Marker,
  useMapEvents,
  useMap,
} from 'react-leaflet'
import type { LeafletMouseEvent } from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { useMapStore } from '../../store/map'
import { useWebSocketStore } from '../../store/websocket'
import { useAuthStore } from '../../store/auth'
import { getPingsNearby } from '../../api/pings'
import { DEFAULT_LAT, DEFAULT_LON, DEFAULT_ZOOM } from '../../hooks/useGeolocation'
import { animalIcon, foodIcon, fedIcon, userIcon, feederIcon } from './markers'
import PingPopup from './PingPopup'

// ── Sous-composant : abonnement WS + chargement des pings à chaque déplacement ──
function MapEventHandler() {
  const { setPings, setSelectedPing, userLat, userLon } = useMapStore()
  const { subscribe, sendPosition, connected } = useWebSocketStore()

  const loadPings = useCallback(
    async (lat: number, lon: number) => {
      try {
        const data = await getPingsNearby(lat, lon, 3000)
        setPings(data.filter((p) => p.is_active))
      } catch {
        // silencieux — la carte reste utilisable sans pings
      }
    },
    [setPings],
  )

  // Charge les pings et s'abonne au WS quand la carte est déplacée
  useMapEvents({
    moveend(e) {
      const center = e.target.getCenter()
      const bounds = e.target.getBounds()
      loadPings(center.lat, center.lng)
      if (connected) {
        subscribe({
          min_lat: bounds.getSouth(),
          max_lat: bounds.getNorth(),
          min_lon: bounds.getWest(),
          max_lon: bounds.getEast(),
        })
      }
      setSelectedPing(null)
    },
  })

  // Envoie la position GPS au WS quand elle change
  useEffect(() => {
    if (userLat !== null && userLon !== null && connected) {
      sendPosition(userLat, userLon)
    }
  }, [userLat, userLon, connected, sendPosition])

  return null
}

// ── Sous-composant : centrage automatique sur la position utilisateur ──────────
function RecenterOnUser() {
  const map = useMap()
  const userLat = useMapStore((s) => s.userLat)
  const userLon = useMapStore((s) => s.userLon)

  useEffect(() => {
    if (userLat !== null && userLon !== null) {
      // Ne recentre qu'une seule fois (premier fix GPS)
      map.setView([userLat, userLon], map.getZoom(), { animate: true })
    }
    // On ne met pas map dans les deps — on veut déclencher uniquement au premier fix
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userLat, userLon])

  return null
}

// ── Sous-composant : intercepte un clic unique sur la carte pour choisir une position ──
function MapClickHandler({ onPick }: { onPick: (lat: number, lon: number) => void }) {
  useMapEvents({
    click(e: LeafletMouseEvent) {
      onPick(e.latlng.lat, e.latlng.lng)
    },
  })
  return null
}

// ── Composant principal ────────────────────────────────────────────────────────
interface MapViewProps {
  /** Si défini, active le mode "pick" : le prochain clic sur la carte appelle ce callback */
  onMapPick?: ((lat: number, lon: number) => void) | null
}

export default function MapView({ onMapPick }: MapViewProps = {}) {
  const { pings, userLat, userLon, feeders, setSelectedPing } = useMapStore()
  const { connect, disconnect } = useWebSocketStore()
  const { user } = useAuthStore()

  // Connexion WS uniquement si l'utilisateur est connecté
  useEffect(() => {
    if (user) {
      connect()
      return () => disconnect()
    }
  }, [user, connect, disconnect])

  const center: [number, number] =
    userLat !== null && userLon !== null
      ? [userLat, userLon]
      : [DEFAULT_LAT, DEFAULT_LON]

  return (
    <MapContainer
      center={center}
      zoom={DEFAULT_ZOOM}
      style={{ width: '100%', height: '100%' }}
      zoomControl={true}
    >
      <TileLayer
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        maxZoom={19}
      />

      <MapEventHandler />
      <RecenterOnUser />
      {onMapPick && <MapClickHandler onPick={onMapPick} />}

      {/* Marqueurs pings */}
      {pings.map((ping) => {
        const icon =
          ping.fed_at ? fedIcon : ping.type === 'animal' ? animalIcon : foodIcon
        return (
          <Marker
            key={ping.id}
            position={[ping.lat, ping.lon]}
            icon={icon}
            eventHandlers={{ click: () => setSelectedPing(ping.id) }}
          >
            <PingPopup ping={ping} />
          </Marker>
        )
      })}

      {/* Marqueur position utilisateur */}
      {userLat !== null && userLon !== null && (
        <Marker position={[userLat, userLon]} icon={userIcon} />
      )}

      {/* Marqueurs feeders actifs */}
      {Object.values(feeders).map((f) => (
        <Marker
          key={f.feeder_id}
          position={[f.lat, f.lon]}
          icon={feederIcon}
        />
      ))}
    </MapContainer>
  )
}
