<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/modules/auth/store'
import { useUserStore } from '../store'
import type { User, UserRole, UserStatusFilter } from '../types'

import DataTable, { type DataTablePageEvent } from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'

import UserFormModal from './UserFormModal.vue'
import ResetPasswordModal from './ResetPasswordModal.vue'
import BatchImportModal from './BatchImportModal.vue'

const router = useRouter()
const authStore = useAuthStore()
const userStore = useUserStore()

const isAdmin = computed(() => authStore.currentUser?.role === 'admin')
const isTeacher = computed(() => authStore.currentUser?.role === 'teacher')

// Filtres locals
const searchInput = ref('')
let searchTimeout: ReturnType<typeof setTimeout> | null = null

const selectedRole = ref<UserRole | undefined>(undefined)
const selectedStatus = ref<UserStatusFilter>('active')

const roleFilterOptions = computed(() => {
  if (isAdmin.value) {
    return [
      { label: 'Tots els rols', value: undefined },
      { label: 'Administradors', value: 'admin' as UserRole },
      { label: 'Professors', value: 'teacher' as UserRole },
      { label: 'Alumnes', value: 'student' as UserRole }
    ]
  }
  // Si és professor només gestiona alumnes
  return [{ label: 'Alumnes', value: 'student' as UserRole }]
})

const statusFilterOptions = [
  { label: 'Actius', value: 'active' as UserStatusFilter },
  { label: 'Inactius', value: 'inactive' as UserStatusFilter },
  { label: 'Tots', value: 'all' as UserStatusFilter }
]

// Modals i estats
const showFormModal = ref(false)
const showResetPasswordModal = ref(false)
const showBatchImportModal = ref(false)
const showDeleteConfirmModal = ref(false)

const selectedUser = ref<User | null>(null)
const userToDelete = ref<User | null>(null)

const successFeedback = ref<string | null>(null)
const errorFeedback = ref<string | null>(null)

// Carrega inicial
onMounted(() => {
  // Si és teacher, forcem el filtre de rol a student
  if (isTeacher.value) {
    selectedRole.value = 'student'
    userStore.setRoleFilter('student')
  } else {
    userStore.fetchUsers()
  }
})

// Gestors de cerca i filtres
function handleSearchInput() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    userStore.setSearch(searchInput.value)
  }, 350)
}

function handleRoleChange() {
  userStore.setRoleFilter(selectedRole.value)
}

function handleStatusChange() {
  userStore.setStatusFilter(selectedStatus.value)
}

function handlePageChange(event: DataTablePageEvent) {
  const newPage = Math.floor(event.first / event.rows) + 1
  userStore.setPage(newPage, event.rows)
}

function handleResetFilters() {
  searchInput.value = ''
  selectedStatus.value = 'active'
  if (isTeacher.value) {
    selectedRole.value = 'student'
    userStore.resetFilters()
    userStore.setRoleFilter('student')
  } else {
    selectedRole.value = undefined
    userStore.resetFilters()
  }
}

// Accions sobre usuaris
function openCreateModal() {
  selectedUser.value = null
  showFormModal.value = true
}

function openEditModal(user: User) {
  selectedUser.value = { ...user }
  showFormModal.value = true
}

function openResetPasswordModal(user: User) {
  selectedUser.value = { ...user }
  showResetPasswordModal.value = true
}

function openDeleteConfirm(user: User) {
  userToDelete.value = user
  showDeleteConfirmModal.value = true
}

async function confirmDeleteUser() {
  if (!userToDelete.value) return

  try {
    const msg = await userStore.deleteUser(userToDelete.value.id)
    successFeedback.value = msg || 'Usuari donat de baixa correctament.'
    showDeleteConfirmModal.value = false
    userToDelete.value = null
  } catch (err: any) {
    errorFeedback.value =
      err.response?.data?.error?.message ||
      err.message ||
      'Error en donar de baixa l’usuari.'
  }
}

function handleUserSaved(savedUser: User) {
  successFeedback.value = `Usuari "${savedUser.firstName} ${savedUser.lastName}" desat correctament.`
}

function handleBatchImported() {
  successFeedback.value = 'Procés d’importació massiva completat.'
}

function navigateToDetail(user: User) {
  router.push(`/users/${user.id}`)
}

function getRoleSeverity(role: UserRole) {
  switch (role) {
    case 'admin':
      return 'danger'
    case 'teacher':
      return 'info'
    case 'student':
      return 'secondary'
    default:
      return undefined
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
  <div class="users-view-container">
    <!-- Capçalera de pàgina -->
    <div class="page-header">
      <div>
        <h1 class="page-title">Gestió d'Usuaris</h1>
        <p class="page-subtitle">
          {{ isAdmin ? 'Administra tots els comptes de la plataforma Encertia' : 'Gestiona els alumnes matriculats' }}
        </p>
      </div>

      <div class="header-actions">
        <Button
          label="Importar CSV"
          icon="pi pi-file-import"
          severity="secondary"
          outlined
          @click="showBatchImportModal = true"
          data-testid="btn-open-batch-import"
        />
        <Button
          label="Nou Usuari"
          icon="pi pi-user-plus"
          severity="primary"
          @click="openCreateModal"
          data-testid="btn-open-create-user"
        />
      </div>
    </div>

    <!-- Alertes de feedback -->
    <Message
      v-if="successFeedback"
      severity="success"
      :closable="true"
      class="mb-3"
      @close="successFeedback = null"
      data-testid="msg-success-feedback"
    >
      {{ successFeedback }}
    </Message>

    <Message
      v-if="errorFeedback || userStore.error"
      severity="error"
      :closable="true"
      class="mb-3"
      @close="errorFeedback = null; userStore.clearError()"
      data-testid="msg-error-feedback"
    >
      {{ errorFeedback || userStore.error }}
    </Message>

    <!-- Barra de filtres -->
    <div class="filters-bar">
      <div class="search-input-wrapper">
        <i class="pi pi-search search-icon" />
        <InputText
          v-model="searchInput"
          placeholder="Cercar per nom o correu..."
          class="search-input"
          @input="handleSearchInput"
          data-testid="input-search-users"
        />
        <Button
          v-if="searchInput"
          icon="pi pi-times"
          text
          rounded
          severity="secondary"
          class="clear-search-btn"
          @click="searchInput = ''; userStore.setSearch('')"
        />
      </div>

      <div class="filter-controls">
        <Select
          v-if="isAdmin"
          v-model="selectedRole"
          :options="roleFilterOptions"
          optionLabel="label"
          optionValue="value"
          placeholder="Filtrar per rol"
          class="filter-select"
          @change="handleRoleChange"
          data-testid="filter-role-select"
        />

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
          :loading="userStore.isLoading"
          tooltip="Refrescar llista"
          @click="userStore.fetchUsers()"
          data-testid="btn-refresh-users"
        />
      </div>
    </div>

    <!-- Taula d'Usuaris -->
    <div class="table-card">
      <DataTable
        :value="userStore.userList"
        :loading="userStore.isLoading"
        lazy
        paginator
        :rows="userStore.pageSize"
        :totalRecords="userStore.totalUsers"
        :first="(userStore.currentPage - 1) * userStore.pageSize"
        :rowsPerPageOptions="[10, 20, 50, 100]"
        @page="handlePageChange"
        tableStyle="min-width: 60rem"
        stripedRows
        data-testid="users-data-table"
      >
        <template #empty>
          <div class="empty-state">
            <i class="pi pi-users empty-icon" />
            <p>No s’ha trobat cap usuari amb els criteris seleccionats.</p>
          </div>
        </template>

        <Column header="Nom i Cognoms" style="min-width: 14rem">
          <template #body="{ data }">
            <div class="user-cell" @click="navigateToDetail(data)">
              <div class="user-avatar-sm">
                {{ data.firstName.charAt(0) }}{{ data.lastName.charAt(0) }}
              </div>
              <div>
                <span class="user-full-name">{{ data.firstName }} {{ data.lastName }}</span>
                <div class="user-email-sub">{{ data.email }}</div>
              </div>
            </div>
          </template>
        </Column>

        <Column field="role" header="Rol" style="width: 8rem">
          <template #body="{ data }">
            <Tag :value="data.role" :severity="getRoleSeverity(data.role)" />
          </template>
        </Column>

        <Column field="isActive" header="Estat" style="width: 7rem">
          <template #body="{ data }">
            <Tag
              :value="data.isActive ? 'Actiu' : 'Inactiu'"
              :severity="data.isActive ? 'success' : 'warn'"
            />
          </template>
        </Column>

        <Column field="createdAt" header="Data d'alta" style="width: 9rem">
          <template #body="{ data }">
            {{ formatDate(data.createdAt) }}
          </template>
        </Column>

        <Column header="Accions" style="width: 11rem; text-align: right">
          <template #body="{ data }">
            <div class="actions-wrapper">
              <Button
                icon="pi pi-eye"
                text
                rounded
                severity="secondary"
                size="small"
                tooltip="Veure Detall"
                @click="navigateToDetail(data)"
                data-testid="btn-view-user"
              />
              <Button
                icon="pi pi-pencil"
                text
                rounded
                severity="info"
                size="small"
                tooltip="Editar Usuari"
                @click="openEditModal(data)"
                data-testid="btn-edit-user"
              />
              <Button
                icon="pi pi-key"
                text
                rounded
                severity="warn"
                size="small"
                tooltip="Restablir Contrasenya"
                @click="openResetPasswordModal(data)"
                data-testid="btn-reset-password"
              />
              <Button
                v-if="isAdmin"
                icon="pi pi-trash"
                text
                rounded
                severity="danger"
                size="small"
                tooltip="Donar de Baixa"
                @click="openDeleteConfirm(data)"
                data-testid="btn-delete-user"
              />
            </div>
          </template>
        </Column>
      </DataTable>
    </div>

    <!-- Modals -->
    <UserFormModal
      v-model:visible="showFormModal"
      :user="selectedUser"
      @saved="handleUserSaved"
    />

    <ResetPasswordModal
      v-model:visible="showResetPasswordModal"
      :user="selectedUser"
      @success="(msg) => (successFeedback = msg)"
    />

    <BatchImportModal
      v-model:visible="showBatchImportModal"
      @imported="handleBatchImported"
    />

    <!-- Diàleg de confirmació de Baixa -->
    <Dialog
      v-model:visible="showDeleteConfirmModal"
      modal
      header="Confirmar Baixa d'Usuari"
      :style="{ width: '90vw', maxWidth: '440px' }"
      data-testid="delete-confirm-dialog"
    >
      <div class="delete-confirm-content">
        <i class="pi pi-exclamation-triangle warning-icon" />
        <div>
          <p>
            Estàs segur que vols donar de baixa l’usuari
            <strong>{{ userToDelete?.firstName }} {{ userToDelete?.lastName }}</strong>
            ({{ userToDelete?.email }})?
          </p>
          <p class="soft-delete-hint">
            Aquesta acció aplicarà una baixa lògica (soft-delete), revocant l'accés immediatament sense esborrar el seu historial acadèmic.
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
          :loading="userStore.isSubmitting"
          @click="confirmDeleteUser"
          data-testid="btn-confirm-delete"
        />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.users-view-container {
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
}

.filter-select {
  min-width: 160px;
}

.table-card {
  background-color: #ffffff;
  border-radius: 0.75rem;
  border: 1px solid #e2e8f0;
  overflow: hidden;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
}

.user-cell:hover .user-full-name {
  color: #2563eb;
  text-decoration: underline;
}

.user-avatar-sm {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 50%;
  background: #e0e7ff;
  color: #3730a3;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 700;
  text-transform: uppercase;
}

.user-full-name {
  font-weight: 600;
  color: #1e293b;
}

.user-email-sub {
  font-size: 0.8rem;
  color: #64748b;
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
