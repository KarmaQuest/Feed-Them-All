// src/pages/admin/sections/BadgesSection.tsx — CRUD des badges.
//
// Editeur de condition structure : xp_threshold ou action_count.
// Les deux types sont supportes par le moteur gamification backend.
//   xp_threshold  : badge debloque quand users.xp >= value
//   action_count  : badge debloque quand COUNT(xp_log.action = X) >= value
// Pour des badges cumulables (niveau 1, 2, 3 pour la meme action),
// creer plusieurs badges avec la meme action mais des valeurs differentes.
import { useState, useEffect } from 'react'
import {
  listBadges,
  createBadge,
  updateBadge,
  deleteBadge,
  type AdminBadge,
} from '../../../api/admin'

interface ConditionForm {
  type: 'xp_threshold' | 'action_count'
  value: string
  action: string
}

const EMPTY_COND: ConditionForm = { type: 'xp_threshold', value: '100', action: '' }
const EMPTY_BADGE_FIELDS = { slug: '', label: '', description: '' }

function condFormToJSON(c: ConditionForm): object {
  if (c.type === 'xp_threshold') return { type: 'xp_threshold', value: parseInt(c.value, 10) || 0 }
  return { type: 'action_count', action: c.action, value: parseInt(c.value, 10) || 0 }
}

function condJSONToForm(raw: unknown): ConditionForm {
  if (!raw || typeof raw !== 'object') return { ...EMPTY_COND }
  const c = raw as { type?: string; value?: number; action?: string }
  if (c.type === 'action_count') {
    return { type: 'action_count', value: String(c.value ?? 1), action: c.action ?? '' }
  }
  return { type: 'xp_threshold', value: String(c.value ?? 100), action: '' }
}

function condLabel(raw: unknown): string {
  if (!raw || typeof raw !== 'object') return '?'
  const c = raw as { type?: string; value?: number; action?: string }
  if (c.type === 'xp_threshold') return String(c.value) + ' XP'
  if (c.type === 'action_count') return String(c.value) + 'x ' + c.action
  return JSON.stringify(raw)
}

export default function BadgesSection() {
  const [badges, setBadges] = useState<AdminBadge[]>([])
  const [loading, setLoading] = useState(false)
  const [modal, setModal] = useState<null | { mode: 'create' | 'edit'; data: typeof EMPTY_BADGE_FIELDS; cond: ConditionForm; id?: string }>(null)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    listBadges().then((d) => { setBadges(d); setLoading(false) })
  }, [])

  function openCreate() {
    setModal({ mode: 'create', data: { ...EMPTY_BADGE_FIELDS }, cond: { ...EMPTY_COND } })
  }

  function openEdit(b: AdminBadge) {
    setModal({
      mode: 'edit',
      data: { slug: b.slug, label: b.label, description: b.description },
      cond: condJSONToForm(b.condition),
      id: b.id,
    })
  }

  function closeModal() { setModal(null) }

  async function handleSave() {
    if (!modal) return
    const payload = {
      ...modal.data,
      condition: condFormToJSON(modal.cond),
    }
    setSaving(true)
    try {
      if (modal.mode === 'create') {
        const { id } = await createBadge(payload as Omit<AdminBadge, 'id'>)
        setBadges((prev) => [...prev, { id, ...payload } as AdminBadge])
      } else if (modal.id) {
        await updateBadge(modal.id, payload as Omit<AdminBadge, 'id'>)
        setBadges((prev) => prev.map((b) => b.id === modal.id ? { id: modal.id!, ...payload } as AdminBadge : b))
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

      <button className="btn-add" onClick={openCreate}>+ Nouveau badge</button>

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
              <tr className="loading-row"><td colSpan={5}>Chargement...</td></tr>
            ) : badges.length === 0 ? (
              <tr className="loading-row"><td colSpan={5}>Aucun badge</td></tr>
            ) : badges.map((b) => (
              <tr key={b.id}>
                <td><code style={{ color: '#818cf8' }}>{b.slug}</code></td>
                <td style={{ fontWeight: 600 }}>{b.label}</td>
                <td className="text-muted">{b.description || '---'}</td>
                <td>
                  <span style={{ background: '#1e2340', padding: '0.2rem 0.5rem', borderRadius: '4px', fontSize: '0.8rem', color: '#c4c9e2' }}>
                    {condLabel(b.condition)}
                  </span>
                </td>
                <td style={{ display: 'flex', gap: '0.5rem' }}>
                  <button className="btn-save" onClick={() => openEdit(b)}>Modifier</button>
                  <button className="btn-danger" disabled={deleting === b.id} onClick={() => handleDelete(b.id)}>
                    {deleting === b.id ? '...' : 'Supprimer'}
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

            <div className="modal-form-grid">
              <label>Slug</label>
              <input className="inline-input" placeholder="ex: feeder_10"
                value={modal.data.slug}
                onChange={(e) => setModal({ ...modal, data: { ...modal.data, slug: e.target.value } })} />

              <label>Label</label>
              <input className="inline-input" placeholder="ex: Nourrisseur confirme"
                value={modal.data.label}
                onChange={(e) => setModal({ ...modal, data: { ...modal.data, label: e.target.value } })} />

              <label>Description</label>
              <input className="inline-input"
                value={modal.data.description}
                onChange={(e) => setModal({ ...modal, data: { ...modal.data, description: e.target.value } })} />

              <label>Type de condition</label>
              <select className="inline-select" value={modal.cond.type}
                onChange={(e) => setModal({ ...modal, cond: { ...modal.cond, type: e.target.value as 'xp_threshold' | 'action_count' } })}>
                <option value="xp_threshold">XP minimum (xp_threshold)</option>
                <option value="action_count">Nombre d'actions (action_count)</option>
              </select>

              {modal.cond.type === 'action_count' && (
                <>
                  <label>Nom de l'action</label>
                  <input className="inline-input" placeholder="ex: feed, report_animal"
                    value={modal.cond.action}
                    onChange={(e) => setModal({ ...modal, cond: { ...modal.cond, action: e.target.value } })} />
                </>
              )}

              <label>{modal.cond.type === 'xp_threshold' ? 'XP requis' : 'Nombre requis'}</label>
              <input className="inline-input" type="number" min={1}
                value={modal.cond.value}
                onChange={(e) => setModal({ ...modal, cond: { ...modal.cond, value: e.target.value } })} />
            </div>

            <div style={{ marginTop: '0.75rem', padding: '0.5rem 0.75rem', background: '#1a1f2e', borderRadius: '6px', fontSize: '0.8rem', color: '#6b7280' }}>
              JSON : <code style={{ color: '#818cf8' }}>{JSON.stringify(condFormToJSON(modal.cond))}</code>
              {modal.cond.type === 'action_count' && (
                <p style={{ marginTop: '0.4rem', color: '#6b7280', margin: '0.3rem 0 0' }}>
                  Badges cumulables : creer plusieurs badges avec la meme action et des valeurs differentes (ex: feed x10, feed x50, feed x100).
                </p>
              )}
            </div>

            <div className="modal-actions">
              <button className="btn-cancel" onClick={closeModal}>Annuler</button>
              <button className="btn-submit" disabled={saving} onClick={handleSave}>
                {saving ? '...' : modal.mode === 'create' ? 'Creer' : 'Sauvegarder'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
