// src/components/map/MapSidebar.tsx — Sidebar slideout droite de la carte.
//
// Remplace la modal SignalForm et le PingPopup Leaflet.
// Trois panneaux gérés en interne :
//   'nav'    — navigation utilisateur (défaut)
//   'signal' — SignalForm (créer un ping)
//   'ping'   — détails d'un ping + liste des activités
//
// Le bouton d'ouverture est le logo dans MapPage (icon-only topbar).
// Le panneau 'ping' s'ouvre automatiquement quand selectedPingId change.
import { useState, useEffect } from 'react'
import { useAuthStore } from '../../store/auth'
import { useMapStore } from '../../store/map'
import {
  getPingMedia,
  getPingFeedings,
  confirmPing,
  type Ping,
  type PingMedia,
  type FeedingEvent,
} from '../../api/pings'
import SignalForm from './SignalForm'
import FeedForm from './FeedForm'

type Panel = 'nav' | 'signal' | 'ping'

const ANIMAL_LABELS: Record<string, string> = {
  cat: 'Chat',
  dog: 'Chien',
  other: 'Autre animal',
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

interface Props {
  open: boolean
  onClose: () => void
  /** Mode pick carte actif (masque la sidebar sans la démonter) */
  pickMode: boolean
  pickedLat: number | null
  pickedLon: number | null
  onRequestMapPick: () => void
  onPingCreated: (ping: Ping) => void
  onLogout: () => void
}

export default function MapSidebar({
  open,
  onClose,
  pickMode,
  pickedLat,
  pickedLon,
  onRequestMapPick,
  onPingCreated,
  onLogout,
}: Props) {
  const { user } = useAuthStore()
  const { pings, selectedPingId, setSelectedPing, updatePing } = useMapStore()

  const [panel, setPanel] = useState<Panel>('nav')

  // Bascule automatiquement vers le panneau ping quand un marqueur est cliqué
  useEffect(() => {
    if (selectedPingId) setPanel('ping')
  }, [selectedPingId])

  const selectedPing = pings.find((p) => p.id === selectedPingId) ?? null

  // ── Panneau nav ────────────────────────────────────────────────────────────
  function NavPanel() {
    return (
      <div className="msb-panel">
        <h3 className="msb-panel__title">Menu</h3>
        {user ? (
          <div className="msb-nav">
            <div className="msb-user-badge">
              <span className="msb-user-badge__name">{user.username}</span>
              <span className="msb-user-badge__role">{user.role}</span>
            </div>
            <button
              className="msb-btn msb-btn--primary"
              onClick={() => setPanel('signal')}
            >
              + Signaler un animal
            </button>
            {user.role === 'admin' && (
              <a href="/admin" className="msb-btn msb-btn--secondary">
                ⚙ Dashboard admin
              </a>
            )}
            <button className="msb-btn msb-btn--ghost" onClick={onLogout}>
              Déconnexion
            </button>
          </div>
        ) : (
          <div className="msb-nav">
            <p className="msb-nav__hint">
              Connecte-toi pour signaler des animaux et suivre les nourrissages.
            </p>
            <a href="/user-login" className="msb-btn msb-btn--primary">
              Connexion
            </a>
            <a href="/register" className="msb-btn msb-btn--secondary">
              Créer un compte
            </a>
          </div>
        )}
      </div>
    )
  }

  // ── Panneau signal ─────────────────────────────────────────────────────────
  function SignalPanel() {
    return (
      <div className="msb-panel">
        <button
          className="msb-back"
          onClick={() => setPanel('nav')}
        >
          ← Retour
        </button>
        <SignalForm
          onDone={(ping) => {
            onPingCreated(ping)
            setPanel('nav')
          }}
          onCancel={() => setPanel('nav')}
          onRequestMapPick={onRequestMapPick}
          pickedLat={pickedLat}
          pickedLon={pickedLon}
          inline
        />
      </div>
    )
  }

  // ── Panneau ping ───────────────────────────────────────────────────────────
  function PingPanel({ ping }: { ping: Ping }) {
    const [media, setMedia] = useState<PingMedia[]>([])
    const [feedings, setFeedings] = useState<FeedingEvent[]>([])
    const [confirming, setConfirming] = useState(false)
    const [showFeedForm, setShowFeedForm] = useState(false)

    useEffect(() => {
      getPingMedia(ping.id).then(setMedia).catch(() => {})
      getPingFeedings(ping.id).then(setFeedings).catch(() => {})
    }, [ping.id])

    async function handleConfirm() {
      if (!user) return
      setConfirming(true)
      try { await confirmPing(ping.id) } finally { setConfirming(false) }
    }

    function onFedDone(updated: Ping) {
      updatePing(updated)
      setShowFeedForm(false)
      getPingFeedings(ping.id).then(setFeedings).catch(() => {})
    }

    const isAnimal = ping.type === 'animal'

    return (
      <div className="msb-panel">
        <button className="msb-back" onClick={() => { setSelectedPing(null); setPanel('nav') }}>
          ← Retour
        </button>

        {/* En-tête ping */}
        <div className="msb-ping-header">
          <div className="msb-ping-header__type">
            {isAnimal
              ? ANIMAL_LABELS[ping.animal_type ?? 'other']
              : 'Nourriture disponible'}
            {isAnimal && ping.animal_count > 1 && (
              <span className="msb-ping-header__count"> ×{ping.animal_count}</span>
            )}
          </div>
          <div className="msb-ping-header__date">
            Signalé le {formatDate(ping.created_at)}
          </div>
        </div>

        {/* Photos */}
        {media.length > 0 && (
          <div className="msb-media">
            {media.slice(0, 4).map((m) => (
              <img key={m.id} src={`/api${m.url}`} alt="" className="msb-media__img" />
            ))}
          </div>
        )}

        {/* Actions */}
        {user && !showFeedForm && (
          <div className="msb-ping-actions">
            {isAnimal && (
              <button
                className="msb-btn msb-btn--primary"
                onClick={() => setShowFeedForm(true)}
              >
                🍽 J'ai nourri
              </button>
            )}
            <button
              className="msb-btn msb-btn--secondary"
              disabled={confirming}
              onClick={handleConfirm}
            >
              👁 Confirmer présence
            </button>
          </div>
        )}

        {showFeedForm && (
          <FeedForm
            ping={ping}
            onDone={onFedDone}
            onCancel={() => setShowFeedForm(false)}
          />
        )}

        {/* Activités */}
        <div className="msb-activities">
          <h4 className="msb-activities__title">
            Activités — {ping.id.slice(0, 8)}…
          </h4>
          <p className="msb-activities__meta">
            {ping.type} · {ping.animal_type ?? ''} · Créé le{' '}
            {new Date(ping.created_at).toLocaleDateString('fr-FR')}
          </p>

          {feedings.length === 0 ? (
            <p className="msb-activities__empty">Aucune activité enregistrée.</p>
          ) : (
            <div className="msb-activity-list">
              {feedings.map((f, i) => (
                <div key={f.id} className="msb-activity-item">
                  <div className="msb-activity-item__head">
                    <span className="msb-activity-item__num">#{i + 1}</span>
                    <span className="msb-activity-item__user">par {f.username}</span>
                    <span className="msb-activity-item__date">{formatDate(f.fed_at)}</span>
                  </div>
                  {f.animal_count_seen != null && (
                    <p className="msb-activity-item__count">
                      {f.animal_count_seen} animal(s) vu(s)
                    </p>
                  )}
                  {f.note && (
                    <p className="msb-activity-item__note">"{f.note}"</p>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Ajouter une activité */}
          {user && !showFeedForm && isAnimal && (
            <button
              className="msb-btn msb-btn--ghost msb-btn--small"
              onClick={() => setShowFeedForm(true)}
            >
              + Ajouter une activité
            </button>
          )}
        </div>
      </div>
    )
  }

  const isVisible = open && !pickMode

  return (
    <>
      {/* Overlay semi-transparent (mobile) */}
      {isVisible && (
        <div className="msb-overlay" onClick={onClose} />
      )}

      {/* Sidebar */}
      <aside className={`msb${isVisible ? ' msb--open' : ''}`}>
        {/* Header */}
        <div className="msb-header">
          <img src="/logo.png" alt="FeedThemAll" className="msb-header__logo" />
          <span className="msb-header__title">FeedThemAll</span>
          <button className="msb-header__close" onClick={onClose} aria-label="Fermer">✕</button>
        </div>

        {/* Corps */}
        <div className="msb-body">
          {panel === 'nav' && <NavPanel />}
          {panel === 'signal' && <SignalPanel />}
          {panel === 'ping' && selectedPing && <PingPanel ping={selectedPing} />}
          {panel === 'ping' && !selectedPing && <NavPanel />}
        </div>
      </aside>
    </>
  )
}
