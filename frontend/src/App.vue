<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import { useAuthStore } from '@/modules/auth/store'
import AppNavbar from '@/components/AppNavbar.vue'

const route = useRoute()
const authStore = useAuthStore()

const showNavbar = computed(() => {
  if (!authStore.isAuthenticated) return false
  if (route.meta.hideNavbar) return false
  if (route.path.startsWith('/play') || route.path.startsWith('/matches/')) return false
  return true
})
</script>

<template>
  <div id="encertia-root">
    <AppNavbar v-if="showNavbar" />
    <main class="main-content">
      <RouterView />
    </main>
  </div>
</template>

<style>
/* Reset i estils base */
*,
*::before,
*::after {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family:
    system-ui,
    -apple-system,
    BlinkMacSystemFont,
    'Segoe UI',
    Roboto,
    Oxygen,
    Ubuntu,
    Cantarell,
    'Open Sans',
    'Helvetica Neue',
    sans-serif;
  color: #1e293b;
  background-color: #f8fafc;
  line-height: 1.5;
  -webkit-font-smoothing: antialiased;
}

#encertia-root {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.main-content {
  flex: 1;
}
</style>
