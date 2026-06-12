---
name: pixel-art-design
description: Pixel art asset creation workflow for FeedThemAll. Use for creating sprites, spritesheets, markers, and UI elements using Aseprite and PixelLab.ai, following the Pokémon FireRed/LeafGreen visual style.
---

# Pixel Art Design — FeedThemAll

## When to Use This Skill
- Creating or editing sprite assets for the game (avatars, animals, markers, UI)
- Exporting spritesheets from Aseprite for use in React/Leaflet
- Generating base sprites with PixelLab.ai
- Ensuring visual consistency across all assets
- Implementing pixel art in the React frontend (CSS, Leaflet markers, animations)

## Visual Direction
Style reference: **Pokémon FireRed / LeafGreen (GBA)** — top-down view, pixel art, rétro-cute.
- Characters are small (~16×24px or 32×48px at ×2 scale)
- Clean black outlines, limited palette
- Animations are simple: idle + 2-3 walk frames per direction

## Palette (Pico-8 — 16 colors)
**This palette is mandatory for all assets** — import it in Aseprite and set as the default project palette.

| Role | Hex | Use |
|---|---|---|
| Black | `#000000` | Outlines |
| Dark blue | `#1D2B53` | Shadows, deep outlines |
| Dark purple | `#7E2553` | Accent dark |
| Dark green | `#008751` | Grass, terrain |
| Brown | `#AB5236` | Dirt, paths |
| Dark grey | `#5F574F` | Stone, shadows |
| Light grey | `#C2C3C7` | Highlights |
| White | `#FFF1E8` | Light skin, eyes |
| Red | `#FF004D` | Accent, heart icons |
| Orange | `#FFA300` | Food, warm tones |
| Yellow | `#FFEC27` | Stars, XP, gleams |
| Green | `#00E436` | Active zones, health |
| Blue | `#29ADFF` | Water, UI |
| Lavender | `#83769C` | UI backgrounds |
| Pink | `#FF77A8` | Cat tones, feminine |
| Peach | `#FFCCAA` | Skin tones |

## Spritesheet Format (Characters)

### Grid Layout
```
       Frame 0   Frame 1   Frame 2
Row 0: [DOWN  ] [DOWN-L ] [DOWN-R ]   ← Facing down (toward viewer)
Row 1: [LEFT  ] [LEFT-L ] [LEFT-R ]   ← Facing left
Row 2: [RIGHT ] [RIGHT-L] [RIGHT-R]   ← Facing right
Row 3: [UP    ] [UP-L   ] [UP-R   ]   ← Facing up (away from viewer)
```

### Sprite Dimensions
- **Base size**: 16×24px (width × height) per frame
- **Display size**: ×2 = 32×48px (CSS `image-rendering: pixelated`)
- **Total spritesheet**: 48×96px (3 frames × 4 rows)

### Aseprite Export Settings
```
File > Export Sprite Sheet
  ├── Layout: By Rows
  ├── Sheet Type: Horizontal Strip (or Grid)
  ├── Columns: 3
  ├── Rows: 4
  ├── Export: PNG + JSON (Data)
  └── Scale: 1x (scaling handled in CSS)
```

## Workflow: PixelLab.ai → Aseprite → React

### Step 1: Generate base in PixelLab.ai
1. Go to https://www.pixellab.ai/
2. Set palette to Pico-8 (paste hex values if needed)
3. Prompt: "16x24 pixel art RPG character sprite, top-down view, [description], Pokémon GBA style, black outline"
4. Generate the **idle/facing-down frame** first
5. Download PNG

### Step 2: Refine in Aseprite
1. Open the PixelLab.ai PNG in Aseprite
2. Set canvas to 48×96px (full spritesheet)
3. Paste idle frame in Row 0, Frame 0
4. Manually draw the 3 other direction frames (mirror/adapt)
5. Animate walk cycle: duplicate idle frame, shift legs 2px down/up for step-left and step-right
6. Enforce palette: `Sprite > Color Mode > Indexed` then `Palette > Fit to Palette`
7. Export as spritesheet PNG + JSON

### Step 3: Add to project
```
frontend-web/public/assets/sprites/
├── characters/
│   ├── default.png        ← base feeder avatar
│   ├── chef.png           ← giver avatar (chef hat/apron)
│   └── default.json       ← Aseprite JSON (frame positions)
├── animals/
│   ├── cat.png            ← stray cat (idle + 1 anim)
│   └── dog.png            ← stray dog (idle + 1 anim)
└── markers/
    ├── bowl.png           ← food ping marker (32×32)
    ├── paw.png            ← animal ping marker (32×32)
    └── star.png           ← recently fed zone (32×32)
```

## React Integration

### Image rendering (no blur on pixels)
```css
img.sprite, canvas.sprite {
  image-rendering: pixelated;
  image-rendering: crisp-edges; /* Firefox */
}
```

### Leaflet marker with pixel art
```tsx
const icon = L.divIcon({
  html: `<img
    src="/assets/sprites/markers/paw.png"
    style="image-rendering:pixelated;width:32px;height:32px"
  />`,
  iconSize: [32, 32],
  iconAnchor: [16, 32],
  className: '',
});
```

### CSS walk animation (spritesheet stepping)
```css
/* Character facing down, 3-frame walk cycle */
.sprite-walk-down {
  width: 32px;      /* 16px × 2 */
  height: 48px;     /* 24px × 2 */
  background-image: url('/assets/sprites/characters/default.png');
  background-size: 96px 192px; /* full sheet × 2 */
  background-position: 0px 0px; /* row 0 = down */
  animation: walk-cycle 0.45s steps(3) infinite;
}
@keyframes walk-cycle {
  from { background-position-x: 0px; }
  to   { background-position-x: -96px; }
}
```

## MVP Asset Checklist
- [ ] Palette Pico-8 importée dans Aseprite
- [ ] Avatar Feeder — default (4 directions × 3 frames)
- [ ] Avatar Giver — chef (4 directions × 3 frames)
- [ ] Chat errant — cat.png (idle + walk, 1 direction suffis pour MVP)
- [ ] Chien errant — dog.png (idle + walk)
- [ ] Marqueur gamelle — bowl.png (32×32, statique)
- [ ] Marqueur patte — paw.png (32×32, statique)
- [ ] Marqueur étoile — star.png (32×32, statique ou 2-frame glitter)

## Quality Rules
- Always work at **1×** in Aseprite — no anti-aliasing
- Maximum **16 colors** per sprite (Pico-8 palette only)
- Black outline (`#000000`) on all character sprites
- No subpixel gaps: sprites must tile cleanly on the map
- Test visibility on both light (day) and dark (night) OSM tile backgrounds
