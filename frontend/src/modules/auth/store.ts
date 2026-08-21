import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  TokenPair,
  User
} from './types'
import * as authApi from './api'
import { ACCESS_TOKEN_KEY, REFRESH_TOKEN_KEY } from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(localStorage.getItem(ACCESS_TOKEN_KEY))
  const refreshToken = ref<string | null>(localStorage.getItem(REFRESH_TOKEN_KEY))
  const isLoading = ref(false)
  const isInitialized = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const isAuthenticated = computed(() => Boolean(accessToken.value && user.value))
  const currentUser = computed(() => user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isTeacher = computed(() => user.value?.role === 'teacher')
  const isStudent = computed(() => user.value?.role === 'student')
  const fullName = computed(() => {
    if (!user.value) return ''
    return `${user.value.firstName} ${user.value.lastName}`.trim()
  })

  // Actions
  function setTokens(tokens: TokenPair) {
    accessToken.value = tokens.accessToken
    refreshToken.value = tokens.refreshToken
    localStorage.setItem(ACCESS_TOKEN_KEY, tokens.accessToken)
    localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refreshToken)
  }

  function setUser(newUser: User | null) {
    user.value = newUser
  }

  function clearAuth() {
    user.value = null
    accessToken.value = null
    refreshToken.value = null
    error.value = null
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
  }

  async function login(credentials: LoginRequest): Promise<AuthResponse> {
    isLoading.value = true
    error.value = null
    try {
      const response = await authApi.login(credentials)
      setTokens(response.tokens)
      setUser(response.user)
      return response
    } catch (err: any) {
      const errMsg =
        err.response?.data?.error?.message ||
        err.message ||
        'Error en iniciar sessió.'
      error.value = errMsg
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function register(userData: RegisterRequest): Promise<AuthResponse> {
    isLoading.value = true
    error.value = null
    try {
      const response = await authApi.register(userData)
      setTokens(response.tokens)
      setUser(response.user)
      return response
    } catch (err: any) {
      const errMsg =
        err.response?.data?.error?.message ||
        err.message ||
        'Error en registrar el compte.'
      error.value = errMsg
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function logout(): Promise<void> {
    isLoading.value = true
    try {
      if (refreshToken.value) {
        await authApi.logout({ refreshToken: refreshToken.value })
      } else {
        await authApi.logout()
      }
    } catch {
      // Ignorem errors al logout per assegurar neteja local
    } finally {
      clearAuth()
      isLoading.value = false
    }
  }

  async function fetchMe(): Promise<User | null> {
    if (!accessToken.value) {
      clearAuth()
      return null
    }

    isLoading.value = true
    try {
      const response = await authApi.getCurrentUser()
      setUser(response.user)
      return response.user
    } catch (err) {
      clearAuth()
      return null
    } finally {
      isLoading.value = false
    }
  }

  async function initAuth(): Promise<void> {
    if (isInitialized.value) return

    if (accessToken.value) {
      await fetchMe()
    }
    isInitialized.value = true
  }

  return {
    // State
    user,
    accessToken,
    refreshToken,
    isLoading,
    isInitialized,
    error,
    // Getters
    isAuthenticated,
    currentUser,
    isAdmin,
    isTeacher,
    isStudent,
    fullName,
    // Actions
    setTokens,
    setUser,
    clearAuth,
    login,
    register,
    logout,
    fetchMe,
    initAuth
  }
})
