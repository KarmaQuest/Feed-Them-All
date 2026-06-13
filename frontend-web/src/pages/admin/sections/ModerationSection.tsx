// src/pages/admin/sections/ModerationSection.tsx — Moderation des pings.
//
// Tableau des pings avec filtres (actifs / signales).
// Bouton "Desactiver" force is_active=false sans verifier le proprietaire.
// Bouton "Creer ping" pour injecter un ping manuellement (admin/test).
//
// Filtres :
//   Actifs seulement   : pings dont is_active=true (en cours)
//   Signales seulement : pings dont report_count > 0
import { useState, useEffect, useCallback } from 'react'
import { listPingsAdmin, forceDeactivatePing, createPingAdmin, type AdminPing } from '../../../api/admin'
import { useAuthStore } from '../../../store/auth'

export default function ModerationSection() {
  const currentUser = useAuthStore((s) => s.user)
  const [pings, setPings] = useState<AdminPing[]>([])
  const [loading, setLoading] = useState(false)
  const [activeOnly, setActiveOnly] = useState(true)
  const [flaggedOnly, setFlaggedOnly] = useState(false)
  const [deactivating, setDeactivating] = useState<string | null>(null)
  const [modal, setModal] = useState(false)
  const [newPing, setNewPing] = useState({ user_id: '', type: 'animal', lat: '', lon: '' })
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')

  const fetchPings = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listPingsAdmin({ active: activeOnly || undefined, flagged: flaggedOnly || undefined })
      setPings(data)
    } finally {
      setLoading(false)
    }
  }, [activeOnly, flaggedOnly])

  useEffect(() => { fetchPings() }, [fetchPings])

  async function handleDeactivate(id: string) {
    setDeactivating(id)
    try {
      await forceDeactivatePing(id)
      setPings((prev) => prev.map((p) => p.id === id ? { ...p, is_active: false } : p))
    } finally {
      setDeactivating(null)
    }
  }

  async function handleCreate() {
    setCreateError('')
    const lat = parseFloat(newPing.lat)
    const lon = parseFloat(newPing.lon)
    if (!newPing.user_id) { setCreateError("L'ID utilisateur est requis."); return }
    if (isNaN(lat) || isNaN(lon)) { setCreateError('Latitude et longitude doivent etre des nombres.'); return }
    setCreating(true)
    try {
      await createPingAdmin({ user_id: newPing.user_id, type: newPing.type, lat, lon })
      setModal(false)
      setNewPing({ user_id: currentUser?.id ?? '', type: 'animal', lat: '', lon: '' })
      fetchPings()
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
      setCreateError(msg ?? 'Erreur lors de la creation.')
    } finally {
      setCreating(false)
    }
  }

  return (
    <div>
      <h2 className="admin-section-title">Moderation des Pings</h2>

      <div className="section-toolbar">
        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: '#9ca3af', fontSize: '0.875rem', cursor: 'pointer' }}>
            <input type="checkbox" checked={activeOnly} onChange={(e) => setActiveOnly(e.target.checked)} />
            Actifs seulement
          </label>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: '#9ca3af', fontSize: '0.875rem', cursor: 'pointer' }}>
            <input type="checkbox" checked={flaggedOnly} onChange={(e) => setFlaggedOnly(e.target.checked)} />
            Signales seulement
          </label>
        </div>
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <span className="text-muted">{pings.length} pings</span>
          <button className="btn-add" onClick={() => {
            setModal(true)
            setCreateError('')
            setNewPing({ user_id: currentUser?.id ?? '', type: 'animal', lat: '', lon: '' })
          }}>
            + Creer ping
          </button>
        </div>
      </div>

      <div className="admin-table-wrapper" style={{ marginTop: '0.75rem' }}>
        <table className="admin-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Type</th>
              <th>Cree par</th>
              <th>Reports</th>
              <th>Statut</th>
              <th>Cree le</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr className="loading-row"><td colSpan={7}>Chargement...</td></tr>
            ) : pings.length === 0 ? (
              <tr className="loading-row"><td colSpan={7}>Aucun ping trouve</td></tr>
            ) : pings.map((p) => (
              <tr key={p.id}>
                <td className="text-mono text-muted">{p.id.slice(0, 8)}...</td>
                <td>
                  <span className="badge-role" style={{
                    background: p.type === 'animal' ? 'rgba(16,185,129,0.14)' : 'rgba(245,158,11,0.14)',
                    color: p.type === 'animal' ? '#34d399' : '#fbbf24',
                  }}>
                    {p.type}
                  </span>
                </td>
                <td className="text-mono text-muted">{p.created_by.slice(0, 8)}...</td>
                <td>
                  {p.report_count > 0 ? (
                    <span className="text-red" style={{ fontWeight: 600 }}>! {p.report_count}</span>
                  ) : (
                    <span className="text-muted">0</span>
                  )}
                </td>
                <td>
                  <span className={`badge-status ${p.is_active ? 'active' : 'inactive'}`} />
                  {p.is_active ? <span className="text-green">Actif</span> : <span className="text-muted">Inactif</span>}
                </td>
                <td className="text-muted">{p.created_at.slice(0, 10)}</td>
                <td>
                  {p.is_active ? (
                    <button className="btn-danger" disabled={deactivating === p.id} onClick={() => handleDeactivate(p.id)}>
                      {deactivating === p.id ? '...' : 'Desactiver'}
                    </button>
                  ) : (
                    <span className="text-muted">---</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {modal && (
        <div className="modal-overlay" onClick={() => setModal(false)}>
          <div className="modal-box" onClick={(e) => e.stopPropagation()}>
            <h3>Creer un ping</h3>
            <div className="modal-form-grid">
              <label>ID utilisateur</label>
              <input className="inline-input" placeholder="UUID de l'utilisateur"
                value={newPing.user_id} onChange={(e) => setNewPing({ ...newPing, user_id: e.target.value })} />
              <label>Type</label>
              <select className="inline-select" value={newPing.type}
                onChange={(e) => setNewPing({ ...newPing, type: e.target.value })}>
                <option value="animal">animal</option>
                <option value="food">food</option>
              </select>
              <label>Latitude</label>
              <input className="inline-input" type="number" step="any" placeholder="ex: 21.027763"
                value={newPing.lat} onChange={(e) => setNewPing({ ...newPing, lat: e.target.value })} />
              <label>Longitude</label>
              <input className="inline-input" type="number" step="any" placeholder="ex: 105.834160"
                value={newPing.lon} onChange={(e) => setNewPing({ ...newPing, lon: e.target.value })} />
            </div>
            {createError && <p className="error-msg">{createError}</p>}
            <div className="modal-actions">
              <button className="btn-cancel" onClick={() => setModal(false)}>Annuler</button>
              <button className="btn-submit" disabled={creating} onClick={handleCreate}>
                {creating ? 'Creation...' : 'Creer'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
