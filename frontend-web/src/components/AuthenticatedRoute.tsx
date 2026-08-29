// src/components/AuthenticatedRoute.tsx — Redirige vers /user-login si non connecté.
// Utilisé pour les pages qui nécessitent une session (profile, avatar, shop).
import { Navigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'

interface Props {
  children: React.ReactNode
}

export default function AuthenticatedRoute({ children }: Props) {
  const { user, initialized } = useAuthStore()
  if (!initialized) return null
  if (!user) {
    return <Navigate to="/user-login" replace />
  }
  return <>{children}</>
}
