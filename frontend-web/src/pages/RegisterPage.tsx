// src/pages/RegisterPage.tsx — Page d'inscription utilisateur.
//
// Sélection de rôle :
//   - Feeder et Giver peuvent être cochés ensemble
//   - Association est exclusif (décoche Feeder et Giver)
//   - Au moins un rôle est requis
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
  const [isFeeder, setIsFeeder] = useState(true)
  const [isGiver, setIsGiver] = useState(false)
  const [isAssociation, setIsAssociation] = useState(false)

  function toggleFeeder(checked: boolean) {
    setIsFeeder(checked)
    if (checked) setIsAssociation(false)
  }
  function toggleGiver(checked: boolean) {
    setIsGiver(checked)
    if (checked) setIsAssociation(false)
  }
  function toggleAssociation(checked: boolean) {
    setIsAssociation(checked)
    if (checked) { setIsFeeder(false); setIsGiver(false) }
  }

  async function handleSubmit({ email, password, username }: AuthFormFields) {
    const roles: string[] = []
    if (isFeeder) roles.push('feeder')
    if (isGiver) roles.push('giver')
    if (isAssociation) roles.push('association')
    if (roles.length === 0) { setError('Sélectionne au moins un rôle.'); return }

    setError('')
    setLoading(true)
    try {
      const data = await register(username ?? '', email, password, roles)
      loginStore({ id: data.user.id, username: data.user.username, role: data.user.role })
      navigate('/')
    } catch (err: unknown) {
      const apiMsg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      setError(apiMsg ?? 'Erreur lors de la création du compte.')
    } finally {
      setLoading(false)
    }
  }

  const roleSelector = (
    <div className="auth-roles">
      <p className="auth-roles__label">Ton rôle</p>
      <label className="auth-roles__option">
        <input type="checkbox" checked={isFeeder} onChange={(e) => toggleFeeder(e.target.checked)} />
        <span className="auth-roles__icon">🍽️</span>
        <span>
          <strong>Feeder</strong>
          <small>Je nourris les animaux</small>
        </span>
      </label>
      <label className="auth-roles__option">
        <input type="checkbox" checked={isGiver} onChange={(e) => toggleGiver(e.target.checked)} />
        <span className="auth-roles__icon">🎁</span>
        <span>
          <strong>Giver</strong>
          <small>Je fournis de la nourriture</small>
        </span>
      </label>
      <label className={`auth-roles__option auth-roles__option--exclusive${isAssociation ? ' auth-roles__option--active' : ''}`}>
        <input type="checkbox" checked={isAssociation} onChange={(e) => toggleAssociation(e.target.checked)} />
        <span className="auth-roles__icon">🏢</span>
        <span>
          <strong>Association</strong>
          <small>Je représente une association</small>
        </span>
      </label>
    </div>
  )

  return (
    <AuthForm
      subtitle="Créer un compte"
      showUsername
      submitLabel="S'inscrire"
      loading={loading}
      error={error}
      onSubmit={handleSubmit}
      extraFields={roleSelector}
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
