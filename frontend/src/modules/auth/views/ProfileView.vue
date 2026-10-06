<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../store'
import type { SupportedLanguage } from '@/i18n'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Divider from 'primevue/divider'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const authStore = useAuthStore()
const toast = useToast()
const { t, locale } = useI18n()

const user = computed(() => authStore.currentUser)
const isLoggingOut = ref(false)
const isRefreshingProfile = ref(false)

const supportedLangs: { code: SupportedLanguage; label: string }[] = [
  { code: 'ca', label: 'Català (CA)' },
  { code: 'es', label: 'Castellà (ES)' },
  { code: 'en', label: 'Anglès (EN)' }
]

function changeLang(langCode: SupportedLanguage) {
  authStore.updateLanguage(langCode)
}

function formatDate(dateString?: string) {
  if (!dateString) return 'N/D'
  try {
    const date = new Date(dateString)
    return new Intl.DateTimeFormat(locale.value === 'en' ? 'en-US' : locale.value === 'es' ? 'es-ES' : 'ca-ES', {
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(date)
  } catch {
    return dateString
  }
}

async function handleRefresh() {
  isRefreshingProfile.value = true
  try {
    await authStore.fetchMe()
    toast.add({
      severity: 'success',
      summary: t('common.success'),
      detail: t('auth.profile.updatedSuccess'),
      life: 3000
    })
  } catch {
    toast.add({
      severity: 'error',
      summary: t('common.error'),
      detail: t('common.error'),
      life: 4000
    })
  } finally {
    isRefreshingProfile.value = false
  }
}

async function handleLogout() {
  isLoggingOut.value = true
  try {
    await authStore.logout()
    router.push('/login')
  } catch {
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
    <main class="profile-main">
      <div class="profile-wrapper">
        <Card class="profile-card">
          <template #title>
            <div class="profile-header">
              <div class="avatar-circle">
                {{ user?.firstName?.charAt(0) || 'U' }}{{ user?.lastName?.charAt(0) || '' }}
              </div>
              <div class="profile-info-header">
                <div class="profile-name-row">
                  <h2 class="profile-fullname">{{ authStore.fullName || $t('nav.roles.user') }}</h2>
                  <Tag
                    :value="user?.role === 'teacher' ? $t('nav.roles.teacher') : user?.role === 'admin' ? $t('nav.roles.admin') : $t('nav.roles.student')"
                    :severity="user?.role === 'admin' ? 'danger' : user?.role === 'teacher' ? 'info' : 'success'"
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
                <span class="detail-label">{{ $t('auth.profile.id') }}</span>
                <span class="detail-value mono">{{ user?.id }}</span>
              </div>

              <div class="detail-item">
                <span class="detail-label">{{ $t('auth.profile.email') }}</span>
                <span class="detail-value">{{ user?.email }}</span>
              </div>

              <div class="detail-item">
                <span class="detail-label">{{ $t('auth.profile.fullName') }}</span>
                <span class="detail-value">{{ user?.firstName }} {{ user?.lastName }}</span>
              </div>

              <div class="detail-item">
                <span class="detail-label">{{ $t('auth.profile.registeredAt') }}</span>
                <span class="detail-value">{{ formatDate(user?.createdAt) }}</span>
              </div>

              <div class="detail-item full-width-item">
                <span class="detail-label">🌐 {{ $t('auth.profile.language') }}</span>
                <div class="lang-selector-profile">
                  <button
                    v-for="lang in supportedLangs"
                    :key="lang.code"
                    type="button"
                    class="lang-profile-btn"
                    :class="{ active: locale === lang.code }"
                    @click="changeLang(lang.code)"
                  >
                    {{ lang.label }}
                  </button>
                </div>
              </div>
            </div>

            <Divider />

            <!-- Secció segons rol -->
            <div v-if="authStore.isAdmin" class="role-section">
              <h3 class="section-title">{{ $t('auth.profile.adminPanel') }}</h3>
              <p class="section-desc">
                {{ $t('auth.profile.adminDesc') }}
              </p>
              <div class="action-cards">
                <div class="action-card cursor-pointer" @click="router.push('/quizzes')">
                  <i class="pi pi-bolt action-icon"></i>
                  <h4>{{ $t('auth.profile.quizManagement') }}</h4>
                  <p>{{ $t('auth.profile.quizManagementDesc') }}</p>
                </div>
                <div class="action-card cursor-pointer" @click="router.push('/users')">
                  <i class="pi pi-users action-icon"></i>
                  <h4>{{ $t('auth.profile.userManagement') }}</h4>
                  <p>{{ $t('auth.profile.userManagementDesc') }}</p>
                </div>
              </div>
            </div>

            <div v-else-if="authStore.isTeacher" class="role-section">
              <h3 class="section-title">{{ $t('auth.profile.teacherPanel') }}</h3>
              <p class="section-desc">
                {{ $t('auth.profile.teacherDesc') }}
              </p>
              <div class="action-cards">
                <div class="action-card cursor-pointer" @click="router.push('/quizzes')">
                  <i class="pi pi-bolt action-icon"></i>
                  <h4>{{ $t('auth.profile.myGames') }}</h4>
                  <p>{{ $t('auth.profile.myGamesDesc') }}</p>
                </div>
                <div class="action-card cursor-pointer" @click="router.push('/users')">
                  <i class="pi pi-users action-icon"></i>
                  <h4>{{ $t('auth.profile.studentManagement') }}</h4>
                  <p>{{ $t('auth.profile.studentManagementDesc') }}</p>
                </div>
              </div>
            </div>

            <div v-else-if="authStore.isStudent" class="role-section">
              <h3 class="section-title">{{ $t('auth.profile.studentPanel') }}</h3>
              <p class="section-desc">
                {{ $t('auth.profile.studentDesc') }}
              </p>
              <div class="action-cards">
                <div class="action-card cursor-pointer" @click="router.push('/quizzes')">
                  <i class="pi pi-bolt action-icon"></i>
                  <h4>{{ $t('auth.profile.myGames') }}</h4>
                  <p>{{ $t('auth.profile.myGamesDesc') }}</p>
                </div>
              </div>
            </div>
          </template>

          <template #footer>
            <div class="profile-actions">
              <Button
                :label="$t('auth.profile.updateData')"
                icon="pi pi-refresh"
                severity="secondary"
                size="small"
                :loading="isRefreshingProfile"
                @click="handleRefresh"
              />
              <Button
                :label="$t('auth.profile.logout')"
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

.full-width-item {
  grid-column: 1 / -1;
}

.lang-selector-profile {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.25rem;
}

.lang-profile-btn {
  background-color: #f1f5f9;
  border: 1px solid #cbd5e1;
  color: #475569;
  padding: 0.4rem 0.85rem;
  border-radius: 0.5rem;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.lang-profile-btn:hover {
  background-color: #e2e8f0;
  color: #0f172a;
}

.lang-profile-btn.active {
  background-color: #4338ca;
  border-color: #4338ca;
  color: #ffffff;
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
