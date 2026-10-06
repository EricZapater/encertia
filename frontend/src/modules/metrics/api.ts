/**
 * Crides HTTP per al mòdul de mètriques i auditoria (Metrics).
 * Utilitza el client Axios centralitzat src/api/client.ts i respecta contracts/metrics.openapi.yaml.
 */
import apiClient from '@/api/client'
import type {
  MetricsSummaryResponse,
  ApiLatencyMetricsResponse,
  AuditLogsPaginatedResponse,
  AuditLogQueryParams
} from './types'

/**
 * Obté el resum global de mètriques del sistema, usuaris actius i salut del servidor
 */
export async function getMetricsSummary(): Promise<MetricsSummaryResponse> {
  const response = await apiClient.get<MetricsSummaryResponse>('/metrics/summary')
  return response.data
}

/**
 * Obté la latència mitjana, p95, p99 i taxa d'errors per endpoint
 */
export async function getApiLatencyMetrics(): Promise<ApiLatencyMetricsResponse> {
  const response = await apiClient.get<ApiLatencyMetricsResponse>('/metrics/api-latency')
  return response.data
}

/**
 * Obté la llista paginada de registres d'auditoria d'ús
 */
export async function getAuditLogs(
  params?: AuditLogQueryParams
): Promise<AuditLogsPaginatedResponse> {
  const response = await apiClient.get<AuditLogsPaginatedResponse>('/metrics/audit-logs', {
    params
  })
  return response.data
}

/**
 * Descarrega el registre d'auditoria d'ús en format CSV
 */
export async function exportAuditLogsCSV(): Promise<Blob> {
  const response = await apiClient.get('/metrics/audit-logs/export', {
    responseType: 'blob'
  })
  return response.data
}
