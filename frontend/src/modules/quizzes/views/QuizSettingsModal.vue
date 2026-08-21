<script setup lang="ts">
import { ref, watch } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'
import Chips from 'primevue/chips'
import Message from 'primevue/message'
import type { QuizDetail, QuizStatus } from '../types'
import { useQuizStore } from '../store'

const props = defineProps<{
  visible: boolean
  quiz: QuizDetail | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'saved', updatedFields: Partial<QuizDetail>): void
}>()

const quizStore = useQuizStore()

const title = ref('')
const description = ref('')
const status = ref<QuizStatus>('draft')
const tags = ref<string[]>([])
const coverImageUrl = ref<string | null>(null)

const isUploading = ref(false)
const errorMessage = ref<string | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

const statusOptions = [
  { label: 'Esborrany (Draft)', value: 'draft' as QuizStatus },
  { label: 'Publicat (Published)', value: 'published' as QuizStatus },
  { label: 'Arxivat (Archived)', value: 'archived' as QuizStatus }
]

watch(
  () => props.quiz,
  (newQuiz) => {
    if (newQuiz) {
      title.value = newQuiz.title || ''
      description.value = newQuiz.description || ''
      status.value = newQuiz.status || 'draft'
      tags.value = newQuiz.tags ? [...newQuiz.tags] : []
      coverImageUrl.value = newQuiz.coverImageUrl || null
      errorMessage.value = null
    }
  },
  { immediate: true }
)

function handleClose() {
  emit('update:visible', false)
  errorMessage.value = null
}

function triggerFileUpload() {
  fileInput.value?.click()
}

async function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (file.size > 5 * 1024 * 1024) {
    errorMessage.value = "La imatge supera la mida màxima de 5 MB."
    return
  }

  isUploading.value = true
  errorMessage.value = null
  try {
    const url = await quizStore.uploadImage(file)
    coverImageUrl.value = url
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || err.message || "Error en pujar la imatge de portada."
  } finally {
    isUploading.value = false
    if (target) target.value = ''
  }
}

function removeCoverImage() {
  coverImageUrl.value = null
}

function handleSave() {
  if (!title.value.trim()) {
    errorMessage.value = 'El títol del qüestionari és obligatori.'
    return
  }

  emit('saved', {
    title: title.value.trim(),
    description: description.value.trim() || null,
    status: status.value,
    tags: tags.value,
    coverImageUrl: coverImageUrl.value
  })

  handleClose()
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    header="Configuració del Qüestionari"
    :style="{ width: '90vw', maxWidth: '600px' }"
    @update:visible="(val) => emit('update:visible', val)"
    data-testid="quiz-settings-dialog"
  >
    <div class="settings-modal-content">
      <Message v-if="errorMessage" severity="error" class="mb-3" :closable="true" @close="errorMessage = null">
        {{ errorMessage }}
      </Message>

      <div class="form-field">
        <label for="quiz-title" class="field-label">Títol del Joc <span class="required">*</span></label>
        <InputText
          id="quiz-title"
          v-model="title"
          class="w-full"
          placeholder="Ex: Geografia de Catalunya"
          data-testid="input-settings-title"
        />
      </div>

      <div class="form-field">
        <label for="quiz-desc" class="field-label">Descripció</label>
        <Textarea
          id="quiz-desc"
          v-model="description"
          rows="3"
          class="w-full"
          placeholder="Afegeix una breu descripció o instruccions per als jugadors..."
          data-testid="textarea-settings-desc"
        />
      </div>

      <div class="form-row">
        <div class="form-field flex-1">
          <label for="quiz-status" class="field-label">Estat de Publicació</label>
          <Select
            id="quiz-status"
            v-model="status"
            :options="statusOptions"
            optionLabel="label"
            optionValue="value"
            class="w-full"
            data-testid="select-settings-status"
          />
        </div>

        <div class="form-field flex-1">
          <label class="field-label">Etiquetes (Tags)</label>
          <Chips
            v-model="tags"
            placeholder="Escriu i prem Enter..."
            class="w-full"
            data-testid="chips-settings-tags"
          />
        </div>
      </div>

      <!-- Imatge de portada -->
      <div class="form-field">
        <label class="field-label">Imatge de Portada</label>
        
        <div v-if="coverImageUrl" class="cover-preview-wrapper">
          <img :src="coverImageUrl" alt="Portada del qüestionari" class="cover-image-preview" />
          <div class="cover-actions">
            <Button
              label="Canviar Imatge"
              icon="pi pi-image"
              size="small"
              severity="secondary"
              outlined
              :loading="isUploading"
              @click="triggerFileUpload"
            />
            <Button
              label="Eliminar"
              icon="pi pi-trash"
              size="small"
              severity="danger"
              text
              @click="removeCoverImage"
            />
          </div>
        </div>

        <div v-else class="upload-dropzone" @click="triggerFileUpload">
          <i class="pi pi-cloud-upload upload-icon" />
          <p class="upload-text">
            <span v-if="isUploading">Pujant imatge a Cloudflare R2...</span>
            <span v-else>Fes clic per pujar una imatge de portada (PNG, JPG, WEBP, màx 5MB)</span>
          </p>
        </div>

        <input
          ref="fileInput"
          type="file"
          accept="image/png, image/jpeg, image/webp, image/gif"
          class="hidden-file-input"
          @change="handleFileChange"
        />
      </div>
    </div>

    <template #footer>
      <Button
        label="Cancel·lar"
        icon="pi pi-times"
        text
        severity="secondary"
        @click="handleClose"
      />
      <Button
        label="Acceptar canvis"
        icon="pi pi-check"
        severity="primary"
        @click="handleSave"
        data-testid="btn-save-settings"
      />
    </template>
  </Dialog>
</template>

<style scoped>
.settings-modal-content {
  display: flex;
  flex-direction: column;
  gap: 1.15rem;
  padding: 0.5rem 0;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-row {
  display: flex;
  gap: 1rem;
}

.flex-1 {
  flex: 1;
}

.field-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #334155;
}

.required {
  color: #ef4444;
}

.w-full {
  width: 100%;
}

.cover-preview-wrapper {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 0.75rem;
  background-color: #f8fafc;
}

.cover-image-preview {
  width: 100%;
  max-height: 180px;
  object-fit: cover;
  border-radius: 0.375rem;
}

.cover-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.upload-dropzone {
  border: 2px dashed #cbd5e1;
  border-radius: 0.5rem;
  padding: 1.5rem 1rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  cursor: pointer;
  background-color: #f8fafc;
  transition: all 0.2s ease;
}

.upload-dropzone:hover {
  border-color: #3b82f6;
  background-color: #eff6ff;
}

.upload-icon {
  font-size: 1.75rem;
  color: #64748b;
}

.upload-text {
  font-size: 0.825rem;
  color: #64748b;
  margin: 0;
  text-align: center;
}

.hidden-file-input {
  display: none;
}

.mb-3 {
  margin-bottom: 0.75rem;
}

@media (max-width: 640px) {
  .form-row {
    flex-direction: column;
  }
}
</style>
