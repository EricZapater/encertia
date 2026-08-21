import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  Quiz,
  QuizDetail,
  QuizStatus,
  QuizFilters,
  CreateQuizRequest,
  UpdateQuizRequest,
  DuplicateQuizRequest,
  QuizQuestion,
  QuizAnswer
} from './types'
import * as quizApi from './api'

export const useQuizStore = defineStore('quiz', () => {
  // State
  const quizzes = ref<Quiz[]>([])
  const currentQuiz = ref<QuizDetail | null>(null)

  const currentPage = ref(1)
  const pageSize = ref(12)
  const totalCount = ref(0)
  const totalPages = ref(0)

  const search = ref('')
  const statusFilter = ref<QuizStatus | undefined>(undefined)
  const tagFilter = ref<string | undefined>(undefined)

  const isLoading = ref(false)
  const isSaving = ref(false)
  const isUploading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const quizList = computed(() => quizzes.value)
  const hasQuizzes = computed(() => quizzes.value.length > 0)
  const totalQuizzes = computed(() => totalCount.value)

  // Actions
  function clearError() {
    error.value = null
  }

  function setCurrentQuiz(quiz: QuizDetail | null) {
    currentQuiz.value = quiz
  }

  function initNewQuiz(): QuizDetail {
    const defaultAnswers: QuizAnswer[] = [
      { id: 'new-ans-0', text: '', isCorrect: true, orderIndex: 0 },
      { id: 'new-ans-1', text: '', isCorrect: false, orderIndex: 1 },
      { id: 'new-ans-2', text: '', isCorrect: false, orderIndex: 2 },
      { id: 'new-ans-3', text: '', isCorrect: false, orderIndex: 3 }
    ]

    const defaultQuestion: QuizQuestion = {
      id: 'new-q-0',
      text: '',
      imageUrl: null,
      questionType: 'single_choice',
      timeLimitSeconds: 20,
      orderIndex: 0,
      answers: defaultAnswers
    }

    const newQuiz: QuizDetail = {
      id: '',
      creatorId: '',
      title: '',
      description: '',
      coverImageUrl: null,
      status: 'draft',
      tags: [],
      questionCount: 1,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      questions: [defaultQuestion]
    }

    currentQuiz.value = newQuiz
    return newQuiz
  }

  async function fetchQuizzes() {
    isLoading.value = true
    error.value = null
    try {
      const filters: QuizFilters = {
        page: currentPage.value,
        pageSize: pageSize.value,
        search: search.value,
        status: statusFilter.value,
        tag: tagFilter.value
      }
      const response = await quizApi.listQuizzes(filters)
      quizzes.value = response.items || []
      currentPage.value = response.pagination.page
      pageSize.value = response.pagination.pageSize
      totalCount.value = response.pagination.totalCount
      totalPages.value = response.pagination.totalPages
    } catch (err: any) {
      error.value =
        err.response?.data?.error || err.message || 'Error en carregar els qüestionaris.'
      quizzes.value = []
    } finally {
      isLoading.value = false
    }
  }

  function setSearch(newSearch: string) {
    search.value = newSearch
    currentPage.value = 1
    fetchQuizzes()
  }

  function setStatusFilter(newStatus?: QuizStatus) {
    statusFilter.value = newStatus
    currentPage.value = 1
    fetchQuizzes()
  }

  function setTagFilter(newTag?: string) {
    tagFilter.value = newTag
    currentPage.value = 1
    fetchQuizzes()
  }

  function setPage(page: number, size?: number) {
    currentPage.value = page
    if (size) pageSize.value = size
    fetchQuizzes()
  }

  function resetFilters() {
    search.value = ''
    statusFilter.value = undefined
    tagFilter.value = undefined
    currentPage.value = 1
    fetchQuizzes()
  }

  async function fetchQuizDetail(id: string): Promise<QuizDetail> {
    isLoading.value = true
    error.value = null
    try {
      const detail = await quizApi.getQuizById(id)
      currentQuiz.value = detail
      return detail
    } catch (err: any) {
      error.value =
        err.response?.data?.error ||
        err.message ||
        'Error en carregar el detall del qüestionari.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function createQuiz(payload: CreateQuizRequest): Promise<QuizDetail> {
    isSaving.value = true
    error.value = null
    try {
      const created = await quizApi.createQuiz(payload)
      currentQuiz.value = created
      await fetchQuizzes()
      return created
    } catch (err: any) {
      error.value =
        err.response?.data?.error || err.message || 'Error en crear el qüestionari.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function updateQuiz(id: string, payload: UpdateQuizRequest): Promise<QuizDetail> {
    isSaving.value = true
    error.value = null
    try {
      const updated = await quizApi.updateQuiz(id, payload)
      currentQuiz.value = updated
      // Actualitza l'element a la llista si existeix
      const idx = quizzes.value.findIndex((q) => q.id === id)
      if (idx !== -1) {
        quizzes.value[idx] = { ...quizzes.value[idx], ...updated }
      }
      return updated
    } catch (err: any) {
      error.value =
        err.response?.data?.error || err.message || 'Error en actualitzar el qüestionari.'
      throw err
    } finally {
      isSaving.value = false
    }
  }

  async function deleteQuiz(id: string): Promise<string> {
    isLoading.value = true
    error.value = null
    try {
      const res = await quizApi.deleteQuiz(id)
      quizzes.value = quizzes.value.filter((q) => q.id !== id)
      totalCount.value = Math.max(0, totalCount.value - 1)
      return res.message || 'Qüestionari eliminat correctament.'
    } catch (err: any) {
      error.value =
        err.response?.data?.error || err.message || 'Error en eliminar el qüestionari.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function duplicateQuiz(
    id: string,
    payload?: DuplicateQuizRequest
  ): Promise<QuizDetail> {
    isLoading.value = true
    error.value = null
    try {
      const duplicated = await quizApi.duplicateQuiz(id, payload)
      await fetchQuizzes()
      return duplicated
    } catch (err: any) {
      error.value =
        err.response?.data?.error || err.message || 'Error en duplicar el qüestionari.'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function uploadImage(file: File): Promise<string> {
    isUploading.value = true
    error.value = null
    try {
      const res = await quizApi.uploadImage(file)
      return res.url
    } catch (err: any) {
      const status = err.response?.status
      let message: string
      if (status === 413) {
        message = `La imatge és massa gran per ser pujada. La mida màxima permesa és de 5 MB (la teva pesa ${(file.size / (1024 * 1024)).toFixed(1)} MB).`
      } else {
        message = err.response?.data?.error?.message ?? err.response?.data?.error ?? err.message ?? 'Error en pujar la imatge.'
      }
      error.value = message
      const uploadError = new Error(message)
      throw uploadError
    } finally {
      isUploading.value = false
    }
  }

  return {
    // State
    quizzes,
    currentQuiz,
    currentPage,
    pageSize,
    totalCount,
    totalPages,
    search,
    statusFilter,
    tagFilter,
    isLoading,
    isSaving,
    isUploading,
    error,

    // Getters
    quizList,
    hasQuizzes,
    totalQuizzes,

    // Actions
    clearError,
    setCurrentQuiz,
    initNewQuiz,
    fetchQuizzes,
    setSearch,
    setStatusFilter,
    setTagFilter,
    setPage,
    resetFilters,
    fetchQuizDetail,
    createQuiz,
    updateQuiz,
    deleteQuiz,
    duplicateQuiz,
    uploadImage
  }
})
