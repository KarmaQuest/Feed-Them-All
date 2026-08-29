// src/api/shop.ts — Requêtes API de la boutique avatar.
//
// Endpoints publics : GET /shop/items
// Endpoints autentifiés : GET /users/me/inventory, POST /shop/items/{id}/purchase
import { apiClient } from './client'

export interface UnlockCondition {
  type: string
  value: number
  action?: string
}

export interface ShopItem {
  id: string
  slug: string
  name: string
  category: 'skin' | 'outfit' | 'accessory'
  price_cents: number
  currency: string
  unlock_condition: UnlockCondition | null
  is_active: boolean
}

export interface InventoryItem {
  item: ShopItem
  acquired_at: string
  source: 'default' | 'quest' | 'purchase'
}

export interface PurchaseResponse {
  client_secret?: string
  granted?: boolean
}

export async function getCatalogue(): Promise<ShopItem[]> {
  const res = await apiClient.get<ShopItem[]>('/shop/items')
  return res.data
}

export async function getInventory(): Promise<InventoryItem[]> {
  const res = await apiClient.get<InventoryItem[]>('/users/me/inventory')
  return res.data
}

export async function purchaseItem(itemId: string): Promise<PurchaseResponse> {
  const res = await apiClient.post<PurchaseResponse>(`/shop/items/${itemId}/purchase`)
  return res.data
}
