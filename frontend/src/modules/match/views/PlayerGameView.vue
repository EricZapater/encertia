<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/modules/auth/store'
import { useMatchStore } from '../store'
import { KAHOOT_THEMES, type MatchAnswerOption } from '../types'

import Button from 'primevue/button'
import Tag from 'primevue/tag'
import ProgressBar from 'primevue/progressbar'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const matchStore = useMatchStore()

const pin = computed(() => (route.params.pin as string) || matchStore.pin || '')

onMounted(async () => {
  if (!authStore.isInitialized) {
    await authStore.initAuth()
  }

  if (!authStore.isAuthenticated) {
    router.push({
      name: 'login',
      query: { redirect: `/play/${pin.value}` }
    })
    return
  }

  // Si no està connectat o el PIN no coincideix, reconnecta com a jugador
  if (!matchStore.isConnected || matchStore.pin !== pin.value) {
    if (pin.value) {
      try {
        await matchStore.connectAsPlayer(pin.value)
      } catch {
        // Ignora errors de connect
      }
    } else {
      router.push('/play')
    }
  }
})

onUnmounted(() => {
  // Opcionalment desconnectar si surt de la vista
})

function getOptionTheme(index: number) {
  return KAHOOT_THEMES[index % KAHOOT_THEMES.length]
}

function isOptionSelected(optionId: string): boolean {
  return matchStore.mySelectedAnswerIds.includes(optionId)
}

function handleOptionClick(option: MatchAnswerOption) {
  if (matchStore.hasSubmittedAnswer || matchStore.status !== 'question_active') return

  const isMultiple = matchStore.currentQuestion?.type === 'multiple_choice'
  matchStore.selectAnswer(option.id, isMultiple)

  // En modalitat de resposta única, enviem automàticament per agilitat o permetem confirmar
  if (!isMultiple) {
    matchStore.submitAnswer()
  }
}

function handleMultipleSubmit() {
  matchStore.submitAnswer()
}

function handleExit() {
  matchStore.leaveMatch()
  router.push('/play')
}
</script>

<template>
  <div class="player-game-container" :class="`state-${matchStore.status || 'loading'}`">
    <!-- Header de Jugador -->
    <header class="player-header">
      <div class="header-left">
        <span class="player-nick">
          <i class="pi pi-user" /> {{ matchStore.myNickname || authStore.fullName || 'Jugador' }}
        </span>
      </div>
      <div class="header-center">
        <Tag :value="`PIN: ${pin}`" severity="contrast" class="pin-tag" />
      </div>
      <div class="header-right">
        <div class="score-badge">
          <i class="pi pi-star-fill text-amber-400" />
          <span>{{ matchStore.myScore }} pts</span>
        </div>
      </div>
    </header>

    <!-- PANTALLA: EXPULSAT -->
    <div v-if="matchStore.isKicked" class="game-card screen-kicked" data-testid="screen-kicked">
      <div class="feedback-icon kick-icon">
        <i class="pi pi-ban" />
      </div>
      <h2 class="feedback-title">Has estat expulsat</h2>
      <p class="feedback-subtitle">El moderador ha retirat el teu accés a aquesta partida.</p>
      <Button label="Tornar a l'inici" severity="secondary" @click="handleExit" class="mt-4" />
    </div>

    <!-- PANTALLA 1: SALA D'ESPERA (LOBBY) -->
    <div
      v-else-if="!matchStore.status || matchStore.status === 'lobby'"
      class="game-card screen-lobby"
      data-testid="screen-player-lobby"
    >
      <div class="pulse-avatar">
        <i class="pi pi-check" />
      </div>
      <h2 class="lobby-title">Estàs a dins!</h2>
      <p class="lobby-nick">{{ matchStore.myNickname || authStore.fullName }}</p>
      <p class="lobby-instruction">
        Posa't còmode. La partida començarà tan bon punt el moderador doni el tret de sortida.
      </p>

      <div class="lobby-status-pill">
        <i class="pi pi-spin pi-spinner" />
        <span>Esperant que comenci el joc...</span>
      </div>
    </div>

    <!-- PANTALLA 2: PAUSA PRÈVIA (QUESTION PREVIEW) -->
    <div
      v-else-if="matchStore.status === 'question_preview'"
      class="game-card screen-preview"
      data-testid="screen-player-preview"
    >
      <div class="question-meta-badge">
        Pregunta {{ matchStore.questionNumber }} de {{ matchStore.totalQuestions || '?' }}
      </div>

      <h2 class="preview-statement">
        {{ matchStore.currentQuestion?.title || "Atenció a l'enunciat..." }}
      </h2>

      <div v-if="matchStore.currentQuestion?.imageUrl" class="preview-image-wrapper">
        <img
          :src="matchStore.currentQuestion.imageUrl"
          alt="Imatge de la pregunta"
          class="preview-img"
        />
      </div>

      <div class="preview-waiting-box">
        <i class="pi pi-eye preview-icon" />
        <p>Llegeix amb atenció l'enunciat. Les opcions de resposta s'obriran en breus instants!</p>
      </div>
    </div>

    <!-- PANTALLA 3: PREGUNTA ACTIVA (QUESTION ACTIVE) -->
    <div
      v-else-if="matchStore.status === 'question_active'"
      class="screen-active"
      data-testid="screen-player-active"
    >
      <!-- Barra superior de temps -->
      <div class="timer-bar-container">
        <div class="timer-info">
          <span class="timer-count">{{ matchStore.timerSeconds }}s</span>
          <span class="question-pill">
            {{ matchStore.questionNumber }} / {{ matchStore.totalQuestions }}
          </span>
        </div>
        <ProgressBar
          :value="matchStore.timerPercentage"
          :showValue="false"
          class="timer-progress"
        />
      </div>

      <!-- Resposta ja enviada -->
      <div v-if="matchStore.hasSubmittedAnswer" class="submitted-feedback-card" data-testid="msg-answer-submitted">
        <div class="submitted-icon">
          <i class="pi pi-check-circle" />
        </div>
        <h3>Resposta registrada!</h3>
        <p>Esperant que s'acabi el temps o responguin la resta de jugadors...</p>
      </div>

      <!-- Formulari de respostes interactiu -->
      <div v-else class="answers-container">
        <div class="options-grid">
          <button
            v-for="(option, idx) in matchStore.currentQuestion?.options || []"
            :key="option.id"
            type="button"
            class="kahoot-btn"
            :class="{ selected: isOptionSelected(option.id) }"
            :style="{
              backgroundColor: getOptionTheme(idx).color,
              borderColor: getOptionTheme(idx).border
            }"
            @click="handleOptionClick(option)"
            :data-testid="`btn-option-${idx}`"
          >
            <div class="btn-shape-badge">
              <span class="btn-shape-icon">{{ getOptionTheme(idx).shapeIcon }}</span>
            </div>
            <span class="btn-text">{{ option.text }}</span>
            <div v-if="isOptionSelected(option.id)" class="btn-selected-check">
              <i class="pi pi-check" />
            </div>
          </button>
        </div>

        <!-- Botó d'enviament per a preguntes de selecció múltiple -->
        <div
          v-if="matchStore.currentQuestion?.type === 'multiple_choice'"
          class="multiple-submit-container"
        >
          <Button
            label="Confirmar i Enviar Respostes"
            icon="pi pi-send"
            severity="success"
            size="large"
            class="submit-multiple-btn"
            :disabled="matchStore.mySelectedAnswerIds.length === 0"
            @click="handleMultipleSubmit"
            data-testid="btn-submit-multiple-answer"
          />
        </div>
      </div>
    </div>

    <!-- PANTALLA 4: RESULTATS DE PREGUNTA (QUESTION RESULTS) -->
    <div
      v-else-if="matchStore.status === 'question_results'"
      class="game-card screen-results"
      data-testid="screen-player-results"
    >
      <!-- Feedback d'encert o error -->
      <div
        v-if="matchStore.lastAnswerResult?.isCorrect"
        class="result-box correct"
        data-testid="feedback-correct"
      >
        <div class="feedback-icon correct-icon">
          <i class="pi pi-check" />
        </div>
        <h2 class="result-title">Molt bé! Resposta Correcta!</h2>
        <div class="points-awarded">+1 punt</div>
      </div>

      <div v-else class="result-box incorrect" data-testid="feedback-incorrect">
        <div class="feedback-icon incorrect-icon">
          <i class="pi pi-times" />
        </div>
        <h2 class="result-title">Oh no! Resposta Incorrecta</h2>
        <div class="points-awarded">+0 punts</div>
      </div>

      <div class="score-summary-box">
        <p class="score-label">Puntuació total actual</p>
        <p class="score-val">{{ matchStore.myScore }} pts</p>
      </div>

      <p class="results-wait-hint">Mira la pantalla gran per veure les solucions!</p>
    </div>

    <!-- PANTALLA 5: RÀNQUING PARCIAL (LEADERBOARD) -->
    <div
      v-else-if="matchStore.status === 'leaderboard'"
      class="game-card screen-leaderboard"
      data-testid="screen-player-leaderboard"
    >
      <div class="leaderboard-icon">
        <i class="pi pi-chart-bar" />
      </div>
      <h2 class="leaderboard-title">Rànquing de la Partida</h2>
      <p class="leaderboard-subtitle">Mira la pantalla del moderador per veure les posicions!</p>

      <div class="player-score-highlight">
        <span>La teva puntuació:</span>
        <strong>{{ matchStore.myScore }} punts</strong>
      </div>
    </div>

    <!-- PANTALLA 6: PODI I FINAL (FINISHED) -->
    <div
      v-else-if="matchStore.status === 'finished'"
      class="game-card screen-finished"
      data-testid="screen-player-finished"
    >
      <div class="podium-trophy">
        <i class="pi pi-trophy" />
      </div>
      <h2 class="finished-title">Partida Finalitzada!</h2>
      <p class="finished-score">
        Has aconseguit un total de <strong>{{ matchStore.myScore }}</strong> punts.
      </p>

      <Button
        label="Tornar als Jocs"
        icon="pi pi-home"
        severity="primary"
        size="large"
        class="mt-4"
        @click="handleExit"
        data-testid="btn-player-finish-exit"
      />
    </div>
  </div>
</template>

<style scoped>
.player-game-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #0f172a;
  color: #ffffff;
  padding: 1rem;
}

/* Header */
.player-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(10px);
  border-radius: 1rem;
  margin-bottom: 1.5rem;
}

.player-nick {
  font-weight: 700;
  font-size: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.pin-tag {
  font-weight: 700;
  font-size: 0.9rem;
}

.score-badge {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  background: rgba(255, 255, 255, 0.15);
  padding: 0.35rem 0.75rem;
  border-radius: 9999px;
  font-weight: 700;
}

/* Base Game Card */
.game-card {
  max-width: 580px;
  width: 100%;
  margin: auto;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 1.5rem;
  padding: 2.5rem 1.5rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
}

/* Screen Lobby */
.pulse-avatar {
  width: 80px;
  height: 80px;
  background: #22c55e;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2.5rem;
  color: #ffffff;
  margin-bottom: 1.5rem;
  box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.7);
  animation: pulseGreen 2s infinite;
}

@keyframes pulseGreen {
  0% {
    box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.7);
  }
  70% {
    box-shadow: 0 0 0 20px rgba(34, 197, 94, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(34, 197, 94, 0);
  }
}

.lobby-title {
  font-size: 2.25rem;
  font-weight: 800;
  margin: 0 0 0.5rem 0;
}

.lobby-nick {
  font-size: 1.5rem;
  font-weight: 700;
  color: #60a5fa;
  margin: 0 0 1rem 0;
}

.lobby-instruction {
  color: #94a3b8;
  font-size: 0.95rem;
  max-width: 400px;
  line-height: 1.4;
  margin-bottom: 1.5rem;
}

.lobby-status-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  background: rgba(255, 255, 255, 0.1);
  padding: 0.5rem 1.25rem;
  border-radius: 9999px;
  font-size: 0.9rem;
  color: #e2e8f0;
}

/* Screen Preview */
.question-meta-badge {
  background: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.4);
  padding: 0.4rem 1rem;
  border-radius: 9999px;
  font-weight: 700;
  font-size: 0.9rem;
  margin-bottom: 1rem;
}

.preview-statement {
  font-size: 1.75rem;
  font-weight: 800;
  line-height: 1.3;
  margin-bottom: 1.5rem;
}

.preview-image-wrapper {
  width: 100%;
  max-height: 220px;
  border-radius: 1rem;
  overflow: hidden;
  margin-bottom: 1.5rem;
}

.preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-waiting-box {
  background: rgba(255, 255, 255, 0.06);
  border-radius: 1rem;
  padding: 1rem 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  text-align: left;
}

.preview-icon {
  font-size: 2rem;
  color: #38bdf8;
}

/* Screen Active */
.screen-active {
  max-width: 900px;
  width: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  flex: 1;
}

.timer-bar-container {
  margin-bottom: 1.5rem;
}

.timer-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.timer-count {
  font-size: 1.75rem;
  font-weight: 900;
  color: #fbbf24;
}

.question-pill {
  font-size: 0.95rem;
  font-weight: 600;
  color: #94a3b8;
}

.timer-progress {
  height: 12px;
  border-radius: 6px;
  background-color: rgba(255, 255, 255, 0.1);
}

.answers-container {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.options-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  flex: 1;
}

.kahoot-btn {
  display: flex;
  align-items: center;
  padding: 1.5rem;
  border-radius: 1rem;
  border: 3px solid transparent;
  cursor: pointer;
  color: #ffffff;
  font-size: 1.25rem;
  font-weight: 700;
  text-align: left;
  position: relative;
  transition: transform 0.1s ease, box-shadow 0.15s ease, filter 0.15s ease;
  min-height: 100px;
}

.kahoot-btn:hover {
  filter: brightness(1.1);
  transform: translateY(-2px);
}

.kahoot-btn:active {
  transform: scale(0.98);
}

.kahoot-btn.selected {
  border-color: #ffffff;
  box-shadow: 0 0 20px rgba(255, 255, 255, 0.6);
}

.btn-shape-badge {
  font-size: 1.75rem;
  margin-right: 1rem;
  display: flex;
  align-items: center;
}

.btn-text {
  flex: 1;
  line-height: 1.3;
}

.btn-selected-check {
  background: #ffffff;
  color: #0f172a;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 900;
  margin-left: 0.5rem;
}

.multiple-submit-container {
  margin-top: 1.5rem;
  display: flex;
  justify-content: center;
}

.submit-multiple-btn {
  padding: 1rem 2rem;
  font-size: 1.15rem;
  font-weight: 700;
}

.submitted-feedback-card {
  margin: auto;
  text-align: center;
  padding: 3rem 1.5rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 1.5rem;
}

.submitted-icon {
  font-size: 4rem;
  color: #38bdf8;
  margin-bottom: 1rem;
}

/* Results Feedback */
.result-box {
  width: 100%;
  padding: 1.5rem;
  border-radius: 1rem;
  margin-bottom: 1.5rem;
}

.result-box.correct {
  background: rgba(34, 197, 94, 0.15);
  border: 2px solid #22c55e;
}

.result-box.incorrect {
  background: rgba(239, 68, 68, 0.15);
  border: 2px solid #ef4444;
}

.feedback-icon {
  font-size: 3.5rem;
  margin-bottom: 0.5rem;
}

.correct-icon {
  color: #22c55e;
}

.incorrect-icon {
  color: #ef4444;
}

.result-title {
  font-size: 1.75rem;
  font-weight: 800;
  margin-bottom: 0.5rem;
}

.points-awarded {
  font-size: 1.5rem;
  font-weight: 700;
}

.score-summary-box {
  background: rgba(255, 255, 255, 0.08);
  border-radius: 1rem;
  padding: 1rem 2rem;
  margin-bottom: 1rem;
}

.score-label {
  font-size: 0.85rem;
  color: #94a3b8;
}

.score-val {
  font-size: 2rem;
  font-weight: 800;
  color: #fbbf24;
  margin: 0;
}

.results-wait-hint {
  color: #64748b;
  font-size: 0.9rem;
}

/* Screen Leaderboard / Finished */
.leaderboard-icon,
.podium-trophy {
  font-size: 4rem;
  color: #fbbf24;
  margin-bottom: 1rem;
}

.player-score-highlight {
  background: rgba(255, 255, 255, 0.1);
  padding: 0.75rem 1.5rem;
  border-radius: 9999px;
  display: flex;
  gap: 0.75rem;
  font-size: 1.1rem;
  margin-top: 1.5rem;
}

@media (max-width: 640px) {
  .options-grid {
    grid-template-columns: 1fr;
  }
}
</style>
