import { describe, it, expect, vi, beforeEach } from 'vitest'
import apiClient from '@/api/client'
import {
  listMaterials,
  createMaterial,
  uploadMaterialFile,
  getMaterial,
  updateMaterial,
  deleteMaterial,
  recordMaterialView,
  getMaterialViewsReport,
  listUnitMaterials,
  linkMaterialToUnit,
  unlinkMaterialFromUnit
} from '../api'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

describe('Materials API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('listMaterials calls GET /materials with params', async () => {
    const mockResponse = {
      data: {
        items: [],
        total: 0,
        page: 1,
        pageSize: 10,
        totalPages: 0
      }
    }
    vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

    const result = await listMaterials({ page: 1, pageSize: 10, search: 'PDF', materialType: 'document' })

    expect(apiClient.get).toHaveBeenCalledWith('/materials', {
      params: {
        page: 1,
        pageSize: 10,
        search: 'PDF',
        materialType: 'document'
      }
    })
    expect(result).toEqual(mockResponse.data)
  })

  it('createMaterial calls POST /materials with payload', async () => {
    const payload = { title: 'Material 1', materialType: 'document' as const, fileUrl: 'https://example.com/doc.pdf' }
    const mockCreated = { id: 'm-1', ...payload, teacherId: 't-1', pageCount: 5, createdAt: '', updatedAt: '' }
    vi.mocked(apiClient.post).mockResolvedValue({ data: mockCreated })

    const result = await createMaterial(payload)
    expect(apiClient.post).toHaveBeenCalledWith('/materials', payload)
    expect(result).toEqual(mockCreated)
  })

  it('uploadMaterialFile calls POST /materials/upload with FormData', async () => {
    const mockResponse = {
      data: {
        fileUrl: 'https://example.com/doc.pdf',
        fileName: 'doc.pdf',
        fileSizeBytes: 1024,
        mimeType: 'application/pdf',
        pageCount: 10
      }
    }
    vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

    const file = new File(['dummy'], 'doc.pdf', { type: 'application/pdf' })
    const result = await uploadMaterialFile(file)

    expect(apiClient.post).toHaveBeenCalledWith('/materials/upload', expect.any(FormData), {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    expect(result).toEqual(mockResponse.data)
  })

  it('getMaterial calls GET /materials/:id', async () => {
    const mockMat = { id: 'm-1', title: 'Material 1' }
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockMat })

    const result = await getMaterial('m-1')
    expect(apiClient.get).toHaveBeenCalledWith('/materials/m-1')
    expect(result).toEqual(mockMat)
  })

  it('updateMaterial calls PUT /materials/:id', async () => {
    const payload = { title: 'Material Actualitzat' }
    const mockUpdated = { id: 'm-1', title: 'Material Actualitzat' }
    vi.mocked(apiClient.put).mockResolvedValue({ data: mockUpdated })

    const result = await updateMaterial('m-1', payload)
    expect(apiClient.put).toHaveBeenCalledWith('/materials/m-1', payload)
    expect(result).toEqual(mockUpdated)
  })

  it('deleteMaterial calls DELETE /materials/:id', async () => {
    vi.mocked(apiClient.delete).mockResolvedValue({})

    await deleteMaterial('m-1')
    expect(apiClient.delete).toHaveBeenCalledWith('/materials/m-1')
  })

  it('recordMaterialView calls POST /materials/:id/views', async () => {
    vi.mocked(apiClient.post).mockResolvedValue({ data: { success: true, viewCount: 1 } })

    const result = await recordMaterialView('m-1')
    expect(apiClient.post).toHaveBeenCalledWith('/materials/m-1/views')
    expect(result).toEqual({ success: true, viewCount: 1 })
  })

  it('getMaterialViewsReport calls GET /materials/:id/views', async () => {
    const mockReport = { materialId: 'm-1', totalViews: 5, totalStudentsViewed: 2, studentViews: [] }
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockReport })

    const result = await getMaterialViewsReport('m-1')
    expect(apiClient.get).toHaveBeenCalledWith('/materials/m-1/views')
    expect(result).toEqual(mockReport)
  })

  it('listUnitMaterials, linkMaterialToUnit, and unlinkMaterialFromUnit call correct endpoints', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: [] })
    vi.mocked(apiClient.post).mockResolvedValue({})
    vi.mocked(apiClient.delete).mockResolvedValue({})

    await listUnitMaterials('c-1', 'u-1')
    expect(apiClient.get).toHaveBeenCalledWith('/courses/c-1/units/u-1/materials')

    await linkMaterialToUnit('c-1', 'u-1', 'm-1', 0)
    expect(apiClient.post).toHaveBeenCalledWith('/courses/c-1/units/u-1/materials', { materialId: 'm-1', orderIndex: 0 })

    await unlinkMaterialFromUnit('c-1', 'u-1', 'm-1')
    expect(apiClient.delete).toHaveBeenCalledWith('/courses/c-1/units/u-1/materials/m-1')
  })
})
