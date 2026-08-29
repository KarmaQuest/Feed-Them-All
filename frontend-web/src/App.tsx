// src/App.tsx — Router principal de l'application FeedThemAll.
//
// Routes :
//   /            -> MapPage (publique — carte interactive)
//   /user-login  -> UserLoginPage (connexion utilisateur, tous rôles)
//   /register    -> RegisterPage (inscription, rôle feeder par défaut)
//   /login       -> LoginPage (connexion admin uniquement)
//   /admin       -> AdminPage (protégée, rôle admin requis)
//   /profile     -> ProfilePage (propre profil, JWT requis)
//   /profile/:id -> ProfilePage (profil public d'un autre utilisateur)
//   /avatar      -> redirige vers /profile
//   /shop        -> ShopPage (boutique avatar, JWT requis)
//   /quests      -> QuestsPage (items gratuits à débloquer, JWT requis)
//   /*           -> Redirige vers /
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { useEffect } from 'react'
import LoginPage from './pages/LoginPage'
import UserLoginPage from './pages/UserLoginPage'
import RegisterPage from './pages/RegisterPage'
import AdminPage from './pages/admin/AdminPage'
import MapPage from './pages/MapPage'
import ProfilePage from './pages/ProfilePage'
import ShopPage from './pages/ShopPage'
import QuestsPage from './pages/QuestsPage'
import ProtectedRoute from './components/ProtectedRoute'
import AuthenticatedRoute from './components/AuthenticatedRoute'
import { useAuthStore } from './store/auth'

// Composant interne pour accéder à useNavigate (doit être enfant de BrowserRouter)
function AppRoutes() {
  const initialize = useAuthStore(s => s.initialize)
  const logout = useAuthStore(s => s.logout)
  const navigate = useNavigate()

  // Au démarrage : tente de restaurer la session via le refresh token cookie
  useEffect(() => {
    initialize()
  }, [initialize])

  // Écoute l'event dispatché par l'intercepteur axios quand le refresh échoue.
  // Utilise React Router navigate (SPA, sans rechargement) pour éviter la boucle
  // deco/reco Firefox causée par window.location.href (rechargement complet).
  useEffect(() => {
    function handleSessionExpired() {
      logout()
      navigate('/user-login', { replace: true })
    }
    window.addEventListener('auth:session-expired', handleSessionExpired)
    return () => window.removeEventListener('auth:session-expired', handleSessionExpired)
  }, [logout, navigate])

  return (
    <Routes>
      <Route path="/" element={<MapPage />} />
      <Route path="/user-login" element={<UserLoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/profile" element={<ProfilePage />} />
      <Route path="/profile/:id" element={<ProfilePage />} />
      <Route path="/avatar" element={<Navigate to="/profile" replace />} />
      <Route path="/shop" element={<AuthenticatedRoute><ShopPage /></AuthenticatedRoute>} />
      <Route path="/quests" element={<AuthenticatedRoute><QuestsPage /></AuthenticatedRoute>} />
      <Route
        path="/admin"
        element={
          <ProtectedRoute>
            <AdminPage />
          </ProtectedRoute>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  )
}
