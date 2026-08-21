import { describe, it, expect, vi, beforeEach } from 'vitest'
import apiClient from '@/api/client'
import {
  listQuizzes,
  getQuizById,
  createQuiz,
  updateQuiz,
  deleteQuiz,
  duplicateQuiz,
  uploadImage
} from '../api'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

describe('Quizzes API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('listQuizzes calls GET /quizzes with query parameters', async () => {
    const mockResponse = {
      data: {
        items: [],
        pagination: { page: 1, pageSize: 12, totalCount: 0, totalPages: 0 }
      }
    }
    vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

    const result = await listQuizzes({
      page: 2,
      pageSize: 10,
      search: 'historia',
      status: 'published',
      tag: 'secundaria'
    })

    expect(apiClient.get).toHaveBeenCalledWith('/quizzes', {
      params: {
        page: 2,
        pageSize: 10,
        search: 'historia',
        status: 'published',
        tag: 'secundaria'
      }
    })
    expect(result).toEqual(mockResponse.data)
  })

  it('getQuizById calls GET /quizzes/:id', async () => {
    const mockQuizDetail = {
      id: 'quiz-123',
      creatorId: 'user-1',
      title: 'Ciències',
      status: 'published',
      tags: ['ciencies'],
      questionCount: 1,
      createdAt: '2026-08-21T10:00:00Z',
      updatedAt: '2026-08-21T10:00:00Z',
      questions: []
    }
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockQuizDetail })

    const result = await getQuizById('quiz-123')
    expect(apiClient.get).toHaveBeenCalledWith('/quizzes/quiz-123')
    expect(result).toEqual(mockQuizDetail)
  })

  it('createQuiz calls POST /quizzes with payload', async () => {
    const payload = {
      title: 'Geografia Catalana',
      description: 'Descripció del joc',
      status: 'draft' as const,
      tags: ['geo'],
      questions: []
    }
    const mockCreated = { id: 'new-id', ...payload, questionCount: 0, createdAt: '', updatedAt: '' }
    vi.mocked(apiClient.post).mockResolvedValue({ data: mockCreated })

    const result = await createQuiz(payload)
    expect(apiClient.post).toHaveBeenCalledWith('/quizzes', payload)
    expect(result).toEqual(mockCreated)
  })

  it('updateQuiz calls PUT /quizzes/:id with payload', async () => {
    const payload = {
      title: 'Geografia Catalana V2',
      status: 'published' as const
    }
    const mockUpdated = { id: 'quiz-123', ...payload }
    vi.mocked(apiClient.put).mockResolvedValue({ data: mockUpdated })

    const result = await updateQuiz('quiz-123', payload)
    expect(apiClient.put).toHaveBeenCalledWith('/quizzes/quiz-123', payload)
    expect(result).toEqual(mockUpdated)
  })

  it('deleteQuiz calls DELETE /quizzes/:id', async () => {
    const mockResponse = { data: { message: 'Qüestionari eliminat correctament' } }
    vi.mocked(apiClient.delete).mockResolvedValue(mockResponse)

    const result = await deleteQuiz('quiz-123')
    expect(apiClient.delete).toHaveBeenCalledWith('/quizzes/quiz-123')
    expect(result).toEqual(mockResponse.data)
  })

  it('duplicateQuiz calls POST /quizzes/:id/duplicate with includeAnswers flag', async () => {
    const mockDuplicated = { id: 'copy-123', title: '[Còpia] Ciències' }
    vi.mocked(apiClient.post).mockResolvedValue({ data: mockDuplicated })

    const result = await duplicateQuiz('quiz-123', {
      title: 'Còpia Ciències Grup A',
      includeAnswers: true
    })

    expect(apiClient.post).toHaveBeenCalledWith('/quizzes/quiz-123/duplicate', {
      title: 'Còpia Ciències Grup A',
      includeAnswers: true
    })
    expect(result).toEqual(mockDuplicated)
  })

  it('uploadImage sends multipart/form-data to /uploads/images', async () => {
    const mockFile = new File(['fake-image-content'], 'quiz-cover.png', { type: 'image/png' })
    const mockUploadResponse = {
      data: {
        url: 'https://pub-r2.encertia.cat/uploads/quiz-cover.png',
        key: 'uploads/quiz-cover.png'
      }
    }
    vi.mocked(apiClient.post).mockResolvedValue(mockUploadResponse)

    const result = await uploadImage(mockFile)

    expect(apiClient.post).toHaveBeenCalledWith(
      '/uploads/images',
      expect.any(FormData),
      expect.objectContaining({
        headers: { 'Content-Type': 'multipart/form-data' }
      })
    )
    expect(result).toEqual(mockUploadResponse.data)
  })
})
