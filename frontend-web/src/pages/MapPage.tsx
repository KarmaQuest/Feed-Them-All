// src/pages/MapPage.tsx — Page principale : la carte interactive.
//
// Affiche la carte plein écran avec :
//   - MapView (Leaflet + marqueurs + WS)
//   - Barre de contrôle flottante (logo, bouton Signaler)
//   - Toast géolocalisation refusée
//   - Modal SignalForm
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import MapView from '../components/map/MapView'
import SignalForm from '../components/map/SignalForm'
import { useGeolocation } from '../hooks/useGeolocation'
import { useMapStore } from '../store/map'
import { useAuthStore } from '../store/auth'
import { useWebSocketStore } from '../store/websocket'
import { logout } from '../api/auth'
import type { Ping } from '../api/pings'
import './MapPage.css'

export default function MapPage() {
  const { loading: geoLoading, error: geoError } = useGeolocation()
  const { pings, addPing } = useMapStore()
  const { user, logout: logoutStore } = useAuthStore()
  const connected = useWebSocketStore((s) => s.connected)
  const navigate = useNavigate()

  const [showSignalForm, setShowSignalForm] = useState(false)
  const [pickMode, setPickMode] = useState(false)
  const [pickedLat, setPickedLat] = useState<number | null>(null)
  const [pickedLon, setPickedLon] = useState<number | null>(null)

  function handleRequestMapPick() {
    setPickMode(true)
  }

  function handleMapPick(lat: number, lon: number) {
    setPickedLat(lat)
    setPickedLon(lon)
    setPickMode(false) // un seul clic suffit, on désactive le mode pick
  }

  async function handleLogout() {
    try { await logout() } catch { /* ignore */ }
    logoutStore()
    navigate('/')
  }

  function onPingCreated(ping: Ping) {
    addPing(ping)
    setShowSignalForm(false)
    setPickedLat(null)
    setPickedLon(null)
  }

  const animalCount = pings.filter((p) => p.type === 'animal').length
  const foodCount = pings.filter((p) => p.type === 'food').length

  return (
    <div className="map-page">
      {/* Carte plein écran */}
      <div className={`map-container${pickMode ? ' map-container--pick' : ''}`}>
        <MapView onMapPick={pickMode ? handleMapPick : null} />
      </div>

      {/* Barre de contrôle flottante en haut */}
      <div className="map-topbar">
        <div className="map-topbar__brand">
          <img src="/logo.png" alt="FeedThemAll" className="map-topbar__logo" />
          <span className="map-topbar__name">FeedThemAll</span>
        </div>

        <div className="map-topbar__stats">
          <span className="map-stat map-stat--animal">🐾 {animalCount}</span>
          <span className="map-stat map-stat--food">🍖 {foodCount}</span>
          {user && (
            <span className={`map-stat map-stat--ws ${connected ? 'map-stat--online' : 'map-stat--offline'}`}>
              {connected ? '● Live' : '○ Hors ligne'}
            </span>
          )}
        </div>

        <div className="map-topbar__actions">
          {user ? (
            <>
              <button
                className="map-btn map-btn--signal"
                onClick={() => setShowSignalForm(true)}
              >
                + Signaler
              </button>
              <div className="map-topbar__user">
                <span className="map-topbar__username">{user.username}</span>
                {user.role === 'admin' && (
                  <a href="/admin" className="map-btn map-btn--admin">
                    ⚙ Admin
                  </a>
                )}
                <button className="map-btn map-btn--logout" onClick={handleLogout}>
                  Déconnexion
                </button>
              </div>
            </>
          ) : (
            <>
              <a href="/user-login" className="map-btn map-btn--login">
                Connexion
              </a>
              <a href="/register" className="map-btn map-btn--signal">
                S'inscrire
              </a>
            </>
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

      {/* Modal signalement */}
      {showSignalForm && (
        <SignalForm
          onDone={onPingCreated}
          onCancel={() => { setShowSignalForm(false); setPickMode(false) }}
          onRequestMapPick={handleRequestMapPick}
          pickedLat={pickedLat}
          pickedLon={pickedLon}
        />
      )}

      {/* Indicateur mode pick (clic sur carte) */}
      {pickMode && (
        <div className="map-pick-hint">
          🗺 Cliquez sur la carte pour placer le marqueur
          <button
            className="map-pick-hint__cancel"
            onClick={() => setPickMode(false)}
          >
            ✕
          </button>
        </div>
      )}
    </div>
  )
}
