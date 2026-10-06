<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCourseStore } from '../store'
import { listQuizzes } from '@/modules/quizzes/api'
import type { Quiz } from '@/modules/quizzes/types'
import type { CreateScriptBlockRequest, ScriptBlock, ScriptBlockType } from '../types'

import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'

const route = useRoute()
const router = useRouter()
const courseStore = useCourseStore()
const toast = useToast()

const courseId = route.params.courseId as string
const unitId = route.params.unitId as string

// Unit Metadata Form
const unitTitle = ref('')
const unitDescription = ref('')

// Quizzes Vincular Modal
const showLinkQuizModal = ref(false)
const availableQuizzes = ref<Quiz[]>([])
const selectedQuizToLink = ref<string | null>(null)
const isLoadingQuizzes = ref(false)

// Script Blocks State
const scriptBlocksLocal = ref<CreateScriptBlockRequest[]>([])
const showAddBlockModal = ref(false)
const editingBlockIndex = ref<number | null>(null)

const newBlock = ref<CreateScriptBlockRequest>({
  blockType: 'material',
  title: '',
  description: '',
  pdfUrl: '',
  startPage: 1,
  endPage: 1,
  quizId: undefined,
  durationMinutes: 5,
  orderIndex: 0
})

const blockTypeOptions = [
  { label: 'Material PDF / Contingut', value: 'material' as ScriptBlockType },
  { label: 'Qüestionari Interactiu (Quiz)', value: 'quiz' as ScriptBlockType },
  { label: 'Pausa / Descans (Break)', value: 'break' as ScriptBlockType }
]

watch(() => courseStore.error, (err) => {
  if (err) {
    toast.add({ severity: 'error', summary: 'Error', detail: err, life: 4000 })
    courseStore.clearError()
  }
})

onMounted(async () => {
  if (courseId && unitId) {
    const unit = await courseStore.fetchCourseUnit(courseId, unitId)
    unitTitle.value = unit.title
    unitDescription.value = unit.description || ''
    initScriptBlocks(unit.scriptBlocks || [])
  }
})

function initScriptBlocks(blocks: ScriptBlock[]) {
  scriptBlocksLocal.value = blocks.map((b, idx) => ({
    blockType: b.blockType,
    orderIndex: b.orderIndex ?? idx,
    title: b.title,
    description: b.description || undefined,
    materialId: b.materialId || undefined,
    pdfUrl: b.pdfUrl || undefined,
    startPage: b.startPage || undefined,
    endPage: b.endPage || undefined,
    quizId: b.quizId || undefined,
    durationMinutes: b.durationMinutes || undefined
  }))
}

// Update Unit Header Metadata
async function handleSaveMetadata() {
  if (!unitTitle.value.trim()) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'El títol no pot estar buit', life: 3000 })
    return
  }

  try {
    await courseStore.updateCourseUnit(courseId, unitId, {
      title: unitTitle.value.trim(),
      description: unitDescription.value.trim() || undefined
    })
    toast.add({ severity: 'success', summary: 'Èxit', detail: 'Metadades de la unitat actualitzades', life: 3000 })
  } catch (_e) {
    // Handled by store
  }
}

// Link Quiz
async function openLinkQuizModal() {
  selectedQuizToLink.value = null
  showLinkQuizModal.value = true
  isLoadingQuizzes.value = true
  try {
    const res = await listQuizzes({ pageSize: 100 })
    const linkedIds = new Set((courseStore.currentUnit?.linkedQuizzes || []).map((q) => q.id))
    availableQuizzes.value = (res.items || []).filter((q) => !linkedIds.has(q.id))
  } catch (_e) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Error en carregar els qüestionaris', life: 3000 })
  } finally {
    isLoadingQuizzes.value = false
  }
}

async function handleLinkQuiz() {
  if (!selectedQuizToLink.value) return
  try {
    await courseStore.linkQuizToUnit(courseId, unitId, selectedQuizToLink.value)
    toast.add({ severity: 'success', summary: 'Èxit', detail: 'Qüestionari vinculat a la unitat', life: 3000 })
    showLinkQuizModal.value = false
  } catch (_e) {
    // Handled by store
  }
}

async function handleUnlinkQuiz(quizId: string) {
  try {
    await courseStore.unlinkQuizFromUnit(courseId, unitId, quizId)
    toast.add({ severity: 'success', summary: 'Desvinculat', detail: 'Qüestionari desvinculat', life: 3000 })
  } catch (_e) {
    // Handled by store
  }
}

// Script Block Actions
function openAddBlockModal(indexToEdit?: number) {
  if (indexToEdit !== undefined) {
    editingBlockIndex.value = indexToEdit
    const b = scriptBlocksLocal.value[indexToEdit]
    newBlock.value = { ...b }
  } else {
    editingBlockIndex.value = null
    newBlock.value = {
      blockType: 'material',
      title: '',
      description: '',
      pdfUrl: '',
      startPage: 1,
      endPage: 1,
      quizId: undefined,
      durationMinutes: 5,
      orderIndex: scriptBlocksLocal.value.length
    }
  }
  showAddBlockModal.value = true
}

function handleSaveBlockLocal() {
  if (!newBlock.value.title.trim()) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'El títol del bloc és obligatori', life: 3000 })
    return
  }

  const blockPayload: CreateScriptBlockRequest = {
    blockType: newBlock.value.blockType,
    orderIndex: newBlock.value.orderIndex,
    title: newBlock.value.title.trim(),
    description: newBlock.value.description?.trim() || undefined,
    pdfUrl: newBlock.value.blockType === 'material' ? newBlock.value.pdfUrl || undefined : undefined,
    startPage: newBlock.value.blockType === 'material' ? newBlock.value.startPage || undefined : undefined,
    endPage: newBlock.value.blockType === 'material' ? newBlock.value.endPage || undefined : undefined,
    quizId: newBlock.value.blockType === 'quiz' ? newBlock.value.quizId || undefined : undefined,
    durationMinutes: newBlock.value.blockType === 'break' ? newBlock.value.durationMinutes || undefined : undefined
  }

  if (editingBlockIndex.value !== null) {
    scriptBlocksLocal.value[editingBlockIndex.value] = blockPayload
  } else {
    blockPayload.orderIndex = scriptBlocksLocal.value.length
    scriptBlocksLocal.value.push(blockPayload)
  }

  showAddBlockModal.value = false
}

function moveBlock(index: number, direction: 'up' | 'down') {
  const targetIndex = direction === 'up' ? index - 1 : index + 1
  if (targetIndex < 0 || targetIndex >= scriptBlocksLocal.value.length) return

  const temp = scriptBlocksLocal.value[index]
  scriptBlocksLocal.value[index] = scriptBlocksLocal.value[targetIndex]
  scriptBlocksLocal.value[targetIndex] = temp

  // Re-index
  scriptBlocksLocal.value.forEach((b, i) => {
    b.orderIndex = i
  })
}

function removeBlock(index: number) {
  scriptBlocksLocal.value.splice(index, 1)
  scriptBlocksLocal.value.forEach((b, i) => {
    b.orderIndex = i
  })
}

async function handleSaveScript() {
  try {
    const preparedBlocks = scriptBlocksLocal.value.map((b, idx) => ({
      ...b,
      orderIndex: idx
    }))
    await courseStore.updateUnitScript(courseId, unitId, preparedBlocks)
    toast.add({ severity: 'success', summary: 'Èxit', detail: 'Guió de classe desat correctament', life: 3000 })
  } catch (_e) {
    // Handled by store
  }
}

function navigateToPlayer() {
  router.push(`/courses/${courseId}/units/${unitId}/play`)
}

function goBackToCourse() {
  router.push(`/courses/${courseId}`)
}

function blockTypeSeverity(type: ScriptBlockType) {
  switch (type) {
    case 'material':
      return 'info'
    case 'quiz':
      return 'success'
    case 'break':
      return 'warn'
  }
}

function blockTypeLabel(type: ScriptBlockType) {
  switch (type) {
    case 'material':
      return 'Material PDF'
    case 'quiz':
      return 'Qüestionari'
    case 'break':
      return 'Descans'
  }
}
</script>

<template>
  <div class="unit-editor-view">
    <!-- Navigation Top Bar -->
    <div class="top-nav">
      <Button
        label="Tornar al curs"
        icon="pi pi-arrow-left"
        severity="secondary"
        text
        @click="goBackToCourse"
      />
      <div class="nav-right-actions">
        <Button
          label="Reproduir Guió"
          icon="pi pi-play"
          severity="success"
          @click="navigateToPlayer"
        />
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="courseStore.isLoading && !courseStore.currentUnit" class="loading-container">
      <i class="pi pi-spin pi-spinner spinner-icon"></i>
      <p>Carregant la unitat didàctica...</p>
    </div>

    <template v-else-if="courseStore.currentUnit">
      <!-- Section 1: Unit Metadata -->
      <section class="editor-card">
        <div class="card-title-row">
          <h2>Informació de la Unitat Didàctica</h2>
          <Button
            label="Desar canvis"
            icon="pi pi-save"
            severity="primary"
            size="small"
            :loading="courseStore.isSaving"
            @click="handleSaveMetadata"
          />
        </div>

        <div class="form-grid">
          <div class="form-field">
            <label for="unit-editor-title">Títol de la unitat *</label>
            <InputText id="unit-editor-title" v-model="unitTitle" />
          </div>

          <div class="form-field">
            <label for="unit-editor-desc">Descripció</label>
            <Textarea id="unit-editor-desc" v-model="unitDescription" rows="2" />
          </div>
        </div>
      </section>

      <!-- Section 2: Linked Quizzes -->
      <section class="editor-card">
        <div class="card-title-row">
          <div>
            <h2>Qüestionaris Vinculats</h2>
            <p class="section-subtitle">
              Relació N:N de qüestionaris de la plataforma associats a aquesta unitat
            </p>
          </div>
          <Button
            label="Vincular Qüestionari"
            icon="pi pi-link"
            severity="secondary"
            size="small"
            @click="openLinkQuizModal"
          />
        </div>

        <div
          v-if="!courseStore.currentUnit.linkedQuizzes || courseStore.currentUnit.linkedQuizzes.length === 0"
          class="empty-linked-quizzes"
        >
          <i class="pi pi-info-circle mr-2"></i>
          <span>No hi ha cap qüestionari vinculat actualment.</span>
        </div>

        <div v-else class="linked-quizzes-grid">
          <div
            v-for="quiz in courseStore.currentUnit.linkedQuizzes"
            :key="quiz.id"
            class="linked-quiz-chip"
          >
            <i class="pi pi-check-square quiz-chip-icon"></i>
            <div class="quiz-chip-info">
              <span class="quiz-chip-title">{{ quiz.title }}</span>
              <span class="quiz-chip-questions">{{ quiz.questionsCount ?? 0 }} preguntes</span>
            </div>
            <Button
              icon="pi pi-times"
              severity="danger"
              text
              rounded
              size="small"
              title="Desvincular"
              @click="handleUnlinkQuiz(quiz.id)"
            />
          </div>
        </div>
      </section>

      <!-- Section 3: Script Sequence Editor -->
      <section class="editor-card">
        <div class="card-title-row">
          <div>
            <h2>Guió de Classe Seqüencial</h2>
            <p class="section-subtitle">
              Ordre de la sessió: materials de lectura, qüestionaris de control i pauses
            </p>
          </div>
          <div class="script-header-btns">
            <Button
              label="Afegir Bloc"
              icon="pi pi-plus"
              severity="primary"
              size="small"
              @click="openAddBlockModal()"
            />
            <Button
              label="Desar Guió"
              icon="pi pi-save"
              severity="success"
              size="small"
              :loading="courseStore.isSaving"
              @click="handleSaveScript"
            />
          </div>
        </div>

        <div v-if="scriptBlocksLocal.length === 0" class="empty-script-state">
          <i class="pi pi-list-check empty-script-icon"></i>
          <p>El guió de classe està buit.</p>
          <Button
            label="Afegir primer bloc"
            icon="pi pi-plus"
            severity="primary"
            class="mt-2"
            @click="openAddBlockModal()"
          />
        </div>

        <div v-else class="script-blocks-list">
          <div
            v-for="(block, idx) in scriptBlocksLocal"
            :key="idx"
            class="script-block-item"
          >
            <div class="block-drag-badge">
              <span>{{ idx + 1 }}</span>
            </div>

            <div class="block-content">
              <div class="block-header">
                <Tag
                  :severity="blockTypeSeverity(block.blockType)"
                  :value="blockTypeLabel(block.blockType)"
                />
                <h3 class="block-title">{{ block.title }}</h3>
              </div>

              <p v-if="block.description" class="block-desc">{{ block.description }}</p>

              <!-- Details depending on block type -->
              <div class="block-details-pills">
                <span v-if="block.blockType === 'material' && block.pdfUrl" class="detail-pill">
                  <i class="pi pi-file-pdf mr-1"></i> PDF: {{ block.pdfUrl }}
                </span>
                <span v-if="block.blockType === 'material' && (block.startPage || block.endPage)" class="detail-pill">
                  <i class="pi pi-book mr-1"></i> Pàgines: {{ block.startPage || 1 }} - {{ block.endPage || 1 }}
                </span>

                <span v-if="block.blockType === 'quiz' && block.quizId" class="detail-pill">
                  <i class="pi pi-check-square mr-1"></i> Quiz ID: {{ block.quizId }}
                </span>

                <span v-if="block.blockType === 'break'" class="detail-pill">
                  <i class="pi pi-clock mr-1"></i> Durada: {{ block.durationMinutes || 5 }} minuts
                </span>
              </div>
            </div>

            <div class="block-actions">
              <div class="reorder-btns">
                <Button
                  icon="pi pi-chevron-up"
                  severity="secondary"
                  text
                  size="small"
                  :disabled="idx === 0"
                  @click="moveBlock(idx, 'up')"
                />
                <Button
                  icon="pi pi-chevron-down"
                  severity="secondary"
                  text
                  size="small"
                  :disabled="idx === scriptBlocksLocal.length - 1"
                  @click="moveBlock(idx, 'down')"
                />
              </div>

              <Button
                icon="pi pi-pencil"
                severity="secondary"
                text
                size="small"
                title="Editar bloc"
                @click="openAddBlockModal(idx)"
              />

              <Button
                icon="pi pi-trash"
                severity="danger"
                text
                size="small"
                title="Eliminar bloc"
                @click="removeBlock(idx)"
              />
            </div>
          </div>
        </div>
      </section>
    </template>

    <!-- Modal Vincular Qüestionari -->
    <Dialog
      v-model:visible="showLinkQuizModal"
      header="Vincular Qüestionari a la Unitat"
      :modal="true"
      :style="{ width: '480px' }"
    >
      <div class="form-container">
        <p class="modal-intro">
          Selecciona un qüestionari disponible a la plataforma per afegir-lo com a recurs de la unitat.
        </p>

        <div v-if="isLoadingQuizzes" class="text-center py-4">
          <i class="pi pi-spin pi-spinner text-2xl text-indigo-600"></i>
        </div>

        <div v-else-if="availableQuizzes.length === 0" class="empty-available">
          <p>No hi ha qüestionaris addicionals disponibles per vincular.</p>
        </div>

        <div v-else class="form-field">
          <label for="quiz-select">Selecciona el qüestionari</label>
          <Select
            id="quiz-select"
            v-model="selectedQuizToLink"
            :options="availableQuizzes"
            optionLabel="title"
            optionValue="id"
            placeholder="Tria un qüestionari..."
            class="w-full"
          />
        </div>
      </div>
      <template #footer>
        <Button label="Cancel·lar" severity="secondary" text @click="showLinkQuizModal = false" />
        <Button
          label="Vincular"
          icon="pi pi-link"
          severity="primary"
          :disabled="!selectedQuizToLink"
          :loading="courseStore.isSaving"
          @click="handleLinkQuiz"
        />
      </template>
    </Dialog>

    <!-- Modal Afegir / Editar Bloc de Guió -->
    <Dialog
      v-model:visible="showAddBlockModal"
      :header="editingBlockIndex !== null ? 'Editar Bloc de Guió' : 'Afegir Bloc al Guió'"
      :modal="true"
      :style="{ width: '520px' }"
    >
      <div class="form-container">
        <div class="form-field">
          <label for="block-type">Tipus de bloc *</label>
          <Select
            id="block-type"
            v-model="newBlock.blockType"
            :options="blockTypeOptions"
            optionLabel="label"
            optionValue="value"
            class="w-full"
          />
        </div>

        <div class="form-field">
          <label for="block-title">Títol del bloc *</label>
          <InputText
            id="block-title"
            v-model="newBlock.title"
            placeholder="Ex: Introducció teòrica PDF / Descans de 10 min..."
          />
        </div>

        <div class="form-field">
          <label for="block-desc">Descripció o Instruccions</label>
          <Textarea
            id="block-desc"
            v-model="newBlock.description"
            rows="2"
            placeholder="Instruccions per als alumnes..."
          />
        </div>

        <!-- Conditional Fields: Material PDF -->
        <template v-if="newBlock.blockType === 'material'">
          <div class="form-field">
            <label for="block-pdf">URL del document PDF</label>
            <InputText
              id="block-pdf"
              v-model="newBlock.pdfUrl"
              placeholder="https://.../document.pdf"
            />
          </div>
          <div class="form-grid-2">
            <div class="form-field">
              <label for="block-start-page">Pàgina inicial</label>
              <InputNumber id="block-start-page" v-model="newBlock.startPage" :min="1" />
            </div>
            <div class="form-field">
              <label for="block-end-page">Pàgina final</label>
              <InputNumber id="block-end-page" v-model="newBlock.endPage" :min="1" />
            </div>
          </div>
        </template>

        <!-- Conditional Fields: Quiz -->
        <template v-if="newBlock.blockType === 'quiz'">
          <div class="form-field">
            <label for="block-quiz-id">Qüestionari Vinculat</label>
            <Select
              id="block-quiz-id"
              v-model="newBlock.quizId"
              :options="courseStore.currentUnit?.linkedQuizzes || []"
              optionLabel="title"
              optionValue="id"
              placeholder="Selecciona un qüestionari vinculat..."
              class="w-full"
            />
          </div>
        </template>

        <!-- Conditional Fields: Break -->
        <template v-if="newBlock.blockType === 'break'">
          <div class="form-field">
            <label for="block-duration">Durada del descans (minuts)</label>
            <InputNumber id="block-duration" v-model="newBlock.durationMinutes" :min="1" :max="120" />
          </div>
        </template>
      </div>

      <template #footer>
        <Button label="Cancel·lar" severity="secondary" text @click="showAddBlockModal = false" />
        <Button
          label="Desar Bloc"
          icon="pi pi-check"
          severity="primary"
          @click="handleSaveBlockLocal"
        />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.unit-editor-view {
  max-width: 1100px;
  margin: 0 auto;
  padding: 1.5rem;
}

.top-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.loading-container {
  text-align: center;
  padding: 4rem 1rem;
  background: #ffffff;
  border-radius: 0.75rem;
  border: 1px dashed #cbd5e1;
}

.spinner-icon {
  font-size: 2.5rem;
  color: #6366f1;
  margin-bottom: 1rem;
}

.editor-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.card-title-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.25rem;

  h2 {
    font-size: 1.3rem;
    font-weight: 700;
    color: #0f172a;
    margin: 0;
  }
}

.section-subtitle {
  color: #64748b;
  font-size: 0.875rem;
  margin: 0.2rem 0 0 0;
}

.form-grid {
  display: flex;
  flex-direction: column;
  gap: 1rem;
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

.empty-linked-quizzes {
  padding: 1.5rem;
  background: #f8fafc;
  border-radius: 0.5rem;
  color: #64748b;
  font-size: 0.9rem;
}

.linked-quizzes-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.linked-quiz-chip {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  padding: 0.4rem 0.75rem;
  border-radius: 0.5rem;
}

.quiz-chip-icon {
  font-size: 1.1rem;
  color: #4f46e5;
}

.quiz-chip-info {
  display: flex;
  flex-direction: column;
}

.quiz-chip-title {
  font-weight: 600;
  font-size: 0.875rem;
  color: #1e293b;
}

.quiz-chip-questions {
  font-size: 0.75rem;
  color: #64748b;
}

.script-header-btns {
  display: flex;
  gap: 0.5rem;
}

.empty-script-state {
  text-align: center;
  padding: 3rem 1rem;
  color: #64748b;
  border: 1px dashed #cbd5e1;
  border-radius: 0.5rem;
}

.empty-script-icon {
  font-size: 2.5rem;
  color: #94a3b8;
  margin-bottom: 0.75rem;
}

.script-blocks-list {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.script-block-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1rem;
  background: #ffffff;
}

.block-drag-badge {
  width: 2rem;
  height: 2rem;
  background: #f1f5f9;
  color: #475569;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.875rem;
  flex-shrink: 0;
}

.block-content {
  flex: 1;
}

.block-header {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 0.25rem;
}

.block-title {
  font-size: 1.05rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.block-desc {
  font-size: 0.875rem;
  color: #64748b;
  margin: 0 0 0.5rem 0;
}

.block-details-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.detail-pill {
  font-size: 0.775rem;
  color: #475569;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  padding: 0.2rem 0.5rem;
  border-radius: 0.25rem;
}

.block-actions {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.reorder-btns {
  display: flex;
  flex-direction: column;
}

.form-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.modal-intro {
  font-size: 0.9rem;
  color: #475569;
  margin: 0;
}
</style>
