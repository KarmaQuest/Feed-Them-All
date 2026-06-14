// src/pages/RegisterPage.tsx — Page d'inscription utilisateur.
//
// Crée un compte (rôle feeder par défaut), connecte automatiquement
// et redirige vers / (la carte) centrée sur la position de l'utilisateur.
// Lien vers /user-login pour les utilisateurs déjà inscrits.
import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { register } from '../api/auth'
import { useAuthStore } from '../store/auth'
import AuthForm, { type AuthFormFields } from '../components/AuthForm'

export default function RegisterPage() {
  const navigate = useNavigate()
  const loginStore = useAuthStore((s) => s.login)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit({ email, password, username }: AuthFormFields) {
    setError('')
    setLoading(true)
    try {
      const data = await register(username ?? '', email, password)
      loginStore({ id: data.user.id, username: data.user.username, role: data.user.role })
      navigate('/')
    } catch (err: unknown) {
      const apiMsg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      setError(apiMsg ?? 'Erreur lors de la création du compte.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <AuthForm
      subtitle="Créer un compte"
      showUsername
      submitLabel="S'inscrire"
      loading={loading}
      error={error}
      onSubmit={handleSubmit}
      footer={
        <p>
          Déjà inscrit ?{' '}
          <Link to="/user-login" className="auth-link">
            Se connecter
          </Link>
        </p>
      }
    />
  )
}
