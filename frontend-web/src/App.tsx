// src/App.tsx — Router principal de l'application FeedThemAll.
//
// Routes :
//   /       -> MapPage (publique — carte interactive)
//   /login  -> LoginPage (publique)
//   /admin  -> AdminPage (protégée, rôle admin requis)
//   /*      -> Redirige vers /
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import AdminPage from './pages/admin/AdminPage'
import MapPage from './pages/MapPage'
import ProtectedRoute from './components/ProtectedRoute'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<MapPage />} />
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
