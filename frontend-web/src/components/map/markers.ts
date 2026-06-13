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

// ── Ping animal (patte) ───────────────────────────────────────────────────────
const pawSVG = `
<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32">
  <rect width="32" height="32" rx="6" fill="#1a1d27" stroke="#6366f1" stroke-width="2"/>
  <text x="16" y="22" font-size="16" text-anchor="middle">🐾</text>
</svg>`

export const animalIcon = L.divIcon({
  html: `<div style="image-rendering:pixelated">${pawSVG}</div>`,
  iconSize: [32, 32],
  iconAnchor: [16, 32],
  popupAnchor: [0, -34],
  className: '',
})

// ── Ping nourriture (gamelle) ──────────────────────────────────────────────────
const bowlSVG = `
<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32">
  <rect width="32" height="32" rx="6" fill="#1a1d27" stroke="#fbbf24" stroke-width="2"/>
  <text x="16" y="22" font-size="16" text-anchor="middle">🍖</text>
</svg>`

export const foodIcon = L.divIcon({
  html: `<div style="image-rendering:pixelated">${bowlSVG}</div>`,
  iconSize: [32, 32],
  iconAnchor: [16, 32],
  popupAnchor: [0, -34],
  className: '',
})

// ── Ping nourri (étoile verte) ─────────────────────────────────────────────────
const fedSVG = `
<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32">
  <rect width="32" height="32" rx="6" fill="#1a1d27" stroke="#34d399" stroke-width="2"/>
  <text x="16" y="22" font-size="16" text-anchor="middle">✅</text>
</svg>`

export const fedIcon = L.divIcon({
  html: `<div style="image-rendering:pixelated">${fedSVG}</div>`,
  iconSize: [32, 32],
  iconAnchor: [16, 32],
  popupAnchor: [0, -34],
  className: '',
})

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
