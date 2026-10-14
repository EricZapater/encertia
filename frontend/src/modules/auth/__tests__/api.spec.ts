import { describe, it, expect, vi, beforeEach } from 'vitest'
import apiClient from '@/api/client'
import * as authApi from '../api'
import type { LoginRequest, RegisterRequest } from '../types'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn()
  }
}))

describe('Auth API client functions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('login', () => {
    it('calls POST /auth/login with email and password when deviceToken is not provided', async () => {
      const mockResponse = {
        data: {
          user: {
            id: 'u1',
            email: 'teacher@encertia.cat',
            firstName: 'Joan',
            lastName: 'Docent',
            role: 'teacher',
            createdAt: '2026-08-21T10:00:00Z'
          },
          tokens: {
            accessToken: 'acc_123',
            refreshToken: 'ref_123',
            tokenType: 'Bearer',
            expiresIn: 900
          }
        }
      }
      vi.mocked(apiClient.post).mockResolvedValueOnce(mockResponse)

      const payload: LoginRequest = {
        email: 'teacher@encertia.cat',
        password: 'password123'
      }

      const result = await authApi.login(payload)

      expect(apiClient.post).toHaveBeenCalledWith('/auth/login', {
        email: 'teacher@encertia.cat',
        password: 'password123'
      })
      expect(result).toEqual(mockResponse.data)
    })

    it('calls POST /auth/login with deviceToken included when present', async () => {
      const mockResponse = {
        data: {
          user: {
            id: 'u1',
            email: 'teacher@encertia.cat',
            firstName: 'Joan',
            lastName: 'Docent',
            role: 'teacher',
            createdAt: '2026-08-21T10:00:00Z'
          },
          tokens: {
            accessToken: 'acc_123',
            refreshToken: 'ref_123',
            tokenType: 'Bearer',
            expiresIn: 900
          }
        }
      }
      vi.mocked(apiClient.post).mockResolvedValueOnce(mockResponse)

      const payload: LoginRequest = {
        email: 'teacher@encertia.cat',
        password: 'password123',
        deviceToken: 'dbt_test_token_123'
      }

      const result = await authApi.login(payload)

      expect(apiClient.post).toHaveBeenCalledWith('/auth/login', {
        email: 'teacher@encertia.cat',
        password: 'password123',
        deviceToken: 'dbt_test_token_123'
      })
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('register', () => {
    it('calls POST /auth/register with user payload', async () => {
      const mockResponse = {
        data: {
          user: {
            id: 'u2',
            email: 'student@encertia.cat',
            firstName: 'Alumne',
            lastName: 'Nou',
            role: 'student',
            createdAt: '2026-08-21T10:00:00Z'
          },
          tokens: {
            accessToken: 'acc_456',
            refreshToken: 'ref_456',
            tokenType: 'Bearer',
            expiresIn: 900
          }
        }
      }
      vi.mocked(apiClient.post).mockResolvedValueOnce(mockResponse)

      const payload: RegisterRequest = {
        email: 'student@encertia.cat',
        password: 'password123',
        firstName: 'Alumne',
        lastName: 'Nou',
        role: 'student'
      }

      const result = await authApi.register(payload)

      expect(apiClient.post).toHaveBeenCalledWith('/auth/register', payload)
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('refreshToken', () => {
    it('calls POST /auth/refresh with refreshToken', async () => {
      const mockResponse = {
        data: {
          accessToken: 'acc_new',
          refreshToken: 'ref_new',
          tokenType: 'Bearer',
          expiresIn: 900
        }
      }
      vi.mocked(apiClient.post).mockResolvedValueOnce(mockResponse)

      const result = await authApi.refreshToken({ refreshToken: 'ref_old' })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/refresh', { refreshToken: 'ref_old' })
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('logout', () => {
    it('calls POST /auth/logout with refreshToken', async () => {
      const mockResponse = { data: { message: 'Logged out' } }
      vi.mocked(apiClient.post).mockResolvedValueOnce(mockResponse)

      const result = await authApi.logout({ refreshToken: 'ref_123' })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/logout', { refreshToken: 'ref_123' })
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('getCurrentUser', () => {
    it('calls GET /auth/me', async () => {
      const mockResponse = {
        data: {
          user: {
            id: 'u1',
            email: 'test@encertia.cat',
            firstName: 'Test',
            lastName: 'User',
            role: 'teacher',
            createdAt: '2026-08-21T10:00:00Z'
          }
        }
      }
      vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse)

      const result = await authApi.getCurrentUser()

      expect(apiClient.get).toHaveBeenCalledWith('/auth/me')
      expect(result).toEqual(mockResponse.data)
    })
  })
})
