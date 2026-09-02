import { describe, it, expect, vi, beforeEach } from 'vitest'
import apiClient from '@/api/client'
import {
  getMetricsSummary,
  getApiLatencyMetrics,
  getAuditLogs,
  exportAuditLogsCSV
} from '../api'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn()
  }
}))

describe('Metrics API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getMetricsSummary calls GET /metrics/summary', async () => {
    const mockSummary = {
      activeUsersToday: 42,
      totalRequestsToday: 1250,
      avgResponseTimeMs: 45.2,
      errorRatePercentage: 0.5,
      matchesPlayedTotal: 18,
      materialsReadTotal: 95,
      systemHealth: {
        uptimeSeconds: 86400,
        goroutines: 24,
        memoryUsageMb: 128.5,
        dbOpenConnections: 10,
        dbInUseConnections: 2
      }
    }
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockSummary })

    const result = await getMetricsSummary()
    expect(apiClient.get).toHaveBeenCalledWith('/metrics/summary')
    expect(result).toEqual(mockSummary)
  })

  it('getApiLatencyMetrics calls GET /metrics/api-latency', async () => {
    const mockLatency = {
      endpoints: [
        {
          method: 'GET',
          endpoint: '/api/v1/courses',
          requestCount: 300,
          avgDurationMs: 25.4,
          p95DurationMs: 80.1,
          p99DurationMs: 150.2,
          errorCount: 0
        }
      ]
    }
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockLatency })

    const result = await getApiLatencyMetrics()
    expect(apiClient.get).toHaveBeenCalledWith('/metrics/api-latency')
    expect(result).toEqual(mockLatency)
  })

  it('getAuditLogs calls GET /metrics/audit-logs with query parameters', async () => {
    const mockAudit = {
      items: [
        {
          id: 'log-1',
          action: 'login',
          module: 'auth',
          endpoint: '/api/v1/auth/login',
          method: 'POST',
          statusCode: 200,
          durationMs: 40,
          createdAt: '2026-09-02T10:00:00Z'
        }
      ],
      total: 1,
      page: 1,
      pageSize: 20,
      totalPages: 1
    }
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockAudit })

    const params = { page: 1, pageSize: 20, search: 'login', module: 'auth' }
    const result = await getAuditLogs(params)

    expect(apiClient.get).toHaveBeenCalledWith('/metrics/audit-logs', { params })
    expect(result).toEqual(mockAudit)
  })

  it('exportAuditLogsCSV calls GET /metrics/audit-logs/export with responseType blob', async () => {
    const mockBlob = new Blob(['csv,content'], { type: 'text/csv' })
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockBlob })

    const result = await exportAuditLogsCSV()
    expect(apiClient.get).toHaveBeenCalledWith('/metrics/audit-logs/export', {
      responseType: 'blob'
    })
    expect(result).toEqual(mockBlob)
  })
})
