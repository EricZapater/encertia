<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useCourseStore } from '../store'
import { useAuthStore } from '@/modules/auth/store'
import type { Course, CourseStatus, CreateCourseRequest } from '../types'

import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import Paginator, { type PageState } from 'primevue/paginator'
import { useToast } from 'primevue/usetoast'

import { useI18n } from 'vue-i18n'

const router = useRouter()
const courseStore = useCourseStore()
const authStore = useAuthStore()
const toast = useToast()
useI18n()

const canCreateOrManage = computed(() => authStore.isAdmin || authStore.isTeacher)

// Filtres
const searchInput = ref('')
let searchTimeout: ReturnType<typeof setTimeout> | null = null
const selectedStatus = ref<CourseStatus | undefined>(undefined)

const statusFilterOptions = [
  { label: 'Tots els estats', value: undefined },
  { label: 'Esborrany', value: 'draft' as CourseStatus },
  { label: 'Actiu', value: 'active' as CourseStatus },
  { label: 'Arxivat', value: 'archived' as CourseStatus }
]

// Modal de Creació de Curs
const showCreateModal = ref(false)
const newCourse = ref<CreateCourseRequest>({
  title: '',
  code: '',
  description: '',
  status: 'draft',
  startDate: '',
  endDate: ''
})
const createFormErrors = ref<Record<string, string>>({})

// Modal Confirmació d'Eliminació
const showDeleteConfirmModal = ref(false)
const courseToDelete = ref<Course | null>(null)

watch(() => courseStore.error, (err) => {
  if (err) {
    toast.add({ severity: 'error', summary: 'Error', detail: err, life: 4000 })
    courseStore.clearError()
  }
})

onMounted(() => {
  courseStore.fetchCourses()
})

function handleSearchInput() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    courseStore.setSearch(searchInput.value)
  }, 350)
}

function handleStatusChange() {
  courseStore.setStatusFilter(selectedStatus.value)
}

function handleResetFilters() {
  searchInput.value = ''
  selectedStatus.value = undefined
  courseStore.resetFilters()
}

function handlePageChange(event: PageState) {
  courseStore.setPage(event.page + 1, event.rows)
}

function openCreateModal() {
  newCourse.value = {
    title: '',
    code: '',
    description: '',
    status: 'draft',
    startDate: '',
    endDate: ''
  }
  createFormErrors.value = {}
  showCreateModal.value = true
}

function validateCreateForm(): boolean {
  createFormErrors.value = {}
  if (!newCourse.value.title.trim()) {
    createFormErrors.value.title = 'El títol del curs és obligatori'
  }
  if (!newCourse.value.code.trim()) {
    createFormErrors.value.code = 'El codi del curs és obligatori'
  }
  return Object.keys(createFormErrors.value).length === 0
}

async function handleCreateCourse() {
  if (!validateCreateForm()) return
  try {
    const payload: CreateCourseRequest = {
      title: newCourse.value.title.trim(),
      code: newCourse.value.code.trim(),
      description: newCourse.value.description?.trim() || undefined,
      status: newCourse.value.status || 'draft',
      startDate: newCourse.value.startDate || undefined,
      endDate: newCourse.value.endDate || undefined
    }
    await courseStore.createCourse(payload)
    toast.add({ severity: 'success', summary: 'Èxit', detail: 'Curs creat correctament', life: 3000 })
    showCreateModal.value = false
  } catch (_e) {
    // Handled by watcher/store
  }
}

function confirmDeleteCourse(course: Course) {
  courseToDelete.value = course
  showDeleteConfirmModal.value = true
}

async function handleDeleteCourse() {
  if (!courseToDelete.value) return
  try {
    await courseStore.deleteCourse(courseToDelete.value.id)
    toast.add({ severity: 'success', summary: 'Eliminat', detail: 'Curs eliminat correctament', life: 3000 })
    showDeleteConfirmModal.value = false
  } catch (_e) {
    // Handled by store
  }
}

function navigateToDetail(id: string) {
  router.push(`/courses/${id}`)
}

function statusSeverity(status: CourseStatus) {
  switch (status) {
    case 'active':
      return 'success'
    case 'draft':
      return 'warn'
    case 'archived':
      return 'secondary'
    default:
      return 'info'
  }
}

function statusLabel(status: CourseStatus) {
  switch (status) {
    case 'active':
      return 'Actiu'
    case 'draft':
      return 'Esborrany'
    case 'archived':
      return 'Arxivat'
    default:
      return status
  }
}
</script>

<template>
  <div class="courses-list-view">
    <!-- Header -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">{{ $t('courses.title') }}</h1>
        <p class="page-subtitle">{{ $t('courses.subtitle') }}</p>
      </div>
      <div class="header-actions">
        <Button
          v-if="canCreateOrManage"
          :label="$t('courses.create')"
          icon="pi pi-plus"
          severity="primary"
          class="create-btn"
          @click="openCreateModal"
        />
      </div>
    </header>

    <!-- Filters Section -->
    <section class="filters-card">
      <div class="filter-group">
        <span class="p-input-icon-left search-input-wrapper">
          <i class="pi pi-search" />
          <InputText
            v-model="searchInput"
            placeholder="Cercar per títol o codi..."
            class="search-input"
            @input="handleSearchInput"
          />
        </span>
      </div>

      <div class="filter-group">
        <Select
          v-model="selectedStatus"
          :options="statusFilterOptions"
          optionLabel="label"
          optionValue="value"
          placeholder="Estat del curs"
          class="status-select"
          @change="handleStatusChange"
        />
      </div>

      <Button
        v-if="searchInput || selectedStatus"
        label="Netejar"
        icon="pi pi-filter-slash"
        severity="secondary"
        text
        @click="handleResetFilters"
      />
    </section>

    <!-- Loading State -->
    <div v-if="courseStore.isLoading" class="loading-container">
      <i class="pi pi-spin pi-spinner spinner-icon"></i>
      <p>Carregant cursos...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="courseStore.courses.length === 0" class="empty-container">
      <i class="pi pi-book empty-icon"></i>
      <h3>No s'han trobat cursos</h3>
      <p>No hi ha cap curs disponible amb els filtres actuals.</p>
      <Button
        v-if="canCreateOrManage"
        label="Crear el primer curs"
        icon="pi pi-plus"
        severity="primary"
        class="mt-3"
        @click="openCreateModal"
      />
    </div>

    <!-- Courses Grid -->
    <div v-else class="courses-grid">
      <div
        v-for="course in courseStore.courses"
        :key="course.id"
        class="course-card"
        @click="navigateToDetail(course.id)"
      >
        <div class="card-header">
          <span class="course-code">{{ course.code }}</span>
          <Tag :severity="statusSeverity(course.status)" :value="statusLabel(course.status)" />
        </div>

        <h2 class="course-title">{{ course.title }}</h2>
        <p class="course-description">
          {{ course.description || 'Sense descripció.' }}
        </p>

        <div class="course-meta">
          <div v-if="course.teacherName" class="meta-item">
            <i class="pi pi-user"></i>
            <span>{{ course.teacherName }}</span>
          </div>
          <div class="meta-item">
            <i class="pi pi-list"></i>
            <span>{{ course.unitsCount ?? 0 }} unitats</span>
          </div>
          <div class="meta-item">
            <i class="pi pi-users"></i>
            <span>{{ course.enrolledStudentsCount ?? 0 }} alumnes</span>
          </div>
        </div>

        <div class="card-footer" @click.stop>
          <Button
            label="Entrar"
            icon="pi pi-arrow-right"
            iconPos="right"
            severity="primary"
            size="small"
            @click="navigateToDetail(course.id)"
          />
          <Button
            v-if="canCreateOrManage"
            icon="pi pi-trash"
            severity="danger"
            text
            rounded
            size="small"
            title="Eliminar Curs"
            @click="confirmDeleteCourse(course)"
          />
        </div>
      </div>
    </div>

    <!-- Paginator -->
    <div v-if="courseStore.totalCount > courseStore.pageSize" class="pagination-wrapper">
      <Paginator
        :rows="courseStore.pageSize"
        :totalRecords="courseStore.totalCount"
        :first="(courseStore.currentPage - 1) * courseStore.pageSize"
        @page="handlePageChange"
      />
    </div>

    <!-- Modal Nou Curs -->
    <Dialog
      v-model:visible="showCreateModal"
      header="Crear Nou Curs"
      :modal="true"
      class="create-course-dialog"
      :style="{ width: '500px' }"
    >
      <div class="form-container">
        <div class="form-field">
          <label for="course-title">Títol del curs *</label>
          <InputText
            id="course-title"
            v-model="newCourse.title"
            placeholder="Ex: Història de la Ciència"
            :class="{ 'p-invalid': createFormErrors.title }"
          />
          <small v-if="createFormErrors.title" class="p-error">{{ createFormErrors.title }}</small>
        </div>

        <div class="form-field">
          <label for="course-code">Codi del curs *</label>
          <InputText
            id="course-code"
            v-model="newCourse.code"
            placeholder="Ex: HIS-101"
            :class="{ 'p-invalid': createFormErrors.code }"
          />
          <small v-if="createFormErrors.code" class="p-error">{{ createFormErrors.code }}</small>
        </div>

        <div class="form-field">
          <label for="course-description">Descripció</label>
          <Textarea
            id="course-description"
            v-model="newCourse.description"
            rows="3"
            placeholder="Descripció general dels objectius del curs..."
          />
        </div>

        <div class="form-field">
          <label for="course-status">Estat inicial</label>
          <Select
            id="course-status"
            v-model="newCourse.status"
            :options="[
              { label: 'Esborrany', value: 'draft' },
              { label: 'Actiu', value: 'active' },
              { label: 'Arxivat', value: 'archived' }
            ]"
            optionLabel="label"
            optionValue="value"
          />
        </div>

        <div class="form-grid-2">
          <div class="form-field">
            <label for="course-start">Data d'inici</label>
            <InputText
              id="course-start"
              v-model="newCourse.startDate"
              type="date"
            />
          </div>
          <div class="form-field">
            <label for="course-end">Data de fi</label>
            <InputText
              id="course-end"
              v-model="newCourse.endDate"
              type="date"
            />
          </div>
        </div>
      </div>

      <template #footer>
        <Button label="Cancel·lar" severity="secondary" text @click="showCreateModal = false" />
        <Button
          label="Crear Curs"
          icon="pi pi-check"
          severity="primary"
          :loading="courseStore.isSaving"
          @click="handleCreateCourse"
        />
      </template>
    </Dialog>

    <!-- Modal Confirmació Eliminació -->
    <Dialog
      v-model:visible="showDeleteConfirmModal"
      header="Confirmar Eliminació"
      :modal="true"
      :style="{ width: '420px' }"
    >
      <div class="confirm-content">
        <i class="pi pi-exclamation-triangle warning-icon"></i>
        <p>
          Estàs segur que vols eliminar el curs
          <strong>"{{ courseToDelete?.title }}"</strong> ({{ courseToDelete?.code }})?
        </p>
      </div>
      <template #footer>
        <Button label="Cancel·lar" severity="secondary" text @click="showDeleteConfirmModal = false" />
        <Button
          label="Eliminar"
          icon="pi pi-trash"
          severity="danger"
          :loading="courseStore.isLoading"
          @click="handleDeleteCourse"
        />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.courses-list-view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 1.5rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}

.page-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.25rem 0;
}

.page-subtitle {
  color: #64748b;
  margin: 0;
  font-size: 0.95rem;
}

.filters-card {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1rem;
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1.5rem;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  position: relative;
}

.search-input-wrapper i {
  position: absolute;
  left: 0.75rem;
  color: #94a3b8;
}

.search-input {
  padding-left: 2.25rem !important;
  width: 280px;
}

.status-select {
  width: 180px;
}

.loading-container,
.empty-container {
  text-align: center;
  padding: 4rem 1rem;
  background: #ffffff;
  border-radius: 0.75rem;
  border: 1px dashed #cbd5e1;
}

.spinner-icon,
.empty-icon {
  font-size: 2.5rem;
  color: #6366f1;
  margin-bottom: 1rem;
}

.courses-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.25rem;
}

.course-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.course-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-color: #c7d2fe;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.course-code {
  font-weight: 700;
  font-size: 0.85rem;
  color: #4f46e5;
  background: #e0e7ff;
  padding: 0.2rem 0.5rem;
  border-radius: 0.375rem;
}

.course-title {
  font-size: 1.15rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 0.5rem 0;
  line-height: 1.3;
}

.course-description {
  color: #64748b;
  font-size: 0.875rem;
  margin: 0 0 1rem 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  height: 2.6rem;
}

.course-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.85rem;
  padding-top: 0.75rem;
  border-top: 1px solid #f1f5f9;
  margin-bottom: 1rem;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
  color: #64748b;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 0.5rem;
}

.pagination-wrapper {
  margin-top: 2rem;
  display: flex;
  justify-content: center;
}

.form-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-field label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #334155;
}

.form-grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.confirm-content {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.warning-icon {
  font-size: 2rem;
  color: #f59e0b;
}
</style>
