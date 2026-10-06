<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/modules/auth/store'
import { useUserStore } from '../store'
import type { User } from '../types'

import Card from 'primevue/card'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import ProgressSpinner from 'primevue/progressspinner'
import { useToast } from 'primevue/usetoast'

import UserFormModal from './UserFormModal.vue'
import ResetPasswordModal from './ResetPasswordModal.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const userStore = useUserStore()
const toast = useToast()

const userId = computed(() => route.params.id as string)
const user = ref<User | null>(null)
const isLoading = ref(true)
const errorMessage = ref<string | null>(null)
const successMessage = ref<string | null>(null)

watch(successMessage, (msg) => {
  if (msg) {
    toast.add({ severity: 'success', summary: 'Èxit', detail: msg, life: 3000 })
  }
})

watch(errorMessage, (msg) => {
  if (msg) {
    toast.add({ severity: 'error', summary: 'Error', detail: msg, life: 4000 })
  }
})

const showEditModal = ref(false)
const showResetPasswordModal = ref(false)

const isAdmin = computed(() => authStore.currentUser?.role === 'admin')
const isSelf = computed(() => authStore.currentUser?.id === userId.value)
const canEdit = computed(() => isAdmin.value || isSelf.value)

async function loadUser() {
  if (!userId.value) return
  isLoading.value = true
  errorMessage.value = null
  try {
    user.value = await userStore.fetchUserById(userId.value)
  } catch (err: any) {
    errorMessage.value =
      err.response?.data?.error?.message ||
      err.message ||
      'No s’ha pogut carregar el detall de l’usuari.'
  } finally {
    isLoading.value = false
  }
}

function handleUserSaved(updatedUser: User) {
  user.value = updatedUser
  successMessage.value = 'Dades d’usuari actualitzades correctament.'
}

function handlePasswordReset(msg: string) {
  successMessage.value = msg
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleString('ca-ES', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch {
    return dateStr
  }
}

onMounted(() => {
  loadUser()
})
</script>

<template>
  <div class="user-detail-container">
    <div class="page-header">
      <Button
        icon="pi pi-arrow-left"
        label="Tornar al llistat"
        text
        severity="secondary"
        @click="router.push('/users')"
        data-testid="btn-back-users"
      />
    </div>

    <div v-if="isLoading" class="loading-wrapper">
      <ProgressSpinner strokeWidth="4" />
    </div>

    <div v-else-if="user" class="detail-card-wrapper">
      <Card class="user-card">
        <template #title>
          <div class="user-header">
            <div class="avatar-circle">
              {{ user.firstName.charAt(0) }}{{ user.lastName.charAt(0) }}
            </div>
            <div class="user-meta">
              <h2 class="user-name">{{ user.firstName }} {{ user.lastName }}</h2>
              <p class="user-email">{{ user.email }}</p>
              <div class="badges-row">
                <Tag
                  :value="user.role"
                  :severity="user.role === 'admin' ? 'danger' : user.role === 'teacher' ? 'info' : 'secondary'"
                />
                <Tag
                  :value="user.isActive ? 'Actiu' : 'Inactiu'"
                  :severity="user.isActive ? 'success' : 'warn'"
                />
              </div>
            </div>
          </div>
        </template>

        <template #content>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">Identificador (UUID)</span>
              <span class="info-value font-mono">{{ user.id }}</span>
            </div>

            <div class="info-item">
              <span class="info-label">Data de Registre</span>
              <span class="info-value">{{ formatDate(user.createdAt) }}</span>
            </div>

            <div class="info-item">
              <span class="info-label">Darrera Actualització</span>
              <span class="info-value">{{ formatDate(user.updatedAt) }}</span>
            </div>
          </div>
        </template>

        <template #footer>
          <div class="card-footer-actions">
            <Button
              v-if="canEdit"
              label="Editar Dades"
              icon="pi pi-pencil"
              severity="primary"
              @click="showEditModal = true"
              data-testid="btn-edit-user-detail"
            />
            <Button
              v-if="isAdmin"
              label="Restablir Contrasenya"
              icon="pi pi-key"
              severity="warning"
              outlined
              @click="showResetPasswordModal = true"
              data-testid="btn-reset-pwd-detail"
            />
          </div>
        </template>
      </Card>
    </div>

    <!-- Modals -->
    <UserFormModal
      v-model:visible="showEditModal"
      :user="user"
      @saved="handleUserSaved"
    />

    <ResetPasswordModal
      v-model:visible="showResetPasswordModal"
      :user="user"
      @success="handlePasswordReset"
    />
  </div>
</template>

<style scoped>
.user-detail-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.page-header {
  margin-bottom: 1.5rem;
}

.loading-wrapper {
  display: flex;
  justify-content: center;
  padding: 4rem;
}

.user-card {
  border-radius: 1rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1);
}

.user-header {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #f1f5f9;
}

.avatar-circle {
  width: 4rem;
  height: 4rem;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6 0%, #1e40af 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 700;
  text-transform: uppercase;
}

.user-name {
  margin: 0;
  font-size: 1.5rem;
  color: #0f172a;
}

.user-email {
  margin: 0.2rem 0 0.5rem 0;
  font-size: 0.95rem;
  color: #64748b;
}

.badges-row {
  display: flex;
  gap: 0.5rem;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
  padding: 1rem 0;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-label {
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #94a3b8;
}

.info-value {
  font-size: 0.95rem;
  color: #334155;
}

.font-mono {
  font-family: monospace;
  font-size: 0.85rem;
}

.card-footer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding-top: 1rem;
  border-top: 1px solid #f1f5f9;
}

.mb-3 {
  margin-bottom: 1rem;
}
</style>
