// src/pages/admin/sections/XPSection.tsx — Dashboard XP & Levels.
//
// Deux tableaux : xp_actions (inline edit) + level_thresholds (inline edit).
// Le changement des paliers de level recharge la config côté serveur (users.Service).
import { useState, useEffect } from 'react'
import {
  listXPActions,
  updateXPAction,
  listThresholds,
  replaceThresholds,
  type AdminXPAction,
  type LevelThreshold,
} from '../../../api/admin'

export default function XPSection() {
  const [actions, setActions] = useState<AdminXPAction[]>([])
  const [editActions, setEditActions] = useState<Record<string, { xp_value: string; daily_limit: string }>>({})
  const [editThresholds, setEditThresholds] = useState<LevelThreshold[]>([])
  const [savingAction, setSavingAction] = useState<string | null>(null)
  const [savingThresholds, setSavingThresholds] = useState(false)
  const [savedAction, setSavedAction] = useState<string | null>(null)
  const [savedThresholds, setSavedThresholds] = useState(false)

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

  return (
    <div>
      <h2 className="admin-section-title">XP & Levels</h2>

      {/* XP Actions */}
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
              <tr className="loading-row"><td colSpan={4}>Chargement…</td></tr>
            ) : actions.map((a) => (
              <tr key={a.action}>
                <td><code style={{ color: '#818cf8' }}>{a.action}</code></td>
                <td>
                  <input
                    className="inline-input"
                    type="number"
                    min={0}
                    value={editActions[a.action]?.xp_value ?? a.xp_value}
                    onChange={(e) =>
                      setEditActions((prev) => ({ ...prev, [a.action]: { ...prev[a.action], xp_value: e.target.value } }))
                    }
                  />
                </td>
                <td>
                  <input
                    className="inline-input"
                    type="number"
                    min={1}
                    value={editActions[a.action]?.daily_limit ?? a.daily_limit}
                    onChange={(e) =>
                      setEditActions((prev) => ({ ...prev, [a.action]: { ...prev[a.action], daily_limit: e.target.value } }))
                    }
                  />
                </td>
                <td>
                  <button
                    className="btn-save"
                    disabled={savingAction === a.action}
                    onClick={() => saveAction(a.action)}
                  >
                    {savingAction === a.action ? '…' : savedAction === a.action ? '✓ Sauvé' : 'Sauvegarder'}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Level Thresholds */}
      <div className="section-toolbar">
        <h3 style={{ color: '#9ca3af', fontSize: '0.875rem', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.04em', margin: 0 }}>
          Paliers de Level
        </h3>
        <button className="btn-save" disabled={savingThresholds} onClick={saveThresholds}>
          {savingThresholds ? '…' : savedThresholds ? '✓ Sauvegardé' : 'Sauvegarder tout'}
        </button>
      </div>
      <div className="admin-table-wrapper" style={{ marginTop: '0.75rem' }}>
        <table className="admin-table">
          <thead>
            <tr>
              <th>Level</th>
              <th>XP minimum</th>
            </tr>
          </thead>
          <tbody>
            {editThresholds.length === 0 ? (
              <tr className="loading-row"><td colSpan={2}>Chargement…</td></tr>
            ) : editThresholds.map((t, i) => (
              <tr key={t.level}>
                <td style={{ fontWeight: 600, color: '#818cf8' }}>Niveau {t.level}</td>
                <td>
                  <input
                    className="inline-input"
                    type="number"
                    min={0}
                    value={t.min_xp}
                    onChange={(e) => {
                      const v = parseInt(e.target.value, 10)
                      setEditThresholds((prev) => prev.map((x, idx) => idx === i ? { ...x, min_xp: isNaN(v) ? 0 : v } : x))
                    }}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
