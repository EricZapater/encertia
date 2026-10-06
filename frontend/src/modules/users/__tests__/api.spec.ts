import { describe, it, expect, vi, beforeEach } from 'vitest'
import apiClient from '@/api/client'
import * as userApi from '../api'
import type {
  BatchCreateUsersRequest,
  CreateUserRequest,
  ResetPasswordRequest,
  UpdateUserRequest,
  UserListParams
} from '../types'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

describe('User API client functions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('listUsers', () => {
    it('calls GET /users with expected query parameters', async () => {
      const mockResponse = {
        data: {
          items: [
            {
              id: '1',
              email: 'test@encertia.cat',
              firstName: 'Test',
              lastName: 'User',
              role: 'student',
              isActive: true,
              createdAt: '2026-08-21T10:00:00Z'
            }
          ],
          pagination: { page: 1, pageSize: 20, totalCount: 1, totalPages: 1 }
        }
      }
      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse)

      const params: UserListParams = {
        page: 2,
        pageSize: 10,
        search: 'test',
        role: 'student',
        status: 'active'
      }

      const result = await userApi.listUsers(params)

      expect(apiClient.get).toHaveBeenCalledWith('/users', { params })
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('getUserById', () => {
    it('calls GET /users/:id with UUID', async () => {
      const mockResponse = {
        data: {
          user: {
            id: 'uuid-123',
            email: 'marc@encertia.cat',
            firstName: 'Marc',
            lastName: 'Rovira',
            role: 'student',
            isActive: true,
            createdAt: '2026-08-21T10:00:00Z'
          }
        }
      }
      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse)

      const result = await userApi.getUserById('uuid-123')

      expect(apiClient.get).toHaveBeenCalledWith('/users/uuid-123')
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('createUser', () => {
    it('calls POST /users with create payload', async () => {
      const payload: CreateUserRequest = {
        email: 'laia@encertia.cat',
        password: 'Password123!',
        firstName: 'Laia',
        lastName: 'Sole',
        role: 'teacher'
      }
      const mockResponse = {
        data: {
          user: {
            id: 'new-id',
            ...payload,
            isActive: true,
            createdAt: '2026-08-21T10:00:00Z'
          }
        }
      }
      vi.mocked(apiClient.post).mockResolvedValueOnce(mockResponse)

      const result = await userApi.createUser(payload)

      expect(apiClient.post).toHaveBeenCalledWith('/users', payload)
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('updateUser', () => {
    it('calls PUT /users/:id with update payload', async () => {
      const payload: UpdateUserRequest = {
        firstName: 'Laia Modificada',
        lastName: 'Sole Costa',
        email: 'laia.mod@encertia.cat',
        role: 'admin',
        isActive: false
      }
      const mockResponse = {
        data: {
          user: {
            id: 'uuid-123',
            ...payload,
            createdAt: '2026-08-21T10:00:00Z'
          }
        }
      }
      vi.mocked(apiClient.put).mockResolvedValueOnce(mockResponse)

      const result = await userApi.updateUser('uuid-123', payload)

      expect(apiClient.put).toHaveBeenCalledWith('/users/uuid-123', payload)
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('deleteUser', () => {
    it('calls DELETE /users/:id', async () => {
      const mockResponse = {
        data: { message: 'Usuari donat de baixa correctament.' }
      }
      vi.mocked(apiClient.delete).mockResolvedValueOnce(mockResponse)

      const result = await userApi.deleteUser('uuid-123')

      expect(apiClient.delete).toHaveBeenCalledWith('/users/uuid-123')
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('resetUserPassword', () => {
    it('calls POST /users/:id/password with newPassword', async () => {
      const payload: ResetPasswordRequest = {
        newPassword: 'NovaContrasenya456!'
      }
      const mockResponse = {
        data: { message: 'Contrasenya actualitzada correctament.' }
      }
      vi.mocked(apiClient.post).mockResolvedValueOnce(mockResponse)

      const result = await userApi.resetUserPassword('uuid-123', payload)

      expect(apiClient.post).toHaveBeenCalledWith('/users/uuid-123/password', payload)
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('batchCreateUsers', () => {
    it('calls POST /users/batch with batch items array', async () => {
      const payload: BatchCreateUsersRequest = {
        users: [
          {
            email: 'alumne1@encertia.cat',
            firstName: 'Alumne',
            lastName: 'U',
            role: 'student'
          }
        ]
      }
      const mockResponse = {
        data: {
          totalRequested: 1,
          createdCount: 1,
          failedCount: 0,
          createdUsers: [
            {
              id: 'u-1',
              email: 'alumne1@encertia.cat',
              firstName: 'Alumne',
              lastName: 'U',
              role: 'student',
              isActive: true,
              createdAt: '2026-08-21T10:00:00Z'
            }
          ],
          errors: []
        }
      }
      vi.mocked(apiClient.post).mockResolvedValueOnce(mockResponse)

      const result = await userApi.batchCreateUsers(payload)

      expect(apiClient.post).toHaveBeenCalledWith('/users/batch', payload)
      expect(result).toEqual(mockResponse.data)
    })
  })
})
