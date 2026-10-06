import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  Material,
  MaterialType,
  MaterialQueryParams,
  CreateMaterialRequest,
  UpdateMaterialRequest,
  UploadFileResponse,
  MaterialViewsReportResponse
} from './types'
import * as materialsApi from './api'

export const useMaterialStore = defineStore('material', () => {
  // State
  const materials = ref<Material[]>([])
  const currentMaterial = ref<Material | null>(null)
  const currentViewsReport = ref<MaterialViewsReportResponse | null>(null)
  const unitMaterials = ref<Material[]>([])

  const currentPage = ref(1)
  const pageSize = ref(10)
  const totalCount = ref(0)
  const totalPages = ref(0)

  const search = ref('')
  const materialTypeFilter = ref<MaterialType | undefined>(undefined)

  const isLoading = ref(false)
  const isSaving = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const materialList = computed(() => materials.value)
  const hasMaterials = computed(() => materials.value.length > 0)

  // Actions
  function clearError() {
    error.value = null
  }

  function setCurrentMaterial(material: Material | null) {
    currentMaterial.value = material
  }

  async function fetchMaterials(params?: MaterialQueryParams) {
    isLoading.value = true
    error.value = null
    try {
      const queryParams: MaterialQueryParams = {
        page: params?.page ?? currentPage.value,
        pageSize: params?.pageSize ?? pageSize.value,
        search: params?.search ?? search.value,
        materialType: params?.materialType ?? materialTypeFilter.value
      }
      const response = await materialsApi.listMaterials(queryParams)
      materials.value = response.items || []
      currentPage.value = response.page
      pageSize.value = response.pageSize
      totalCount.value = response.total
      totalPages.value = response.totalPages
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar els materials.'
      materials.value = []
    } finally {
      isLoading.value = false
    }
  }

  function setSearch(newSearch: string) {
    search.value = newSearch
    currentPage.value = 1
    fetchMaterials()
  }

  function setTypeFilter(type?: MaterialType) {
    materialTypeFilter.value = type
    currentPage.value = 1
    fetchMaterials()
  }

  function setPage(page: number, size?: number) {
    currentPage.value = page
    if (size) pageSize.value = size
    fetchMaterials()
  }

  function resetFilters() {
    search.value = ''
    materialTypeFilter.value = undefined
    currentPage.value = 1
    fetchMaterials()
  }

  async function fetchMaterialDetail(id: string): Promise<Material> {
    isLoading.value = true
    error.value = null
    try {
      const detail = await materialsApi.getMaterial(id)
      currentMaterial.value = detail
      return detail
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar el detall del material.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function uploadFile(file: File): Promise<UploadFileResponse> {
    isSaving.value = true
    error.value = null
    try {
      return await materialsApi.uploadMaterialFile(file)
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en pujar el fitxer.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function createMaterial(payload: CreateMaterialRequest): Promise<Material> {
    isSaving.value = true
    error.value = null
    try {
      const created = await materialsApi.createMaterial(payload)
      await fetchMaterials()
      return created
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en crear el material.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function updateMaterial(id: string, payload: UpdateMaterialRequest): Promise<Material> {
    isSaving.value = true
    error.value = null
    try {
      const updated = await materialsApi.updateMaterial(id, payload)
      if (currentMaterial.value && currentMaterial.value.id === id) {
        currentMaterial.value = { ...currentMaterial.value, ...updated }
      }
      const idx = materials.value.findIndex((m) => m.id === id)
      if (idx !== -1) {
        materials.value[idx] = { ...materials.value[idx], ...updated }
      }
      return updated
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en actualitzar el material.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function deleteMaterial(id: string): Promise<void> {
    isLoading.value = true
    error.value = null
    try {
      await materialsApi.deleteMaterial(id)
      materials.value = materials.value.filter((m) => m.id !== id)
      totalCount.value = Math.max(0, totalCount.value - 1)
      if (currentMaterial.value?.id === id) {
        currentMaterial.value = null
      }
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en eliminar el material.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function recordView(id: string) {
    try {
      return await materialsApi.recordMaterialView(id)
    } catch (err: any) {
      console.error('Failed to record material view:', err)
    }
  }

  async function fetchViewsReport(id: string): Promise<MaterialViewsReportResponse> {
    isLoading.value = true
    error.value = null
    try {
      const report = await materialsApi.getMaterialViewsReport(id)
      currentViewsReport.value = report
      return report
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar l\'informe de visualitzacions.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchUnitMaterials(courseId: string, unitId: string): Promise<Material[]> {
    isLoading.value = true
    error.value = null
    try {
      const items = await materialsApi.listUnitMaterials(courseId, unitId)
      unitMaterials.value = items
      return items
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar els materials de la unitat.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function linkMaterialToUnit(
    courseId: string,
    unitId: string,
    materialId: string,
    orderIndex?: number
  ): Promise<void> {
    isSaving.value = true
    error.value = null
    try {
      await materialsApi.linkMaterialToUnit(courseId, unitId, materialId, orderIndex)
      await fetchUnitMaterials(courseId, unitId)
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en vincular el material a la unitat.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function unlinkMaterialFromUnit(
    courseId: string,
    unitId: string,
    materialId: string
  ): Promise<void> {
    isSaving.value = true
    error.value = null
    try {
      await materialsApi.unlinkMaterialFromUnit(courseId, unitId, materialId)
      unitMaterials.value = unitMaterials.value.filter((m) => m.id !== materialId)
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en desvincular el material.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  return {
    // State
    materials,
    currentMaterial,
    currentViewsReport,
    unitMaterials,
    currentPage,
    pageSize,
    totalCount,
    totalPages,
    search,
    materialTypeFilter,
    isLoading,
    isSaving,
    error,

    // Getters
    materialList,
    hasMaterials,

    // Actions
    clearError,
    setCurrentMaterial,
    fetchMaterials,
    setSearch,
    setTypeFilter,
    setPage,
    resetFilters,
    fetchMaterialDetail,
    uploadFile,
    createMaterial,
    updateMaterial,
    deleteMaterial,
    recordView,
    fetchViewsReport,
    fetchUnitMaterials,
    linkMaterialToUnit,
    unlinkMaterialFromUnit
  }
})
