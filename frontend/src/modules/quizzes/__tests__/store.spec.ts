import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useQuizStore } from '../store'
import * as quizApi from '../api'
import type { Quiz, QuizDetail } from '../types'

vi.mock('../api')

describe('Quiz Pinia Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.mocked(quizApi.listQuizzes).mockResolvedValue({
      items: [],
      pagination: { page: 1, pageSize: 12, totalCount: 0, totalPages: 0 }
    })
  })

  it('initializes with default state', () => {
    const store = useQuizStore()
    expect(store.quizzes).toEqual([])
    expect(store.currentQuiz).toBeNull()
    expect(store.currentPage).toBe(1)
    expect(store.pageSize).toBe(12)
    expect(store.totalCount).toBe(0)
    expect(store.search).toBe('')
    expect(store.statusFilter).toBeUndefined()
    expect(store.tagFilter).toBeUndefined()
    expect(store.isLoading).toBe(false)
    expect(store.isSaving).toBe(false)
  })

  it('initNewQuiz sets a fresh empty quiz structure with 1 question and 4 Kahoot answers', () => {
    const store = useQuizStore()
    const newQuiz = store.initNewQuiz()

    expect(newQuiz.title).toBe('')
    expect(newQuiz.status).toBe('draft')
    expect(newQuiz.questions.length).toBe(1)
    expect(newQuiz.questions[0].answers.length).toBe(4)
    expect(newQuiz.questions[0].answers[0].isCorrect).toBe(true)
    expect(store.currentQuiz).toEqual(newQuiz)
  })

  it('fetchQuizzes updates quizzes and pagination on success', async () => {
    const mockList: Quiz[] = [
      {
        id: 'quiz-1',
        creatorId: 'user-1',
        title: 'Geografia',
        status: 'published',
        tags: ['geo'],
        questionCount: 5,
        createdAt: '2026-08-21T10:00:00Z',
        updatedAt: '2026-08-21T10:00:00Z'
      }
    ]
    vi.mocked(quizApi.listQuizzes).mockResolvedValue({
      items: mockList,
      pagination: { page: 1, pageSize: 12, totalCount: 1, totalPages: 1 }
    })

    const store = useQuizStore()
    await store.fetchQuizzes()

    expect(store.quizzes).toEqual(mockList)
    expect(store.totalCount).toBe(1)
    expect(store.totalPages).toBe(1)
    expect(store.hasQuizzes).toBe(true)
    expect(store.isLoading).toBe(false)
  })

  it('setSearch updates search term and triggers fetch', () => {
    const store = useQuizStore()
    store.setSearch('historia')
    expect(store.search).toBe('historia')
    expect(store.currentPage).toBe(1)
    expect(quizApi.listQuizzes).toHaveBeenCalled()
  })

  it('setStatusFilter updates status filter and triggers fetch', () => {
    const store = useQuizStore()
    store.setStatusFilter('published')
    expect(store.statusFilter).toBe('published')
    expect(quizApi.listQuizzes).toHaveBeenCalled()
  })

  it('resetFilters clears all filters and resets page to 1', () => {
    const store = useQuizStore()
    store.search = 'query'
    store.statusFilter = 'draft'
    store.tagFilter = 'primaria'
    store.currentPage = 3

    store.resetFilters()

    expect(store.search).toBe('')
    expect(store.statusFilter).toBeUndefined()
    expect(store.tagFilter).toBeUndefined()
    expect(store.currentPage).toBe(1)
    expect(quizApi.listQuizzes).toHaveBeenCalled()
  })

  it('createQuiz calls api and triggers fetchQuizzes', async () => {
    const store = useQuizStore()
    const createdQuiz: QuizDetail = {
      id: 'created-1',
      creatorId: 'user-1',
      title: 'Història',
      status: 'draft',
      tags: [],
      questionCount: 0,
      createdAt: '',
      updatedAt: '',
      questions: []
    }
    vi.mocked(quizApi.createQuiz).mockResolvedValue(createdQuiz)
    vi.mocked(quizApi.listQuizzes).mockResolvedValue({
      items: [createdQuiz],
      pagination: { page: 1, pageSize: 12, totalCount: 1, totalPages: 1 }
    })

    const result = await store.createQuiz({ title: 'Història' })
    expect(result).toEqual(createdQuiz)
    expect(store.currentQuiz).toEqual(createdQuiz)
    expect(store.quizzes.length).toBe(1)
  })

  it('updateQuiz updates currentQuiz and local list item', async () => {
    const store = useQuizStore()
    store.quizzes = [
      {
        id: 'quiz-1',
        creatorId: 'user-1',
        title: 'Original Title',
        status: 'draft',
        tags: [],
        questionCount: 1,
        createdAt: '',
        updatedAt: ''
      }
    ]

    const updatedDetail: QuizDetail = {
      id: 'quiz-1',
      creatorId: 'user-1',
      title: 'Updated Title',
      status: 'published',
      tags: ['nova'],
      questionCount: 1,
      createdAt: '',
      updatedAt: '',
      questions: []
    }
    vi.mocked(quizApi.updateQuiz).mockResolvedValue(updatedDetail)

    const result = await store.updateQuiz('quiz-1', { title: 'Updated Title', status: 'published' })
    expect(result.title).toBe('Updated Title')
    expect(store.quizzes[0].title).toBe('Updated Title')
    expect(store.quizzes[0].status).toBe('published')
  })

  it('deleteQuiz removes item from store quizzes list', async () => {
    const store = useQuizStore()
    store.quizzes = [
      {
        id: 'quiz-1',
        creatorId: 'user-1',
        title: 'To Delete',
        status: 'draft',
        tags: [],
        questionCount: 0,
        createdAt: '',
        updatedAt: ''
      }
    ]
    store.totalCount = 1
    vi.mocked(quizApi.deleteQuiz).mockResolvedValue({ message: 'Qüestionari eliminat' })

    const msg = await store.deleteQuiz('quiz-1')
    expect(msg).toBe('Qüestionari eliminat')
    expect(store.quizzes).toHaveLength(0)
    expect(store.totalCount).toBe(0)
  })

  it('duplicateQuiz calls duplicate endpoint and re-fetches quizzes', async () => {
    const store = useQuizStore()
    const duplicated: QuizDetail = {
      id: 'dup-1',
      creatorId: 'user-1',
      title: '[Còpia] Original',
      status: 'draft',
      tags: [],
      questionCount: 2,
      createdAt: '',
      updatedAt: '',
      questions: []
    }
    vi.mocked(quizApi.duplicateQuiz).mockResolvedValue(duplicated)
    vi.mocked(quizApi.listQuizzes).mockResolvedValue({
      items: [duplicated],
      pagination: { page: 1, pageSize: 12, totalCount: 1, totalPages: 1 }
    })

    const res = await store.duplicateQuiz('orig-1', { includeAnswers: true })
    expect(res).toEqual(duplicated)
    expect(store.quizzes).toContainEqual(duplicated)
  })

  it('uploadImage returns public URL', async () => {
    const store = useQuizStore()
    const mockFile = new File(['data'], 'img.jpg', { type: 'image/jpeg' })
    vi.mocked(quizApi.uploadImage).mockResolvedValue({
      url: 'https://pub-r2.encertia.cat/img.jpg',
      key: 'img.jpg'
    })

    const url = await store.uploadImage(mockFile)
    expect(url).toBe('https://pub-r2.encertia.cat/img.jpg')
  })
})
