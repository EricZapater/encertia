<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Divider from 'primevue/divider'
import Message from 'primevue/message'

const router = useRouter()
const authStore = useAuthStore()

const user = computed(() => authStore.currentUser)
const isLoggingOut = ref(false)
const isRefreshingProfile = ref(false)
const successMessage = ref<string | null>(null)
const errorMessage = ref<string | null>(null)

function formatDate(dateString?: string) {
  if (!dateString) return 'N/D'
  try {
    const date = new Date(dateString)
    return new Intl.DateTimeFormat('ca-ES', {
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(date)
  } catch {
    return dateString
  }
}

async function handleRefresh() {
  isRefreshingProfile.value = true
  errorMessage.value = null
  successMessage.value = null
  try {
    await authStore.fetchMe()
    successMessage.value = 'Dades del perfil actualitzades correctament.'
    setTimeout(() => {
      successMessage.value = null
    }, 3000)
  } catch (err: any) {
    errorMessage.value = 'No s’han pogut actualitzar les dades del perfil.'
  } finally {
    isRefreshingProfile.value = false
  }
}

async function handleLogout() {
  isLoggingOut.value = true
  try {
    await authStore.logout()
    router.push('/login')
  } catch (err) {
    router.push('/login')
  } finally {
    isLoggingOut.value = false
  }
}

onMounted(() => {
  if (!authStore.currentUser) {
    authStore.fetchMe()
  }
})
</script>

<template>
  <div class="profile-container">
    <header class="app-header">
      <div class="header-content">
        <div class="brand">
          <span class="brand-name">Encertia</span>
          <span class="brand-badge">Educació</span>
        </div>
        <div class="header-actions">
          <Button
            v-if="authStore.isAdmin || authStore.isTeacher"
            label="Gestió d'Usuaris"
            icon="pi pi-users"
            severity="primary"
            size="small"
            @click="router.push('/users')"
            data-testid="btn-nav-users"
          />
          <Button
            label="Tancar Sessió"
            icon="pi pi-sign-out"
            severity="secondary"
            outlined
            size="small"
            :loading="isLoggingOut"
            @click="handleLogout"
          />
        </div>
      </div>
    </header>

    <main class="profile-main">
      <div class="profile-wrapper">
        <Message v-if="successMessage" severity="success" class="mb-4">
          {{ successMessage }}
        </Message>
        <Message v-if="errorMessage" severity="error" class="mb-4">
          {{ errorMessage }}
        </Message>

        <Card class="profile-card">
          <template #title>
            <div class="profile-header">
              <div class="avatar-circle">
                {{ user?.firstName?.charAt(0) || 'U' }}{{ user?.lastName?.charAt(0) || '' }}
              </div>
              <div class="profile-info-header">
                <div class="profile-name-row">
                  <h2 class="profile-fullname">{{ authStore.fullName || 'Usuari' }}</h2>
                  <Tag
                    :value="user?.role === 'teacher' ? 'Professor / Docent' : 'Alumne / Estudiant'"
                    :severity="user?.role === 'teacher' ? 'info' : 'success'"
                    class="role-tag"
                  />
                </div>
                <p class="profile-email">{{ user?.email }}</p>
              </div>
            </div>
          </template>

          <template #content>
            <Divider />

            <div class="details-grid">
              <div class="detail-item">
                <span class="detail-label">Identificador (UUID)</span>
                <span class="detail-value mono">{{ user?.id }}</span>
              </div>

              <div class="detail-item">
                <span class="detail-label">Correu Electrònic</span>
                <span class="detail-value">{{ user?.email }}</span>
              </div>

              <div class="detail-item">
                <span class="detail-label">Nom Complet</span>
                <span class="detail-value">{{ user?.firstName }} {{ user?.lastName }}</span>
              </div>

              <div class="detail-item">
                <span class="detail-label">Data de Registre</span>
                <span class="detail-value">{{ formatDate(user?.createdAt) }}</span>
              </div>
            </div>

            <Divider />

            <!-- Secció segons rol -->
            <div v-if="authStore.isAdmin" class="role-section">
              <h3 class="section-title">Panell d'Administrador</h3>
              <p class="section-desc">
                Gestió global de la plataforma, usuaris, rols i configuració del sistema.
              </p>
              <div class="action-cards">
                <div class="action-card cursor-pointer" @click="router.push('/users')">
                  <i class="pi pi-users action-icon"></i>
                  <h4>Gestió d'Usuaris</h4>
                  <p>Administra comptes d'administradors, professors i estudiants.</p>
                </div>
              </div>
            </div>

            <div v-else-if="authStore.isTeacher" class="role-section">
              <h3 class="section-title">Panell del Professor</h3>
              <p class="section-desc">
                Des d'aquí pots gestionar els teus cursos, matèries, continguts i alumnes.
              </p>
              <div class="action-cards">
                <div class="action-card cursor-pointer" @click="router.push('/users')">
                  <i class="pi pi-users action-icon"></i>
                  <h4>Gestió d'Alumnes</h4>
                  <p>Consulta la llista d'alumnes i fes importacions massives en CSV.</p>
                </div>
                <div class="action-card">
                  <i class="pi pi-book action-icon"></i>
                  <h4>Cursos i Classes</h4>
                  <p>Gestiona les assignatures i grups.</p>
                </div>
                <div class="action-card">
                  <i class="pi pi-check-square action-icon"></i>
                  <h4>Qüestionaris</h4>
                  <p>Crea, edita i avalua proves per als teus estudiants.</p>
                </div>
              </div>
            </div>

            <div v-else-if="authStore.isStudent" class="role-section">
              <h3 class="section-title">Panell de l'Estudiant</h3>
              <p class="section-desc">
                Accedeix als teus cursos matriculats i respon als qüestionaris actius.
              </p>
              <div class="action-cards">
                <div class="action-card">
                  <i class="pi pi-book action-icon"></i>
                  <h4>Els meus Cursos</h4>
                  <p>Revisa els materials d'estudi i activitats pendents.</p>
                </div>
                <div class="action-card">
                  <i class="pi pi-chart-line action-icon"></i>
                  <h4>Avaluacions</h4>
                  <p>Fes els qüestionaris assignats i consulta els teus resultats.</p>
                </div>
              </div>
            </div>
          </template>

          <template #footer>
            <div class="profile-actions">
              <Button
                label="Actualitzar Dades"
                icon="pi pi-refresh"
                severity="secondary"
                size="small"
                :loading="isRefreshingProfile"
                @click="handleRefresh"
              />
              <Button
                label="Tancar Sessió"
                icon="pi pi-sign-out"
                severity="danger"
                outlined
                size="small"
                :loading="isLoggingOut"
                @click="handleLogout"
              />
            </div>
          </template>
        </Card>
      </div>
    </main>
  </div>
</template>

<style scoped>
.profile-container {
  min-height: 100vh;
  background-color: #f8fafc;
  display: flex;
  flex-direction: column;
}

.app-header {
  background-color: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  padding: 0.75rem 1.5rem;
}

.header-content {
  max-width: 960px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.brand-name {
  font-size: 1.25rem;
  font-weight: 800;
  color: #1e3a8a;
  letter-spacing: -0.02em;
}

.brand-badge {
  font-size: 0.75rem;
  font-weight: 600;
  background-color: #e0e7ff;
  color: #3730a3;
  padding: 0.15rem 0.5rem;
  border-radius: 9999px;
}

.profile-main {
  flex: 1;
  padding: 2rem 1.5rem;
  display: flex;
  justify-content: center;
}

.profile-wrapper {
  width: 100%;
  max-width: 800px;
}

.mb-4 {
  margin-bottom: 1rem;
}

.profile-card {
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  border-radius: 1rem;
  border: 1px solid #e2e8f0;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

.avatar-circle {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: #ffffff;
  font-size: 1.5rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  text-transform: uppercase;
}

.profile-info-header {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.profile-name-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.profile-fullname {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.role-tag {
  font-size: 0.75rem;
  font-weight: 600;
}

.profile-email {
  font-size: 0.925rem;
  color: #64748b;
  margin: 0;
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 1.25rem;
  margin: 1rem 0;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #94a3b8;
}

.detail-value {
  font-size: 0.95rem;
  font-weight: 500;
  color: #1e293b;
  word-break: break-all;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.85rem;
  color: #475569;
}

.role-section {
  margin-top: 1rem;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.25rem 0;
}

.section-desc {
  font-size: 0.875rem;
  color: #64748b;
  margin: 0 0 1rem 0;
}

.action-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.action-card {
  padding: 1rem;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.action-icon {
  font-size: 1.5rem;
  color: #2563eb;
  margin-bottom: 0.25rem;
}

.action-card h4 {
  font-size: 0.95rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.action-card p {
  font-size: 0.825rem;
  color: #64748b;
  margin: 0;
}

.profile-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding-top: 0.5rem;
}
</style>
