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
          to="/groups"
          class="nav-link"
          :class="{ active: isRouteActive('/groups') }"
        >
          <i class="pi pi-users"></i>
          <span>{{ $t('nav.groups') }}</span>
        </router-link>

        <router-link
          v-if="canAccessTeacherFeatures"
          to="/users"
          class="nav-link"
          :class="{ active: isRouteActive('/users') }"
        >
          <i class="pi pi-user-plus"></i>
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

        <router-link
          v-if="authStore.isAdmin"
          to="/admin/devices"
          class="nav-link"
          :class="{ active: isRouteActive('/admin/devices') }"
          data-testid="nav-link-devices"
        >
          <i class="pi pi-desktop"></i>
          <span>{{ $t('nav.devices') }}</span>
        </router-link>
      </nav>

      <!-- Right User Controls (Desplegable) -->
      <div class="user-menu-wrapper" ref="userMenuRef">
        <button
          type="button"
          class="user-menu-trigger"
          :class="{ active: isDropdownOpen }"
          @click.stop="toggleDropdown"
          :aria-expanded="isDropdownOpen"
          aria-haspopup="true"
          data-testid="user-menu-trigger"
          :title="$t('nav.profile')"
        >
          <div class="user-avatar-circle">
            <i class="pi pi-user"></i>
          </div>
          <span class="user-name">{{ authStore.fullName || authStore.currentUser?.email }}</span>
          <i class="pi pi-chevron-down chevron-icon" :class="{ 'chevron-rotated': isDropdownOpen }"></i>
        </button>

        <!-- Menú desplegable -->
        <Transition name="dropdown-fade">
          <div v-if="isDropdownOpen" class="user-dropdown-menu" role="menu">
            <!-- Capçalera del desplegable: Nom, Email i Rol -->
            <div class="dropdown-header">
              <div class="dropdown-user-info">
                <span class="dropdown-user-name">{{ authStore.fullName || 'Usuari' }}</span>
                <span class="dropdown-user-email">{{ authStore.currentUser?.email }}</span>
              </div>
              <Tag :severity="roleSeverity" class="role-tag">
                {{ roleLabel }}
              </Tag>
            </div>

            <div class="dropdown-divider"></div>

            <!-- Enllaç a El meu perfil -->
            <button
              type="button"
              class="dropdown-item"
              @click="goToProfile"
              data-testid="dropdown-profile-btn"
            >
              <i class="pi pi-user dropdown-item-icon"></i>
              <span>{{ $t('nav.profile') }}</span>
            </button>

            <div class="dropdown-divider"></div>

            <!-- Selector d'idioma -->
            <div class="dropdown-section">
              <div class="dropdown-section-title">
                <i class="pi pi-globe"></i>
                <span>Idioma</span>
              </div>
              <div class="lang-selector-group">
                <button
                  v-for="lang in supportedLangs"
                  :key="lang.code"
                  type="button"
                  class="lang-pill"
                  :class="{ active: locale === lang.code }"
                  @click="changeLang(lang.code)"
                >
                  {{ lang.label }}
                </button>
              </div>
            </div>

            <div class="dropdown-divider"></div>

            <!-- Botó de tancar sessió -->
            <button
              type="button"
              class="dropdown-item logout-item"
              @click="handleLogout"
              data-testid="dropdown-logout-btn"
            >
              <i class="pi pi-sign-out dropdown-item-icon"></i>
              <span>{{ $t('nav.logout') }}</span>
            </button>
          </div>
        </Transition>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/modules/auth/store'
import type { SupportedLanguage } from '@/i18n'
import Tag from 'primevue/tag'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { t, locale } = useI18n()

// Gestió de l'estat del menú desplegable d'usuari
const userMenuRef = ref<HTMLElement | null>(null)
const isDropdownOpen = ref(false)

function toggleDropdown() {
  isDropdownOpen.value = !isDropdownOpen.value
}

function closeDropdown() {
  isDropdownOpen.value = false
}

function handleClickOutside(event: MouseEvent) {
  if (userMenuRef.value && !userMenuRef.value.contains(event.target as Node)) {
    closeDropdown()
  }
}

function handleKeyDown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    closeDropdown()
  }
}

onMounted(() => {
  window.addEventListener('click', handleClickOutside)
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('click', handleClickOutside)
  window.removeEventListener('keydown', handleKeyDown)
})

function goToProfile() {
  closeDropdown()
  router.push('/profile')
}

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
  closeDropdown()
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

.user-menu-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.user-menu-trigger {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.35rem 0.65rem;
  background-color: transparent;
  border: 1px solid transparent;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
}

.user-menu-trigger:hover,
.user-menu-trigger.active {
  background-color: #f1f5f9;
  border-color: #e2e8f0;
}

.user-avatar-circle {
  width: 2rem;
  height: 2rem;
  border-radius: 50%;
  background-color: #e0e7ff;
  color: #4f46e5;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.95rem;
  flex-shrink: 0;
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

.chevron-icon {
  font-size: 0.75rem;
  color: #64748b;
  transition: transform 0.2s ease;
}

.chevron-rotated {
  transform: rotate(180deg);
}

/* Menú Desplegable */
.user-dropdown-menu {
  position: absolute;
  top: calc(100% + 0.5rem);
  right: 0;
  width: 250px;
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.05);
  z-index: 1100;
  padding: 0.5rem 0;
}

.dropdown-header {
  padding: 0.6rem 1rem 0.5rem;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
}

.dropdown-user-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dropdown-user-name {
  font-size: 0.9rem;
  font-weight: 700;
  color: #0f172a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dropdown-user-email {
  font-size: 0.75rem;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-top: 0.1rem;
}

.role-tag {
  font-size: 0.7rem !important;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  padding: 0.1rem 0.4rem !important;
  flex-shrink: 0;
}

.dropdown-divider {
  height: 1px;
  background-color: #f1f5f9;
  margin: 0.35rem 0;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  width: 100%;
  padding: 0.55rem 1rem;
  background: none;
  border: none;
  font-size: 0.875rem;
  font-weight: 500;
  color: #334155;
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
}

.dropdown-item:hover {
  background-color: #f8fafc;
  color: #0f172a;
}

.dropdown-item-icon {
  font-size: 0.95rem;
  color: #64748b;
}

.logout-item {
  color: #dc2626;
}

.logout-item .dropdown-item-icon {
  color: #dc2626;
}

.logout-item:hover {
  background-color: #fef2f2;
  color: #b91c1c;
}

.dropdown-section {
  padding: 0.4rem 1rem 0.5rem;
}

.dropdown-section-title {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 0.4rem;
}

.dropdown-section-title i {
  font-size: 0.8rem;
}

.lang-selector-group {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.lang-pill {
  flex: 1;
  background-color: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  padding: 0.25rem 0.4rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: #475569;
  cursor: pointer;
  text-align: center;
  transition: all 0.15s ease;
}

.lang-pill:hover {
  background-color: #e2e8f0;
  color: #0f172a;
}

.lang-pill.active {
  background-color: #6366f1;
  border-color: #6366f1;
  color: #ffffff;
}

/* Animació d'obertura/tancament */
.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: all 0.15s ease-out;
}

.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
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
