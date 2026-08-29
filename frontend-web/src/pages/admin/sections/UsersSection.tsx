// src/pages/admin/sections/UsersSection.tsx — Gestion des utilisateurs.
//
// Liste paginee + filtre search. Inline edit pour role et is_banned.
// Modale pour creer un nouvel utilisateur.
import { useState, useEffect, useCallback } from 'react'
import { listUsers, updateUser, createUser, deleteUser, type AdminUser } from '../../../api/admin'
import { useAuthStore } from '../../../store/auth'

const EMPTY_NEW_USER = { email: '', username: '', password: '', role: 'feeder' }

export default function UsersSection() {
  const currentUser = useAuthStore(s => s.user)
  const [users, setUsers] = useState<AdminUser[]>([])
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState<string | null>(null)
  const [deleting, setDeleting] = useState<string | null>(null)
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)
  const [modal, setModal] = useState(false)
  const [newUser, setNewUser] = useState({ ...EMPTY_NEW_USER })
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')

  const fetchUsers = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listUsers(page, search)
      setUsers(data)
    } finally {
      setLoading(false)
    }
  }, [page, search])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  async function handleRoleChange(user: AdminUser, role: string) {
    setSaving(user.id)
    try {
      await updateUser(user.id, { role })
      setUsers(u => u.map(x => (x.id === user.id ? { ...x, role } : x)))
    } finally {
      setSaving(null)
    }
  }

  async function toggleBan(user: AdminUser) {
    setSaving(user.id)
    try {
      await updateUser(user.id, { is_banned: !user.is_banned })
      setUsers(u => u.map(x => (x.id === user.id ? { ...x, is_banned: !x.is_banned } : x)))
    } finally {
      setSaving(null)
    }
  }

  async function handleDelete(id: string) {
    setDeleting(id)
    try {
      await deleteUser(id)
      setUsers(u => u.filter(x => x.id !== id))
      setConfirmDelete(null)
    } finally {
      setDeleting(null)
    }
  }

  async function handleCreate() {
    setCreateError('')
    if (!newUser.email || !newUser.username || !newUser.password) {
      setCreateError('Tous les champs sont requis.')
      return
    }
    setCreating(true)
    try {
      await createUser(newUser)
      setModal(false)
      setNewUser({ ...EMPTY_NEW_USER })
      fetchUsers()
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
      setCreateError(msg ?? 'Erreur lors de la creation.')
    } finally {
      setCreating(false)
    }
  }

  return (
    <div>
      <h2 className="admin-section-title">Utilisateurs</h2>
      <div className="section-toolbar">
        <input
          className="search-input"
          type="search"
          placeholder="Rechercher..."
          value={search}
          onChange={e => {
            setSearch(e.target.value)
            setPage(1)
          }}
        />
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            className="btn btn--style-yellow"
            onClick={() => {
              setModal(true)
              setCreateError('')
            }}
          >
            + Nouvel utilisateur
          </button>
          <button className="btn-page" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
            &larr; Préc.
          </button>
          <span className="text-muted pagination" style={{ lineHeight: '1.6', fontSize: '0.8rem' }}>
            Page {page}
          </span>
          <button
            className="btn-page"
            disabled={users.length < 20}
            onClick={() => setPage(p => p + 1)}
          >
            Suiv. &rarr;
          </button>
        </div>
      </div>

      <div className="admin-table-wrapper">
        <table className="admin-table">
          <thead>
            <tr>
              <th>Utilisateur</th>
              <th>Email</th>
              <th>Role</th>
              <th>XP</th>
              <th>Statut</th>
              <th>Cree le</th>
              <th>Bannir</th>
              <th>Supprimer</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr className="loading-row">
                <td colSpan={7}>Chargement...</td>
              </tr>
            ) : users.length === 0 ? (
              <tr className="loading-row">
                <td colSpan={7}>Aucun utilisateur trouve</td>
              </tr>
            ) : (
              users.map(u => (
                <tr key={u.id}>
                  <td>
                    <span style={{ fontWeight: 600, color: '#e8eaf6' }}>{u.username}</span>
                    <br />
                    <span className="text-mono text-muted">{u.id.slice(0, 8)}...</span>
                  </td>
                  <td className="text-muted">{u.email}</td>
                  <td>
                    <select
                      className="inline-select"
                      value={u.role}
                      disabled={saving === u.id}
                      onChange={e => handleRoleChange(u, e.target.value)}
                    >
                      <option value="feeder">feeder</option>
                      <option value="giver">giver</option>
                      <option value="association">association</option>
                      <option value="admin">admin</option>
                    </select>
                  </td>
                  <td>{u.xp.toLocaleString()}</td>
                  <td>
                    {u.is_banned ? (
                      <span className="text-red">Banni</span>
                    ) : (
                      <span className="text-green">Actif</span>
                    )}
                  </td>
                  <td className="text-muted">{u.created_at.slice(0, 10)}</td>
                  <td>
                    <button
                      className={u.is_banned ? 'btn btn--style-yellow btn--sm' : 'btn btn--style-red btn--sm'}
                      disabled={saving === u.id}
                      onClick={() => toggleBan(u)}
                    >
                      {saving === u.id ? '...' : u.is_banned ? 'Debannir' : 'Bannir'}
                    </button>
                  </td>
                  <td>
                    {currentUser?.id === u.id ? (
                      <span className="text-muted">—</span>
                    ) : confirmDelete === u.id ? (
                      <span style={{ display: 'flex', gap: '0.3rem' }}>
                        <button
                          className="btn btn--style-red btn--sm"
                          disabled={deleting === u.id}
                          onClick={() => handleDelete(u.id)}
                        >
                          {deleting === u.id ? '...' : 'Confirmer'}
                        </button>
                        <button className="btn-cancel" onClick={() => setConfirmDelete(null)}>
                          ✕
                        </button>
                      </span>
                    ) : (
                      <button className="btn btn--style-red btn--sm" onClick={() => setConfirmDelete(u.id)}>
                        Supprimer
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {modal && (
        <div className="modal-overlay" onClick={() => setModal(false)}>
          <div className="modal-box" onClick={e => e.stopPropagation()}>
            <h3>Nouvel utilisateur</h3>
            <div className="modal-form-grid">
              <label>Email</label>
              <input
                className="inline-input"
                type="email"
                value={newUser.email}
                onChange={e => setNewUser({ ...newUser, email: e.target.value })}
              />
              <label>Nom d'utilisateur</label>
              <input
                className="inline-input"
                value={newUser.username}
                onChange={e => setNewUser({ ...newUser, username: e.target.value })}
              />
              <label>Mot de passe</label>
              <input
                className="inline-input"
                type="password"
                value={newUser.password}
                onChange={e => setNewUser({ ...newUser, password: e.target.value })}
              />
              <label>Role</label>
              <select
                className="inline-select"
                value={newUser.role}
                onChange={e => setNewUser({ ...newUser, role: e.target.value })}
              >
                <option value="feeder">feeder</option>
                <option value="giver">giver</option>
                <option value="association">association</option>
                <option value="admin">admin</option>
              </select>
            </div>
            {createError && <p className="error-msg">{createError}</p>}
            <div className="modal-actions">
              <button className="btn btn--style-red" onClick={() => setModal(false)}>
                Annuler
              </button>
              <button className="btn btn--style-green" disabled={creating} onClick={handleCreate}>
                {creating ? 'Creation...' : 'Creer'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
