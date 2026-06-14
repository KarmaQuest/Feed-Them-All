// src/pages/MapPage.tsx — Page principale : la carte interactive.
//
// Affiche la carte plein écran avec :
//   - MapView (Leaflet + marqueurs + WS)
//   - Topbar flottante minimaliste (logo bouton sidebar + stats)
//   - MapSidebar slideout droite (nav / signal / ping detail)
//   - Bandeau pick-mode (clic pour placer un marqueur)
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import MapView from '../components/map/MapView'
import MapSidebar from '../components/map/MapSidebar'
import { useGeolocation } from '../hooks/useGeolocation'
import { useMapStore } from '../store/map'
import { useAuthStore } from '../store/auth'
import { useWebSocketStore } from '../store/websocket'
import { logout } from '../api/auth'
import type { Ping } from '../api/pings'
import './MapPage.css'

export default function MapPage() {
  const { loading: geoLoading, error: geoError } = useGeolocation()
  const { pings, addPing, setSelectedPing } = useMapStore()
  const { user, logout: logoutStore } = useAuthStore()
  const connected = useWebSocketStore((s) => s.connected)
  const navigate = useNavigate()

  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [pickMode, setPickMode] = useState(false)
  const [pickedLat, setPickedLat] = useState<number | null>(null)
  const [pickedLon, setPickedLon] = useState<number | null>(null)

  function handleRequestMapPick() {
    setPickMode(true)
    setSidebarOpen(false)
  }

  function handleMapPick(lat: number, lon: number) {
    setPickedLat(lat)
    setPickedLon(lon)
    setPickMode(false)
    setSidebarOpen(true)
  }

  function handleMarkerClick(pingId: string) {
    setSelectedPing(pingId)
    setSidebarOpen(true)
  }

  async function handleLogout() {
    try { await logout() } catch { /* ignore */ }
    logoutStore()
    navigate('/')
  }

  function onPingCreated(ping: Ping) {
    addPing(ping)
    setPickedLat(null)
    setPickedLon(null)
  }

  const animalCount = pings.filter((p) => p.type === 'animal').length
  const foodCount = pings.filter((p) => p.type === 'food').length

  return (
    <div className="map-page">
      {/* Carte plein écran */}
      <div className={`map-container${pickMode ? ' map-container--pick' : ''}`}>
        <MapView
          onMapPick={pickMode ? handleMapPick : null}
          onMarkerClick={handleMarkerClick}
        />
      </div>

      {/* Topbar minimaliste : logo toggle + stats */}
      <div className="map-topbar">
        <button
          className="map-topbar__toggle"
          onClick={() => setSidebarOpen((v) => !v)}
          aria-label="Ouvrir le menu"
        >
          <img src="/logo.png" alt="FeedThemAll" className="map-topbar__logo" />
        </button>

        <div className="map-topbar__stats">
          <span className="map-stat map-stat--animal">🐾 {animalCount}</span>
          <span className="map-stat map-stat--food">🍖 {foodCount}</span>
          {user && (
            <span className={`map-stat map-stat--ws ${connected ? 'map-stat--online' : 'map-stat--offline'}`}>
              {connected ? '● Live' : '○ Hors ligne'}
            </span>
          )}
        </div>
      </div>

      {/* Loader géolocalisation */}
      {geoLoading && (
        <div className="map-geo-loader">
          <span className="map-geo-loader__spinner" />
          <span>Recherche de votre position…</span>
        </div>
      )}

      {/* Toast géolocalisation */}
      {geoError && (
        <div className="map-toast map-toast--warning">
          📍 {geoError}
        </div>
      )}

      {/* Bandeau mode pick */}
      {pickMode && (
        <div className="map-pick-hint">
          🗺 Cliquez sur la carte pour placer le marqueur
          <button
            className="map-pick-hint__cancel"
            onClick={() => { setPickMode(false); setSidebarOpen(true) }}
          >
            ✕
          </button>
        </div>
      )}

      {/* Sidebar slideout droite */}
      <MapSidebar
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        pickMode={pickMode}
        pickedLat={pickedLat}
        pickedLon={pickedLon}
        onRequestMapPick={handleRequestMapPick}
        onPingCreated={onPingCreated}
        onLogout={handleLogout}
      />
    </div>
  )
}
