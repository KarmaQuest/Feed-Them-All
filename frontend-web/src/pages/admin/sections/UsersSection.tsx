// src/pages/admin/sections/UsersSection.tsx — Gestion des utilisateurs.
//
// Liste paginée + filtre search. Inline edit pour role et is_banned.
import { useState, useEffect, useCallback } from 'react'
import { listUsers, updateUser, type AdminUser } from '../../../api/admin'

export default function UsersSection() {
  const [users, setUsers] = useState<AdminUser[]>([])
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState<string | null>(null) // user id being saved

  const fetchUsers = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listUsers(page, search)
      setUsers(data)
    } finally {
      setLoading(false)
    }
  }, [page, search])

  useEffect(() => { fetchUsers() }, [fetchUsers])

  async function handleRoleChange(user: AdminUser, role: string) {
    setSaving(user.id)
    try {
      await updateUser(user.id, { role })
      setUsers((u) => u.map((x) => (x.id === user.id ? { ...x, role } : x)))
    } finally {
      setSaving(null)
    }
  }

  async function toggleBan(user: AdminUser) {
    setSaving(user.id)
    try {
      await updateUser(user.id, { is_banned: !user.is_banned })
      setUsers((u) =>
        u.map((x) => (x.id === user.id ? { ...x, is_banned: !x.is_banned } : x)),
      )
    } finally {
      setSaving(null)
    }
  }

  return (
    <div>
      <h2 className="admin-section-title">Utilisateurs</h2>
      <div className="section-toolbar">
        <input
          className="search-input"
          type="search"
          placeholder="Rechercher…"
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1) }}
        />
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            className="btn-cancel"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
          >
            ← Préc.
          </button>
          <span className="text-muted" style={{ lineHeight: '2.1' }}>
            Page {page}
          </span>
          <button
            className="btn-cancel"
            disabled={users.length < 20}
            onClick={() => setPage((p) => p + 1)}
          >
            Suiv. →
          </button>
        </div>
      </div>

      <div className="admin-table-wrapper">
        <table className="admin-table">
          <thead>
            <tr>
              <th>Utilisateur</th>
              <th>Email</th>
              <th>Rôle</th>
              <th>XP</th>
              <th>Statut</th>
              <th>Créé le</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr className="loading-row"><td colSpan={7}>Chargement…</td></tr>
            ) : users.length === 0 ? (
              <tr className="loading-row"><td colSpan={7}>Aucun utilisateur trouvé</td></tr>
            ) : users.map((u) => (
              <tr key={u.id}>
                <td>
                  <span style={{ fontWeight: 600, color: '#e8eaf6' }}>{u.username}</span>
                  <br />
                  <span className="text-mono text-muted">{u.id.slice(0, 8)}…</span>
                </td>
                <td className="text-muted">{u.email}</td>
                <td>
                  <select
                    className="inline-select"
                    value={u.role}
                    disabled={saving === u.id}
                    onChange={(e) => handleRoleChange(u, e.target.value)}
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
                    <span className="text-red">🚫 Banni</span>
                  ) : (
                    <span className="text-green">✓ Actif</span>
                  )}
                </td>
                <td className="text-muted">{u.created_at.slice(0, 10)}</td>
                <td>
                  <button
                    className={u.is_banned ? 'btn-save' : 'btn-danger'}
                    disabled={saving === u.id}
                    onClick={() => toggleBan(u)}
                  >
                    {saving === u.id ? '…' : u.is_banned ? 'Débannir' : 'Bannir'}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
