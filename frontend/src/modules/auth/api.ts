import apiClient from '@/api/client'
import type {
  AuthResponse,
  LoginRequest,
  LogoutRequest,
  MessageResponse,
  RefreshTokenRequest,
  RegisterRequest,
  TokenPair,
  UserResponse
} from './types'

/**
 * Registra un nou usuari a Encertia.
 * POST /auth/register
 */
export async function register(payload: RegisterRequest): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>('/auth/register', payload)
  return response.data
}

/**
 * Inicia sessió amb correu i contrasenya.
 * POST /auth/login
 */
export async function login(payload: LoginRequest): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>('/auth/login', payload)
  return response.data
}

/**
 * Renova el parell de tokens JWT emprant el refresh token.
 * POST /auth/refresh
 */
export async function refreshToken(payload: RefreshTokenRequest): Promise<TokenPair> {
  const response = await apiClient.post<TokenPair>('/auth/refresh', payload)
  return response.data
}

/**
 * Tanca la sessió de l'usuari i invalida el refresh token.
 * POST /auth/logout
 */
export async function logout(payload?: LogoutRequest): Promise<MessageResponse> {
  const response = await apiClient.post<MessageResponse>('/auth/logout', payload || {})
  return response.data
}

/**
 * Obté el perfil de l'usuari autenticat actual.
 * GET /auth/me
 */
export async function getCurrentUser(): Promise<UserResponse> {
  const response = await apiClient.get<UserResponse>('/auth/me')
  return response.data
}
