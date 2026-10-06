<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAuthStore } from '@/modules/auth/store'
import { useGroupStore } from '../store'
import type { Group, GroupStudent } from '../types'

import DataTable, { type DataTablePageEvent } from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { useI18n } from 'vue-i18n'

const authStore = useAuthStore()
const groupStore = useGroupStore()
const toast = useToast()
const { t } = useI18n()

const canManage = computed(() => authStore.isAdmin || authStore.isTeacher)

// Filter state
const searchInput = ref('')
const academicYearInput = ref('')
let searchTimeout: ReturnType<typeof setTimeout> | null = null

// Modals
const showFormModal = ref(false)
const showDeleteConfirmModal = ref(false)
const showStudentsModal = ref(false)

const isEditing = ref(false)
const selectedGroup = ref<Group | null>(null)
const groupToDelete = ref<Group | null>(null)
const activeGroupForStudents = ref<Group | null>(null)

// Form fields
const formName = ref('')
const formAcademicYear = ref('26-27')
const formCourseId = ref('')
const formErrors = ref<{ name?: string }>({})

// Student management modal fields
const newStudentIdsText = ref('')
const studentError = ref<string | null>(null)

const successFeedback = ref<string | null>(null)
const errorFeedback = ref<string | null>(null)

watch(successFeedback, (msg) => {
  if (msg) {
    toast.add({ severity: 'success', summary: t('common.success') || 'Èxit', detail: msg, life: 3000 })
    successFeedback.value = null
  }
})

watch([errorFeedback, () => groupStore.error], ([err1, err2]) => {
  const err = err1 || err2
  if (err) {
    toast.add({ severity: 'error', summary: t('common.error') || 'Error', detail: err, life: 4000 })
    errorFeedback.value = null
    groupStore.clearError()
  }
})

onMounted(() => {
  groupStore.fetchGroups()
})

function handleSearchInput() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    groupStore.setSearch(searchInput.value)
  }, 350)
}

function handleAcademicYearFilter() {
  groupStore.setAcademicYearFilter(academicYearInput.value)
}

function handlePageChange(event: DataTablePageEvent) {
  const newPage = Math.floor(event.first / event.rows) + 1
  groupStore.setPage(newPage, event.rows)
}

function handleResetFilters() {
  searchInput.value = ''
  academicYearInput.value = ''
  groupStore.resetFilters()
}

// Form actions
function openCreateModal() {
  isEditing.value = false
  selectedGroup.value = null
  formName.value = ''
  formAcademicYear.value = '26-27'
  formCourseId.value = ''
  formErrors.value = {}
  showFormModal.value = true
}

function openEditModal(group: Group) {
  isEditing.value = true
  selectedGroup.value = group
  formName.value = group.name
  formAcademicYear.value = group.academicYear || '26-27'
  formCourseId.value = group.courseId || ''
  formErrors.value = {}
  showFormModal.value = true
}

function validateForm(): boolean {
  formErrors.value = {}
  if (!formName.value.trim()) {
    formErrors.value.name = 'El nom del grup és obligatori.'
    return false
  }
  return true
}

async function handleSaveGroup() {
  if (!validateForm()) return

  try {
    const payload = {
      name: formName.value.trim(),
      academicYear: formAcademicYear.value.trim() || '26-27',
      courseId: formCourseId.value.trim() || null
    }

    if (isEditing.value && selectedGroup.value) {
      await groupStore.updateGroup(selectedGroup.value.id, payload)
      successFeedback.value = `Grup "${payload.name}" actualitzat correctament.`
    } else {
      await groupStore.createGroup(payload)
      successFeedback.value = `Grup "${payload.name}" creat correctament.`
    }
    showFormModal.value = false
  } catch (err: any) {
    errorFeedback.value = err.response?.data?.message || err.message || 'Error en desar el grup.'
  }
}

// Delete actions
function openDeleteConfirm(group: Group) {
  groupToDelete.value = group
  showDeleteConfirmModal.value = true
}

async function confirmDeleteGroup() {
  if (!groupToDelete.value) return
  try {
    await groupStore.deleteGroup(groupToDelete.value.id)
    successFeedback.value = `Grup "${groupToDelete.value.name}" donat de baixa correctament.`
    showDeleteConfirmModal.value = false
    groupToDelete.value = null
  } catch (err: any) {
    errorFeedback.value = err.response?.data?.message || err.message || 'Error en eliminar el grup.'
  }
}

// Students management
async function openStudentsModal(group: Group) {
  activeGroupForStudents.value = group
  newStudentIdsText.value = ''
  studentError.value = null
  showStudentsModal.value = true
  try {
    await groupStore.fetchGroupStudents(group.id)
  } catch (err: any) {
    studentError.value = 'Error en carregar els alumnes del grup.'
  }
}

async function handleAddStudents() {
  if (!activeGroupForStudents.value) return
  const ids = newStudentIdsText.value
    .split(/[\n,]+/)
    .map((s) => s.trim())
    .filter(Boolean)

  if (ids.length === 0) {
    studentError.value = 'Introdueix almenys un UUID d\'alumne.'
    return
  }

  try {
    studentError.value = null
    await groupStore.addStudentsToGroup(activeGroupForStudents.value.id, ids)
    newStudentIdsText.value = ''
    successFeedback.value = 'Alumnes afegits al grup satisfactòriament.'
  } catch (err: any) {
    studentError.value = err.response?.data?.message || err.message || 'Error en afegir alumnes.'
  }
}

async function handleRemoveStudent(student: GroupStudent) {
  if (!activeGroupForStudents.value) return
  try {
    await groupStore.removeStudentFromGroup(activeGroupForStudents.value.id, student.id)
    successFeedback.value = `Alumne "${student.fullName || student.email}" desassignat del grup.`
  } catch (err: any) {
    studentError.value = err.response?.data?.message || err.message || 'Error en desassignar l\'alumne.'
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
  <div class="groups-view-container">
    <!-- Capçalera de pàgina -->
    <div class="page-header">
      <div>
        <h1 class="page-title">Gestió de Grups</h1>
        <p class="page-subtitle">
          Organitza els alumnes en grups per curs acadèmic i assignatures
        </p>
      </div>

      <div v-if="canManage" class="header-actions">
        <Button
          label="Nou Grup"
          icon="pi pi-plus"
          severity="primary"
          @click="openCreateModal"
          data-testid="btn-open-create-group"
        />
      </div>
    </div>

    <!-- Barra de filtres -->
    <div class="filters-bar">
      <div class="search-input-wrapper">
        <i class="pi pi-search search-icon" />
        <InputText
          v-model="searchInput"
          placeholder="Cercar per nom de grup..."
          class="search-input"
          @input="handleSearchInput"
          data-testid="input-search-groups"
        />
        <Button
          v-if="searchInput"
          icon="pi pi-times"
          text
          rounded
          severity="secondary"
          class="clear-search-btn"
          @click="searchInput = ''; groupStore.setSearch('')"
        />
      </div>

      <div class="filter-controls">
        <InputText
          v-model="academicYearInput"
          placeholder="Curs (ex. 26-27)"
          class="year-filter-input"
          @change="handleAcademicYearFilter"
          data-testid="filter-academic-year-input"
        />

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
          :loading="groupStore.isLoading"
          tooltip="Refrescar llista"
          @click="groupStore.fetchGroups()"
          data-testid="btn-refresh-groups"
        />
      </div>
    </div>

    <!-- Taula de Grups -->
    <div class="table-card">
      <DataTable
        :value="groupStore.groupList"
        :loading="groupStore.isLoading"
        lazy
        paginator
        :rows="groupStore.pageSize"
        :totalRecords="groupStore.totalCount"
        :first="(groupStore.currentPage - 1) * groupStore.pageSize"
        :rowsPerPageOptions="[10, 20, 50]"
        @page="handlePageChange"
        tableStyle="min-width: 50rem"
        stripedRows
        data-testid="groups-data-table"
      >
        <template #empty>
          <div class="empty-state">
            <i class="pi pi-folder-open empty-icon" />
            <p>No s'ha trobat cap grup amb els criteris seleccionats.</p>
          </div>
        </template>

        <Column header="Nom del Grup" style="min-width: 14rem">
          <template #body="{ data }">
            <div class="group-cell">
              <div class="group-icon-sm">
                <i class="pi pi-users"></i>
              </div>
              <div>
                <span class="group-name">{{ data.name }}</span>
                <div v-if="data.courseTitle" class="group-course-sub">
                  {{ data.courseTitle }}
                </div>
              </div>
            </div>
          </template>
        </Column>

        <Column field="academicYear" header="Curs Acadèmic" style="width: 10rem">
          <template #body="{ data }">
            <span class="badge-year">{{ data.academicYear || '26-27' }}</span>
          </template>
        </Column>

        <Column field="studentCount" header="Alumnes" style="width: 8rem">
          <template #body="{ data }">
            <span class="student-count-badge">
              <i class="pi pi-user"></i> {{ data.studentCount ?? 0 }}
            </span>
          </template>
        </Column>

        <Column field="createdAt" header="Data de creació" style="width: 10rem">
          <template #body="{ data }">
            {{ formatDate(data.createdAt) }}
          </template>
        </Column>

        <Column header="Accions" style="width: 11rem; text-align: right">
          <template #body="{ data }">
            <div class="actions-wrapper">
              <Button
                icon="pi pi-users"
                text
                rounded
                severity="info"
                size="small"
                tooltip="Gestionar Alumnes"
                @click="openStudentsModal(data)"
                data-testid="btn-manage-students"
              />
              <Button
                v-if="canManage"
                icon="pi pi-pencil"
                text
                rounded
                severity="secondary"
                size="small"
                tooltip="Editar Grup"
                @click="openEditModal(data)"
                data-testid="btn-edit-group"
              />
              <Button
                v-if="canManage"
                icon="pi pi-trash"
                text
                rounded
                severity="danger"
                size="small"
                tooltip="Esborrar Grup"
                @click="openDeleteConfirm(data)"
                data-testid="btn-delete-group"
              />
            </div>
          </template>
        </Column>
      </DataTable>
    </div>

    <!-- Modal Formulari Creació / Edició -->
    <Dialog
      v-model:visible="showFormModal"
      modal
      :header="isEditing ? 'Editar Grup' : 'Crear Nou Grup'"
      :style="{ width: '90vw', maxWidth: '500px' }"
      data-testid="group-form-dialog"
    >
      <div class="form-container">
        <div class="form-field">
          <label for="group-name" class="form-label">Nom del Grup *</label>
          <InputText
            id="group-name"
            v-model="formName"
            placeholder="Nom del grup (ex. Grup B&F, 1A...)"
            :class="{ 'p-invalid': formErrors.name }"
            class="full-width"
            data-testid="input-group-name"
          />
          <small v-if="formErrors.name" class="error-text">{{ formErrors.name }}</small>
        </div>

        <div class="form-field">
          <label for="group-academic-year" class="form-label">Curs Acadèmic</label>
          <InputText
            id="group-academic-year"
            v-model="formAcademicYear"
            placeholder="26-27"
            class="full-width"
            data-testid="input-group-academic-year"
          />
          <small class="hint-text">Proposta per defecte: "26-27"</small>
        </div>

        <div class="form-field">
          <label for="group-course-id" class="form-label">UUID Curs Associat (opcional)</label>
          <InputText
            id="group-course-id"
            v-model="formCourseId"
            placeholder="UUID del curs/assignatura..."
            class="full-width"
            data-testid="input-group-course-id"
          />
        </div>
      </div>

      <template #footer>
        <Button
          label="Cancel·lar"
          icon="pi pi-times"
          text
          severity="secondary"
          @click="showFormModal = false"
        />
        <Button
          :label="isEditing ? 'Actualitzar' : 'Crear Grup'"
          icon="pi pi-check"
          severity="primary"
          :loading="groupStore.isSaving"
          @click="handleSaveGroup"
          data-testid="btn-save-group"
        />
      </template>
    </Dialog>

    <!-- Modal Confirmació Esborrat -->
    <Dialog
      v-model:visible="showDeleteConfirmModal"
      modal
      header="Confirmar Baixa de Grup"
      :style="{ width: '90vw', maxWidth: '440px' }"
      data-testid="delete-group-dialog"
    >
      <div class="delete-confirm-content">
        <i class="pi pi-exclamation-triangle warning-icon" />
        <div>
          <p>
            Estàs segur que vols donar de baixa el grup
            <strong>{{ groupToDelete?.name }}</strong>?
          </p>
          <p class="soft-delete-hint">
            Aquesta acció realitza una baixa lògica (soft-delete) preservant la informació acadèmica dels alumnes.
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
          label="Confirmar Baixa"
          icon="pi pi-trash"
          severity="danger"
          :loading="groupStore.isLoading"
          @click="confirmDeleteGroup"
          data-testid="btn-confirm-delete-group"
        />
      </template>
    </Dialog>

    <!-- Modal Gestió d'Alumnes del Grup -->
    <Dialog
      v-model:visible="showStudentsModal"
      modal
      :header="`Alumnes del Grup: ${activeGroupForStudents?.name || ''}`"
      :style="{ width: '90vw', maxWidth: '640px' }"
      data-testid="manage-students-dialog"
    >
      <div class="students-modal-content">
        <div v-if="canManage" class="add-students-section">
          <h3>Assignar nous alumnes</h3>
          <p class="section-desc">Introdueix els UUIDs dels alumnes separats per comes o noves línies:</p>
          <textarea
            v-model="newStudentIdsText"
            placeholder="UUID1, UUID2..."
            rows="3"
            class="student-ids-textarea"
            data-testid="textarea-add-students"
          ></textarea>
          <div class="add-btn-wrapper">
            <Button
              label="Afegir Alumnes"
              icon="pi pi-user-plus"
              severity="primary"
              size="small"
              :loading="groupStore.isSaving"
              @click="handleAddStudents"
              data-testid="btn-add-students-submit"
            />
          </div>
          <small v-if="studentError" class="error-text mt-1">{{ studentError }}</small>
        </div>

        <div class="students-list-section">
          <h3>Llista d'alumnes assignats ({{ groupStore.groupStudents.length }})</h3>
          <div v-if="groupStore.groupStudents.length === 0" class="no-students">
            <p>No hi ha cap alumne assignat a aquest grup actualment.</p>
          </div>
          <ul v-else class="students-list">
            <li v-for="student in groupStore.groupStudents" :key="student.id" class="student-item">
              <div class="student-info">
                <i class="pi pi-user student-icon"></i>
                <div>
                  <div class="student-name">{{ student.fullName || 'Alumne sense nom' }}</div>
                  <div class="student-email">{{ student.email }}</div>
                </div>
              </div>
              <Button
                v-if="canManage"
                icon="pi pi-trash"
                text
                rounded
                severity="danger"
                size="small"
                tooltip="Desassignar"
                @click="handleRemoveStudent(student)"
                data-testid="btn-remove-student"
              />
            </li>
          </ul>
        </div>
      </div>
      <template #footer>
        <Button
          label="Tancar"
          icon="pi pi-times"
          text
          severity="secondary"
          @click="showStudentsModal = false"
        />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.groups-view-container {
  max-width: 1200px;
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
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  min-width: 240px;
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
}

.year-filter-input {
  width: 140px;
}

.table-card {
  background-color: #ffffff;
  border-radius: 0.75rem;
  border: 1px solid #e2e8f0;
  overflow: hidden;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
}

.group-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.group-icon-sm {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 0.5rem;
  background: #e0e7ff;
  color: #4338ca;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem;
}

.group-name {
  font-weight: 600;
  color: #1e293b;
}

.group-course-sub {
  font-size: 0.8rem;
  color: #64748b;
}

.badge-year {
  background-color: #f1f5f9;
  color: #334155;
  font-weight: 600;
  padding: 0.25rem 0.6rem;
  border-radius: 0.375rem;
  font-size: 0.85rem;
  border: 1px solid #cbd5e1;
}

.student-count-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.9rem;
  color: #475569;
}

.actions-wrapper {
  display: flex;
  justify-content: flex-end;
  gap: 0.25rem;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 3rem 1rem;
  color: #64748b;
}

.empty-icon {
  font-size: 2.5rem;
  color: #cbd5e1;
  margin-bottom: 0.75rem;
}

.form-container {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 0.5rem 0;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-label {
  font-weight: 600;
  font-size: 0.9rem;
  color: #1e293b;
}

.full-width {
  width: 100%;
}

.error-text {
  color: #ef4444;
  font-size: 0.8rem;
}

.hint-text {
  color: #64748b;
  font-size: 0.8rem;
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

.students-modal-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 0.5rem 0;
}

.add-students-section {
  background-color: #f8fafc;
  padding: 1rem;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
}

.add-students-section h3,
.students-list-section h3 {
  font-size: 1rem;
  font-weight: 600;
  margin: 0 0 0.5rem 0;
  color: #0f172a;
}

.section-desc {
  font-size: 0.85rem;
  color: #64748b;
  margin: 0 0 0.5rem 0;
}

.student-ids-textarea {
  width: 100%;
  border: 1px solid #cbd5e1;
  border-radius: 0.375rem;
  padding: 0.5rem;
  font-size: 0.9rem;
  font-family: inherit;
  resize: vertical;
}

.add-btn-wrapper {
  margin-top: 0.5rem;
  display: flex;
  justify-content: flex-end;
}

.no-students {
  padding: 1.5rem;
  text-align: center;
  color: #64748b;
  font-size: 0.9rem;
  background-color: #f8fafc;
  border-radius: 0.5rem;
}

.students-list {
  list-style: none;
  padding: 0;
  margin: 0;
  max-height: 250px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.student-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.6rem 0.8rem;
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
}

.student-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.student-icon {
  font-size: 1.1rem;
  color: #64748b;
  background-color: #f1f5f9;
  padding: 0.4rem;
  border-radius: 50%;
}

.student-name {
  font-weight: 600;
  font-size: 0.9rem;
  color: #1e293b;
}

.student-email {
  font-size: 0.8rem;
  color: #64748b;
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
  .filter-controls {
    flex-wrap: wrap;
  }
}
</style>
