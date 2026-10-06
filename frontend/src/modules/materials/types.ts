export type MaterialType = 'document' | 'video'
export type VideoProvider = 'youtube' | 'vimeo' | 'external'

export interface Material {
  id: string
  title: string
  description?: string | null
  materialType: MaterialType
  fileUrl?: string | null
  fileName?: string | null
  fileSizeBytes?: number | null
  mimeType?: string | null
  pageCount: number
  videoUrl?: string | null
  videoProvider?: VideoProvider | string | null
  teacherId: string
  createdAt: string
  updatedAt: string
}

export interface MaterialListResponse {
  items: Material[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface MaterialQueryParams {
  page?: number
  pageSize?: number
  search?: string
  materialType?: MaterialType
}

export interface CreateMaterialRequest {
  title: string
  description?: string
  materialType: MaterialType
  fileUrl?: string
  fileName?: string
  fileSizeBytes?: number
  mimeType?: string
  pageCount?: number
  videoUrl?: string
  videoProvider?: VideoProvider
}

export interface UpdateMaterialRequest {
  title?: string
  description?: string
  fileUrl?: string
  fileName?: string
  fileSizeBytes?: number
  mimeType?: string
  pageCount?: number
  videoUrl?: string
  videoProvider?: VideoProvider
}

export interface UploadFileResponse {
  fileUrl: string
  fileName: string
  fileSizeBytes: number
  mimeType: string
  pageCount: number
}

export interface StudentMaterialView {
  studentId: string
  studentName: string
  studentEmail: string
  viewCount: number
  lastViewedAt: string
}

export interface MaterialViewsReportResponse {
  materialId: string
  totalViews: number
  totalStudentsViewed: number
  studentViews: StudentMaterialView[]
}

export interface RecordMaterialViewResponse {
  success: boolean
  viewCount: number
}
