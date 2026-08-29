// src/pages/UserLoginPage.tsx — Page de connexion utilisateur (tous rôles).
//
// Différente de LoginPage (admin) : pas de vérification de rôle,
// redirige vers / (la carte) après connexion réussie.
// Lien vers /register pour les nouveaux utilisateurs.
import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { login } from '../api/auth'
import { useAuthStore } from '../store/auth'
import AuthForm, { type AuthFormFields } from '../components/AuthForm'

export default function UserLoginPage() {
  const navigate = useNavigate()
  const loginStore = useAuthStore(s => s.login)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit({ email, password }: AuthFormFields) {
    setError('')
    setLoading(true)
    try {
      const data = await login(email, password)
      loginStore({ id: data.user.id, username: data.user.username, role: data.user.role, roles: data.user.roles, avatar_config: data.user.avatar_config })
      navigate('/')
    } catch {
      setError('Email ou mot de passe incorrect.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <AuthForm
      subtitle="Connexion"
      submitLabel="Se connecter"
      loading={loading}
      error={error}
      onSubmit={handleSubmit}
      footer={
        <p>
          Pas encore de compte ?{' '}
          <Link to="/register" className="auth-link">
            S'inscrire
          </Link>
        </p>
      }
    />
  )
}
