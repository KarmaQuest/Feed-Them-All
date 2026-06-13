// src/pages/admin/sections/XPSection.tsx — Dashboard XP et Levels.
//
// Deux tableaux : xp_actions (inline edit + creation) + level_thresholds (inline edit + ajout).
// Le changement des paliers de level recharge la config cote serveur (users.Service).
import { useState, useEffect } from 'react'
import {
  listXPActions,
  updateXPAction,
  createXPAction,
  listThresholds,
  replaceThresholds,
  type AdminXPAction,
  type LevelThreshold,
} from '../../../api/admin'

const EMPTY_NEW_ACTION = { action: '', xp_value: '10', daily_limit: '5' }

export default function XPSection() {
  const [actions, setActions] = useState<AdminXPAction[]>([])
  const [editActions, setEditActions] = useState<Record<string, { xp_value: string; daily_limit: string }>>({})
  const [editThresholds, setEditThresholds] = useState<LevelThreshold[]>([])
  const [savingAction, setSavingAction] = useState<string | null>(null)
  const [savingThresholds, setSavingThresholds] = useState(false)
  const [savedAction, setSavedAction] = useState<string | null>(null)
  const [savedThresholds, setSavedThresholds] = useState(false)
  const [newAction, setNewAction] = useState({ ...EMPTY_NEW_ACTION })
  const [creatingAction, setCreatingAction] = useState(false)
  const [createActionError, setCreateActionError] = useState('')

  useEffect(() => {
    listXPActions().then((data) => {
      setActions(data)
      const init: Record<string, { xp_value: string; daily_limit: string }> = {}
      data.forEach((a) => { init[a.action] = { xp_value: String(a.xp_value), daily_limit: String(a.daily_limit) } })
      setEditActions(init)
    })
    listThresholds().then((data) => { setEditThresholds(data.map((t) => ({ ...t }))) })
  }, [])

  async function saveAction(action: string) {
    const vals = editActions[action]
    if (!vals) return
    setSavingAction(action)
    try {
      await updateXPAction(action, {
        xp_value: parseInt(vals.xp_value, 10),
        daily_limit: parseInt(vals.daily_limit, 10),
      })
      setActions((prev) => prev.map((a) => a.action === action
        ? { ...a, xp_value: parseInt(vals.xp_value, 10), daily_limit: parseInt(vals.daily_limit, 10) }
        : a))
      setSavedAction(action)
      setTimeout(() => setSavedAction(null), 1500)
    } finally {
      setSavingAction(null)
    }
  }

  async function handleCreateAction() {
    setCreateActionError('')
    if (!newAction.action) { setCreateActionError("Le nom de l'action est requis."); return }
    setCreatingAction(true)
    try {
      await createXPAction({
        action: newAction.action,
        xp_value: parseInt(newAction.xp_value, 10) || 0,
        daily_limit: parseInt(newAction.daily_limit, 10) || 1,
      })
      const created: AdminXPAction = {
        action: newAction.action,
        xp_value: parseInt(newAction.xp_value, 10) || 0,
        daily_limit: parseInt(newAction.daily_limit, 10) || 1,
      }
      setActions((prev) => [...prev, created])
      setEditActions((prev) => ({ ...prev, [created.action]: { xp_value: newAction.xp_value, daily_limit: newAction.daily_limit } }))
      setNewAction({ ...EMPTY_NEW_ACTION })
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
      setCreateActionError(msg ?? 'Erreur lors de la creation.')
    } finally {
      setCreatingAction(false)
    }
  }

  async function saveThresholds() {
    setSavingThresholds(true)
    try {
      await replaceThresholds(editThresholds)
      setSavedThresholds(true)
      setTimeout(() => setSavedThresholds(false), 1500)
    } finally {
      setSavingThresholds(false)
    }
  }

  function addThresholdRow() {
    const maxLevel = editThresholds.length > 0
      ? Math.max(...editThresholds.map((t) => t.level))
      : 0
    setEditThresholds((prev) => [...prev, { level: maxLevel + 1, min_xp: 0 }])
  }

  function removeThresholdRow(i: number) {
    setEditThresholds((prev) => prev.filter((_, idx) => idx !== i))
  }

  return (
    <div>
      <h2 className="admin-section-title">XP &amp; Levels</h2>

      <h3 style={{ color: '#9ca3af', fontSize: '0.875rem', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.04em', marginBottom: '0.75rem' }}>
        Actions XP
      </h3>
      <div className="admin-table-wrapper" style={{ marginBottom: '2rem' }}>
        <table className="admin-table">
          <thead>
            <tr>
              <th>Action</th>
              <th>XP / occurrence</th>
              <th>Limite / jour</th>
              <th>Sauvegarder</th>
            </tr>
          </thead>
          <tbody>
            {actions.length === 0 ? (
              <tr className="loading-row"><td colSpan={4}>Chargement...</td></tr>
            ) : actions.map((a) => (
              <tr key={a.action}>
                <td><code style={{ color: '#818cf8' }}>{a.action}</code></td>
                <td>
                  <input className="inline-input" type="number" min={0}
                    value={editActions[a.action]?.xp_value ?? a.xp_value}
                    onChange={(e) =>
                      setEditActions((prev) => ({ ...prev, [a.action]: { ...prev[a.action], xp_value: e.target.value } }))
                    }
                  />
                </td>
                <td>
                  <input className="inline-input" type="number" min={1}
                    value={editActions[a.action]?.daily_limit ?? a.daily_limit}
                    onChange={(e) =>
                      setEditActions((prev) => ({ ...prev, [a.action]: { ...prev[a.action], daily_limit: e.target.value } }))
                    }
                  />
                </td>
                <td>
                  <button className="btn-save" disabled={savingAction === a.action} onClick={() => saveAction(a.action)}>
                    {savingAction === a.action ? '...' : savedAction === a.action ? 'Sauve' : 'Sauvegarder'}
                  </button>
                </td>
              </tr>
            ))}
            <tr style={{ background: '#1a1f2e' }}>
              <td>
                <input className="inline-input" placeholder="nom_action" value={newAction.action}
                  onChange={(e) => setNewAction({ ...newAction, action: e.target.value })} />
              </td>
              <td>
                <input className="inline-input" type="number" min={0} value={newAction.xp_value}
                  onChange={(e) => setNewAction({ ...newAction, xp_value: e.target.value })} />
              </td>
              <td>
                <input className="inline-input" type="number" min={1} value={newAction.daily_limit}
                  onChange={(e) => setNewAction({ ...newAction, daily_limit: e.target.value })} />
              </td>
              <td>
                <button className="btn-add" disabled={creatingAction} onClick={handleCreateAction}>
                  {creatingAction ? '...' : '+ Creer'}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        {createActionError && <p className="error-msg">{createActionError}</p>}
      </div>

      <div className="section-toolbar">
        <h3 style={{ color: '#9ca3af', fontSize: '0.875rem', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.04em', margin: 0 }}>
          Paliers de Level
        </h3>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button className="btn-add" onClick={addThresholdRow}>+ Ajouter un palier</button>
          <button className="btn-save" disabled={savingThresholds} onClick={saveThresholds}>
            {savingThresholds ? '...' : savedThresholds ? 'Sauvegarde' : 'Sauvegarder tout'}
          </button>
        </div>
      </div>
      <div className="admin-table-wrapper" style={{ marginTop: '0.75rem' }}>
        <table className="admin-table">
          <thead>
            <tr>
              <th>Level</th>
              <th>XP minimum</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {editThresholds.length === 0 ? (
              <tr className="loading-row"><td colSpan={3}>Chargement...</td></tr>
            ) : editThresholds.map((t, i) => (
              <tr key={i}>
                <td>
                  <input className="inline-input" type="number" min={1} value={t.level}
                    onChange={(e) => {
                      const v = parseInt(e.target.value, 10)
                      setEditThresholds((prev) => prev.map((x, idx) => idx === i ? { ...x, level: isNaN(v) ? 1 : v } : x))
                    }}
                  />
                </td>
                <td>
                  <input className="inline-input" type="number" min={0} value={t.min_xp}
                    onChange={(e) => {
                      const v = parseInt(e.target.value, 10)
                      setEditThresholds((prev) => prev.map((x, idx) => idx === i ? { ...x, min_xp: isNaN(v) ? 0 : v } : x))
                    }}
                  />
                </td>
                <td>
                  <button className="btn-danger" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem' }}
                    onClick={() => removeThresholdRow(i)}>
                    x
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
