// src/components/map/MapSidebar.tsx
//
// Trois panneaux gérés via un state 'panel' :
//   'nav'    — navigation utilisateur (stats, login, signaler, admin, déconnexion)
//   'signal' — SignalForm (créer un ping)
//   'ping'   — détails d'un ping + activités + formulaire nourrissage
//
// IMPORTANT : NavPanel, SignalPanel et PingPanel sont déclarés AU NIVEAU MODULE
// (hors de MapSidebar) pour éviter que React ne les recrée à chaque rendu,
// ce qui causerait un démontage/remontage et la perte des états locaux.
import { useState, useEffect } from 'react'
import { useAuthStore } from '../../store/auth'
import { useMapStore } from '../../store/map'
import { useWebSocketStore } from '../../store/websocket'
import {
  getPingMedia,
  getPingFeedings,
  confirmPing,
  deactivatePing,
  updatePing as apiUpdatePing,
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

// ── Panneau nav ────────────────────────────────────────────────────────────────
interface NavPanelProps {
  onSignal: () => void
  onLogout: () => void
}
function NavPanel({ onSignal, onLogout }: NavPanelProps) {
  const { user } = useAuthStore()
  const { pings } = useMapStore()
  const connected = useWebSocketStore((s) => s.connected)
  const animalCount = pings.filter((p) => p.type === 'animal').length
  const foodCount = pings.filter((p) => p.type === 'food').length

  return (
    <div className="msb-panel">
      <div className="msb-stats">
        <span className="msb-stat msb-stat--animal">🐾 {animalCount} signalement{animalCount > 1 ? 's' : ''}</span>
        <span className="msb-stat msb-stat--food">🍖 {foodCount} nourriture{foodCount > 1 ? 's' : ''}</span>
        {user && (
          <span className={`msb-stat msb-stat--ws ${connected ? 'msb-stat--online' : 'msb-stat--offline'}`}>
            {connected ? '● Live' : '○ Hors ligne'}
          </span>
        )}
      </div>

      <div className="msb-divider" />

      {user ? (
        <div className="msb-nav">
          <div className="msb-user-badge">
            <span className="msb-user-badge__name">{user.username}</span>
            <span className="msb-user-badge__role">{user.role}</span>
          </div>
          <a href={`/profile`} className="msb-btn msb-btn--secondary">👤 Mon profil</a>
          <button className="msb-btn msb-btn--primary" onClick={onSignal}>
            + Signaler un animal
          </button>
          {user.role === 'admin' && (
            <a href="/admin" className="msb-btn msb-btn--secondary">⚙️ Dashboard admin</a>
          )}
          <button className="msb-btn msb-btn--ghost" onClick={onLogout}>Déconnexion</button>
        </div>
      ) : (
        <div className="msb-nav">
          <p className="msb-nav__hint">
            Connecte-toi pour signaler des animaux et suivre les nourrissages.
          </p>
          <a href="/user-login" className="msb-btn msb-btn--primary">Connexion</a>
          <a href="/register" className="msb-btn msb-btn--secondary">Créer un compte</a>
        </div>
      )}
    </div>
  )
}

// ── Panneau signal ─────────────────────────────────────────────────────────────
interface SignalPanelProps {
  onBack: () => void
  onDone: (ping: Ping) => void
  onRequestMapPick: () => void
  pickedLat: number | null
  pickedLon: number | null
}
function SignalPanel({ onBack, onDone, onRequestMapPick, pickedLat, pickedLon }: SignalPanelProps) {
  return (
    <div className="msb-panel">
      <button className="msb-back" onClick={onBack}>← Retour</button>
      <SignalForm
        onDone={onDone}
        onCancel={onBack}
        onRequestMapPick={onRequestMapPick}
        pickedLat={pickedLat}
        pickedLon={pickedLon}
        inline
      />
    </div>
  )
}

// ── Panneau ping ───────────────────────────────────────────────────────────────
interface PingPanelProps {
  ping: Ping
  onBack: () => void
}
function PingPanel({ ping, onBack }: PingPanelProps) {
  const { user } = useAuthStore()
  const { updatePing, removePing, setSelectedPing } = useMapStore()

  const [media, setMedia] = useState<PingMedia[]>([])
  const [feedings, setFeedings] = useState<FeedingEvent[]>([])
  const [confirming, setConfirming] = useState(false)
  const [confirmDone, setConfirmDone] = useState(false)
  const [showFeedForm, setShowFeedForm] = useState(false)

  // Edit state (Point 2)
  const [editing, setEditing] = useState(false)
  const [editAnimalType, setEditAnimalType] = useState<string>(ping.animal_type ?? 'other')
  const [editAnimalCount, setEditAnimalCount] = useState<number>(ping.animal_count)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const isOwner = user?.id === ping.created_by

  useEffect(() => {
    getPingMedia(ping.id).then(setMedia).catch(() => {})
    getPingFeedings(ping.id).then(setFeedings).catch(() => {})
  }, [ping.id])

  async function handleConfirm() {
    if (!user || confirming) return
    setConfirming(true)
    try {
      await confirmPing(ping.id)
      setConfirmDone(true)
    } finally {
      setConfirming(false)
    }
  }

  function onFedDone(updated: Ping) {
    updatePing(updated)
    setShowFeedForm(false)
    getPingFeedings(ping.id).then(setFeedings).catch(() => {})
  }

  async function handleSaveEdit() {
    if (saving) return
    setSaving(true)
    try {
      const updated = await apiUpdatePing(ping.id, {
        animal_type: editAnimalType as 'cat' | 'dog' | 'other',
        animal_count: editAnimalCount,
      })
      updatePing(updated)
      setEditing(false)
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (deleting) return
    setDeleting(true)
    try {
      await deactivatePing(ping.id)
      removePing(ping.id)
      setSelectedPing(null)
      onBack()
    } finally {
      setDeleting(false)
    }
  }

  const isAnimal = ping.type === 'animal'

  return (
    <div className="msb-panel">
      <button className="msb-back" onClick={onBack}>← Retour</button>

      {/* En-tête ping */}
      <div className="msb-ping-header">
        <div className="msb-ping-header__type">
          {isAnimal ? (ANIMAL_LABELS[ping.animal_type ?? 'other'] ?? 'Animal') : 'Nourriture disponible'}
          {isAnimal && ping.animal_count > 1 && (
            <span className="msb-ping-header__count"> ×{ping.animal_count}</span>
          )}
        </div>
        <div className="msb-ping-header__date">Signalé le {formatDate(ping.created_at)}</div>
        {isOwner && !editing && (
          <button className="msb-btn msb-btn--ghost msb-btn--small" onClick={() => setEditing(true)}>
            ✏️ Modifier
          </button>
        )}
      </div>

      {/* Formulaire d'édition (Point 2) */}
      {editing && isAnimal && (
        <div className="msb-edit-form">
          <label className="msb-edit-form__label">
            Espèce
            <select
              className="msb-edit-form__select"
              value={editAnimalType}
              onChange={(e) => setEditAnimalType(e.target.value)}
            >
              <option value="cat">Chat</option>
              <option value="dog">Chien</option>
              <option value="other">Autre</option>
            </select>
          </label>
          <label className="msb-edit-form__label">
            Nombre
            <input
              type="number"
              className="msb-edit-form__input"
              min={1}
              max={100}
              value={editAnimalCount}
              onChange={(e) => setEditAnimalCount(Number(e.target.value))}
            />
          </label>
          <div className="msb-edit-form__actions">
            <button className="msb-btn msb-btn--primary msb-btn--small" onClick={handleSaveEdit} disabled={saving}>
              {saving ? '...' : 'Enregistrer'}
            </button>
            <button className="msb-btn msb-btn--ghost msb-btn--small" onClick={() => setEditing(false)}>
              Annuler
            </button>
          </div>
        </div>
      )}

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
            <button className="msb-btn msb-btn--primary" onClick={() => setShowFeedForm(true)}>
              🍽️ J'ai nourri
            </button>
          )}
          <button
            className="msb-btn msb-btn--secondary"
            disabled={confirming || confirmDone}
            onClick={handleConfirm}
          >
            {confirmDone ? '✔ Présence confirmée' : confirming ? '...' : '👍 Confirmer présence'}
          </button>
          {isOwner && (
            confirmDelete ? (
              <div className="msb-delete-confirm">
                <span>Supprimer ce ping ?</span>
                <button className="msb-btn msb-btn--danger msb-btn--small" onClick={handleDelete} disabled={deleting}>
                  {deleting ? '...' : 'Confirmer'}
                </button>
                <button className="msb-btn msb-btn--ghost msb-btn--small" onClick={() => setConfirmDelete(false)}>
                  Annuler
                </button>
              </div>
            ) : (
              <button className="msb-btn msb-btn--ghost msb-btn--small" onClick={() => setConfirmDelete(true)}>
                🗑️ Supprimer
              </button>
            )
          )}
        </div>
      )}

      {/* Formulaire nourrissage */}
      {showFeedForm && (
        <FeedForm
          ping={ping}
          onDone={onFedDone}
          onCancel={() => setShowFeedForm(false)}
        />
      )}

      {/* Activités */}
      <div className="msb-activities">
        <h4 className="msb-activities__title">Activités</h4>

        {feedings.length === 0 ? (
          <p className="msb-activities__empty">Aucune activité enregistrée.</p>
        ) : (
          <div className="msb-activity-list">
            {feedings.map((f) => (
              <div key={f.id} className={`msb-activity-item${f.event_type === 'signal' ? ' msb-activity-item--signal' : ''}`}>
                <div className="msb-activity-item__head">
                  <span className="msb-activity-item__icon">
                    {f.event_type === 'signal' ? '📍' : '🍽️'}
                  </span>
                  <span className="msb-activity-item__user">
                    {f.event_type === 'signal' ? `Signalé par ${f.username}` : `Nourri par ${f.username}`}
                  </span>
                  <span className="msb-activity-item__date">{formatDate(f.fed_at)}</span>
                </div>
                {f.animal_count_seen != null && (
                  <p className="msb-activity-item__count">{f.animal_count_seen} animal(s) vu(s)</p>
                )}
                {f.note && <p className="msb-activity-item__note">"{f.note}"</p>}
              </div>
            ))}
          </div>
        )}

        {user && !showFeedForm && isAnimal && (
          <button className="msb-btn msb-btn--ghost msb-btn--small" onClick={() => setShowFeedForm(true)}>
            + Ajouter une activité
          </button>
        )}
      </div>
    </div>
  )
}

// ── Composant principal MapSidebar ─────────────────────────────────────────────
interface Props {
  open: boolean
  onClose: () => void
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
  const { pings, selectedPingId, setSelectedPing } = useMapStore()
  const [panel, setPanel] = useState<Panel>('nav')

  useEffect(() => {
    if (selectedPingId) setPanel('ping')
  }, [selectedPingId])

  const selectedPing = pings.find((p) => p.id === selectedPingId) ?? null
  const isVisible = open && !pickMode

  function handleBack() {
    setSelectedPing(null)
    setPanel('nav')
  }

  return (
    <>
      {isVisible && <div className="msb-overlay" onClick={onClose} />}

      <aside className={`msb${isVisible ? ' msb--open' : ''}`}>
        <div className="msb-header">
          <img src="/logo.png" alt="FeedThemAll" className="msb-header__logo" />
          <span className="msb-header__title">FeedThemAll</span>
          <button className="msb-header__close" onClick={onClose} aria-label="Fermer">✕</button>
        </div>

        <div className="msb-body">
          {panel === 'nav' && (
            <NavPanel onSignal={() => setPanel('signal')} onLogout={onLogout} />
          )}
          {panel === 'signal' && (
            <SignalPanel
              onBack={() => setPanel('nav')}
              onDone={(ping) => {
                onPingCreated(ping)
                setSelectedPing(ping.id)
                setPanel('ping')
              }}
              onRequestMapPick={onRequestMapPick}
              pickedLat={pickedLat}
              pickedLon={pickedLon}
            />
          )}
          {panel === 'ping' && selectedPing && (
            <PingPanel ping={selectedPing} onBack={handleBack} />
          )}
          {panel === 'ping' && !selectedPing && (
            <NavPanel onSignal={() => setPanel('signal')} onLogout={onLogout} />
          )}
        </div>
      </aside>
    </>
  )
}
