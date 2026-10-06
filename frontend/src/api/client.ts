import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import type { RefreshTokenRequest, TokenPair } from '@/modules/auth/types'

export const ACCESS_TOKEN_KEY = 'encertia_access_token'
export const REFRESH_TOKEN_KEY = 'encertia_refresh_token'

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  headers: {
    'Content-Type': 'application/json'
  }
})

// Variables per gestionar la cua de peticions durant el refresc de token
let isRefreshing = false
let failedQueue: Array<{
  resolve: (token: string) => void
  reject: (error: unknown) => void
}> = []

const processQueue = (error: unknown, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else if (token) {
      prom.resolve(token)
    }
  })
  failedQueue = []
}

// Request Interceptor: Injecta el Bearer token si està disponible
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem(ACCESS_TOKEN_KEY)
    if (token && !config.headers.Authorization) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response Interceptor: Gestiona errors 401 i renovació automàtica del token
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    if (!originalRequest) {
      return Promise.reject(error)
    }

    // Evitar bucle si l'error 401 ve de login o refresh
    const isAuthEndpoint =
      originalRequest.url?.includes('/auth/login') ||
      originalRequest.url?.includes('/auth/refresh') ||
      originalRequest.url?.includes('/auth/register')

    if (error.response?.status === 401 && !originalRequest._retry && !isAuthEndpoint) {
      if (isRefreshing) {
        // Si ja s'està renovant el token, afegim la petició a la cua
        return new Promise<string>((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
          .then((token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            return apiClient(originalRequest)
          })
          .catch((err) => Promise.reject(err))
      }

      originalRequest._retry = true
      isRefreshing = true

      const currentRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
      if (!currentRefreshToken) {
        isRefreshing = false
        localStorage.removeItem(ACCESS_TOKEN_KEY)
        localStorage.removeItem(REFRESH_TOKEN_KEY)
        return Promise.reject(error)
      }

      try {
        const refreshPayload: RefreshTokenRequest = {
          refreshToken: currentRefreshToken
        }

        // Crida directa a l'endpoint de refresc sense passar per l'interceptor
        const response = await axios.post<TokenPair>(
          `${apiClient.defaults.baseURL}/auth/refresh`,
          refreshPayload,
          {
            headers: { 'Content-Type': 'application/json' }
          }
        )

        const { accessToken, refreshToken: newRefreshToken } = response.data
        localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
        if (newRefreshToken) {
          localStorage.setItem(REFRESH_TOKEN_KEY, newRefreshToken)
        }

        apiClient.defaults.headers.common.Authorization = `Bearer ${accessToken}`
        processQueue(null, accessToken)

        originalRequest.headers.Authorization = `Bearer ${accessToken}`
        return apiClient(originalRequest)
      } catch (refreshError) {
        processQueue(refreshError, null)
        localStorage.removeItem(ACCESS_TOKEN_KEY)
        localStorage.removeItem(REFRESH_TOKEN_KEY)
        delete apiClient.defaults.headers.common.Authorization

        // Opcionalment redirigir al login si no som al login
        if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
          window.location.href = '/login'
        }

        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  }
)

export default apiClient
