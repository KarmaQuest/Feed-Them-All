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
import BadgesSection from './sections/BadgesSection'
import ShopSection from './sections/ShopSection'
import ModerationSection from './sections/ModerationSection'
import './AdminPage.css'

type Section = 'users' | 'xp' | 'badges' | 'shop' | 'moderation'

const NAV: { id: Section; label: string; icon: string }[] = [
  { id: 'users', label: 'Utilisateurs', icon: '👥' },
  { id: 'xp', label: 'XP & Levels', icon: '⚡' },
  { id: 'badges', label: 'Badges', icon: '🏅' },
  { id: 'shop', label: 'Boutique', icon: '🛒' },
  { id: 'moderation', label: 'Modération', icon: '🛡' },
]

export default function AdminPage() {
  const [active, setActive] = useState<Section>('users')
  const navigate = useNavigate()
  const { user, logout: logoutStore } = useAuthStore()

  async function handleLogout() {
    try { await logout() } catch { /* ignore */ }
    logoutStore()
    navigate('/login')
  }

  function renderSection() {
    switch (active) {
      case 'users': return <UsersSection />
      case 'xp': return <XPSection />
      case 'badges': return <BadgesSection />
      case 'shop': return <ShopSection />
      case 'moderation': return <ModerationSection />
    }
  }

  return (
    <div className="admin-layout">
      <aside className="admin-sidebar">
        <div className="sidebar-logo">
          <h2>FeedThemAll</h2>
          <span>Admin Dashboard</span>
        </div>
        <nav className="sidebar-nav">
          {NAV.map((n) => (
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
          <button className="btn-logout" onClick={handleLogout}>Déconnexion</button>
        </div>
      </aside>
      <main className="admin-main">
        {renderSection()}
      </main>
    </div>
  )
}
