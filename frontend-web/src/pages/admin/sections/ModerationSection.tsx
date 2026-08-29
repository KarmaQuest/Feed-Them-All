// src/pages/admin/sections/ModerationSection.tsx — Moderation des pings.
//
// Tableau des pings avec filtres (actifs / signales).
// Bouton "Desactiver" force is_active=false sans verifier le proprietaire.
// Bouton "Activites" ouvre une popup listant les feeding events avec edit/delete.
// Bouton "Commentaires" ouvre une popup listant les commentaires avec edit/delete.
// Bouton "Creer ping" pour injecter un ping manuellement (admin/test).
import { useState, useEffect, useCallback } from 'react'
import {
  listPingsAdmin,
  forceDeactivatePing,
  createPingAdmin,
  listFeedingEventsAdmin,
  createFeedingEventAdmin,
  updateFeedingEvent,
  deleteFeedingEvent,
  listCommentsAdmin,
  createComment,
  updateComment,
  deleteComment,
  type AdminPing,
  type AdminFeedingEvent,
  type AdminComment,
} from '../../../api/admin'
import { useAuthStore } from '../../../store/auth'

const ANIMAL_TYPES = ['cat', 'dog', 'other']

export default function ModerationSection() {
  const currentUser = useAuthStore(s => s.user)
  const [pings, setPings] = useState<AdminPing[]>([])
  const [loading, setLoading] = useState(false)
  const [activeOnly, setActiveOnly] = useState(true)
  const [flaggedOnly, setFlaggedOnly] = useState(false)
  const [deactivating, setDeactivating] = useState<string | null>(null)

  // ── Create ping modal ──────────────────────────────────────────────────────
  const [modal, setModal] = useState(false)
  const [newPing, setNewPing] = useState({
    user_id: '',
    type: 'animal',
    lat: '',
    lon: '',
    animal_type: 'cat',
    animal_count: '1',
  })
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')

  // ── Feeding events popup ────────────────────────────────────────────────────
  const [historyPing, setHistoryPing] = useState<AdminPing | null>(null)
  const [historyEvents, setHistoryEvents] = useState<AdminFeedingEvent[]>([])
  const [historyLoading, setHistoryLoading] = useState(false)
  const [editingEvent, setEditingEvent] = useState<{
    id: string
    note: string
    count: string
  } | null>(null)
  const [deletingEvent, setDeletingEvent] = useState<string | null>(null)
  const [newEvent, setNewEvent] = useState({ note: '', count: '' })
  const [addingEvent, setAddingEvent] = useState(false)

  // ── Comments popup ────────────────────────────────────────────────────────
  const [commentPing, setCommentPing] = useState<AdminPing | null>(null)
  const [comments, setComments] = useState<AdminComment[]>([])
  const [commentsLoading, setCommentsLoading] = useState(false)
  const [editingComment, setEditingComment] = useState<{ id: string; content: string } | null>(null)
  const [deletingComment, setDeletingComment] = useState<string | null>(null)
  const [newCommentText, setNewCommentText] = useState('')
  const [addingComment, setAddingComment] = useState(false)

  const fetchPings = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listPingsAdmin({
        active: activeOnly || undefined,
        flagged: flaggedOnly || undefined,
      })
      setPings(data)
    } finally {
      setLoading(false)
    }
  }, [activeOnly, flaggedOnly])

  useEffect(() => {
    fetchPings()
  }, [fetchPings])

  async function handleDeactivate(id: string) {
    setDeactivating(id)
    try {
      await forceDeactivatePing(id)
      setPings(prev => prev.map(p => (p.id === id ? { ...p, is_active: false } : p)))
    } finally {
      setDeactivating(null)
    }
  }

  async function handleOpenHistory(ping: AdminPing) {
    setHistoryPing(ping)
    setHistoryEvents([])
    setEditingEvent(null)
    setNewEvent({ note: '', count: '' })
    setHistoryLoading(true)
    try {
      const events = await listFeedingEventsAdmin(ping.id)
      setHistoryEvents(events)
    } finally {
      setHistoryLoading(false)
    }
  }

  async function handleAddEvent() {
    if (!historyPing) return
    setAddingEvent(true)
    try {
      const body: { note?: string | null; animal_count_seen?: number | null } = {}
      if (newEvent.note) body.note = newEvent.note
      if (newEvent.count) body.animal_count_seen = parseInt(newEvent.count, 10)
      const created = await createFeedingEventAdmin(historyPing.id, body)
      setHistoryEvents(prev => [created, ...prev])
      setNewEvent({ note: '', count: '' })
    } finally {
      setAddingEvent(false)
    }
  }

  async function handleDeleteEvent(id: string) {
    setDeletingEvent(id)
    try {
      await deleteFeedingEvent(id)
      setHistoryEvents(prev => prev.filter(e => e.id !== id))
    } finally {
      setDeletingEvent(null)
    }
  }

  async function handleSaveEvent() {
    if (!editingEvent) return
    await updateFeedingEvent(editingEvent.id, {
      note: editingEvent.note || null,
      animal_count_seen: editingEvent.count ? parseInt(editingEvent.count, 10) : null,
    })
    setHistoryEvents(prev =>
      prev.map(e =>
        e.id === editingEvent.id
          ? {
              ...e,
              note: editingEvent.note || null,
              animal_count_seen: editingEvent.count ? parseInt(editingEvent.count, 10) : null,
            }
          : e
      )
    )
    setEditingEvent(null)
  }

  async function handleOpenComments(ping: AdminPing) {
    setCommentPing(ping)
    setComments([])
    setEditingComment(null)
    setNewCommentText('')
    setCommentsLoading(true)
    try {
      const data = await listCommentsAdmin(ping.id)
      setComments(data)
    } finally {
      setCommentsLoading(false)
    }
  }

  async function handleAddComment() {
    if (!commentPing || !newCommentText.trim()) return
    setAddingComment(true)
    try {
      const created = await createComment(commentPing.id, newCommentText.trim())
      setComments(prev => [created, ...prev])
      setNewCommentText('')
    } finally {
      setAddingComment(false)
    }
  }

  async function handleDeleteComment(id: string) {
    setDeletingComment(id)
    try {
      await deleteComment(id)
      setComments(prev => prev.filter(c => c.id !== id))
    } finally {
      setDeletingComment(null)
    }
  }

  async function handleSaveComment() {
    if (!editingComment) return
    await updateComment(editingComment.id, editingComment.content)
    setComments(prev =>
      prev.map(c => (c.id === editingComment.id ? { ...c, content: editingComment.content } : c))
    )
    setEditingComment(null)
  }

  async function handleCreate() {
    setCreateError('')
    const lat = parseFloat(newPing.lat)
    const lon = parseFloat(newPing.lon)
    const count = parseInt(newPing.animal_count, 10)
    if (!newPing.user_id) {
      setCreateError("L'ID utilisateur est requis.")
      return
    }
    if (isNaN(lat) || isNaN(lon)) {
      setCreateError('Latitude et longitude doivent etre des nombres.')
      return
    }
    if (newPing.type === 'animal' && (isNaN(count) || count < 1)) {
      setCreateError("Le nombre d'animaux doit etre >= 1.")
      return
    }
    setCreating(true)
    try {
      const body: Parameters<typeof createPingAdmin>[0] = {
        user_id: newPing.user_id,
        type: newPing.type,
        lat,
        lon,
      }
      if (newPing.type === 'animal') {
        body.animal_type = newPing.animal_type
        body.animal_count = count
      }
      await createPingAdmin(body)
      setModal(false)
      setNewPing({
        user_id: currentUser?.id ?? '',
        type: 'animal',
        lat: '',
        lon: '',
        animal_type: 'cat',
        animal_count: '1',
      })
      fetchPings()
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
      setCreateError(msg ?? 'Erreur lors de la creation.')
    } finally {
      setCreating(false)
    }
  }

  function formatDateTime(iso: string) {
    const d = new Date(iso)
    return (
      d.toLocaleDateString('fr-FR') +
      ' ' +
      d.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })
    )
  }

  return (
    <div>
      <h2 className="admin-section-title">Moderation des Pings</h2>

      <div className="section-toolbar">
        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
          <label
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              color: '#9ca3af',
              fontSize: '0.875rem',
              cursor: 'pointer',
            }}
          >
            <input
              type="checkbox"
              checked={activeOnly}
              onChange={e => setActiveOnly(e.target.checked)}
            />
            Actifs seulement
          </label>
          <label
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              color: '#9ca3af',
              fontSize: '0.875rem',
              cursor: 'pointer',
            }}
          >
            <input
              type="checkbox"
              checked={flaggedOnly}
              onChange={e => setFlaggedOnly(e.target.checked)}
            />
            Signales seulement
          </label>
        </div>
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <span className="text-muted">{pings.length} pings</span>
          <button
            className="btn btn--style-yellow"
            onClick={() => {
              setModal(true)
              setCreateError('')
              setNewPing({
                user_id: currentUser?.id ?? '',
                type: 'animal',
                lat: '',
                lon: '',
                animal_type: 'cat',
                animal_count: '1',
              })
            }}
          >
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
              <th>Animal</th>
              <th>Cree par</th>
              <th>Reports</th>
              <th>Statut</th>
              <th>Cree le</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr className="loading-row">
                <td colSpan={8}>Chargement...</td>
              </tr>
            ) : pings.length === 0 ? (
              <tr className="loading-row">
                <td colSpan={8}>Aucun ping trouve</td>
              </tr>
            ) : (
              pings.map(p => (
                <tr key={p.id}>
                  <td className="text-mono text-muted">{p.id.slice(0, 8)}...</td>
                  <td>
                    <span
                      className="badge-role"
                      style={{
                        background:
                          p.type === 'animal' ? 'rgba(16,185,129,0.14)' : 'rgba(245,158,11,0.14)',
                        color: p.type === 'animal' ? '#34d399' : '#fbbf24',
                      }}
                    >
                      {p.type}
                    </span>
                  </td>
                  <td className="text-muted" style={{ fontSize: '0.8rem' }}>
                    {p.animal_type ? (
                      <span>
                        {p.animal_type}{' '}
                        {p.animal_count != null && p.animal_count > 1 ? `×${p.animal_count}` : ''}
                      </span>
                    ) : (
                      <span>—</span>
                    )}
                  </td>
                  <td className="text-mono text-muted">{p.created_by.slice(0, 8)}...</td>
                  <td>
                    {p.report_count > 0 ? (
                      <span className="text-red" style={{ fontWeight: 600 }}>
                        ! {p.report_count}
                      </span>
                    ) : (
                      <span className="text-muted">0</span>
                    )}
                  </td>
                  <td>
                    <span className={`badge-status ${p.is_active ? 'active' : 'inactive'}`} />
                    {p.is_active ? (
                      <span className="text-green">Actif</span>
                    ) : (
                      <span className="text-muted">Inactif</span>
                    )}
                  </td>
                  <td className="text-muted">{p.created_at.slice(0, 10)}</td>
                  <td style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap' }}>
                    <button
                      className="btn-secondary"
                      style={{ fontSize: '0.75rem', padding: '0.25rem 0.6rem' }}
                      onClick={() => handleOpenHistory(p)}
                    >
                      Activites
                    </button>
                    <button
                      className="btn-secondary"
                      style={{
                        fontSize: '0.75rem',
                        padding: '0.25rem 0.6rem',
                        background: 'rgba(59,130,246,0.15)',
                        color: '#93c5fd',
                      }}
                      onClick={() => handleOpenComments(p)}
                    >
                      Commentaires
                    </button>
                    {p.is_active ? (
                      <button
                        className="btn btn--style-red btn--sm"
                        disabled={deactivating === p.id}
                        onClick={() => handleDeactivate(p.id)}
                      >
                        {deactivating === p.id ? '...' : 'Desactiver'}
                      </button>
                    ) : (
                      <span className="text-muted">---</span>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* ── Create ping modal ─────────────────────────────────────────────── */}
      {modal && (
        <div className="modal-overlay" onClick={() => setModal(false)}>
          <div className="modal-box" onClick={e => e.stopPropagation()}>
            <h3>Creer un ping</h3>
            <div className="modal-form-grid">
              <label>ID utilisateur</label>
              <input
                className="inline-input"
                placeholder="UUID de l'utilisateur"
                value={newPing.user_id}
                onChange={e => setNewPing({ ...newPing, user_id: e.target.value })}
              />
              <label>Type</label>
              <select
                className="inline-select"
                value={newPing.type}
                onChange={e => setNewPing({ ...newPing, type: e.target.value })}
              >
                <option value="animal">animal</option>
                <option value="food">food</option>
              </select>
              {newPing.type === 'animal' && (
                <>
                  <label>Type d'animal</label>
                  <select
                    className="inline-select"
                    value={newPing.animal_type}
                    onChange={e => setNewPing({ ...newPing, animal_type: e.target.value })}
                  >
                    {ANIMAL_TYPES.map(t => (
                      <option key={t} value={t}>
                        {t}
                      </option>
                    ))}
                  </select>
                  <label>Nombre d'animaux</label>
                  <input
                    className="inline-input"
                    type="number"
                    min="1"
                    max="100"
                    value={newPing.animal_count}
                    onChange={e => setNewPing({ ...newPing, animal_count: e.target.value })}
                  />
                </>
              )}
              <label>Latitude</label>
              <input
                className="inline-input"
                type="number"
                step="any"
                placeholder="ex: 21.027763"
                value={newPing.lat}
                onChange={e => setNewPing({ ...newPing, lat: e.target.value })}
              />
              <label>Longitude</label>
              <input
                className="inline-input"
                type="number"
                step="any"
                placeholder="ex: 105.834160"
                value={newPing.lon}
                onChange={e => setNewPing({ ...newPing, lon: e.target.value })}
              />
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

      {/* ── Feeding events popup ──────────────────────────────────────────── */}
      {historyPing && (
        <div
          className="modal-overlay"
          onClick={() => {
            setHistoryPing(null)
            setEditingEvent(null)
          }}
        >
          <div
            className="modal-box"
            style={{ minWidth: '580px', maxWidth: '720px' }}
            onClick={e => e.stopPropagation()}
          >
            <h3 style={{ marginBottom: '0.25rem' }}>Activites — {historyPing.id.slice(0, 8)}...</h3>
            <p className="text-muted" style={{ fontSize: '0.8rem', marginBottom: '1rem' }}>
              {historyPing.type}
              {historyPing.animal_type ? ` · ${historyPing.animal_type}` : ''}
              {historyPing.animal_count && historyPing.animal_count > 1
                ? ` ×${historyPing.animal_count}`
                : ''}
              {' · '}Cree le {historyPing.created_at.slice(0, 10)}
            </p>

            {historyLoading ? (
              <p className="text-muted" style={{ textAlign: 'center', padding: '1.5rem 0' }}>
                Chargement...
              </p>
            ) : historyEvents.length === 0 ? (
              <p className="text-muted" style={{ textAlign: 'center', padding: '1.5rem 0' }}>
                Aucun nourrissage enregistre pour ce ping.
              </p>
            ) : (
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '0.6rem',
                  maxHeight: '400px',
                  overflowY: 'auto',
                }}
              >
                {historyEvents.map((ev, i) => (
                  <div
                    key={ev.id}
                    style={{
                      background: 'rgba(255,255,255,0.04)',
                      borderRadius: '8px',
                      padding: '0.75rem 1rem',
                      borderLeft: '3px solid #6366f1',
                    }}
                  >
                    {editingEvent?.id === ev.id ? (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                          <label style={{ fontSize: '0.75rem', color: '#9ca3af', width: '6rem' }}>
                            Note
                          </label>
                          <input
                            className="inline-input"
                            style={{ flex: 1 }}
                            value={editingEvent.note}
                            onChange={e =>
                              setEditingEvent({ ...editingEvent, note: e.target.value })
                            }
                          />
                        </div>
                        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                          <label style={{ fontSize: '0.75rem', color: '#9ca3af', width: '6rem' }}>
                            Animaux vus
                          </label>
                          <input
                            className="inline-input"
                            style={{ width: '5rem' }}
                            type="number"
                            min="1"
                            max="100"
                            value={editingEvent.count}
                            onChange={e =>
                              setEditingEvent({ ...editingEvent, count: e.target.value })
                            }
                          />
                        </div>
                        <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.3rem' }}>
              <button
                className="btn btn--style-green btn--sm"
                onClick={handleSaveEvent}
              >
                Sauver
              </button>
              <button
                className="btn btn--style-red btn--sm"
                onClick={() => setEditingEvent(null)}
              >
                Annuler
              </button>
                        </div>
                      </div>
                    ) : (
                      <>
                        <div
                          style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'flex-start',
                          }}
                        >
                          <div>
                            <span
                              style={{ fontWeight: 600, color: '#c4b5fd', fontSize: '0.85rem' }}
                            >
                              #{historyEvents.length - i}
                            </span>
                            <span
                              className="text-muted"
                              style={{ fontSize: '0.8rem', marginLeft: '0.5rem' }}
                            >
                              par {ev.username}
                            </span>
                          </div>
                          <div style={{ display: 'flex', gap: '0.4rem', alignItems: 'center' }}>
                            <span className="text-muted" style={{ fontSize: '0.8rem' }}>
                              {formatDateTime(ev.fed_at)}
                            </span>
                            <button
                              className="btn-secondary"
                              style={{ fontSize: '0.7rem', padding: '0.15rem 0.5rem' }}
                              onClick={() =>
                                setEditingEvent({
                                  id: ev.id,
                                  note: ev.note ?? '',
                                  count:
                                    ev.animal_count_seen != null
                                      ? String(ev.animal_count_seen)
                                      : '',
                                })
                              }
                            >
                              Editer
                            </button>
                            <button
                              className="btn btn--style-red btn--sm"
                              disabled={deletingEvent === ev.id}
                              onClick={() => handleDeleteEvent(ev.id)}
                            >
                              {deletingEvent === ev.id ? '...' : 'Suppr.'}
                            </button>
                          </div>
                        </div>
                        {ev.animal_count_seen != null && (
                          <p style={{ fontSize: '0.8rem', color: '#34d399', marginTop: '0.3rem' }}>
                            {ev.animal_count_seen} animal(s) vu(s)
                          </p>
                        )}
                        {ev.note && (
                          <p
                            style={{
                              fontSize: '0.85rem',
                              color: '#e5e7eb',
                              marginTop: '0.3rem',
                              fontStyle: 'italic',
                            }}
                          >
                            "{ev.note}"
                          </p>
                        )}
                      </>
                    )}
                  </div>
                ))}
              </div>
            )}

            {/* Formulaire ajout activité */}
            <div
              style={{
                marginTop: '1rem',
                padding: '0.75rem 1rem',
                background: 'rgba(99,102,241,0.08)',
                borderRadius: '8px',
                border: '1px dashed rgba(99,102,241,0.3)',
              }}
            >
              <p
                style={{
                  fontSize: '0.8rem',
                  color: '#a5b4fc',
                  marginBottom: '0.5rem',
                  fontWeight: 600,
                }}
              >
                + Ajouter une activite
              </p>
              <div
                style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'flex-end' }}
              >
                <div
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '0.25rem',
                    flex: 1,
                    minWidth: '120px',
                  }}
                >
                  <label style={{ fontSize: '0.7rem', color: '#9ca3af' }}>Note</label>
                  <input
                    className="inline-input"
                    placeholder="Ex: 3 chats nourris"
                    value={newEvent.note}
                    onChange={e => setNewEvent({ ...newEvent, note: e.target.value })}
                  />
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                  <label style={{ fontSize: '0.7rem', color: '#9ca3af' }}>Animaux vus</label>
                  <input
                    className="inline-input"
                    style={{ width: '5rem' }}
                    type="number"
                    min="1"
                    max="100"
                    placeholder="—"
                    value={newEvent.count}
                    onChange={e => setNewEvent({ ...newEvent, count: e.target.value })}
                  />
                </div>
            <button
              className="btn btn--style-green btn--sm"
              disabled={addingEvent}
              onClick={handleAddEvent}
            >
              {addingEvent ? '...' : 'Ajouter'}
            </button>
              </div>
            </div>

            <div className="modal-actions" style={{ marginTop: '1rem' }}>
              <button
                className="btn-cancel"
                onClick={() => {
                  setHistoryPing(null)
                  setEditingEvent(null)
                }}
              >
                Fermer
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Comments popup ────────────────────────────────────────────────── */}
      {commentPing && (
        <div
          className="modal-overlay"
          onClick={() => {
            setCommentPing(null)
            setEditingComment(null)
          }}
        >
          <div
            className="modal-box"
            style={{ minWidth: '580px', maxWidth: '720px' }}
            onClick={e => e.stopPropagation()}
          >
            <h3 style={{ marginBottom: '0.25rem' }}>
              Commentaires — {commentPing.id.slice(0, 8)}...
            </h3>
            <p className="text-muted" style={{ fontSize: '0.8rem', marginBottom: '1rem' }}>
              Cree le {commentPing.created_at.slice(0, 10)} · {comments.length} commentaire(s)
            </p>

            {commentsLoading ? (
              <p className="text-muted" style={{ textAlign: 'center', padding: '1.5rem 0' }}>
                Chargement...
              </p>
            ) : comments.length === 0 ? (
              <p className="text-muted" style={{ textAlign: 'center', padding: '1.5rem 0' }}>
                Aucun commentaire pour ce ping.
              </p>
            ) : (
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '0.6rem',
                  maxHeight: '400px',
                  overflowY: 'auto',
                }}
              >
                {comments.map(c => (
                  <div
                    key={c.id}
                    style={{
                      background: 'rgba(255,255,255,0.04)',
                      borderRadius: '8px',
                      padding: '0.75rem 1rem',
                      borderLeft: '3px solid #3b82f6',
                    }}
                  >
                    {editingComment?.id === c.id ? (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                        <textarea
                          className="inline-input"
                          style={{ resize: 'vertical', minHeight: '60px', fontFamily: 'inherit' }}
                          maxLength={500}
                          value={editingComment.content}
                          onChange={e =>
                            setEditingComment({ ...editingComment, content: e.target.value })
                          }
                        />
                        <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button
              className="btn btn--style-green btn--sm"
              onClick={handleSaveComment}
            >
              Sauver
            </button>
            <button
              className="btn btn--style-red btn--sm"
              onClick={() => setEditingComment(null)}
            >
              Annuler
            </button>
                        </div>
                      </div>
                    ) : (
                      <>
                        <div
                          style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'flex-start',
                            marginBottom: '0.35rem',
                          }}
                        >
                          <div>
                            <span
                              style={{ fontWeight: 600, color: '#93c5fd', fontSize: '0.85rem' }}
                            >
                              {c.username}
                            </span>
                            <span
                              className="text-muted"
                              style={{ fontSize: '0.75rem', marginLeft: '0.5rem' }}
                            >
                              {formatDateTime(c.created_at)}
                            </span>
                            {c.updated_at !== c.created_at && (
                              <span
                                className="text-muted"
                                style={{ fontSize: '0.7rem', marginLeft: '0.4rem' }}
                              >
                                (modifie)
                              </span>
                            )}
                          </div>
                          <div style={{ display: 'flex', gap: '0.4rem' }}>
                            <button
                              className="btn-secondary"
                              style={{ fontSize: '0.7rem', padding: '0.15rem 0.5rem' }}
                              onClick={() => setEditingComment({ id: c.id, content: c.content })}
                            >
                              Editer
                            </button>
                            <button
                              className="btn btn--style-red btn--sm"
                              disabled={deletingComment === c.id}
                              onClick={() => handleDeleteComment(c.id)}
                            >
                              {deletingComment === c.id ? '...' : 'Suppr.'}
                            </button>
                          </div>
                        </div>
                        <p style={{ fontSize: '0.9rem', color: '#e5e7eb', margin: 0 }}>
                          {c.content}
                        </p>
                      </>
                    )}
                  </div>
                ))}
              </div>
            )}

            {/* Formulaire ajout commentaire */}
            <div
              style={{
                marginTop: '1rem',
                padding: '0.75rem 1rem',
                background: 'rgba(59,130,246,0.08)',
                borderRadius: '8px',
                border: '1px dashed rgba(59,130,246,0.3)',
              }}
            >
              <p
                style={{
                  fontSize: '0.8rem',
                  color: '#93c5fd',
                  marginBottom: '0.5rem',
                  fontWeight: 600,
                }}
              >
                + Ajouter un commentaire
              </p>
              <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'flex-end' }}>
                <textarea
                  className="inline-input"
                  style={{ flex: 1, resize: 'vertical', minHeight: '52px', fontFamily: 'inherit' }}
                  maxLength={500}
                  placeholder="Contenu du commentaire..."
                  value={newCommentText}
                  onChange={e => setNewCommentText(e.target.value)}
                />
            <button
              className="btn btn--style-green btn--sm"
              disabled={addingComment || !newCommentText.trim()}
              onClick={handleAddComment}
            >
              {addingComment ? '...' : 'Ajouter'}
            </button>
              </div>
            </div>

            <div className="modal-actions" style={{ marginTop: '1rem' }}>
              <button
                className="btn-cancel"
                onClick={() => {
                  setCommentPing(null)
                  setEditingComment(null)
                }}
              >
                Fermer
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
