// src/pages/admin/AdminPage.tsx — Layout principal du dashboard admin.
//
// Sidebar avec 5 sections. Section active gérée en state local.
// Bouton logout dans le footer de la sidebar.
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { logout } from '../../api/auth'
import { useAuthStore } from '../../store/auth'
import UsersSection from './sections/UsersSection'
import XPSection from './sections/XPSection'
import LevelsSection from './sections/LevelsSection'
import BadgesSection from './sections/BadgesSection'
import ShopSection from './sections/ShopSection'
import ModerationSection from './sections/ModerationSection'
import SpritesSection from './sections/SpritesSection'
import './AdminPage.css'

type Section = 'users' | 'xp' | 'levels' | 'badges' | 'shop' | 'moderation' | 'sprites'

const NAV: { id: Section; label: string; icon: string }[] = [
  { id: 'users', label: 'Utilisateurs', icon: '👥' },
  { id: 'sprites', label: 'Sprites', icon: '🎨' },
  { id: 'xp', label: 'Actions XP', icon: '⚡' },
  { id: 'levels', label: 'Paliers de Level', icon: '📊' },
  { id: 'badges', label: 'Badges', icon: '🏅' },
  { id: 'shop', label: 'Boutique', icon: '🛒' },
  { id: 'moderation', label: 'Modération', icon: '🛡' },
]

export default function AdminPage() {
  const [active, setActive] = useState<Section>('users')
  const navigate = useNavigate()
  const { user, logout: logoutStore } = useAuthStore()

  async function handleLogout() {
    try {
      await logout()
    } catch {
      /* ignore */
    }
    logoutStore()
    navigate('/login')
  }

  function renderSection() {
    switch (active) {
      case 'users':
        return <UsersSection />
      case 'xp':
        return <XPSection />
      case 'levels':
        return <LevelsSection />
      case 'badges':
        return <BadgesSection />
      case 'shop':
        return <ShopSection />
      case 'moderation':
        return <ModerationSection />
      case 'sprites':
        return <SpritesSection />
    }
  }

  return (
    <div className="admin-layout">
      <aside className="admin-sidebar">
        <div className="sidebar-logo">
          <img
            src="/logo.png"
            alt="FeedThemAll"
            style={{ width: '72px', height: '72px', objectFit: 'contain' }}
          />
          <span>Admin Dashboard</span>
        </div>
        <nav className="sidebar-nav">
          {NAV.map(n => (
            <button
              key={n.id}
              className={`nav-item ${active === n.id ? 'active' : ''}`}
              onClick={() => setActive(n.id)}
            >
              <span className="nav-icon">{n.icon}</span>
              {n.label}
            </button>
          ))}
        </nav>
        <div className="sidebar-footer">
          <span>{user?.username}</span>
          <button className="btn btn--style-red" onClick={handleLogout}>
            Déconnexion
          </button>
        </div>
      </aside>
      <main className="admin-main">{renderSection()}</main>
    </div>
  )
}
