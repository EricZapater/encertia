import apiClient from '@/api/client'
import type {
  QuizDetail,
  QuizListResponse,
  QuizFilters,
  CreateQuizRequest,
  UpdateQuizRequest,
  DuplicateQuizRequest,
  UploadImageResponse
} from './types'

/**
 * Llista els qüestionaris de l'usuari amb filtres i paginació
 */
export async function listQuizzes(filters?: QuizFilters): Promise<QuizListResponse> {
  const params: Record<string, string | number> = {}
  if (filters?.page) params.page = filters.page
  if (filters?.pageSize) params.pageSize = filters.pageSize
  if (filters?.search && filters.search.trim()) params.search = filters.search.trim()
  if (filters?.status) params.status = filters.status
  if (filters?.tag && filters.tag.trim()) params.tag = filters.tag.trim()

  const response = await apiClient.get<QuizListResponse>('/quizzes', { params })
  return response.data
}

/**
 * Obté el detall complet d'un qüestionari amb les seves preguntes i respostes
 */
export async function getQuizById(id: string): Promise<QuizDetail> {
  const response = await apiClient.get<QuizDetail>(`/quizzes/${id}`)
  return response.data
}

/**
 * Crea un nou qüestionari
 */
export async function createQuiz(payload: CreateQuizRequest): Promise<QuizDetail> {
  const response = await apiClient.post<QuizDetail>('/quizzes', payload)
  return response.data
}

/**
 * Actualitza les metadades i contingut d'un qüestionari
 */
export async function updateQuiz(id: string, payload: UpdateQuizRequest): Promise<QuizDetail> {
  const response = await apiClient.put<QuizDetail>(`/quizzes/${id}`, payload)
  return response.data
}

/**
 * Elimina un qüestionari (soft delete)
 */
export async function deleteQuiz(id: string): Promise<{ message: string }> {
  const response = await apiClient.delete<{ message: string }>(`/quizzes/${id}`)
  return response.data
}

/**
 * Duplica un qüestionari existent amb les seves preguntes (i opcionalment respostes)
 */
export async function duplicateQuiz(
  id: string,
  payload?: DuplicateQuizRequest
): Promise<QuizDetail> {
  const response = await apiClient.post<QuizDetail>(`/quizzes/${id}/duplicate`, payload || {})
  return response.data
}

/**
 * Puja una imatge a Cloudflare R2 (portada o pregunta)
 */
export async function uploadImage(file: File): Promise<UploadImageResponse> {
  const formData = new FormData()
  formData.append('file', file)

  const response = await apiClient.post<UploadImageResponse>('/uploads/images', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
  return response.data
}
