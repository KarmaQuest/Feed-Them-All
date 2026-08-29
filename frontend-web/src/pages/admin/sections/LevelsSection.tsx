// src/pages/admin/sections/LevelsSection.tsx — Gestion des paliers de level.
//
// Tableau éditable des paliers XP. Ajout/suppression de lignes.
// Envoi en une seule requête PUT (remplacement complet du tableau).
import { useState, useEffect } from 'react'
import { listThresholds, replaceThresholds, type LevelThreshold } from '../../../api/admin'

export default function LevelsSection() {
  const [editThresholds, setEditThresholds] = useState<LevelThreshold[]>([])
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    listThresholds().then(data => setEditThresholds(data.map(t => ({ ...t }))))
  }, [])

  async function saveThresholds() {
    setSaving(true)
    try {
      await replaceThresholds(editThresholds)
      setSaved(true)
      setTimeout(() => setSaved(false), 1500)
    } finally {
      setSaving(false)
    }
  }

  function addRow() {
    const maxLevel = editThresholds.length > 0 ? Math.max(...editThresholds.map(t => t.level)) : 0
    setEditThresholds(prev => [...prev, { level: maxLevel + 1, min_xp: 0 }])
  }

  function removeRow(i: number) {
    setEditThresholds(prev => prev.filter((_, idx) => idx !== i))
  }

  return (
    <div>
      <h2 className="admin-section-title">Paliers de Level</h2>

      <div className="section-toolbar">
        <span />
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button className="btn btn--style-yellow" onClick={addRow}>
            + Ajouter un palier
          </button>
          <button className="btn btn--style-green" disabled={saving} onClick={saveThresholds}>
            {saving ? '...' : saved ? 'Sauvegardé ✓' : 'Sauvegarder tout'}
          </button>
        </div>
      </div>

      <div className="admin-table-wrapper">
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
              <tr className="loading-row">
                <td colSpan={3}>Chargement…</td>
              </tr>
            ) : (
              editThresholds.map((t, i) => (
                <tr key={i}>
                  <td>
                    <input
                      className="inline-input"
                      type="number"
                      min={1}
                      value={t.level}
                      onChange={e => {
                        const v = parseInt(e.target.value, 10)
                        setEditThresholds(prev =>
                          prev.map((x, idx) => (idx === i ? { ...x, level: isNaN(v) ? 1 : v } : x))
                        )
                      }}
                    />
                  </td>
                  <td>
                    <input
                      className="inline-input"
                      type="number"
                      min={0}
                      value={t.min_xp}
                      onChange={e => {
                        const v = parseInt(e.target.value, 10)
                        setEditThresholds(prev =>
                          prev.map((x, idx) => (idx === i ? { ...x, min_xp: isNaN(v) ? 0 : v } : x))
                        )
                      }}
                    />
                  </td>
                  <td>
                    <button
                      className="btn-danger"
                      style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem' }}
                      onClick={() => removeRow(i)}
                    >
                      ✕
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
