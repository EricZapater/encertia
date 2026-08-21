<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <div id="encertia-root">
    <RouterView />
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
}
</style>
