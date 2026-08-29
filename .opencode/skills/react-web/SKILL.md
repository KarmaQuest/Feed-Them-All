---
name: react-web
description: Expert React + TypeScript frontend development for FeedThemAll web app. Use for all code in frontend-web/ — components, pages, Leaflet map, avatar rendering, API calls, and Zustand store.
---

# React Web — FeedThemAll

## When to Use This Skill
- Writing or reviewing any file in `frontend-web/`
- Building Leaflet map components with pixel art markers
- Creating or updating React components or pages
- Managing state with Zustand
- Calling the Go backend API
- Implementing WebSocket client for real-time map updates

## Project Structure
```
frontend-web/
├── src/
│   ├── api/              # All fetch/axios calls — never call fetch directly in components
│   ├── components/
│   │   ├── map/          # MapView, PingMarker, AvatarMarker, MapControls
│   │   ├── avatar/       # AvatarSprite, AvatarCustomizer
│   │   └── ui/           # Button, Modal, XPBar, DialogBox (RPG style)
│   ├── pages/            # Home (map), Login, Register, Profile, Avatar
│   ├── store/            # Zustand stores (auth, map, websocket)
│   └── main.tsx
└── public/
    └── assets/sprites/
        ├── characters/   # {skin_name}.png spritesheets
        ├── animals/      # cat.png, dog.png
        └── markers/      # bowl.png, paw.png, star.png
```

## Code Conventions

### TypeScript
- No `any` — use proper interfaces and types
- Shared types imported from `../../shared/types/`
- Props always typed with an interface:
```tsx
interface PingMarkerProps {
  ping: Ping;
  onClick: (id: string) => void;
}
```

### Components
- Functional components only — no class components
- Custom hooks for reusable logic (`useGeolocation`, `useWebSocket`, `usePings`)
- Keep components small: if a component exceeds ~150 lines, split it

### API Layer (`src/api/`)
```ts
// src/api/pings.ts
export async function getPingsNearby(lat: number, lon: number, radius: number): Promise<Ping[]> {
  const res = await apiClient.get(`/pings?lat=${lat}&lon=${lon}&radius=${radius}`);
  return res.data;
}
// Never: fetch('/api/pings') directly inside a component
```

### Zustand Store
```ts
// src/store/auth.ts
interface AuthStore {
  user: User | null;
  isLogged: boolean;
  login: (user: User) => void;
  logout: () => void;
}
export const useAuthStore = create<AuthStore>((set) => ({
  user: null,
  isLogged: false,
  login: (user) => set({ user, isLogged: true }),
  logout: () => set({ user: null, isLogged: false }),
}));
```

## Leaflet Map

### Setup
```tsx
import { MapContainer, TileLayer } from 'react-leaflet';

<MapContainer center={[userLat, userLon]} zoom={15}>
  <TileLayer
    url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
    attribution='© OpenStreetMap contributors'
  />
</MapContainer>
```

### Pixel Art Markers
```tsx
import L from 'leaflet';

const pawIcon = L.icon({
  iconUrl: '/assets/sprites/markers/paw.png',
  iconSize: [32, 32],
  iconAnchor: [16, 32],
  popupAnchor: [0, -32],
});

// Avatar marker (player sprite)
const avatarIcon = (spriteUrl: string) => L.divIcon({
  html: `<img src="${spriteUrl}" style="image-rendering:pixelated;width:32px;height:48px" />`,
  iconSize: [32, 48],
  iconAnchor: [16, 48],
  className: '', // clear default Leaflet styles
});
```

### Real-time WebSocket Updates
```ts
// src/store/websocket.ts
// On new ping received → add to map store
// On ping removed → remove from map store
// On feeder position update → update feeder marker position
```

## Avatar Sprite System

### Spritesheet Format (Pokémon-style top-down)
```
Row 0: facing DOWN  (3 frames: idle, step-left, step-right)
Row 1: facing LEFT  (3 frames)
Row 2: facing RIGHT (3 frames)
Row 3: facing UP    (3 frames)
```

### AvatarSprite Component
```tsx
// Config stored in DB as JSON: { skin: 'default', outfit: 'chef', accessory: 'bag' }
// Component composes layers: base sprite + outfit overlay + accessory overlay
<AvatarSprite config={user.avatarConfig} direction="down" isMoving={false} />
```

### CSS Animation (spritesheet stepping)
```css
.avatar-sprite {
  image-rendering: pixelated;
  background-image: url('/assets/sprites/characters/default.png');
  width: 32px;
  height: 48px;
  animation: walk 0.4s steps(3) infinite; /* 3 frames per row */
}
```

## Security
- Never store JWT access token in localStorage — use memory (Zustand) only
- Refresh token is HttpOnly cookie — handled automatically by the browser
- Sanitize any user-generated content before rendering (no dangerouslySetInnerHTML)

## Running Locally
```bash
cd frontend-web
npm install
npm run dev      # http://localhost:3000
npm run lint     # ESLint check
npm run test     # Vitest
```
