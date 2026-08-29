---
name: react-native
description: Expert React Native + Expo development for FeedThemAll mobile app. Use for all code in frontend-mobile/ — navigation, native GPS, map integration, camera, and shared API/types with the web app.
---

# React Native (Expo) — FeedThemAll

## When to Use This Skill
- Writing or reviewing any file in `frontend-mobile/`
- Setting up Expo navigation, screens, and layouts
- Accessing native GPS (`expo-location`)
- Integrating the map on mobile (React Native Maps or WebView+Leaflet)
- Using the camera for proof-of-feeding photos
- Sharing API logic and TypeScript types with `frontend-web/`

## Project Structure
```
frontend-mobile/
└── src/
    ├── api/          # Symlink or shared copy from shared/ — same calls as web
    ├── components/   # Mobile-specific components
    ├── screens/      # Map, Login, Register, Profile, AvatarCustomizer
    ├── navigation/   # Expo Router or React Navigation stack
    └── store/        # Same Zustand stores as web (shared if possible)
```

## Setup
```bash
cd frontend-mobile
npx create-expo-app . --template blank-typescript
npx expo install expo-location expo-camera expo-image-picker
npm run start   # Expo Dev Tools
```

## Shared Code with Web
Types from `shared/types/` must be used in both apps:
```ts
// shared/types/ping.ts
export interface Ping {
  id: string;
  type: 'animal' | 'food';
  lat: number;
  lon: number;
  createdAt: string;
  isActive: boolean;
}
```
API functions in `src/api/` should be identical to those in `frontend-web/src/api/`.

## Native GPS (expo-location)
```ts
import * as Location from 'expo-location';

// Always request permission before use
const { status } = await Location.requestForegroundPermissionsAsync();
if (status !== 'granted') {
  // Show fallback UI — never crash
  return;
}

const location = await Location.getCurrentPositionAsync({
  accuracy: Location.Accuracy.High,
});
// location.coords.latitude, location.coords.longitude
```

For real-time avatar position updates:
```ts
Location.watchPositionAsync(
  { accuracy: Location.Accuracy.High, timeInterval: 1000, distanceInterval: 5 },
  (loc) => wsStore.sendPosition(loc.coords)
);
```

## Map Integration

**Option A — React Native Maps** (recommended for performance):
```bash
npx expo install react-native-maps
```
```tsx
import MapView, { Marker } from 'react-native-maps';
<MapView style={{ flex: 1 }} initialRegion={region}>
  {pings.map(p => (
    <Marker key={p.id} coordinate={{ latitude: p.lat, longitude: p.lon }}
      image={require('../assets/sprites/markers/paw.png')} />
  ))}
</MapView>
```

**Option B — Leaflet via WebView** (simpler, shares web map code):
```tsx
import { WebView } from 'react-native-webview';
// Serve the React web map as a local HTML file injected into WebView
// Use postMessage to pass GPS position and receive ping tap events
```

## Camera (proof-of-feeding photos)
```ts
import * as ImagePicker from 'expo-image-picker';

const result = await ImagePicker.launchCameraAsync({
  mediaTypes: ImagePicker.MediaTypeOptions.Images,
  quality: 0.7,        // Compress before upload
  allowsEditing: true,
  aspect: [4, 3],
});
if (!result.canceled) {
  await uploadPingMedia(pingId, result.assets[0].uri);
}
```

## Avatar Pixel Art on Mobile
- Use `<Image>` with `resizeMode="contain"` — pixelated rendering not natively supported, but acceptable at ×2/×3 scale
- For animation: use `Animated.Image` stepping through frames, or a small spritesheet lib like `react-native-sprite-animator`

## Performance Rules (React Native)
- Use `FlatList` instead of `ScrollView` for any list of pings or players
- Memoize heavy components with `React.memo` and `useMemo`
- Avoid inline styles in render — define `StyleSheet.create` outside the component
- Keep the JS bridge calls minimal in hot paths (map updates, GPS)

## Navigation (Expo Router)
```
src/
└── app/
    ├── (tabs)/
    │   ├── index.tsx      # Map screen (main tab)
    │   ├── profile.tsx    # Profile + XP + badges
    │   └── avatar.tsx     # Avatar customizer
    ├── login.tsx
    └── register.tsx
```

## Security
- Same JWT/cookie strategy as web — access token in memory, refresh via cookie
- Camera/location permissions: always explain WHY before requesting (`permissionsRequestable` pattern)
- Never log GPS coordinates

## Testing
```bash
npx expo start --ios     # iOS simulator
npx expo start --android # Android emulator
# E2E: Detox (phase 2)
```
