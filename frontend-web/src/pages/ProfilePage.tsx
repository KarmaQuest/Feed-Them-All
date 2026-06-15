// src/pages/ProfilePage.tsx — Page de profil utilisateur.
//
// /profile       → propre profil (utilise user.id du store auth)
// /profile/:id   → profil public d'un autre utilisateur
//
// Avatar : carré coloré avec initiales (couleur dérivée du username)
// XP bar : progression vers le level suivant
// Si profil privé + non-propriétaire → écran "Profil privé"
import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useAuthStore } from '../store/auth'
import { getUserProfile, updatePrivacy, type UserProfile, type PrivateProfile } from '../api/users'
import './ProfilePage.css'

// Palette de couleurs pour les avatars initiales
const AVATAR_COLORS = [
  '#6366f1', '#8b5cf6', '#ec4899', '#ef4444',
  '#f59e0b', '#10b981', '#06b6d4', '#3b82f6',
]

function getAvatarColor(username: string): string {
  let hash = 0
  for (let i = 0; i < username.length; i++) {
    hash = username.charCodeAt(i) + ((hash << 5) - hash)
  }
  return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length]
}

function getInitials(username: string): string {
  return username.slice(0, 2).toUpperCase()
}

function XpBar({ xp, level, thresholds }: { xp: number; level: number; thresholds: number[] }) {
  const currentMin = thresholds[level - 1] ?? 0
  const nextMin = thresholds[level] ?? null
  const progress = nextMin != null
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
        <p className="prof-xp__next">encore {xpToNext.toLocaleString()} XP pour le niveau {level + 1}</p>
      ) : (
        <p className="prof-xp__next">Niveau maximum atteint !</p>
      )}
    </div>
  )
}

// Paliers par défaut (synchronisés avec le backend)
const DEFAULT_THRESHOLDS = [0, 100, 250, 500, 900, 1400, 2100, 3000, 4500, 7000]

export default function ProfilePage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user, initialized } = useAuthStore()

  const [profile, setProfile] = useState<UserProfile | PrivateProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [privacyLoading, setPrivacyLoading] = useState(false)

  // Détermine l'ID cible : si pas de :id dans l'URL → propre profil
  const targetId = id ?? user?.id ?? null

  const isOwnProfile = user?.id === targetId

  useEffect(() => {
    if (!initialized) return
    if (!targetId) {
      // Non connecté et pas d'ID dans l'URL → rediriger vers login
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
          <Link to="/" className="prof-back-link">← Retour à la carte</Link>
        </div>
      </div>
    )
  }

  // Profil privé vu par un tiers
  if (profile.is_private && !isOwnProfile) {
    return (
      <div className="prof-page">
        <div className="prof-private">
          <div
            className="prof-avatar"
            style={{ background: getAvatarColor(profile.username) }}
          >
            {getInitials(profile.username)}
          </div>
          <h2 className="prof-username">{profile.username}</h2>
          <p className="prof-private__msg">🔒 Ce profil est privé</p>
          <p className="prof-private__level">Niveau {profile.level}</p>
          <Link to="/" className="prof-back-link">← Retour à la carte</Link>
        </div>
      </div>
    )
  }

  // Profil complet
  const full = profile as UserProfile

  return (
    <div className="prof-page">
      <div className="prof-card">
        {/* Header */}
        <div className="prof-header">
          <Link to="/" className="prof-back">← Carte</Link>
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
          <div
            className="prof-avatar prof-avatar--large"
            style={{ background: getAvatarColor(full.username) }}
          >
            {getInitials(full.username)}
          </div>
          <div className="prof-identity__info">
            <h1 className="prof-username">{full.username}</h1>
            <span className="prof-role">{full.role}</span>
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
            <span className="prof-stat-item__label">Nourrissage{full.nb_feedings > 1 ? 's' : ''}</span>
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
              {full.badges.map((b) => (
                <div key={b.slug} className="prof-badge" title={b.label}>
                  <span className="prof-badge__icon">🏅</span>
                  <span className="prof-badge__label">{b.label}</span>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}
