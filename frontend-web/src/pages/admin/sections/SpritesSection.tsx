import { useEffect, useState, useRef, useCallback } from 'react'
import { listSprites, uploadSprite, deleteSprite, type SpriteEntry } from '../../../api/admin'

const SPRITE_URL = '/api/sprites'

function flattenEntries(entries: SpriteEntry[], prefix = ''): SpriteEntry[] {
  const result: SpriteEntry[] = []
  for (const e of entries) {
    const path = prefix ? `${prefix}/${e.name}` : e.name
    result.push({ ...e, path })
    if (e.children) {
      result.push(...flattenEntries(e.children, path))
    }
  }
  return result
}

/** Guess frame count from sprite filename: spritesheet → 4 frames, else 1 */
function guessFrameCount(name: string, path: string): number {
  if (name !== 'spritesheet.png') return 1
  return 4
}

/** Guessed background-size X for a spritesheet (4 frames × 64px = 256px) */
function guessSheetWidth(_path: string): number {
  return 256
}

export default function SpritesSection() {
  const [entries, setEntries] = useState<SpriteEntry[]>([])
  const [flat, setFlat] = useState<SpriteEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [uploading, setUploading] = useState(false)
  const [uploadDest, setUploadDest] = useState('default/characters/male')
  const fileRef = useRef<HTMLInputElement>(null)

  // animation preview state
  const [animPath, setAnimPath] = useState<string | null>(null)
  const [animPlaying, setAnimPlaying] = useState(false)

  async function load() {
    setLoading(true)
    try {
      const data = await listSprites()
      setEntries(data)
      setFlat(flattenEntries(data))
    } catch {
      setError('Erreur lors du chargement')
    }
    setLoading(false)
  }

  useEffect(() => { load() }, [])

  async function handleUpload() {
    const file = fileRef.current?.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      await uploadSprite(file, uploadDest)
      fileRef.current!.value = ''
      await load()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Erreur upload')
    }
    setUploading(false)
  }

  async function handleDelete(path: string) {
    if (!window.confirm(`Supprimer ${path} ?`)) return
    try {
      await deleteSprite(path)
      await load()
    } catch {
      setError('Erreur suppression')
    }
  }

  function toggleAnim(path: string, name: string) {
    if (name !== 'spritesheet.png') return
    if (animPath === path) {
      setAnimPath(null)
    } else {
      setAnimPath(path)
      setAnimPlaying(true)
    }
  }

  function renderTree(children: SpriteEntry[], depth = 0): JSX.Element[] {
    return children.flatMap(e => {
      const indent = { paddingLeft: `${depth * 20 + 8}px` }
      const isImg = !e.is_dir && e.name.endsWith('.png')
      const isSheet = e.name === 'spritesheet.png'
      const fullPath = e.path.replace(/\\/g, '/')
      return [
        <div key={fullPath} className="sprite-row" style={indent}>
          <span className="sprite-icon">{e.is_dir ? '📁' : '🖼'}</span>
          <span className="sprite-name">{e.name}</span>
          {e.is_dir && e.size !== undefined && (
            <span className="sprite-count">{e.children?.length ?? 0} items</span>
          )}
          {isImg && !isSheet && (
            <img
              src={`${SPRITE_URL}/${fullPath}`}
              alt={e.name}
              className="sprite-preview"
            />
          )}
          {isSheet && (
            <div
              className="sprite-preview sprite-preview--anim"
              style={{
                backgroundImage: `url(${SPRITE_URL}/${fullPath})`,
                backgroundSize: `${guessSheetWidth(fullPath)}px 64px`,
                animation: animPath === fullPath && animPlaying
                  ? `sprite-step 1s steps(${guessFrameCount(e.name, fullPath)}) infinite`
                  : 'none',
              }}
              onMouseEnter={() => { setAnimPath(fullPath); setAnimPlaying(true) }}
              onMouseLeave={() => { setAnimPath(null); setAnimPlaying(false) }}
            />
          )}
          {isSheet && (
            <span className="sprite-anim-badge">SPRITESHEET</span>
          )}
          {!e.is_dir && (
            <button className="btn btn--style-red btn--sm" onClick={() => handleDelete(fullPath)}>
              Suppr
            </button>
          )}
        </div>,
        ...(e.children ? renderTree(e.children, depth + 1) : []),
      ]
    })
  }

  return (
    <div className="section-card sprites-section">
      <h2>Sprites</h2>

      <div className="upload-box">
        <h4>Uploader un sprite</h4>
        <div className="upload-form">
          <input ref={fileRef} type="file" accept=".png" />
          <select value={uploadDest} onChange={e => setUploadDest(e.target.value)}>
            <option value="default/characters/male">default/characters/male</option>
            <option value="default/characters/female">default/characters/female</option>
            <option value="default/markers">default/markers</option>
            <option value="default/animals/dogs">default/animals/dogs</option>
            <option value="default/animals/cats">default/animals/cats</option>
            <option value="shop">shop/{'{slug}'}</option>
          </select>
          <button className="btn btn--style-yellow" onClick={handleUpload} disabled={uploading}>
            {uploading ? 'Upload...' : 'Upload'}
          </button>
        </div>
      </div>

      {error && <div className="alert-error">{error}</div>}

      <div className="sprite-tree">
        {loading ? <p>Chargement...</p> : renderTree(entries)}
      </div>
    </div>
  )
}
