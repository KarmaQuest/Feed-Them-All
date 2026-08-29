// src/pages/AvatarPage.tsx — Page de personnalisation d'avatar.
//
// Sélecteur visuel de skin, tenue, accessoire avec preview en direct.
// Les items disponibles viennent de GET /users/me/inventory.
// Les changements sont sauvegardés via PATCH /users/me/avatar.
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'
import AvatarSprite from '../components/avatar/AvatarSprite'
import { getInventory, type ShopItem, type InventoryItem } from '../api/shop'
import { getUserProfile, updateAvatar as apiUpdateAvatar } from '../api/users'
import './AvatarPage.css'

const CATEGORY_LABELS: Record<string, string> = {
  skin: 'Skin',
  outfit: 'Tenue',
  accessory: 'Accessoire',
}

export default function AvatarPage() {
  const { user } = useAuthStore()
  const navigate = useNavigate()

  const [inventory, setInventory] = useState<InventoryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const [selected, setSelected] = useState<Record<string, string>>({
    skin: 'skin_default',
    outfit: 'outfit_default',
    accessory: 'accessory_none',
  })
  const [gender, setGender] = useState<'male' | 'female'>('male')

  // Load inventory and current config
  useEffect(() => {
    if (!user) { navigate('/user-login'); return }

    async function load() {
      setLoading(true)
      try {
        const [inv, profile] = await Promise.all([
          getInventory(),
          getUserProfile(user.id),
        ])
        setInventory(inv)

        // Restore saved config
        if ('avatar_config' in profile && profile.avatar_config) {
          const cfg = profile.avatar_config as Record<string, string>
          if (cfg.skin) setSelected(prev => ({ ...prev, skin: cfg.skin! }))
          if (cfg.outfit) setSelected(prev => ({ ...prev, outfit: cfg.outfit! }))
          if (cfg.accessory) setSelected(prev => ({ ...prev, accessory: cfg.accessory! }))
          if (cfg.gender === 'female') setGender('female')
        }
      } catch {
        setError('Erreur chargement')
      }
      setLoading(false)
    }
    load()
  }, [user, navigate])

  // Group inventory items by category
  const byCategory = inventory.reduce<Record<string, ShopItem[]>>((acc, inv) => {
    const cat = inv.item.category
    if (!acc[cat]) acc[cat] = []
    acc[cat].push(inv.item)
    return acc
  }, {})

  async function handleSave() {
    setSaving(true)
    setError('')
    setSuccess('')
    try {
      await apiUpdateAvatar({
        gender,
        skin: selected.skin,
        outfit: selected.outfit,
        accessory: selected.accessory,
      })
      setSuccess('Avatar sauvegardé !')
      setTimeout(() => setSuccess(''), 3000)
    } catch {
      setError('Erreur lors de la sauvegarde')
    }
    setSaving(false)
  }

  if (loading) {
    return (
      <div className="avatar-page">
        <div className="avatar-page__loader">Chargement…</div>
      </div>
    )
  }

  return (
    <div className="avatar-page">
      <div className="avatar-page__header">
        <button className="avatar-page__back" onClick={() => navigate(-1)}>← Retour</button>
        <h1 className="avatar-page__title">Personnaliser mon avatar</h1>
      </div>

      <div className="avatar-page__content">
        {/* Preview panel */}
        <div className="avatar-page__preview">
          <div className="avatar-page__avatar-box">
            <AvatarSprite
              config={{
                gender,
                skin: selected.skin,
                outfit: selected.outfit,
                accessory: selected.accessory,
              }}
              size="lg"
            />
          </div>

          {/* Gender toggle */}
          <div className="avatar-page__gender">
            <button
              className={`avatar-page__gender-btn ${gender === 'male' ? 'active' : ''}`}
              onClick={() => setGender('male')}
            >
              ♂ Masculin
            </button>
            <button
              className={`avatar-page__gender-btn ${gender === 'female' ? 'active' : ''}`}
              onClick={() => setGender('female')}
            >
              ♀ Féminin
            </button>
          </div>
        </div>

        {/* Selector panel */}
        <div className="avatar-page__selectors">
          {['skin', 'outfit', 'accessory'].map(category => (
            <div key={category} className="avatar-page__category">
              <h3 className="avatar-page__category-title">{CATEGORY_LABELS[category]}</h3>
              <div className="avatar-page__item-grid">
                {(byCategory[category] ?? []).map(item => {
                  const isSelected = selected[category] === item.slug
                  const hasSprite = category !== 'skin' // skins have no sprite
                  const spriteUrl = hasSprite
                    ? `/api/sprites/shop/${item.slug}/south.png`
                    : undefined

                  return (
                    <button
                      key={item.id}
                      className={`avatar-page__item ${isSelected ? 'selected' : ''}`}
                      onClick={() => setSelected(prev => ({ ...prev, [category]: item.slug }))}
                      title={item.name}
                    >
                      {spriteUrl ? (
                        <img
                          src={spriteUrl}
                          alt={item.name}
                          className="avatar-page__item-img"
                          onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
                        />
                      ) : (
                        <div className="avatar-page__item-placeholder">?</div>
                      )}
                      <span className="avatar-page__item-name">{item.name}</span>
                    </button>
                  )
                })}
                {!byCategory[category]?.length && (
                  <p className="avatar-page__empty">Aucun item disponible</p>
                )}
              </div>
            </div>
          ))}

          {error && <p className="avatar-page__error">{error}</p>}
          {success && <p className="avatar-page__success">{success}</p>}

          <div className="avatar-page__actions">
            <button
              className="avatar-page__btn avatar-page__btn--save"
              disabled={saving}
              onClick={handleSave}
            >
              {saving ? '…' : 'Sauvegarder'}
            </button>
            <button
              className="btn btn--style-red"
              onClick={() => navigate(-1)}
            >
              Annuler
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
