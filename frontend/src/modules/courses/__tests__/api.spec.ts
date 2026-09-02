import { describe, it, expect, vi, beforeEach } from 'vitest'
import apiClient from '@/api/client'
import {
  listCourses,
  createCourse,
  getCourseById,
  updateCourse,
  deleteCourse,
  getCourseStudents,
  enrollStudents,
  unenrollStudent,
  listCourseUnits,
  createCourseUnit,
  reorderCourseUnits,
  getCourseUnit,
  updateCourseUnit,
  deleteCourseUnit,
  linkQuizToUnit,
  unlinkQuizFromUnit,
  getUnitScript,
  updateUnitScript
} from '../api'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

describe('Courses API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('listCourses calls GET /courses with query parameters', async () => {
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

    const result = await listCourses({
      page: 1,
      pageSize: 10,
      search: 'mates',
      status: 'active'
    })

    expect(apiClient.get).toHaveBeenCalledWith('/courses', {
      params: {
        page: 1,
        pageSize: 10,
        search: 'mates',
        status: 'active'
      }
    })
    expect(result).toEqual(mockResponse.data)
  })

  it('createCourse calls POST /courses with payload', async () => {
    const payload = { title: 'Matemàtiques', code: 'MAT-101' }
    const mockCreated = { id: 'c-1', ...payload, status: 'draft', teacherId: 't-1', createdAt: '', updatedAt: '' }
    vi.mocked(apiClient.post).mockResolvedValue({ data: mockCreated })

    const result = await createCourse(payload)
    expect(apiClient.post).toHaveBeenCalledWith('/courses', payload)
    expect(result).toEqual(mockCreated)
  })

  it('getCourseById calls GET /courses/:id', async () => {
    const mockCourse = { id: 'c-1', title: 'Matemàtiques', code: 'MAT-101' }
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockCourse })

    const result = await getCourseById('c-1')
    expect(apiClient.get).toHaveBeenCalledWith('/courses/c-1')
    expect(result).toEqual(mockCourse)
  })

  it('updateCourse calls PUT /courses/:id with payload', async () => {
    const payload = { title: 'Matemàtiques Avançades' }
    const mockUpdated = { id: 'c-1', title: 'Matemàtiques Avançades' }
    vi.mocked(apiClient.put).mockResolvedValue({ data: mockUpdated })

    const result = await updateCourse('c-1', payload)
    expect(apiClient.put).toHaveBeenCalledWith('/courses/c-1', payload)
    expect(result).toEqual(mockUpdated)
  })

  it('deleteCourse calls DELETE /courses/:id', async () => {
    vi.mocked(apiClient.delete).mockResolvedValue({})

    await deleteCourse('c-1')
    expect(apiClient.delete).toHaveBeenCalledWith('/courses/c-1')
  })

  it('enrollStudents calls POST /courses/:id/students', async () => {
    const payload = { studentIds: ['s-1', 's-2'] }
    const mockResponse = { data: { courseId: 'c-1', total: 2, students: [] } }
    vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

    const result = await enrollStudents('c-1', payload)
    expect(apiClient.post).toHaveBeenCalledWith('/courses/c-1/students', payload)
    expect(result).toEqual(mockResponse.data)
  })

  it('unenrollStudent calls DELETE /courses/:id/students/:studentId', async () => {
    vi.mocked(apiClient.delete).mockResolvedValue({})

    await unenrollStudent('c-1', 's-1')
    expect(apiClient.delete).toHaveBeenCalledWith('/courses/c-1/students/s-1')
  })

  it('createCourseUnit calls POST /courses/:courseId/units', async () => {
    const payload = { title: 'Tema 1', orderIndex: 0 }
    const mockUnit = { id: 'u-1', courseId: 'c-1', ...payload }
    vi.mocked(apiClient.post).mockResolvedValue({ data: mockUnit })

    const result = await createCourseUnit('c-1', payload)
    expect(apiClient.post).toHaveBeenCalledWith('/courses/c-1/units', payload)
    expect(result).toEqual(mockUnit)
  })

  it('reorderCourseUnits calls PUT /courses/:courseId/units/reorder', async () => {
    const unitIds = ['u-2', 'u-1']
    vi.mocked(apiClient.put).mockResolvedValue({ data: [] })

    await reorderCourseUnits('c-1', unitIds)
    expect(apiClient.put).toHaveBeenCalledWith('/courses/c-1/units/reorder', unitIds)
  })

  it('linkQuizToUnit and unlinkQuizFromUnit call correct endpoints', async () => {
    vi.mocked(apiClient.post).mockResolvedValue({})
    vi.mocked(apiClient.delete).mockResolvedValue({})

    await linkQuizToUnit('c-1', 'u-1', 'q-1')
    expect(apiClient.post).toHaveBeenCalledWith('/courses/c-1/units/u-1/quizzes', { quizId: 'q-1' })

    await unlinkQuizFromUnit('c-1', 'u-1', 'q-1')
    expect(apiClient.delete).toHaveBeenCalledWith('/courses/c-1/units/u-1/quizzes/q-1')
  })

  it('getUnitScript and updateUnitScript call correct endpoints', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: [] })
    vi.mocked(apiClient.put).mockResolvedValue({ data: [] })

    await getUnitScript('c-1', 'u-1')
    expect(apiClient.get).toHaveBeenCalledWith('/courses/c-1/units/u-1/script')

    const blocks = [{ blockType: 'material' as const, orderIndex: 0, title: 'Intro' }]
    await updateUnitScript('c-1', 'u-1', blocks)
    expect(apiClient.put).toHaveBeenCalledWith('/courses/c-1/units/u-1/script', blocks)
  })
})
