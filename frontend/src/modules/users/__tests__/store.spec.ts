import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useUserStore } from '../store'
import * as userApi from '../api'
import type {
  BatchCreateUsersResponse,
  CreateUserRequest,
  UpdateUserRequest,
  User,
  UserListResponse
} from '../types'

vi.mock('../api')

describe('useUserStore Pinia Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('has initial default state', () => {
    const store = useUserStore()

    expect(store.users).toEqual([])
    expect(store.selectedUser).toBeNull()
    expect(store.pagination).toEqual({
      page: 1,
      pageSize: 20,
      totalCount: 0,
      totalPages: 0
    })
    expect(store.filters).toEqual({
      search: '',
      role: undefined,
      status: 'active'
    })
    expect(store.isLoading).toBe(false)
    expect(store.isSubmitting).toBe(false)
    expect(store.error).toBeNull()
    expect(store.hasUsers).toBe(false)
  })

  describe('fetchUsers', () => {
    it('fetches users list and updates store state', async () => {
      const store = useUserStore()
      const mockUsers: User[] = [
        {
          id: 'user-1',
          email: 'marc@encertia.cat',
          firstName: 'Marc',
          lastName: 'Rovira',
          role: 'student',
          isActive: true,
          createdAt: '2026-08-21T10:00:00Z'
        }
      ]
      const mockResponse: UserListResponse = {
        items: mockUsers,
        pagination: {
          page: 1,
          pageSize: 20,
          totalCount: 1,
          totalPages: 1
        }
      }

      vi.mocked(userApi.listUsers).mockResolvedValueOnce(mockResponse)

      await store.fetchUsers()

      expect(userApi.listUsers).toHaveBeenCalledWith({
        page: 1,
        pageSize: 20,
        status: 'active'
      })
      expect(store.users).toEqual(mockUsers)
      expect(store.totalUsers).toBe(1)
      expect(store.hasUsers).toBe(true)
      expect(store.isLoading).toBe(false)
    })

    it('handles fetch errors correctly', async () => {
      const store = useUserStore()
      vi.mocked(userApi.listUsers).mockRejectedValueOnce({
        response: { data: { error: { message: 'Error de servidor' } } }
      })

      await expect(store.fetchUsers()).rejects.toBeDefined()
      expect(store.error).toBe('Error de servidor')
      expect(store.isLoading).toBe(false)
    })
  })

  describe('filter and pagination actions', () => {
    it('setPage updates page and triggers fetchUsers', async () => {
      const store = useUserStore()
      vi.mocked(userApi.listUsers).mockResolvedValue({
        items: [],
        pagination: { page: 2, pageSize: 10, totalCount: 20, totalPages: 2 }
      })

      await store.setPage(2, 10)

      expect(store.pagination.page).toBe(2)
      expect(store.pagination.pageSize).toBe(10)
      expect(userApi.listUsers).toHaveBeenCalled()
    })

    it('setSearch updates search filter, resets page to 1 and fetches', async () => {
      const store = useUserStore()
      store.pagination.page = 3
      vi.mocked(userApi.listUsers).mockResolvedValue({
        items: [],
        pagination: { page: 1, pageSize: 20, totalCount: 0, totalPages: 0 }
      })

      await store.setSearch('maria')

      expect(store.filters.search).toBe('maria')
      expect(store.pagination.page).toBe(1)
      expect(userApi.listUsers).toHaveBeenCalledWith(
        expect.objectContaining({ search: 'maria', page: 1 })
      )
    })

    it('setRoleFilter updates role and resets page', async () => {
      const store = useUserStore()
      store.pagination.page = 4
      vi.mocked(userApi.listUsers).mockResolvedValue({
        items: [],
        pagination: { page: 1, pageSize: 20, totalCount: 0, totalPages: 0 }
      })

      await store.setRoleFilter('teacher')

      expect(store.filters.role).toBe('teacher')
      expect(store.pagination.page).toBe(1)
      expect(userApi.listUsers).toHaveBeenCalledWith(
        expect.objectContaining({ role: 'teacher', page: 1 })
      )
    })

    it('setStatusFilter updates status filter and resets page', async () => {
      const store = useUserStore()
      vi.mocked(userApi.listUsers).mockResolvedValue({
        items: [],
        pagination: { page: 1, pageSize: 20, totalCount: 0, totalPages: 0 }
      })

      await store.setStatusFilter('inactive')

      expect(store.filters.status).toBe('inactive')
      expect(userApi.listUsers).toHaveBeenCalledWith(
        expect.objectContaining({ status: 'inactive', page: 1 })
      )
    })

    it('resetFilters restores default filters and page', async () => {
      const store = useUserStore()
      store.filters.search = 'query'
      store.filters.role = 'admin'
      store.filters.status = 'all'
      store.pagination.page = 5

      vi.mocked(userApi.listUsers).mockResolvedValue({
        items: [],
        pagination: { page: 1, pageSize: 20, totalCount: 0, totalPages: 0 }
      })

      await store.resetFilters()

      expect(store.filters.search).toBe('')
      expect(store.filters.role).toBeUndefined()
      expect(store.filters.status).toBe('active')
      expect(store.pagination.page).toBe(1)
    })
  })

  describe('CRUD actions', () => {
    it('createUser calls api and refreshes users list', async () => {
      const store = useUserStore()
      const newUserData: CreateUserRequest = {
        email: 'clara@encertia.cat',
        password: 'Password123!',
        firstName: 'Clara',
        lastName: 'Vidal',
        role: 'student'
      }
      const createdUser: User = {
        id: 'u-clara',
        ...newUserData,
        isActive: true,
        createdAt: '2026-08-21T10:00:00Z'
      }

      vi.mocked(userApi.createUser).mockResolvedValueOnce({ user: createdUser })
      vi.mocked(userApi.listUsers).mockResolvedValueOnce({
        items: [createdUser],
        pagination: { page: 1, pageSize: 20, totalCount: 1, totalPages: 1 }
      })

      const result = await store.createUser(newUserData)

      expect(userApi.createUser).toHaveBeenCalledWith(newUserData)
      expect(userApi.listUsers).toHaveBeenCalled()
      expect(result).toEqual(createdUser)
    })

    it('updateUser calls api and updates local state in users array', async () => {
      const store = useUserStore()
      const existingUser: User = {
        id: 'u-1',
        email: 'antic@encertia.cat',
        firstName: 'Antic',
        lastName: 'Nom',
        role: 'student',
        isActive: true,
        createdAt: '2026-08-21T10:00:00Z'
      }
      store.users = [existingUser]

      const updateData: UpdateUserRequest = {
        firstName: 'Nou Nom'
      }
      const updatedUser: User = {
        ...existingUser,
        firstName: 'Nou Nom'
      }

      vi.mocked(userApi.updateUser).mockResolvedValueOnce({ user: updatedUser })

      const result = await store.updateUser('u-1', updateData)

      expect(userApi.updateUser).toHaveBeenCalledWith('u-1', updateData)
      expect(store.users[0].firstName).toBe('Nou Nom')
      expect(result).toEqual(updatedUser)
    })

    it('deleteUser calls api and refreshes list', async () => {
      const store = useUserStore()
      vi.mocked(userApi.deleteUser).mockResolvedValueOnce({
        message: 'Usuari donat de baixa correctament.'
      })
      vi.mocked(userApi.listUsers).mockResolvedValueOnce({
        items: [],
        pagination: { page: 1, pageSize: 20, totalCount: 0, totalPages: 0 }
      })

      const msg = await store.deleteUser('u-1')

      expect(userApi.deleteUser).toHaveBeenCalledWith('u-1')
      expect(userApi.listUsers).toHaveBeenCalled()
      expect(msg).toBe('Usuari donat de baixa correctament.')
    })

    it('resetUserPassword calls api with new password', async () => {
      const store = useUserStore()
      vi.mocked(userApi.resetUserPassword).mockResolvedValueOnce({
        message: 'Contrasenya actualitzada correctament.'
      })

      const msg = await store.resetUserPassword('u-1', {
        newPassword: 'NovaPassword123!'
      })

      expect(userApi.resetUserPassword).toHaveBeenCalledWith('u-1', {
        newPassword: 'NovaPassword123!'
      })
      expect(msg).toBe('Contrasenya actualitzada correctament.')
    })

    it('batchImport calls api, stores lastBatchResult and refreshes list', async () => {
      const store = useUserStore()
      const mockBatchResponse: BatchCreateUsersResponse = {
        totalRequested: 2,
        createdCount: 2,
        failedCount: 0,
        createdUsers: [],
        errors: []
      }

      vi.mocked(userApi.batchCreateUsers).mockResolvedValueOnce(mockBatchResponse)
      vi.mocked(userApi.listUsers).mockResolvedValueOnce({
        items: [],
        pagination: { page: 1, pageSize: 20, totalCount: 2, totalPages: 1 }
      })

      const result = await store.batchImport({
        users: [
          { email: 'u1@encertia.cat', firstName: 'U1', lastName: 'L1', role: 'student' },
          { email: 'u2@encertia.cat', firstName: 'U2', lastName: 'L2', role: 'student' }
        ]
      })

      expect(userApi.batchCreateUsers).toHaveBeenCalled()
      expect(store.lastBatchResult).toEqual(mockBatchResponse)
      expect(result).toEqual(mockBatchResponse)
    })
  })
})
