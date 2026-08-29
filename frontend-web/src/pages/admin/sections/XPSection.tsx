// src/pages/admin/sections/XPSection.tsx — Actions XP.
//
// Tableau des xp_actions : inline edit (xp_value + daily_limit) + création.
// Les paliers de level sont dans LevelsSection.tsx.
import { useState, useEffect } from 'react'
import {
  listXPActions,
  updateXPAction,
  createXPAction,
  type AdminXPAction,
} from '../../../api/admin'

const EMPTY_NEW_ACTION = { action: '', xp_value: '10', daily_limit: '5' }

export default function XPSection() {
  const [actions, setActions] = useState<AdminXPAction[]>([])
  const [editActions, setEditActions] = useState<
    Record<string, { xp_value: string; daily_limit: string }>
  >({})
  const [savingAction, setSavingAction] = useState<string | null>(null)
  const [savedAction, setSavedAction] = useState<string | null>(null)
  const [newAction, setNewAction] = useState({ ...EMPTY_NEW_ACTION })
  const [creatingAction, setCreatingAction] = useState(false)
  const [createActionError, setCreateActionError] = useState('')

  useEffect(() => {
    listXPActions().then(data => {
      setActions(data)
      const init: Record<string, { xp_value: string; daily_limit: string }> = {}
      data.forEach(a => {
        init[a.action] = { xp_value: String(a.xp_value), daily_limit: String(a.daily_limit) }
      })
      setEditActions(init)
    })
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
      setActions(prev =>
        prev.map(a =>
          a.action === action
            ? {
                ...a,
                xp_value: parseInt(vals.xp_value, 10),
                daily_limit: parseInt(vals.daily_limit, 10),
              }
            : a
        )
      )
      setSavedAction(action)
      setTimeout(() => setSavedAction(null), 1500)
    } finally {
      setSavingAction(null)
    }
  }

  async function handleCreateAction() {
    setCreateActionError('')
    if (!newAction.action) {
      setCreateActionError("Le nom de l'action est requis.")
      return
    }
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
      setActions(prev => [...prev, created])
      setEditActions(prev => ({
        ...prev,
        [created.action]: { xp_value: newAction.xp_value, daily_limit: newAction.daily_limit },
      }))
      setNewAction({ ...EMPTY_NEW_ACTION })
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
      setCreateActionError(msg ?? 'Erreur lors de la creation.')
    } finally {
      setCreatingAction(false)
    }
  }

  return (
    <div>
      <h2 className="admin-section-title">Actions XP</h2>

      <div className="admin-table-wrapper">
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
              <tr className="loading-row">
                <td colSpan={4}>Chargement...</td>
              </tr>
            ) : (
              actions.map(a => (
                <tr key={a.action}>
                  <td>
                    <code style={{ color: '#818cf8', fontSize: '1.2em' }}>{a.action}</code>
                  </td>
                  <td>
                    <input
                      className="inline-input"
                      type="number"
                      min={0}
                      value={editActions[a.action]?.xp_value ?? a.xp_value}
                      onChange={e =>
                        setEditActions(prev => ({
                          ...prev,
                          [a.action]: { ...prev[a.action], xp_value: e.target.value },
                        }))
                      }
                    />
                  </td>
                  <td>
                    <input
                      className="inline-input"
                      type="number"
                      min={1}
                      value={editActions[a.action]?.daily_limit ?? a.daily_limit}
                      onChange={e =>
                        setEditActions(prev => ({
                          ...prev,
                          [a.action]: { ...prev[a.action], daily_limit: e.target.value },
                        }))
                      }
                    />
                  </td>
                  <td>
                    <button
                      className="btn btn--style-green btn--sm"
                      disabled={savingAction === a.action}
                      onClick={() => saveAction(a.action)}
                    >
                      {savingAction === a.action
                        ? '...'
                        : savedAction === a.action
                          ? 'Sauve'
                          : 'Sauvegarder'}
                    </button>
                  </td>
                </tr>
              ))
            )}
            <tr style={{ background: '#1a1f2e' }}>
              <td>
                <input
                  className="inline-input"
                  placeholder="nom_action"
                  value={newAction.action}
                  onChange={e => setNewAction({ ...newAction, action: e.target.value })}
                />
              </td>
              <td>
                <input
                  className="inline-input"
                  type="number"
                  min={0}
                  value={newAction.xp_value}
                  onChange={e => setNewAction({ ...newAction, xp_value: e.target.value })}
                />
              </td>
              <td>
                <input
                  className="inline-input"
                  type="number"
                  min={1}
                  value={newAction.daily_limit}
                  onChange={e => setNewAction({ ...newAction, daily_limit: e.target.value })}
                />
              </td>
              <td>
                <button className="btn btn--style-yellow btn--sm" disabled={creatingAction} onClick={handleCreateAction}>
                  {creatingAction ? '...' : '+ Creer'}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        {createActionError && <p className="error-msg">{createActionError}</p>}
      </div>
    </div>
  )
}
