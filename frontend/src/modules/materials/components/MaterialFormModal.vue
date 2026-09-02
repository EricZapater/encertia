<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useMaterialStore } from '../store'
import type { Material, MaterialType, VideoProvider, CreateMaterialRequest, UpdateMaterialRequest } from '../types'

import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'

const props = defineProps<{
  visible: boolean
  materialToEdit?: Material | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'saved'): void
}>()

const materialStore = useMaterialStore()
const toast = useToast()

const materialType = ref<MaterialType>('document')
const title = ref('')
const description = ref('')

// Campos para documento
const fileUrl = ref('')
const fileName = ref('')
const fileSizeBytes = ref<number | undefined>(undefined)
const mimeType = ref('')
const pageCount = ref<number>(1)
const selectedFile = ref<File | null>(null)
const isUploadingFile = ref(false)

// Campos para vídeo
const videoUrl = ref('')
const videoProvider = ref<VideoProvider>('youtube')

const isEditMode = computed(() => !!props.materialToEdit)

const typeOptions = [
  { label: 'Document (PDF, Word, PPT)', value: 'document' as MaterialType },
  { label: 'Vídeo (YouTube, Vimeo...)', value: 'video' as MaterialType }
]

const providerOptions = [
  { label: 'YouTube', value: 'youtube' as VideoProvider },
  { label: 'Vimeo', value: 'vimeo' as VideoProvider },
  { label: 'Extern / Altres', value: 'external' as VideoProvider }
]

watch(
  () => props.visible,
  (newVal) => {
    if (newVal) {
      if (props.materialToEdit) {
        materialType.value = props.materialToEdit.materialType
        title.value = props.materialToEdit.title
        description.value = props.materialToEdit.description || ''
        fileUrl.value = props.materialToEdit.fileUrl || ''
        fileName.value = props.materialToEdit.fileName || ''
        fileSizeBytes.value = props.materialToEdit.fileSizeBytes || undefined
        mimeType.value = props.materialToEdit.mimeType || ''
        pageCount.value = props.materialToEdit.pageCount || 1
        videoUrl.value = props.materialToEdit.videoUrl || ''
        videoProvider.value = (props.materialToEdit.videoProvider as VideoProvider) || 'youtube'
      } else {
        resetForm()
      }
    }
  }
)

watch(videoUrl, (url) => {
  if (!url) return
  if (url.includes('youtube.com') || url.includes('youtu.be')) {
    videoProvider.value = 'youtube'
  } else if (url.includes('vimeo.com')) {
    videoProvider.value = 'vimeo'
  }
})

function resetForm() {
  materialType.value = 'document'
  title.value = ''
  description.value = ''
  fileUrl.value = ''
  fileName.value = ''
  fileSizeBytes.value = undefined
  mimeType.value = ''
  pageCount.value = 1
  selectedFile.value = null
  videoUrl.value = ''
  videoProvider.value = 'youtube'
}

function handleClose() {
  emit('update:visible', false)
}

async function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  if (!target.files || target.files.length === 0) return
  const file = target.files[0]
  selectedFile.value = file

  // Auto upload file
  isUploadingFile.value = true
  try {
    const uploadRes = await materialStore.uploadFile(file)
    fileUrl.value = uploadRes.fileUrl
    fileName.value = uploadRes.fileName
    fileSizeBytes.value = uploadRes.fileSizeBytes
    mimeType.value = uploadRes.mimeType
    if (uploadRes.pageCount && uploadRes.pageCount > 0) {
      pageCount.value = uploadRes.pageCount
    }
    if (!title.value.trim()) {
      title.value = file.name.replace(/\.[^/.]+$/, '')
    }
    toast.add({ severity: 'success', summary: 'Fitxer pujat', detail: 'Fitxer pujat correctament', life: 3000 })
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Error de pujada', detail: e.message || 'Error en pujar el fitxer', life: 4000 })
  } finally {
    isUploadingFile.value = false
  }
}

async function handleSubmit() {
  if (!title.value.trim()) {
    toast.add({ severity: 'error', summary: 'Error de validació', detail: 'El títol és obligatori', life: 3000 })
    return
  }

  if (materialType.value === 'document' && !fileUrl.value.trim()) {
    toast.add({ severity: 'error', summary: 'Error de validació', detail: 'Cal pujar un fitxer o indicar la URL del document', life: 3000 })
    return
  }

  if (materialType.value === 'video' && !videoUrl.value.trim()) {
    toast.add({ severity: 'error', summary: 'Error de validació', detail: 'La URL del vídeo és obligatòria', life: 3000 })
    return
  }

  try {
    if (isEditMode.value && props.materialToEdit) {
      const updatePayload: UpdateMaterialRequest = {
        title: title.value.trim(),
        description: description.value.trim() || undefined,
        fileUrl: materialType.value === 'document' ? fileUrl.value.trim() || undefined : undefined,
        fileName: materialType.value === 'document' ? fileName.value.trim() || undefined : undefined,
        fileSizeBytes: materialType.value === 'document' ? fileSizeBytes.value : undefined,
        mimeType: materialType.value === 'document' ? mimeType.value.trim() || undefined : undefined,
        pageCount: materialType.value === 'document' ? pageCount.value : undefined,
        videoUrl: materialType.value === 'video' ? videoUrl.value.trim() || undefined : undefined,
        videoProvider: materialType.value === 'video' ? videoProvider.value : undefined
      }
      await materialStore.updateMaterial(props.materialToEdit.id, updatePayload)
      toast.add({ severity: 'success', summary: 'Èxit', detail: 'Material actualitzat correctament', life: 3000 })
    } else {
      const createPayload: CreateMaterialRequest = {
        title: title.value.trim(),
        description: description.value.trim() || undefined,
        materialType: materialType.value,
        fileUrl: materialType.value === 'document' ? fileUrl.value.trim() || undefined : undefined,
        fileName: materialType.value === 'document' ? fileName.value.trim() || undefined : undefined,
        fileSizeBytes: materialType.value === 'document' ? fileSizeBytes.value : undefined,
        mimeType: materialType.value === 'document' ? mimeType.value.trim() || undefined : undefined,
        pageCount: materialType.value === 'document' ? pageCount.value : undefined,
        videoUrl: materialType.value === 'video' ? videoUrl.value.trim() || undefined : undefined,
        videoProvider: materialType.value === 'video' ? videoProvider.value : undefined
      }
      await materialStore.createMaterial(createPayload)
      toast.add({ severity: 'success', summary: 'Èxit', detail: 'Material creat correctament', life: 3000 })
    }

    emit('saved')
    handleClose()
  } catch (_e) {
    // Error handled in store
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    :header="isEditMode ? 'Editar Material Didàctic' : 'Nou Material Didàctic'"
    :modal="true"
    :style="{ width: '560px' }"
    @update:visible="emit('update:visible', $event)"
  >
    <div class="form-container">
      <div v-if="!isEditMode" class="form-field">
        <label for="material-type-select">Tipus de Material *</label>
        <Select
          id="material-type-select"
          v-model="materialType"
          :options="typeOptions"
          optionLabel="label"
          optionValue="value"
          class="w-full"
        />
      </div>

      <div class="form-field">
        <label for="material-title">Títol del Material *</label>
        <InputText
          id="material-title"
          v-model="title"
          placeholder="Ex: Tema 1 - Introducció al curs PDF"
        />
      </div>

      <div class="form-field">
        <label for="material-desc">Descripció</label>
        <Textarea
          id="material-desc"
          v-model="description"
          rows="3"
          placeholder="Resum o instruccions de lectura..."
        />
      </div>

      <!-- Camps per a Document (PDF/Word/PPT) -->
      <template v-if="materialType === 'document'">
        <div class="form-field">
          <label>Pujar fitxer (PDF, DOCX, PPTX)</label>
          <div class="file-upload-box">
            <input
              type="file"
              accept=".pdf,.doc,.docx,.ppt,.pptx"
              class="hidden-file-input"
              id="material-file-input"
              @change="handleFileChange"
            />
            <label for="material-file-input" class="file-upload-button">
              <i class="pi pi-upload mr-2"></i>
              <span>{{ isUploadingFile ? 'Pujant fitxer...' : 'Seleccionar o arrossegar fitxer' }}</span>
            </label>
            <span v-if="fileName" class="selected-filename">
              <i class="pi pi-file-pdf mr-1"></i> {{ fileName }}
            </span>
          </div>
        </div>

        <div class="form-field">
          <label for="material-file-url">URL del fitxer (o directament la font)</label>
          <InputText
            id="material-file-url"
            v-model="fileUrl"
            placeholder="https://.../document.pdf"
          />
        </div>

        <div class="form-grid-2">
          <div class="form-field">
            <label for="material-file-name">Nom del fitxer</label>
            <InputText id="material-file-name" v-model="fileName" placeholder="document.pdf" />
          </div>
          <div class="form-field">
            <label for="material-page-count">Total de pàgines</label>
            <InputNumber id="material-page-count" v-model="pageCount" :min="1" />
          </div>
        </div>
      </template>

      <!-- Camps per a Vídeo (YouTube, Vimeo) -->
      <template v-if="materialType === 'video'">
        <div class="form-field">
          <label for="material-video-url">URL del vídeo *</label>
          <InputText
            id="material-video-url"
            v-model="videoUrl"
            placeholder="https://www.youtube.com/watch?v=... o https://vimeo.com/..."
          />
        </div>

        <div class="form-field">
          <label for="material-video-provider">Plataforma de vídeo</label>
          <Select
            id="material-video-provider"
            v-model="videoProvider"
            :options="providerOptions"
            optionLabel="label"
            optionValue="value"
            class="w-full"
          />
        </div>
      </template>
    </div>

    <template #footer>
      <Button label="Cancel·lar" severity="secondary" text @click="handleClose" />
      <Button
        :label="isEditMode ? 'Guardar Canvis' : 'Crear Material'"
        icon="pi pi-check"
        severity="primary"
        :loading="materialStore.isSaving || isUploadingFile"
        @click="handleSubmit"
      />
    </template>
  </Dialog>
</template>

<style scoped>
.form-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-top: 0.5rem;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;

  label {
    font-size: 0.875rem;
    font-weight: 600;
    color: #334155;
  }
}

.form-grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.file-upload-box {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  border: 2px dashed #cbd5e1;
  border-radius: 0.5rem;
  padding: 1rem;
  text-align: center;
  background: #f8fafc;
}

.hidden-file-input {
  display: none;
}

.file-upload-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.6rem 1rem;
  background-color: #6366f1;
  color: #ffffff;
  border-radius: 0.375rem;
  font-weight: 600;
  font-size: 0.9rem;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.file-upload-button:hover {
  background-color: #4f46e5;
}

.selected-filename {
  font-size: 0.85rem;
  color: #475569;
  font-weight: 500;
}
</style>
