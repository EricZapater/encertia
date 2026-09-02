import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import MetricsDashboardView from '../views/MetricsDashboardView.vue'
import { useMetricsStore } from '../store'

vi.mock('../api', () => ({
  getMetricsSummary: vi.fn().mockResolvedValue({
    activeUsersToday: 25,
    totalRequestsToday: 1200,
    avgResponseTimeMs: 42.5,
    errorRatePercentage: 0.1,
    matchesPlayedTotal: 15,
    materialsReadTotal: 80,
    systemHealth: {
      uptimeSeconds: 7200,
      goroutines: 18,
      memoryUsageMb: 95.4,
      dbOpenConnections: 8,
      dbInUseConnections: 2
    }
  }),
  getApiLatencyMetrics: vi.fn().mockResolvedValue({
    endpoints: [
      {
        method: 'GET',
        endpoint: '/api/v1/metrics/summary',
        requestCount: 100,
        avgDurationMs: 12.0,
        p95DurationMs: 25.0,
        p99DurationMs: 40.0,
        errorCount: 0
      }
    ]
  }),
  getAuditLogs: vi.fn().mockResolvedValue({
    items: [],
    total: 0,
    page: 1,
    pageSize: 20,
    totalPages: 0
  }),
  exportAuditLogsCSV: vi.fn().mockResolvedValue(new Blob(['csv'], { type: 'text/csv' }))
}))

describe('MetricsDashboardView.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders dashboard title and elements correctly', async () => {
    const store = useMetricsStore()

    const wrapper = mount(MetricsDashboardView, {
      global: {
        plugins: [PrimeVue],
        stubs: {
          DataTable: true,
          Column: true,
          InputText: true,
          Select: true,
          Button: true,
          Tag: true
        }
      }
    })

    // Esperar que les promeses del store es resolguin
    await store.fetchAll()

    expect(wrapper.find('[data-testid="metrics-dashboard"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Mètriques i Auditoria del Sistema')
    expect(wrapper.find('[data-testid="btn-export-csv"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="audit-data-table"]').exists()).toBe(true)
  })
})
