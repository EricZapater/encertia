<template>
  <div class="quiz-evaluation-container p-4">
    <div class="flex flex-column sm:flex-row justify-content-between align-items-start sm:align-items-center gap-3 mb-4">
      <div>
        <Button :label="$t('evaluations.actions.back')" icon="pi pi-arrow-left" class="p-button-text mb-2" @click="router.push('/evaluations')" />
        <h1 class="text-2xl font-bold m-0" v-if="evalData">{{ evalData.quizTitle }}</h1>
      </div>
    </div>

    <!-- Barra de filtres superior -->
    <div class="filter-bar flex flex-column sm:flex-row flex-wrap gap-3 mb-4 align-items-center bg-surface-50 dark:bg-surface-800 p-3 border-round">
      <div class="flex-1 min-w-12rem">
        <label class="block text-sm font-medium mb-1">{{ $t('evaluations.filters.group') }}</label>
        <Select
          v-model="selectedGroupId"
          :options="groupOptions"
          optionLabel="label"
          optionValue="value"
          :placeholder="$t('evaluations.filters.allGroups')"
          showClear
          class="w-full"
          @change="onGroupChange"
        />
      </div>

      <div class="flex-1 min-w-12rem">
        <label class="block text-sm font-medium mb-1">{{ $t('evaluations.filters.game') }}</label>
        <Select
          v-model="selectedQuizId"
          :options="quizOptions"
          optionLabel="label"
          optionValue="value"
          :placeholder="$t('evaluations.filters.selectGame')"
          showClear
          class="w-full"
          @change="onQuizChange"
        />
      </div>

      <div class="flex-1 min-w-12rem">
        <label class="block text-sm font-medium mb-1">{{ $t('evaluations.filters.date') }}</label>
        <DatePicker
          v-model="selectedDate"
          :placeholder="$t('evaluations.filters.datePlaceholder')"
          dateFormat="dd/mm/yy"
          showIcon
          showClear
          class="w-full"
        />
      </div>
    </div>

    <div v-if="store.isLoading" class="text-center p-4">
      <i class="pi pi-spin pi-spinner text-2xl"></i>
    </div>

    <div v-else-if="evalData">
      <!-- Secció A: Estadístiques Globals -->
      <Panel :header="$t('evaluations.panels.globalStats')" :toggleable="true" class="mb-4">
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
              <div class="flex flex-column gap-1">
                <div
                  v-for="item in slotProps.data.answerDistribution"
                  :key="item.answerId"
                  class="text-xs"
                >
                  <span :class="{ 'font-bold text-green-600': item.isCorrect }">
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
      <Panel :header="$t('evaluations.panels.studentResults')" :toggleable="true" class="mb-4">
        <DataTable :value="filteredStudents" class="p-datatable-sm" paginator :rows="10">
          <template #empty>{{ $t('common.noResults') }}</template>
          <Column field="studentName" :header="$t('evaluations.table.studentName')" sortable />
          <Column field="groupName" :header="$t('evaluations.table.group')" sortable>
            <template #body="slotProps">
              <span>{{ slotProps.data.groupName || '-' }}</span>
            </template>
          </Column>
          <Column field="matchesCount" :header="$t('evaluations.table.matches')" sortable align="center" />
          <Column :header="$t('evaluations.table.calculatedGrade')" sortable field="calculatedGrade" align="center">
            <template #body="slotProps">
              <span class="text-gray-600 font-bold">{{ slotProps.data.calculatedGrade.toFixed(2) }}</span>
            </template>
          </Column>
          <Column :header="$t('evaluations.table.finalGrade')" align="center">
            <template #body="slotProps">
              <Tag v-if="slotProps.data.isGraded" severity="success">
                {{ slotProps.data.finalGrade?.toFixed(2) }}
              </Tag>
              <span v-else class="text-gray-400 font-italic">{{ $t('evaluations.table.pending') }}</span>
            </template>
          </Column>
          <Column :header="$t('evaluations.table.actions')" align="center">
            <template #body="slotProps">
              <Button
                :label="slotProps.data.isGraded ? $t('evaluations.actions.editGrade') : $t('evaluations.actions.grade')"
                :icon="slotProps.data.isGraded ? 'pi pi-pencil' : 'pi pi-check-circle'"
                :class="slotProps.data.isGraded ? 'p-button-sm p-button-outlined' : 'p-button-sm p-button-success'"
                @click="navigateToStudentEvaluation(slotProps.data.studentId)"
              />
            </template>
          </Column>
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

const quizId = computed(() => route.params.quizId as string)
const evalData = computed(() => store.activeQuizEvaluation)

const filteredStudents = computed(() => {
  if (!evalData.value) return []
  let list = evalData.value.students
  if (selectedGroupId.value) {
    list = list.filter((s) => !s.groupId || s.groupId === selectedGroupId.value)
  }
  if (selectedDate.value) {
    const filterDate = new Date(selectedDate.value)
    list = list.filter((s) => {
      const dateStr = s.lastMatchAt || s.matchDate
      if (!dateStr) return false
      const matchDate = new Date(dateStr)
      return (
        matchDate.getFullYear() === filterDate.getFullYear() &&
        matchDate.getMonth() === filterDate.getMonth() &&
        matchDate.getDate() === filterDate.getDate()
      )
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

function navigateToStudentEvaluation(studentId: string) {
  router.push(`/evaluations/quizzes/${quizId.value}/students/${studentId}`)
}
</script>
