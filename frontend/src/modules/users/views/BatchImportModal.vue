<script setup lang="ts">
import { ref, watch } from 'vue'
import { useUserStore } from '../store'
import { parseUsersCsv, type CsvParseResult } from '../utils/csvParser'
import type { BatchCreateUsersResponse } from '../types'

import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import ProgressBar from 'primevue/progressbar'
import { useToast } from 'primevue/usetoast'

defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'imported', response: BatchCreateUsersResponse): void
}>()

const userStore = useUserStore()
const toast = useToast()

const fileInputRef = ref<HTMLInputElement | null>(null)
const fileName = ref<string | null>(null)
const parseResult = ref<CsvParseResult | null>(null)
const isParsing = ref(false)
const isSubmitting = ref(false)
const importResult = ref<BatchCreateUsersResponse | null>(null)
const globalError = ref<string | null>(null)

watch(globalError, (err) => {
  if (err) {
    toast.add({ severity: 'error', summary: 'Error d’Importació', detail: err, life: 4000 })
  }
})

// Pas del procés: 'select' -> 'preview' -> 'result'
const step = ref<'select' | 'preview' | 'result'>('select')

function resetState() {
  fileName.value = null
  parseResult.value = null
  importResult.value = null
  globalError.value = null
  step.value = 'select'
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

function closeModal() {
  emit('update:visible', false)
  resetState()
}

function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  fileName.value = file.name
  globalError.value = null
  isParsing.value = true

  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const text = e.target?.result as string
      const result = parseUsersCsv(text, 'student')
      parseResult.value = result

      if (result.errors.length > 0) {
        globalError.value = result.errors.join(' ')
      } else if (result.totalCount === 0) {
        globalError.value = 'El fitxer no conté cap fila d’usuaris per importar.'
      } else {
        step.value = 'preview'
      }
    } catch (err: any) {
      globalError.value = `Error en processar el fitxer: ${err.message}`
    } finally {
      isParsing.value = false
    }
  }

  reader.onerror = () => {
    globalError.value = 'Error en llegir el fitxer del disc.'
    isParsing.value = false
  }

  reader.readAsText(file)
}

function triggerFileInput() {
  fileInputRef.value?.click()
}

async function handleImportSubmit() {
  if (!parseResult.value || parseResult.value.validItems.length === 0) {
    globalError.value = 'No hi ha usuaris vàlids per importar.'
    return
  }

  isSubmitting.value = true
  globalError.value = null

  try {
    const response = await userStore.batchImport({
      users: parseResult.value.validItems
    })
    importResult.value = response
    step.value = 'result'
    emit('imported', response)
  } catch (err: any) {
    globalError.value =
      err.response?.data?.error?.message ||
      err.message ||
      'Error en enviar el lot d’usuaris al servidor.'
  } finally {
    isSubmitting.value = false
  }
}

function downloadExampleCsv() {
  const exampleContent =
    'email,firstName,lastName,password\n' +
    'joan.puig@encertia.cat,Joan,Puig,AlumnePass123!\n' +
    'clara.vidal@encertia.cat,Clara,Vidal,\n' +
    'arnau.costa@encertia.cat,Arnau,Costa,Provisional2026!'

  const blob = new Blob([exampleContent], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.setAttribute('href', url)
  link.setAttribute('download', 'plantilla_alumnes_encertia.csv')
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
</script>

<template>
  <Dialog
    :visible="visible"
    @update:visible="(val) => emit('update:visible', val)"
    modal
    header="Importació Massiva d'Alumnes (CSV)"
    :style="{ width: '92vw', maxWidth: '820px' }"
    :closable="!isSubmitting"
    data-testid="batch-import-dialog"
  >
    <div class="batch-import-container">
      <!-- PAS 1: SELECCIÓ DE FITXER -->
      <div v-if="step === 'select'" class="step-select">
        <div class="upload-dropzone" @click="triggerFileInput">
          <i class="pi pi-file-excel dropzone-icon" />
          <h3>Selecciona o arrossega el fitxer CSV</h3>
          <p class="dropzone-hint">
            Formats admesos: <code>.csv</code> delimitat per comes, punts i comes o tabulacions.
          </p>
          <input
            ref="fileInputRef"
            type="file"
            accept=".csv,text/csv,text/plain"
            style="display: none"
            @change="handleFileSelect"
          />
          <Button
            type="button"
            label="Examinar Fitxer"
            icon="pi pi-folder-open"
            class="mt-3"
            @click.stop="triggerFileInput"
          />
        </div>

        <div class="template-download-box">
          <div>
            <strong>Necessites una plantilla de referència?</strong>
            <p>Pots descarregar un fitxer CSV d'exemple preparat per omplir.</p>
          </div>
          <Button
            label="Descarregar Plantilla CSV"
            icon="pi pi-download"
            severity="secondary"
            outlined
            size="small"
            @click="downloadExampleCsv"
          />
        </div>
      </div>

      <!-- PAS 2: PREVISUALITZACIÓ -->
      <div v-else-if="step === 'preview'" class="step-preview">
        <div class="preview-summary">
          <div class="summary-item">
            <span class="summary-label">Fitxer:</span>
            <strong>{{ fileName }}</strong>
          </div>
          <div class="summary-badges">
            <Tag severity="success" :value="`Vàlids: ${parseResult?.validCount}`" />
            <Tag
              v-if="parseResult && parseResult.invalidCount > 0"
              severity="danger"
              :value="`Errors de validació: ${parseResult.invalidCount}`"
            />
            <Tag severity="info" :value="`Total files: ${parseResult?.totalCount}`" />
          </div>
        </div>

        <Message v-if="parseResult && parseResult.invalidCount > 0" severity="warn" :closable="false">
          Hi ha files que contenen errors de format. Només s'importaran les files vàlides ({{ parseResult.validCount }} de {{ parseResult.totalCount }}).
        </Message>

        <div class="preview-table-wrapper">
          <DataTable
            :value="parseResult?.rows || []"
            paginator
            :rows="5"
            :rowsPerPageOptions="[5, 10, 25]"
            size="small"
            tableStyle="min-width: 50rem"
            data-testid="csv-preview-table"
          >
            <Column field="rowNumber" header="#" style="width: 4rem" />
            <Column header="Estat" style="width: 7rem">
              <template #body="{ data }">
                <Tag
                  :severity="data.isValid ? 'success' : 'danger'"
                  :value="data.isValid ? 'Vàlid' : 'Invàlid'"
                />
              </template>
            </Column>
            <Column field="data.firstName" header="Nom" />
            <Column field="data.lastName" header="Cognoms" />
            <Column field="data.email" header="Correu Electrònic" />
            <Column header="Observacions / Errors">
              <template #body="{ data }">
                <span v-if="data.isValid" class="text-valid">Tot correcte</span>
                <ul v-else class="error-list">
                  <li v-for="(err, idx) in data.validationErrors" :key="idx">{{ err }}</li>
                </ul>
              </template>
            </Column>
          </DataTable>
        </div>
      </div>

      <!-- PAS 3: INFORME DE RESULTATS -->
      <div v-else-if="step === 'result'" class="step-result">
        <div class="result-header">
          <i
            class="pi"
            :class="importResult?.failedCount === 0 ? 'pi-check-circle success-icon' : 'pi-exclamation-triangle warn-icon'"
          />
          <h3 class="result-title">
            {{ importResult?.failedCount === 0 ? 'Importació completada amb èxit!' : 'Importació finalitzada amb observacions' }}
          </h3>
        </div>

        <div class="result-stats-grid">
          <div class="stat-card">
            <span class="stat-num">{{ importResult?.totalRequested }}</span>
            <span class="stat-label">Total Sol·licitats</span>
          </div>
          <div class="stat-card success-stat">
            <span class="stat-num">{{ importResult?.createdCount }}</span>
            <span class="stat-label">Alumnes Creats</span>
          </div>
          <div class="stat-card danger-stat">
            <span class="stat-num">{{ importResult?.failedCount }}</span>
            <span class="stat-label">Errors de Servidor</span>
          </div>
        </div>

        <!-- Llista d'errors retornats pel servidor -->
        <div v-if="importResult && importResult.errors.length > 0" class="server-errors-box">
          <h4>Detall dels errors detectats pel servidor:</h4>
          <DataTable :value="importResult.errors" size="small" paginator :rows="5">
            <Column field="row" header="Fila" style="width: 5rem" />
            <Column field="email" header="Correu Electrònic" />
            <Column field="error" header="Motiu de l'error">
              <template #body="{ data }">
                <Tag severity="danger" :value="data.error" />
              </template>
            </Column>
          </DataTable>
        </div>
      </div>

      <!-- Barra de progrés -->
      <ProgressBar v-if="isSubmitting" mode="indeterminate" style="height: 4px; margin-top: 1rem" />

      <!-- Accions del modal -->
      <div class="dialog-actions">
        <Button
          v-if="step === 'preview'"
          label="Canviar Fitxer"
          icon="pi pi-arrow-left"
          severity="secondary"
          text
          @click="resetState"
          :disabled="isSubmitting"
        />

        <Button
          v-if="step !== 'result'"
          label="Cancel·lar"
          icon="pi pi-times"
          severity="secondary"
          text
          @click="closeModal"
          :disabled="isSubmitting"
        />

        <Button
          v-if="step === 'preview'"
          :label="`Importar ${parseResult?.validCount || 0} Alumnes`"
          icon="pi pi-upload"
          :loading="isSubmitting"
          :disabled="!parseResult || parseResult.validCount === 0"
          @click="handleImportSubmit"
          data-testid="btn-confirm-import"
        />

        <Button
          v-if="step === 'result'"
          label="Tancar"
          icon="pi pi-check"
          severity="primary"
          @click="closeModal"
          data-testid="btn-close-result"
        />
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
.batch-import-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.upload-dropzone {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2.5rem 1.5rem;
  border: 2px dashed #cbd5e1;
  border-radius: 0.75rem;
  background-color: #f8fafc;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: center;
}

.upload-dropzone:hover {
  border-color: #3b82f6;
  background-color: #eff6ff;
}

.dropzone-icon {
  font-size: 2.75rem;
  color: #3b82f6;
  margin-bottom: 0.75rem;
}

.dropzone-hint {
  font-size: 0.875rem;
  color: #64748b;
  margin-top: 0.25rem;
}

.template-download-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem;
  background-color: #f1f5f9;
  border-radius: 0.5rem;
  margin-top: 1rem;
}

.template-download-box p {
  margin: 0.25rem 0 0 0;
  font-size: 0.85rem;
  color: #64748b;
}

.preview-summary {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background-color: #f8fafc;
  border-radius: 0.5rem;
  margin-bottom: 0.5rem;
}

.summary-badges {
  display: flex;
  gap: 0.5rem;
}

.text-valid {
  color: #10b981;
  font-weight: 500;
  font-size: 0.85rem;
}

.error-list {
  margin: 0;
  padding-left: 1.25rem;
  color: #ef4444;
  font-size: 0.8rem;
}

.result-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 1rem 0;
}

.success-icon {
  font-size: 3rem;
  color: #10b981;
}

.warn-icon {
  font-size: 3rem;
  color: #f59e0b;
}

.result-title {
  margin: 0.5rem 0 0 0;
  font-size: 1.25rem;
  color: #1e293b;
}

.result-stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  margin: 1rem 0;
}

.stat-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1rem;
  background-color: #f8fafc;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
}

.stat-num {
  font-size: 1.75rem;
  font-weight: 700;
  color: #1e293b;
}

.stat-label {
  font-size: 0.8rem;
  color: #64748b;
}

.success-stat {
  border-color: #bbf7d0;
  background-color: #f0fdf4;
}

.success-stat .stat-num {
  color: #16a34a;
}

.danger-stat {
  border-color: #fecaca;
  background-color: #fef2f2;
}

.danger-stat .stat-num {
  color: #dc2626;
}

.server-errors-box {
  margin-top: 1rem;
}

.server-errors-box h4 {
  margin: 0 0 0.5rem 0;
  font-size: 0.95rem;
  color: #991b1b;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #f1f5f9;
}

@media (max-width: 600px) {
  .template-download-box {
    flex-direction: column;
    align-items: flex-start;
  }
  .preview-summary {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }
  .result-stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
