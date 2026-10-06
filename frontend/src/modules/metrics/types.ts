/**
 * Tipus TypeScript per al mòdul de mètriques i auditoria (Metrics).
 * Transcrits fidelment del contracte OpenAPI contracts/metrics.openapi.yaml
 */

export interface SystemHealth {
  uptimeSeconds: number
  goroutines: number
  memoryUsageMb: number
  dbOpenConnections: number
  dbInUseConnections: number
}

export interface MetricsSummaryResponse {
  activeUsersToday: number
  totalRequestsToday: number
  avgResponseTimeMs: number
  errorRatePercentage: number
  matchesPlayedTotal: number
  materialsReadTotal: number
  systemHealth: SystemHealth
}

export interface EndpointLatencyItem {
  method: string
  endpoint: string
  requestCount: number
  avgDurationMs: number
  p95DurationMs: number
  p99DurationMs: number
  errorCount: number
}

export interface ApiLatencyMetricsResponse {
  endpoints: EndpointLatencyItem[]
}

export interface AuditLogItem {
  id: string
  userId?: string
  userEmail?: string
  userRole?: string
  action: string
  module: string
  endpoint: string
  method: string
  statusCode: number
  durationMs: number
  ipAddress?: string
  createdAt: string
}

export interface AuditLogsPaginatedResponse {
  items: AuditLogItem[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface AuditLogQueryParams {
  page?: number
  pageSize?: number
  search?: string
  module?: string
  userId?: string
}
