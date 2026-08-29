// src/pages/MapPage.tsx — Page principale : la carte interactive.
//
// Affiche la carte plein écran avec :
//   - MapView (Leaflet + marqueurs + WS)
//   - Bouton FAB flottant (logo, haut droite) pour ouvrir la sidebar
//   - MapSidebar slideout droite (nav / ping detail)
//   - SignalForm en modal centré (33vw)
//   - Bandeau pick-mode (clic pour placer un marqueur)
import { useState, useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import MapView from '../components/map/MapView'
import MapSidebar from '../components/map/MapSidebar'
import SignalForm from '../components/map/SignalForm'
import { useGeolocation } from '../hooks/useGeolocation'
import { useMapStore } from '../store/map'
import { useAuthStore } from '../store/auth'
import { logout } from '../api/auth'
import type { Ping } from '../api/pings'
import './MapPage.css'

export default function MapPage() {
  const { loading: geoLoading, error: geoError } = useGeolocation()
  const { addPing, setSelectedPing } = useMapStore()
  const { logout: logoutStore } = useAuthStore()
  const navigate = useNavigate()

  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [signalModalOpen, setSignalModalOpen] = useState(false)
  const location = useLocation()

  useEffect(() => {
    if ((location.state as { fromMap?: boolean })?.fromMap) {
      setSidebarOpen(true)
    }
  }, [location.state])
  const [pickMode, setPickMode] = useState(false)
  const [pickedLat, setPickedLat] = useState<number | null>(null)
  const [pickedLon, setPickedLon] = useState<number | null>(null)
  // Track if SignalForm has ever been opened (keep mounted for state preservation)
  const [signalEverOpened, setSignalEverOpened] = useState(false)

  function handleOpenSignalModal() {
    setSidebarOpen(false)
    setSignalEverOpened(true)
    setTimeout(() => setSignalModalOpen(true), 300)
  }

  function handleRequestMapPick() {
    setPickMode(true)
    setSignalModalOpen(false) // hide modal, don't unmount
  }

  function handleMapPick(lat: number, lon: number) {
    setPickedLat(lat)
    setPickedLon(lon)
    setPickMode(false)
    // Re-show signal modal after picking
    setTimeout(() => setSignalModalOpen(true), 100)
  }

  function handleMarkerClick(pingId: string) {
    setSelectedPing(pingId)
    setSidebarOpen(true)
  }

  async function handleLogout() {
    try {
      await logout()
    } catch {
      /* ignore */
    }
    logoutStore()
    navigate('/')
  }

  function handlePingCreated(ping: Ping) {
    addPing(ping)
    setSignalModalOpen(false)
    setSignalEverOpened(false)
    setPickedLat(null)
    setPickedLon(null)
    setSelectedPing(ping.id)
    setSidebarOpen(true)
  }

  function handleSignalCancel() {
    setSignalModalOpen(false)
    setSignalEverOpened(false)
    setPickedLat(null)
    setPickedLon(null)
  }

  return (
    <div className="map-page">
      {/* Carte plein écran */}
      <div className={`map-container${pickMode ? ' map-container--pick' : ''}`}>
        <MapView onMapPick={pickMode ? handleMapPick : null} onMarkerClick={handleMarkerClick} />
      </div>

      {/* Bouton FAB — logo flottant en haut à droite pour ouvrir la sidebar */}
      {!sidebarOpen && !signalModalOpen && (
        <button
          className="map-fab"
          onClick={() => setSidebarOpen(true)}
          aria-label="Ouvrir le menu"
        >
          <img src="/logo.png" alt="FeedThemAll" className="map-fab__logo" />
        </button>
      )}

      {/* Loader géolocalisation */}
      {geoLoading && (
        <div className="map-geo-loader">
          <span className="map-geo-loader__spinner" />
          <span className="map-geo-loader__label">Recherche de votre position…</span>
        </div>
      )}

      {/* Toast géolocalisation */}
      {geoError && <div className="map-toast map-toast--warning">📍 {geoError}</div>}

      {/* Bandeau mode pick */}
      {pickMode && (
        <div className="map-pick-hint">
          🗺 Cliquez sur la carte pour placer le marqueur
          <button
            className="map-pick-hint__cancel"
            onClick={() => {
              setPickMode(false)
              setSignalModalOpen(true)
            }}
          >
            ✕
          </button>
        </div>
      )}

      {/* SignalForm modal centré — toujours monté une fois ouvert, caché via CSS */}
      {signalEverOpened && (
        <div className={`signal-modal-wrapper${signalModalOpen ? '' : ' signal-modal-wrapper--hidden'}`}>
          <SignalForm
            onDone={handlePingCreated}
            onCancel={handleSignalCancel}
            onRequestMapPick={handleRequestMapPick}
            pickedLat={pickedLat}
            pickedLon={pickedLon}
          />
        </div>
      )}

      {/* Sidebar slideout droite */}
      <MapSidebar
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onOpenSignalModal={handleOpenSignalModal}
        onLogout={handleLogout}
      />
    </div>
  )
}
