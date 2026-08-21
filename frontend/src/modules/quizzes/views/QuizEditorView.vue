<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuizStore } from '../store'
import type {
  QuizDetail,
  QuizQuestion,
  QuizAnswer,
  QuestionType,
  TimeLimitSeconds,
  SaveQuestionInput,
  SaveAnswerInput
} from '../types'
import { TIME_LIMIT_OPTIONS, KAHOOT_THEME_SHAPES } from '../types'

import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'

import QuizSettingsModal from './QuizSettingsModal.vue'
import QuizPreviewModal from './QuizPreviewModal.vue'

const route = useRoute()
const router = useRouter()
const quizStore = useQuizStore()

const quizId = computed(() => route.params.id as string | undefined)
const isNewQuiz = computed(() => !quizId.value || quizId.value === 'new')

// Estat de treball local del qüestionari
const currentQuiz = ref<QuizDetail>({
  id: '',
  creatorId: '',
  title: '',
  description: '',
  coverImageUrl: null,
  status: 'draft',
  tags: [],
  questionCount: 0,
  createdAt: '',
  updatedAt: '',
  questions: []
})

const selectedQuestionIndex = ref(0)
const isUploadingImage = ref(false)
const questionFileInput = ref<HTMLInputElement | null>(null)

// Modals
const showSettingsModal = ref(false)
const showPreviewModal = ref(false)

const feedbackMessage = ref<{ type: 'success' | 'error'; text: string } | null>(null)

const activeQuestion = computed<QuizQuestion | undefined>(() => {
  return currentQuiz.value.questions[selectedQuestionIndex.value]
})

const questionTypeOptions = [
  { label: 'Una sola resposta correcta', value: 'single_choice' as QuestionType },
  { label: 'Múltiples respostes correctes', value: 'multiple_choice' as QuestionType }
]

const timeLimitOptions = TIME_LIMIT_OPTIONS.map((sec) => ({
  label: `${sec} segons`,
  value: sec as TimeLimitSeconds
}))

onMounted(async () => {
  if (isNewQuiz.value) {
    const initialized = quizStore.initNewQuiz()
    currentQuiz.value = JSON.parse(JSON.stringify(initialized))
    selectedQuestionIndex.value = 0
  } else if (quizId.value) {
    try {
      const fetched = await quizStore.fetchQuizDetail(quizId.value)
      currentQuiz.value = JSON.parse(JSON.stringify(fetched))
      if (!currentQuiz.value.questions || currentQuiz.value.questions.length === 0) {
        addQuestion()
      }
      selectedQuestionIndex.value = 0
    } catch (err: any) {
      feedbackMessage.value = {
        type: 'error',
        text: err.response?.data?.error || err.message || 'Error en carregar el qüestionari.'
      }
    }
  }
})

// Gestió de Preguntes
function addQuestion() {
  const newIndex = currentQuiz.value.questions.length
  const newQuestion: QuizQuestion = {
    id: `q-temp-${Date.now()}-${Math.random().toString(36).substring(2, 7)}`,
    text: '',
    imageUrl: null,
    questionType: 'single_choice',
    timeLimitSeconds: 20,
    orderIndex: newIndex,
    answers: [
      { id: `ans-${Date.now()}-0`, text: '', isCorrect: true, orderIndex: 0 },
      { id: `ans-${Date.now()}-1`, text: '', isCorrect: false, orderIndex: 1 },
      { id: `ans-${Date.now()}-2`, text: '', isCorrect: false, orderIndex: 2 },
      { id: `ans-${Date.now()}-3`, text: '', isCorrect: false, orderIndex: 3 }
    ]
  }

  currentQuiz.value.questions.push(newQuestion)
  currentQuiz.value.questionCount = currentQuiz.value.questions.length
  selectedQuestionIndex.value = newIndex
}

function duplicateQuestion(index: number) {
  const source = currentQuiz.value.questions[index]
  if (!source) return

  const clonedAnswers: QuizAnswer[] = source.answers.map((a, i) => ({
    ...a,
    id: `ans-${Date.now()}-${i}`,
    orderIndex: i
  }))

  const clonedQuestion: QuizQuestion = {
    ...source,
    id: `q-temp-${Date.now()}-${Math.random().toString(36).substring(2, 7)}`,
    text: source.text ? `${source.text} (Còpia)` : '',
    orderIndex: index + 1,
    answers: clonedAnswers
  }

  currentQuiz.value.questions.splice(index + 1, 0, clonedQuestion)
  reindexQuestions()
  selectedQuestionIndex.value = index + 1
}

function removeQuestion(index: number) {
  if (currentQuiz.value.questions.length <= 1) {
    feedbackMessage.value = {
      type: 'error',
      text: 'El qüestionari ha de tenir com a mínim 1 pregunta.'
    }
    return
  }

  currentQuiz.value.questions.splice(index, 1)
  reindexQuestions()
  if (selectedQuestionIndex.value >= currentQuiz.value.questions.length) {
    selectedQuestionIndex.value = currentQuiz.value.questions.length - 1
  }
}

function moveQuestion(index: number, direction: 'up' | 'down') {
  const targetIndex = direction === 'up' ? index - 1 : index + 1
  if (targetIndex < 0 || targetIndex >= currentQuiz.value.questions.length) return

  const item = currentQuiz.value.questions.splice(index, 1)[0]
  currentQuiz.value.questions.splice(targetIndex, 0, item)
  reindexQuestions()
  selectedQuestionIndex.value = targetIndex
}

function reindexQuestions() {
  currentQuiz.value.questions.forEach((q, idx) => {
    q.orderIndex = idx
  })
  currentQuiz.value.questionCount = currentQuiz.value.questions.length
}

// Gestió de Respostes
function addAnswerOption() {
  if (!activeQuestion.value) return
  if (activeQuestion.value.answers.length >= 6) {
    feedbackMessage.value = {
      type: 'error',
      text: 'El màxim permès és de 6 opcions de resposta per pregunta.'
    }
    return
  }

  const nextIndex = activeQuestion.value.answers.length
  activeQuestion.value.answers.push({
    id: `ans-${Date.now()}-${nextIndex}`,
    text: '',
    isCorrect: false,
    orderIndex: nextIndex
  })
}

function removeAnswerOption(ansIndex: number) {
  if (!activeQuestion.value) return
  if (activeQuestion.value.answers.length <= 2) {
    feedbackMessage.value = {
      type: 'error',
      text: 'Com a mínim calen 2 opcions de resposta.'
    }
    return
  }

  activeQuestion.value.answers.splice(ansIndex, 1)
  activeQuestion.value.answers.forEach((ans, i) => {
    ans.orderIndex = i
  })

  // Si hem eliminat la correcta en single choice, assegurar que n'hi hagi una
  if (activeQuestion.value.questionType === 'single_choice') {
    const hasCorrect = activeQuestion.value.answers.some((a) => a.isCorrect)
    if (!hasCorrect && activeQuestion.value.answers.length > 0) {
      activeQuestion.value.answers[0].isCorrect = true
    }
  }
}

function toggleAnswerCorrect(ansIndex: number) {
  if (!activeQuestion.value) return
  const ans = activeQuestion.value.answers[ansIndex]
  if (!ans) return

  if (activeQuestion.value.questionType === 'single_choice') {
    // Només 1 pot ser correcta
    activeQuestion.value.answers.forEach((a, idx) => {
      a.isCorrect = idx === ansIndex
    })
  } else {
    // Multiple choice: canvi lliure de boolean
    ans.isCorrect = !ans.isCorrect
  }
}

function handleQuestionTypeChange(newType: QuestionType) {
  if (!activeQuestion.value) return
  activeQuestion.value.questionType = newType

  if (newType === 'single_choice') {
    // Assegura exactament 1 correcta
    let firstFound = false
    activeQuestion.value.answers.forEach((a) => {
      if (a.isCorrect && !firstFound) {
        firstFound = true
      } else {
        a.isCorrect = false
      }
    })
    if (!firstFound && activeQuestion.value.answers.length > 0) {
      activeQuestion.value.answers[0].isCorrect = true
    }
  }
}

// Pujada d'imatges de la pregunta
function triggerQuestionImageUpload() {
  questionFileInput.value?.click()
}

async function handleQuestionImageChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file || !activeQuestion.value) return

  if (file.size > 5 * 1024 * 1024) {
    feedbackMessage.value = {
      type: 'error',
      text: 'La imatge supera la mida màxima de 5 MB.'
    }
    return
  }

  isUploadingImage.value = true
  feedbackMessage.value = null
  try {
    const url = await quizStore.uploadImage(file)
    activeQuestion.value.imageUrl = url
  } catch (err: any) {
    feedbackMessage.value = {
      type: 'error',
      text: err.message || "Error en pujar la imatge de la pregunta."
    }
  } finally {
    isUploadingImage.value = false
    if (target) target.value = ''
  }
}

function removeQuestionImage() {
  if (activeQuestion.value) {
    activeQuestion.value.imageUrl = null
  }
}

// Actualització des de modal de configuració
async function handleSettingsSaved(updatedFields: Partial<QuizDetail>) {
  // Primer actualitzem l'estat local per reflectir els canvis a la UI immediatament
  Object.assign(currentQuiz.value, updatedFields)

  // Si el quiz ja existeix al servidor, persistim les metadades via PUT
  if (!isNewQuiz.value && quizId.value) {
    try {
      const updated = await quizStore.updateQuiz(quizId.value, {
        title: currentQuiz.value.title,
        description: currentQuiz.value.description,
        coverImageUrl: currentQuiz.value.coverImageUrl,
        status: currentQuiz.value.status,
        tags: currentQuiz.value.tags
        // No enviem `questions` → el backend conserva les preguntes existents
      })
      // Sincronitzem amb la resposta del servidor
      Object.assign(currentQuiz.value, updated)
      feedbackMessage.value = {
        type: 'success',
        text: `Configuració desada. Estat: "${updated.status}".`
      }
    } catch (err: any) {
      feedbackMessage.value = {
        type: 'error',
        text: err.message || 'Error en desar la configuració del qüestionari.'
      }
    }
  } else {
    // Quiz nou: només actualitzem l'estat local (es desarà al crear)
    feedbackMessage.value = {
      type: 'success',
      text: 'Configuració del qüestionari actualitzada.'
    }
  }
}

// Validació i Desar
function validateQuiz(): string | null {
  if (!currentQuiz.value.title || currentQuiz.value.title.trim().length < 3) {
    return 'El títol del qüestionari ha de tenir com a mínim 3 caràcters.'
  }

  if (currentQuiz.value.questions.length === 0) {
    return 'El qüestionari ha de tenir almenys 1 pregunta.'
  }

  for (let i = 0; i < currentQuiz.value.questions.length; i++) {
    const q = currentQuiz.value.questions[i]
    if (!q.text || !q.text.trim()) {
      return `La pregunta #${i + 1} no té enunciat.`
    }

    if (!q.answers || q.answers.length < 2) {
      return `La pregunta #${i + 1} ha de tenir com a mínim 2 opcions de resposta.`
    }

    for (let j = 0; j < q.answers.length; j++) {
      if (!q.answers[j].text || !q.answers[j].text.trim()) {
        return `A la pregunta #${i + 1}, l'opció de resposta #${j + 1} està buida.`
      }
    }

    const correctAnswers = q.answers.filter((a) => a.isCorrect)
    if (q.questionType === 'single_choice') {
      if (correctAnswers.length !== 1) {
        return `A la pregunta #${i + 1} ("${q.text}"), cal marcar exactament 1 resposta com a correcta.`
      }
    } else {
      if (correctAnswers.length === 0) {
        return `A la pregunta #${i + 1} ("${q.text}"), cal marcar com a mínim 1 resposta correcta.`
      }
    }
  }

  return null
}

function cleanUUID(id?: string | null): string | undefined {
  if (!id) return undefined
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
  return uuidRegex.test(id) ? id : undefined
}

async function handleSaveQuiz() {
  feedbackMessage.value = null
  const validationError = validateQuiz()
  if (validationError) {
    feedbackMessage.value = { type: 'error', text: validationError }
    return
  }

  const formattedQuestions: SaveQuestionInput[] = currentQuiz.value.questions.map(
    (q, qIdx) => ({
      id: cleanUUID(q.id),
      text: q.text.trim(),
      imageUrl: q.imageUrl,
      questionType: q.questionType,
      timeLimitSeconds: q.timeLimitSeconds,
      orderIndex: qIdx,
      answers: q.answers.map((a, aIdx) => ({
        id: cleanUUID(a.id),
        text: a.text.trim(),
        isCorrect: a.isCorrect,
        orderIndex: aIdx
      })) as SaveAnswerInput[]
    })
  )

  try {
    if (isNewQuiz.value) {
      const created = await quizStore.createQuiz({
        title: currentQuiz.value.title.trim(),
        description: currentQuiz.value.description?.trim() || null,
        coverImageUrl: currentQuiz.value.coverImageUrl,
        status: currentQuiz.value.status,
        tags: currentQuiz.value.tags,
        questions: formattedQuestions
      })
      feedbackMessage.value = {
        type: 'success',
        text: `Qüestionari "${created.title}" creat correctament!`
      }
      // Naveguem a la ruta d'edició del quiz creat
      router.replace(`/quizzes/${created.id}/edit`)
    } else if (quizId.value) {
      const updated = await quizStore.updateQuiz(quizId.value, {
        title: currentQuiz.value.title.trim(),
        description: currentQuiz.value.description?.trim() || null,
        coverImageUrl: currentQuiz.value.coverImageUrl,
        status: currentQuiz.value.status,
        tags: currentQuiz.value.tags,
        questions: formattedQuestions
      })
      feedbackMessage.value = {
        type: 'success',
        text: `Qüestionari "${updated.title}" desat correctament.`
      }
    }
  } catch (err: any) {
    feedbackMessage.value = {
      type: 'error',
      text: err.response?.data?.error || err.message || 'Error en desar el qüestionari.'
    }
  }
}

function handleExit() {
  router.push('/quizzes')
}

function getShape(index: number) {
  return KAHOOT_THEME_SHAPES[index % KAHOOT_THEME_SHAPES.length]
}
</script>

<template>
  <div class="quiz-editor-layout">
    <!-- Barra Superior -->
    <header class="editor-topbar">
      <div class="topbar-left">
        <Button
          icon="pi pi-arrow-left"
          text
          rounded
          severity="secondary"
          tooltip="Tornar a la llista"
          @click="handleExit"
          data-testid="btn-exit-editor"
        />

        <div class="quiz-meta-info" @click="showSettingsModal = true">
          <h2 class="quiz-main-title" :title="currentQuiz.title || 'Sense títol'">
            {{ currentQuiz.title || 'Nou Qüestionari (Fes clic per titular)' }}
          </h2>
          <div class="quiz-submeta">
            <Tag
              :value="currentQuiz.status === 'published' ? 'Publicat' : 'Esborrany'"
              :severity="currentQuiz.status === 'published' ? 'success' : 'warn'"
            />
            <span class="settings-hint"><i class="pi pi-cog" /> Configuració</span>
          </div>
        </div>
      </div>

      <div class="topbar-right">
        <Button
          label="Configuració"
          icon="pi pi-cog"
          severity="secondary"
          outlined
          size="small"
          @click="showSettingsModal = true"
          data-testid="btn-open-settings"
        />

        <Button
          label="Previsualitzar"
          icon="pi pi-play"
          severity="secondary"
          size="small"
          @click="showPreviewModal = true"
          data-testid="btn-open-preview"
        />

        <Button
          label="Desar Joc"
          icon="pi pi-save"
          severity="primary"
          size="small"
          :loading="quizStore.isSaving"
          @click="handleSaveQuiz"
          data-testid="btn-save-quiz"
        />
      </div>
    </header>

    <!-- Alertes de feedback -->
    <div v-if="feedbackMessage" class="feedback-banner">
      <Message
        :severity="feedbackMessage.type"
        :closable="true"
        @close="feedbackMessage = null"
        data-testid="editor-feedback-msg"
      >
        {{ feedbackMessage.text }}
      </Message>
    </div>

    <!-- Cos de l'Editor: Panell Esquerre + Panell Central -->
    <div class="editor-workspace">
      <!-- Panell Lateral Esquerre: Llista de Preguntes -->
      <aside class="questions-sidebar" data-testid="questions-sidebar">
        <div class="sidebar-header">
          <span class="sidebar-title">Preguntes ({{ currentQuiz.questions.length }})</span>
          <Button
            label="Afegir"
            icon="pi pi-plus"
            size="small"
            severity="primary"
            text
            @click="addQuestion"
            data-testid="btn-add-question"
          />
        </div>

        <div class="questions-scroll-list">
          <div
            v-for="(question, qIdx) in currentQuiz.questions"
            :key="question.id || qIdx"
            class="question-thumb-card"
            :class="{ 'is-active': selectedQuestionIndex === qIdx }"
            @click="selectedQuestionIndex = qIdx"
            :data-testid="`question-thumb-${qIdx}`"
          >
            <div class="thumb-header">
              <span class="thumb-number">{{ qIdx + 1 }}</span>
              <span class="thumb-type-badge">{{ question.timeLimitSeconds }}s</span>
            </div>

            <div class="thumb-preview-box">
              <img
                v-if="question.imageUrl"
                :src="question.imageUrl"
                alt="Miniatura"
                class="thumb-img"
              />
              <span v-else class="thumb-snippet">
                {{ question.text || '(Enunciat buit)' }}
              </span>
            </div>

            <!-- Botons d'acció de la miniatura -->
            <div class="thumb-actions" @click.stop>
              <Button
                icon="pi pi-arrow-up"
                text
                rounded
                size="small"
                severity="secondary"
                :disabled="qIdx === 0"
                tooltip="Pujar ordre"
                @click="moveQuestion(qIdx, 'up')"
                data-testid="btn-move-up-question"
              />
              <Button
                icon="pi pi-arrow-down"
                text
                rounded
                size="small"
                severity="secondary"
                :disabled="qIdx === currentQuiz.questions.length - 1"
                tooltip="Baixar ordre"
                @click="moveQuestion(qIdx, 'down')"
                data-testid="btn-move-down-question"
              />
              <Button
                icon="pi pi-copy"
                text
                rounded
                size="small"
                severity="secondary"
                tooltip="Duplicar pregunta"
                @click="duplicateQuestion(qIdx)"
                data-testid="btn-duplicate-question"
              />
              <Button
                icon="pi pi-trash"
                text
                rounded
                size="small"
                severity="danger"
                :disabled="currentQuiz.questions.length <= 1"
                tooltip="Eliminar pregunta"
                @click="removeQuestion(qIdx)"
                data-testid="btn-delete-question"
              />
            </div>
          </div>
        </div>

        <div class="sidebar-footer">
          <Button
            label="+ Afegir Pregunta"
            severity="secondary"
            outlined
            class="w-full"
            @click="addQuestion"
            data-testid="btn-add-question-bottom"
          />
        </div>
      </aside>

      <!-- Panell Central: Editor de la Pregunta Seleccionada -->
      <main class="question-main-editor" v-if="activeQuestion" data-testid="active-question-editor">
        <!-- Barra de controls de la pregunta -->
        <div class="question-controls-bar">
          <div class="control-group">
            <label class="control-label">Tipus de pregunta:</label>
            <Select
              :modelValue="activeQuestion.questionType"
              :options="questionTypeOptions"
              optionLabel="label"
              optionValue="value"
              class="control-select"
              @update:modelValue="handleQuestionTypeChange"
              data-testid="select-question-type"
            />
          </div>

          <div class="control-group">
            <label class="control-label">Temps límit:</label>
            <Select
              v-model="activeQuestion.timeLimitSeconds"
              :options="timeLimitOptions"
              optionLabel="label"
              optionValue="value"
              class="control-select"
              data-testid="select-time-limit"
            />
          </div>
        </div>

        <!-- Input d'Enunciat de la pregunta -->
        <div class="question-prompt-container">
          <Textarea
            v-model="activeQuestion.text"
            rows="2"
            autoResize
            class="question-prompt-input"
            placeholder="Escriu aquí l'enunciat de la pregunta..."
            data-testid="input-question-text"
          />
        </div>

        <!-- Zona d'Imatge de la Pregunta -->
        <div class="question-media-zone">
          <div v-if="activeQuestion.imageUrl" class="question-image-preview-box">
            <img :src="activeQuestion.imageUrl" alt="Imatge de la pregunta" class="question-active-img" />
            <div class="image-overlay-actions">
              <Button
                icon="pi pi-pencil"
                severity="secondary"
                size="small"
                tooltip="Canviar Imatge"
                @click="triggerQuestionImageUpload"
              />
              <Button
                icon="pi pi-trash"
                severity="danger"
                size="small"
                tooltip="Eliminar Imatge"
                @click="removeQuestionImage"
                data-testid="btn-remove-question-image"
              />
            </div>
          </div>

          <div v-else class="question-image-dropzone" @click="triggerQuestionImageUpload">
            <i class="pi pi-image dropzone-icon" />
            <span class="dropzone-label">
              <span v-if="isUploadingImage">Pujant imatge a Cloudflare R2...</span>
              <span v-else>Afegir Imatge a la pregunta (opcional)</span>
            </span>
          </div>

          <input
            ref="questionFileInput"
            type="file"
            accept="image/png, image/jpeg, image/webp, image/gif"
            class="hidden-file-input"
            @change="handleQuestionImageChange"
          />
        </div>

        <!-- Graella de Respostes Kahoot -->
        <div class="answers-editor-section">
          <div class="answers-header-info">
            <span class="answers-title">Opcions de Resposta ({{ activeQuestion.answers.length }}/6)</span>
            <span class="answers-hint">
              {{ activeQuestion.questionType === 'single_choice' ? 'Fes clic al cercle per marcar la resposta correcta' : 'Marca les caselles de totes les respostes correctes' }}
            </span>
          </div>

          <div class="kahoot-editor-grid">
            <div
              v-for="(answer, aIdx) in activeQuestion.answers"
              :key="answer.id || aIdx"
              class="kahoot-editor-card"
              :style="{
                backgroundColor: getShape(aIdx).color,
                borderColor: getShape(aIdx).borderColor
              }"
              :data-testid="`editor-answer-card-${aIdx}`"
            >
              <!-- Símbol de la forma Kahoot -->
              <div class="card-shape-badge">
                <span class="shape-char">{{ getShape(aIdx).symbol }}</span>
              </div>

              <!-- Input de text de la resposta -->
              <div class="card-input-wrapper">
                <InputText
                  v-model="answer.text"
                  class="answer-text-input"
                  :placeholder="`Escriu la resposta ${aIdx + 1}...`"
                  :data-testid="`input-answer-text-${aIdx}`"
                />
              </div>

              <!-- Botó de selecció de Correcta -->
              <button
                type="button"
                class="correct-toggle-btn"
                :class="{ 'is-correct-active': answer.isCorrect }"
                @click="toggleAnswerCorrect(aIdx)"
                :title="answer.isCorrect ? 'Resposta Correcta' : 'Marcar com a Correcta'"
                :data-testid="`btn-toggle-correct-${aIdx}`"
              >
                <i v-if="answer.isCorrect" class="pi pi-check" />
              </button>

              <!-- Botó d'eliminar opció (si > 2) -->
              <button
                v-if="activeQuestion.answers.length > 2"
                type="button"
                class="remove-answer-btn"
                @click="removeAnswerOption(aIdx)"
                title="Eliminar aquesta opció"
                :data-testid="`btn-remove-answer-${aIdx}`"
              >
                <i class="pi pi-times" />
              </button>
            </div>
          </div>

          <!-- Botó d'afegir nova resposta si < 6 -->
          <div v-if="activeQuestion.answers.length < 6" class="add-answer-row">
            <Button
              label="+ Afegir opció de resposta"
              icon="pi pi-plus"
              severity="secondary"
              outlined
              size="small"
              @click="addAnswerOption"
              data-testid="btn-add-answer-option"
            />
          </div>
        </div>
      </main>
    </div>

    <!-- Modals -->
    <QuizSettingsModal
      v-model:visible="showSettingsModal"
      :quiz="currentQuiz"
      @saved="handleSettingsSaved"
    />

    <QuizPreviewModal
      v-model:visible="showPreviewModal"
      :quiz="currentQuiz"
    />
  </div>
</template>

<style scoped>
.quiz-editor-layout {
  min-height: 100vh;
  background-color: #f1f5f9;
  display: flex;
  flex-direction: column;
}

.editor-topbar {
  background-color: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  padding: 0.6rem 1.25rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.quiz-meta-info {
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.quiz-main-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
  max-width: 380px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.quiz-submeta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.settings-hint {
  font-size: 0.75rem;
  color: #64748b;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.feedback-banner {
  padding: 0.5rem 1.25rem 0 1.25rem;
}

.editor-workspace {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* Panell Lateral de Preguntes */
.questions-sidebar {
  width: 260px;
  background-color: #ffffff;
  border-right: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sidebar-title {
  font-size: 0.875rem;
  font-weight: 700;
  color: #334155;
}

.questions-scroll-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.question-thumb-card {
  border: 2px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 0.5rem;
  background-color: #f8fafc;
  cursor: pointer;
  transition: all 0.15s ease;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.question-thumb-card:hover {
  border-color: #cbd5e1;
}

.question-thumb-card.is-active {
  border-color: #3b82f6;
  background-color: #eff6ff;
}

.thumb-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.thumb-number {
  font-size: 0.8rem;
  font-weight: 700;
  color: #1e293b;
}

.thumb-type-badge {
  font-size: 0.7rem;
  font-weight: 600;
  background-color: #e2e8f0;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  color: #475569;
}

.thumb-preview-box {
  height: 48px;
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem 0.4rem;
}

.thumb-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.thumb-snippet {
  font-size: 0.75rem;
  color: #64748b;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-align: center;
}

.thumb-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.15rem;
}

.sidebar-footer {
  padding: 0.75rem;
  border-top: 1px solid #f1f5f9;
}

/* Panell Central d'Edició */
.question-main-editor {
  flex: 1;
  overflow-y: auto;
  padding: 1.25rem 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  max-width: 980px;
  margin: 0 auto;
  width: 100%;
}

.question-controls-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.control-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.control-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: #475569;
}

.control-select {
  min-width: 200px;
}

.question-prompt-container {
  width: 100%;
}

.question-prompt-input {
  width: 100%;
  font-size: 1.25rem;
  font-weight: 600;
  text-align: center;
  padding: 1rem;
  border-radius: 0.75rem;
  border: 1px solid #cbd5e1;
  background-color: #ffffff;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

.question-media-zone {
  display: flex;
  justify-content: center;
}

.question-image-preview-box {
  position: relative;
  max-width: 320px;
  max-height: 180px;
  border-radius: 0.5rem;
  overflow: hidden;
  border: 1px solid #cbd5e1;
  box-shadow: 0 2px 4px rgba(0,0,0,0.08);
}

.question-active-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-overlay-actions {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  display: flex;
  gap: 0.35rem;
  background: rgba(0, 0, 0, 0.4);
  padding: 0.25rem;
  border-radius: 0.35rem;
}

.question-image-dropzone {
  border: 2px dashed #cbd5e1;
  border-radius: 0.5rem;
  padding: 1rem 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  background-color: #ffffff;
  transition: all 0.2s ease;
}

.question-image-dropzone:hover {
  border-color: #3b82f6;
  background-color: #eff6ff;
}

.dropzone-icon {
  font-size: 1.25rem;
  color: #64748b;
}

.dropzone-label {
  font-size: 0.85rem;
  color: #64748b;
  font-weight: 500;
}

.hidden-file-input {
  display: none;
}

/* Graella d'Opcions de Resposta Kahoot */
.answers-editor-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.answers-header-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.answers-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: #1e293b;
}

.answers-hint {
  font-size: 0.8rem;
  color: #64748b;
}

.kahoot-editor-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.85rem;
}

.kahoot-editor-card {
  display: flex;
  align-items: center;
  border-radius: 0.5rem;
  padding: 0.5rem 0.75rem;
  box-shadow: 0 2px 4px rgba(0,0,0,0.12);
  gap: 0.5rem;
  position: relative;
}

.card-shape-badge {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 1.25rem;
  flex-shrink: 0;
}

.card-input-wrapper {
  flex: 1;
}

.answer-text-input {
  width: 100%;
  background-color: rgba(255, 255, 255, 0.92);
  border: 1px solid transparent;
  font-weight: 600;
  color: #0f172a;
}

.answer-text-input:focus {
  background-color: #ffffff;
}

.correct-toggle-btn {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.8);
  background-color: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #ffffff;
  font-size: 1.1rem;
  flex-shrink: 0;
  transition: all 0.15s ease;
}

.correct-toggle-btn.is-correct-active {
  background-color: #22c55e;
  border-color: #ffffff;
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.6);
}

.remove-answer-btn {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  border: none;
  background: rgba(0, 0, 0, 0.2);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 0.8rem;
  flex-shrink: 0;
  transition: background 0.15s ease;
}

.remove-answer-btn:hover {
  background: rgba(0, 0, 0, 0.4);
}

.add-answer-row {
  display: flex;
  justify-content: center;
  margin-top: 0.25rem;
}

.w-full {
  width: 100%;
}

@media (max-width: 768px) {
  .editor-workspace {
    flex-direction: column;
  }
  .questions-sidebar {
    width: 100%;
    max-height: 180px;
    border-right: none;
    border-bottom: 1px solid #e2e8f0;
  }
  .questions-scroll-list {
    flex-direction: row;
  }
  .question-thumb-card {
    min-width: 140px;
  }
  .kahoot-editor-grid {
    grid-template-columns: 1fr;
  }
}
</style>
