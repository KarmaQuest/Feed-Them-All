// src/components/map/SignalForm.tsx — Modal centré "Signaler un animal".
//
// S'affiche en popup centré (85% largeur max) avec :
//   - Sélecteur de type : Animal / Nourriture
//   - Grille de sprites animaux avec barre de recherche (race)
//   - Choix position : GPS ou clic sur carte
//   - Boutons Annuler (ferme le modal) / Signaler (submit)
import { useState, useEffect, useMemo } from 'react'
import { createPing, listAnimalBreeds, type Ping, type AnimalGroup } from '../../api/pings'
import { useMapStore } from '../../store/map'
import { useAuthStore } from '../../store/auth'

interface Props {
  onDone: (ping: Ping) => void
  onCancel: () => void
  /** Appelé pour passer en mode "clic sur carte" */
  onRequestMapPick: () => void
  /** Position choisie via clic sur la carte (injectée depuis MapPage) */
  pickedLat?: number | null
  pickedLon?: number | null
}

const TYPE_LABELS: Record<string, string> = {
  dogs: 'Chiens',
  cats: 'Chats',
}

export default function SignalForm({ onDone, onCancel, onRequestMapPick, pickedLat, pickedLon }: Props) {
  const { user } = useAuthStore()
  const { userLat, userLon } = useMapStore()

  const [type, setType] = useState<'animal' | 'food'>('animal')
  const [animalType, setAnimalType] = useState<'cat' | 'dog' | 'other'>('cat')
  const [selectedBreed, setSelectedBreed] = useState<string | null>(null)
  const [animalCount, setAnimalCount] = useState(1)
  const [posMode, setPosMode] = useState<'gps' | 'map'>('gps')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [breedSearch, setBreedSearch] = useState('')
  const [breedGroups, setBreedGroups] = useState<AnimalGroup[]>([])
  const [breedsLoading, setBreedsLoading] = useState(true)

  // Load breeds on mount
  useEffect(() => {
    setBreedsLoading(true)
    listAnimalBreeds()
      .then(setBreedGroups)
      .catch(() => setBreedGroups([]))
      .finally(() => setBreedsLoading(false))
  }, [])

  // Filter breeds by search text
  const filteredGroups = useMemo(() => {
    if (!breedSearch.trim()) return breedGroups
    const q = breedSearch.toLowerCase().trim()
    return breedGroups
      .map(g => ({
        ...g,
        breeds: g.breeds.filter(b => b.toLowerCase().includes(q)),
      }))
      .filter(g => g.breeds.length > 0)
  }, [breedGroups, breedSearch])

  // Map animal type string to the groups we have
  const currentTypeKey = animalType === 'dog' ? 'dogs' : animalType === 'cat' ? 'cats' : null

  const effectiveLat = posMode === 'map' && pickedLat != null ? pickedLat : userLat
  const effectiveLon = posMode === 'map' && pickedLon != null ? pickedLon : userLon

  async function handleSubmit() {
    if (!user) {
      setError('Vous devez être connecté.')
      return
    }
    if (effectiveLat == null || effectiveLon == null) {
      setError('Position non disponible.')
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
          animal_breed: selectedBreed ?? undefined,
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

  function handlePickOnMap() {
    setPosMode('map')
    onRequestMapPick()
  }

  return (
    <div className="signal-modal-overlay" onClick={onCancel}>
      <div className="signal-modal" onClick={e => e.stopPropagation()}>
        {/* Header */}
        <div className="signal-modal__header">
          <h2 className="signal-modal__title">Signaler</h2>
          <button className="signal-modal__close" onClick={onCancel}>✕</button>
        </div>

        <div className="signal-modal__body">
          {/* Type selector */}
          <div className="signal-modal__row">
            <label className="signal-modal__label">Type</label>
            <div className="signal-modal__type-toggle">
              <button
                className={`signal-modal__type-btn ${type === 'animal' ? 'active' : ''}`}
                onClick={() => setType('animal')}
              >
                🐾 Animal
              </button>
              <button
                className={`signal-modal__type-btn ${type === 'food' ? 'active' : ''}`}
                onClick={() => setType('food')}
              >
                🍖 Nourriture
              </button>
            </div>
          </div>

          {type === 'animal' && (
            <>
              {/* Species selector */}
              <div className="signal-modal__row">
                <label className="signal-modal__label">Espèce</label>
                <div className="signal-modal__type-toggle">
                  <button
                    className={`signal-modal__type-btn ${animalType === 'cat' ? 'active' : ''}`}
                    onClick={() => { setAnimalType('cat'); setSelectedBreed(null) }}
                  >
                    😺 Chat
                  </button>
                  <button
                    className={`signal-modal__type-btn ${animalType === 'dog' ? 'active' : ''}`}
                    onClick={() => { setAnimalType('dog'); setSelectedBreed(null) }}
                  >
                    🐶 Chien
                  </button>
                  <button
                    className={`signal-modal__type-btn ${animalType === 'other' ? 'active' : ''}`}
                    onClick={() => { setAnimalType('other'); setSelectedBreed(null) }}
                  >
                    ❓ Autre
                  </button>
                </div>
              </div>

              {/* Breed grid (only for cat/dog) */}
              {currentTypeKey && (
                <div className="signal-modal__row signal-modal__row--breeds">
                  <label className="signal-modal__label">Race (optionnel)</label>

                  {/* Search bar */}
                  <div className="signal-modal__search">
                    <input
                      type="text"
                      className="signal-modal__search-input"
                      placeholder="Rechercher une race..."
                      value={breedSearch}
                      onChange={e => setBreedSearch(e.target.value)}
                    />
                    {breedSearch && (
                      <button
                        className="signal-modal__search-clear"
                        onClick={() => setBreedSearch('')}
                      >
                        ✕
                      </button>
                    )}
                  </div>

                  {/* Sprite grid */}
                  <div className="signal-modal__sprite-grid">
                    {breedsLoading ? (
                      <p className="signal-modal__empty">Chargement des sprites…</p>
                    ) : filteredGroups.length === 0 ? (
                      <p className="signal-modal__empty">Aucun sprite animal disponible</p>
                    ) : (
                      filteredGroups.map(group => (
                        group.breeds.map(breed => {
                          const spriteSrc = `/api/sprites/default/animals/${group.type}/${breed}.png`
                          const isSelected = selectedBreed === breed
                          return (
                            <button
                              key={breed}
                              className={`signal-modal__sprite-item ${isSelected ? 'selected' : ''}`}
                              onClick={() => setSelectedBreed(isSelected ? null : breed)}
                              title={breed}
                            >
                              <img
                                src={spriteSrc}
                                alt={breed}
                                className="signal-modal__sprite-img"
                              />
                              <span className="signal-modal__sprite-name">{breed}</span>
                            </button>
                          )
                        })
                      ))
                    )}
                  </div>
                </div>
              )}

              {/* Animal count */}
              <div className="signal-modal__row">
                <label className="signal-modal__label">Nombre d'animaux</label>
                <input
                  type="number"
                  className="signal-modal__input"
                  min={1}
                  max={50}
                  value={animalCount}
                  onChange={e => setAnimalCount(Number(e.target.value))}
                />
              </div>
            </>
          )}

          {/* Position */}
          <div className="signal-modal__row">
            <label className="signal-modal__label">Position</label>
            <div className="signal-modal__type-toggle">
              <button
                className={`signal-modal__type-btn ${posMode === 'gps' ? 'active' : ''}`}
                onClick={() => setPosMode('gps')}
              >
                📍 Ma position
              </button>
              <button
                className={`signal-modal__type-btn ${posMode === 'map' ? 'active' : ''}`}
                onClick={handlePickOnMap}
              >
                🗺 Choisir sur la carte
              </button>
            </div>
          </div>

          {/* Coords display */}
          <div className="signal-modal__coords">
            {posMode === 'gps' && effectiveLat != null ? (
              <span>📍 {effectiveLat.toFixed(5)}, {effectiveLon?.toFixed(5)}</span>
            ) : posMode === 'map' && pickedLat != null ? (
              <span className="signal-modal__coords--picked">
                ✔ {pickedLat.toFixed(5)}, {pickedLon?.toFixed(5)}
              </span>
            ) : posMode === 'map' ? (
              <span className="signal-modal__coords--hint">
                Cliquez sur la carte pour choisir l'emplacement
              </span>
            ) : (
              <span className="signal-modal__coords--warn">Position GPS indisponible</span>
            )}
          </div>

          {error && <p className="signal-modal__error">{error}</p>}
        </div>

        {/* Footer */}
        <div className="signal-modal__footer">
          <button className="btn btn--style-red" onClick={onCancel}>
            Annuler
          </button>
          <button
            className="btn btn--style-yellow"
            disabled={submitting}
            onClick={handleSubmit}
          >
            {submitting ? '…' : 'Signaler'}
          </button>
        </div>
      </div>
    </div>
  )
}
