import apiClient from '@/api/client'
import type {
  GroupResponse,
  GroupListResponse,
  GroupFilters,
  CreateGroupRequest,
  UpdateGroupRequest,
  GroupStudentListResponse,
  AddStudentsToGroupRequest
} from './types'

/**
 * Llista els grups amb filtres i paginació
 */
export async function listGroups(filters?: GroupFilters): Promise<GroupListResponse> {
  const params: Record<string, string | number> = {}
  if (filters?.page) params.page = filters.page
  if (filters?.pageSize) params.pageSize = filters.pageSize
  if (filters?.search && filters.search.trim()) params.search = filters.search.trim()
  if (filters?.academicYear && filters.academicYear.trim()) params.academicYear = filters.academicYear.trim()
  if (filters?.courseId) params.courseId = filters.courseId

  const response = await apiClient.get<GroupListResponse>('/groups', { params })
  return response.data
}

/**
 * Crea un nou grup
 */
export async function createGroup(payload: CreateGroupRequest): Promise<GroupResponse> {
  const response = await apiClient.post<GroupResponse>('/groups', payload)
  return response.data
}

/**
 * Obté el detall d'un grup per ID
 */
export async function getGroupById(id: string): Promise<GroupResponse> {
  const response = await apiClient.get<GroupResponse>(`/groups/${id}`)
  return response.data
}

/**
 * Actualitza un grup existent
 */
export async function updateGroup(id: string, payload: UpdateGroupRequest): Promise<GroupResponse> {
  const response = await apiClient.put<GroupResponse>(`/groups/${id}`, payload)
  return response.data
}

/**
 * Esborra un grup (soft-delete)
 */
export async function deleteGroup(id: string): Promise<void> {
  await apiClient.delete(`/groups/${id}`)
}

/**
 * Llista els alumnes d'un grup
 */
export async function listGroupStudents(id: string): Promise<GroupStudentListResponse> {
  const response = await apiClient.get<GroupStudentListResponse>(`/groups/${id}/students`)
  return response.data
}

/**
 * Assigna alumnes al grup
 */
export async function addStudentsToGroup(
  id: string,
  payload: AddStudentsToGroupRequest
): Promise<GroupStudentListResponse> {
  const response = await apiClient.post<GroupStudentListResponse>(`/groups/${id}/students`, payload)
  return response.data
}

/**
 * Desassigna un alumne del grup
 */
export async function removeStudentFromGroup(id: string, studentId: string): Promise<void> {
  await apiClient.delete(`/groups/${id}/students/${studentId}`)
}
