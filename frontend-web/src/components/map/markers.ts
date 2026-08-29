// src/components/map/markers.ts — Icônes Leaflet pour les marqueurs de la carte.
//
// Utilise des L.divIcon avec du SVG inline pour un rendu pixel art immédiat,
// sans dépendance à des assets PNG (qui arriveront en Phase 8).
// image-rendering: pixelated garantit le style rétro même sur écrans HiDPI.
import L from 'leaflet'

// Corrige le bug Leaflet/Vite : les icônes par défaut pointent vers des assets manquants
delete (L.Icon.Default.prototype as unknown as Record<string, unknown>)._getIconUrl
L.Icon.Default.mergeOptions({
  iconUrl: '',
  shadowUrl: '',
})

// ── Ping icons with breed support and sprite fallback ──────────────────────

const PING_ICON_SIZE = { width: 48, height: 48, anchor: [24, 48] as [number, number] }

// Mapping from ping type / status → sprite filename (without extension)
const ICON_MAP: Record<string, string> = {
  animal: 'paw',
  food: 'bowl',
  fed: 'check',
}

interface PingIconOptions {
  animal_type?: string | null
  animal_breed?: string | null
}

function pingIcon(
  typeOrStatus: string,
  fallbackSvg: string,
  strokeColor: string,
  opts?: PingIconOptions,
): L.DivIcon {
  const spriteName = ICON_MAP[typeOrStatus] || 'paw'
  const breed = opts?.animal_breed
  const animalType = opts?.animal_type
  let spriteSrc: string
  if (breed && animalType && (animalType === 'cat' || animalType === 'dog')) {
    spriteSrc = `/api/sprites/default/animals/${animalType}s/${breed}.png`
  } else {
    spriteSrc = `/api/sprites/default/markers/${spriteName}.png`
  }
  const borderPx = 8
  return L.divIcon({
    html: `<div style="position:relative;width:48px;height:48px;image-rendering:pixelated">
      <img
        src="${spriteSrc}"
        style="width:48px;height:48px;image-rendering:pixelated;border-radius:${borderPx}px;border:0"
        onerror="
          this.style.display='none';
          this.nextElementSibling.style.display='flex';
        "
      />
      <div style="display:none;width:48px;height:48px;border-radius:${borderPx}px;background:#1a1d27;border:1px solid ${strokeColor};align-items:center;justify-content:center;font-size:24px;position:absolute;top:0;left:0">
        ${fallbackSvg}
      </div>
    </div>`,
    iconSize: [PING_ICON_SIZE.width, PING_ICON_SIZE.height],
    iconAnchor: PING_ICON_SIZE.anchor,
    popupAnchor: [0, -50],
    className: '',
  })
}

export function animalIcon(opts?: PingIconOptions) {
  return pingIcon('animal', `<span>🐾</span>`, '#6366f1', opts)
}

export function foodIcon(opts?: PingIconOptions) {
  return pingIcon('food', `<span>🍖</span>`, '#fbbf24', opts)
}

export function fedIcon(opts?: PingIconOptions) {
  return pingIcon('fed', `<span>✅</span>`, '#34d399', opts)
}

// ── Position utilisateur ───────────────────────────────────────────────────────
export const userIcon = L.divIcon({
  html: `<div style="
    width:14px;height:14px;
    background:#6366f1;
    border:3px solid #fff;
    border-radius:50%;
    box-shadow:0 0 0 3px rgba(99,102,241,0.35);
  "></div>`,
  iconSize: [14, 14],
  iconAnchor: [7, 7],
  className: '',
})

// ── Avatar du Feeder connecté ──────────────────────────────────────────────────
export function createAvatarIcon(config?: Record<string, unknown> | null, size = 48): L.DivIcon {
  const outfit = config?.outfit as string | undefined
  const gender = (config?.gender as string) === 'female' ? 'female' : 'male'
  const src = outfit
    ? `/api/sprites/shop/${outfit}/south.png`
    : `/api/sprites/default/characters/${gender}/south.png`
  return L.divIcon({
    html: `<img
      src="${src}"
      style="width:${size}px;height:${size}px;image-rendering:pixelated"
      onerror="this.style.display='none'"
    />`,
    iconSize: [size, size],
    iconAnchor: [size / 2, size / 2],
    className: '',
  })
}

// ── Feeder actif (avatar générique) ───────────────────────────────────────────
export const feederIcon = L.divIcon({
  html: `<div style="
    width:20px;height:20px;
    background:#818cf8;
    border:2px solid #fff;
    border-radius:50%;
    opacity:0.75;
  "></div>`,
  iconSize: [20, 20],
  iconAnchor: [10, 10],
  className: '',
})
