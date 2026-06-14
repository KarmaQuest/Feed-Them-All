// src/components/map/SignalForm.tsx — Formulaire "Signaler un animal".
//
// Deux modes de position :
//   "gps"  → coordonnées pré-remplies depuis la géolocalisation browser
//   "map"  → l'utilisateur clique sur la carte pour choisir l'emplacement
import { useState } from 'react'
import { createPing, type Ping } from '../../api/pings'
import { useMapStore } from '../../store/map'
import { useAuthStore } from '../../store/auth'

interface Props {
  onDone: (ping: Ping) => void
  onCancel: () => void
  /** Appelé par MapPage pour passer en mode "clic sur carte" */
  onRequestMapPick: () => void
  /** Position choisie via clic sur la carte (injectée depuis MapPage) */
  pickedLat?: number | null
  pickedLon?: number | null
}

export default function SignalForm({ onDone, onCancel, onRequestMapPick, pickedLat, pickedLon }: Props) {
  const { user } = useAuthStore()
  const { userLat, userLon } = useMapStore()

  const [type, setType] = useState<'animal' | 'food'>('animal')
  const [animalType, setAnimalType] = useState<'cat' | 'dog' | 'other'>('cat')
  const [animalCount, setAnimalCount] = useState(1)
  const [posMode, setPosMode] = useState<'gps' | 'map'>('gps')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  // Coordonnées effectives selon le mode
  const effectiveLat = posMode === 'map' && pickedLat != null
    ? pickedLat
    : userLat
  const effectiveLon = posMode === 'map' && pickedLon != null
    ? pickedLon
    : userLon

  async function handleSubmit() {
    if (!user) { setError('Vous devez être connecté.'); return }
    if (effectiveLat == null || effectiveLon == null) {
      setError('Position non disponible. Activez la géolocalisation ou cliquez sur la carte.')
      return
    }
    setError('')
    setSubmitting(true)
    try {
      const ping = await createPing({
        type,
        lat: effectiveLat,
        lon: effectiveLon,
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

        {/* Choix du mode de position */}
        <div className="modal-field">
          <label>Position</label>
          <div className="signal-pos-toggle">
            <button
              type="button"
              className={`signal-pos-btn ${posMode === 'gps' ? 'active' : ''}`}
              onClick={() => setPosMode('gps')}
            >
              📍 Ma position GPS
            </button>
            <button
              type="button"
              className={`signal-pos-btn ${posMode === 'map' ? 'active' : ''}`}
              onClick={() => { setPosMode('map'); onRequestMapPick() }}
            >
              🗺 Choisir sur la carte
            </button>
          </div>
        </div>

        {/* Affichage des coordonnées effectives */}
        <div className="signal-coords">
          {posMode === 'gps' && effectiveLat != null ? (
            <span>📍 {effectiveLat.toFixed(5)}, {effectiveLon?.toFixed(5)}</span>
          ) : posMode === 'map' && pickedLat != null ? (
            <span className="signal-coords--picked">✔ {pickedLat.toFixed(5)}, {pickedLon?.toFixed(5)}</span>
          ) : posMode === 'map' ? (
            <span className="signal-coords--hint">Cliquez sur la carte pour choisir l'emplacement</span>
          ) : (
            <span className="signal-coords--warn">Position GPS indisponible</span>
          )}
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
