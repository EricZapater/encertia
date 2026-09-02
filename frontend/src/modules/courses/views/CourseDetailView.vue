<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCourseStore } from '../store'
import { useAuthStore } from '@/modules/auth/store'
import { listUsers } from '@/modules/users/api'
import type { CourseUnit, CreateCourseUnitRequest, EnrolledStudent } from '../types'

import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import MultiSelect from 'primevue/multiselect'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import { useToast } from 'primevue/usetoast'
import type { User } from '@/modules/users/types'

const route = useRoute()
const router = useRouter()
const courseStore = useCourseStore()
const authStore = useAuthStore()
const toast = useToast()

const courseId = route.params.id as string
const activeTab = ref('units')

const canManage = computed(() => authStore.isAdmin || authStore.isTeacher)

// Modal Nova Unitat
const showCreateUnitModal = ref(false)
const newUnitTitle = ref('')
const newUnitDescription = ref('')
const createUnitError = ref('')

// Modal Matricular Alumnes
const showEnrollModal = ref(false)
const availableStudents = ref<User[]>([])
const selectedStudentIds = ref<string[]>([])
const isLoadingStudents = ref(false)

// Confirmació eliminació d'unitat
const showDeleteUnitConfirm = ref(false)
const unitToDelete = ref<CourseUnit | null>(null)

// Confirmació desmatriculació
const showUnenrollConfirm = ref(false)
const studentToUnenroll = ref<EnrolledStudent | null>(null)

watch(() => courseStore.error, (err) => {
  if (err) {
    toast.add({ severity: 'error', summary: 'Error', detail: err, life: 4000 })
    courseStore.clearError()
  }
})

onMounted(async () => {
  if (courseId) {
    await courseStore.fetchCourseDetail(courseId)
    await courseStore.fetchCourseStudents(courseId)
  }
})

function statusSeverity(status?: string) {
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

function statusLabel(status?: string) {
  switch (status) {
    case 'active':
      return 'Actiu'
    case 'draft':
      return 'Esborrany'
    case 'archived':
      return 'Arxivat'
    default:
      return status || ''
  }
}

// Unitats
function openCreateUnitModal() {
  newUnitTitle.value = ''
  newUnitDescription.value = ''
  createUnitError.value = ''
  showCreateUnitModal.value = true
}

async function handleCreateUnit() {
  if (!newUnitTitle.value.trim()) {
    createUnitError.value = 'El títol de la unitat és obligatori'
    return
  }

  try {
    const currentUnits = courseStore.currentCourse?.units || []
    const nextOrderIndex = currentUnits.length
    const payload: CreateCourseUnitRequest = {
      title: newUnitTitle.value.trim(),
      description: newUnitDescription.value.trim() || undefined,
      orderIndex: nextOrderIndex
    }
    await courseStore.createCourseUnit(courseId, payload)
    toast.add({ severity: 'success', summary: 'Èxit', detail: 'Unitat didàctica creada', life: 3000 })
    showCreateUnitModal.value = false
  } catch (_e) {
    // Handled by store
  }
}

function confirmDeleteUnit(unit: CourseUnit) {
  unitToDelete.value = unit
  showDeleteUnitConfirm.value = true
}

async function handleDeleteUnit() {
  if (!unitToDelete.value) return
  try {
    await courseStore.deleteCourseUnit(courseId, unitToDelete.value.id)
    toast.add({ severity: 'success', summary: 'Eliminada', detail: 'Unitat didàctica eliminada', life: 3000 })
    showDeleteUnitConfirm.value = false
  } catch (_e) {
    // Handled by store
  }
}

async function moveUnit(index: number, direction: 'up' | 'down') {
  const units = [...(courseStore.currentCourse?.units || [])]
  const targetIndex = direction === 'up' ? index - 1 : index + 1
  if (targetIndex < 0 || targetIndex >= units.length) return

  const temp = units[index]
  units[index] = units[targetIndex]
  units[targetIndex] = temp

  const reorderedIds = units.map((u) => u.id)
  try {
    await courseStore.reorderCourseUnits(courseId, reorderedIds)
    toast.add({ severity: 'success', summary: 'Reordenat', detail: 'Ordre de les unitats actualitzat', life: 2000 })
  } catch (_e) {
    // Handled by store
  }
}

function navigateToUnitEdit(unitId: string) {
  router.push(`/courses/${courseId}/units/${unitId}/edit`)
}

function navigateToUnitPlay(unitId: string) {
  router.push(`/courses/${courseId}/units/${unitId}/play`)
}

// Matriculació
async function openEnrollModal() {
  selectedStudentIds.value = []
  showEnrollModal.value = true
  isLoadingStudents.value = true
  try {
    const res = await listUsers({ role: 'student', pageSize: 100 })
    // Filtrem els que ja estan matriculats
    const alreadyEnrolledIds = new Set(courseStore.enrolledStudents.map((s) => s.id))
    availableStudents.value = (res.items || []).filter((u) => !alreadyEnrolledIds.has(u.id))
  } catch (_e) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Error en carregar alumnes', life: 3000 })
  } finally {
    isLoadingStudents.value = false
  }
}

async function handleEnrollStudents() {
  if (selectedStudentIds.value.length === 0) return
  try {
    await courseStore.enrollStudents(courseId, selectedStudentIds.value)
    toast.add({ severity: 'success', summary: 'Èxit', detail: 'Alumnes matriculats correctament', life: 3000 })
    showEnrollModal.value = false
  } catch (_e) {
    // Handled by store
  }
}

function confirmUnenrollStudent(student: EnrolledStudent) {
  studentToUnenroll.value = student
  showUnenrollConfirm.value = true
}

async function handleUnenrollStudent() {
  if (!studentToUnenroll.value) return
  try {
    await courseStore.unenrollStudent(courseId, studentToUnenroll.value.id)
    toast.add({ severity: 'success', summary: 'Desmatriculat', detail: 'Alumne desmatriculat del curs', life: 3000 })
    showUnenrollConfirm.value = false
  } catch (_e) {
    // Handled by store
  }
}

function goBack() {
  router.push('/courses')
}
</script>

<template>
  <div class="course-detail-view">
    <!-- Top Bar Navigation -->
    <div class="top-nav">
      <Button
        label="Tornar al llistat"
        icon="pi pi-arrow-left"
        severity="secondary"
        text
        @click="goBack"
      />
    </div>

    <!-- Loading State -->
    <div v-if="courseStore.isLoading && !courseStore.currentCourse" class="loading-container">
      <i class="pi pi-spin pi-spinner spinner-icon"></i>
      <p>Carregant el curs...</p>
    </div>

    <template v-else-if="courseStore.currentCourse">
      <!-- Course Summary Header -->
      <header class="course-summary-card">
        <div class="summary-main">
          <div class="summary-title-wrapper">
            <span class="course-code-badge">{{ courseStore.currentCourse.code }}</span>
            <h1 class="course-title">{{ courseStore.currentCourse.title }}</h1>
            <Tag
              :severity="statusSeverity(courseStore.currentCourse.status)"
              :value="statusLabel(courseStore.currentCourse.status)"
            />
          </div>
          <p class="course-description">
            {{ courseStore.currentCourse.description || 'Sense descripció.' }}
          </p>
        </div>

        <div class="summary-details">
          <div v-if="courseStore.currentCourse.teacherName" class="detail-item">
            <i class="pi pi-user detail-icon"></i>
            <div>
              <div class="detail-label">Professorat</div>
              <div class="detail-value">{{ courseStore.currentCourse.teacherName }}</div>
            </div>
          </div>

          <div class="detail-item">
            <i class="pi pi-list detail-icon"></i>
            <div>
              <div class="detail-label">Unitats Didàctiques</div>
              <div class="detail-value">{{ courseStore.currentCourse.units?.length || 0 }} unitats</div>
            </div>
          </div>

          <div class="detail-item">
            <i class="pi pi-users detail-icon"></i>
            <div>
              <div class="detail-label">Alumnes Matriculats</div>
              <div class="detail-value">{{ courseStore.enrolledStudents.length }} alumnes</div>
            </div>
          </div>
        </div>
      </header>

      <!-- Tabs Section -->
      <div class="tabs-container">
        <Tabs v-model:value="activeTab">
          <TabList>
            <Tab value="units">
              <i class="pi pi-book mr-2"></i>
              Unitats Didàctiques ({{ courseStore.currentCourse.units?.length || 0 }})
            </Tab>
            <Tab value="students">
              <i class="pi pi-users mr-2"></i>
              Alumnes Matriculats ({{ courseStore.enrolledStudents.length }})
            </Tab>
          </TabList>

          <TabPanels>
            <!-- TAB 1: UNITATS DIDÀCTIQUES -->
            <TabPanel value="units">
              <div class="tab-actions">
                <h2 class="section-title">Programa del Curs</h2>
                <Button
                  v-if="canManage"
                  label="Nova Unitat"
                  icon="pi pi-plus"
                  severity="primary"
                  @click="openCreateUnitModal"
                />
              </div>

              <!-- Units List -->
              <div
                v-if="!courseStore.currentCourse.units || courseStore.currentCourse.units.length === 0"
                class="empty-tab-state"
              >
                <i class="pi pi-book empty-tab-icon"></i>
                <p>Aquest curs encara no té cap unitat didàctica.</p>
                <Button
                  v-if="canManage"
                  label="Crear la primera unitat"
                  icon="pi pi-plus"
                  severity="primary"
                  class="mt-2"
                  @click="openCreateUnitModal"
                />
              </div>

              <div v-else class="units-list">
                <div
                  v-for="(unit, idx) in courseStore.currentCourse.units"
                  :key="unit.id"
                  class="unit-card"
                >
                  <div class="unit-index">
                    <span>{{ idx + 1 }}</span>
                  </div>

                  <div class="unit-info">
                    <h3 class="unit-title">{{ unit.title }}</h3>
                    <p class="unit-desc">{{ unit.description || 'Sense descripció.' }}</p>
                    <div class="unit-meta">
                      <span class="meta-tag">
                        <i class="pi pi-file mr-1"></i> {{ unit.blocksCount ?? 0 }} blocs de guió
                      </span>
                      <span class="meta-tag">
                        <i class="pi pi-question-circle mr-1"></i> {{ unit.quizzesCount ?? 0 }} qüestionaris
                      </span>
                    </div>
                  </div>

                  <div class="unit-actions">
                    <!-- Unit Order Buttons -->
                    <div v-if="canManage" class="reorder-btns">
                      <Button
                        icon="pi pi-chevron-up"
                        severity="secondary"
                        text
                        size="small"
                        :disabled="idx === 0"
                        title="Moure amunt"
                        @click="moveUnit(idx, 'up')"
                      />
                      <Button
                        icon="pi pi-chevron-down"
                        severity="secondary"
                        text
                        size="small"
                        :disabled="idx === courseStore.currentCourse.units.length - 1"
                        title="Moure avall"
                        @click="moveUnit(idx, 'down')"
                      />
                    </div>

                    <!-- Play Script Button -->
                    <Button
                      label="Guió de classe"
                      icon="pi pi-play"
                      severity="success"
                      size="small"
                      @click="navigateToUnitPlay(unit.id)"
                    />

                    <!-- Edit Unit Button (Teachers/Admins) -->
                    <Button
                      v-if="canManage"
                      label="Editar"
                      icon="pi pi-pencil"
                      severity="secondary"
                      size="small"
                      @click="navigateToUnitEdit(unit.id)"
                    />

                    <!-- Delete Unit Button -->
                    <Button
                      v-if="canManage"
                      icon="pi pi-trash"
                      severity="danger"
                      text
                      rounded
                      size="small"
                      title="Eliminar unitat"
                      @click="confirmDeleteUnit(unit)"
                    />
                  </div>
                </div>
              </div>
            </TabPanel>

            <!-- TAB 2: ALUMNES MATRICULATS -->
            <TabPanel value="students">
              <div class="tab-actions">
                <h2 class="section-title">Llista d'Alumnes Inscrits</h2>
                <Button
                  v-if="canManage"
                  label="Matricular Alumnes"
                  icon="pi pi-user-plus"
                  severity="primary"
                  @click="openEnrollModal"
                />
              </div>

              <div v-if="courseStore.enrolledStudents.length === 0" class="empty-tab-state">
                <i class="pi pi-users empty-tab-icon"></i>
                <p>No hi ha cap alumne matriculat en aquest curs.</p>
                <Button
                  v-if="canManage"
                  label="Matricular alumnes"
                  icon="pi pi-user-plus"
                  severity="primary"
                  class="mt-2"
                  @click="openEnrollModal"
                />
              </div>

              <div v-else class="students-table-wrapper">
                <table class="students-table">
                  <thead>
                    <tr>
                      <th>Nom de l'Alumne</th>
                      <th>Correu Electrònic</th>
                      <th>Data de Matriculació</th>
                      <th v-if="canManage" class="text-right">Accions</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="student in courseStore.enrolledStudents" :key="student.id">
                      <td>
                        <div class="student-name-col">
                          <i class="pi pi-user student-avatar"></i>
                          <span>
                            {{ student.firstName || student.lastName ? `${student.firstName || ''} ${student.lastName || ''}` : 'Alumne' }}
                          </span>
                        </div>
                      </td>
                      <td>{{ student.email }}</td>
                      <td>{{ new Date(student.enrolledAt).toLocaleDateString() }}</td>
                      <td v-if="canManage" class="text-right">
                        <Button
                          icon="pi pi-user-minus"
                          label="Desmatricular"
                          severity="danger"
                          text
                          size="small"
                          @click="confirmUnenrollStudent(student)"
                        />
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </div>
    </template>

    <!-- Modal Nova Unitat Didàctica -->
    <Dialog
      v-model:visible="showCreateUnitModal"
      header="Crear Unitat Didàctica"
      :modal="true"
      :style="{ width: '480px' }"
    >
      <div class="form-container">
        <div class="form-field">
          <label for="unit-title">Títol de la unitat *</label>
          <InputText
            id="unit-title"
            v-model="newUnitTitle"
            placeholder="Ex: Tema 1: Fonaments i Conceptes Clau"
            :class="{ 'p-invalid': createUnitError }"
          />
          <small v-if="createUnitError" class="p-error">{{ createUnitError }}</small>
        </div>

        <div class="form-field">
          <label for="unit-desc">Descripció</label>
          <Textarea
            id="unit-desc"
            v-model="newUnitDescription"
            rows="3"
            placeholder="Resum dels continguts o objectius d'aquesta unitat..."
          />
        </div>
      </div>
      <template #footer>
        <Button label="Cancel·lar" severity="secondary" text @click="showCreateUnitModal = false" />
        <Button
          label="Crear Unitat"
          icon="pi pi-check"
          severity="primary"
          :loading="courseStore.isSaving"
          @click="handleCreateUnit"
        />
      </template>
    </Dialog>

    <!-- Modal Matricular Alumnes -->
    <Dialog
      v-model:visible="showEnrollModal"
      header="Matricular Alumnes al Curs"
      :modal="true"
      :style="{ width: '520px' }"
    >
      <div class="form-container">
        <p class="modal-intro">
          Selecciona un o més alumnes registrats a la plataforma per afegir-los al curs.
        </p>

        <div v-if="isLoadingStudents" class="text-center py-4">
          <i class="pi pi-spin pi-spinner text-2xl text-indigo-600"></i>
        </div>

        <div v-else-if="availableStudents.length === 0" class="empty-available">
          <p>Tots els alumnes de la plataforma ja estan matriculats en aquest curs.</p>
        </div>

        <div v-else class="form-field">
          <label for="student-select">Selecciona alumnes</label>
          <MultiSelect
            id="student-select"
            v-model="selectedStudentIds"
            :options="availableStudents"
            optionLabel="email"
            optionValue="id"
            placeholder="Tria els alumnes..."
            filter
            class="w-full"
            display="chip"
          >
            <template #option="slotProps">
              <div class="student-option">
                <span>{{ slotProps.option.firstName }} {{ slotProps.option.lastName }}</span>
                <small class="text-gray-500">({{ slotProps.option.email }})</small>
              </div>
            </template>
          </MultiSelect>
        </div>
      </div>
      <template #footer>
        <Button label="Cancel·lar" severity="secondary" text @click="showEnrollModal = false" />
        <Button
          label="Matricular"
          icon="pi pi-user-plus"
          severity="primary"
          :disabled="selectedStudentIds.length === 0"
          :loading="courseStore.isSaving"
          @click="handleEnrollStudents"
        />
      </template>
    </Dialog>

    <!-- Modal Confirmar Esborrar Unitat -->
    <Dialog
      v-model:visible="showDeleteUnitConfirm"
      header="Confirmar Eliminació d'Unitat"
      :modal="true"
      :style="{ width: '420px' }"
    >
      <div class="confirm-content">
        <i class="pi pi-exclamation-triangle warning-icon"></i>
        <p>
          Vols eliminar la unitat didàctica <strong>"{{ unitToDelete?.title }}"</strong>?
        </p>
      </div>
      <template #footer>
        <Button label="Cancel·lar" severity="secondary" text @click="showDeleteUnitConfirm = false" />
        <Button
          label="Eliminar"
          icon="pi pi-trash"
          severity="danger"
          :loading="courseStore.isLoading"
          @click="handleDeleteUnit"
        />
      </template>
    </Dialog>

    <!-- Modal Confirmar Desmatricular -->
    <Dialog
      v-model:visible="showUnenrollConfirm"
      header="Confirmar Desmatriculació"
      :modal="true"
      :style="{ width: '420px' }"
    >
      <div class="confirm-content">
        <i class="pi pi-exclamation-triangle warning-icon"></i>
        <p>
          Estàs segur que vols desmatricular l'alumne
          <strong>{{ studentToUnenroll?.email }}</strong> del curs?
        </p>
      </div>
      <template #footer>
        <Button label="Cancel·lar" severity="secondary" text @click="showUnenrollConfirm = false" />
        <Button
          label="Desmatricular"
          icon="pi pi-user-minus"
          severity="danger"
          :loading="courseStore.isSaving"
          @click="handleUnenrollStudent"
        />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.course-detail-view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 1.5rem;
}

.top-nav {
  margin-bottom: 1rem;
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

.course-summary-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.summary-title-wrapper {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
  flex-wrap: wrap;
}

.course-code-badge {
  font-weight: 700;
  font-size: 0.9rem;
  color: #4f46e5;
  background: #e0e7ff;
  padding: 0.25rem 0.6rem;
  border-radius: 0.375rem;
}

.course-title {
  font-size: 1.6rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.course-description {
  color: #475569;
  font-size: 0.95rem;
  margin: 0;
  line-height: 1.5;
}

.summary-details {
  display: flex;
  gap: 2rem;
  padding-top: 1rem;
  border-top: 1px solid #f1f5f9;
  flex-wrap: wrap;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.detail-icon {
  font-size: 1.25rem;
  color: #6366f1;
  background: #f0fdf4;
  padding: 0.5rem;
  border-radius: 0.5rem;
}

.detail-label {
  font-size: 0.75rem;
  color: #64748b;
  text-transform: uppercase;
  font-weight: 600;
}

.detail-value {
  font-weight: 600;
  color: #1e293b;
  font-size: 0.95rem;
}

.tabs-container {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1.25rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.tab-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
  margin-top: 0.5rem;
}

.section-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.empty-tab-state {
  text-align: center;
  padding: 3rem 1rem;
  color: #64748b;
  border: 1px dashed #e2e8f0;
  border-radius: 0.5rem;
}

.empty-tab-icon {
  font-size: 2.5rem;
  color: #94a3b8;
  margin-bottom: 0.75rem;
}

.units-list {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.unit-card {
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  transition: background-color 0.15s ease;
}

.unit-card:hover {
  background-color: #f8fafc;
}

.unit-index {
  width: 2.25rem;
  height: 2.25rem;
  background: #e0e7ff;
  color: #4338ca;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.95rem;
  flex-shrink: 0;
}

.unit-info {
  flex: 1;
}

.unit-title {
  font-size: 1.05rem;
  font-weight: 600;
  color: #0f172a;
  margin: 0 0 0.25rem 0;
}

.unit-desc {
  font-size: 0.875rem;
  color: #64748b;
  margin: 0 0 0.5rem 0;
}

.unit-meta {
  display: flex;
  gap: 1rem;
}

.meta-tag {
  font-size: 0.775rem;
  color: #475569;
  background: #f1f5f9;
  padding: 0.15rem 0.4rem;
  border-radius: 0.25rem;
}

.unit-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.reorder-btns {
  display: flex;
  flex-direction: column;
  margin-right: 0.25rem;
}

.students-table-wrapper {
  overflow-x: auto;
}

.students-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.students-table th {
  background: #f8fafc;
  color: #475569;
  font-weight: 600;
  font-size: 0.85rem;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #e2e8f0;
}

.students-table td {
  padding: 0.85rem 1rem;
  border-bottom: 1px solid #f1f5f9;
  font-size: 0.9rem;
  color: #1e293b;
}

.student-name-col {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
}

.student-avatar {
  background: #f1f5f9;
  color: #64748b;
  padding: 0.35rem;
  border-radius: 50%;
  font-size: 0.85rem;
}

.student-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
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
