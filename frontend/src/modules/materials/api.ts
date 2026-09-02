import apiClient from '@/api/client'
import type {
  Material,
  MaterialListResponse,
  MaterialQueryParams,
  CreateMaterialRequest,
  UpdateMaterialRequest,
  UploadFileResponse,
  RecordMaterialViewResponse,
  MaterialViewsReportResponse
} from './types'

/**
 * Llista els materials didàctics amb opcions de filtre i paginació
 */
export async function listMaterials(params?: MaterialQueryParams): Promise<MaterialListResponse> {
  const queryParams: Record<string, string | number> = {}
  if (params?.page) queryParams.page = params.page
  if (params?.pageSize) queryParams.pageSize = params.pageSize
  if (params?.search && params.search.trim()) queryParams.search = params.search.trim()
  if (params?.materialType) queryParams.materialType = params.materialType

  const response = await apiClient.get<MaterialListResponse>('/materials', { params: queryParams })
  return response.data
}

/**
 * Crea un nou material didàctic (PDF/video)
 */
export async function createMaterial(payload: CreateMaterialRequest): Promise<Material> {
  const response = await apiClient.post<Material>('/materials', payload)
  return response.data
}

/**
 * Pujar un fitxer de document (PDF/Word/PPT)
 */
export async function uploadMaterialFile(file: File): Promise<UploadFileResponse> {
  const formData = new FormData()
  formData.append('file', file)

  const response = await apiClient.post<UploadFileResponse>('/materials/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
  return response.data
}

/**
 * Obté el detall d'un material didàctic per ID
 */
export async function getMaterial(id: string): Promise<Material> {
  const response = await apiClient.get<Material>(`/materials/${id}`)
  return response.data
}

/**
 * Actualitza un material existent
 */
export async function updateMaterial(id: string, payload: UpdateMaterialRequest): Promise<Material> {
  const response = await apiClient.put<Material>(`/materials/${id}`, payload)
  return response.data
}

/**
 * Elimina un material didàctic (soft delete)
 */
export async function deleteMaterial(id: string): Promise<void> {
  await apiClient.delete(`/materials/${id}`)
}

/**
 * Enregistra una visualització per part de l'alumne
 */
export async function recordMaterialView(id: string): Promise<RecordMaterialViewResponse> {
  const response = await apiClient.post<RecordMaterialViewResponse>(`/materials/${id}/views`)
  return response.data
}

/**
 * Obté l'informe d'accessos dels alumnes a un material (Professor/Admin)
 */
export async function getMaterialViewsReport(id: string): Promise<MaterialViewsReportResponse> {
  const response = await apiClient.get<MaterialViewsReportResponse>(`/materials/${id}/views`)
  return response.data
}

/**
 * Llista els materials associats a una unitat didàctica
 */
export async function listUnitMaterials(courseId: string, unitId: string): Promise<Material[]> {
  const response = await apiClient.get<Material[]>(`/courses/${courseId}/units/${unitId}/materials`)
  return response.data
}

/**
 * Vincular un material existent a una unitat didàctica (N:N)
 */
export async function linkMaterialToUnit(
  courseId: string,
  unitId: string,
  materialId: string,
  orderIndex?: number
): Promise<void> {
  await apiClient.post(`/courses/${courseId}/units/${unitId}/materials`, {
    materialId,
    orderIndex
  })
}

/**
 * Desvincular un material d'una unitat didàctica
 */
export async function unlinkMaterialFromUnit(
  courseId: string,
  unitId: string,
  materialId: string
): Promise<void> {
  await apiClient.delete(`/courses/${courseId}/units/${unitId}/materials/${materialId}`)
}
