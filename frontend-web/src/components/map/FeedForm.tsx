// src/components/map/FeedForm.tsx — Formulaire inline "J'ai nourri".
//
// Affiché à l'intérieur du PingPopup.
// Permet d'ajouter une note et un nombre d'animaux vus.
// Upload photo optionnel après la soumission.
import { useState, useRef } from 'react'
import { markFed, uploadPingMedia, type Ping } from '../../api/pings'

interface Props {
  ping: Ping
  onDone: (updatedPing: Ping) => void
  onCancel: () => void
}

export default function FeedForm({ ping, onDone, onCancel }: Props) {
  const [note, setNote] = useState('')
  const [count, setCount] = useState(ping.animal_count || 1)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [uploadWarning, setUploadWarning] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)

  async function handleSubmit() {
    setError('')
    setUploadWarning('')
    setSubmitting(true)
    try {
      await markFed(ping.id, note || undefined, count)
    } catch {
      setError('Erreur lors de l\'enregistrement du nourrissage.')
      setSubmitting(false)
      return
    }

    // Upload photo séparé — un échec n'annule pas le nourrissage
    const file = fileRef.current?.files?.[0]
    if (file) {
      try {
        await uploadPingMedia(ping.id, file)
      } catch {
        setUploadWarning('Nourrissage enregistré, mais l\'upload de la photo a échoué.')
      }
    }

    setSubmitting(false)
    const updated: Ping = { ...ping, fed_at: new Date().toISOString() }
    onDone(updated)
  }

  return (
    <div className="feed-form">
      <p className="feed-form__title">Enregistrer un nourrissage</p>

      <label className="feed-form__label">Animaux vus</label>
      <input
        type="number"
        min={1}
        max={50}
        value={count}
        onChange={(e) => setCount(Number(e.target.value))}
        className="feed-form__input"
      />

      <label className="feed-form__label">Note (optionnel)</label>
      <textarea
        value={note}
        onChange={(e) => setNote(e.target.value)}
        placeholder="Ex : 2 chats nourris, eau laissée..."
        maxLength={300}
        className="feed-form__textarea"
      />

      <label className="feed-form__label">Photo (optionnel)</label>
      <input
        ref={fileRef}
        type="file"
        accept="image/*"
        className="feed-form__file"
      />

      {error && <p className="feed-form__error">{error}</p>}
      {uploadWarning && <p className="feed-form__warning">{uploadWarning}</p>}

      <div className="feed-form__actions">
        <button
          className="ping-popup__btn ping-popup__btn--feed"
          disabled={submitting}
          onClick={handleSubmit}
        >
          {submitting ? '...' : 'Valider'}
        </button>
        <button
          className="ping-popup__btn ping-popup__btn--cancel"
          onClick={onCancel}
        >
          Annuler
        </button>
      </div>
    </div>
  )
}
