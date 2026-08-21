/**
 * Tipus TypeScript per al mòdul d'usuaris (User).
 * Transcrits fidelment del contracte OpenAPI contracts/user.openapi.yaml
 */

/** Rol de l'usuari a la plataforma Encertia */
export type UserRole = 'admin' | 'teacher' | 'student'

/** Estat de filtratge dels usuaris */
export type UserStatusFilter = 'active' | 'inactive' | 'all'

/** Entitat d'usuari */
export interface User {
  id: string
  email: string
  firstName: string
  lastName: string
  role: UserRole
  isActive: boolean
  createdAt: string
  updatedAt?: string
}

/** Petició de creació individual d'usuari */
export interface CreateUserRequest {
  email: string
  password: string
  firstName: string
  lastName: string
  role: UserRole
}

/** Petició de modificació d'un usuari */
export interface UpdateUserRequest {
  email?: string
  firstName?: string
  lastName?: string
  role?: UserRole
  isActive?: boolean
}

/** Petició de reseteig administratiu de contrasenya */
export interface ResetPasswordRequest {
  newPassword: string
}

/** Element d'usuari per a l'alta massiva (Batch / CSV) */
export interface BatchUserItem {
  email: string
  firstName: string
  lastName: string
  role: UserRole
  password?: string
}

/** Petició d'alta massiva d'usuaris */
export interface BatchCreateUsersRequest {
  users: BatchUserItem[]
}

/** Error associat a una fila d'importació massiva */
export interface BatchItemError {
  row: number
  email: string
  error: string
}

/** Resposta de l'alta massiva d'usuaris */
export interface BatchCreateUsersResponse {
  totalRequested: number
  createdCount: number
  failedCount: number
  createdUsers: User[]
  errors: BatchItemError[]
}

/** Metadades de paginació */
export interface PaginationMetadata {
  page: number
  pageSize: number
  totalCount: number
  totalPages: number
}

/** Resposta paginada del llistat d'usuaris */
export interface UserListResponse {
  items: User[]
  pagination: PaginationMetadata
}

/** Paràmetres per a la consulta de llistat d'usuaris */
export interface UserListParams {
  page?: number
  pageSize?: number
  search?: string
  role?: UserRole
  status?: UserStatusFilter
}

/** Resposta que conté un únic usuari */
export interface UserResponse {
  user: User
}

/** Resposta genèrica amb missatge informatiu */
export interface MessageResponse {
  message: string
}

/** Detall d'error del servidor */
export interface ErrorDetail {
  code: string
  message: string
  details?: Record<string, unknown>
}

/** Resposta d'error de l'API */
export interface ErrorResponse {
  error: ErrorDetail
}
