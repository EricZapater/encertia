<template>
  <div class="video-downloader-container" data-testid="video-downloader-view">
    <!-- Capçalera Principal -->
    <header class="page-header">
      <div class="header-text">
        <h1 class="page-title">{{ $t('downloader.title') }}</h1>
        <p class="page-subtitle">{{ $t('downloader.subtitle') }}</p>
      </div>
    </header>

    <!-- Targeta d'Entrada d'Enllaç -->
    <div class="downloader-main-card">
      <div class="input-section">
        <label for="video-url-input" class="input-label">
          <i class="pi pi-link"></i>
          <span>Enllaç del vídeo</span>
        </label>

        <div class="input-group">
          <div class="input-wrapper">
            <i class="pi pi-search input-prefix-icon"></i>
            <InputText
              id="video-url-input"
              v-model="videoUrl"
              class="video-url-field"
              :placeholder="$t('downloader.urlPlaceholder')"
              :disabled="isInspecting || isDownloading"
              @keydown.enter="handleInspect"
              data-testid="input-video-url"
            />
            <button
              v-if="videoUrl"
              type="button"
              class="clear-input-btn"
              @click="clearUrl"
              :title="$t('downloader.clear')"
              data-testid="btn-clear-url"
            >
              <i class="pi pi-times"></i>
            </button>
          </div>

          <Button
            type="button"
            icon="pi pi-clipboard"
            :label="$t('downloader.paste')"
            severity="secondary"
            outlined
            class="action-btn"
            @click="handlePaste"
            data-testid="btn-paste-url"
          />

          <Button
            type="button"
            icon="pi pi-search"
            :label="isInspecting ? $t('downloader.inspecting') : $t('downloader.inspectBtn')"
            :loading="isInspecting"
            severity="secondary"
            class="action-btn"
            :disabled="!isValidUrl || isDownloading"
            @click="handleInspect"
            data-testid="btn-inspect-video"
          />

          <Button
            type="button"
            icon="pi pi-download"
            :label="isDownloading ? $t('downloader.downloading') : $t('downloader.downloadBtn')"
            :loading="isDownloading"
            severity="primary"
            class="action-btn download-main-btn"
            :disabled="!isValidUrl || isInspecting"
            @click="handleDownload"
            data-testid="btn-direct-download"
          />
        </div>

        <!-- Missatge d'error si n'hi ha -->
        <div v-if="errorMessage" class="error-banner" data-testid="downloader-error-message">
          <i class="pi pi-exclamation-circle error-icon"></i>
          <span>{{ errorMessage }}</span>
        </div>

        <!-- Badges de xarxes socials compatibles -->
        <div class="supported-platforms">
          <span class="platforms-label">Compatible amb:</span>
          <div class="platform-tags">
            <span class="platform-chip platform-youtube">
              <i class="pi pi-youtube"></i> YouTube
            </span>
            <span class="platform-chip platform-instagram">
              <i class="pi pi-instagram"></i> Instagram Reels
            </span>
            <span class="platform-chip platform-linkedin">
              <i class="pi pi-linkedin"></i> LinkedIn
            </span>
            <span class="platform-chip platform-tiktok">
              <i class="pi pi-video"></i> TikTok
            </span>
            <span class="platform-chip platform-twitter">
              <i class="pi pi-twitter"></i> X / Twitter
            </span>
            <span class="platform-chip platform-facebook">
              <i class="pi pi-facebook"></i> Facebook
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Estat de Càrrega / Inspecció -->
    <div v-if="isInspecting" class="loading-state" data-testid="inspecting-loader">
      <ProgressSpinner style="width: 48px; height: 48px" strokeWidth="4" />
      <p class="loading-text">{{ $t('downloader.inspecting') }}</p>
    </div>

    <!-- Estat de Descàrrega en curs -->
    <div v-if="isDownloading" class="downloading-banner" data-testid="downloading-banner">
      <ProgressSpinner style="width: 32px; height: 32px" strokeWidth="4" />
      <div class="downloading-info">
        <span class="downloading-title">{{ $t('downloader.downloading') }}</span>
        <span class="downloading-subtitle">Processant el fitxer MP4 amb màxima qualitat...</span>
      </div>
    </div>

    <!-- Fitxa de Previsualització del Vídeo -->
    <div v-if="videoInfo && !isInspecting" class="video-preview-card" data-testid="card-video-preview">
      <div class="preview-media">
        <div class="thumbnail-wrapper">
          <img
            v-if="videoInfo.thumbnail"
            :src="videoInfo.thumbnail"
            :alt="videoInfo.title"
            class="video-thumbnail"
            @error="handleImageError"
          />
          <div v-else class="thumbnail-placeholder">
            <i class="pi pi-video placeholder-icon"></i>
          </div>
          <span v-if="formattedDuration" class="duration-badge">
            {{ formattedDuration }}
          </span>
        </div>
      </div>

      <div class="preview-details">
        <div class="preview-header">
          <Tag :severity="getPlatformSeverity(videoInfo.extractor)" class="platform-tag" data-testid="video-platform">
            {{ formatExtractor(videoInfo.extractor) }}
          </Tag>
          <span v-if="videoInfo.uploader" class="uploader-name">
            <i class="pi pi-user"></i> {{ videoInfo.uploader }}
          </span>
        </div>

        <h2 class="video-title" data-testid="video-title">
          {{ videoInfo.title || $t('downloader.unnamedVideo') }}
        </h2>

        <p v-if="videoInfo.description" class="video-description">
          {{ truncate(videoInfo.description, 200) }}
        </p>

        <div class="preview-actions">
          <Button
            type="button"
            icon="pi pi-download"
            :label="isDownloading ? $t('downloader.downloading') : $t('downloader.downloadBtn')"
            :loading="isDownloading"
            severity="primary"
            class="btn-preview-download"
            data-testid="btn-download-mp4"
            @click="handleDownload"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import ProgressSpinner from 'primevue/progressspinner'
import type { VideoInfo } from '../types'
import { fetchVideoInfo, downloadVideoFile } from '../api'

const { t } = useI18n()
const toast = useToast()

const videoUrl = ref('')
const isInspecting = ref(false)
const isDownloading = ref(false)
const videoInfo = ref<VideoInfo | null>(null)
const errorMessage = ref('')

const isValidUrl = computed(() => {
  const trimmed = videoUrl.value.trim()
  if (!trimmed) return false
  try {
    const parsed = new URL(trimmed)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
})

const formattedDuration = computed(() => {
  if (!videoInfo.value?.duration || videoInfo.value.duration <= 0) return ''
  const total = Math.floor(videoInfo.value.duration)
  const m = Math.floor(total / 60)
  const s = total % 60
  return `${m}:${s < 10 ? '0' : ''}${s}`
})

async function handlePaste() {
  try {
    if (navigator.clipboard && navigator.clipboard.readText) {
      const text = await navigator.clipboard.readText()
      if (text) {
        videoUrl.value = text.trim()
        errorMessage.value = ''
        if (isValidUrl.value) {
          handleInspect()
        }
      }
    }
  } catch {
    // Si l'accés al portapapers no està permès, ignora
  }
}

function clearUrl() {
  videoUrl.value = ''
  videoInfo.value = null
  errorMessage.value = ''
}

async function handleInspect() {
  if (!isValidUrl.value || isInspecting.value) return

  errorMessage.value = ''
  isInspecting.value = true
  videoInfo.value = null

  try {
    const info = await fetchVideoInfo(videoUrl.value.trim())
    videoInfo.value = info
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || t('downloader.downloadError')
    errorMessage.value = msg
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: msg,
      life: 5000
    })
  } finally {
    isInspecting.value = false
  }
}

async function handleDownload() {
  if (!isValidUrl.value || isDownloading.value) return

  errorMessage.value = ''
  isDownloading.value = true

  try {
    const { blob, filename } = await downloadVideoFile(videoUrl.value.trim())

    // Disparar descàrrega al navegador
    const blobUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = blobUrl
    link.download = filename || 'video.mp4'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(blobUrl)

    toast.add({
      severity: 'success',
      summary: 'Descàrrega completada',
      detail: t('downloader.downloadSuccess'),
      life: 5000
    })
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || t('downloader.downloadError')
    errorMessage.value = msg
    toast.add({
      severity: 'error',
      summary: 'Error de descàrrega',
      detail: msg,
      life: 5000
    })
  } finally {
    isDownloading.value = false
  }
}

function handleImageError(e: Event) {
  const target = e.target as HTMLImageElement
  if (target) {
    target.style.display = 'none'
  }
}

function formatExtractor(extractor?: string): string {
  if (!extractor) return 'Vídeo'
  const lower = extractor.toLowerCase()
  if (lower.includes('youtube')) return 'YouTube'
  if (lower.includes('instagram')) return 'Instagram'
  if (lower.includes('linkedin')) return 'LinkedIn'
  if (lower.includes('tiktok')) return 'TikTok'
  if (lower.includes('twitter')) return 'Twitter / X'
  return extractor
}

function getPlatformSeverity(extractor?: string): 'info' | 'success' | 'warn' | 'danger' | 'secondary' {
  if (!extractor) return 'secondary'
  const lower = extractor.toLowerCase()
  if (lower.includes('youtube')) return 'danger'
  if (lower.includes('instagram')) return 'warn'
  if (lower.includes('linkedin')) return 'info'
  if (lower.includes('tiktok')) return 'secondary'
  return 'info'
}

function truncate(str: string, maxLen: number): string {
  if (!str) return ''
  return str.length > maxLen ? str.slice(0, maxLen) + '...' : str
}
</script>

<style scoped>
.video-downloader-container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 2rem 1.5rem;
}

.page-header {
  margin-bottom: 2rem;
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text-color, #1e293b);
  margin: 0 0 0.5rem 0;
}

.page-subtitle {
  font-size: 1rem;
  color: var(--text-color-secondary, #64748b);
  margin: 0;
}

/* Targeta Principal */
.downloader-main-card {
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--surface-border, #e2e8f0);
  border-radius: 12px;
  padding: 1.75rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  margin-bottom: 2rem;
}

.input-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-color, #1e293b);
  margin-bottom: 0.75rem;
}

.input-group {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  flex-wrap: wrap;
}

.input-wrapper {
  flex: 1;
  min-width: 280px;
  position: relative;
  display: flex;
  align-items: center;
}

.input-prefix-icon {
  position: absolute;
  left: 1rem;
  color: var(--text-color-secondary, #94a3b8);
  font-size: 1rem;
  pointer-events: none;
}

.video-url-field {
  width: 100%;
  padding-left: 2.75rem;
  padding-right: 2.5rem;
  border-radius: 8px;
  height: 44px;
}

.clear-input-btn {
  position: absolute;
  right: 0.75rem;
  background: transparent;
  border: none;
  color: var(--text-color-secondary, #94a3b8);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem;
  border-radius: 50%;
}

.clear-input-btn:hover {
  color: var(--text-color, #1e293b);
  background: var(--surface-hover, #f1f5f9);
}

.action-btn {
  height: 44px;
  border-radius: 8px;
  font-weight: 600;
}

.download-main-btn {
  min-width: 150px;
}

/* Error Banner */
.error-banner {
  margin-top: 1rem;
  padding: 0.75rem 1rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #b91c1c;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
}

.error-icon {
  font-size: 1.1rem;
}

/* Xarxes Suportades */
.supported-platforms {
  margin-top: 1.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--surface-border, #f1f5f9);
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.platforms-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-color-secondary, #64748b);
}

.platform-tags {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.platform-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.25rem 0.65rem;
  font-size: 0.8rem;
  font-weight: 500;
  border-radius: 999px;
  background: var(--surface-ground, #f8fafc);
  color: var(--text-color, #475569);
  border: 1px solid var(--surface-border, #e2e8f0);
}

.platform-youtube i { color: #ef4444; }
.platform-instagram i { color: #ec4899; }
.platform-linkedin i { color: #0284c7; }
.platform-tiktok i { color: #000000; }
.platform-twitter i { color: #1d9bf0; }
.platform-facebook i { color: #2563eb; }

/* Loading State */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  gap: 1rem;
}

.loading-text {
  font-size: 1rem;
  color: var(--text-color-secondary, #64748b);
  margin: 0;
}

/* Downloading Banner */
.downloading-banner {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 12px;
  margin-bottom: 2rem;
}

.downloading-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.downloading-title {
  font-weight: 600;
  color: #1d4ed8;
  font-size: 0.95rem;
}

.downloading-subtitle {
  font-size: 0.85rem;
  color: #3b82f6;
}

/* Video Preview Card */
.video-preview-card {
  display: flex;
  gap: 1.5rem;
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--surface-border, #e2e8f0);
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transition: transform 0.2s, box-shadow 0.2s;
}

@media (max-width: 768px) {
  .video-preview-card {
    flex-direction: column;
  }
}

.preview-media {
  flex-shrink: 0;
}

.thumbnail-wrapper {
  position: relative;
  width: 280px;
  aspect-ratio: 16 / 9;
  border-radius: 8px;
  overflow: hidden;
  background: #0f172a;
  display: flex;
  align-items: center;
  justify-content: center;
}

@media (max-width: 768px) {
  .thumbnail-wrapper {
    width: 100%;
  }
}

.video-thumbnail {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.thumbnail-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.placeholder-icon {
  font-size: 3rem;
  color: #475569;
}

.duration-badge {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.8);
  color: #ffffff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.preview-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.preview-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.uploader-name {
  font-size: 0.85rem;
  color: var(--text-color-secondary, #64748b);
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.video-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-color, #1e293b);
  margin: 0 0 0.5rem 0;
  line-height: 1.4;
}

.video-description {
  font-size: 0.9rem;
  color: var(--text-color-secondary, #64748b);
  margin: 0 0 1.25rem 0;
  line-height: 1.5;
}

.preview-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.btn-preview-download {
  height: 42px;
  border-radius: 8px;
  font-weight: 600;
  padding: 0 1.5rem;
}
</style>
