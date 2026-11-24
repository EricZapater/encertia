<template>
  <div class="quiz-evaluation-container">
    <div class="page-header">
      <div class="header-content">
        <Button
          :label="$t('evaluations.actions.back')"
          icon="pi pi-arrow-left"
          text
          class="back-btn p-button-text"
          @click="router.push('/evaluations')"
          data-testid="btn-back-to-evaluations"
        />
        <h1 class="page-title" v-if="evalData">{{ evalData.quizTitle }}</h1>
      </div>
    </div>

    <!-- Panell de filtres d'avaluació -->
    <div class="evaluation-filter-panel filter-panel-card">
      <div class="filter-panel-header">
        <div class="filter-title">
          <i class="pi pi-filter"></i>
          <span>{{ $t('evaluations.filters.title') }}</span>
        </div>
        <Button
          :label="$t('evaluations.filters.clear')"
          icon="pi pi-filter-slash"
          text
          class="filter-clear-btn p-button-text p-button-sm"
          @click="clearFilters"
          data-testid="btn-clear-quiz-filters"
        />
      </div>

      <div class="filter-grid">
        <div class="filter-item">
          <label class="filter-label">{{ $t('evaluations.filters.group') }}</label>
          <Select
            v-model="selectedGroupId"
            :options="groupOptions"
            optionLabel="label"
            optionValue="value"
            :placeholder="$t('evaluations.filters.allGroups')"
            showClear
            class="filter-select"
            @change="onGroupChange"
            data-testid="select-filter-group"
          />
        </div>

        <div class="filter-item">
          <label class="filter-label">{{ $t('evaluations.filters.game') }}</label>
          <Select
            v-model="selectedQuizId"
            :options="quizOptions"
            optionLabel="label"
            optionValue="value"
            :placeholder="$t('evaluations.filters.selectGame')"
            showClear
            class="filter-select"
            @change="onQuizChange"
            data-testid="select-filter-quiz"
          />
        </div>

        <div class="filter-item">
          <label class="filter-label">{{ $t('evaluations.filters.date') }}</label>
          <DatePicker
            v-model="selectedDate"
            :placeholder="$t('evaluations.filters.datePlaceholder')"
            dateFormat="dd/mm/yy"
            showIcon
            showClear
            class="filter-datepicker"
            data-testid="datepicker-filter-date"
          />
        </div>
      </div>
    </div>

    <div v-if="store.isLoading" class="loading-state">
      <i class="pi pi-spin pi-spinner loading-spinner"></i>
    </div>

    <div v-else-if="evalData" class="evaluation-content">
      <!-- Secció A: Estadístiques Globals -->
      <Panel :header="$t('evaluations.panels.globalStats')" :toggleable="true" class="stats-panel mb-4">
        <DataTable :value="evalData.stats" class="p-datatable-sm">
          <template #empty>{{ $t('common.noResults') }}</template>
          <Column field="questionIndex" :header="$t('evaluations.table.questionNumber')" style="width: 50px" />
          <Column field="questionText" :header="$t('evaluations.table.question')" />
          <Column :header="$t('evaluations.table.hitRate')" align="center">
            <template #body="slotProps">
              <Tag
                :severity="slotProps.data.hitRate >= 0.7 ? 'success' : slotProps.data.hitRate >= 0.5 ? 'warning' : 'danger'"
              >
                {{ (slotProps.data.hitRate * 100).toFixed(0) }}%
              </Tag>
            </template>
          </Column>
          <Column :header="$t('evaluations.table.avgResponseTime')" align="center">
            <template #body="slotProps">
              {{ (slotProps.data.avgResponseTimeMs / 1000).toFixed(1) }}s
            </template>
          </Column>
          <Column :header="$t('evaluations.table.answerDistribution')">
            <template #body="slotProps">
              <div class="answer-distribution-list">
                <div
                  v-for="item in slotProps.data.answerDistribution"
                  :key="item.answerId"
                  class="answer-distribution-item"
                >
                  <span :class="{ 'correct-answer': item.isCorrect }">
                    {{ item.answerText }}: {{ (item.percentage * 100).toFixed(0) }}% ({{ item.count }})
                  </span>
                </div>
              </div>
            </template>
          </Column>
          <Column field="noAnswerCount" :header="$t('evaluations.table.noAnswer')" align="center" />
        </DataTable>
      </Panel>

      <!-- Secció B: Taula d'Alumnes -->
      <Panel :header="$t('evaluations.panels.studentResults')" :toggleable="true" class="results-panel mb-4">
        <DataTable
          v-model:expandedRows="expandedRows"
          :value="filteredStudents"
          dataKey="studentId"
          class="p-datatable-sm"
          paginator
          :rows="10"
          @row-expand="onRowExpand"
        >
          <template #empty>{{ $t('common.noResults') }}</template>
          <Column expander style="width: 3.5rem" />
          <Column field="studentName" :header="$t('evaluations.table.studentName')" sortable />
          <Column field="groupName" :header="$t('evaluations.table.group')" sortable>
            <template #body="slotProps">
              <span>{{ slotProps.data.groupName || '-' }}</span>
            </template>
          </Column>
          <Column field="matchesCount" :header="$t('evaluations.table.matches')" sortable align="center" />
          <Column :header="$t('evaluations.table.calculatedGrade')" sortable field="calculatedGrade" align="center">
            <template #body="slotProps">
              <span class="calculated-grade-text">{{ slotProps.data.calculatedGrade.toFixed(2) }}</span>
            </template>
          </Column>
          <Column :header="$t('evaluations.table.finalGrade')" align="center">
            <template #body="slotProps">
              <Tag v-if="slotProps.data.isGraded" severity="success">
                {{ slotProps.data.finalGrade?.toFixed(2) }}
              </Tag>
              <span v-else class="pending-grade-text">{{ $t('evaluations.table.pending') }}</span>
            </template>
          </Column>
          <Column :header="$t('evaluations.table.answers') || 'Respostes (OK/KO)'" align="center" style="min-width: 180px">
            <template #body="slotProps">
              <div class="answers-col-container">
                <Tag
                  v-if="slotProps.data.answers && slotProps.data.answers.length > 0"
                  :severity="slotProps.data.answers.filter((a: any) => a.isCorrect).length === slotProps.data.answers.length ? 'success' : 'info'"
                  class="answers-count-badge"
                >
                  {{ slotProps.data.answers.filter((a: any) => a.isCorrect).length }}/{{ slotProps.data.answers.length }} OK
                </Tag>
                <Button
                  :label="expandedRows[slotProps.data.studentId] ? 'Amagar' : ($t('evaluations.actions.viewAnswers') || 'Veure Respostes')"
                  :icon="expandedRows[slotProps.data.studentId] ? 'pi pi-chevron-up' : 'pi pi-list-check'"
                  size="small"
                  severity="info"
                  outlined
                  @click="toggleStudentRow(slotProps.data)"
                  :data-testid="`btn-toggle-answers-${slotProps.data.studentId}`"
                />
              </div>
            </template>
          </Column>
          <Column :header="$t('evaluations.table.actions')" align="center">
            <template #body="slotProps">
              <Button
                :label="slotProps.data.isGraded ? $t('evaluations.actions.editGrade') : $t('evaluations.actions.grade')"
                :icon="slotProps.data.isGraded ? 'pi pi-pencil' : 'pi pi-check-circle'"
                :class="slotProps.data.isGraded ? 'p-button-sm p-button-outlined' : 'p-button-sm p-button-success'"
                @click="navigateToStudentEvaluation(slotProps.data.studentId)"
                :data-testid="`btn-grade-${slotProps.data.studentId}`"
              />
            </template>
          </Column>

          <!-- Fila expandible amb detall de respostes OK/KO -->
          <template #expansion="slotProps">
            <div class="student-expansion-container" :data-testid="`student-answers-${slotProps.data.studentId}`">
              <div v-if="loadingStudentMap[slotProps.data.studentId] && getStudentMatches(slotProps.data).length === 0" class="expansion-loading">
                <i class="pi pi-spin pi-spinner mr-2"></i> Carregant respostes de l'alumne...
              </div>
              <div v-else-if="errorStudentMap[slotProps.data.studentId] && getStudentMatches(slotProps.data).length === 0" class="expansion-error">
                <span><i class="pi pi-exclamation-triangle mr-2"></i> Error: {{ errorStudentMap[slotProps.data.studentId] }}</span>
                <Button label="Reintentar" icon="pi pi-refresh" size="small" severity="danger" text @click="loadStudentAnswers(slotProps.data.studentId)" />
              </div>
              <div v-else-if="getStudentMatches(slotProps.data).length > 0" class="student-matches-wrapper">
                <div
                  v-for="(match, mIdx) in getStudentMatches(slotProps.data)"
                  :key="match.matchId"
                  class="match-evaluation-card"
                >
                  <div class="match-card-header">
                    <div class="match-meta">
                      <span class="match-badge">Partida {{ getStudentMatches(slotProps.data).length - mIdx }}</span>
                      <span class="match-date"><i class="pi pi-calendar mr-1"></i>{{ formatDate(match.matchDate) }}</span>
                    </div>
                    <div class="match-stats">
                      <Tag :severity="match.score === match.totalQuestions ? 'success' : match.score >= match.totalQuestions / 2 ? 'info' : 'warn'">
                        {{ match.score }} / {{ match.totalQuestions }} encerts ({{ Math.round((match.score / (match.totalQuestions || 1)) * 100) }}%)
                      </Tag>
                    </div>
                  </div>

                  <DataTable :value="match.answers" class="p-datatable-sm answers-subtable">
                    <Column field="questionIndex" header="#" style="width: 50px" align="center">
                      <template #body="ansProps">
                        <strong>P{{ ansProps.data.questionIndex + 1 }}</strong>
                      </template>
                    </Column>
                    <Column field="questionText" header="Pregunta">
                      <template #body="ansProps">
                        <FormattedQuestionText :text="ansProps.data.questionText" />
                      </template>
                    </Column>
                    <Column header="Resultat" align="center" style="width: 120px">
                      <template #body="ansProps">
                        <Tag
                          :severity="ansProps.data.isCorrect ? 'success' : 'danger'"
                          class="ok-ko-tag"
                          :data-testid="ansProps.data.isCorrect ? 'badge-answer-ok' : 'badge-answer-ko'"
                        >
                          <i :class="ansProps.data.isCorrect ? 'pi pi-check mr-1' : 'pi pi-times mr-1'"></i>
                          {{ ansProps.data.isCorrect ? 'OK' : 'KO' }}
                        </Tag>
                      </template>
                    </Column>
                    <Column header="Temps de Resposta" align="center" style="width: 140px">
                      <template #body="ansProps">
                        <span>{{ (ansProps.data.responseTimeMs / 1000).toFixed(1) }}s</span>
                      </template>
                    </Column>
                  </DataTable>
                </div>
              </div>
              <div v-else class="p-3 text-muted">
                Aquest alumne no té cap partida registrada per a aquest qüestionari.
              </div>
            </div>
          </template>
        </DataTable>
      </Panel>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useEvaluationStore } from '../store'
import { useGroupStore } from '@/modules/groups/store'
import { useQuizStore } from '@/modules/quizzes/store'
import { getStudentEvaluation } from '../api'
import type { StudentEvaluationDetail, StudentEvaluationSummary, StudentMatchResult } from '../types'
import FormattedQuestionText from '@/components/FormattedQuestionText.vue'
import Button from 'primevue/button'
import Panel from 'primevue/panel'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Select from 'primevue/select'
import DatePicker from 'primevue/datepicker'
import { useToast } from 'primevue/usetoast'

const route = useRoute()
const router = useRouter()
const store = useEvaluationStore()
const groupStore = useGroupStore()
const quizStore = useQuizStore()
const toast = useToast()

const selectedGroupId = ref<string | null>(null)
const selectedQuizId = ref<string | null>((route.params.quizId as string) || null)
const selectedDate = ref<Date | null>(null)

// Estat d'expansió de files i memòria cau de respostes dels alumnes
const expandedRows = ref<Record<string, boolean>>({})
const studentDetailsCache = ref<Record<string, StudentEvaluationDetail>>({})
const loadingStudentMap = ref<Record<string, boolean>>({})
const errorStudentMap = ref<Record<string, string>>({})

const quizId = computed(() => (route.params.quizId as string) || selectedQuizId.value || evalData.value?.quizId || '')

async function loadStudentAnswers(studentId: string): Promise<StudentEvaluationDetail | null> {
  if (studentDetailsCache.value[studentId]) {
    return studentDetailsCache.value[studentId]
  }
  loadingStudentMap.value[studentId] = true
  delete errorStudentMap.value[studentId]
  try {
    const qId = quizId.value
    const detail = await getStudentEvaluation(qId, studentId)
    studentDetailsCache.value[studentId] = detail
    return detail
  } catch (err: any) {
    console.error('Error carregant respostes de l alumne:', err)
    errorStudentMap.value[studentId] = err?.message || 'Error carregant dades'
    return null
  } finally {
    loadingStudentMap.value[studentId] = false
  }
}

function getStudentMatches(student: StudentEvaluationSummary): StudentMatchResult[] {
  const cached = studentDetailsCache.value[student.studentId]
  if (cached?.matches && cached.matches.length > 0) {
    return cached.matches
  }
  if (student.answers && student.answers.length > 0) {
    const total = student.answers.length
    const score = student.answers.filter((a) => a.isCorrect).length
    return [
      {
        matchId: `m-cached-${student.studentId}`,
        matchDate: student.lastMatchAt || student.matchDate || new Date().toISOString(),
        score,
        totalQuestions: total,
        answers: student.answers
      }
    ]
  }
  return []
}

async function onRowExpand(event: any) {
  const student = event.data as StudentEvaluationSummary
  if (student?.studentId) {
    await loadStudentAnswers(student.studentId)
  }
}

async function toggleStudentRow(student: StudentEvaluationSummary) {
  if (expandedRows.value[student.studentId]) {
    const copy = { ...expandedRows.value }
    delete copy[student.studentId]
    expandedRows.value = copy
  } else {
    expandedRows.value = { ...expandedRows.value, [student.studentId]: true }
    await loadStudentAnswers(student.studentId)
  }
}

function formatDate(d: string | Date | undefined): string {
  if (!d) return ''
  const date = typeof d === 'string' ? new Date(d) : d
  if (isNaN(date.getTime())) return ''
  return date.toLocaleDateString('ca-ES', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const groupOptions = computed(() => {
  return groupStore.groups.map((g) => ({
    label: g.academicYear ? `${g.name} (${g.academicYear})` : g.name,
    value: g.id
  }))
})

const quizOptions = computed(() => {
  const map = new Map<string, string>()
  quizStore.quizzes.forEach((q) => map.set(q.id, q.title))
  store.evaluationsList.forEach((e) => {
    if (!map.has(e.quizId)) {
      map.set(e.quizId, e.quizTitle)
    }
  })
  if (evalData.value && !map.has(evalData.value.quizId)) {
    map.set(evalData.value.quizId, evalData.value.quizTitle)
  }
  return Array.from(map.entries()).map(([id, title]) => ({
    label: title,
    value: id
  }))
})

const evalData = computed(() => store.activeQuizEvaluation)

function toLocalDateString(d: Date | string | null | undefined): string {
  if (!d) return ''
  const date = typeof d === 'string' ? new Date(d) : d
  if (isNaN(date.getTime())) return ''
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const filteredStudents = computed(() => {
  if (!evalData.value) return []
  let list = evalData.value.students
  if (selectedGroupId.value) {
    list = list.filter((s) => !s.groupId || s.groupId === selectedGroupId.value)
  }
  if (selectedDate.value) {
    list = list.filter((s) => {
      return toLocalDateString(s.lastMatchAt || s.matchDate) === toLocalDateString(selectedDate.value)
    })
  }
  return list
})

watch(
  () => route.params.quizId,
  (newQuizId) => {
    if (newQuizId && typeof newQuizId === 'string') {
      selectedQuizId.value = newQuizId
      store.fetchQuizEvaluation(newQuizId, selectedGroupId.value || undefined)
    }
  }
)

watch(
  () => store.error,
  (err) => {
    if (err) {
      toast.add({ severity: 'error', summary: 'Error', detail: err, life: 4000 })
      store.clearError()
    }
  }
)

onMounted(() => {
  groupStore.fetchGroups()
  quizStore.fetchQuizzes()
  store.fetchEvaluationsList()
  if (quizId.value) {
    store.fetchQuizEvaluation(quizId.value, selectedGroupId.value || undefined)
  }
})

function onGroupChange() {
  if (quizId.value) {
    store.fetchQuizEvaluation(quizId.value, selectedGroupId.value || undefined)
  }
}

function onQuizChange() {
  if (selectedQuizId.value && selectedQuizId.value !== quizId.value) {
    router.push(`/evaluations/quizzes/${selectedQuizId.value}`)
  }
}

function clearFilters() {
  selectedGroupId.value = null
  selectedDate.value = null
  if (quizId.value) {
    store.fetchQuizEvaluation(quizId.value)
  }
}

function navigateToStudentEvaluation(studentId: string) {
  router.push(`/evaluations/quizzes/${quizId.value}/students/${studentId}`)
}
</script>

<style scoped>
.quiz-evaluation-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1.5rem 1rem;
}

.page-header {
  margin-bottom: 1.5rem;
}

.header-content {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.5rem;
}

.back-btn {
  padding-left: 0;
}

.page-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

/* Panell de filtres */
.evaluation-filter-panel,
.filter-panel-card {
  background-color: #ffffff;
  border-radius: 0.75rem;
  border: 1px solid #e2e8f0;
  padding: 1.25rem;
  margin-bottom: 1.5rem;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
}

.filter-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid #f1f5f9;
}

.filter-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  font-size: 1rem;
  color: #1e293b;
}

.filter-title i {
  color: #3b82f6;
  font-size: 1rem;
}

.filter-clear-btn {
  font-size: 0.85rem;
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
  align-items: end;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.filter-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #334155;
}

.filter-select,
.filter-datepicker {
  width: 100%;
}

.loading-state {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 4rem 1rem;
}

.loading-spinner {
  font-size: 2rem;
  color: #3b82f6;
}

.evaluation-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.stats-panel,
.results-panel {
  margin-bottom: 0;
}

.answer-distribution-list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.answer-distribution-item {
  font-size: 0.75rem;
}

.correct-answer {
  font-weight: 700;
  color: #16a34a;
}

.calculated-grade-text {
  font-weight: 700;
  color: #475569;
}

.pending-grade-text {
  color: #94a3b8;
  font-style: italic;
}

/* Expansió de Respostes d'Alumne */
.student-expansion-container {
  padding: 1rem;
  background-color: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  margin: 0.5rem 0;
}

.expansion-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  color: #64748b;
  font-weight: 500;
}

.expansion-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  color: #ef4444;
  font-weight: 500;
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
}

.answers-col-container {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.answers-count-badge {
  font-size: 0.75rem;
  font-weight: 700;
}

.match-evaluation-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 0.75rem 1rem;
  margin-bottom: 0.75rem;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.match-evaluation-card:last-child {
  margin-bottom: 0;
}

.match-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 0.6rem;
  margin-bottom: 0.6rem;
  border-bottom: 1px solid #f1f5f9;
}

.match-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.match-badge {
  font-weight: 700;
  font-size: 0.85rem;
  color: #1e293b;
}

.match-date {
  font-size: 0.8rem;
  color: #64748b;
}

.ok-ko-tag {
  font-weight: 800;
  font-size: 0.85rem;
  letter-spacing: 0.05em;
  padding: 0.2rem 0.5rem;
}

.answers-subtable {
  background: #ffffff;
}

/* Suport Mode Fosc */
:global(.dark-mode) .page-title {
  color: #f8fafc;
}

:global(.dark-mode) .evaluation-filter-panel,
:global(.dark-mode) .filter-panel-card {
  background-color: #1e293b;
  border-color: #334155;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.3);
}

:global(.dark-mode) .filter-panel-header {
  border-bottom-color: #334155;
}

:global(.dark-mode) .filter-title {
  color: #f1f5f9;
}

:global(.dark-mode) .filter-title i {
  color: #60a5fa;
}

:global(.dark-mode) .filter-label {
  color: #94a3b8;
}

:global(.dark-mode) .calculated-grade-text {
  color: #cbd5e1;
}

:global(.dark-mode) .pending-grade-text {
  color: #64748b;
}

:global(.dark-mode) .correct-answer {
  color: #4ade80;
}

:global(.dark-mode) .student-expansion-container {
  background-color: #0f172a;
  border-color: #334155;
}

:global(.dark-mode) .expansion-error {
  background-color: #450a0a;
  border-color: #7f1d1d;
  color: #f87171;
}

:global(.dark-mode) .match-evaluation-card {
  background-color: #1e293b;
  border-color: #334155;
}

:global(.dark-mode) .match-card-header {
  border-bottom-color: #334155;
}

:global(.dark-mode) .match-badge {
  color: #f1f5f9;
}

:global(.dark-mode) .match-date {
  color: #94a3b8;
}

:global(.dark-mode) .answers-subtable {
  background-color: #1e293b;
}

@media (prefers-color-scheme: dark) {
  :global(:not(.light-mode)) .page-title {
    color: #f8fafc;
  }

  :global(:not(.light-mode)) .evaluation-filter-panel,
  :global(:not(.light-mode)) .filter-panel-card {
    background-color: #1e293b;
    border-color: #334155;
  }

  :global(:not(.light-mode)) .filter-panel-header {
    border-bottom-color: #334155;
  }

  :global(:not(.light-mode)) .filter-title {
    color: #f1f5f9;
  }

  :global(:not(.light-mode)) .filter-title i {
    color: #60a5fa;
  }

  :global(:not(.light-mode)) .filter-label {
    color: #94a3b8;
  }

  :global(:not(.light-mode)) .calculated-grade-text {
    color: #cbd5e1;
  }

  :global(:not(.light-mode)) .pending-grade-text {
    color: #64748b;
  }

  :global(:not(.light-mode)) .correct-answer {
    color: #4ade80;
  }
}
</style>
