import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  Course,
  CourseDetail,
  CourseStatus,
  CourseFilters,
  CreateCourseRequest,
  UpdateCourseRequest,
  EnrolledStudent,
  CourseUnit,
  CourseUnitDetail,
  CreateCourseUnitRequest,
  UpdateCourseUnitRequest,
  ScriptBlock,
  CreateScriptBlockRequest
} from './types'
import * as coursesApi from './api'

export const useCourseStore = defineStore('course', () => {
  // State
  const courses = ref<Course[]>([])
  const currentCourse = ref<CourseDetail | null>(null)
  const currentUnit = ref<CourseUnitDetail | null>(null)
  const enrolledStudents = ref<EnrolledStudent[]>([])

  const currentPage = ref(1)
  const pageSize = ref(10)
  const totalCount = ref(0)
  const totalPages = ref(0)

  const search = ref('')
  const statusFilter = ref<CourseStatus | undefined>(undefined)

  const isLoading = ref(false)
  const isSaving = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const courseList = computed(() => courses.value)
  const hasCourses = computed(() => courses.value.length > 0)

  // Actions
  function clearError() {
    error.value = null
  }

  function setCurrentCourse(course: CourseDetail | null) {
    currentCourse.value = course
  }

  function setCurrentUnit(unit: CourseUnitDetail | null) {
    currentUnit.value = unit
  }

  async function fetchCourses() {
    isLoading.value = true
    error.value = null
    try {
      const filters: CourseFilters = {
        page: currentPage.value,
        pageSize: pageSize.value,
        search: search.value,
        status: statusFilter.value
      }
      const response = await coursesApi.listCourses(filters)
      courses.value = response.items || []
      currentPage.value = response.page
      pageSize.value = response.pageSize
      totalCount.value = response.total
      totalPages.value = response.totalPages
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar els cursos.'
      courses.value = []
    } finally {
      isLoading.value = false
    }
  }

  function setSearch(newSearch: string) {
    search.value = newSearch
    currentPage.value = 1
    fetchCourses()
  }

  function setStatusFilter(newStatus?: CourseStatus) {
    statusFilter.value = newStatus
    currentPage.value = 1
    fetchCourses()
  }

  function setPage(page: number, size?: number) {
    currentPage.value = page
    if (size) pageSize.value = size
    fetchCourses()
  }

  function resetFilters() {
    search.value = ''
    statusFilter.value = undefined
    currentPage.value = 1
    fetchCourses()
  }

  async function fetchCourseDetail(id: string): Promise<CourseDetail> {
    isLoading.value = true
    error.value = null
    try {
      const detail = await coursesApi.getCourseById(id)
      currentCourse.value = detail
      return detail
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar el curs.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function createCourse(payload: CreateCourseRequest): Promise<Course> {
    isSaving.value = true
    error.value = null
    try {
      const created = await coursesApi.createCourse(payload)
      await fetchCourses()
      return created
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en crear el curs.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function updateCourse(id: string, payload: UpdateCourseRequest): Promise<Course> {
    isSaving.value = true
    error.value = null
    try {
      const updated = await coursesApi.updateCourse(id, payload)
      if (currentCourse.value && currentCourse.value.id === id) {
        currentCourse.value = { ...currentCourse.value, ...updated }
      }
      const idx = courses.value.findIndex((c) => c.id === id)
      if (idx !== -1) {
        courses.value[idx] = { ...courses.value[idx], ...updated }
      }
      return updated
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en actualitzar el curs.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function deleteCourse(id: string): Promise<void> {
    isLoading.value = true
    error.value = null
    try {
      await coursesApi.deleteCourse(id)
      courses.value = courses.value.filter((c) => c.id !== id)
      totalCount.value = Math.max(0, totalCount.value - 1)
      if (currentCourse.value?.id === id) {
        currentCourse.value = null
      }
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en eliminar el curs.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchCourseStudents(id: string): Promise<EnrolledStudent[]> {
    isLoading.value = true
    error.value = null
    try {
      const response = await coursesApi.getCourseStudents(id)
      enrolledStudents.value = response.students || []
      return response.students
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar els alumnes del curs.'
      enrolledStudents.value = []
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function enrollStudents(id: string, studentIds: string[]): Promise<EnrolledStudent[]> {
    isSaving.value = true
    error.value = null
    try {
      const response = await coursesApi.enrollStudents(id, { studentIds })
      enrolledStudents.value = response.students || []
      if (currentCourse.value && currentCourse.value.id === id) {
        currentCourse.value.enrolledStudentsCount = response.total
      }
      return response.students
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en matricular alumnes.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function unenrollStudent(id: string, studentId: string): Promise<void> {
    isSaving.value = true
    error.value = null
    try {
      await coursesApi.unenrollStudent(id, studentId)
      enrolledStudents.value = enrolledStudents.value.filter((s) => s.id !== studentId)
      if (currentCourse.value && currentCourse.value.id === id && currentCourse.value.enrolledStudentsCount) {
        currentCourse.value.enrolledStudentsCount = Math.max(0, currentCourse.value.enrolledStudentsCount - 1)
      }
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en desmatricular l\'alumne.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function fetchCourseUnits(courseId: string): Promise<CourseUnit[]> {
    isLoading.value = true
    error.value = null
    try {
      const units = await coursesApi.listCourseUnits(courseId)
      if (currentCourse.value && currentCourse.value.id === courseId) {
        currentCourse.value.units = units
        currentCourse.value.unitsCount = units.length
      }
      return units
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar les unitats del curs.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function createCourseUnit(
    courseId: string,
    payload: CreateCourseUnitRequest
  ): Promise<CourseUnit> {
    isSaving.value = true
    error.value = null
    try {
      const created = await coursesApi.createCourseUnit(courseId, payload)
      if (currentCourse.value && currentCourse.value.id === courseId) {
        if (!currentCourse.value.units) currentCourse.value.units = []
        currentCourse.value.units.push(created)
        currentCourse.value.unitsCount = currentCourse.value.units.length
      }
      return created
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en crear la unitat didàctica.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function reorderCourseUnits(courseId: string, unitIds: string[]): Promise<CourseUnit[]> {
    isSaving.value = true
    error.value = null
    try {
      const reordered = await coursesApi.reorderCourseUnits(courseId, unitIds)
      if (currentCourse.value && currentCourse.value.id === courseId) {
        currentCourse.value.units = reordered
      }
      return reordered
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en reordenar les unitats.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function fetchCourseUnit(courseId: string, unitId: string): Promise<CourseUnitDetail> {
    isLoading.value = true
    error.value = null
    try {
      const unit = await coursesApi.getCourseUnit(courseId, unitId)
      currentUnit.value = unit
      return unit
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar la unitat didàctica.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function updateCourseUnit(
    courseId: string,
    unitId: string,
    payload: UpdateCourseUnitRequest
  ): Promise<CourseUnit> {
    isSaving.value = true
    error.value = null
    try {
      const updated = await coursesApi.updateCourseUnit(courseId, unitId, payload)
      if (currentUnit.value && currentUnit.value.id === unitId) {
        currentUnit.value = { ...currentUnit.value, ...updated }
      }
      if (currentCourse.value && currentCourse.value.id === courseId && currentCourse.value.units) {
        const idx = currentCourse.value.units.findIndex((u) => u.id === unitId)
        if (idx !== -1) {
          currentCourse.value.units[idx] = { ...currentCourse.value.units[idx], ...updated }
        }
      }
      return updated
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en actualitzar la unitat.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function deleteCourseUnit(courseId: string, unitId: string): Promise<void> {
    isLoading.value = true
    error.value = null
    try {
      await coursesApi.deleteCourseUnit(courseId, unitId)
      if (currentCourse.value && currentCourse.value.units) {
        currentCourse.value.units = currentCourse.value.units.filter((u) => u.id !== unitId)
        currentCourse.value.unitsCount = currentCourse.value.units.length
      }
      if (currentUnit.value?.id === unitId) {
        currentUnit.value = null
      }
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en eliminar la unitat.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function linkQuizToUnit(
    courseId: string,
    unitId: string,
    quizId: string
  ): Promise<void> {
    isSaving.value = true
    error.value = null
    try {
      await coursesApi.linkQuizToUnit(courseId, unitId, quizId)
      await fetchCourseUnit(courseId, unitId)
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en vincular el qüestionari.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function unlinkQuizFromUnit(
    courseId: string,
    unitId: string,
    quizId: string
  ): Promise<void> {
    isSaving.value = true
    error.value = null
    try {
      await coursesApi.unlinkQuizFromUnit(courseId, unitId, quizId)
      if (currentUnit.value && currentUnit.value.id === unitId && currentUnit.value.linkedQuizzes) {
        currentUnit.value.linkedQuizzes = currentUnit.value.linkedQuizzes.filter((q) => q.id !== quizId)
      }
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en desvincular el qüestionari.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function fetchUnitScript(courseId: string, unitId: string): Promise<ScriptBlock[]> {
    isLoading.value = true
    error.value = null
    try {
      const blocks = await coursesApi.getUnitScript(courseId, unitId)
      if (currentUnit.value && currentUnit.value.id === unitId) {
        currentUnit.value.scriptBlocks = blocks
      }
      return blocks
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en carregar el guió de classe.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function updateUnitScript(
    courseId: string,
    unitId: string,
    blocks: CreateScriptBlockRequest[]
  ): Promise<ScriptBlock[]> {
    isSaving.value = true
    error.value = null
    try {
      const updatedBlocks = await coursesApi.updateUnitScript(courseId, unitId, blocks)
      if (currentUnit.value && currentUnit.value.id === unitId) {
        currentUnit.value.scriptBlocks = updatedBlocks
      }
      return updatedBlocks
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Error en actualitzar el guió de classe.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  return {
    // State
    courses,
    currentCourse,
    currentUnit,
    enrolledStudents,
    currentPage,
    pageSize,
    totalCount,
    totalPages,
    search,
    statusFilter,
    isLoading,
    isSaving,
    error,

    // Getters
    courseList,
    hasCourses,

    // Actions
    clearError,
    setCurrentCourse,
    setCurrentUnit,
    fetchCourses,
    setSearch,
    setStatusFilter,
    setPage,
    resetFilters,
    fetchCourseDetail,
    createCourse,
    updateCourse,
    deleteCourse,
    fetchCourseStudents,
    enrollStudents,
    unenrollStudent,
    fetchCourseUnits,
    createCourseUnit,
    reorderCourseUnits,
    fetchCourseUnit,
    updateCourseUnit,
    deleteCourseUnit,
    linkQuizToUnit,
    unlinkQuizFromUnit,
    fetchUnitScript,
    updateUnitScript
  }
})
