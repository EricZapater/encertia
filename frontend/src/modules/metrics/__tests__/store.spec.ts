import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMetricsStore } from '../store'
import * as metricsApi from '../api'

vi.mock('../api')

describe('useMetricsStore Pinia Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('initial state is correct', () => {
    const store = useMetricsStore()
    expect(store.summary).toBeNull()
    expect(store.apiLatency).toBeNull()
    expect(store.auditLogs).toEqual([])
    expect(store.auditPagination.page).toBe(1)
    expect(store.auditPagination.pageSize).toBe(20)
    expect(store.isLoadingSummary).toBe(false)
    expect(store.isLoadingLatency).toBe(false)
    expect(store.isLoadingAudit).toBe(false)
    expect(store.error).toBeNull()
  })

  it('fetchSummary updates summary state', async () => {
    const mockSummary = {
      activeUsersToday: 10,
      totalRequestsToday: 500,
      avgResponseTimeMs: 30.5,
      errorRatePercentage: 0,
      matchesPlayedTotal: 5,
      materialsReadTotal: 20,
      systemHealth: {
        uptimeSeconds: 3600,
        goroutines: 12,
        memoryUsageMb: 64,
        dbOpenConnections: 5,
        dbInUseConnections: 1
      }
    }
    vi.mocked(metricsApi.getMetricsSummary).mockResolvedValue(mockSummary)

    const store = useMetricsStore()
    await store.fetchSummary()

    expect(metricsApi.getMetricsSummary).toHaveBeenCalled()
    expect(store.summary).toEqual(mockSummary)
    expect(store.hasSummary).toBe(true)
  })

  it('fetchApiLatency updates apiLatency state', async () => {
    const mockLatency = {
      endpoints: [
        {
          method: 'GET',
          endpoint: '/metrics/summary',
          requestCount: 50,
          avgDurationMs: 15.2,
          p95DurationMs: 35.0,
          p99DurationMs: 50.0,
          errorCount: 0
        }
      ]
    }
    vi.mocked(metricsApi.getApiLatencyMetrics).mockResolvedValue(mockLatency)

    const store = useMetricsStore()
    await store.fetchApiLatency()

    expect(metricsApi.getApiLatencyMetrics).toHaveBeenCalled()
    expect(store.apiLatency).toEqual(mockLatency)
    expect(store.latencyEndpoints).toHaveLength(1)
  })

  it('fetchAuditLogs updates auditLogs and auditPagination', async () => {
    const mockAudit = {
      items: [
        {
          id: 'log-1',
          userId: 'u-1',
          userEmail: 'admin@encertia.cat',
          userRole: 'admin',
          action: 'view_metrics',
          module: 'metrics',
          endpoint: '/metrics/summary',
          method: 'GET',
          statusCode: 200,
          durationMs: 15,
          ipAddress: '127.0.0.1',
          createdAt: '2026-09-02T12:00:00Z'
        }
      ],
      total: 1,
      page: 1,
      pageSize: 20,
      totalPages: 1
    }
    vi.mocked(metricsApi.getAuditLogs).mockResolvedValue(mockAudit)

    const store = useMetricsStore()
    await store.fetchAuditLogs()

    expect(metricsApi.getAuditLogs).toHaveBeenCalled()
    expect(store.auditLogs).toEqual(mockAudit.items)
    expect(store.totalAuditLogs).toBe(1)
  })

  it('exportAuditLogs calls exportAuditLogsCSV api method', async () => {
    const mockBlob = new Blob(['csv'], { type: 'text/csv' })
    vi.mocked(metricsApi.exportAuditLogsCSV).mockResolvedValue(mockBlob)

    // Mock URL methods for browser environment in Vitest
    window.URL.createObjectURL = vi.fn().mockReturnValue('blob:http://localhost/123')
    window.URL.revokeObjectURL = vi.fn()

    const store = useMetricsStore()
    await store.exportAuditLogs()

    expect(metricsApi.exportAuditLogsCSV).toHaveBeenCalled()
    expect(window.URL.createObjectURL).toHaveBeenCalledWith(mockBlob)
  })
})
