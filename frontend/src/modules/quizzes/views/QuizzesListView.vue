<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useQuizStore } from '../store'
import { useMatchStore } from '@/modules/match/store'
import type { Quiz, QuizStatus } from '../types'

import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import Paginator, { type PageState } from 'primevue/paginator'
import { useToast } from 'primevue/usetoast'

import DuplicateQuizModal from './DuplicateQuizModal.vue'

const router = useRouter()
const quizStore = useQuizStore()
const matchStore = useMatchStore()
const toast = useToast()
const isLaunchingMatch = ref<Record<string, boolean>>({})

// Filtres
const searchInput = ref('')
let searchTimeout: ReturnType<typeof setTimeout> | null = null

const selectedStatus = ref<QuizStatus | undefined>(undefined)
const selectedTag = ref('')

const statusFilterOptions = [
  { label: 'Tots els estats', value: undefined },
  { label: 'Esborranys (Draft)', value: 'draft' as QuizStatus },
  { label: 'Publicats (Published)', value: 'published' as QuizStatus },
  { label: 'Arxivats (Archived)', value: 'archived' as QuizStatus }
]

// Modals
const showDuplicateModal = ref(false)
const showDeleteConfirmModal = ref(false)
const selectedQuiz = ref<Quiz | null>(null)
const quizToDelete = ref<Quiz | null>(null)

const successFeedback = ref<string | null>(null)
const errorFeedback = ref<string | null>(null)

watch(successFeedback, (msg) => {
  if (msg) {
    toast.add({ severity: 'success', summary: 'Èxit', detail: msg, life: 3000 })
  }
})

watch([errorFeedback, () => quizStore.error], ([err1, err2]) => {
  const err = err1 || err2
  if (err) {
    toast.add({ severity: 'error', summary: 'Error', detail: err, life: 4000 })
    errorFeedback.value = null
    quizStore.clearError()
  }
})

onMounted(() => {
  quizStore.fetchQuizzes()
})

function handleSearchInput() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    quizStore.setSearch(searchInput.value)
  }, 350)
}

function handleStatusChange() {
  quizStore.setStatusFilter(selectedStatus.value)
}

function handleTagChange() {
  quizStore.setTagFilter(selectedTag.value.trim() || undefined)
}

function handleResetFilters() {
  searchInput.value = ''
  selectedStatus.value = undefined
  selectedTag.value = ''
  quizStore.resetFilters()
}

function handlePageChange(event: PageState) {
  const newPage = event.page + 1
  quizStore.setPage(newPage, event.rows)
}

function navigateToCreate() {
  router.push('/quizzes/new')
}

function navigateToEdit(quiz: Quiz) {
  router.push(`/quizzes/${quiz.id}/edit`)
}

async function handleLaunchMatch(quiz: Quiz) {
  try {
    isLaunchingMatch.value[quiz.id] = true
    const res = await matchStore.initHostMatch(quiz.id)
    router.push(`/matches/${res.id}/host`)
  } catch (err: any) {
    errorFeedback.value =
      err.response?.data?.error?.message || err.message || 'Error en iniciar la partida.'
  } finally {
    isLaunchingMatch.value[quiz.id] = false
  }
}

function openDuplicateModal(quiz: Quiz) {
  selectedQuiz.value = quiz
  showDuplicateModal.value = true
}

function handleDuplicated(copy: Quiz) {
  successFeedback.value = `S'ha duplicat el qüestionari "${copy.title}" correctament en estat esborrany.`
}

function openDeleteConfirm(quiz: Quiz) {
  quizToDelete.value = quiz
  showDeleteConfirmModal.value = true
}

async function confirmDeleteQuiz() {
  if (!quizToDelete.value) return
  try {
    const msg = await quizStore.deleteQuiz(quizToDelete.value.id)
    successFeedback.value = msg || 'Qüestionari eliminat correctament.'
    showDeleteConfirmModal.value = false
    quizToDelete.value = null
  } catch (err: any) {
    errorFeedback.value =
      err.response?.data?.error || err.message || 'Error en eliminar el qüestionari.'
  }
}

async function handleToggleStatus(quiz: Quiz, newStatus: QuizStatus) {
  try {
    await quizStore.updateQuiz(quiz.id, {
      title: quiz.title,
      description: quiz.description,
      coverImageUrl: quiz.coverImageUrl,
      status: newStatus,
      tags: quiz.tags
    })
    quiz.status = newStatus
    successFeedback.value = `Estat actualitzat a "${newStatus}".`
  } catch (err: any) {
    errorFeedback.value =
      err.response?.data?.error || err.message || "Error en canviar l'estat del qüestionari."
  }
}

function getStatusSeverity(status: QuizStatus) {
  switch (status) {
    case 'published':
      return 'success'
    case 'draft':
      return 'warn'
    case 'archived':
      return 'secondary'
    default:
      return undefined
  }
}

function getStatusLabel(status: QuizStatus) {
  switch (status) {
    case 'published':
      return 'Publicat'
    case 'draft':
      return 'Esborrany'
    case 'archived':
      return 'Arxivat'
    default:
      return status
  }
}

function formatDate(dateStr: string) {
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
</script>

<template>
  <div class="quizzes-view-container">
    <!-- Capçalera de pàgina -->
    <div class="page-header">
      <div>
        <h1 class="page-title">Els meus Jocs i Qüestionaris</h1>
        <p class="page-subtitle">
          Crea, gestiona i comparteix activitats interactives estil Kahoot per a la teva aula
        </p>
      </div>

      <div class="header-actions">
        <Button
          label="Nou Joc"
          icon="pi pi-plus"
          severity="primary"
          @click="navigateToCreate"
          data-testid="btn-create-quiz"
        />
      </div>
    </div>

    <!-- Barra de filtres -->
    <div class="filters-bar">
      <div class="search-input-wrapper">
        <i class="pi pi-search search-icon" />
        <InputText
          v-model="searchInput"
          placeholder="Cercar per títol o descripció..."
          class="search-input"
          @input="handleSearchInput"
          data-testid="input-search-quizzes"
        />
        <Button
          v-if="searchInput"
          icon="pi pi-times"
          text
          rounded
          severity="secondary"
          class="clear-search-btn"
          @click="searchInput = ''; quizStore.setSearch('')"
        />
      </div>

      <div class="filter-controls">
        <Select
          v-model="selectedStatus"
          :options="statusFilterOptions"
          optionLabel="label"
          optionValue="value"
          placeholder="Filtrar per estat"
          class="filter-select"
          @change="handleStatusChange"
          data-testid="filter-status-select"
        />

        <div class="tag-input-wrapper">
          <InputText
            v-model="selectedTag"
            placeholder="Filtrar per tag..."
            class="tag-input"
            @keyup.enter="handleTagChange"
            @blur="handleTagChange"
            data-testid="filter-tag-input"
          />
        </div>

        <Button
          icon="pi pi-filter-slash"
          text
          severity="secondary"
          tooltip="Netejar filtres"
          @click="handleResetFilters"
          data-testid="btn-reset-filters"
        />

        <Button
          icon="pi pi-refresh"
          text
          severity="secondary"
          :loading="quizStore.isLoading"
          tooltip="Refrescar llista"
          @click="quizStore.fetchQuizzes()"
          data-testid="btn-refresh-quizzes"
        />
      </div>
    </div>

    <!-- Estat de Càrrega -->
    <div v-if="quizStore.isLoading && !quizStore.hasQuizzes" class="loading-state">
      <i class="pi pi-spin pi-spinner loading-icon" />
      <p>Carregant qüestionaris...</p>
    </div>

    <!-- Estat buit -->
    <div v-else-if="!quizStore.hasQuizzes" class="empty-state" data-testid="quizzes-empty-state">
      <i class="pi pi-folder-open empty-icon" />
      <h3>No s'ha trobat cap qüestionari</h3>
      <p>Pots crear el teu primer joc interactiu fent clic al botó "Nou Joc".</p>
      <Button
        label="Crear el meu primer joc"
        icon="pi pi-plus"
        severity="primary"
        class="mt-2"
        @click="navigateToCreate"
      />
    </div>

    <!-- Graella de Targetes de Jocs -->
    <div v-else class="quiz-grid" data-testid="quizzes-grid">
      <div
        v-for="quiz in quizStore.quizList"
        :key="quiz.id"
        class="quiz-card"
        :data-testid="`quiz-card-${quiz.id}`"
      >
        <!-- Imatge de portada -->
        <div class="card-cover-container" @click="navigateToEdit(quiz)">
          <img
            v-if="quiz.coverImageUrl"
            :src="quiz.coverImageUrl"
            :alt="quiz.title"
            class="card-cover-img"
          />
          <div v-else class="card-cover-placeholder">
            <i class="pi pi-bolt placeholder-icon" />
            <span class="placeholder-text">Encertia Quiz</span>
          </div>

          <div class="card-badge-top">
            <Tag :value="getStatusLabel(quiz.status)" :severity="getStatusSeverity(quiz.status)" />
          </div>
        </div>

        <!-- Contingut de la targeta -->
        <div class="card-body">
          <h3 class="card-title" @click="navigateToEdit(quiz)" :title="quiz.title">
            {{ quiz.title }}
          </h3>

          <p class="card-description">
            {{ quiz.description || 'Sense descripció.' }}
          </p>

          <!-- Tags -->
          <div v-if="quiz.tags && quiz.tags.length > 0" class="card-tags">
            <span v-for="tag in quiz.tags" :key="tag" class="tag-pill">
              #{{ tag }}
            </span>
          </div>

          <!-- Metadades -->
          <div class="card-meta">
            <span class="meta-item">
              <i class="pi pi-list" /> {{ quiz.questionCount }} {{ quiz.questionCount === 1 ? 'pregunta' : 'preguntes' }}
            </span>
            <span class="meta-item">
              <i class="pi pi-calendar" /> {{ formatDate(quiz.updatedAt) }}
            </span>
          </div>
        </div>

        <!-- Accions de la targeta -->
        <div class="card-actions">
          <div class="card-left-actions">
            <Button
              v-if="quiz.status === 'published'"
              label="Llançar"
              icon="pi pi-play"
              size="small"
              severity="success"
              :loading="Boolean(isLaunchingMatch[quiz.id])"
              @click="handleLaunchMatch(quiz)"
              data-testid="btn-launch-match"
            />
            <Button
              label="Editar"
              icon="pi pi-pencil"
              size="small"
              severity="primary"
              outlined
              @click="navigateToEdit(quiz)"
              data-testid="btn-edit-quiz"
            />
          </div>

          <div class="card-action-icons">
            <Button
              v-if="quiz.status !== 'published'"
              icon="pi pi-check-circle"
              text
              rounded
              severity="success"
              size="small"
              tooltip="Publicar"
              @click="handleToggleStatus(quiz, 'published')"
              data-testid="btn-publish-quiz"
            />
            <Button
              v-else
              icon="pi pi-file-edit"
              text
              rounded
              severity="warn"
              size="small"
              tooltip="Passar a Esborrany"
              @click="handleToggleStatus(quiz, 'draft')"
              data-testid="btn-unpublish-quiz"
            />
            <Button
              icon="pi pi-copy"
              text
              rounded
              severity="secondary"
              size="small"
              tooltip="Duplicar"
              @click="openDuplicateModal(quiz)"
              data-testid="btn-duplicate-quiz"
            />
            <Button
              icon="pi pi-trash"
              text
              rounded
              severity="danger"
              size="small"
              tooltip="Eliminar"
              @click="openDeleteConfirm(quiz)"
              data-testid="btn-delete-quiz"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Paginador -->
    <div v-if="quizStore.totalQuizzes > quizStore.pageSize" class="pagination-wrapper">
      <Paginator
        :rows="quizStore.pageSize"
        :totalRecords="quizStore.totalQuizzes"
        :first="(quizStore.currentPage - 1) * quizStore.pageSize"
        @page="handlePageChange"
      />
    </div>

    <!-- Modal de Duplicació -->
    <DuplicateQuizModal
      v-model:visible="showDuplicateModal"
      :quiz="selectedQuiz"
      @duplicated="handleDuplicated"
    />

    <!-- Diàleg de confirmació d'Eliminació -->
    <Dialog
      v-model:visible="showDeleteConfirmModal"
      modal
      header="Confirmar Eliminació"
      :style="{ width: '90vw', maxWidth: '440px' }"
      data-testid="delete-quiz-dialog"
    >
      <div class="delete-confirm-content">
        <i class="pi pi-exclamation-triangle warning-icon" />
        <div>
          <p>
            Estàs segur que vols eliminar el qüestionari
            <strong>{{ quizToDelete?.title }}</strong>?
          </p>
          <p class="soft-delete-hint">
            Aquesta acció mourà el qüestionari a la paperera (soft-delete).
          </p>
        </div>
      </div>
      <template #footer>
        <Button
          label="Cancel·lar"
          icon="pi pi-times"
          text
          severity="secondary"
          @click="showDeleteConfirmModal = false"
        />
        <Button
          label="Eliminar"
          icon="pi pi-trash"
          severity="danger"
          :loading="quizStore.isLoading"
          @click="confirmDeleteQuiz"
          data-testid="btn-confirm-delete-quiz"
        />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.quizzes-view-container {
  max-width: 1240px;
  margin: 0 auto;
  padding: 1.5rem 1rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
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

.filters-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
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
  flex-wrap: wrap;
}

.filter-select {
  min-width: 180px;
}

.tag-input-wrapper {
  max-width: 160px;
}

.tag-input {
  width: 100%;
}

.quiz-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1.25rem;
  margin-bottom: 2rem;
}

.quiz-card {
  background-color: #ffffff;
  border-radius: 0.75rem;
  border: 1px solid #e2e8f0;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.quiz-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.08);
}

.card-cover-container {
  height: 140px;
  position: relative;
  background-color: #f1f5f9;
  cursor: pointer;
  overflow: hidden;
}

.card-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.card-cover-placeholder {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #3b82f6 0%, #1e40af 100%);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  gap: 0.35rem;
}

.placeholder-icon {
  font-size: 2rem;
}

.placeholder-text {
  font-size: 0.85rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.card-badge-top {
  position: absolute;
  top: 0.6rem;
  right: 0.6rem;
}

.card-body {
  padding: 1rem;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.card-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-title:hover {
  color: #2563eb;
}

.card-description {
  font-size: 0.85rem;
  color: #64748b;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.35;
  min-height: 2.3em;
}

.card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-top: 0.25rem;
}

.tag-pill {
  font-size: 0.725rem;
  background-color: #f1f5f9;
  color: #475569;
  padding: 0.15rem 0.45rem;
  border-radius: 9999px;
  font-weight: 500;
}

.card-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.8rem;
  color: #94a3b8;
  margin-top: auto;
  padding-top: 0.75rem;
  border-top: 1px solid #f1f5f9;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.card-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background-color: #fafafa;
  border-top: 1px solid #f1f5f9;
}

.card-left-actions {
  display: flex;
  gap: 0.4rem;
  align-items: center;
}

.card-action-icons {
  display: flex;
  gap: 0.25rem;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 1.5rem;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 1rem;
  color: #64748b;
  text-align: center;
  gap: 0.5rem;
}

.loading-icon,
.empty-icon {
  font-size: 2.5rem;
  color: #cbd5e1;
  margin-bottom: 0.5rem;
}

.delete-confirm-content {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.5rem 0;
}

.warning-icon {
  font-size: 2rem;
  color: #ef4444;
  margin-top: 0.25rem;
}

.soft-delete-hint {
  font-size: 0.825rem;
  color: #64748b;
  margin-top: 0.5rem;
}

.mb-3 {
  margin-bottom: 1rem;
}

.mt-2 {
  margin-top: 0.5rem;
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
}
</style>
