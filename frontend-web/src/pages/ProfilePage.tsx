// src/pages/ProfilePage.tsx — Page de profil utilisateur.
//
// /profile       → propre profil (utilise user.id du store auth)
// /profile/:id   → profil public d'un autre utilisateur
//
// Avatar : AvatarSprite avec config stockée en DB (gender, skin, etc.)
// XP bar : progression vers le level suivant
// Si profil privé + non-propriétaire → écran "Profil privé"
// Si propre profil → section personnalisation avatar en dessous des badges
import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useAuthStore } from '../store/auth'
import { getUserProfile, updatePrivacy, updateAvatar, type UserProfile, type PrivateProfile } from '../api/users'
import { getInventory, type ShopItem, type InventoryItem } from '../api/shop'
import AvatarSprite from '../components/avatar/AvatarSprite'
import { formatRoles } from '../utils/roles'
import './ProfilePage.css'

function XpBar({ xp, level, thresholds }: { xp: number; level: number; thresholds: number[] }) {
  const currentMin = thresholds[level - 1] ?? 0
  const nextMin = thresholds[level] ?? null
  const progress =
    nextMin != null
      ? Math.min(100, Math.round(((xp - currentMin) / (nextMin - currentMin)) * 100))
      : 100
  const xpToNext = nextMin != null ? nextMin - xp : 0

  return (
    <div className="prof-xp">
      <div className="prof-xp__header">
        <span className="prof-xp__label">Niveau {level}</span>
        <span className="prof-xp__value">{xp.toLocaleString()} XP</span>
      </div>
      <div className="prof-xp__bar-track">
        <div className="prof-xp__bar-fill" style={{ width: `${progress}%` }} />
      </div>
      {nextMin != null ? (
        <p className="prof-xp__next">
          encore {xpToNext.toLocaleString()} XP pour le niveau {level + 1}
        </p>
      ) : (
        <p className="prof-xp__next">Niveau maximum atteint !</p>
      )}
    </div>
  )
}

const DEFAULT_THRESHOLDS = [0, 100, 250, 500, 900, 1400, 2100, 3000, 4500, 7000]
const CATEGORY_LABELS: Record<string, string> = {
  skin: 'Skin',
  outfit: 'Tenue',
  accessory: 'Accessoire',
}

export default function ProfilePage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user, initialized } = useAuthStore()

  const [profile, setProfile] = useState<UserProfile | PrivateProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [privacyLoading, setPrivacyLoading] = useState(false)

  // Avatar customization state (own profile only)
  const [inventory, setInventory] = useState<InventoryItem[]>([])
  const [selected, setSelected] = useState<Record<string, string>>({
    skin: 'skin_default',
    outfit: 'outfit_default',
    accessory: 'accessory_none',
  })
  const [gender, setGender] = useState<'male' | 'female'>('male')
  const [saving, setSaving] = useState(false)
  const [avatarSaved, setAvatarSaved] = useState(false)

  const targetId = id ?? user?.id ?? null
  const isOwnProfile = user?.id === targetId

  useEffect(() => {
    if (!initialized) return
    if (!targetId) {
      navigate('/user-login', { replace: true })
      return
    }
    setLoading(true)
    setError('')
    getUserProfile(targetId)
      .then(setProfile)
      .catch(() => setError('Profil introuvable.'))
      .finally(() => setLoading(false))
  }, [targetId, initialized, navigate])

  // Load inventory + restore avatar config for own profile
  useEffect(() => {
    if (!isOwnProfile || !profile || !('avatar_config' in profile)) return
    const cfg = (profile as UserProfile).avatar_config as Record<string, string> | undefined
    async function load() {
      try {
        const inv = await getInventory()
        setInventory(inv)
      } catch {}
      if (cfg) {
        if (cfg.skin) setSelected(prev => ({ ...prev, skin: cfg.skin! }))
        if (cfg.outfit) setSelected(prev => ({ ...prev, outfit: cfg.outfit! }))
        if (cfg.accessory) setSelected(prev => ({ ...prev, accessory: cfg.accessory! }))
        if (cfg.gender === 'female') setGender('female')
      }
    }
    load()
  }, [isOwnProfile, profile])

  async function handleTogglePrivacy() {
    if (!profile || !('nb_pings' in profile) || privacyLoading) return
    setPrivacyLoading(true)
    try {
      await updatePrivacy(!profile.is_private)
      setProfile({ ...profile, is_private: !profile.is_private })
    } finally {
      setPrivacyLoading(false)
    }
  }

  async function handleSaveAvatar() {
    setSaving(true)
    setAvatarSaved(false)
    try {
      await updateAvatar({ gender, skin: selected.skin, outfit: selected.outfit, accessory: selected.accessory })
      setAvatarSaved(true)
      setTimeout(() => setAvatarSaved(false), 3000)
    } catch {}
    setSaving(false)
  }

  if (!initialized || loading) {
    return (
      <div className="prof-page">
        <div className="prof-loading">Chargement…</div>
      </div>
    )
  }

  if (error || !profile) {
    return (
      <div className="prof-page">
        <div className="prof-error">
          <p>{error || 'Profil introuvable.'}</p>
          <Link to="/" className="prof-back-link">
            ← Retour à la carte
          </Link>
        </div>
      </div>
    )
  }

  if (profile.is_private && !isOwnProfile) {
    return (
      <div className="prof-page">
        <div className="prof-private">
          <AvatarSprite
            config={'avatar_config' in profile ? (profile as UserProfile).avatar_config : undefined}
            size="lg"
          />
          <h2 className="prof-username">{profile.username}</h2>
          <p className="prof-private__msg">🔒 Ce profil est privé</p>
          <p className="prof-private__level">Niveau {profile.level}</p>
          <Link to="/" className="prof-back-link">
            ← Retour à la carte
          </Link>
        </div>
      </div>
    )
  }

  const full = profile as UserProfile

  // Group inventory by category
  const byCategory = inventory.reduce<Record<string, ShopItem[]>>((acc, inv) => {
    const cat = inv.item.category
    if (!acc[cat]) acc[cat] = []
    acc[cat].push(inv.item)
    return acc
  }, {})

  return (
    <div className="prof-page">
      <div className="prof-card">
        {/* Header */}
        <div className="prof-header">
          <Link to="/" className="prof-back">
            ← Carte
          </Link>
          {isOwnProfile && (
            <button
              className={`prof-privacy-btn${full.is_private ? ' prof-privacy-btn--private' : ''}`}
              onClick={handleTogglePrivacy}
              disabled={privacyLoading}
            >
              {privacyLoading ? '…' : full.is_private ? '🔒 Privé' : '🌐 Public'}
            </button>
          )}
        </div>

        {/* Avatar + identité */}
        <div className="prof-identity">
          <AvatarSprite config={full.avatar_config} size="lg" />
          <div className="prof-identity__info">
            <h1 className="prof-username">{full.username}</h1>
            <span className="prof-role">{formatRoles(full.roles, full.role)}</span>
          </div>
        </div>

        {/* XP Bar */}
        <XpBar xp={full.xp} level={full.level} thresholds={DEFAULT_THRESHOLDS} />

        {/* Stats */}
        <div className="prof-stats">
          <div className="prof-stat-item">
            <span className="prof-stat-item__icon">📍</span>
            <span className="prof-stat-item__value">{full.nb_pings}</span>
            <span className="prof-stat-item__label">Signal{full.nb_pings > 1 ? 's' : ''}</span>
          </div>
          <div className="prof-stat-item">
            <span className="prof-stat-item__icon">🍽️</span>
            <span className="prof-stat-item__value">{full.nb_feedings}</span>
            <span className="prof-stat-item__label">
              Nourrissage{full.nb_feedings > 1 ? 's' : ''}
            </span>
          </div>
          <div className="prof-stat-item">
            <span className="prof-stat-item__icon">🏅</span>
            <span className="prof-stat-item__value">{full.badges.length}</span>
            <span className="prof-stat-item__label">Badge{full.badges.length > 1 ? 's' : ''}</span>
          </div>
        </div>

        {/* Badges */}
        <section className="prof-section">
          <h3 className="prof-section__title">Badges</h3>
          {full.badges.length === 0 ? (
            <p className="prof-empty">Aucun badge pour l'instant.</p>
          ) : (
            <div className="prof-badges">
              {full.badges.map(b => (
                <div key={b.slug} className="prof-badge" title={b.label}>
                  <span className="prof-badge__icon">🏅</span>
                  <span className="prof-badge__label">{b.label}</span>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* ── Avatar customization (own profile only) ───────────────────── */}
        {isOwnProfile && (
          <section className="prof-section prof-avatar-custom">
            <h3 className="prof-section__title">Personnaliser mon avatar</h3>

            <div className="prof-avatar-custom__preview">
              <div className="prof-avatar-custom__box">
                <AvatarSprite
                  config={{ gender, skin: selected.skin, outfit: selected.outfit, accessory: selected.accessory }}
                  size="lg"
                />
              </div>
              <div className="prof-avatar-custom__gender">
                <button
                  className={`prof-avatar-custom__gender-btn ${gender === 'male' ? 'active' : ''}`}
                  onClick={() => setGender('male')}
                >
                  ♂ Masculin
                </button>
                <button
                  className={`prof-avatar-custom__gender-btn ${gender === 'female' ? 'active' : ''}`}
                  onClick={() => setGender('female')}
                >
                  ♀ Féminin
                </button>
              </div>
            </div>

            {['skin', 'outfit', 'accessory'].map(category => (
              <div key={category} className="prof-avatar-custom__category">
                <h4 className="prof-avatar-custom__cat-title">{CATEGORY_LABELS[category]}</h4>
                <div className="prof-avatar-custom__grid">
                  {(byCategory[category] ?? []).map(item => {
                    const isSelected = selected[category] === item.slug
                    const spriteUrl = category !== 'skin'
                      ? `/api/sprites/shop/${item.slug}/south.png`
                      : undefined
                    return (
                      <button
                        key={item.id}
                        className={`prof-avatar-custom__item ${isSelected ? 'selected' : ''}`}
                        onClick={() => setSelected(prev => ({ ...prev, [category]: item.slug }))}
                        title={item.name}
                      >
                        {spriteUrl ? (
                          <img
                            src={spriteUrl}
                            alt={item.name}
                            className="prof-avatar-custom__item-img"
                            onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
                          />
                        ) : (
                          <div className="prof-avatar-custom__item-placeholder">?</div>
                        )}
                        <span className="prof-avatar-custom__item-name">{item.name}</span>
                      </button>
                    )
                  })}
                  {!byCategory[category]?.length && (
                    <p className="prof-avatar-custom__empty">Aucun item disponible</p>
                  )}
                </div>
              </div>
            ))}

            <div className="prof-avatar-custom__actions">
              <button
                className="btn btn--style-yellow"
                disabled={saving}
                onClick={handleSaveAvatar}
              >
                {saving ? '…' : 'Sauvegarder'}
              </button>
              {avatarSaved && <span className="prof-avatar-custom__saved">✓ Sauvegardé !</span>}
            </div>
          </section>
        )}
      </div>
    </div>
  )
}
