// src/components/ProtectedRoute.tsx — Redirige vers /login si non connecté ou non admin.
import { Navigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'

interface Props {
  children: React.ReactNode
}

export default function ProtectedRoute({ children }: Props) {
  const user = useAuthStore((s) => s.user)
  if (!user || user.role !== 'admin') {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}
