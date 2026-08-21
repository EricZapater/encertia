<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import type { QuizDetail, QuizQuestion } from '../types'
import { KAHOOT_THEME_SHAPES } from '../types'

const props = defineProps<{
  visible: boolean
  quiz: QuizDetail | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
}>()

const currentQuestionIndex = ref(0)
const timeLeft = ref(20)
const timerRunning = ref(false)
let timerInterval: ReturnType<typeof setInterval> | null = null

const selectedAnswerIds = ref<string[]>([])
const isAnswerSubmitted = ref(false)
const isQuizCompleted = ref(false)
const correctCount = ref(0)

const questions = computed<QuizQuestion[]>(() => props.quiz?.questions || [])
const currentQuestion = computed<QuizQuestion | null>(() => {
  if (questions.value.length === 0) return null
  return questions.value[currentQuestionIndex.value] || null
})

watch(
  () => props.visible,
  (val) => {
    if (val) {
      startQuiz()
    } else {
      stopTimer()
    }
  }
)

function startQuiz() {
  currentQuestionIndex.value = 0
  correctCount.value = 0
  isQuizCompleted.value = false
  loadQuestion(0)
}

function loadQuestion(index: number) {
  stopTimer()
  if (index >= questions.value.length) {
    isQuizCompleted.value = true
    return
  }

  currentQuestionIndex.value = index
  selectedAnswerIds.value = []
  isAnswerSubmitted.value = false

  const q = questions.value[index]
  timeLeft.value = q?.timeLimitSeconds || 20
  startTimer()
}

function startTimer() {
  timerRunning.value = true
  if (timerInterval) clearInterval(timerInterval)
  timerInterval = setInterval(() => {
    if (timeLeft.value > 1) {
      timeLeft.value--
    } else {
      timeLeft.value = 0
      stopTimer()
      // Temps exhaurit
      if (!isAnswerSubmitted.value) {
        submitAnswer()
      }
    }
  }, 1000)
}

function stopTimer() {
  timerRunning.value = false
  if (timerInterval) {
    clearInterval(timerInterval)
    timerInterval = null
  }
}

function handleOptionClick(answerId: string) {
  if (isAnswerSubmitted.value) return

  const q = currentQuestion.value
  if (!q) return

  if (q.questionType === 'single_choice') {
    selectedAnswerIds.value = [answerId]
    submitAnswer()
  } else {
    // multiple_choice
    const idx = selectedAnswerIds.value.indexOf(answerId)
    if (idx !== -1) {
      selectedAnswerIds.value.splice(idx, 1)
    } else {
      selectedAnswerIds.value.push(answerId)
    }
  }
}

function submitAnswer() {
  if (isAnswerSubmitted.value) return
  stopTimer()
  isAnswerSubmitted.value = true

  const q = currentQuestion.value
  if (!q) return

  const correctAnswers = q.answers.filter((a) => a.isCorrect).map((a) => a.id)
  const userSelected = [...selectedAnswerIds.value].sort()
  const expected = [...correctAnswers].sort()

  const isAllCorrect =
    userSelected.length === expected.length &&
    userSelected.every((id, idx) => id === expected[idx])

  if (isAllCorrect) {
    correctCount.value++
  }
}

function handleNextQuestion() {
  loadQuestion(currentQuestionIndex.value + 1)
}

function handleClose() {
  stopTimer()
  emit('update:visible', false)
}

function getShape(index: number) {
  return KAHOOT_THEME_SHAPES[index % KAHOOT_THEME_SHAPES.length]
}

onUnmounted(() => {
  stopTimer()
})
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    :header="`Previsualització: ${quiz?.title || 'Qüestionari'}`"
    :style="{ width: '95vw', maxWidth: '840px' }"
    @update:visible="handleClose"
    data-testid="quiz-preview-dialog"
  >
    <div class="preview-modal-body">
      <!-- Pantalla Final de Resultats -->
      <div v-if="isQuizCompleted" class="completed-screen" data-testid="preview-completed-screen">
        <div class="completed-badge">
          <i class="pi pi-trophy trophy-icon" />
        </div>
        <h2 class="completed-title">Simulació Completada!</h2>
        <p class="completed-score">
          Has encertat <strong>{{ correctCount }}</strong> de <strong>{{ questions.length }}</strong> preguntes.
        </p>
        <div class="score-percent">
          {{ questions.length > 0 ? Math.round((correctCount / questions.length) * 100) : 0 }}% d'encert
        </div>
        <div class="completed-actions">
          <Button
            label="Tornar a jugar"
            icon="pi pi-refresh"
            severity="secondary"
            @click="startQuiz"
            data-testid="btn-restart-preview"
          />
          <Button
            label="Sortir de la simulació"
            icon="pi pi-check"
            severity="primary"
            @click="handleClose"
            data-testid="btn-close-preview"
          />
        </div>
      </div>

      <!-- Pantalla sense preguntes -->
      <div v-else-if="!currentQuestion" class="empty-preview">
        <i class="pi pi-info-circle empty-icon" />
        <p>Aquest qüestionari no té cap pregunta per previsualitzar.</p>
        <Button label="Tancar" severity="secondary" @click="handleClose" />
      </div>

      <!-- Pregunta Activa -->
      <div v-else class="question-screen" data-testid="preview-question-screen">
        <!-- Barra d'estat de la pregunta -->
        <div class="question-header-bar">
          <div class="question-index-tag">
            Pregunta {{ currentQuestionIndex + 1 }} de {{ questions.length }}
          </div>

          <div class="timer-box" :class="{ 'timer-urgent': timeLeft <= 5 }">
            <i class="pi pi-stopwatch timer-icon" />
            <span class="timer-value" data-testid="preview-timer-value">{{ timeLeft }}s</span>
          </div>

          <Tag
            :value="currentQuestion.questionType === 'single_choice' ? 'Una sola correcta' : 'Múltiples correctes'"
            :severity="currentQuestion.questionType === 'single_choice' ? 'info' : 'warn'"
          />
        </div>

        <!-- Text de la pregunta -->
        <div class="question-content">
          <h3 class="question-text" data-testid="preview-question-text">{{ currentQuestion.text || '(Enunciat de la pregunta)' }}</h3>

          <div v-if="currentQuestion.imageUrl" class="question-image-container">
            <img :src="currentQuestion.imageUrl" alt="Imatge de la pregunta" class="question-image" />
          </div>
        </div>

        <!-- Graella de Respostes Kahoot -->
        <div class="answers-grid" :class="`answers-count-${currentQuestion.answers.length}`">
          <button
            v-for="(answer, idx) in currentQuestion.answers"
            :key="answer.id || idx"
            class="kahoot-answer-card"
            :style="{
              backgroundColor: getShape(idx).color,
              borderColor: getShape(idx).borderColor
            }"
            :class="{
              'is-selected': selectedAnswerIds.includes(answer.id),
              'is-revealed-correct': isAnswerSubmitted && answer.isCorrect,
              'is-revealed-incorrect': isAnswerSubmitted && selectedAnswerIds.includes(answer.id) && !answer.isCorrect,
              'is-dimmed': isAnswerSubmitted && !answer.isCorrect && !selectedAnswerIds.includes(answer.id)
            }"
            :disabled="isAnswerSubmitted"
            @click="handleOptionClick(answer.id)"
            :data-testid="`preview-answer-btn-${idx}`"
          >
            <div class="card-left-shape">
              <span class="shape-symbol">{{ getShape(idx).symbol }}</span>
            </div>
            <div class="card-answer-text">
              {{ answer.text || `Opció ${idx + 1}` }}
            </div>
            <div class="card-status-icon">
              <i
                v-if="isAnswerSubmitted && answer.isCorrect"
                class="pi pi-check-circle correct-icon"
              />
              <i
                v-else-if="isAnswerSubmitted && selectedAnswerIds.includes(answer.id) && !answer.isCorrect"
                class="pi pi-times-circle wrong-icon"
              />
            </div>
          </button>
        </div>

        <!-- Botons de navegació / Confirmació per a multiple choice -->
        <div class="preview-footer-actions">
          <Button
            v-if="currentQuestion.questionType === 'multiple_choice' && !isAnswerSubmitted"
            label="Enviar Resposta"
            icon="pi pi-check"
            severity="primary"
            :disabled="selectedAnswerIds.length === 0"
            @click="submitAnswer"
            data-testid="btn-submit-preview-multiple"
          />

          <Button
            v-if="isAnswerSubmitted"
            :label="currentQuestionIndex + 1 >= questions.length ? 'Finalitzar simulació' : 'Següent Pregunta'"
            icon="pi pi-arrow-right"
            severity="success"
            @click="handleNextQuestion"
            data-testid="btn-next-preview-question"
          />
        </div>
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
.preview-modal-body {
  min-height: 420px;
  display: flex;
  flex-direction: column;
}

.question-screen {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.question-header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid #e2e8f0;
}

.question-index-tag {
  font-weight: 700;
  color: #475569;
  font-size: 0.95rem;
}

.timer-box {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  background-color: #f1f5f9;
  border: 1px solid #cbd5e1;
  padding: 0.35rem 0.85rem;
  border-radius: 9999px;
  font-weight: 700;
  color: #0f172a;
}

.timer-urgent {
  background-color: #fee2e2;
  border-color: #ef4444;
  color: #b91c1c;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.timer-icon {
  font-size: 1rem;
}

.timer-value {
  font-size: 1.05rem;
}

.question-content {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.question-text {
  font-size: 1.35rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
  line-height: 1.35;
}

.question-image-container {
  max-width: 380px;
  max-height: 200px;
  border-radius: 0.5rem;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.question-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.answers-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.85rem;
}

.kahoot-answer-card {
  display: flex;
  align-items: center;
  border: 2px solid transparent;
  border-radius: 0.5rem;
  color: #ffffff;
  padding: 0.85rem 1rem;
  cursor: pointer;
  box-shadow: 0 3px 6px rgba(0, 0, 0, 0.15);
  transition: transform 0.15s ease, box-shadow 0.15s ease, filter 0.2s ease;
  min-height: 68px;
  position: relative;
  text-align: left;
}

.kahoot-answer-card:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 12px rgba(0, 0, 0, 0.25);
  filter: brightness(1.05);
}

.kahoot-answer-card.is-selected {
  outline: 4px solid #0f172a;
}

.kahoot-answer-card.is-revealed-correct {
  filter: brightness(1.15);
  border-color: #22c55e;
  box-shadow: 0 0 0 4px #22c55e;
}

.kahoot-answer-card.is-revealed-incorrect {
  opacity: 0.6;
  border-color: #ef4444;
}

.kahoot-answer-card.is-dimmed {
  opacity: 0.4;
  filter: grayscale(40%);
}

.card-left-shape {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.3rem;
  margin-right: 0.75rem;
  flex-shrink: 0;
}

.card-answer-text {
  flex: 1;
  font-size: 1.05rem;
  font-weight: 600;
  text-shadow: 0 1px 2px rgba(0,0,0,0.3);
  word-break: break-word;
}

.card-status-icon {
  font-size: 1.5rem;
  margin-left: 0.5rem;
}

.correct-icon {
  color: #ffffff;
}

.wrong-icon {
  color: #ffffff;
}

.preview-footer-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 0.5rem;
}

.completed-screen {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  text-align: center;
  gap: 1rem;
}

.completed-badge {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background-color: #fef3c7;
  display: flex;
  align-items: center;
  justify-content: center;
}

.trophy-icon {
  font-size: 2.25rem;
  color: #d97706;
}

.completed-title {
  font-size: 1.75rem;
  font-weight: 800;
  color: #0f172a;
  margin: 0;
}

.completed-score {
  font-size: 1.1rem;
  color: #475569;
  margin: 0;
}

.score-percent {
  font-size: 2rem;
  font-weight: 800;
  color: #2563eb;
}

.completed-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1rem;
}

.empty-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 1rem;
  gap: 1rem;
  color: #64748b;
}

.empty-icon {
  font-size: 2.5rem;
  color: #94a3b8;
}

@media (max-width: 640px) {
  .answers-grid {
    grid-template-columns: 1fr;
  }
}
</style>
