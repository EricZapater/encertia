import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMaterialStore } from '../store'
import * as materialsApi from '../api'

vi.mock('../api')

describe('useMaterialStore Pinia Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('initial state is correct', () => {
    const store = useMaterialStore()
    expect(store.materials).toEqual([])
    expect(store.currentMaterial).toBeNull()
    expect(store.currentViewsReport).toBeNull()
    expect(store.unitMaterials).toEqual([])
    expect(store.currentPage).toBe(1)
    expect(store.isLoading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('fetchMaterials updates materials and pagination in store', async () => {
    const mockList = {
      items: [
        {
          id: 'm-1',
          title: 'Apunts Tema 1',
          materialType: 'document' as const,
          fileUrl: 'https://example.com/tema1.pdf',
          pageCount: 10,
          teacherId: 't-1',
          createdAt: '',
          updatedAt: ''
        }
      ],
      total: 1,
      page: 1,
      pageSize: 10,
      totalPages: 1
    }
    vi.mocked(materialsApi.listMaterials).mockResolvedValue(mockList)

    const store = useMaterialStore()
    await store.fetchMaterials()

    expect(store.materials).toEqual(mockList.items)
    expect(store.totalCount).toBe(1)
    expect(store.isLoading).toBe(false)
  })

  it('createMaterial calls api and refreshes list', async () => {
    const store = useMaterialStore()
    const newMat = {
      id: 'm-2',
      title: 'Vídeo Explicatiu',
      materialType: 'video' as const,
      videoUrl: 'https://youtube.com/watch?v=123',
      videoProvider: 'youtube',
      pageCount: 0,
      teacherId: 't-1',
      createdAt: '',
      updatedAt: ''
    }
    vi.mocked(materialsApi.createMaterial).mockResolvedValue(newMat)
    vi.mocked(materialsApi.listMaterials).mockResolvedValue({
      items: [newMat],
      total: 1,
      page: 1,
      pageSize: 10,
      totalPages: 1
    })

    const result = await store.createMaterial({
      title: 'Vídeo Explicatiu',
      materialType: 'video',
      videoUrl: 'https://youtube.com/watch?v=123'
    })

    expect(materialsApi.createMaterial).toHaveBeenCalledWith({
      title: 'Vídeo Explicatiu',
      materialType: 'video',
      videoUrl: 'https://youtube.com/watch?v=123'
    })
    expect(result).toEqual(newMat)
    expect(store.materials).toHaveLength(1)
  })

  it('deleteMaterial removes material from store state', async () => {
    const store = useMaterialStore()
    store.materials = [
      {
        id: 'm-1',
        title: 'Document PDF',
        materialType: 'document',
        pageCount: 5,
        teacherId: 't-1',
        createdAt: '',
        updatedAt: ''
      }
    ]
    store.totalCount = 1

    vi.mocked(materialsApi.deleteMaterial).mockResolvedValue()

    await store.deleteMaterial('m-1')

    expect(materialsApi.deleteMaterial).toHaveBeenCalledWith('m-1')
    expect(store.materials).toEqual([])
    expect(store.totalCount).toBe(0)
  })

  it('fetchViewsReport updates currentViewsReport in store', async () => {
    const store = useMaterialStore()
    const mockReport = {
      materialId: 'm-1',
      totalViews: 10,
      totalStudentsViewed: 3,
      studentViews: []
    }
    vi.mocked(materialsApi.getMaterialViewsReport).mockResolvedValue(mockReport)

    await store.fetchViewsReport('m-1')

    expect(store.currentViewsReport).toEqual(mockReport)
  })
})
