import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../store'
import * as authApi from '../api'

vi.mock('../api', () => ({
  login: vi.fn(),
  register: vi.fn(),
  logout: vi.fn(),
  getCurrentUser: vi.fn()
}))

describe('Auth Store (useAuthStore)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('initializes with default values including null deviceToken', () => {
    const store = useAuthStore()
    expect(store.user).toBeNull()
    expect(store.accessToken).toBeNull()
    expect(store.refreshToken).toBeNull()
    expect(store.deviceToken).toBeNull()
    expect(store.isAuthenticated).toBe(false)
  })

  it('sets and clears deviceToken correctly', () => {
    const store = useAuthStore()
    store.setDeviceToken('dbt_token_abc')
    expect(store.deviceToken).toBe('dbt_token_abc')

    store.clearAuth()
    expect(store.deviceToken).toBeNull()
  })

  it('transmits deviceToken from credentials during login', async () => {
    const mockAuthResponse = {
      user: {
        id: 'u1',
        email: 'teacher@encertia.cat',
        firstName: 'Joan',
        lastName: 'Docent',
        role: 'teacher' as const,
        createdAt: '2026-08-21T10:00:00Z'
      },
      tokens: {
        accessToken: 'access_1',
        refreshToken: 'refresh_1',
        tokenType: 'Bearer',
        expiresIn: 900
      }
    }
    vi.mocked(authApi.login).mockResolvedValueOnce(mockAuthResponse)

    const store = useAuthStore()
    const result = await store.login({
      email: 'teacher@encertia.cat',
      password: 'password123',
      deviceToken: 'dbt_from_credentials'
    })

    expect(authApi.login).toHaveBeenCalledWith({
      email: 'teacher@encertia.cat',
      password: 'password123',
      deviceToken: 'dbt_from_credentials'
    })
    expect(result).toEqual(mockAuthResponse)
    expect(store.user).toEqual(mockAuthResponse.user)
    expect(store.accessToken).toBe('access_1')
    expect(store.isAuthenticated).toBe(true)
  })

  it('transmits deviceToken from store state if omitted in credentials', async () => {
    const mockAuthResponse = {
      user: {
        id: 'u1',
        email: 'teacher@encertia.cat',
        firstName: 'Joan',
        lastName: 'Docent',
        role: 'teacher' as const,
        createdAt: '2026-08-21T10:00:00Z'
      },
      tokens: {
        accessToken: 'access_1',
        refreshToken: 'refresh_1',
        tokenType: 'Bearer',
        expiresIn: 900
      }
    }
    vi.mocked(authApi.login).mockResolvedValueOnce(mockAuthResponse)

    const store = useAuthStore()
    store.setDeviceToken('dbt_from_store_state')

    await store.login({
      email: 'teacher@encertia.cat',
      password: 'password123'
    })

    expect(authApi.login).toHaveBeenCalledWith({
      email: 'teacher@encertia.cat',
      password: 'password123',
      deviceToken: 'dbt_from_store_state'
    })
  })

  it('does not send deviceToken when not present in credentials nor store', async () => {
    const mockAuthResponse = {
      user: {
        id: 'u1',
        email: 'student@encertia.cat',
        firstName: 'Student',
        lastName: 'One',
        role: 'student' as const,
        createdAt: '2026-08-21T10:00:00Z'
      },
      tokens: {
        accessToken: 'access_2',
        refreshToken: 'refresh_2',
        tokenType: 'Bearer',
        expiresIn: 900
      }
    }
    vi.mocked(authApi.login).mockResolvedValueOnce(mockAuthResponse)

    const store = useAuthStore()
    await store.login({
      email: 'student@encertia.cat',
      password: 'password123'
    })

    expect(authApi.login).toHaveBeenCalledWith({
      email: 'student@encertia.cat',
      password: 'password123'
    })
  })

  it('handles login failure and stores error message', async () => {
    vi.mocked(authApi.login).mockRejectedValueOnce({
      response: {
        data: {
          error: {
            code: 'DEVICE_AUTHORIZATION_REQUIRED',
            message: 'Aquest compte de docent requereix un dispositiu autoritzat.'
          }
        }
      }
    })

    const store = useAuthStore()
    await expect(
      store.login({
        email: 'teacher@encertia.cat',
        password: 'wrong'
      })
    ).rejects.toBeDefined()

    expect(store.error).toBe('Aquest compte de docent requereix un dispositiu autoritzat.')
    expect(store.isLoading).toBe(false)
  })
})
