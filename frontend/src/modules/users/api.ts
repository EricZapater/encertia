/**
 * Crides HTTP per al mòdul d'usuaris (User).
 * Utilitza el client Axios centralitzat src/api/client.ts i respecta contracts/user.openapi.yaml.
 */
import apiClient from '@/api/client'
import type {
  BatchCreateUsersRequest,
  BatchCreateUsersResponse,
  CreateUserRequest,
  MessageResponse,
  ResetPasswordRequest,
  UpdateUserRequest,
  UserListParams,
  UserListResponse,
  UserResponse
} from './types'

/**
 * Obté el llistat paginat d'usuaris amb filtres opcionals
 */
export async function listUsers(params?: UserListParams): Promise<UserListResponse> {
  const response = await apiClient.get<UserListResponse>('/users', {
    params
  })
  return response.data
}

/**
 * Consulta un usuari pel seu ID (UUID)
 */
export async function getUserById(id: string): Promise<UserResponse> {
  const response = await apiClient.get<UserResponse>(`/users/${id}`)
  return response.data
}

/**
 * Crea un usuari nou des de l'aplicació
 */
export async function createUser(data: CreateUserRequest): Promise<UserResponse> {
  const response = await apiClient.post<UserResponse>('/users', data)
  return response.data
}

/**
 * Modifica les dades d'un usuari existent
 */
export async function updateUser(id: string, data: UpdateUserRequest): Promise<UserResponse> {
  const response = await apiClient.put<UserResponse>(`/users/${id}`, data)
  return response.data
}

/**
 * Aplica baixa lògica (soft-delete) a un usuari
 */
export async function deleteUser(id: string): Promise<MessageResponse> {
  const response = await apiClient.delete<MessageResponse>(`/users/${id}`)
  return response.data
}

/**
 * Reseteja administrativament la contrasenya d'un usuari
 */
export async function resetUserPassword(
  id: string,
  data: ResetPasswordRequest
): Promise<MessageResponse> {
  const response = await apiClient.post<MessageResponse>(`/users/${id}/password`, data)
  return response.data
}

/**
 * Realitza l'alta massiva d'usuaris (Batch / CSV)
 */
export async function batchCreateUsers(
  data: BatchCreateUsersRequest
): Promise<BatchCreateUsersResponse> {
  const response = await apiClient.post<BatchCreateUsersResponse>('/users/batch', data)
  return response.data
}
