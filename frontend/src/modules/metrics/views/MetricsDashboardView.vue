<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useMetricsStore } from '../store'
import type { AuditLogItem, EndpointLatencyItem } from '../types'

import DataTable, { type DataTablePageEvent } from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { useI18n } from 'vue-i18n'

const metricsStore = useMetricsStore()
const toast = useToast()
const { t } = useI18n()

// Filtre de cerca local amb debounce
const auditSearchInput = ref('')
let auditSearchTimeout: ReturnType<typeof setTimeout> | null = null

const selectedModule = ref<string | undefined>(undefined)

const moduleFilterOptions = computed(() => [
  { label: t('metrics.audit.allModules'), value: undefined },
  { label: 'auth', value: 'auth' },
  { label: 'users', value: 'users' },
  { label: 'quizzes', value: 'quizzes' },
  { label: 'courses', value: 'courses' },
  { label: 'materials', value: 'materials' },
  { label: 'match', value: 'match' },
  { label: 'evaluations', value: 'evaluations' },
  { label: 'metrics', value: 'metrics' }
])

watch(
  () => metricsStore.error,
  (err) => {
    if (err) {
      toast.add({ severity: 'error', summary: t('common.error'), detail: err, life: 4000 })
      metricsStore.clearError()
    }
  }
)

onMounted(() => {
  metricsStore.fetchAll()
})

function handleAuditSearchInput() {
  if (auditSearchTimeout) clearTimeout(auditSearchTimeout)
  auditSearchTimeout = setTimeout(() => {
    metricsStore.setAuditSearch(auditSearchInput.value)
  }, 350)
}

function handleModuleChange() {
  metricsStore.setAuditModuleFilter(selectedModule.value)
}

function handleAuditPageChange(event: DataTablePageEvent) {
  const newPage = Math.floor(event.first / event.rows) + 1
  metricsStore.setAuditPage(newPage, event.rows)
}

function handleResetAuditFilters() {
  auditSearchInput.value = ''
  selectedModule.value = undefined
  metricsStore.resetAuditFilters()
}

async function handleExportCSV() {
  try {
    await metricsStore.exportAuditLogs()
    toast.add({
      severity: 'success',
      summary: t('common.success'),
      detail: t('metrics.audit.exportSuccess', 'Fitxer CSV exportat amb èxit.'),
      life: 3000
    })
  } catch {
    // Error is handled via store watcher
  }
}

function formatUptime(seconds?: number): string {
  if (!seconds || seconds <= 0) return '0s'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60

  const parts: string[] = []
  if (days > 0) parts.push(`${days}d`)
  if (hours > 0) parts.push(`${hours}h`)
  if (mins > 0) parts.push(`${mins}m`)
  if (parts.length === 0 || secs > 0) parts.push(`${secs}s`)
  return parts.join(' ')
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleString('ca-ES', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  } catch {
    return dateStr
  }
}

function getMethodSeverity(method: string) {
  switch (method?.toUpperCase()) {
    case 'GET':
      return 'success'
    case 'POST':
      return 'info'
    case 'PUT':
    case 'PATCH':
      return 'warn'
    case 'DELETE':
      return 'danger'
    default:
      return 'secondary'
  }
}

function getStatusSeverity(status: number) {
  if (status >= 200 && status < 300) return 'success'
  if (status >= 300 && status < 400) return 'info'
  if (status >= 400 && status < 500) return 'warn'
  if (status >= 500) return 'danger'
  return 'secondary'
}
</script>

<template>
  <div class="metrics-view-container" data-testid="metrics-dashboard">
    <!-- Capçalera de pàgina -->
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ $t('metrics.title') }}</h1>
        <p class="page-subtitle">{{ $t('metrics.subtitle') }}</p>
      </div>

      <div class="header-actions">
        <Button
          icon="pi pi-refresh"
          label="Refrescar Dades"
          severity="secondary"
          outlined
          :loading="metricsStore.isLoadingAny"
          @click="metricsStore.fetchAll()"
          data-testid="btn-refresh-metrics"
        />
      </div>
    </div>

    <!-- Mètriques de Resum Global (Summary Cards) -->
    <div class="metrics-grid">
      <div class="metric-card">
        <div class="card-icon icon-users"><i class="pi pi-users" /></div>
        <div class="card-content">
          <span class="card-label">{{ $t('metrics.summary.activeUsers') }}</span>
          <span class="card-value">{{ metricsStore.summary?.activeUsersToday ?? '-' }}</span>
        </div>
      </div>

      <div class="metric-card">
        <div class="card-icon icon-requests"><i class="pi pi-arrow-right-arrow-left" /></div>
        <div class="card-content">
          <span class="card-label">{{ $t('metrics.summary.totalRequests') }}</span>
          <span class="card-value">{{ metricsStore.summary?.totalRequestsToday ?? '-' }}</span>
        </div>
      </div>

      <div class="metric-card">
        <div class="card-icon icon-latency"><i class="pi pi-clock" /></div>
        <div class="card-content">
          <span class="card-label">{{ $t('metrics.summary.avgResponseTime') }}</span>
          <span class="card-value">
            {{ metricsStore.summary?.avgResponseTimeMs?.toFixed(1) ?? '-' }} ms
          </span>
        </div>
      </div>

      <div class="metric-card">
        <div class="card-icon icon-errors"><i class="pi pi-exclamation-triangle" /></div>
        <div class="card-content">
          <span class="card-label">{{ $t('metrics.summary.errorRate') }}</span>
          <span class="card-value" :class="{ 'text-danger': (metricsStore.summary?.errorRatePercentage ?? 0) > 5 }">
            {{ metricsStore.summary?.errorRatePercentage?.toFixed(2) ?? '-' }}%
          </span>
        </div>
      </div>

      <div class="metric-card">
        <div class="card-icon icon-matches"><i class="pi pi-play" /></div>
        <div class="card-content">
          <span class="card-label">{{ $t('metrics.summary.matchesPlayed') }}</span>
          <span class="card-value">{{ metricsStore.summary?.matchesPlayedTotal ?? '-' }}</span>
        </div>
      </div>

      <div class="metric-card">
        <div class="card-icon icon-materials"><i class="pi pi-file" /></div>
        <div class="card-content">
          <span class="card-label">{{ $t('metrics.summary.materialsRead') }}</span>
          <span class="card-value">{{ metricsStore.summary?.materialsReadTotal ?? '-' }}</span>
        </div>
      </div>
    </div>

    <!-- Mètriques de Salut del Servidor (System Health) -->
    <div class="section-card health-section">
      <div class="section-header">
        <div class="section-title-box">
          <i class="pi pi-server section-icon" />
          <h2 class="section-title">{{ $t('metrics.health.title') }}</h2>
        </div>
      </div>

      <div class="health-grid">
        <div class="health-item">
          <span class="health-label">{{ $t('metrics.health.uptime') }}</span>
          <span class="health-value">
            {{ formatUptime(metricsStore.summary?.systemHealth?.uptimeSeconds) }}
          </span>
        </div>
        <div class="health-item">
          <span class="health-label">{{ $t('metrics.health.goroutines') }}</span>
          <span class="health-value">
            {{ metricsStore.summary?.systemHealth?.goroutines ?? '-' }}
          </span>
        </div>
        <div class="health-item">
          <span class="health-label">{{ $t('metrics.health.memory') }}</span>
          <span class="health-value">
            {{ metricsStore.summary?.systemHealth?.memoryUsageMb?.toFixed(1) ?? '-' }} MB
          </span>
        </div>
        <div class="health-item">
          <span class="health-label">{{ $t('metrics.health.dbConnections') }}</span>
          <span class="health-value">
            {{ metricsStore.summary?.systemHealth?.dbInUseConnections ?? 0 }} /
            {{ metricsStore.summary?.systemHealth?.dbOpenConnections ?? 0 }}
          </span>
        </div>
      </div>
    </div>

    <!-- Rànquing de Latència d'API (P95 / P99) -->
    <div class="section-card">
      <div class="section-header">
        <div class="section-title-box">
          <i class="pi pi-chart-bar section-icon" />
          <div>
            <h2 class="section-title">{{ $t('metrics.latency.title') }}</h2>
            <p class="section-subtitle">{{ $t('metrics.latency.subtitle') }}</p>
          </div>
        </div>
      </div>

      <DataTable
        :value="metricsStore.latencyEndpoints"
        :loading="metricsStore.isLoadingLatency"
        stripedRows
        tableStyle="min-width: 50rem"
        data-testid="latency-data-table"
      >
        <template #empty>
          <div class="empty-state">
            <i class="pi pi-inbox empty-icon" />
            <p>No hi ha dades de latència disponibles.</p>
          </div>
        </template>

        <Column field="method" :header="$t('metrics.latency.method')" style="width: 7rem">
          <template #body="{ data }: { data: EndpointLatencyItem }">
            <Tag :value="data.method" :severity="getMethodSeverity(data.method)" />
          </template>
        </Column>

        <Column field="endpoint" :header="$t('metrics.latency.endpoint')" style="min-width: 14rem">
          <template #body="{ data }: { data: EndpointLatencyItem }">
            <code class="endpoint-code">{{ data.endpoint }}</code>
          </template>
        </Column>

        <Column field="requestCount" :header="$t('metrics.latency.requests')" style="width: 8rem; text-align: right">
          <template #body="{ data }: { data: EndpointLatencyItem }">
            {{ data.requestCount }}
          </template>
        </Column>

        <Column field="avgDurationMs" :header="$t('metrics.latency.avgMs')" style="width: 9rem; text-align: right">
          <template #body="{ data }: { data: EndpointLatencyItem }">
            {{ data.avgDurationMs.toFixed(1) }} ms
          </template>
        </Column>

        <Column field="p95DurationMs" :header="$t('metrics.latency.p95Ms')" style="width: 9rem; text-align: right">
          <template #body="{ data }: { data: EndpointLatencyItem }">
            <span class="p-highlight">{{ data.p95DurationMs.toFixed(1) }} ms</span>
          </template>
        </Column>

        <Column field="p99DurationMs" :header="$t('metrics.latency.p99Ms')" style="width: 9rem; text-align: right">
          <template #body="{ data }: { data: EndpointLatencyItem }">
            <span class="p99-highlight">{{ data.p99DurationMs.toFixed(1) }} ms</span>
          </template>
        </Column>

        <Column field="errorCount" :header="$t('metrics.latency.errors')" style="width: 7rem; text-align: right">
          <template #body="{ data }: { data: EndpointLatencyItem }">
            <span :class="{ 'text-danger fw-bold': data.errorCount > 0 }">{{ data.errorCount }}</span>
          </template>
        </Column>
      </DataTable>
    </div>

    <!-- Registre d'Auditoria d'Ús (Audit Logs) -->
    <div class="section-card">
      <div class="section-header flex-between">
        <div class="section-title-box">
          <i class="pi pi-list section-icon" />
          <div>
            <h2 class="section-title">{{ $t('metrics.audit.title') }}</h2>
            <p class="section-subtitle">{{ $t('metrics.audit.subtitle') }}</p>
          </div>
        </div>

        <Button
          :label="$t('metrics.audit.exportCsv')"
          icon="pi pi-download"
          severity="primary"
          :loading="metricsStore.isExporting"
          @click="handleExportCSV"
          data-testid="btn-export-csv"
        />
      </div>

      <!-- Filtres d'Auditoria -->
      <div class="filters-bar">
        <div class="search-input-wrapper">
          <i class="pi pi-search search-icon" />
          <InputText
            v-model="auditSearchInput"
            :placeholder="$t('metrics.audit.searchPlaceholder')"
            class="search-input"
            @input="handleAuditSearchInput"
            data-testid="input-audit-search"
          />
          <Button
            v-if="auditSearchInput"
            icon="pi pi-times"
            text
            rounded
            severity="secondary"
            class="clear-search-btn"
            @click="auditSearchInput = ''; metricsStore.setAuditSearch('')"
          />
        </div>

        <div class="filter-controls">
          <Select
            v-model="selectedModule"
            :options="moduleFilterOptions"
            optionLabel="label"
            optionValue="value"
            :placeholder="$t('metrics.audit.filterModule')"
            class="filter-select"
            @change="handleModuleChange"
            data-testid="filter-module-select"
          />

          <Button
            icon="pi pi-filter-slash"
            text
            severity="secondary"
            tooltip="Netejar filtres"
            @click="handleResetAuditFilters"
            data-testid="btn-reset-audit-filters"
          />

          <Button
            icon="pi pi-refresh"
            text
            severity="secondary"
            :loading="metricsStore.isLoadingAudit"
            tooltip="Refrescar auditoria"
            @click="metricsStore.fetchAuditLogs()"
            data-testid="btn-refresh-audit"
          />
        </div>
      </div>

      <!-- Taula d'Auditoria -->
      <DataTable
        :value="metricsStore.auditLogs"
        :loading="metricsStore.isLoadingAudit"
        lazy
        paginator
        :rows="metricsStore.auditPageSize"
        :totalRecords="metricsStore.totalAuditLogs"
        :first="(metricsStore.currentAuditPage - 1) * metricsStore.auditPageSize"
        :rowsPerPageOptions="[10, 20, 50, 100]"
        @page="handleAuditPageChange"
        tableStyle="min-width: 60rem"
        stripedRows
        data-testid="audit-data-table"
      >
        <template #empty>
          <div class="empty-state">
            <i class="pi pi-history empty-icon" />
            <p>{{ $t('metrics.audit.empty') }}</p>
          </div>
        </template>

        <Column field="createdAt" :header="$t('metrics.audit.table.date')" style="width: 11rem">
          <template #body="{ data }: { data: AuditLogItem }">
            <span class="date-text">{{ formatDate(data.createdAt) }}</span>
          </template>
        </Column>

        <Column field="userEmail" :header="$t('metrics.audit.table.user')" style="min-width: 12rem">
          <template #body="{ data }: { data: AuditLogItem }">
            <div class="user-audit-info">
              <span class="user-email-text">{{ data.userEmail || data.userId || 'Anònim' }}</span>
              <Tag v-if="data.userRole" :value="data.userRole" severity="secondary" class="role-micro-tag" />
            </div>
          </template>
        </Column>

        <Column field="module" :header="$t('metrics.audit.table.module')" style="width: 8rem">
          <template #body="{ data }: { data: AuditLogItem }">
            <Tag :value="data.module" severity="info" />
          </template>
        </Column>

        <Column field="action" :header="$t('metrics.audit.table.action')" style="min-width: 10rem">
          <template #body="{ data }: { data: AuditLogItem }">
            <span class="action-text">{{ data.action }}</span>
          </template>
        </Column>

        <Column field="endpoint" :header="$t('metrics.audit.table.endpoint')" style="min-width: 14rem">
          <template #body="{ data }: { data: AuditLogItem }">
            <div class="endpoint-cell">
              <Tag :value="data.method" :severity="getMethodSeverity(data.method)" class="method-tag" />
              <code class="endpoint-code">{{ data.endpoint }}</code>
            </div>
          </template>
        </Column>

        <Column field="statusCode" :header="$t('metrics.audit.table.status')" style="width: 6rem; text-align: center">
          <template #body="{ data }: { data: AuditLogItem }">
            <Tag :value="String(data.statusCode)" :severity="getStatusSeverity(data.statusCode)" />
          </template>
        </Column>

        <Column field="durationMs" :header="$t('metrics.audit.table.duration')" style="width: 7rem; text-align: right">
          <template #body="{ data }: { data: AuditLogItem }">
            {{ data.durationMs }} ms
          </template>
        </Column>

        <Column field="ipAddress" :header="$t('metrics.audit.table.ip')" style="width: 9rem">
          <template #body="{ data }: { data: AuditLogItem }">
            <span class="ip-text">{{ data.ipAddress || '-' }}</span>
          </template>
        </Column>
      </DataTable>
    </div>
  </div>
</template>

<style scoped>
.metrics-view-container {
  max-width: 1280px;
  margin: 0 auto;
  padding: 1.5rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.page-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.page-subtitle {
  font-size: 0.9rem;
  color: #64748b;
  margin: 0.25rem 0 0 0;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
}

/* Grid Mètriques de Resum */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
}

.metric-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1.25rem 1rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
}

.card-icon {
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.icon-users { background: #e0e7ff; color: #4338ca; }
.icon-requests { background: #dcfce7; color: #15803d; }
.icon-latency { background: #fef3c7; color: #b45309; }
.icon-errors { background: #fee2e2; color: #b91c1c; }
.icon-matches { background: #f3e8ff; color: #6b21a8; }
.icon-materials { background: #e0f2fe; color: #0369a1; }

.card-content {
  display: flex;
  flex-direction: column;
}

.card-label {
  font-size: 0.8rem;
  color: #64748b;
  font-weight: 500;
}

.card-value {
  font-size: 1.35rem;
  font-weight: 700;
  color: #0f172a;
  margin-top: 0.15rem;
}

.text-danger {
  color: #ef4444;
}

.fw-bold {
  font-weight: 700;
}

/* Secció de Salut del Servidor */
.section-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1.25rem;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
}

.section-header {
  margin-bottom: 1rem;
}

.flex-between {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.section-title-box {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.section-icon {
  font-size: 1.35rem;
  color: #6366f1;
}

.section-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.section-subtitle {
  font-size: 0.825rem;
  color: #64748b;
  margin: 0.15rem 0 0 0;
}

.health-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.health-item {
  background: #f8fafc;
  border: 1px solid #f1f5f9;
  border-radius: 0.5rem;
  padding: 0.85rem 1rem;
  display: flex;
  flex-direction: column;
}

.health-label {
  font-size: 0.8rem;
  color: #64748b;
  font-weight: 500;
}

.health-value {
  font-size: 1.1rem;
  font-weight: 600;
  color: #1e293b;
  margin-top: 0.25rem;
}

/* Latència i Taula d'Auditoria */
.endpoint-code {
  font-family: monospace;
  font-size: 0.85rem;
  background: #f1f5f9;
  padding: 0.2rem 0.4rem;
  border-radius: 0.25rem;
  color: #334155;
}

.p-highlight {
  color: #d97706;
  font-weight: 600;
}

.p99-highlight {
  color: #dc2626;
  font-weight: 700;
}

.filters-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  min-width: 260px;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  color: #94a3b8;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding-left: 2.25rem;
}

.clear-search-btn {
  position: absolute;
  right: 0.25rem;
}

.filter-controls {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.filter-select {
  min-width: 160px;
}

.date-text {
  font-size: 0.825rem;
  color: #475569;
}

.user-audit-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.user-email-text {
  font-size: 0.85rem;
  color: #1e293b;
  font-weight: 500;
}

.role-micro-tag {
  font-size: 0.65rem !important;
  padding: 0.05rem 0.3rem !important;
}

.action-text {
  font-size: 0.85rem;
  color: #334155;
}

.endpoint-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.method-tag {
  font-size: 0.7rem !important;
}

.ip-text {
  font-size: 0.8rem;
  color: #64748b;
  font-family: monospace;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2.5rem 1rem;
  color: #64748b;
}

.empty-icon {
  font-size: 2.25rem;
  color: #cbd5e1;
  margin-bottom: 0.5rem;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
  .filters-bar {
    flex-direction: column;
    align-items: stretch;
  }
  .filter-controls {
    flex-wrap: wrap;
  }
}
</style>
