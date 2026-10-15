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
import DatePicker from 'primevue/datepicker'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { useI18n } from 'vue-i18n'

const metricsStore = useMetricsStore()
const toast = useToast()
const { t } = useI18n()

// Pestanya activa ('latency' o 'requests')
const activeTab = ref<'latency' | 'requests'>('latency')

// Filtre de dates (per defecte: avui)
const startDate = ref<Date | null>(new Date())
const endDate = ref<Date | null>(new Date())

// Modal de detall de petició
const showDetailModal = ref(false)
const selectedLog = ref<AuditLogItem | null>(null)

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

function toLocalDateStr(d: Date | null): string | undefined {
  if (!d) return undefined
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function handleDateFilterChange() {
  const s = toLocalDateStr(startDate.value)
  const e = toLocalDateStr(endDate.value)
  metricsStore.setAuditDateRange(s, e)
}

function setTodayFilter() {
  startDate.value = new Date()
  endDate.value = new Date()
  handleDateFilterChange()
}

function handleResetAuditFilters() {
  auditSearchInput.value = ''
  selectedModule.value = undefined
  startDate.value = new Date()
  endDate.value = new Date()
  metricsStore.resetAuditFilters()
}

function openLogDetail(log: AuditLogItem) {
  selectedLog.value = log
  showDetailModal.value = true
}

function formatJSON(raw?: string | null): string {
  if (!raw) return ''
  try {
    const parsed = JSON.parse(raw)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return raw
  }
}

async function copyText(text?: string | null) {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    toast.add({
      severity: 'info',
      summary: t('common.success', 'Copiat'),
      detail: t('metrics.requests.copied', 'Copiat al porta-retalls'),
      life: 2500
    })
  } catch {
    // Fallback ignorable
  }
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

    <!-- Secció Unificada: Rànquing de Latència & Peticions Fetes amb Pestanyes -->
    <div class="section-card tabs-card">
      <div class="section-header flex-between">
        <div class="section-title-box">
          <i :class="activeTab === 'latency' ? 'pi pi-chart-bar' : 'pi pi-list'" class="section-icon" />
          <div>
            <h2 class="section-title">
              {{ activeTab === 'latency' ? $t('metrics.latency.title') : $t('metrics.requests.title') }}
            </h2>
            <p class="section-subtitle">
              {{ activeTab === 'latency' ? $t('metrics.latency.subtitle') : $t('metrics.requests.subtitle') }}
            </p>
          </div>
        </div>

        <div class="header-right-actions">
          <!-- Selector de Pestanyes (Pills) -->
          <div class="segment-tabs" role="tablist">
            <button
              type="button"
              class="segment-tab-btn"
              :class="{ active: activeTab === 'latency' }"
              @click="activeTab = 'latency'"
              data-testid="tab-latency"
              role="tab"
              :aria-selected="activeTab === 'latency'"
            >
              <i class="pi pi-chart-bar tab-icon" />
              <span>{{ $t('metrics.tabs.latency') }}</span>
            </button>
            <button
              type="button"
              class="segment-tab-btn"
              :class="{ active: activeTab === 'requests' }"
              @click="activeTab = 'requests'"
              data-testid="tab-requests"
              role="tab"
              :aria-selected="activeTab === 'requests'"
            >
              <i class="pi pi-list tab-icon" />
              <span>{{ $t('metrics.tabs.requests') }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Contingut Pestanya 1: Rànquing de Latència -->
      <div v-show="activeTab === 'latency'" class="tab-pane">
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

      <!-- Contingut Pestanya 2: Peticions Fetes (amb filtre d'interval de dates per defecte avui) -->
      <div v-show="activeTab === 'requests'" class="tab-pane">
        <!-- Barra de filtres completa -->
        <div class="filters-bar requests-filters-bar">
          <!-- Interval de dates (per defecte: avui) -->
          <div class="date-interval-group">
            <div class="date-input-box">
              <label class="filter-mini-label">{{ $t('metrics.requests.fromDate') }}</label>
              <DatePicker
                v-model="startDate"
                dateFormat="dd/mm/yy"
                showIcon
                showClear
                class="filter-datepicker"
                data-testid="filter-start-date"
                @update:modelValue="handleDateFilterChange"
              />
            </div>

            <div class="date-input-box">
              <label class="filter-mini-label">{{ $t('metrics.requests.toDate') }}</label>
              <DatePicker
                v-model="endDate"
                dateFormat="dd/mm/yy"
                showIcon
                showClear
                class="filter-datepicker"
                data-testid="filter-end-date"
                @update:modelValue="handleDateFilterChange"
              />
            </div>

            <Button
              :label="$t('metrics.requests.today')"
              icon="pi pi-calendar"
              severity="secondary"
              outlined
              size="small"
              class="today-btn"
              @click="setTodayFilter"
              data-testid="btn-filter-today"
            />
          </div>

          <!-- Cerca per text lliure -->
          <div class="search-input-wrapper">
            <i class="pi pi-search search-icon" />
            <InputText
              v-model="auditSearchInput"
              :placeholder="$t('metrics.requests.searchPlaceholder')"
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

          <!-- Controls de mòdul, neteja i exportació -->
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
              tooltip="Refrescar peticions"
              @click="metricsStore.fetchAuditLogs()"
              data-testid="btn-refresh-audit"
            />

            <Button
              :label="$t('metrics.audit.exportCsv')"
              icon="pi pi-download"
              severity="primary"
              size="small"
              :loading="metricsStore.isExporting"
              @click="handleExportCSV"
              data-testid="btn-export-csv"
            />
          </div>
        </div>

        <!-- Taula de Peticions Fetes -->
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
          tableStyle="min-width: 65rem"
          stripedRows
          data-testid="audit-data-table"
        >
          <template #empty>
            <div class="empty-state">
              <i class="pi pi-history empty-icon" />
              <p>{{ $t('metrics.requests.empty') }}</p>
            </div>
          </template>

          <Column field="createdAt" :header="$t('metrics.audit.table.date')" style="width: 10.5rem">
            <template #body="{ data }: { data: AuditLogItem }">
              <span class="date-text">{{ formatDate(data.createdAt) }}</span>
            </template>
          </Column>

          <Column field="userEmail" :header="$t('metrics.audit.table.user')" style="min-width: 11rem">
            <template #body="{ data }: { data: AuditLogItem }">
              <div class="user-audit-info">
                <span class="user-email-text">{{ data.userEmail || data.userId || 'Anònim' }}</span>
                <Tag v-if="data.userRole" :value="data.userRole" severity="secondary" class="role-micro-tag" />
              </div>
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

          <Column field="statusCode" :header="$t('metrics.audit.table.status')" style="width: 5.5rem; text-align: center">
            <template #body="{ data }: { data: AuditLogItem }">
              <Tag :value="String(data.statusCode)" :severity="getStatusSeverity(data.statusCode)" />
            </template>
          </Column>

          <Column field="durationMs" :header="$t('metrics.audit.table.duration')" style="width: 6.5rem; text-align: right">
            <template #body="{ data }: { data: AuditLogItem }">
              <span :class="{ 'latency-warning': data.durationMs > 500, 'latency-danger': data.durationMs > 1000 }">
                {{ data.durationMs }} ms
              </span>
            </template>
          </Column>

          <Column :header="$t('metrics.requests.payload')" style="width: 7.5rem; text-align: center">
            <template #body="{ data }: { data: AuditLogItem }">
              <Tag
                v-if="data.payload"
                :value="$t('metrics.requests.hasPayload')"
                severity="info"
                class="cursor-pointer"
                @click="openLogDetail(data)"
              />
              <span v-else class="text-muted text-xs">{{ $t('metrics.requests.noPayload') }}</span>
            </template>
          </Column>

          <Column :header="$t('metrics.requests.response')" style="width: 7.5rem; text-align: center">
            <template #body="{ data }: { data: AuditLogItem }">
              <Tag
                v-if="data.responseBody"
                :value="$t('metrics.requests.hasResponse')"
                severity="success"
                class="cursor-pointer"
                @click="openLogDetail(data)"
              />
              <span v-else class="text-muted text-xs">{{ $t('metrics.requests.noResponse') }}</span>
            </template>
          </Column>

          <Column :header="$t('common.actions')" style="width: 6.5rem; text-align: center">
            <template #body="{ data }: { data: AuditLogItem }">
              <Button
                icon="pi pi-search"
                :label="$t('metrics.requests.viewDetail')"
                size="small"
                text
                severity="secondary"
                @click="openLogDetail(data)"
                data-testid="btn-view-request-detail"
              />
            </template>
          </Column>
        </DataTable>
      </div>
    </div>

    <!-- Modal de Detall de la Petició HTTP -->
    <Dialog
      v-model:visible="showDetailModal"
      modal
      :header="$t('metrics.requests.detailModalTitle')"
      :style="{ width: '92vw', maxWidth: '850px' }"
      class="request-detail-dialog"
      data-testid="request-detail-modal"
    >
      <div v-if="selectedLog" class="request-detail-container">
        <!-- Barra resumida de dades clau -->
        <div class="detail-summary-bar">
          <div class="summary-pill">
            <span class="pill-label">{{ $t('metrics.audit.table.date') }}</span>
            <span class="pill-value">{{ formatDate(selectedLog.createdAt) }}</span>
          </div>

          <div class="summary-pill">
            <span class="pill-label">{{ $t('metrics.audit.table.user') }}</span>
            <span class="pill-value">{{ selectedLog.userEmail || selectedLog.userId || 'Anònim' }}</span>
            <Tag v-if="selectedLog.userRole" :value="selectedLog.userRole" severity="secondary" class="role-micro-tag" />
          </div>

          <div class="summary-pill">
            <span class="pill-label">{{ $t('metrics.audit.table.endpoint') }}</span>
            <div class="endpoint-pill-val">
              <Tag :value="selectedLog.method" :severity="getMethodSeverity(selectedLog.method)" class="method-tag" />
              <code class="endpoint-code">{{ selectedLog.endpoint }}</code>
            </div>
          </div>

          <div class="summary-pill">
            <span class="pill-label">{{ $t('metrics.audit.table.status') }}</span>
            <Tag :value="String(selectedLog.statusCode)" :severity="getStatusSeverity(selectedLog.statusCode)" />
          </div>

          <div class="summary-pill">
            <span class="pill-label">{{ $t('metrics.audit.table.duration') }}</span>
            <span class="pill-value duration-highlight">{{ selectedLog.durationMs }} ms</span>
          </div>

          <div v-if="selectedLog.ipAddress" class="summary-pill">
            <span class="pill-label">{{ $t('metrics.audit.table.ip') }}</span>
            <span class="pill-value">{{ selectedLog.ipAddress }}</span>
          </div>
        </div>

        <!-- Secció Payload (Petició) -->
        <div class="detail-block">
          <div class="block-header">
            <div class="block-title-box">
              <i class="pi pi-sign-in" />
              <h4 class="block-title">{{ $t('metrics.requests.payloadTitle') }}</h4>
            </div>
            <Button
              v-if="selectedLog.payload"
              icon="pi pi-copy"
              label="Copiar"
              text
              size="small"
              severity="secondary"
              @click="copyText(selectedLog.payload)"
            />
          </div>
          <div v-if="selectedLog.payload" class="code-box">
            <pre><code>{{ formatJSON(selectedLog.payload) }}</code></pre>
          </div>
          <p v-else class="empty-block-text">
            {{ $t('metrics.requests.emptyPayload') }}
          </p>
        </div>

        <!-- Secció Resposta (Servidor) -->
        <div class="detail-block">
          <div class="block-header">
            <div class="block-title-box">
              <i class="pi pi-sign-out" />
              <h4 class="block-title">{{ $t('metrics.requests.responseTitle') }}</h4>
            </div>
            <Button
              v-if="selectedLog.responseBody"
              icon="pi pi-copy"
              label="Copiar"
              text
              size="small"
              severity="secondary"
              @click="copyText(selectedLog.responseBody)"
            />
          </div>
          <div v-if="selectedLog.responseBody" class="code-box">
            <pre><code>{{ formatJSON(selectedLog.responseBody) }}</code></pre>
          </div>
          <p v-else class="empty-block-text">
            {{ $t('metrics.requests.emptyResponse') }}
          </p>
        </div>
      </div>

      <template #footer>
        <Button
          :label="$t('common.close', 'Tancar')"
          icon="pi pi-times"
          severity="secondary"
          @click="showDetailModal = false"
        />
      </template>
    </Dialog>
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
  .date-interval-group {
    flex-direction: column;
    align-items: stretch;
  }
  .filter-datepicker {
    width: 100% !important;
  }
}

/* Pestanyes (Pills Segmentades) */
.segment-tabs {
  display: inline-flex;
  background: #f1f5f9;
  padding: 0.25rem;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
}

.segment-tab-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  background: transparent;
  border: none;
  padding: 0.45rem 0.9rem;
  border-radius: 0.375rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s ease;
}

.segment-tab-btn:hover {
  color: #1e293b;
}

.segment-tab-btn.active {
  background: #ffffff;
  color: #4f46e5;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.tab-icon {
  font-size: 0.95rem;
}

.tab-pane {
  animation: fadeIn 0.2s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(3px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Filtres d'interval de dates */
.requests-filters-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 1rem;
  margin-bottom: 1.25rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #f1f5f9;
}

.date-interval-group {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.date-input-box {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.filter-mini-label {
  font-size: 0.725rem;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
}

.filter-datepicker {
  width: 140px;
}

.today-btn {
  height: 38px;
  font-weight: 600;
}

/* Indicadors de latència i tags de dades */
.latency-warning {
  color: #d97706;
  font-weight: 600;
}

.latency-danger {
  color: #dc2626;
  font-weight: 700;
}

.cursor-pointer {
  cursor: pointer;
}

.text-muted {
  color: #94a3b8;
}

.text-xs {
  font-size: 0.75rem;
}

/* Modal de detall de la petició */
.request-detail-container {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.detail-summary-bar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.75rem;
  background: #f8fafc;
  padding: 1rem;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
}

.summary-pill {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.pill-label {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  color: #64748b;
}

.pill-value {
  font-size: 0.875rem;
  font-weight: 500;
  color: #0f172a;
}

.duration-highlight {
  font-weight: 700;
  color: #6366f1;
}

.endpoint-pill-val {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  flex-wrap: wrap;
}

.detail-block {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.block-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.block-title-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #1e293b;
}

.block-title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 700;
}

.code-box {
  background: #0f172a;
  color: #f8fafc;
  padding: 1rem;
  border-radius: 0.5rem;
  max-height: 260px;
  overflow: auto;
  border: 1px solid #1e293b;
}

.code-box pre {
  margin: 0;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 0.8rem;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}

.empty-block-text {
  margin: 0.25rem 0;
  font-size: 0.85rem;
  color: #94a3b8;
  font-style: italic;
}

/* Dark mode */
:global(.dark-mode) .segment-tabs {
  background: #1e293b;
  border-color: #334155;
}

:global(.dark-mode) .segment-tab-btn {
  color: #94a3b8;
}

:global(.dark-mode) .segment-tab-btn.active {
  background: #334155;
  color: #818cf8;
}

:global(.dark-mode) .detail-summary-bar {
  background: #1e293b;
  border-color: #334155;
}

:global(.dark-mode) .pill-value {
  color: #f1f5f9;
}

:global(.dark-mode) .block-title-box {
  color: #f1f5f9;
}
</style>
