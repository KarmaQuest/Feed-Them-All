// src/api/auth.ts — Auth API calls (login, logout, refresh).
import { apiClient, setAccessToken } from './client'

export interface LoginResponse {
  access_token: string
  user_id: string
  username: string
  role: string
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  const res = await apiClient.post<LoginResponse>('/auth/login', { email, password })
  setAccessToken(res.data.access_token)
  return res.data
}

export async function logout(): Promise<void> {
  await apiClient.post('/auth/logout')
  setAccessToken(null)
}
