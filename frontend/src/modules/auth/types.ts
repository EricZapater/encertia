/**
 * Tipus TypeScript per al mòdul d'autenticació (Auth).
 * Derivats fidelment del contracte OpenAPI contracts/auth.openapi.yaml
 */

/** Rol de l'usuari dins de la plataforma */
export type UserRole = 'admin' | 'teacher' | 'student'

export type SupportedLanguage = 'ca' | 'es' | 'en'

/** Entitat d'usuari */
export interface User {
  id: string
  email: string
  firstName: string
  lastName: string
  role: UserRole
  language?: SupportedLanguage
  isActive?: boolean
  createdAt: string
  updatedAt?: string
}

/** Petició de registre d'un nou usuari */
export interface RegisterRequest {
  email: string
  password: string
  firstName: string
  lastName: string
  role: UserRole
  language?: SupportedLanguage
}

/** Petició d'inici de sessió */
export interface LoginRequest {
  email: string
  password: string
}

/** Petició de renovació de tokens */
export interface RefreshTokenRequest {
  refreshToken: string
}

/** Petició opcional de tancament de sessió */
export interface LogoutRequest {
  refreshToken?: string
}

/** Parell de tokens JWT i metadades */
export interface TokenPair {
  accessToken: string
  refreshToken: string
  tokenType: string
  expiresIn: number
}

/** Resposta d'autenticació amb usuari i tokens */
export interface AuthResponse {
  user: User
  tokens: TokenPair
}

/** Resposta d'obtenció de l'usuari actual */
export interface UserResponse {
  user: User
}

/** Resposta genèrica amb missatge informatiu */
export interface MessageResponse {
  message: string
}

/** Detall d'un error retornat per l'API */
export interface ErrorDetail {
  code: string
  message: string
  details?: Record<string, unknown>
}

/** Resposta unificada d'error */
export interface ErrorResponse {
  error: ErrorDetail
}
