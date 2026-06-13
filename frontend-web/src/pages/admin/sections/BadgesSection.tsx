// src/pages/admin/sections/BadgesSection.tsx — CRUD des badges.
//
// Tableau avec tous les badges. Bouton "Nouveau badge" ouvre une modale de création.
// Bouton "Modifier" ouvre la même modale en mode édition.
// Bouton "Supprimer" avec confirmation inline.
import { useState, useEffect } from 'react'
import {
  listBadges,
  createBadge,
  updateBadge,
  deleteBadge,
  type AdminBadge,
} from '../../../api/admin'

const EMPTY_BADGE: Omit<AdminBadge, 'id'> = {
  slug: '',
  label: '',
  description: '',
  condition: { type: 'xp_threshold', value: 100 },
}

export default function BadgesSection() {
  const [badges, setBadges] = useState<AdminBadge[]>([])
  const [loading, setLoading] = useState(false)
  const [modal, setModal] = useState<null | { mode: 'create' | 'edit'; data: Omit<AdminBadge, 'id'>; id?: string }>(null)
  const [conditionRaw, setConditionRaw] = useState('')
  const [condError, setCondError] = useState('')
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    listBadges().then((d) => { setBadges(d); setLoading(false) })
  }, [])

  function openCreate() {
    setModal({ mode: 'create', data: { ...EMPTY_BADGE } })
    setConditionRaw(JSON.stringify(EMPTY_BADGE.condition, null, 2))
    setCondError('')
  }

  function openEdit(b: AdminBadge) {
    setModal({ mode: 'edit', data: { slug: b.slug, label: b.label, description: b.description, condition: b.condition }, id: b.id })
    setConditionRaw(JSON.stringify(b.condition, null, 2))
    setCondError('')
  }

  function closeModal() { setModal(null); setCondError('') }

  function updateField(field: keyof Omit<AdminBadge, 'id'>, value: unknown) {
    if (!modal) return
    setModal({ ...modal, data: { ...modal.data, [field]: value } })
  }

  async function handleSave() {
    if (!modal) return
    let condition: unknown
    try {
      condition = JSON.parse(conditionRaw)
      setCondError('')
    } catch {
      setCondError('JSON invalide')
      return
    }
    const payload = { ...modal.data, condition }
    setSaving(true)
    try {
      if (modal.mode === 'create') {
        const { id } = await createBadge(payload as Omit<AdminBadge, 'id'>)
        setBadges((prev) => [...prev, { id, ...payload } as AdminBadge])
      } else if (modal.id) {
        await updateBadge(modal.id, payload as Omit<AdminBadge, 'id'>)
        setBadges((prev) => prev.map((b) => b.id === modal.id ? { id: modal.id, ...payload } as AdminBadge : b))
      }
      closeModal()
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id: string) {
    setDeleting(id)
    try {
      await deleteBadge(id)
      setBadges((prev) => prev.filter((b) => b.id !== id))
    } finally {
      setDeleting(null)
    }
  }

  return (
    <div>
      <h2 className="admin-section-title">Badges</h2>

      <button className="btn-add" onClick={openCreate}>＋ Nouveau badge</button>

      <div className="admin-table-wrapper">
        <table className="admin-table">
          <thead>
            <tr>
              <th>Slug</th>
              <th>Label</th>
              <th>Description</th>
              <th>Condition</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr className="loading-row"><td colSpan={5}>Chargement…</td></tr>
            ) : badges.length === 0 ? (
              <tr className="loading-row"><td colSpan={5}>Aucun badge</td></tr>
            ) : badges.map((b) => (
              <tr key={b.id}>
                <td><code style={{ color: '#818cf8' }}>{b.slug}</code></td>
                <td style={{ fontWeight: 600 }}>{b.label}</td>
                <td className="text-muted">{b.description || '—'}</td>
                <td><code className="text-mono text-muted">{JSON.stringify(b.condition)}</code></td>
                <td style={{ display: 'flex', gap: '0.5rem' }}>
                  <button className="btn-save" onClick={() => openEdit(b)}>Modifier</button>
                  <button
                    className="btn-danger"
                    disabled={deleting === b.id}
                    onClick={() => handleDelete(b.id)}
                  >
                    {deleting === b.id ? '…' : 'Supprimer'}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {modal && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-box" onClick={(e) => e.stopPropagation()}>
            <h3>{modal.mode === 'create' ? 'Nouveau badge' : 'Modifier le badge'}</h3>

            <div className="modal-field">
              <label>Slug</label>
              <input
                type="text"
                value={modal.data.slug}
                onChange={(e) => updateField('slug', e.target.value)}
                placeholder="ex: feeder_10"
              />
            </div>
            <div className="modal-field">
              <label>Label</label>
              <input
                type="text"
                value={modal.data.label}
                onChange={(e) => updateField('label', e.target.value)}
                placeholder="ex: Nourrisseur confirmé"
              />
            </div>
            <div className="modal-field">
              <label>Description</label>
              <input
                type="text"
                value={modal.data.description}
                onChange={(e) => updateField('description', e.target.value)}
              />
            </div>
            <div className="modal-field">
              <label>Condition (JSON)</label>
              <textarea
                value={conditionRaw}
                onChange={(e) => setConditionRaw(e.target.value)}
                rows={4}
              />
              {condError && <span className="text-red" style={{ fontSize: '0.8rem' }}>{condError}</span>}
              <div className="text-muted" style={{ marginTop: '0.3rem' }}>
                Exemples :<br />
                <code>{`{"type":"xp_threshold","value":500}`}</code><br />
                <code>{`{"type":"action_count","action":"feed","value":10}`}</code>
              </div>
            </div>

            <div className="modal-actions">
              <button className="btn-cancel" onClick={closeModal}>Annuler</button>
              <button className="btn-submit" disabled={saving} onClick={handleSave}>
                {saving ? '…' : modal.mode === 'create' ? 'Créer' : 'Sauvegarder'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
