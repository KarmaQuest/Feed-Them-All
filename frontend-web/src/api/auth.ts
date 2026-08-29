// src/api/auth.ts — Auth API calls (login, register, logout, refresh).
import { apiClient, setAccessToken } from './client'

export interface AuthResponse {
  access_token: string
  user: {
    id: string
    email: string
    username: string
    role: string
    roles: string[]
    avatar_config?: Record<string, unknown>
  }
}

/** @deprecated use AuthResponse */
export type LoginResponse = AuthResponse

export async function login(email: string, password: string): Promise<AuthResponse> {
  const res = await apiClient.post<AuthResponse>('/auth/login', { email, password })
  setAccessToken(res.data.access_token)
  return res.data
}

export async function register(
  username: string,
  email: string,
  password: string,
  roles: string[] = ['feeder'],
  avatarConfig?: Record<string, unknown>
): Promise<AuthResponse> {
  const res = await apiClient.post<AuthResponse>('/auth/register', {
    username,
    email,
    password,
    roles,
    avatar_config: avatarConfig,
  })
  setAccessToken(res.data.access_token)
  return res.data
}

export async function logout(): Promise<void> {
  await apiClient.post('/auth/logout')
  setAccessToken(null)
}
