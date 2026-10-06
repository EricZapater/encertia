import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  Group,
  GroupStudent,
  GroupFilters,
  CreateGroupRequest,
  UpdateGroupRequest
} from './types'
import * as groupsApi from './api'

export const useGroupStore = defineStore('group', () => {
  // State
  const groups = ref<Group[]>([])
  const currentGroup = ref<Group | null>(null)
  const groupStudents = ref<GroupStudent[]>([])

  const currentPage = ref(1)
  const pageSize = ref(10)
  const totalCount = ref(0)
  const totalPages = ref(0)

  const search = ref('')
  const academicYearFilter = ref('')
  const courseIdFilter = ref('')

  const isLoading = ref(false)
  const isSaving = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const groupList = computed(() => groups.value)
  const hasGroups = computed(() => groups.value.length > 0)

  // Actions
  function clearError() {
    error.value = null
  }

  async function fetchGroups() {
    isLoading.value = true
    error.value = null
    try {
      const filters: GroupFilters = {
        page: currentPage.value,
        pageSize: pageSize.value,
        search: search.value,
        academicYear: academicYearFilter.value,
        courseId: courseIdFilter.value
      }
      const response = await groupsApi.listGroups(filters)
      groups.value = response.items || []
      currentPage.value = response.page
      pageSize.value = response.pageSize
      totalCount.value = response.total
      totalPages.value = response.totalPages
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar els grups.'
      groups.value = []
    } finally {
      isLoading.value = false
    }
  }

  function setSearch(newSearch: string) {
    search.value = newSearch
    currentPage.value = 1
    fetchGroups()
  }

  function setAcademicYearFilter(newYear: string) {
    academicYearFilter.value = newYear
    currentPage.value = 1
    fetchGroups()
  }

  function setCourseIdFilter(newCourseId: string) {
    courseIdFilter.value = newCourseId
    currentPage.value = 1
    fetchGroups()
  }

  function setPage(page: number, size?: number) {
    currentPage.value = page
    if (size) pageSize.value = size
    fetchGroups()
  }

  function resetFilters() {
    search.value = ''
    academicYearFilter.value = ''
    courseIdFilter.value = ''
    currentPage.value = 1
    fetchGroups()
  }

  async function fetchGroupDetail(id: string): Promise<Group> {
    isLoading.value = true
    error.value = null
    try {
      const response = await groupsApi.getGroupById(id)
      currentGroup.value = response.data
      return response.data
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar el grup.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function createGroup(payload: CreateGroupRequest): Promise<Group> {
    isSaving.value = true
    error.value = null
    try {
      const response = await groupsApi.createGroup(payload)
      await fetchGroups()
      return response.data
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en crear el grup.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function updateGroup(id: string, payload: UpdateGroupRequest): Promise<Group> {
    isSaving.value = true
    error.value = null
    try {
      const response = await groupsApi.updateGroup(id, payload)
      const updated = response.data
      if (currentGroup.value && currentGroup.value.id === id) {
        currentGroup.value = { ...currentGroup.value, ...updated }
      }
      const idx = groups.value.findIndex((g) => g.id === id)
      if (idx !== -1) {
        groups.value[idx] = { ...groups.value[idx], ...updated }
      }
      return updated
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en actualitzar el grup.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function deleteGroup(id: string): Promise<void> {
    isLoading.value = true
    error.value = null
    try {
      await groupsApi.deleteGroup(id)
      groups.value = groups.value.filter((g) => g.id !== id)
      totalCount.value = Math.max(0, totalCount.value - 1)
      if (currentGroup.value?.id === id) {
        currentGroup.value = null
      }
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en eliminar el grup.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchGroupStudents(id: string): Promise<GroupStudent[]> {
    isLoading.value = true
    error.value = null
    try {
      const response = await groupsApi.listGroupStudents(id)
      groupStudents.value = response.items || []
      return response.items || []
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar els alumnes del grup.'
      groupStudents.value = []
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function addStudentsToGroup(id: string, studentIds: string[]): Promise<GroupStudent[]> {
    isSaving.value = true
    error.value = null
    try {
      const response = await groupsApi.addStudentsToGroup(id, { studentIds })
      groupStudents.value = response.items || []
      const idx = groups.value.findIndex((g) => g.id === id)
      if (idx !== -1) {
        groups.value[idx].studentCount = response.total
      }
      return response.items || []
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en assignar alumnes.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function removeStudentFromGroup(id: string, studentId: string): Promise<void> {
    isSaving.value = true
    error.value = null
    try {
      await groupsApi.removeStudentFromGroup(id, studentId)
      groupStudents.value = groupStudents.value.filter((s) => s.id !== studentId)
      const idx = groups.value.findIndex((g) => g.id === id)
      if (idx !== -1 && groups.value[idx].studentCount) {
        groups.value[idx].studentCount = Math.max(0, groups.value[idx].studentCount! - 1)
      }
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en desassignar l\'alumne.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  return {
    // State
    groups,
    currentGroup,
    groupStudents,
    currentPage,
    pageSize,
    totalCount,
    totalPages,
    search,
    academicYearFilter,
    courseIdFilter,
    isLoading,
    isSaving,
    error,

    // Getters
    groupList,
    hasGroups,

    // Actions
    clearError,
    fetchGroups,
    setSearch,
    setAcademicYearFilter,
    setCourseIdFilter,
    setPage,
    resetFilters,
    fetchGroupDetail,
    createGroup,
    updateGroup,
    deleteGroup,
    fetchGroupStudents,
    addStudentsToGroup,
    removeStudentFromGroup
  }
})
