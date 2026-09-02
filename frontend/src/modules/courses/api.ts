import apiClient from '@/api/client'
import type {
  Course,
  CourseDetail,
  CourseListResponse,
  CourseFilters,
  CreateCourseRequest,
  UpdateCourseRequest,
  CourseStudentsResponse,
  EnrollStudentsRequest,
  CourseUnit,
  CourseUnitDetail,
  CreateCourseUnitRequest,
  UpdateCourseUnitRequest,
  ScriptBlock,
  CreateScriptBlockRequest
} from './types'

/**
 * Llista els cursos amb filtres i paginació
 */
export async function listCourses(filters?: CourseFilters): Promise<CourseListResponse> {
  const params: Record<string, string | number> = {}
  if (filters?.page) params.page = filters.page
  if (filters?.pageSize) params.pageSize = filters.pageSize
  if (filters?.search && filters.search.trim()) params.search = filters.search.trim()
  if (filters?.status) params.status = filters.status

  const response = await apiClient.get<CourseListResponse>('/courses', { params })
  return response.data
}

/**
 * Crea un nou curs (professors i admins)
 */
export async function createCourse(payload: CreateCourseRequest): Promise<Course> {
  const response = await apiClient.post<Course>('/courses', payload)
  return response.data
}

/**
 * Obté el detall d'un curs amb les seves unitats
 */
export async function getCourseById(id: string): Promise<CourseDetail> {
  const response = await apiClient.get<CourseDetail>(`/courses/${id}`)
  return response.data
}

/**
 * Actualitza les dades d'un curs
 */
export async function updateCourse(id: string, payload: UpdateCourseRequest): Promise<Course> {
  const response = await apiClient.put<Course>(`/courses/${id}`, payload)
  return response.data
}

/**
 * Elimina un curs (soft delete)
 */
export async function deleteCourse(id: string): Promise<void> {
  await apiClient.delete(`/courses/${id}`)
}

/**
 * Llista els alumnes matriculats a un curs
 */
export async function getCourseStudents(id: string): Promise<CourseStudentsResponse> {
  const response = await apiClient.get<CourseStudentsResponse>(`/courses/${id}/students`)
  return response.data
}

/**
 * Matricula alumnes a un curs
 */
export async function enrollStudents(
  id: string,
  payload: EnrollStudentsRequest
): Promise<CourseStudentsResponse> {
  const response = await apiClient.post<CourseStudentsResponse>(`/courses/${id}/students`, payload)
  return response.data
}

/**
 * Desmatricula un alumne d'un curs
 */
export async function unenrollStudent(id: string, studentId: string): Promise<void> {
  await apiClient.delete(`/courses/${id}/students/${studentId}`)
}

/**
 * Llista les unitats didàctiques d'un curs
 */
export async function listCourseUnits(courseId: string): Promise<CourseUnit[]> {
  const response = await apiClient.get<CourseUnit[]>(`/courses/${courseId}/units`)
  return response.data
}

/**
 * Crea una unitat didàctica a un curs
 */
export async function createCourseUnit(
  courseId: string,
  payload: CreateCourseUnitRequest
): Promise<CourseUnit> {
  const response = await apiClient.post<CourseUnit>(`/courses/${courseId}/units`, payload)
  return response.data
}

/**
 * Reordena les unitats didàctiques d'un curs
 */
export async function reorderCourseUnits(
  courseId: string,
  unitIds: string[]
): Promise<CourseUnit[]> {
  const response = await apiClient.put<CourseUnit[]>(`/courses/${courseId}/units/reorder`, unitIds)
  return response.data
}

/**
 * Obté el detall d'una unitat didàctica amb qüestionaris i guió
 */
export async function getCourseUnit(
  courseId: string,
  unitId: string
): Promise<CourseUnitDetail> {
  const response = await apiClient.get<CourseUnitDetail>(`/courses/${courseId}/units/${unitId}`)
  return response.data
}

/**
 * Actualitza una unitat didàctica
 */
export async function updateCourseUnit(
  courseId: string,
  unitId: string,
  payload: UpdateCourseUnitRequest
): Promise<CourseUnit> {
  const response = await apiClient.put<CourseUnit>(
    `/courses/${courseId}/units/${unitId}`,
    payload
  )
  return response.data
}

/**
 * Elimina una unitat didàctica (soft delete)
 */
export async function deleteCourseUnit(courseId: string, unitId: string): Promise<void> {
  await apiClient.delete(`/courses/${courseId}/units/${unitId}`)
}

/**
 * Vincula un qüestionari a una unitat didàctica
 */
export async function linkQuizToUnit(
  courseId: string,
  unitId: string,
  quizId: string
): Promise<void> {
  await apiClient.post(`/courses/${courseId}/units/${unitId}/quizzes`, { quizId })
}

/**
 * Desvincula un qüestionari d'una unitat didàctica
 */
export async function unlinkQuizFromUnit(
  courseId: string,
  unitId: string,
  quizId: string
): Promise<void> {
  await apiClient.delete(`/courses/${courseId}/units/${unitId}/quizzes/${quizId}`)
}

/**
 * Obté el guió de classe d'una unitat
 */
export async function getUnitScript(courseId: string, unitId: string): Promise<ScriptBlock[]> {
  const response = await apiClient.get<ScriptBlock[]>(
    `/courses/${courseId}/units/${unitId}/script`
  )
  return response.data
}

/**
 * Desar / actualitza la seqüència de guió de classe
 */
export async function updateUnitScript(
  courseId: string,
  unitId: string,
  blocks: CreateScriptBlockRequest[]
): Promise<ScriptBlock[]> {
  const response = await apiClient.put<ScriptBlock[]>(
    `/courses/${courseId}/units/${unitId}/script`,
    blocks
  )
  return response.data
}
