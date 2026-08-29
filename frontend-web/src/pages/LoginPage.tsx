// src/pages/LoginPage.tsx — Page de connexion admin.
//
// Utilise AuthForm. Vérifie que le rôle est 'admin' avant de rediriger vers /admin.
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { login } from '../api/auth'
import { useAuthStore } from '../store/auth'
import AuthForm, { type AuthFormFields } from '../components/AuthForm'

export default function LoginPage() {
  const navigate = useNavigate()
  const loginStore = useAuthStore(s => s.login)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit({ email, password }: AuthFormFields) {
    setError('')
    setLoading(true)
    try {
      const data = await login(email, password)
      if (data.user.role !== 'admin') {
        setError('Accès refusé — compte admin requis.')
        return
      }
      loginStore({ id: data.user.id, username: data.user.username, role: data.user.role, roles: data.user.roles })
      navigate('/admin')
    } catch {
      setError('Email ou mot de passe incorrect.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <AuthForm
      subtitle="Dashboard Admin"
      submitLabel="Se connecter"
      loading={loading}
      error={error}
      onSubmit={handleSubmit}
    />
  )
}
