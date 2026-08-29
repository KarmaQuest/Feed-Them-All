// src/utils/roles.ts — Utilitaires d'affichage des rôles utilisateur.

export const ROLE_LABELS: Record<string, string> = {
  feeder: 'Feeder',
  giver: 'Giver',
  admin: 'Admin',
  association: 'Association',
}

export function formatRoles(roles: string[] | undefined, fallback?: string): string {
  if (!roles || roles.length === 0) return fallback ?? '—'
  return roles.map(r => ROLE_LABELS[r] ?? r).join(', ')
}
