// src/components/map/MapSidebar.tsx
//
// Panneaux gérés via un state 'panel' :
//   'nav'  — navigation utilisateur (stats, signaler, admin, déconnexion)
//   'ping' — détails d'un ping + activités + formulaire nourrissage
//
// Le formulaire de signalement est désormais un modal centré (SignalForm)
// géré depuis MapPage. Le bouton "Signaler" appelle onOpenSignalModal().
import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuthStore } from '../../store/auth'
import AvatarSprite from '../avatar/AvatarSprite'
import { formatRoles } from '../../utils/roles'
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
import FeedForm from './FeedForm'

type Panel = 'nav' | 'ping'

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
  onOpenSignalModal: () => void
  onLogout: () => void
}
function NavPanel({ onOpenSignalModal, onLogout }: NavPanelProps) {
  const { user } = useAuthStore()
  const { pings } = useMapStore()
  const connected = useWebSocketStore(s => s.connected)
  const animalCount = pings.filter(p => p.type === 'animal').length
  const foodCount = pings.filter(p => p.type === 'food').length

  return (
    <div className="msb-panel">
      <div className="msb-stats">
        <span className="msb-stat msb-stat--animal">
          🐾 {animalCount} signalement{animalCount > 1 ? 's' : ''}
        </span>
        <span className="msb-stat msb-stat--food">
          🍖 {foodCount} nourriture{foodCount > 1 ? 's' : ''}
        </span>
        {user && (
          <span
            className={`msb-stat msb-stat--ws ${connected ? 'msb-stat--online' : 'msb-stat--offline'}`}
          >
            {connected ? '● Live' : '○ Hors ligne'}
          </span>
        )}
      </div>

      <div className="msb-divider" />

      {user ? (
        <div className="msb-nav">
          <div className="msb-user-badge">
            <AvatarSprite config={user.avatar_config} size="sm" />
            <div className="msb-user-badge__info">
              <span className="msb-user-badge__name">{user.username}</span>
              <span className="msb-user-badge__role">{formatRoles(user.roles, user.role)}</span>
            </div>
          </div>
          <Link to="/profile" state={{ fromMap: true }} className="btn btn--style-yellow btn--full">
            👤 Mon profil
          </Link>
          <Link to="/quests" state={{ fromMap: true }} className="btn btn--style-yellow btn--full">
            🎯 Quêtes
          </Link>
          <Link to="/shop" state={{ fromMap: true }} className="btn btn--style-yellow btn--full">
            🛒 Boutique
          </Link>
          <button className="btn btn--style-yellow btn--full" onClick={onOpenSignalModal}>
            + Signaler un animal
          </button>
          {user.role === 'admin' && (
            <Link to="/admin" state={{ fromMap: true }} className="btn btn--style-yellow btn--full">
              ⚙️ Dashboard admin
            </Link>
          )}
          <button className="btn btn--style-red btn--full" onClick={onLogout}>
            Déconnexion
          </button>
        </div>
      ) : (
        <div className="msb-nav">
          <p className="msb-nav__hint">
            Connecte-toi pour signaler des animaux et suivre les nourrissages.
          </p>
          <Link to="/user-login" className="btn btn--style-yellow btn--full">
            Connexion
          </Link>
          <Link to="/register" className="btn btn--style-yellow btn--full">
            Créer un compte
          </Link>
        </div>
      )}
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
  const [feedingsLoading, setFeedingsLoading] = useState(true)
  const [confirming, setConfirming] = useState(false)
  const [confirmDone, setConfirmDone] = useState(false)
  const [showFeedForm, setShowFeedForm] = useState(false)

  const [editing, setEditing] = useState(false)
  const [editAnimalType, setEditAnimalType] = useState<string>(ping.animal_type ?? 'other')
  const [editAnimalCount, setEditAnimalCount] = useState<number>(ping.animal_count)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const isOwner = user?.id === ping.created_by

  useEffect(() => {
    getPingMedia(ping.id)
      .then(setMedia)
      .catch(() => {})
    setFeedingsLoading(true)
    getPingFeedings(ping.id)
      .then(setFeedings)
      .catch(() => {})
      .finally(() => setFeedingsLoading(false))
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
    setFeedingsLoading(true)
    getPingFeedings(ping.id)
      .then(setFeedings)
      .catch(() => {})
      .finally(() => setFeedingsLoading(false))
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
      <button className="msb-back" onClick={onBack}>
        ← Retour
      </button>

      <div className="msb-ping-header">
        <div className="msb-ping-header__type">
          {isAnimal
            ? (ANIMAL_LABELS[ping.animal_type ?? 'other'] ?? 'Animal')
            : 'Nourriture disponible'}
          {isAnimal && ping.animal_count > 1 && (
            <span className="msb-ping-header__count"> ×{ping.animal_count}</span>
          )}
        </div>
        {ping.animal_breed && (
          <div className="msb-ping-header__breed">Race : {ping.animal_breed}</div>
        )}
        <div className="msb-ping-header__date">Signalé le {formatDate(ping.created_at)}</div>
        {isOwner && !editing && (
          <button
            className="msb-btn msb-btn--ghost msb-btn--small"
            onClick={() => setEditing(true)}
          >
            ✏️ Modifier
          </button>
        )}
      </div>

      {editing && isAnimal && (
        <div className="msb-edit-form">
          <label className="msb-edit-form__label">
            Espèce
            <select
              className="msb-edit-form__select"
              value={editAnimalType}
              onChange={e => setEditAnimalType(e.target.value)}
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
              onChange={e => setEditAnimalCount(Number(e.target.value))}
            />
          </label>
          <div className="msb-edit-form__actions">
            <button
              className="btn btn--style-yellow btn--sm"
              onClick={handleSaveEdit}
              disabled={saving}
            >
              {saving ? '...' : 'Enregistrer'}
            </button>
            <button
              className="btn btn--style-red btn--sm"
              onClick={() => setEditing(false)}
            >
              Annuler
            </button>
          </div>
        </div>
      )}

      {media.length > 0 && (
        <div className="msb-media">
          {media.slice(0, 4).map(m => (
            <img key={m.id} src={`/api${m.url}`} alt="" className="msb-media__img" />
          ))}
        </div>
      )}

      {user && !showFeedForm && (
        <div className="msb-ping-actions">
          {isAnimal && (
            <button className="btn btn--style-yellow" onClick={() => setShowFeedForm(true)}>
              🍽️ J'ai nourri
            </button>
          )}
          <button
            className="btn btn--style-yellow"
            disabled={confirming || confirmDone}
            onClick={handleConfirm}
          >
            {confirmDone ? '✔ Présence confirmée' : confirming ? '...' : '👍 Confirmer présence'}
          </button>
          {isOwner &&
            (confirmDelete ? (
              <div className="msb-delete-confirm">
                <span>Supprimer ce ping ?</span>
                <button
                  className="btn btn--style-red btn--sm"
                  onClick={handleDelete}
                  disabled={deleting}
                >
                  {deleting ? '...' : 'Confirmer'}
                </button>
                <button
                  className="btn btn--style-red btn--sm"
                  onClick={() => setConfirmDelete(false)}
                >
                  Annuler
                </button>
              </div>
            ) : (
              <button
                className="btn btn--style-red btn--sm"
                onClick={() => setConfirmDelete(true)}
              >
                🗑️ Supprimer
              </button>
            ))}
        </div>
      )}

      {showFeedForm && (
        <FeedForm ping={ping} onDone={onFedDone} onCancel={() => setShowFeedForm(false)} />
      )}

      <div className="msb-activities">
        <h4 className="msb-activities__title">Activités</h4>

        {feedingsLoading ? (
          <p className="msb-activities__loading">Chargement…</p>
        ) : feedings.length === 0 ? (
          <p className="msb-activities__empty">Aucune activité enregistrée.</p>
        ) : (
          <div className="msb-activity-list">
            {feedings.map(f => (
              <div
                key={f.id}
                className={`msb-activity-item${f.event_type === 'signal' ? ' msb-activity-item--signal' : ''}`}
              >
                <div className="msb-activity-item__head">
                  <span className="msb-activity-item__icon">
                    {f.event_type === 'signal' ? '📍' : '🍽️'}
                  </span>
                  <span className="msb-activity-item__user">
                    {f.event_type === 'signal'
                      ? `Signalé par ${f.username}`
                      : `Nourri par ${f.username}`}
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

// ── Composant principal MapSidebar ─────────────────────────────────────────────
interface Props {
  open: boolean
  onClose: () => void
  onOpenSignalModal: () => void
  onLogout: () => void
}

export default function MapSidebar({
  open,
  onClose,
  onOpenSignalModal,
  onLogout,
}: Props) {
  const { pings, selectedPingId, setSelectedPing } = useMapStore()
  const [panel, setPanel] = useState<Panel>('nav')

  useEffect(() => {
    if (selectedPingId) setPanel('ping')
  }, [selectedPingId])

  const selectedPing = pings.find(p => p.id === selectedPingId) ?? null

  function handleBack() {
    setSelectedPing(null)
    setPanel('nav')
  }

  return (
    <>
      {open && <div className="msb-overlay" onClick={onClose} />}

      <aside className={`msb${open ? ' msb--open' : ''}`}>
        <div className="msb-header">
          <img src="/logo.png" alt="FeedThemAll" className="msb-header__logo" />
          <span className="msb-header__title">FeedThemAll</span>
          <button className="msb-header__close" onClick={onClose} aria-label="Fermer">
            ✕
          </button>
        </div>

        <div className="msb-body">
          {panel === 'nav' && (
            <NavPanel onOpenSignalModal={onOpenSignalModal} onLogout={onLogout} />
          )}
          {panel === 'ping' && selectedPing && (
            <PingPanel ping={selectedPing} onBack={handleBack} />
          )}
          {panel === 'ping' && !selectedPing && (
            <NavPanel onOpenSignalModal={onOpenSignalModal} onLogout={onLogout} />
          )}
        </div>
      </aside>
    </>
  )
}
