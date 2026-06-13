// src/pages/admin/sections/ModerationSection.tsx — Modération des pings.
//
// Tableau des pings avec filtres (actifs / signalés).
// Bouton "Désactiver" force is_active=false sans vérifier le propriétaire.
import { useState, useEffect, useCallback } from 'react'
import { listPingsAdmin, forceDeactivatePing, type AdminPing } from '../../../api/admin'

export default function ModerationSection() {
  const [pings, setPings] = useState<AdminPing[]>([])
  const [loading, setLoading] = useState(false)
  const [activeOnly, setActiveOnly] = useState(true)
  const [flaggedOnly, setFlaggedOnly] = useState(false)
  const [deactivating, setDeactivating] = useState<string | null>(null)

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

  return (
    <div>
      <h2 className="admin-section-title">Modération des Pings</h2>

      <div className="section-toolbar">
        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: '#9ca3af', fontSize: '0.875rem', cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={activeOnly}
              onChange={(e) => setActiveOnly(e.target.checked)}
            />
            Actifs seulement
          </label>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: '#9ca3af', fontSize: '0.875rem', cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={flaggedOnly}
              onChange={(e) => setFlaggedOnly(e.target.checked)}
            />
            Signalés seulement
          </label>
        </div>
        <span className="text-muted">{pings.length} pings</span>
      </div>

      <div className="admin-table-wrapper" style={{ marginTop: '0.75rem' }}>
        <table className="admin-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Type</th>
              <th>Créé par</th>
              <th>Reports</th>
              <th>Statut</th>
              <th>Créé le</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr className="loading-row"><td colSpan={7}>Chargement…</td></tr>
            ) : pings.length === 0 ? (
              <tr className="loading-row"><td colSpan={7}>Aucun ping trouvé</td></tr>
            ) : pings.map((p) => (
              <tr key={p.id}>
                <td className="text-mono text-muted">{p.id.slice(0, 8)}…</td>
                <td>
                  <span className="badge-role" style={{
                    background: p.type === 'animal' ? 'rgba(16,185,129,0.14)' : 'rgba(245,158,11,0.14)',
                    color: p.type === 'animal' ? '#34d399' : '#fbbf24',
                  }}>
                    {p.type}
                  </span>
                </td>
                <td className="text-mono text-muted">{p.created_by.slice(0, 8)}…</td>
                <td>
                  {p.report_count > 0 ? (
                    <span className="text-red" style={{ fontWeight: 600 }}>⚠ {p.report_count}</span>
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
                    <button
                      className="btn-danger"
                      disabled={deactivating === p.id}
                      onClick={() => handleDeactivate(p.id)}
                    >
                      {deactivating === p.id ? '…' : 'Désactiver'}
                    </button>
                  ) : (
                    <span className="text-muted">—</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
