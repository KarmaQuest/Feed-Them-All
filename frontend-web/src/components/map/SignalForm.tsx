// src/components/map/SignalForm.tsx — Formulaire "Signaler un animal".
//
// Modal affichée depuis MapPage quand l'utilisateur clique "Signaler".
// La position GPS est pré-remplie depuis useMapStore (géolocalisation).
// L'utilisateur peut ajuster lat/lon manuellement.
import { useState } from 'react'
import { createPing, type Ping } from '../../api/pings'
import { useMapStore } from '../../store/map'
import { useAuthStore } from '../../store/auth'

interface Props {
  onDone: (ping: Ping) => void
  onCancel: () => void
}

export default function SignalForm({ onDone, onCancel }: Props) {
  const { user } = useAuthStore()
  const { userLat, userLon } = useMapStore()

  const [type, setType] = useState<'animal' | 'food'>('animal')
  const [animalType, setAnimalType] = useState<'cat' | 'dog' | 'other'>('cat')
  const [animalCount, setAnimalCount] = useState(1)
  const [lat, setLat] = useState(userLat?.toFixed(6) ?? '')
  const [lon, setLon] = useState(userLon?.toFixed(6) ?? '')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit() {
    if (!user) { setError('Vous devez être connecté.'); return }
    const parsedLat = parseFloat(lat)
    const parsedLon = parseFloat(lon)
    if (isNaN(parsedLat) || isNaN(parsedLon)) {
      setError('Coordonnées GPS invalides.')
      return
    }
    setError('')
    setSubmitting(true)
    try {
      const ping = await createPing({
        type,
        lat: parsedLat,
        lon: parsedLon,
        ...(type === 'animal' && {
          animal_type: animalType,
          animal_count: animalCount,
        }),
      })
      onDone(ping)
    } catch {
      setError('Erreur lors de la création du ping.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal-box signal-form" onClick={(e) => e.stopPropagation()}>
        <h3>Signaler</h3>

        {/* Type */}
        <div className="modal-field">
          <label>Type</label>
          <select
            className="inline-select"
            value={type}
            onChange={(e) => setType(e.target.value as 'animal' | 'food')}
          >
            <option value="animal">Animal</option>
            <option value="food">Nourriture disponible</option>
          </select>
        </div>

        {/* Sous-type animal */}
        {type === 'animal' && (
          <>
            <div className="modal-field">
              <label>Espèce</label>
              <select
                className="inline-select"
                value={animalType}
                onChange={(e) => setAnimalType(e.target.value as 'cat' | 'dog' | 'other')}
              >
                <option value="cat">Chat</option>
                <option value="dog">Chien</option>
                <option value="other">Autre</option>
              </select>
            </div>
            <div className="modal-field">
              <label>Nombre d'animaux</label>
              <input
                type="number"
                min={1}
                max={50}
                value={animalCount}
                onChange={(e) => setAnimalCount(Number(e.target.value))}
                className="inline-input"
              />
            </div>
          </>
        )}

        {/* Coordonnées */}
        <div className="modal-field">
          <label>Latitude</label>
          <input
            type="text"
            value={lat}
            onChange={(e) => setLat(e.target.value)}
            className="inline-input"
            placeholder="48.8566"
          />
        </div>
        <div className="modal-field">
          <label>Longitude</label>
          <input
            type="text"
            value={lon}
            onChange={(e) => setLon(e.target.value)}
            className="inline-input"
            placeholder="2.3522"
          />
        </div>

        {error && <p className="error-msg">{error}</p>}

        <div className="modal-actions">
          <button className="btn-cancel" onClick={onCancel}>Annuler</button>
          <button className="btn-submit" disabled={submitting} onClick={handleSubmit}>
            {submitting ? '...' : 'Signaler'}
          </button>
        </div>
      </div>
    </div>
  )
}
