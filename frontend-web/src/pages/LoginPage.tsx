// src/pages/LoginPage.tsx — Page de connexion admin.
//
// Seul point d'entrée authentifié : email + password → JWT → store Zustand.
// Redirige vers /admin après connexion réussie.
// Affiche une erreur si les credentials sont invalides ou si le rôle n'est pas admin.
import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { login } from '../api/auth'
import { useAuthStore } from '../store/auth'
import './LoginPage.css'

export default function LoginPage() {
  const navigate = useNavigate()
  const loginStore = useAuthStore((s) => s.login)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = await login(email, password)
      if (data.role !== 'admin') {
        setError('Accès refusé — compte admin requis.')
        setLoading(false)
        return
      }
      loginStore({ id: data.user_id, username: data.username, role: data.role })
      navigate('/admin')
    } catch {
      setError('Email ou mot de passe incorrect.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <h1>FeedThemAll</h1>
        <p>Dashboard Admin</p>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="email"
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="password">Mot de passe</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="current-password"
              required
            />
          </div>
          <button type="submit" className="btn-primary" disabled={loading}>
            {loading ? 'Connexion…' : 'Se connecter'}
          </button>
          {error && <div className="login-error">{error}</div>}
        </form>
      </div>
    </div>
  )
}
