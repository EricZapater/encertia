/**
 * Pinia Store per a la gestió d'usuaris (useUserStore).
 * Gestiona el cicle de vida dels usuaris, llistat paginat, filtres reactius,
 * accions CRUD, reseteig de contrasenya i importació batch.
 */
import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'
import type {
  BatchCreateUsersRequest,
  BatchCreateUsersResponse,
  CreateUserRequest,
  PaginationMetadata,
  ResetPasswordRequest,
  UpdateUserRequest,
  User,
  UserListParams,
  UserRole,
  UserStatusFilter
} from './types'
import * as userApi from './api'

export const useUserStore = defineStore('users', () => {
  // State
  const users = ref<User[]>([])
  const selectedUser = ref<User | null>(null)
  const pagination = reactive<PaginationMetadata>({
    page: 1,
    pageSize: 20,
    totalCount: 0,
    totalPages: 0
  })

  const filters = reactive<{
    search: string
    role?: UserRole
    status: UserStatusFilter
  }>({
    search: '',
    role: undefined,
    status: 'active'
  })

  const isLoading = ref(false)
  const isSubmitting = ref(false)
  const error = ref<string | null>(null)
  const lastBatchResult = ref<BatchCreateUsersResponse | null>(null)

  // Getters
  const userList = computed(() => users.value)
  const totalUsers = computed(() => pagination.totalCount)
  const currentPage = computed(() => pagination.page)
  const pageSize = computed(() => pagination.pageSize)
  const totalPages = computed(() => pagination.totalPages)
  const hasUsers = computed(() => users.value.length > 0)

  // Helper per extreure missatge d'error amigable
  function extractErrorMessage(err: any, defaultMsg: string): string {
    return (
      err.response?.data?.error?.message ||
      err.response?.data?.message ||
      err.message ||
      defaultMsg
    )
  }

  // Actions
  async function fetchUsers(paramsOverride?: UserListParams): Promise<void> {
    isLoading.value = true
    error.value = null

    try {
      const params: UserListParams = {
        page: pagination.page,
        pageSize: pagination.pageSize,
        status: filters.status,
        ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
        ...(filters.role ? { role: filters.role } : {}),
        ...paramsOverride
      }

      const response = await userApi.listUsers(params)
      users.value = response.items
      pagination.page = response.pagination.page
      pagination.pageSize = response.pagination.pageSize
      pagination.totalCount = response.pagination.totalCount
      pagination.totalPages = response.pagination.totalPages
    } catch (err: any) {
      error.value = extractErrorMessage(err, 'Error en carregar el llistat d’usuaris.')
      throw err
    } finally {
      isLoading.value = false
    }
  }

  function setPage(page: number, newPageSize?: number): Promise<void> {
    pagination.page = Math.max(1, page)
    if (newPageSize && newPageSize > 0) {
      pagination.pageSize = newPageSize
    }
    return fetchUsers()
  }

  function setSearch(query: string): Promise<void> {
    filters.search = query
    pagination.page = 1
    return fetchUsers()
  }

  function setRoleFilter(role?: UserRole): Promise<void> {
    filters.role = role
    pagination.page = 1
    return fetchUsers()
  }

  function setStatusFilter(status: UserStatusFilter): Promise<void> {
    filters.status = status
    pagination.page = 1
    return fetchUsers()
  }

  function resetFilters(): Promise<void> {
    filters.search = ''
    filters.role = undefined
    filters.status = 'active'
    pagination.page = 1
    return fetchUsers()
  }

  async function fetchUserById(id: string): Promise<User> {
    isLoading.value = true
    error.value = null
    try {
      const response = await userApi.getUserById(id)
      selectedUser.value = response.user
      return response.user
    } catch (err: any) {
      error.value = extractErrorMessage(err, 'Error en consultar les dades de l’usuari.')
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function createUser(data: CreateUserRequest): Promise<User> {
    isSubmitting.value = true
    error.value = null
    try {
      const response = await userApi.createUser(data)
      // Refresca la pàgina actual per mantenir la coherència amb la paginació
      await fetchUsers()
      return response.user
    } catch (err: any) {
      error.value = extractErrorMessage(err, 'Error en crear l’usuari.')
      throw err
    } finally {
      isSubmitting.value = false
    }
  }

  async function updateUser(id: string, data: UpdateUserRequest): Promise<User> {
    isSubmitting.value = true
    error.value = null
    try {
      const response = await userApi.updateUser(id, data)
      // Actualitza l'element localment si existeix
      const index = users.value.findIndex((u) => u.id === id)
      if (index !== -1) {
        users.value[index] = response.user
      }
      if (selectedUser.value && selectedUser.value.id === id) {
        selectedUser.value = response.user
      }
      return response.user
    } catch (err: any) {
      error.value = extractErrorMessage(err, 'Error en actualitzar l’usuari.')
      throw err
    } finally {
      isSubmitting.value = false
    }
  }

  async function deleteUser(id: string): Promise<string> {
    isSubmitting.value = true
    error.value = null
    try {
      const response = await userApi.deleteUser(id)
      await fetchUsers()
      return response.message
    } catch (err: any) {
      error.value = extractErrorMessage(err, 'Error en donar de baixa l’usuari.')
      throw err
    } finally {
      isSubmitting.value = false
    }
  }

  async function resetUserPassword(id: string, data: ResetPasswordRequest): Promise<string> {
    isSubmitting.value = true
    error.value = null
    try {
      const response = await userApi.resetUserPassword(id, data)
      return response.message
    } catch (err: any) {
      error.value = extractErrorMessage(err, 'Error en restablir la contrasenya.')
      throw err
    } finally {
      isSubmitting.value = false
    }
  }

  async function batchImport(data: BatchCreateUsersRequest): Promise<BatchCreateUsersResponse> {
    isSubmitting.value = true
    error.value = null
    try {
      const response = await userApi.batchCreateUsers(data)
      lastBatchResult.value = response
      await fetchUsers()
      return response
    } catch (err: any) {
      error.value = extractErrorMessage(err, 'Error en processar la importació massiva.')
      throw err
    } finally {
      isSubmitting.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  function clearSelectedUser() {
    selectedUser.value = null
  }

  return {
    // State
    users,
    selectedUser,
    pagination,
    filters,
    isLoading,
    isSubmitting,
    error,
    lastBatchResult,
    // Getters
    userList,
    totalUsers,
    currentPage,
    pageSize,
    totalPages,
    hasUsers,
    // Actions
    fetchUsers,
    setPage,
    setSearch,
    setRoleFilter,
    setStatusFilter,
    resetFilters,
    fetchUserById,
    createUser,
    updateUser,
    deleteUser,
    resetUserPassword,
    batchImport,
    clearError,
    clearSelectedUser
  }
})
