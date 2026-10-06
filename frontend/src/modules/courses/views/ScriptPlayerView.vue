<script setup lang="ts">
import { ref, computed, onMounted, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCourseStore } from '../store'
import type { ScriptBlock } from '../types'

import Button from 'primevue/button'
import ProgressBar from 'primevue/progressbar'
import Tag from 'primevue/tag'
import Card from 'primevue/card'
import { useToast } from 'primevue/usetoast'

const route = useRoute()
const router = useRouter()
const courseStore = useCourseStore()
const toast = useToast()

const courseId = route.params.courseId as string
const unitId = route.params.unitId as string

const currentBlockIndex = ref(0)
const scriptBlocks = ref<ScriptBlock[]>([])

// Break Timer State
const breakSecondsLeft = ref(0)
const isBreakTimerRunning = ref(false)
let breakTimerInterval: ReturnType<typeof setInterval> | null = null

watch(() => courseStore.error, (err) => {
  if (err) {
    toast.add({ severity: 'error', summary: 'Error', detail: err, life: 4000 })
    courseStore.clearError()
  }
})

onMounted(async () => {
  if (courseId && unitId) {
    await courseStore.fetchCourseUnit(courseId, unitId)
    const blocks = await courseStore.fetchUnitScript(courseId, unitId)
    scriptBlocks.value = blocks || []
    if (scriptBlocks.value.length > 0) {
      initCurrentBlock(0)
    }
  }
})

onUnmounted(() => {
  stopBreakTimer()
})

const currentBlock = computed<ScriptBlock | null>(() => {
  if (scriptBlocks.value.length === 0) return null
  return scriptBlocks.value[currentBlockIndex.value] || null
})

const progressPercentage = computed(() => {
  if (scriptBlocks.value.length === 0) return 0
  return Math.round(((currentBlockIndex.value + 1) / scriptBlocks.value.length) * 100)
})

function initCurrentBlock(index: number) {
  stopBreakTimer()
  currentBlockIndex.value = index
  const block = scriptBlocks.value[index]
  if (block && block.blockType === 'break') {
    const mins = block.durationMinutes || 5
    breakSecondsLeft.value = mins * 60
  }
}

function nextBlock() {
  if (currentBlockIndex.value < scriptBlocks.value.length - 1) {
    initCurrentBlock(currentBlockIndex.value + 1)
  }
}

function previousBlock() {
  if (currentBlockIndex.value > 0) {
    initCurrentBlock(currentBlockIndex.value - 1)
  }
}

function startBreakTimer() {
  if (isBreakTimerRunning.value) return
  isBreakTimerRunning.value = true
  breakTimerInterval = setInterval(() => {
    if (breakSecondsLeft.value > 0) {
      breakSecondsLeft.value--
    } else {
      stopBreakTimer()
      toast.add({ severity: 'info', summary: 'Temps esgotat', detail: 'La pausa ha finalitzat!', life: 4000 })
    }
  }, 1000)
}

function stopBreakTimer() {
  isBreakTimerRunning.value = false
  if (breakTimerInterval) {
    clearInterval(breakTimerInterval)
    breakTimerInterval = null
  }
}

function formatBreakTime(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

function navigateToQuiz(quizId?: string | null) {
  if (quizId) {
    router.push(`/play`)
  } else {
    router.push(`/play`)
  }
}

function exitPlayer() {
  router.push(`/courses/${courseId}`)
}

function blockTypeSeverity(type?: string) {
  switch (type) {
    case 'material':
      return 'info'
    case 'quiz':
      return 'success'
    case 'break':
      return 'warn'
    default:
      return 'secondary'
  }
}

function blockTypeLabel(type?: string) {
  switch (type) {
    case 'material':
      return 'Material PDF'
    case 'quiz':
      return 'Qüestionari'
    case 'break':
      return 'Descans'
    default:
      return type || ''
  }
}
</script>

<template>
  <div class="script-player-view">
    <!-- Header Bar -->
    <header class="player-header">
      <div class="header-left">
        <Button
          icon="pi pi-times"
          severity="secondary"
          text
          rounded
          title="Sortir del visor"
          @click="exitPlayer"
        />
        <div class="unit-title-info">
          <span class="course-badge">{{ courseStore.currentCourse?.code || 'Curs' }}</span>
          <h1 class="player-unit-title">{{ courseStore.currentUnit?.title || 'Guió de Classe' }}</h1>
        </div>
      </div>

      <div class="header-progress">
        <span class="progress-text">
          Bloc {{ currentBlockIndex + 1 }} de {{ scriptBlocks.length }}
        </span>
        <ProgressBar :value="progressPercentage" :showValue="false" class="progress-bar" />
      </div>
    </header>

    <!-- Main Content Stage -->
    <main class="player-stage">
      <div v-if="courseStore.isLoading" class="loading-container">
        <i class="pi pi-spin pi-spinner spinner-icon"></i>
        <p>Carregant bloc...</p>
      </div>

      <div v-else-if="!currentBlock" class="empty-stage">
        <i class="pi pi-inbox empty-icon"></i>
        <h2>No hi ha cap bloc al guió de classe</h2>
        <p>La unitat actual no conté blocs de contingut configurats.</p>
        <Button label="Tornar al curs" icon="pi pi-arrow-left" severity="primary" @click="exitPlayer" />
      </div>

      <!-- BLOCK DISPLAY STAGE -->
      <div v-else class="block-card-container">
        <Card class="block-card">
          <template #title>
            <div class="block-header">
              <Tag
                :severity="blockTypeSeverity(currentBlock.blockType)"
                :value="blockTypeLabel(currentBlock.blockType)"
                class="block-tag"
              />
              <h2 class="current-block-title">{{ currentBlock.title }}</h2>
            </div>
          </template>

          <template #content>
            <p v-if="currentBlock.description" class="current-block-desc">
              {{ currentBlock.description }}
            </p>

            <!-- TYPE 1: MATERIAL PDF -->
            <div v-if="currentBlock.blockType === 'material'" class="material-stage">
              <div class="material-info-banner">
                <i class="pi pi-file-pdf pdf-banner-icon"></i>
                <div class="pdf-details">
                  <h3>Document de lectura / Apunts</h3>
                  <p v-if="currentBlock.startPage || currentBlock.endPage">
                    Llegir pàgines: <strong>{{ currentBlock.startPage || 1 }} - {{ currentBlock.endPage || 1 }}</strong>
                  </p>
                </div>
              </div>

              <div v-if="currentBlock.pdfUrl" class="pdf-viewer-placeholder">
                <i class="pi pi-file-pdf pdf-icon"></i>
                <p>Document PDF disponible:</p>
                <a
                  :href="currentBlock.pdfUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="pdf-link-btn"
                >
                  <i class="pi pi-external-link mr-1"></i> Obrir Document PDF
                </a>
              </div>
            </div>

            <!-- TYPE 2: QUIZ -->
            <div v-else-if="currentBlock.blockType === 'quiz'" class="quiz-stage">
              <div class="quiz-interactive-card">
                <i class="pi pi-check-square quiz-stage-icon"></i>
                <h3>Qüestionari de Classe</h3>
                <p v-if="currentBlock.quizTitle">
                  Qüestionari: <strong>{{ currentBlock.quizTitle }}</strong>
                </p>
                <p>Respon les preguntes per comprovar els teus coneixements d'aquest apartat.</p>
                <Button
                  label="Iniciar Qüestionari ara"
                  icon="pi pi-play"
                  severity="success"
                  size="large"
                  class="mt-3"
                  @click="navigateToQuiz(currentBlock.quizId)"
                />
              </div>
            </div>

            <!-- TYPE 3: BREAK -->
            <div v-else-if="currentBlock.blockType === 'break'" class="break-stage">
              <div class="break-timer-card">
                <i class="pi pi-coffee break-stage-icon"></i>
                <h3>Temps de Descans</h3>
                <div class="timer-display">{{ formatBreakTime(breakSecondsLeft) }}</div>

                <div class="timer-controls">
                  <Button
                    v-if="!isBreakTimerRunning"
                    label="Iniciar Descans"
                    icon="pi pi-play"
                    severity="warning"
                    @click="startBreakTimer"
                  />
                  <Button
                    v-else
                    label="Pausar Descans"
                    icon="pi pi-pause"
                    severity="secondary"
                    @click="stopBreakTimer"
                  />
                </div>
              </div>
            </div>
          </template>
        </Card>
      </div>
    </main>

    <!-- Footer Controls Bar -->
    <footer class="player-footer">
      <Button
        label="Anterior"
        icon="pi pi-chevron-left"
        severity="secondary"
        :disabled="currentBlockIndex === 0"
        @click="previousBlock"
      />

      <div class="block-dots">
        <span
          v-for="(_, idx) in scriptBlocks"
          :key="idx"
          class="dot"
          :class="{ active: idx === currentBlockIndex }"
          @click="initCurrentBlock(idx)"
        ></span>
      </div>

      <Button
        v-if="currentBlockIndex < scriptBlocks.length - 1"
        label="Següent"
        icon="pi pi-chevron-right"
        iconPos="right"
        severity="primary"
        @click="nextBlock"
      />
      <Button
        v-else
        label="Finalitzar Unitat"
        icon="pi pi-check"
        iconPos="right"
        severity="success"
        @click="exitPlayer"
      />
    </footer>
  </div>
</template>

<style scoped>
.script-player-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #f8fafc;
}

.player-header {
  height: 4rem;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.unit-title-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.course-badge {
  font-weight: 700;
  font-size: 0.8rem;
  color: #4f46e5;
  background: #e0e7ff;
  padding: 0.2rem 0.5rem;
  border-radius: 0.375rem;
}

.player-unit-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.header-progress {
  display: flex;
  align-items: center;
  gap: 1rem;

  .progress-text {
    font-size: 0.875rem;
    font-weight: 600;
    color: #64748b;
  }

  .progress-bar {
    width: 140px;
    height: 8px;
  }
}

.player-stage {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  overflow-y: auto;
}

.loading-container,
.empty-stage {
  text-align: center;
  color: #64748b;
}

.spinner-icon,
.empty-icon {
  font-size: 3rem;
  color: #6366f1;
  margin-bottom: 1rem;
}

.block-card-container {
  width: 100%;
  max-width: 800px;
}

.block-card {
  border-radius: 1rem;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  border: 1px solid #e2e8f0;
}

.block-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.current-block-title {
  font-size: 1.4rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.current-block-desc {
  font-size: 1rem;
  color: #475569;
  line-height: 1.6;
  margin-bottom: 1.5rem;
}

.material-stage {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.material-info-banner {
  display: flex;
  align-items: center;
  gap: 1rem;
  background: #f0f9ff;
  border: 1px solid #bae6fd;
  padding: 1rem 1.25rem;
  border-radius: 0.75rem;

  h3 {
    margin: 0 0 0.25rem 0;
    font-size: 1rem;
    color: #0369a1;
  }

  p {
    margin: 0;
    font-size: 0.9rem;
    color: #0284c7;
  }
}

.pdf-banner-icon {
  font-size: 2.25rem;
  color: #0284c7;
}

.pdf-viewer-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2.5rem;
  background: #f8fafc;
  border: 2px dashed #cbd5e1;
  border-radius: 0.75rem;

  .pdf-icon {
    font-size: 3rem;
    color: #ef4444;
    margin-bottom: 0.75rem;
  }

  p {
    margin: 0 0 1rem 0;
    color: #475569;
    font-weight: 600;
  }
}

.pdf-link-btn {
  display: inline-flex;
  align-items: center;
  background: #4f46e5;
  color: #ffffff;
  padding: 0.6rem 1.25rem;
  border-radius: 0.5rem;
  text-decoration: none;
  font-weight: 600;
  font-size: 0.9rem;

  &:hover {
    background: #4338ca;
  }
}

.quiz-stage {
  padding: 1rem 0;
}

.quiz-interactive-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 2.5rem;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 0.75rem;

  .quiz-stage-icon {
    font-size: 3.5rem;
    color: #16a34a;
    margin-bottom: 1rem;
  }

  h3 {
    font-size: 1.3rem;
    color: #15803d;
    margin: 0 0 0.5rem 0;
  }

  p {
    color: #166534;
    font-size: 0.95rem;
    margin: 0.25rem 0;
  }
}

.break-stage {
  padding: 1rem 0;
}

.break-timer-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 2.5rem;
  background: #fffbeb;
  border: 1px solid #fef3c7;
  border-radius: 0.75rem;

  .break-stage-icon {
    font-size: 3.5rem;
    color: #d97706;
    margin-bottom: 1rem;
  }

  h3 {
    font-size: 1.3rem;
    color: #b45309;
    margin: 0 0 1rem 0;
  }
}

.timer-display {
  font-size: 4rem;
  font-weight: 800;
  font-family: monospace;
  color: #92400e;
  margin-bottom: 1.5rem;
}

.player-footer {
  height: 4.5rem;
  background: #ffffff;
  border-top: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2rem;
}

.block-dots {
  display: flex;
  gap: 0.5rem;
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #cbd5e1;
  cursor: pointer;
  transition: all 0.2s ease;

  &.active {
    background: #4f46e5;
    transform: scale(1.3);
  }
}
</style>
