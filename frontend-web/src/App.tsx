// src/App.tsx — Router principal de l'application FeedThemAll.
//
// Routes :
//   /            -> MapPage (publique — carte interactive)
//   /user-login  -> UserLoginPage (connexion utilisateur, tous rôles)
//   /register    -> RegisterPage (inscription, rôle feeder par défaut)
//   /login       -> LoginPage (connexion admin uniquement)
//   /admin       -> AdminPage (protégée, rôle admin requis)
//   /*           -> Redirige vers /
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useEffect } from 'react'
import LoginPage from './pages/LoginPage'
import UserLoginPage from './pages/UserLoginPage'
import RegisterPage from './pages/RegisterPage'
import AdminPage from './pages/admin/AdminPage'
import MapPage from './pages/MapPage'
import ProtectedRoute from './components/ProtectedRoute'
import { useAuthStore } from './store/auth'

export default function App() {
  const initialize = useAuthStore((s) => s.initialize)

  // Au démarrage : tente de restaurer la session via le refresh token cookie
  useEffect(() => { initialize() }, [initialize])

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<MapPage />} />
        <Route path="/user-login" element={<UserLoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/login" element={<LoginPage />} />
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
    </BrowserRouter>
  )
}
