<script setup lang="ts">
import { ref, watch } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Checkbox from 'primevue/checkbox'
import Message from 'primevue/message'
import type { Quiz } from '../types'
import { useQuizStore } from '../store'

const props = defineProps<{
  visible: boolean
  quiz: Quiz | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'duplicated', copy: any): void
}>()

const quizStore = useQuizStore()

const copyTitle = ref('')
const includeAnswers = ref(false)
const errorMessage = ref<string | null>(null)
const isSubmitting = ref(false)

watch(
  () => props.quiz,
  (newQuiz) => {
    if (newQuiz) {
      copyTitle.value = `[Còpia] ${newQuiz.title}`
      includeAnswers.value = false
      errorMessage.value = null
    }
  },
  { immediate: true }
)

function handleClose() {
  emit('update:visible', false)
  errorMessage.value = null
}

async function handleDuplicate() {
  if (!props.quiz) return

  if (!copyTitle.value.trim()) {
    errorMessage.value = 'El títol del qüestionari duplicat és obligatori.'
    return
  }

  isSubmitting.value = true
  errorMessage.value = null

  try {
    const duplicated = await quizStore.duplicateQuiz(props.quiz.id, {
      title: copyTitle.value.trim(),
      includeAnswers: includeAnswers.value
    })
    emit('duplicated', duplicated)
    handleClose()
  } catch (err: any) {
    errorMessage.value =
      err.response?.data?.error || err.message || 'Error en duplicar el qüestionari.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    header="Duplicar Qüestionari"
    :style="{ width: '90vw', maxWidth: '500px' }"
    @update:visible="(val) => emit('update:visible', val)"
    data-testid="duplicate-quiz-dialog"
  >
    <div class="duplicate-modal-content">
      <Message v-if="errorMessage" severity="error" class="mb-3" :closable="true" @close="errorMessage = null">
        {{ errorMessage }}
      </Message>

      <p class="modal-description">
        Crea una còpia independent d'aquest qüestionari en estat <strong>Esborrany</strong> per poder adaptar-lo o reutilitzar-lo.
      </p>

      <div class="form-field">
        <label for="copy-title" class="field-label">Títol de la nova còpia <span class="required">*</span></label>
        <InputText
          id="copy-title"
          v-model="copyTitle"
          class="w-full"
          placeholder="Ex: Geografia de Catalunya (Grup B)"
          data-testid="input-duplicate-title"
        />
      </div>

      <div class="checkbox-container">
        <div class="checkbox-row">
          <Checkbox
            v-model="includeAnswers"
            :binary="true"
            inputId="include-answers"
            data-testid="checkbox-include-answers"
          />
          <label for="include-answers" class="checkbox-label">
            Copiar també les opcions de resposta
          </label>
        </div>
        <p class="checkbox-hint">
          Per defecte desactivat: només es copien els enunciats de les preguntes i temps. Si s'activa, es clonaran també totes les respostes i solucions.
        </p>
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
        label="Duplicar Qüestionari"
        icon="pi pi-copy"
        severity="primary"
        :loading="isSubmitting"
        @click="handleDuplicate"
        data-testid="btn-confirm-duplicate"
      />
    </template>
  </Dialog>
</template>

<style scoped>
.duplicate-modal-content {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 0.5rem 0;
}

.modal-description {
  font-size: 0.9rem;
  color: #64748b;
  margin: 0;
  line-height: 1.4;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
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

.checkbox-container {
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 0.85rem 1rem;
}

.checkbox-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.checkbox-label {
  font-size: 0.9rem;
  font-weight: 600;
  color: #1e293b;
  cursor: pointer;
}

.checkbox-hint {
  font-size: 0.8rem;
  color: #64748b;
  margin: 0.35rem 0 0 1.75rem;
}

.mb-3 {
  margin-bottom: 0.75rem;
}
</style>
