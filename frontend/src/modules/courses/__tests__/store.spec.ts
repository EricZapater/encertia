import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useCourseStore } from '../store'
import * as coursesApi from '../api'

vi.mock('../api')

describe('useCourseStore Pinia Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('initial state is correct', () => {
    const store = useCourseStore()
    expect(store.courses).toEqual([])
    expect(store.currentCourse).toBeNull()
    expect(store.currentUnit).toBeNull()
    expect(store.enrolledStudents).toEqual([])
    expect(store.currentPage).toBe(1)
    expect(store.isLoading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('fetchCourses updates courses and pagination in store', async () => {
    const mockList = {
      items: [
        {
          id: 'c-1',
          title: 'Física',
          code: 'FIS-101',
          status: 'active' as const,
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
    vi.mocked(coursesApi.listCourses).mockResolvedValue(mockList)

    const store = useCourseStore()
    await store.fetchCourses()

    expect(store.courses).toEqual(mockList.items)
    expect(store.totalCount).toBe(1)
    expect(store.isLoading).toBe(false)
  })

  it('createCourse calls api and refreshes list', async () => {
    const store = useCourseStore()
    const newCourse = {
      id: 'c-2',
      title: 'Química',
      code: 'QUI-101',
      status: 'draft' as const,
      teacherId: 't-1',
      createdAt: '',
      updatedAt: ''
    }
    vi.mocked(coursesApi.createCourse).mockResolvedValue(newCourse)
    vi.mocked(coursesApi.listCourses).mockResolvedValue({
      items: [newCourse],
      total: 1,
      page: 1,
      pageSize: 10,
      totalPages: 1
    })

    const result = await store.createCourse({ title: 'Química', code: 'QUI-101' })

    expect(coursesApi.createCourse).toHaveBeenCalledWith({ title: 'Química', code: 'QUI-101' })
    expect(result).toEqual(newCourse)
    expect(store.courses).toHaveLength(1)
  })

  it('deleteCourse removes course from store state', async () => {
    const store = useCourseStore()
    store.courses = [
      {
        id: 'c-1',
        title: 'Física',
        code: 'FIS-101',
        status: 'active',
        teacherId: 't-1',
        createdAt: '',
        updatedAt: ''
      }
    ]
    store.totalCount = 1

    vi.mocked(coursesApi.deleteCourse).mockResolvedValue()

    await store.deleteCourse('c-1')

    expect(coursesApi.deleteCourse).toHaveBeenCalledWith('c-1')
    expect(store.courses).toEqual([])
    expect(store.totalCount).toBe(0)
  })

  it('enrollStudents updates enrolledStudents state', async () => {
    const store = useCourseStore()
    const mockStudentsResponse = {
      courseId: 'c-1',
      total: 1,
      students: [
        {
          id: 's-1',
          email: 'student@encertia.cat',
          enrolledAt: '2026-09-02T10:00:00Z'
        }
      ]
    }
    vi.mocked(coursesApi.enrollStudents).mockResolvedValue(mockStudentsResponse)

    await store.enrollStudents('c-1', ['s-1'])

    expect(store.enrolledStudents).toEqual(mockStudentsResponse.students)
  })
})
