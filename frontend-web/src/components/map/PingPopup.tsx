// src/components/map/PingPopup.tsx — Popup affichée au clic sur un marqueur.
//
// Affiche le détail d'un ping : type, date, dernier nourrissage, photos.
// Bouton "J'ai nourri" → ouvre le formulaire FeedForm.
// Bouton "Confirmer présence" → appel direct API.
import { useState, useEffect } from 'react'
import { Popup } from 'react-leaflet'
import {
  confirmPing,
  getPingMedia,
  getPingFeedings,
  type Ping,
  type PingMedia,
  type FeedingEvent,
} from '../../api/pings'
import { useAuthStore } from '../../store/auth'
import { useMapStore } from '../../store/map'
import FeedForm from './FeedForm'

interface Props {
  ping: Ping
}

const ANIMAL_LABELS: Record<string, string> = {
  cat: 'Chat',
  dog: 'Chien',
  other: 'Autre animal',
}

export default function PingPopup({ ping }: Props) {
  const { user } = useAuthStore()
  const updatePing = useMapStore(s => s.updatePing)

  const [media, setMedia] = useState<PingMedia[]>([])
  const [feedings, setFeedings] = useState<FeedingEvent[]>([])
  const [confirming, setConfirming] = useState(false)
  const [showFeedForm, setShowFeedForm] = useState(false)

  useEffect(() => {
    getPingMedia(ping.id)
      .then(setMedia)
      .catch(() => {})
    getPingFeedings(ping.id)
      .then(setFeedings)
      .catch(() => {})
  }, [ping.id])

  async function handleConfirm() {
    if (!user) return
    setConfirming(true)
    try {
      await confirmPing(ping.id)
    } finally {
      setConfirming(false)
    }
  }

  function onFeedDone(updatedPing: Ping) {
    updatePing(updatedPing)
    setShowFeedForm(false)
    getPingFeedings(ping.id)
      .then(setFeedings)
      .catch(() => {})
  }

  const lastFed = feedings[0]
  const isAnimal = ping.type === 'animal'

  return (
    <Popup minWidth={220} maxWidth={300}>
      <div className="ping-popup">
        {/* En-tête */}
        <div className="ping-popup__header">
          <span className={`ping-popup__type ping-popup__type--${ping.type}`}>
            {isAnimal
              ? (ANIMAL_LABELS[ping.animal_type ?? 'other'] ?? 'Animal')
              : 'Nourriture disponible'}
          </span>
          {isAnimal && ping.animal_count > 1 && (
            <span className="ping-popup__count">×{ping.animal_count}</span>
          )}
        </div>

        {/* Date */}
        <p className="ping-popup__date">
          Signalé le{' '}
          {new Date(ping.created_at).toLocaleDateString('fr-FR', {
            day: 'numeric',
            month: 'short',
            hour: '2-digit',
            minute: '2-digit',
          })}
        </p>

        {/* Dernier nourrissage */}
        {lastFed ? (
          <p className="ping-popup__fed ping-popup__fed--ok">
            ✅ Nourri par <strong>{lastFed.username}</strong>{' '}
            {new Date(lastFed.created_at).toLocaleDateString('fr-FR', {
              day: 'numeric',
              month: 'short',
              hour: '2-digit',
              minute: '2-digit',
            })}
            {lastFed.note && <span className="ping-popup__note"> — «{lastFed.note}»</span>}
          </p>
        ) : (
          <p className="ping-popup__fed ping-popup__fed--none">Pas encore nourri</p>
        )}

        {/* Photos */}
        {media.length > 0 && (
          <div className="ping-popup__media">
            {media.slice(0, 3).map(m => (
              <img key={m.id} src={`/api${m.url}`} alt="Photo ping" className="ping-popup__photo" />
            ))}
          </div>
        )}

        {/* Actions (utilisateur connecté seulement) */}
        {user && !showFeedForm && (
          <div className="ping-popup__actions">
            {isAnimal && (
              <button
                className="ping-popup__btn ping-popup__btn--feed"
                onClick={() => setShowFeedForm(true)}
              >
                🍽 J'ai nourri
              </button>
            )}
            <button
              className="ping-popup__btn ping-popup__btn--confirm"
              disabled={confirming}
              onClick={handleConfirm}
            >
              👁 Confirmer présence
            </button>
          </div>
        )}

        {/* Formulaire nourrissage inline */}
        {showFeedForm && (
          <FeedForm ping={ping} onDone={onFeedDone} onCancel={() => setShowFeedForm(false)} />
        )}
      </div>
    </Popup>
  )
}
