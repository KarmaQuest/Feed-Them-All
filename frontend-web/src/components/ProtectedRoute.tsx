// src/components/ProtectedRoute.tsx — Redirige vers /login si non connecté ou non admin.
// Attend que initialize() soit terminé (initialized=true) avant de décider,
// pour éviter une redirection prématurée pendant la restauration de session.
import { Navigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'

interface Props {
  children: React.ReactNode
}

export default function ProtectedRoute({ children }: Props) {
  const { user, initialized } = useAuthStore()
  if (!initialized) return null
  if (!user || user.role !== 'admin') {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}
