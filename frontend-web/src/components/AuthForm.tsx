// src/components/AuthForm.tsx — Formulaire d'authentification partagé (login + register).
//
// Rôle : fournir le rendu visuel commun à toutes les pages d'auth (login admin,
//         login utilisateur, inscription). La logique métier reste dans chaque page.
//
// Props :
//   subtitle     — texte affiché sous le logo (ex : "Connexion", "Créer un compte")
//   showUsername — affiche le champ username (mode register uniquement)
//   submitLabel  — texte du bouton de soumission
//   loading      — désactive le bouton pendant la requête
//   error        — message d'erreur à afficher sous le bouton
//   onSubmit     — callback avec { email, password, username? }
//   footer       — contenu optionnel sous le formulaire (liens de navigation)
import { useState, type FormEvent } from 'react'
import '../pages/LoginPage.css'

export interface AuthFormFields {
  email: string
  password: string
  username?: string
}

interface Props {
  subtitle: string
  showUsername?: boolean
  submitLabel: string
  loading: boolean
  error: string
  onSubmit: (fields: AuthFormFields) => void
  footer?: React.ReactNode
  extraFields?: React.ReactNode
}

export default function AuthForm({
  subtitle,
  showUsername,
  submitLabel,
  loading,
  error,
  onSubmit,
  footer,
  extraFields,
}: Props) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [username, setUsername] = useState('')

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    // Lire depuis le DOM pour capter l'autofill navigateur (sinon React state = vide)
    const form = e.currentTarget
    const emailVal = (form.elements.namedItem('email') as HTMLInputElement).value
    const passwordVal = (form.elements.namedItem('password') as HTMLInputElement).value
    const usernameVal = showUsername
      ? (form.elements.namedItem('username') as HTMLInputElement).value
      : undefined
    onSubmit({
      email: emailVal,
      password: passwordVal,
      ...(showUsername ? { username: usernameVal } : {}),
    })
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <img src="/logo.png" alt="FeedThemAll" className="login-logo" />
        <p>{subtitle}</p>
        <form onSubmit={handleSubmit}>
          {showUsername && (
            <div className="form-group">
              <label htmlFor="username">Nom d'utilisateur</label>
              <input
                id="username"
                name="username"
                type="text"
                value={username}
                onChange={e => setUsername(e.target.value)}
                autoComplete="username"
                required
              />
            </div>
          )}
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              name="email"
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              autoComplete="email"
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="password">Mot de passe</label>
            <input
              id="password"
              name="password"
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              autoComplete={showUsername ? 'new-password' : 'current-password'}
              required
            />
          </div>
          {extraFields}
          <button type="submit" className="btn btn--style-yellow btn--full" disabled={loading}>
            {loading ? '…' : submitLabel}
          </button>
          {error && <div className="login-error">{error}</div>}
        </form>
        {footer && <div className="auth-footer">{footer}</div>}
      </div>
    </div>
  )
}
