// src/App.tsx — Router principal de l'application FeedThemAll.
//
// Routes actives :
//   /login  -> LoginPage (publique)
//   /admin  -> AdminPage (protegee, role admin requis)
//   /*      -> Redirige vers /admin
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import AdminPage from './pages/admin/AdminPage'
import ProtectedRoute from './components/ProtectedRoute'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/admin"
          element={
            <ProtectedRoute>
              <AdminPage />
            </ProtectedRoute>
          }
        />
        <Route path="*" element={<Navigate to="/admin" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
