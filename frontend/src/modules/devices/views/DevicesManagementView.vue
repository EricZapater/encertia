<template>
  <div class="devices-view-container" data-testid="devices-management-view">
    <!-- Capçalera Principal -->
    <header class="page-header">
      <div class="header-text">
        <h1 class="page-title">{{ $t('devices.title') }}</h1>
        <p class="page-subtitle">{{ $t('devices.subtitle') }}</p>
      </div>
      <div class="header-actions">
        <Button
          icon="pi pi-refresh"
          :label="$t('devices.refresh')"
          severity="secondary"
          outlined
          :loading="isLoading"
          @click="loadDevices"
          data-testid="btn-refresh-devices"
        />
      </div>
    </header>

    <!-- Bloc 1: Descàrrega d'Agents i Certificats -->
    <section class="downloads-section">
      <div class="section-header">
        <div class="section-title-wrapper">
          <i class="pi pi-download section-icon"></i>
          <div>
            <h2 class="section-title">{{ $t('devices.downloads.title') }}</h2>
            <p class="section-subtitle">{{ $t('devices.downloads.subtitle') }}</p>
          </div>
        </div>
      </div>

      <div class="cards-grid">
        <div
          v-for="item in downloadCards"
          :key="item.platform"
          class="agent-card"
          :data-testid="`card-agent-${item.platform}`"
        >
          <div class="card-header">
            <div class="icon-title-group">
              <div class="card-icon-badge" :class="`platform-${item.platform}`">
                <i :class="item.icon"></i>
              </div>
              <div>
                <h3 class="card-title">{{ $t(item.titleKey) }}</h3>
                <span v-if="item.badge" class="platform-badge">{{ item.badge }}</span>
              </div>
            </div>
          </div>

          <p class="card-desc">{{ $t(item.descKey) }}</p>

          <div v-if="item.command" class="command-block">
            <span class="command-label">{{ $t(item.instructionKey) }}</span>
            <div class="command-box">
              <code class="command-code">{{ item.command }}</code>
              <Button
                icon="pi pi-copy"
                text
                rounded
                size="small"
                severity="secondary"
                class="copy-btn"
                :title="$t('devices.copiedTitle')"
                @click="copyCommand(item.command)"
              />
            </div>
          </div>

          <div class="card-footer">
            <Button
              :icon="downloadingPlatform === item.platform ? 'pi pi-spin pi-spinner' : 'pi pi-download'"
              :label="$t(item.buttonKey)"
              :loading="downloadingPlatform === item.platform"
              severity="primary"
              class="download-btn"
              :data-testid="`btn-download-${item.platform}`"
              @click="handleDownload(item)"
            />
          </div>
        </div>
      </div>

      <!-- Avís preventiu de descàrrega i Safe Browsing -->
      <div class="download-notice-banner">
        <i class="pi pi-shield notice-icon"></i>
        <div class="notice-text">
          <h4 class="notice-title">{{ $t('devices.notice.title') }}</h4>
          <p class="notice-desc">{{ $t('devices.notice.body') }}</p>
        </div>
      </div>
    </section>

    <!-- Bloc 2: Taula de Dispositius Autoritzats -->
    <section class="table-section">
      <div class="section-header">
        <div class="section-title-wrapper">
          <i class="pi pi-shield section-icon"></i>
          <div>
            <h2 class="section-title">{{ $t('devices.table.title') }}</h2>
            <p class="section-subtitle">{{ $t('devices.table.subtitle') }}</p>
          </div>
        </div>
      </div>

      <div class="table-card">
        <DataTable
          :value="devices"
          :loading="isLoading"
          data-key="id"
          responsive-layout="scroll"
          striped-rows
          class="devices-datatable"
          data-testid="devices-table"
        >
          <template #empty>
            <div class="empty-state">
              <i class="pi pi-desktop empty-icon"></i>
              <p>{{ $t('devices.noDevices') }}</p>
            </div>
          </template>

          <!-- Nom de l'ordinador -->
          <Column field="name" :header="$t('devices.table.name')" style="min-width: 14rem">
            <template #body="{ data }: { data: DeviceSummary }">
              <div class="device-name-cell">
                <i class="pi pi-desktop device-icon"></i>
                <span class="device-name">{{ data.name }}</span>
              </div>
            </template>
          </Column>

          <!-- Professor assignat -->
          <Column field="teacherName" :header="$t('devices.table.teacher')" style="min-width: 16rem">
            <template #body="{ data }: { data: DeviceSummary }">
              <div class="teacher-info-cell">
                <span class="teacher-name">{{ data.teacherName }}</span>
                <span class="teacher-email">{{ data.teacherEmail }}</span>
              </div>
            </template>
          </Column>

          <!-- Estat -->
          <Column field="isActive" :header="$t('devices.table.status')" style="width: 8rem; text-align: center">
            <template #body="{ data }: { data: DeviceSummary }">
              <Tag
                :severity="data.isActive ? 'success' : 'danger'"
                :value="data.isActive ? $t('devices.status.active') : $t('devices.status.revoked')"
                class="status-tag"
              />
            </template>
          </Column>

          <!-- Darrera activitat -->
          <Column field="lastUsedAt" :header="$t('devices.table.lastActivity')" style="min-width: 12rem">
            <template #body="{ data }: { data: DeviceSummary }">
              <span class="activity-text">
                {{ data.lastUsedAt ? formatDateTime(data.lastUsedAt) : $t('devices.never') }}
              </span>
            </template>
          </Column>

          <!-- Data de registre -->
          <Column field="createdAt" :header="$t('devices.table.registeredAt')" style="min-width: 10rem">
            <template #body="{ data }: { data: DeviceSummary }">
              <span class="date-text">{{ formatDate(data.createdAt) }}</span>
            </template>
          </Column>

          <!-- Accions -->
          <Column :header="$t('devices.table.actions')" style="width: 7rem; text-align: center">
            <template #body="{ data }: { data: DeviceSummary }">
              <Button
                v-if="data.isActive"
                icon="pi pi-ban"
                text
                rounded
                severity="danger"
                size="small"
                :title="$t('devices.revokeTooltip')"
                @click="openRevokeDialog(data)"
                data-testid="btn-revoke-device"
              />
              <span v-else class="revoked-label">{{ $t('devices.status.revoked') }}</span>
            </template>
          </Column>
        </DataTable>
      </div>
    </section>

    <!-- Diàleg de Confirmació de Revocació -->
    <Dialog
      v-model:visible="showRevokeModal"
      modal
      :header="$t('devices.revokeModal.title')"
      :style="{ width: '90vw', maxWidth: '480px' }"
      data-testid="revoke-confirm-dialog"
    >
      <div class="revoke-confirm-content">
        <i class="pi pi-exclamation-triangle warning-icon"></i>
        <div>
          <p class="confirm-message">
            {{ $t('devices.revokeModal.confirmText') }}
            <strong>{{ deviceToRevoke?.name }}</strong>
            <span v-if="deviceToRevoke?.teacherName"> ({{ deviceToRevoke?.teacherName }})</span>?
          </p>
          <p class="revoke-hint">
            {{ $t('devices.revokeModal.hint') }}
          </p>
        </div>
      </div>
      <template #footer>
        <Button
          :label="$t('devices.revokeModal.cancel')"
          icon="pi pi-times"
          text
          severity="secondary"
          @click="showRevokeModal = false"
        />
        <Button
          :label="$t('devices.revokeModal.confirm')"
          icon="pi pi-ban"
          severity="danger"
          :loading="isRevoking"
          @click="confirmRevokeDevice"
          data-testid="btn-confirm-revoke"
        />
      </template>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import type { DeviceSummary, PlatformDownloadInfo } from '../types'
import { listDevices, revokeDevice, downloadAgent } from '../api'

const { t } = useI18n()
const toast = useToast()

const devices = ref<DeviceSummary[]>([])
const isLoading = ref(false)
const downloadingPlatform = ref<string | null>(null)

const showRevokeModal = ref(false)
const deviceToRevoke = ref<DeviceSummary | null>(null)
const isRevoking = ref(false)

const downloadCards: PlatformDownloadInfo[] = [
  {
    platform: 'windows',
    filename: 'encertia-agent-windows.zip',
    icon: 'pi pi-microsoft',
    titleKey: 'devices.cards.windows.title',
    descKey: 'devices.cards.windows.desc',
    instructionKey: 'devices.cards.windows.instruction',
    command: '.\\encertia-agent-windows.exe --register',
    buttonKey: 'devices.cards.windows.button',
    badge: 'Windows (ZIP / x86_64)'
  },
  {
    platform: 'darwin',
    filename: 'encertia-agent-darwin',
    icon: 'pi pi-apple',
    titleKey: 'devices.cards.darwin.title',
    descKey: 'devices.cards.darwin.desc',
    instructionKey: 'devices.cards.darwin.instruction',
    command: 'chmod +x encertia-agent-darwin',
    buttonKey: 'devices.cards.darwin.button',
    badge: 'macOS (Universal)'
  },
  {
    platform: 'linux',
    filename: 'encertia-agent-linux',
    icon: 'pi pi-server',
    titleKey: 'devices.cards.linux.title',
    descKey: 'devices.cards.linux.desc',
    instructionKey: 'devices.cards.linux.instruction',
    command: 'chmod +x encertia-agent-linux',
    buttonKey: 'devices.cards.linux.button',
    badge: 'Linux (amd64)'
  },
  {
    platform: 'cert',
    filename: 'encertia-codesign.cer',
    icon: 'pi pi-shield',
    titleKey: 'devices.cards.cert.title',
    descKey: 'devices.cards.cert.desc',
    instructionKey: 'devices.cards.cert.instruction',
    command: 'certutil -addstore -f "ROOT" encertia-codesign.cer',
    buttonKey: 'devices.cards.cert.button',
    badge: 'Certificat X.509'
  }
]

async function loadDevices() {
  isLoading.value = true
  try {
    const data = await listDevices()
    devices.value = data || []
  } catch (err: any) {
    devices.value = []
    toast.add({
      severity: 'error',
      summary: t('devices.downloadErrorTitle'),
      detail: err.response?.data?.error?.message || err.message || 'Error en carregar dispositius',
      life: 4000
    })
  } finally {
    isLoading.value = false
  }
}

async function handleDownload(item: PlatformDownloadInfo) {
  downloadingPlatform.value = item.platform
  try {
    const blob = await downloadAgent(item.platform)
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', item.filename)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    toast.add({
      severity: 'success',
      summary: t('devices.downloadSuccessTitle'),
      detail: item.filename,
      life: 3000
    })
  } catch (err: any) {
    toast.add({
      severity: 'error',
      summary: t('devices.downloadErrorTitle'),
      detail: err.response?.data?.error?.message || t('devices.downloadError'),
      life: 4000
    })
  } finally {
    downloadingPlatform.value = null
  }
}

function openRevokeDialog(device: DeviceSummary) {
  deviceToRevoke.value = device
  showRevokeModal.value = true
}

async function confirmRevokeDevice() {
  if (!deviceToRevoke.value) return
  isRevoking.value = true
  try {
    await revokeDevice(deviceToRevoke.value.id)
    toast.add({
      severity: 'success',
      summary: t('devices.revokeSuccess'),
      life: 3000
    })
    showRevokeModal.value = false
    await loadDevices()
  } catch (err: any) {
    toast.add({
      severity: 'error',
      summary: t('devices.revokeError'),
      detail: err.response?.data?.error?.message || err.message,
      life: 4000
    })
  } finally {
    isRevoking.value = false
    deviceToRevoke.value = null
  }
}

async function copyCommand(cmd?: string) {
  if (!cmd) return
  try {
    await navigator.clipboard.writeText(cmd)
    toast.add({
      severity: 'info',
      summary: t('devices.copiedTitle'),
      detail: cmd,
      life: 2000
    })
  } catch {
    // Silent
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleDateString('ca-ES', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    })
  } catch {
    return dateStr
  }
}

function formatDateTime(dateStr?: string | null): string {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleString('ca-ES', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch {
    return dateStr
  }
}

onMounted(() => {
  loadDevices()
})
</script>

<style scoped>
.devices-view-container {
  max-width: 1280px;
  margin: 0 auto;
  padding: 1.5rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

/* Page Header */
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
  font-size: 0.95rem;
  color: #64748b;
  margin: 0.25rem 0 0 0;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
}

/* Sections */
.downloads-section,
.table-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.section-title-wrapper {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.section-icon {
  font-size: 1.4rem;
  color: #6366f1;
}

.section-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.section-subtitle {
  font-size: 0.85rem;
  color: #64748b;
  margin: 0.15rem 0 0 0;
}

/* Cards Grid */
.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.25rem;
}

.agent-card {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.agent-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.08);
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.icon-title-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.card-icon-badge {
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.35rem;
}

.platform-windows {
  background-color: #eff6ff;
  color: #2563eb;
}

.platform-darwin {
  background-color: #f1f5f9;
  color: #334155;
}

.platform-linux {
  background-color: #fefce8;
  color: #ca8a04;
}

.platform-cert {
  background-color: #ecfdf5;
  color: #059669;
}

.card-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #0f172a;
  margin: 0;
}

.platform-badge {
  font-size: 0.75rem;
  color: #64748b;
  font-weight: 500;
}

.card-desc {
  font-size: 0.875rem;
  color: #475569;
  line-height: 1.45;
  margin: 0;
  flex-grow: 1;
}

.command-block {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 0.65rem 0.75rem;
}

.command-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

.command-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.command-code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.8rem;
  color: #0f172a;
  word-break: break-all;
  user-select: all;
}

.copy-btn {
  flex-shrink: 0;
}

.card-footer {
  margin-top: auto;
}

.download-btn {
  width: 100%;
}

/* Table Card */
.table-card {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  overflow: hidden;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
}

.device-name-cell {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.device-icon {
  font-size: 1.1rem;
  color: #6366f1;
}

.device-name {
  font-weight: 600;
  color: #1e293b;
}

.teacher-info-cell {
  display: flex;
  flex-direction: column;
}

.teacher-name {
  font-weight: 500;
  color: #1e293b;
  font-size: 0.9rem;
}

.teacher-email {
  font-size: 0.8rem;
  color: #64748b;
}

.status-tag {
  font-size: 0.75rem !important;
  font-weight: 600;
}

.activity-text,
.date-text {
  font-size: 0.85rem;
  color: #475569;
}

.revoked-label {
  font-size: 0.75rem;
  color: #94a3b8;
  font-style: italic;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  color: #64748b;
}

.empty-icon {
  font-size: 2.5rem;
  color: #cbd5e1;
  margin-bottom: 0.75rem;
}

/* Revoke Modal */
.revoke-confirm-content {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
  padding: 0.5rem 0;
}

.warning-icon {
  font-size: 2rem;
  color: #ef4444;
  flex-shrink: 0;
  margin-top: 0.25rem;
}

.confirm-message {
  font-size: 0.95rem;
  color: #1e293b;
  margin: 0 0 0.5rem 0;
  line-height: 1.45;
}

.revoke-hint {
  font-size: 0.85rem;
  color: #64748b;
  margin: 0;
  line-height: 1.4;
}

/* Dark Mode */
:global(.dark-mode) .page-title,
:global(.dark-mode) .section-title,
:global(.dark-mode) .card-title,
:global(.dark-mode) .device-name,
:global(.dark-mode) .teacher-name,
:global(.dark-mode) .confirm-message {
  color: #f8fafc;
}

:global(.dark-mode) .page-subtitle,
:global(.dark-mode) .section-subtitle,
:global(.dark-mode) .card-desc,
:global(.dark-mode) .teacher-email,
:global(.dark-mode) .activity-text,
:global(.dark-mode) .date-text,
:global(.dark-mode) .revoke-hint,
:global(.dark-mode) .empty-state {
  color: #94a3b8;
}

:global(.dark-mode) .agent-card,
:global(.dark-mode) .table-card {
  background-color: #1e293b;
  border-color: #334155;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.3);
}

:global(.dark-mode) .command-block {
  background-color: #0f172a;
  border-color: #334155;
}

:global(.dark-mode) .command-code {
  color: #e2e8f0;
}

:global(.dark-mode) .platform-darwin {
  background-color: #334155;
  color: #f1f5f9;
}

:global(.dark-mode) .empty-icon {
  color: #475569;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
  .cards-grid {
    grid-template-columns: 1fr;
  }
}

.download-notice-banner {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-top: 1.5rem;
  padding: 1rem 1.25rem;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-left: 4px solid #6366f1;
  border-radius: 0.75rem;
}

.notice-icon {
  font-size: 1.35rem;
  color: #6366f1;
  margin-top: 0.15rem;
  flex-shrink: 0;
}

.notice-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.notice-title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: #1e293b;
}

.notice-desc {
  margin: 0;
  font-size: 0.875rem;
  color: #64748b;
  line-height: 1.5;
}

:global(.dark-mode) .download-notice-banner {
  background-color: #1e293b;
  border-color: #334155;
  border-left-color: #818cf8;
}

:global(.dark-mode) .notice-title {
  color: #f1f5f9;
}

:global(.dark-mode) .notice-desc {
  color: #94a3b8;
}
</style>
