<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useMaterialStore } from '../store'
import { useAuthStore } from '@/modules/auth/store'
import type { Material, MaterialType } from '../types'

import MaterialFormModal from '../components/MaterialFormModal.vue'
import PdfViewerModal from '../components/PdfViewerModal.vue'
import MaterialViewsReportModal from '../components/MaterialViewsReportModal.vue'

import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Paginator from 'primevue/paginator'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { useI18n } from 'vue-i18n'

const materialStore = useMaterialStore()
const authStore = useAuthStore()
const toast = useToast()
useI18n()

const canManageMaterials = computed(() => authStore.isAdmin || authStore.isTeacher)

// State for Modals
const showFormModal = ref(false)
const materialToEdit = ref<Material | null>(null)

const showViewerModal = ref(false)
const selectedMaterialForViewer = ref<Material | null>(null)

const showReportModal = ref(false)
const selectedMaterialForReport = ref<Material | null>(null)

const showDeleteConfirm = ref(false)
const materialToDelete = ref<Material | null>(null)

// Search & Filter State
const searchInput = ref('')
const selectedType = ref<MaterialType | 'all'>('all')

const typeFilterOptions = [
  { label: 'Tots els tipus', value: 'all' },
  { label: 'Documents (PDF/Word)', value: 'document' },
  { label: 'Vídeos', value: 'video' }
]

watch(() => materialStore.error, (err) => {
  if (err) {
    toast.add({ severity: 'error', summary: 'Error', detail: err, life: 4000 })
    materialStore.clearError()
  }
})

onMounted(() => {
  materialStore.fetchMaterials()
})

function handleSearch() {
  materialStore.setSearch(searchInput.value)
}

function handleFilterTypeChange() {
  const filter = selectedType.value === 'all' ? undefined : (selectedType.value as MaterialType)
  materialStore.setTypeFilter(filter)
}

function handlePageChange(event: { page: number; rows: number }) {
  materialStore.setPage(event.page + 1, event.rows)
}

function openCreateModal() {
  materialToEdit.value = null
  showFormModal.value = true
}

function openEditModal(mat: Material) {
  materialToEdit.value = mat
  showFormModal.value = true
}

function openViewerModal(mat: Material) {
  selectedMaterialForViewer.value = mat
  showViewerModal.value = true
}

function openReportModal(mat: Material) {
  selectedMaterialForReport.value = mat
  showReportModal.value = true
}

function confirmDelete(mat: Material) {
  materialToDelete.value = mat
  showDeleteConfirm.value = true
}

async function handleDelete() {
  if (!materialToDelete.value) return
  try {
    await materialStore.deleteMaterial(materialToDelete.value.id)
    toast.add({ severity: 'success', summary: 'Eliminat', detail: 'Material eliminat correctament', life: 3000 })
    showDeleteConfirm.value = false
  } catch (_e) {
    // Handled in store
  }
}

function formatFileSize(bytes?: number | null): string {
  if (!bytes) return ''
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
</script>

<template>
  <div class="materials-list-view">
    <!-- Header -->
    <div class="header-section">
      <div>
        <h1 class="page-title">{{ $t('materials.title') }}</h1>
        <p class="page-subtitle">
          {{ $t('materials.subtitle') }}
        </p>
      </div>

      <Button
        v-if="canManageMaterials"
        :label="$t('materials.upload')"
        icon="pi pi-plus"
        severity="primary"
        @click="openCreateModal"
      />
    </div>

    <!-- Filters Bar -->
    <div class="filters-card">
      <div class="filters-row">
        <div class="search-box">
          <i class="pi pi-search search-icon"></i>
          <InputText
            v-model="searchInput"
            placeholder="Cercar per títol o descripció..."
            class="search-input"
            @keyup.enter="handleSearch"
          />
          <Button
            label="Cercar"
            severity="secondary"
            size="small"
            @click="handleSearch"
          />
        </div>

        <div class="type-filter-box">
          <label for="type-filter" class="filter-label">Tipus:</label>
          <Select
            id="type-filter"
            v-model="selectedType"
            :options="typeFilterOptions"
            optionLabel="label"
            optionValue="value"
            class="type-select"
            @change="handleFilterTypeChange"
          />
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="materialStore.isLoading" class="loading-state">
      <i class="pi pi-spin pi-spinner spinner-icon"></i>
      <p>Carregant els materials didàctics...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="!materialStore.hasMaterials" class="empty-state">
      <i class="pi pi-folder-open empty-icon"></i>
      <h3>Cap material trobat</h3>
      <p>No hi ha materials didàctics disponibles amb els filtres seleccionats.</p>
      <Button
        v-if="canManageMaterials"
        label="Crear el primer material"
        icon="pi pi-plus"
        severity="primary"
        class="mt-3"
        @click="openCreateModal"
      />
    </div>

    <!-- Materials Grid -->
    <div v-else class="materials-grid">
      <div
        v-for="mat in materialStore.materialList"
        :key="mat.id"
        class="material-card"
      >
        <div class="card-header">
          <div class="type-badge-row">
            <Tag
              :severity="mat.materialType === 'document' ? 'info' : 'success'"
              class="type-tag"
            >
              <i :class="mat.materialType === 'document' ? 'pi pi-file-pdf' : 'pi pi-video'" class="mr-1"></i>
              {{ mat.materialType === 'document' ? 'DOCUMENT' : 'VÍDEO' }}
            </Tag>
            <span v-if="mat.fileSizeBytes" class="file-size">{{ formatFileSize(mat.fileSizeBytes) }}</span>
          </div>

          <h2 class="material-title" :title="mat.title">{{ mat.title }}</h2>
        </div>

        <p class="material-desc">
          {{ mat.description || 'Sense descripció.' }}
        </p>

        <div class="material-meta">
          <div v-if="mat.materialType === 'document'" class="meta-item">
            <i class="pi pi-book"></i>
            <span>{{ mat.pageCount || 1 }} pàgines</span>
          </div>
          <div v-else-if="mat.materialType === 'video'" class="meta-item">
            <i class="pi pi-youtube"></i>
            <span class="capitalize">{{ mat.videoProvider || 'Vídeo' }}</span>
          </div>
        </div>

        <div class="card-actions">
          <Button
            label="Visualitzar"
            icon="pi pi-eye"
            severity="primary"
            size="small"
            class="flex-1"
            @click="openViewerModal(mat)"
          />

          <template v-if="canManageMaterials">
            <Button
              icon="pi pi-chart-bar"
              severity="secondary"
              text
              rounded
              size="small"
              title="Informe d'Accessos"
              @click="openReportModal(mat)"
            />
            <Button
              icon="pi pi-pencil"
              severity="secondary"
              text
              rounded
              size="small"
              title="Editar"
              @click="openEditModal(mat)"
            />
            <Button
              icon="pi pi-trash"
              severity="danger"
              text
              rounded
              size="small"
              title="Eliminar"
              @click="confirmDelete(mat)"
            />
          </template>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="materialStore.totalCount > materialStore.pageSize" class="pagination-box">
      <Paginator
        :rows="materialStore.pageSize"
        :totalRecords="materialStore.totalCount"
        :first="(materialStore.currentPage - 1) * materialStore.pageSize"
        @page="handlePageChange"
      />
    </div>

    <!-- Modals -->
    <MaterialFormModal
      v-model:visible="showFormModal"
      :materialToEdit="materialToEdit"
      @saved="materialStore.fetchMaterials"
    />

    <PdfViewerModal
      v-model:visible="showViewerModal"
      :material="selectedMaterialForViewer"
      :trackView="true"
    />

    <MaterialViewsReportModal
      v-if="selectedMaterialForReport"
      v-model:visible="showReportModal"
      :materialId="selectedMaterialForReport.id"
      :materialTitle="selectedMaterialForReport.title"
    />

    <!-- Dialog Confirm Delete -->
    <Dialog
      v-model:visible="showDeleteConfirm"
      header="Confirmar Eliminació"
      :modal="true"
      :style="{ width: '420px' }"
    >
      <div class="confirmation-content">
        <i class="pi pi-exclamation-triangle warning-icon"></i>
        <span>Estàs segur que vols eliminar el material <strong>{{ materialToDelete?.title }}</strong>?</span>
      </div>
      <template #footer>
        <Button label="Cancel·lar" severity="secondary" text @click="showDeleteConfirm = false" />
        <Button
          label="Eliminar"
          icon="pi pi-trash"
          severity="danger"
          :loading="materialStore.isLoading"
          @click="handleDelete"
        />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.materials-list-view {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1.5rem;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}

.page-title {
  font-size: 1.75rem;
  font-weight: 800;
  color: #0f172a;
  margin: 0;
}

.page-subtitle {
  color: #64748b;
  font-size: 0.95rem;
  margin: 0.25rem 0 0 0;
}

.filters-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1rem;
  margin-bottom: 1.5rem;
}

.filters-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  position: relative;
  flex: 1;
  min-width: 280px;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  color: #94a3b8;
}

.search-input {
  padding-left: 2.25rem !important;
  width: 100%;
}

.type-filter-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #475569;
}

.type-select {
  min-width: 200px;
}

.loading-state,
.empty-state {
  text-align: center;
  padding: 4rem 1rem;
  background: #ffffff;
  border-radius: 0.75rem;
  border: 1px dashed #cbd5e1;
}

.spinner-icon,
.empty-icon {
  font-size: 3rem;
  color: #6366f1;
  margin-bottom: 1rem;
}

.empty-icon {
  color: #94a3b8;
}

.materials-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.25rem;
  margin-bottom: 1.5rem;
}

.material-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.material-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.type-badge-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.type-tag {
  font-size: 0.7rem !important;
  letter-spacing: 0.03em;
}

.file-size {
  font-size: 0.75rem;
  color: #64748b;
}

.material-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: #1e293b;
  margin: 0 0 0.5rem 0;
  line-clamp: 2;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.material-desc {
  font-size: 0.875rem;
  color: #475569;
  margin: 0 0 1rem 0;
  line-clamp: 2;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.material-meta {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.25rem;
  font-size: 0.8rem;
  color: #64748b;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding-top: 0.75rem;
  border-top: 1px solid #f1f5f9;
}

.pagination-box {
  display: flex;
  justify-content: center;
  margin-top: 1.5rem;
}

.confirmation-content {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 0;
}

.warning-icon {
  font-size: 2rem;
  color: #ef4444;
}
</style>
