<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useMaterialStore } from '../store'
import type { Material } from '../types'

import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import Tag from 'primevue/tag'

const props = withDefaults(
  defineProps<{
    visible: boolean
    material?: Material | null
    pdfUrl?: string
    title?: string
    startPage?: number
    endPage?: number
    trackView?: boolean
  }>(),
  {
    visible: false,
    material: null,
    pdfUrl: '',
    title: '',
    startPage: 1,
    endPage: undefined,
    trackView: true
  }
)

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const materialStore = useMaterialStore()

const currentPageNum = ref(1)

const resolvedTitle = computed(() => {
  return props.material?.title || props.title || 'Visor de Material'
})

const resolvedPdfUrl = computed(() => {
  return props.material?.fileUrl || props.pdfUrl || ''
})

const resolvedMaterialType = computed(() => {
  return props.material?.materialType || (props.material?.videoUrl ? 'video' : 'document')
})

const minPage = computed(() => props.startPage || 1)
const maxPage = computed(() => {
  if (props.endPage && props.endPage >= minPage.value) return props.endPage
  if (props.material?.pageCount && props.material.pageCount > 0) return props.material.pageCount
  return 1
})

const currentEmbedUrl = computed(() => {
  if (resolvedMaterialType.value === 'video') {
    const videoUrl = props.material?.videoUrl || ''
    return formatVideoEmbedUrl(videoUrl, props.material?.videoProvider || undefined)
  }

  const url = resolvedPdfUrl.value
  if (!url) return ''
  // Si té pàgina específica, afegim l'ancla #page=X
  return `${url}#page=${currentPageNum.value}`
})

watch(
  () => props.visible,
  (newVal) => {
    if (newVal) {
      currentPageNum.value = props.startPage || 1
      if (props.trackView && props.material?.id) {
        materialStore.recordView(props.material.id)
      }
    }
  }
)

function formatVideoEmbedUrl(url: string, provider?: string): string {
  if (!url) return ''
  if (provider === 'youtube' || url.includes('youtube.com') || url.includes('youtu.be')) {
    const regExp = /^.*(youtu.be\/|v\/|u\/\w\/|embed\/|watch\?v=|\&v=)([^#\&\?]*).*/
    const match = url.match(regExp)
    if (match && match[2].length === 11) {
      return `https://www.youtube.com/embed/${match[2]}`
    }
  } else if (provider === 'vimeo' || url.includes('vimeo.com')) {
    const regExp = /vimeo\.com\/(?:.*#|.*\/)?([0-9]+)/
    const match = url.match(regExp)
    if (match && match[1]) {
      return `https://player.vimeo.com/video/${match[1]}`
    }
  }
  return url
}

function prevPage() {
  if (currentPageNum.value > minPage.value) {
    currentPageNum.value--
  }
}

function nextPage() {
  if (currentPageNum.value < maxPage.value) {
    currentPageNum.value++
  }
}

function handleClose() {
  emit('update:visible', false)
}
</script>

<template>
  <Dialog
    :visible="visible"
    :modal="true"
    :dismissableMask="true"
    :style="{ width: '85vw', maxWidth: '1100px' }"
    @update:visible="emit('update:visible', $event)"
  >
    <template #header>
      <div class="pdf-viewer-header">
        <div class="header-title">
          <Tag :severity="resolvedMaterialType === 'document' ? 'info' : 'success'" class="mr-2">
            {{ resolvedMaterialType === 'document' ? 'PDF' : 'VÍDEO' }}
          </Tag>
          <span class="font-bold text-xl text-slate-900">{{ resolvedTitle }}</span>
        </div>
      </div>
    </template>

    <div class="pdf-viewer-body">
      <!-- Toolbar controls for document paging -->
      <div v-if="resolvedMaterialType === 'document' && maxPage > 1" class="viewer-toolbar">
        <Button
          icon="pi pi-chevron-left"
          label="Anterior"
          severity="secondary"
          size="small"
          :disabled="currentPageNum <= minPage"
          @click="prevPage"
        />
        <span class="page-indicator">
          Pàgina <strong>{{ currentPageNum }}</strong> de <strong>{{ maxPage }}</strong>
        </span>
        <Button
          icon="pi pi-chevron-right"
          label="Següent"
          severity="secondary"
          size="small"
          iconPos="right"
          :disabled="currentPageNum >= maxPage"
          @click="nextPage"
        />
      </div>

      <!-- PDF / Media Frame -->
      <div class="iframe-container">
        <iframe
          v-if="currentEmbedUrl"
          :src="currentEmbedUrl"
          class="pdf-iframe"
          title="Document Material Viewer"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowfullscreen
        ></iframe>

        <div v-else class="no-content-placeholder">
          <i class="pi pi-exclamation-triangle text-4xl text-amber-500 mb-2"></i>
          <p>No hi ha cap URL vàlida per a la previsualització del material.</p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-between items-center w-full">
        <div v-if="resolvedMaterialType === 'document' && resolvedPdfUrl">
          <a
            :href="resolvedPdfUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="download-link"
          >
            <i class="pi pi-external-link mr-1"></i> Obrir en una pestanya nova
          </a>
        </div>
        <div v-else></div>

        <Button label="Tancar" severity="secondary" @click="handleClose" />
      </div>
    </template>
  </Dialog>
</template>

<style scoped>
.pdf-viewer-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.pdf-viewer-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.viewer-toolbar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
  padding: 0.6rem;
  background-color: #f1f5f9;
  border-radius: 0.5rem;
}

.page-indicator {
  font-size: 0.95rem;
  color: #334155;
}

.iframe-container {
  width: 100%;
  height: 70vh;
  min-height: 480px;
  background-color: #1e293b;
  border-radius: 0.5rem;
  overflow: hidden;
}

.pdf-iframe {
  width: 100%;
  height: 100%;
  border: none;
}

.no-content-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #94a3b8;
}

.download-link {
  color: #4f46e5;
  font-size: 0.9rem;
  font-weight: 500;
  text-decoration: none;
}

.download-link:hover {
  text-decoration: underline;
}
</style>
