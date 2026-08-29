import { useState, useEffect, useRef } from 'react'
import {
  listShopItems,
  createShopItem,
  updateShopItem,
  deleteShopItem,
  listSprites,
  uploadShopSprite,
  type AdminShopItem,
  type SpriteEntry,
} from '../../../api/admin'

const EMPTY_ITEM: Omit<AdminShopItem, 'id'> = {
  slug: '',
  name: '',
  category: 'skin',
  price_cents: 0,
  currency: 'usd',
  unlock_condition: null,
  is_active: true,
}

function slugify(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/(^_+|_+$)/g, '')
}

function findUniqueSlug(base: string, existing: AdminShopItem[], currentId?: string): string {
  let slug = base
  let counter = 0
  while (existing.some(it => it.slug === slug && it.id !== currentId)) {
    counter++
    slug = `${base}_${String(counter).padStart(2, '0')}`
  }
  return slug
}

function formatPrice(cents: number, currency: string) {
  if (cents === 0) return <span className="text-muted">Gratuit</span>
  return (
    <span>
      {(cents / 100).toFixed(2)} {currency.toUpperCase()}
    </span>
  )
}

export default function ShopSection() {
  const [items, setItems] = useState<AdminShopItem[]>([])
  const [loading, setLoading] = useState(false)
  const [modal, setModal] = useState<null | {
    mode: 'create' | 'edit'
    data: Omit<AdminShopItem, 'id'>
    id?: string
  }>(null)
  const [condRaw, setCondRaw] = useState('')
  const [condError, setCondError] = useState('')
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState<string | null>(null)

  // Sprite selector state
  const [shopSprites, setShopSprites] = useState<SpriteEntry[]>([])
  const [selectedSprite, setSelectedSprite] = useState<string | null>(null)
  const [slugTouched, setSlugTouched] = useState(false)
  const [uploadingSprite, setUploadingSprite] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)

  // Animated preview
  const [animPreview, setAnimPreview] = useState<string | null>(null)
  const [animPlaying, setAnimPlaying] = useState(true)

  useEffect(() => {
    setLoading(true)
    listShopItems().then(d => {
      setItems(d)
      setLoading(false)
    })
  }, [])

  function openCreate() {
    setModal({ mode: 'create', data: { ...EMPTY_ITEM } })
    setCondRaw('')
    setCondError('')
    setSlugTouched(false)
    setSelectedSprite(null)
    loadShopSprites()
  }

  function openEdit(it: AdminShopItem) {
    setModal({
      mode: 'edit',
      data: {
        slug: it.slug,
        name: it.name,
        category: it.category,
        price_cents: it.price_cents,
        currency: it.currency,
        unlock_condition: it.unlock_condition,
        is_active: it.is_active,
      },
      id: it.id,
    })
    setCondRaw(it.unlock_condition ? JSON.stringify(it.unlock_condition, null, 2) : '')
    setCondError('')
    setSlugTouched(true)
    setSelectedSprite(it.slug)
    loadShopSprites()
  }

  async function loadShopSprites() {
    try {
      const all = await listSprites()
      const shop = all.find(e => e.name === 'shop')
      setShopSprites(shop?.children ?? [])
    } catch {
      setShopSprites([])
    }
  }

  function closeModal() {
    setModal(null)
    setCondError('')
    setAnimPreview(null)
  }

  function updateField(field: keyof Omit<AdminShopItem, 'id'>, value: unknown) {
    if (!modal) return
    setModal({ ...modal, data: { ...modal.data, [field]: value } })
  }

  function handleNameChange(name: string) {
    if (!modal) return
    setModal({ ...modal, data: { ...modal.data, name } })
    if (!slugTouched) {
      const base = slugify(name)
      if (base) {
        const slug = findUniqueSlug(base, items, modal.id)
        setModal({ ...modal, data: { ...modal.data, name, slug } })
      }
    }
  }

  function handleSpriteSelect(spriteName: string) {
    if (!modal) return
    setSelectedSprite(spriteName)
    if (modal.mode === 'create' && !slugTouched) {
      setModal({ ...modal, data: { ...modal.data, slug: spriteName } })
    }
  }

  async function handleSpriteUpload() {
    if (!modal || !fileRef.current?.files?.[0]) return
    const slug = modal.data.slug || slugify(modal.data.name)
    if (!slug) return
    setUploadingSprite(true)
    try {
      await uploadShopSprite(slug, fileRef.current.files[0])
      fileRef.current.value = ''
      await loadShopSprites()
      if (!slugTouched) {
        setSelectedSprite(slug)
      }
    } catch {
      /* ignore */
    }
    setUploadingSprite(false)
  }

  async function handleSave() {
    if (!modal) return
    let unlock_condition: unknown = null
    if (condRaw.trim()) {
      try {
        unlock_condition = JSON.parse(condRaw)
        setCondError('')
      } catch {
        setCondError('JSON invalide')
        return
      }
    }
    const payload = { ...modal.data, unlock_condition }
    setSaving(true)
    try {
      if (modal.mode === 'create') {
        const { id } = await createShopItem(payload as Omit<AdminShopItem, 'id'>)
        setItems(prev => [...prev, { id, ...payload } as AdminShopItem])
      } else if (modal.id) {
        await updateShopItem(modal.id, payload as Omit<AdminShopItem, 'id'>)
        setItems(prev =>
          prev.map(x => (x.id === modal.id ? ({ id: modal.id!, ...payload } as AdminShopItem) : x))
        )
      }
      closeModal()
    } catch {
      /* ignore */
    }
    setSaving(false)
  }

  async function handleDelete(id: string) {
    setDeleting(id)
    try {
      await deleteShopItem(id)
      setItems(prev => prev.filter(x => x.id !== id))
    } finally {
      setDeleting(null)
    }
  }

  const CATEGORIES = ['skin', 'outfit', 'accessory']

  function spriteHasAnim(sprite: SpriteEntry): boolean {
    return !!sprite.children?.some(c => c.name === 'idle' || c.name === 'spritesheet.png')
  }

  function getSpritePreview(sprite: SpriteEntry): string {
    if (sprite.children?.some(c => c.name === 'south.png')) {
      return `/api/sprites/shop/${sprite.name}/south.png`
    }
    const png = sprite.children?.find(c => c.name.endsWith('.png'))
    return png ? `/api/sprites/shop/${sprite.name}/${png.name}` : ''
  }

  return (
    <div>
      <h2 className="admin-section-title">Boutique Skins</h2>
      <button className="btn btn--style-yellow" onClick={openCreate}>
        ＋ Nouvel item
      </button>

      <div className="admin-table-wrapper">
        <table className="admin-table">
          <thead>
            <tr>
              <th>Slug</th>
              <th>Nom</th>
              <th>Catégorie</th>
              <th>Prix</th>
              <th>Condition</th>
              <th>Actif</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr className="loading-row">
                <td colSpan={7}>Chargement…</td>
              </tr>
            ) : items.length === 0 ? (
              <tr className="loading-row">
                <td colSpan={7}>Aucun item</td>
              </tr>
            ) : (
              items.map(it => (
                <tr key={it.id}>
                  <td>
                    <code style={{ color: '#818cf8' }}>{it.slug}</code>
                  </td>
                  <td style={{ fontWeight: 600 }}>{it.name}</td>
                  <td>
                    <span
                      className={`badge-role ${it.category}`}
                      style={{
                        background:
                          it.category === 'skin'
                            ? 'rgba(139,92,246,0.14)'
                            : it.category === 'outfit'
                              ? 'rgba(16,185,129,0.14)'
                              : 'rgba(245,158,11,0.14)',
                        color:
                          it.category === 'skin'
                            ? '#a78bfa'
                            : it.category === 'outfit'
                              ? '#34d399'
                              : '#fbbf24',
                      }}
                    >
                      {it.category}
                    </span>
                  </td>
                  <td>{formatPrice(it.price_cents, it.currency)}</td>
                  <td>
                    {it.unlock_condition ? (
                      <code className="text-mono text-muted">
                        {JSON.stringify(it.unlock_condition)}
                      </code>
                    ) : (
                      <span className="text-muted">—</span>
                    )}
                  </td>
                  <td>
                    <span className={`badge-status ${it.is_active ? 'active' : 'inactive'}`} />
                    {it.is_active ? 'Oui' : 'Non'}
                  </td>
                  <td style={{ display: 'flex', gap: '0.5rem' }}>
                    <button className="btn btn--style-green btn--sm" onClick={() => openEdit(it)}>
                      Modifier
                    </button>
                    <button
                      className="btn btn--style-red btn--sm"
                      disabled={deleting === it.id}
                      onClick={() => handleDelete(it.id)}
                    >
                      {deleting === it.id ? '…' : 'Supprimer'}
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {modal && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-box shop-modal" onClick={e => e.stopPropagation()}>
            <h3>{modal.mode === 'create' ? "Nouvel item" : "Modifier l'item"}</h3>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0 1rem' }}>
              <div className="modal-field">
                <label>Slug</label>
                <input
                  type="text"
                  value={modal.data.slug}
                  onChange={e => { setSlugTouched(true); updateField('slug', e.target.value) }}
                  placeholder="skin_default"
                  readOnly={modal.mode === 'edit'}
                  style={modal.mode === 'edit' ? { opacity: 0.6, cursor: 'not-allowed' } : {}}
                />
                {modal.mode === 'edit' && (
                  <span className="text-muted" style={{ fontSize: '0.7rem' }}>Non modifiable (achats existants)</span>
                )}
              </div>
              <div className="modal-field">
                <label>Nom</label>
                <input
                  type="text"
                  value={modal.data.name}
                  onChange={e => handleNameChange(e.target.value)}
                  placeholder="Skin de base"
                />
              </div>
              <div className="modal-field">
                <label>Catégorie</label>
                <select
                  value={modal.data.category}
                  onChange={e => updateField('category', e.target.value)}
                >
                  {CATEGORIES.map(c => (
                    <option key={c} value={c}>
                      {c}
                    </option>
                  ))}
                </select>
              </div>
              <div className="modal-field">
                <label>Devise</label>
                <select
                  value={modal.data.currency}
                  onChange={e => updateField('currency', e.target.value)}
                >
                  <option value="usd">USD</option>
                  <option value="eur">EUR</option>
                </select>
              </div>
              <div className="modal-field">
                <label>Prix (centimes)</label>
                <input
                  type="number"
                  min={0}
                  value={modal.data.price_cents}
                  onChange={e => updateField('price_cents', parseInt(e.target.value, 10) || 0)}
                />
              </div>
              <div className="modal-field">
                <label>Actif</label>
                <select
                  value={modal.data.is_active ? 'true' : 'false'}
                  onChange={e => updateField('is_active', e.target.value === 'true')}
                >
                  <option value="true">Oui</option>
                  <option value="false">Non</option>
                </select>
              </div>
            </div>

            <div className="modal-field">
              <label>Condition de déverrouillage (JSON, laisser vide si aucune)</label>
              <textarea
                value={condRaw}
                onChange={e => setCondRaw(e.target.value)}
                rows={3}
                placeholder={`{"type":"action_count","action":"feed","value":5}`}
              />
              {condError && (
                <span className="text-red" style={{ fontSize: '0.8rem' }}>
                  {condError}
                </span>
              )}
            </div>

            {/* Sprite selector */}
            <div className="sprite-selector-section">
              <label>Sprite associé</label>
              <div className="sprite-upload-inline">
                <input ref={fileRef} type="file" accept=".png" />
                <button
                  className="btn btn--style-yellow btn--sm"
                  onClick={handleSpriteUpload}
                  disabled={uploadingSprite || !modal.data.slug}
                >
                  {uploadingSprite ? '…' : '＋ Upload'}
                </button>
              </div>
              <div className="sprite-selector-grid">
                {shopSprites.length === 0 ? (
                  <span className="text-muted" style={{ gridColumn: '1 / -1', padding: '0.5rem' }}>
                    Aucun sprite uploadé. Crée un dossier dans l'onglet Sprites ou utilise "＋ Upload".
                  </span>
                ) : (
                  shopSprites.map(s => {
                    const preview = getSpritePreview(s)
                    const isSelected = selectedSprite === s.name
                    const hasAnim = spriteHasAnim(s)
                    return (
                      <div
                        key={s.name}
                        className={`sprite-card ${isSelected ? 'selected' : ''}`}
                        onClick={() => handleSpriteSelect(s.name)}
                      >
                        {preview ? (
                          <img
                            src={preview}
                            alt={s.name}
                            className="sprite-card-img"
                          />
                        ) : (
                          <div className="sprite-card-placeholder">?</div>
                        )}
                        <span className="sprite-card-label">{s.name}</span>
                        {hasAnim && (
                          <button
                            className="sprite-card-anim-btn"
                            title="Voir animation"
                            onClick={e => {
                              e.stopPropagation()
                              const ssrc = `/api/sprites/shop/${s.name}/spritesheet.png`
                              setAnimPreview(animPreview === ssrc ? null : ssrc)
                              setAnimPlaying(true)
                            }}
                          >
                            ▶
                          </button>
                        )}
                      </div>
                    )
                  })
                )}
              </div>
            </div>

            {/* Animation preview modal */}
            {animPreview && (
              <div className="anim-overlay" onClick={() => setAnimPreview(null)}>
                <div className="anim-modal" onClick={e => e.stopPropagation()}>
                  <h4>Animation preview</h4>
                  <div
                    className="anim-stage"
                    onClick={() => setAnimPlaying(p => !p)}
                    style={{
                      width: 64,
                      height: 64,
                      backgroundImage: `url(${animPreview})`,
                      backgroundSize: '256px 64px',
                      animation: animPlaying ? `sprite-step 1s steps(4) infinite` : 'none',
                    }}
                  />
                  <button className="btn-cancel" onClick={() => setAnimPreview(null)}>
                    Fermer
                  </button>
                </div>
              </div>
            )}

            <div className="modal-actions">
              <button className="btn btn--style-red" onClick={closeModal}>
                Annuler
              </button>
              <button className="btn btn--style-green" disabled={saving} onClick={handleSave}>
                {saving ? '…' : modal.mode === 'create' ? 'Créer' : 'Sauvegarder'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
