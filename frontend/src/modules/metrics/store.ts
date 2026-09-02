/**
 * Pinia Store per al mòdul de mètriques i auditoria (useMetricsStore).
 * Carrega el resum de mètriques, latències d'API, registres d'auditoria paginats i exportació CSV.
 */
import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'
import type {
  MetricsSummaryResponse,
  ApiLatencyMetricsResponse,
  AuditLogItem,
  AuditLogsPaginatedResponse,
  AuditLogQueryParams
} from './types'
import * as metricsApi from './api'

export const useMetricsStore = defineStore('metrics', () => {
  // State
  const summary = ref<MetricsSummaryResponse | null>(null)
  const apiLatency = ref<ApiLatencyMetricsResponse | null>(null)
  const auditLogs = ref<AuditLogItem[]>([])

  const auditPagination = reactive({
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0
  })

  const auditFilters = reactive<AuditLogQueryParams>({
    search: '',
    module: undefined,
    userId: undefined
  })

  const isLoadingSummary = ref(false)
  const isLoadingLatency = ref(false)
  const isLoadingAudit = ref(false)
  const isExporting = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const hasSummary = computed(() => summary.value !== null)
  const latencyEndpoints = computed(() => apiLatency.value?.endpoints || [])
  const totalAuditLogs = computed(() => auditPagination.total)
  const currentAuditPage = computed(() => auditPagination.page)
  const auditPageSize = computed(() => auditPagination.pageSize)
  const auditTotalPages = computed(() => auditPagination.totalPages)
  const isLoadingAny = computed(
    () => isLoadingSummary.value || isLoadingLatency.value || isLoadingAudit.value
  )

  // Helper per a missatges d'error
  function extractErrorMessage(err: any, defaultMsg: string): string {
    return (
      err.response?.data?.error?.message ||
      err.response?.data?.message ||
      err.message ||
      defaultMsg
    )
  }

  // Actions
  async function fetchSummary(): Promise<void> {
    isLoadingSummary.value = true
    error.value = null
    try {
      summary.value = await metricsApi.getMetricsSummary()
    } catch (err: any) {
      error.value = extractErrorMessage(err, 'Error en carregar el resum de mètriques.')
      throw err
    } finally {
      isLoadingSummary.value = false
    }
  }

  async function fetchApiLatency(): Promise<void> {
    isLoadingLatency.value = true
    error.value = null
    try {
      apiLatency.value = await metricsApi.getApiLatencyMetrics()
    } catch (err: any) {
      error.value = extractErrorMessage(err, "Error en carregar les mètriques de latència d'API.")
      throw err
    } finally {
      isLoadingLatency.value = false
    }
  }

  async function fetchAuditLogs(paramsOverride?: AuditLogQueryParams): Promise<void> {
    isLoadingAudit.value = true
    error.value = null
    try {
      const params: AuditLogQueryParams = {
        page: auditPagination.page,
        pageSize: auditPagination.pageSize,
        ...(auditFilters.search?.trim() ? { search: auditFilters.search.trim() } : {}),
        ...(auditFilters.module ? { module: auditFilters.module } : {}),
        ...(auditFilters.userId ? { userId: auditFilters.userId } : {}),
        ...paramsOverride
      }

      const response: AuditLogsPaginatedResponse = await metricsApi.getAuditLogs(params)
      auditLogs.value = response.items
      auditPagination.page = response.page
      auditPagination.pageSize = response.pageSize
      auditPagination.total = response.total
      auditPagination.totalPages = response.totalPages
    } catch (err: any) {
      error.value = extractErrorMessage(err, "Error en carregar els registres d'auditoria.")
      throw err
    } finally {
      isLoadingAudit.value = false
    }
  }

  async function fetchAll(): Promise<void> {
    await Promise.all([fetchSummary(), fetchApiLatency(), fetchAuditLogs()])
  }

  function setAuditPage(page: number, newPageSize?: number): Promise<void> {
    auditPagination.page = Math.max(1, page)
    if (newPageSize && newPageSize > 0) {
      auditPagination.pageSize = newPageSize
    }
    return fetchAuditLogs()
  }

  function setAuditSearch(searchQuery: string): Promise<void> {
    auditFilters.search = searchQuery
    auditPagination.page = 1
    return fetchAuditLogs()
  }

  function setAuditModuleFilter(moduleName?: string): Promise<void> {
    auditFilters.module = moduleName
    auditPagination.page = 1
    return fetchAuditLogs()
  }

  function resetAuditFilters(): Promise<void> {
    auditFilters.search = ''
    auditFilters.module = undefined
    auditFilters.userId = undefined
    auditPagination.page = 1
    return fetchAuditLogs()
  }

  async function exportAuditLogs(): Promise<void> {
    isExporting.value = true
    error.value = null
    try {
      const blob = await metricsApi.exportAuditLogsCSV()
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `audit_logs_${new Date().toISOString().slice(0, 10)}.csv`)
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
    } catch (err: any) {
      error.value = extractErrorMessage(err, "Error en exportar els registres d'auditoria en CSV.")
      throw err
    } finally {
      isExporting.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    summary,
    apiLatency,
    auditLogs,
    auditPagination,
    auditFilters,
    isLoadingSummary,
    isLoadingLatency,
    isLoadingAudit,
    isExporting,
    error,
    // Getters
    hasSummary,
    latencyEndpoints,
    totalAuditLogs,
    currentAuditPage,
    auditPageSize,
    auditTotalPages,
    isLoadingAny,
    // Actions
    fetchSummary,
    fetchApiLatency,
    fetchAuditLogs,
    fetchAll,
    setAuditPage,
    setAuditSearch,
    setAuditModuleFilter,
    resetAuditFilters,
    exportAuditLogs,
    clearError
  }
})
