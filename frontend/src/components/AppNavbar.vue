<template>
  <header class="app-navbar-header">
    <div class="navbar-container">
      <!-- Brand Logo -->
      <div class="navbar-brand" @click="router.push('/quizzes')">
        <i class="pi pi-check-square brand-icon"></i>
        <span class="brand-name">Encertia</span>
      </div>

      <!-- Navigation Links -->
      <nav class="navbar-menu">
        <router-link
          to="/quizzes"
          class="nav-link"
          :class="{ active: isRouteActive('/quizzes') }"
        >
          <i class="pi pi-th-large"></i>
          <span>{{ $t('nav.quizzes') }}</span>
        </router-link>

        <router-link
          v-if="canAccessTeacherFeatures"
          to="/evaluations"
          class="nav-link"
          :class="{ active: isRouteActive('/evaluations') }"
        >
          <i class="pi pi-chart-bar"></i>
          <span>{{ $t('nav.evaluations') }}</span>
        </router-link>

        <router-link
          v-if="canAccessTeacherFeatures"
          to="/users"
          class="nav-link"
          :class="{ active: isRouteActive('/users') }"
        >
          <i class="pi pi-users"></i>
          <span>{{ $t('nav.users') }}</span>
        </router-link>

        <router-link
          to="/courses"
          class="nav-link"
          :class="{ active: isRouteActive('/courses') }"
        >
          <i class="pi pi-book"></i>
          <span>{{ $t('nav.courses') }}</span>
        </router-link>

        <router-link
          to="/materials"
          class="nav-link"
          :class="{ active: isRouteActive('/materials') }"
        >
          <i class="pi pi-folder-open"></i>
          <span>{{ $t('nav.materials') }}</span>
        </router-link>

        <router-link
          v-if="canAccessTeacherFeatures"
          to="/help/teacher-manual"
          class="nav-link"
          :class="{ active: isRouteActive('/help/teacher-manual') }"
        >
          <i class="pi pi-question-circle"></i>
          <span>{{ $t('nav.manual') }}</span>
        </router-link>

        <router-link
          v-if="authStore.isAdmin"
          to="/metrics"
          class="nav-link"
          :class="{ active: isRouteActive('/metrics') }"
          data-testid="nav-link-metrics"
        >
          <i class="pi pi-chart-line"></i>
          <span>{{ $t('nav.metrics') }}</span>
        </router-link>
      </nav>

      <!-- Right User Controls -->
      <div class="navbar-user">
        <!-- Language Selector -->
        <div class="lang-selector">
          <span class="lang-globe" title="Idioma">🌐</span>
          <button
            v-for="lang in supportedLangs"
            :key="lang.code"
            type="button"
            class="lang-btn"
            :class="{ active: locale === lang.code }"
            @click="changeLang(lang.code)"
          >
            {{ lang.label }}
          </button>
        </div>

        <Tag :severity="roleSeverity" class="role-tag">
          {{ roleLabel }}
        </Tag>

        <div class="user-profile-link" @click="router.push('/profile')" :title="$t('nav.profile')">
          <i class="pi pi-user user-avatar-icon"></i>
          <span class="user-name">{{ authStore.fullName || authStore.currentUser?.email }}</span>
        </div>

        <Button
          icon="pi pi-sign-out"
          severity="secondary"
          text
          rounded
          class="logout-btn"
          :title="$t('nav.logout')"
          @click="handleLogout"
        />
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/modules/auth/store'
import type { SupportedLanguage } from '@/i18n'
import Button from 'primevue/button'
import Tag from 'primevue/tag'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { t, locale } = useI18n()

const supportedLangs: { code: SupportedLanguage; label: string }[] = [
  { code: 'ca', label: 'CA' },
  { code: 'es', label: 'ES' },
  { code: 'en', label: 'EN' }
]

function changeLang(langCode: SupportedLanguage) {
  authStore.updateLanguage(langCode)
}

const canAccessTeacherFeatures = computed(() => authStore.isAdmin || authStore.isTeacher)

const roleLabel = computed(() => {
  if (authStore.isAdmin) return t('nav.roles.admin')
  if (authStore.isTeacher) return t('nav.roles.teacher')
  if (authStore.isStudent) return t('nav.roles.student')
  return t('nav.roles.user')
})

const roleSeverity = computed(() => {
  if (authStore.isAdmin) return 'danger'
  if (authStore.isTeacher) return 'info'
  return 'success'
})

function isRouteActive(pathPrefix: string): boolean {
  return route.path.startsWith(pathPrefix)
}

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.app-navbar-header {
  background-color: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  position: sticky;
  top: 0;
  z-index: 1000;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
}

.navbar-container {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 1.25rem;
  height: 4rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
}

.navbar-brand {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  cursor: pointer;
  user-select: none;
  transition: opacity 0.15s ease;
}

.navbar-brand:hover {
  opacity: 0.85;
}

.brand-icon {
  font-size: 1.6rem;
  color: #6366f1;
}

.brand-name {
  font-size: 1.35rem;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.02em;
}

.navbar-menu {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  height: 100%;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.85rem;
  border-radius: 0.5rem;
  color: #475569;
  text-decoration: none;
  font-weight: 500;
  font-size: 0.95rem;
  transition: all 0.15s ease;
}

.nav-link i {
  font-size: 1.1rem;
}

.nav-link:hover:not(.disabled) {
  background-color: #f1f5f9;
  color: #0f172a;
}

.nav-link.active {
  background-color: #e0e7ff;
  color: #4338ca;
  font-weight: 600;
}

.nav-link.disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.nav-tag {
  font-size: 0.65rem !important;
  padding: 0.1rem 0.35rem !important;
}

.navbar-user {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.lang-selector {
  display: flex;
  align-items: center;
  gap: 0.2rem;
  background-color: #f1f5f9;
  padding: 0.2rem 0.4rem;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
}

.lang-globe {
  font-size: 0.85rem;
  margin-right: 0.15rem;
}

.lang-btn {
  background: none;
  border: none;
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
  padding: 0.15rem 0.35rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.lang-btn:hover {
  color: #0f172a;
}

.lang-btn.active {
  background-color: #6366f1;
  color: #ffffff;
}

.role-tag {
  font-size: 0.75rem !important;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.user-profile-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  padding: 0.4rem 0.65rem;
  border-radius: 0.5rem;
  transition: background-color 0.15s ease;
}

.user-profile-link:hover {
  background-color: #f1f5f9;
}

.user-avatar-icon {
  font-size: 1.1rem;
  color: #64748b;
  background-color: #f1f5f9;
  padding: 0.4rem;
  border-radius: 50%;
}

.user-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: #1e293b;
  max-width: 160px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.logout-btn {
  color: #64748b !important;
}

.logout-btn:hover {
  color: #ef4444 !important;
  background-color: #fef2f2 !important;
}

@media (max-width: 768px) {
  .navbar-container {
    padding: 0 0.75rem;
  }
  .brand-name,
  .user-name,
  .nav-link span {
    display: none;
  }
  .nav-link.disabled {
    display: none;
  }
}
</style>
