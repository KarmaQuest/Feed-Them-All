// src/hooks/useGeolocation.ts — Hook de géolocalisation browser.
//
// Retourne la position GPS de l'utilisateur via navigator.geolocation.
// - Demande la position au montage.
// - Met à jour useMapStore en cas de succès.
// - Expose error (string) si la géoloc est refusée ou indisponible.
import { useState, useEffect } from 'react'
import { useMapStore } from '../store/map'

interface GeolocationState {
  loading: boolean
  error: string | null
}

// Ville par défaut si la géolocalisation est refusée : Hô-Chi-Minh-Ville
export const DEFAULT_LAT = 10.7769
export const DEFAULT_LON = 106.7009
export const DEFAULT_ZOOM = 13

export function useGeolocation(): GeolocationState {
  const setUserPosition = useMapStore((s) => s.setUserPosition)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!navigator.geolocation) {
      setError('La géolocalisation n\'est pas supportée par ce navigateur.')
      setLoading(false)
      return
    }

    const watchId = navigator.geolocation.watchPosition(
      (pos) => {
        setUserPosition(pos.coords.latitude, pos.coords.longitude)
        setLoading(false)
        setError(null)
      },
      (err) => {
        if (err.code === GeolocationPositionError.PERMISSION_DENIED) {
          setError('Géolocalisation refusée — carte centrée sur Paris.')
        } else {
          setError('Impossible d\'obtenir la position GPS.')
        }
        setLoading(false)
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 5000 },
    )

    return () => navigator.geolocation.clearWatch(watchId)
  }, [setUserPosition])

  return { loading, error }
}
