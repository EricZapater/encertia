<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import QRCode from 'qrcode'
import { useAuthStore } from '@/modules/auth/store'
import { useMatchStore } from '../store'
import { KAHOOT_THEMES, type MatchAnswerOption, type MatchPlayer } from '../types'

import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const matchStore = useMatchStore()

const generatedQrDataUrl = ref<string>('')
const playerToKick = ref<MatchPlayer | null>(null)
const showKickModal = ref(false)

const matchPin = computed(() => matchStore.pin || (route.params.pin as string) || (route.params.id as string) || '')
const playUrl = computed(() => {
  if (!matchPin.value) return ''
  if (typeof window !== 'undefined' && window.location.origin) {
    return `${window.location.origin}/play?pin=${matchPin.value}`
  }
  return `/play?pin=${matchPin.value}`
})
const joinUrl = playUrl

onMounted(async () => {
  if (!authStore.isInitialized) {
    await authStore.initAuth()
  }

  if (!authStore.isAuthenticated) {
    router.push({
      name: 'login',
      query: { redirect: route.fullPath }
    })
    return
  }

  // Si no està connectat com a Host, connecta
  const routeParam = (route.params.id as string) || (route.params.pin as string)
  if (!matchStore.isConnected || matchStore.role !== 'host') {
    if (routeParam) {
      try {
        await matchStore.connectAsHost(routeParam)
      } catch (err) {
        console.error('Error connectant com a Host:', err)
      }
    }
  }

  generateQr()
})

watch(
  () => matchPin.value,
  () => {
    generateQr()
  }
)

async function generateQr() {
  if (!matchPin.value) return
  try {
    generatedQrDataUrl.value = await QRCode.toDataURL(joinUrl.value, {
      width: 280,
      margin: 1,
      color: {
        dark: '#0f172a',
        light: '#ffffff'
      }
    })
  } catch (err) {
    console.error('Error generant codi QR:', err)
  }
}

function getOptionTheme(index: number) {
  return KAHOOT_THEMES[index % KAHOOT_THEMES.length]
}

function getOptionCount(option: MatchAnswerOption): number {
  if (typeof option.count === 'number') return option.count
  if (matchStore.lastAnswerResult?.optionCounts && typeof matchStore.lastAnswerResult.optionCounts[option.id] === 'number') {
    return matchStore.lastAnswerResult.optionCounts[option.id]
  }
  return 0
}

function getOptionCountPercentage(option: MatchAnswerOption): number {
  const total = matchStore.totalPlayers || matchStore.activePlayersCount || 1
  const count = getOptionCount(option)
  return Math.round((count / total) * 100)
}

function isOptionCorrect(option: MatchAnswerOption): boolean {
  if (option.isCorrect) return true
  if (matchStore.lastAnswerResult?.correctOptionIds?.includes(option.id)) return true
  return false
}

function openKickModal(player: MatchPlayer) {
  playerToKick.value = player
  showKickModal.value = true
}

function confirmKickPlayer() {
  if (playerToKick.value) {
    matchStore.kickPlayer(playerToKick.value.id)
    showKickModal.value = false
    playerToKick.value = null
  }
}

function handleStartMatch() {
  matchStore.startMatch()
}

function handleStartTimer() {
  matchStore.startQuestionTimer()
}

function handleShowResults() {
  matchStore.showResults()
}

function handleShowLeaderboard() {
  matchStore.showLeaderboard()
}

function handleNextQuestion() {
  matchStore.nextQuestion()
}

function handleExitHost() {
  matchStore.leaveMatch()
  router.push('/quizzes')
}
</script>

<template>
  <div class="host-container" :class="`phase-${matchStore.status || 'lobby'}`">
    <!-- Header Superior del Moderador -->
    <header class="host-header">
      <div class="host-header-left">
        <span class="quiz-title-badge">
          <i class="pi pi-bolt" /> {{ matchStore.quizTitle || 'Partida Encertia' }}
        </span>
      </div>

      <div class="host-header-center">
        <div class="pin-pill" data-testid="host-pin-display">
          <span class="pin-label">PIN DE LA SALA</span>
          <span class="pin-digits">{{ matchPin }}</span>
        </div>
      </div>

      <div class="host-header-right">
        <div class="players-counter-badge" data-testid="host-players-counter">
          <i class="pi pi-users" />
          <span>{{ matchStore.activePlayersCount }} jugadors</span>
        </div>
        <Button
          icon="pi pi-times"
          severity="secondary"
          rounded
          text
          tooltip="Sortir de la partida"
          @click="handleExitHost"
          data-testid="btn-host-exit"
        />
      </div>
    </header>

    <!-- CONTINGUT PRINCIPAL SEGONS L'ESTAT -->

    <!-- 1. FASE LOBBY (SALA D'ESPERA) -->
    <main
      v-if="!matchStore.status || matchStore.status === 'lobby'"
      class="host-main phase-lobby"
      data-testid="host-phase-lobby"
    >
      <div class="lobby-grid">
        <!-- Panell esquerre: PIN i Codi QR -->
        <div class="lobby-card qr-card">
          <h2 class="qr-instructions">Uneix-te a la partida des del teu mòbil o navegador:</h2>
          <p class="join-link-text">
            Entra a <strong>{{ joinUrl }}</strong>
          </p>

          <div class="qr-code-wrapper">
            <img
              v-if="generatedQrDataUrl || matchStore.qrCodeUrl"
              :src="generatedQrDataUrl || matchStore.qrCodeUrl || ''"
              alt="Codi QR per unir-se"
              class="qr-image"
              data-testid="host-qr-code"
            />
            <div v-else class="qr-placeholder">
              <i class="pi pi-qrcode qr-icon-placeholder" />
            </div>
          </div>

          <div class="pin-big-display">
            <span class="pin-hero">{{ matchPin }}</span>
          </div>

          <div class="start-action-container">
            <Button
              label="Començar Partida"
              icon="pi pi-play"
              size="large"
              severity="success"
              class="btn-start-game"
              :disabled="matchStore.activePlayersCount === 0"
              @click="handleStartMatch"
              data-testid="btn-start-match"
            />
            <small v-if="matchStore.activePlayersCount === 0" class="start-hint">
              Esperant que s'uneixi com a mínim 1 jugador...
            </small>
          </div>
        </div>

        <!-- Panell dret: Graella de jugadors connectats -->
        <div class="lobby-card players-card">
          <div class="players-card-header">
            <h3>Jugadors a la sala ({{ matchStore.activePlayersCount }})</h3>
          </div>

          <div v-if="matchStore.activePlayersCount === 0" class="empty-players-notice">
            <i class="pi pi-user-plus empty-players-icon" />
            <p>Els noms dels alumnes apareixeran aquí quan s'uneixin...</p>
          </div>

          <div v-else class="players-badges-grid" data-testid="host-players-grid">
            <div
              v-for="player in matchStore.players.filter((p) => p.isConnected && !p.isKicked)"
              :key="player.id"
              class="player-badge"
              :data-testid="`player-badge-${player.id}`"
            >
              <span class="player-name">{{ player.nickname }}</span>
              <button
                type="button"
                class="kick-player-btn"
                title="Expulsar jugador"
                @click="openKickModal(player)"
                :data-testid="`btn-kick-${player.id}`"
              >
                <i class="pi pi-times" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- 2. FASE QUESTION PREVIEW (PAUSA PRÈVIA DE LECTURA) -->
    <main
      v-else-if="matchStore.status === 'question_preview'"
      class="host-main phase-preview"
      data-testid="host-phase-preview"
    >
      <div class="preview-center-card">
        <div class="question-index-tag">
          Pregunta {{ matchStore.questionNumber }} de {{ matchStore.totalQuestions }}
        </div>

        <h1 class="host-question-title" data-testid="host-preview-title">
          {{ matchStore.currentQuestion?.title }}
        </h1>

        <div v-if="matchStore.currentQuestion?.imageUrl" class="host-question-img-wrapper">
          <img
            :src="matchStore.currentQuestion.imageUrl"
            alt="Imatge de la pregunta"
            class="host-question-img"
          />
        </div>

        <div class="preview-control-bar">
          <Button
            label="Iniciar Temps"
            icon="pi pi-clock"
            size="large"
            severity="primary"
            class="btn-action-hero"
            @click="handleStartTimer"
            data-testid="btn-start-timer"
          />
        </div>
      </div>
    </main>

    <!-- 3. FASE QUESTION ACTIVE (COMPTE ENRERE I RESPOSTES EN DIRECTE) -->
    <main
      v-else-if="matchStore.status === 'question_active'"
      class="host-main phase-active"
      data-testid="host-phase-active"
    >
      <!-- Barra superior: Enunciat + Timer + Comptador de respostes -->
      <div class="active-top-bar">
        <div class="active-timer-circle" data-testid="host-active-timer">
          <span class="timer-number">{{ matchStore.timerSeconds }}</span>
        </div>

        <h2 class="active-question-heading">{{ matchStore.currentQuestion?.title }}</h2>

        <div class="active-stats-pill" data-testid="host-answers-count">
          <span class="stats-num">{{ matchStore.answeredCount }} / {{ matchStore.activePlayersCount }}</span>
          <span class="stats-lbl">Respostes</span>
        </div>
      </div>

      <!-- Imatge opcional al centre -->
      <div v-if="matchStore.currentQuestion?.imageUrl" class="active-image-strip">
        <img :src="matchStore.currentQuestion.imageUrl" alt="Imatge" class="active-img" />
      </div>

      <!-- Graella de targetes d'opcions -->
      <div class="host-options-grid">
        <div
          v-for="(option, idx) in matchStore.currentQuestion?.options || []"
          :key="option.id"
          class="host-option-card"
          :style="{ backgroundColor: getOptionTheme(idx).color }"
        >
          <div class="host-shape-box">
            <span>{{ getOptionTheme(idx).shapeIcon }}</span>
          </div>
          <span class="host-option-text">{{ option.text }}</span>
        </div>
      </div>

      <!-- Botó d'acció per tancar abans d'hora si es vol -->
      <div class="active-actions-footer">
        <Button
          label="Tancar i Mostrar Resultats"
          icon="pi pi-check-circle"
          severity="warn"
          size="large"
          @click="handleShowResults"
          data-testid="btn-show-results"
        />
      </div>
    </main>

    <!-- 4. FASE QUESTION RESULTS (GRÀFIC DE BARRES I RESPOSTA CORRECTA) -->
    <main
      v-else-if="matchStore.status === 'question_results'"
      class="host-main phase-results"
      data-testid="host-phase-results"
    >
      <div class="results-heading-box">
        <h2 class="results-question-text">{{ matchStore.currentQuestion?.title }}</h2>
      </div>

      <!-- Gràfic de barres estil Kahoot -->
      <div class="results-bars-container" data-testid="host-results-chart">
        <div
          v-for="(option, idx) in matchStore.currentQuestion?.options || []"
          :key="option.id"
          class="bar-column"
          :class="{ 'is-correct-answer': isOptionCorrect(option) }"
        >
          <!-- Indicador de Vots / Recompte -->
          <div class="bar-count-label">{{ getOptionCount(option) }}</div>

          <!-- Barra de color vertical -->
          <div class="bar-track">
            <div
              class="bar-fill"
              :style="{
                height: `${Math.max(12, getOptionCountPercentage(option))}%`,
                backgroundColor: getOptionTheme(idx).color
              }"
            />
          </div>

          <!-- Icona de forma + Icona de correcta/incorrecta -->
          <div class="bar-footer-icon" :style="{ backgroundColor: getOptionTheme(idx).color }">
            <span>{{ getOptionTheme(idx).shapeIcon }}</span>
            <i v-if="isOptionCorrect(option)" class="pi pi-check correct-check-badge" />
          </div>

          <!-- Text de l'opció -->
          <div class="bar-option-text" :title="option.text">
            {{ option.text }}
          </div>
        </div>
      </div>

      <!-- Botons d'acció del moderador -->
      <div class="results-control-footer">
        <Button
          label="Veure Rànquing"
          icon="pi pi-chart-bar"
          size="large"
          severity="primary"
          class="btn-action-hero"
          @click="handleShowLeaderboard"
          data-testid="btn-show-leaderboard"
        />
      </div>
    </main>

    <!-- 5. FASE LEADERBOARD (RÀNQUING PARCIAL) -->
    <main
      v-else-if="matchStore.status === 'leaderboard'"
      class="host-main phase-leaderboard"
      data-testid="host-phase-leaderboard"
    >
      <div class="leaderboard-header-box">
        <h1 class="leaderboard-title">Rànquing Parcial</h1>
        <Tag
          :value="`Pregunta ${matchStore.questionNumber} de ${matchStore.totalQuestions}`"
          severity="info"
        />
      </div>

      <!-- Llista Top de jugadors -->
      <div class="leaderboard-list" data-testid="host-leaderboard-list">
        <div
          v-for="(item, idx) in matchStore.leaderboard.slice(0, 8)"
          :key="item.playerId"
          class="leaderboard-item"
          :class="`rank-${idx + 1}`"
        >
          <div class="rank-badge">{{ idx + 1 }}</div>
          <div class="player-info">
            <span class="player-nickname">{{ item.nickname }}</span>
          </div>
          <div class="player-points">{{ item.score }} pts</div>
        </div>
      </div>

      <!-- Botó de següent pregunta o finalitzar -->
      <div class="leaderboard-footer">
        <Button
          v-if="!matchStore.isLastQuestion"
          label="Següent Pregunta"
          icon="pi pi-arrow-right"
          size="large"
          severity="primary"
          class="btn-action-hero"
          @click="handleNextQuestion"
          data-testid="btn-next-question"
        />
        <Button
          v-else
          label="Finalitzar i Veure Podi"
          icon="pi pi-trophy"
          size="large"
          severity="success"
          class="btn-action-hero"
          @click="handleNextQuestion"
          data-testid="btn-finish-podium"
        />
      </div>
    </main>

    <!-- 6. FASE FINISHED (PODI FINAL 3D I CLASSIFICACIÓ COMPLETA) -->
    <main
      v-else-if="matchStore.status === 'finished'"
      class="host-main phase-finished"
      data-testid="host-phase-finished"
    >
      <div class="podium-header-box">
        <h1 class="podium-title">🎉 Podi de Campions 🎉</h1>
        <p class="podium-subtitle">{{ matchStore.quizTitle }}</p>
      </div>

      <!-- Podi dels 3 primers (2n, 1r, 3r) -->
      <div class="podium-stage" data-testid="host-podium-stage">
        <!-- 2n Lloc -->
        <div v-if="matchStore.podium[1]" class="podium-column silver-column">
          <div class="podium-player">
            <span class="podium-player-name">{{ matchStore.podium[1].nickname }}</span>
            <span class="podium-player-pts">{{ matchStore.podium[1].score }} pts</span>
          </div>
          <div class="podium-pedestal silver-pedestal">
            <span class="pedestal-rank">2</span>
          </div>
        </div>

        <!-- 1r Lloc -->
        <div v-if="matchStore.podium[0]" class="podium-column gold-column">
          <div class="podium-crown">👑</div>
          <div class="podium-player">
            <span class="podium-player-name">{{ matchStore.podium[0].nickname }}</span>
            <span class="podium-player-pts">{{ matchStore.podium[0].score }} pts</span>
          </div>
          <div class="podium-pedestal gold-pedestal">
            <span class="pedestal-rank">1</span>
          </div>
        </div>

        <!-- 3r Lloc -->
        <div v-if="matchStore.podium[2]" class="podium-column bronze-column">
          <div class="podium-player">
            <span class="podium-player-name">{{ matchStore.podium[2].nickname }}</span>
            <span class="podium-player-pts">{{ matchStore.podium[2].score }} pts</span>
          </div>
          <div class="podium-pedestal bronze-pedestal">
            <span class="pedestal-rank">3</span>
          </div>
        </div>
      </div>

      <!-- Botó de sortida -->
      <div class="finished-footer">
        <Button
          label="Tornar als meus Jocs"
          icon="pi pi-home"
          size="large"
          severity="primary"
          @click="handleExitHost"
          data-testid="btn-host-finish-exit"
        />
      </div>
    </main>

    <!-- DIÀLEG DE CONFIRMACIÓ D'EXPULSIÓ -->
    <Dialog
      v-model:visible="showKickModal"
      modal
      header="Expulsar Jugador"
      :style="{ width: '90vw', maxWidth: '400px' }"
      data-testid="modal-kick-player"
    >
      <p>
        Estàs segur que vols expulsar <strong>{{ playerToKick?.nickname }}</strong> d'aquesta partida?
      </p>
      <template #footer>
        <Button label="Cancel·lar" text severity="secondary" @click="showKickModal = false" />
        <Button
          label="Expulsar"
          severity="danger"
          @click="confirmKickPlayer"
          data-testid="btn-confirm-kick"
        />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.host-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #0b1329;
  color: #ffffff;
  font-family: inherit;
}

/* Header */
.host-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1.5rem;
  background: rgba(15, 23, 42, 0.9);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.quiz-title-badge {
  font-weight: 700;
  font-size: 1.1rem;
  color: #38bdf8;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.pin-pill {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: #1e293b;
  border: 1px solid #334155;
  padding: 0.35rem 1rem;
  border-radius: 9999px;
}

.pin-label {
  font-size: 0.75rem;
  color: #94a3b8;
  font-weight: 700;
}

.pin-digits {
  font-size: 1.35rem;
  font-weight: 900;
  letter-spacing: 0.15em;
  color: #f59e0b;
}

.host-header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.players-counter-badge {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 700;
  background: rgba(255, 255, 255, 0.1);
  padding: 0.4rem 0.85rem;
  border-radius: 9999px;
}

/* Main wrapper */
.host-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 1.5rem;
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
}

/* LOBBY */
.lobby-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  flex: 1;
}

.lobby-card {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 1.5rem;
  padding: 2rem;
  display: flex;
  flex-direction: column;
}

.qr-card {
  align-items: center;
  text-align: center;
}

.qr-instructions {
  font-size: 1.35rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
}

.join-link-text {
  font-size: 1rem;
  color: #38bdf8;
  margin-bottom: 1.5rem;
}

.qr-code-wrapper {
  background: #ffffff;
  padding: 0.75rem;
  border-radius: 1rem;
  margin-bottom: 1.5rem;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
}

.qr-image {
  width: 220px;
  height: 220px;
  display: block;
}

.pin-big-display {
  margin-bottom: 1.5rem;
}

.pin-hero {
  font-size: 3.5rem;
  font-weight: 900;
  letter-spacing: 0.2em;
  color: #f59e0b;
  text-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
}

.start-action-container {
  width: 100%;
  max-width: 320px;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.btn-start-game {
  width: 100%;
  font-size: 1.25rem;
  font-weight: 800;
  padding: 1rem;
}

.start-hint {
  color: #94a3b8;
  font-size: 0.8rem;
}

.players-card-header {
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding-bottom: 1rem;
  margin-bottom: 1rem;
}

.empty-players-notice {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #64748b;
  text-align: center;
  gap: 1rem;
}

.empty-players-icon {
  font-size: 3rem;
}

.players-badges-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-content: flex-start;
  overflow-y: auto;
  max-height: 500px;
}

.player-badge {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
  padding: 0.5rem 0.85rem;
  border-radius: 9999px;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 700;
  font-size: 1rem;
  animation: fadeIn 0.2s ease;
}

.kick-player-btn {
  background: transparent;
  border: none;
  color: #ef4444;
  cursor: pointer;
  padding: 0.2rem;
  display: flex;
  align-items: center;
  border-radius: 50%;
}

.kick-player-btn:hover {
  background: rgba(239, 68, 68, 0.2);
}

/* PREVIEW */
.preview-center-card {
  max-width: 900px;
  width: 100%;
  margin: auto;
  text-align: center;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 1.5rem;
  padding: 3rem 2rem;
}

.question-index-tag {
  display: inline-block;
  background: rgba(56, 189, 248, 0.2);
  color: #38bdf8;
  font-weight: 800;
  padding: 0.4rem 1.2rem;
  border-radius: 9999px;
  margin-bottom: 1.5rem;
}

.host-question-title {
  font-size: 2.5rem;
  font-weight: 900;
  line-height: 1.3;
  margin-bottom: 2rem;
}

.host-question-img-wrapper {
  max-height: 320px;
  border-radius: 1rem;
  overflow: hidden;
  margin-bottom: 2rem;
}

.host-question-img {
  max-height: 320px;
  object-fit: cover;
  border-radius: 1rem;
}

.btn-action-hero {
  font-size: 1.35rem;
  font-weight: 800;
  padding: 1rem 2.5rem;
}

/* ACTIVE */
.active-top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  background: rgba(255, 255, 255, 0.06);
  padding: 1rem 1.5rem;
  border-radius: 1rem;
  margin-bottom: 1.5rem;
}

.active-timer-circle {
  width: 70px;
  height: 70px;
  border-radius: 50%;
  background: #f59e0b;
  color: #0f172a;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  font-weight: 900;
}

.active-question-heading {
  font-size: 1.75rem;
  font-weight: 800;
  flex: 1;
  text-align: center;
  margin: 0;
}

.active-stats-pill {
  display: flex;
  flex-direction: column;
  align-items: center;
  background: rgba(255, 255, 255, 0.1);
  padding: 0.5rem 1.25rem;
  border-radius: 0.75rem;
}

.stats-num {
  font-size: 1.5rem;
  font-weight: 900;
  color: #38bdf8;
}

.stats-lbl {
  font-size: 0.75rem;
  color: #94a3b8;
}

.active-image-strip {
  text-align: center;
  margin-bottom: 1.5rem;
}

.active-img {
  max-height: 200px;
  border-radius: 0.75rem;
}

.host-options-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.25rem;
  flex: 1;
}

.host-option-card {
  display: flex;
  align-items: center;
  padding: 1.5rem;
  border-radius: 1rem;
  color: #ffffff;
  font-size: 1.5rem;
  font-weight: 700;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
}

.host-shape-box {
  font-size: 2rem;
  margin-right: 1.25rem;
}

.host-option-text {
  flex: 1;
}

.active-actions-footer {
  margin-top: 1.5rem;
  display: flex;
  justify-content: flex-end;
}

/* RESULTS */
.results-heading-box {
  text-align: center;
  margin-bottom: 2rem;
}

.results-question-text {
  font-size: 2rem;
  font-weight: 800;
}

.results-bars-container {
  display: flex;
  justify-content: center;
  align-items: flex-end;
  gap: 2rem;
  flex: 1;
  min-height: 340px;
  padding-bottom: 1rem;
}

.bar-column {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 140px;
  transition: transform 0.2s ease;
}

.bar-column.is-correct-answer {
  transform: scale(1.05);
}

.bar-count-label {
  font-size: 1.5rem;
  font-weight: 900;
  margin-bottom: 0.5rem;
}

.bar-track {
  width: 80px;
  height: 220px;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 0.75rem;
  display: flex;
  align-items: flex-end;
  overflow: hidden;
}

.bar-fill {
  width: 100%;
  border-radius: 0.75rem;
  transition: height 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.bar-footer-icon {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  margin-top: 0.75rem;
  position: relative;
}

.correct-check-badge {
  position: absolute;
  top: -8px;
  right: -8px;
  background: #22c55e;
  color: #ffffff;
  font-size: 0.9rem;
  padding: 0.3rem;
  border-radius: 50%;
  border: 2px solid #0b1329;
}

.bar-option-text {
  margin-top: 0.5rem;
  font-size: 0.9rem;
  font-weight: 600;
  text-align: center;
  color: #cbd5e1;
  max-width: 130px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.results-control-footer {
  display: flex;
  justify-content: center;
  margin-top: 1.5rem;
}

/* LEADERBOARD */
.leaderboard-header-box {
  text-align: center;
  margin-bottom: 2rem;
}

.leaderboard-title {
  font-size: 2.5rem;
  font-weight: 900;
  margin-bottom: 0.5rem;
}

.leaderboard-list {
  max-width: 700px;
  width: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  flex: 1;
}

.leaderboard-item {
  display: flex;
  align-items: center;
  padding: 1rem 1.5rem;
  background: rgba(255, 255, 255, 0.07);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 1rem;
}

.rank-badge {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #334155;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 900;
  font-size: 1.25rem;
  margin-right: 1.25rem;
}

.rank-1 .rank-badge {
  background: #f59e0b;
  color: #0f172a;
}
.rank-2 .rank-badge {
  background: #94a3b8;
  color: #0f172a;
}
.rank-3 .rank-badge {
  background: #d97706;
  color: #ffffff;
}

.player-info {
  flex: 1;
}

.player-nickname {
  font-size: 1.25rem;
  font-weight: 700;
}

.player-points {
  font-size: 1.35rem;
  font-weight: 900;
  color: #38bdf8;
}

.leaderboard-footer {
  display: flex;
  justify-content: center;
  margin-top: 2rem;
}

/* FINISHED / PODIUM */
.podium-header-box {
  text-align: center;
  margin-bottom: 2rem;
}

.podium-title {
  font-size: 3rem;
  font-weight: 900;
  color: #fbbf24;
}

.podium-subtitle {
  font-size: 1.25rem;
  color: #94a3b8;
}

.podium-stage {
  display: flex;
  justify-content: center;
  align-items: flex-end;
  gap: 1.5rem;
  flex: 1;
  min-height: 380px;
  margin-bottom: 2rem;
}

.podium-column {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 180px;
}

.podium-crown {
  font-size: 3rem;
  margin-bottom: -0.5rem;
  animation: bounce 1.5s infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.podium-player {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 0.75rem;
}

.podium-player-name {
  font-size: 1.35rem;
  font-weight: 800;
}

.podium-player-pts {
  font-size: 1.1rem;
  font-weight: 700;
  color: #38bdf8;
}

.podium-pedestal {
  width: 100%;
  border-radius: 1rem 1rem 0 0;
  display: flex;
  justify-content: center;
  padding-top: 1rem;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
}

.gold-pedestal {
  height: 240px;
  background: linear-gradient(180deg, #f59e0b 0%, #b45309 100%);
}

.silver-pedestal {
  height: 180px;
  background: linear-gradient(180deg, #94a3b8 0%, #475569 100%);
}

.bronze-pedestal {
  height: 130px;
  background: linear-gradient(180deg, #d97706 0%, #78350f 100%);
}

.pedestal-rank {
  font-size: 3rem;
  font-weight: 900;
  color: rgba(255, 255, 255, 0.9);
}

.finished-footer {
  display: flex;
  justify-content: center;
  margin-top: 1rem;
}

@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1); }
}

@media (max-width: 900px) {
  .lobby-grid {
    grid-template-columns: 1fr;
  }
  .host-options-grid {
    grid-template-columns: 1fr;
  }
}
</style>
