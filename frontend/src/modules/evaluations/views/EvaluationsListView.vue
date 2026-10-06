<template>
  <div class="evaluations-list-container p-4">
    <div class="flex flex-column sm:flex-row justify-content-between align-items-start sm:align-items-center gap-3 mb-4">
      <h1 class="text-2xl font-bold m-0">{{ $t('evaluations.title') }}</h1>
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
          :placeholder="$t('evaluations.filters.allGames')"
          showClear
          class="w-full"
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

    <DataTable
      :value="filteredEvaluations"
      :loading="store.isLoading"
      responsiveLayout="scroll"
      class="p-datatable-sm"
      paginator
      :rows="10"
    >
      <template #empty>{{ $t('common.noResults') }}</template>
      <Column field="quizTitle" :header="$t('evaluations.table.quizTitle')" sortable />
      <Column field="groupName" :header="$t('evaluations.table.group')" sortable>
        <template #body="slotProps">
          <span>{{ slotProps.data.groupName || '-' }}</span>
        </template>
      </Column>
      <Column field="totalMatches" :header="$t('evaluations.table.totalMatches')" sortable align="center" />
      <Column field="totalStudents" :header="$t('evaluations.table.totalStudents')" sortable align="center" />
      <Column :header="$t('evaluations.table.gradedCount')" align="center">
        <template #body="slotProps">
          <span>{{ slotProps.data.gradedCount }} / {{ slotProps.data.totalStudents }}</span>
        </template>
      </Column>
      <Column :header="$t('evaluations.table.lastMatchAt')" sortable field="lastMatchAt">
        <template #body="slotProps">
          {{ formatDate(slotProps.data.lastMatchAt) }}
        </template>
      </Column>
      <Column :header="$t('evaluations.table.actions')" align="center">
        <template #body="slotProps">
          <Button
            :label="$t('evaluations.actions.viewEvaluation')"
            icon="pi pi-eye"
            class="p-button-sm p-button-outlined"
            @click="navigateToQuizEvaluation(slotProps.data.quizId)"
          />
        </template>
      </Column>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useEvaluationStore } from '../store'
import { useGroupStore } from '@/modules/groups/store'
import { useQuizStore } from '@/modules/quizzes/store'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Select from 'primevue/select'
import DatePicker from 'primevue/datepicker'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const store = useEvaluationStore()
const groupStore = useGroupStore()
const quizStore = useQuizStore()
const toast = useToast()

const selectedGroupId = ref<string | null>(null)
const selectedQuizId = ref<string | null>(null)
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
  return Array.from(map.entries()).map(([id, title]) => ({
    label: title,
    value: id
  }))
})

const filteredEvaluations = computed(() => {
  return store.evaluationsList.filter((item) => {
    if (selectedGroupId.value) {
      if (item.groupId && item.groupId !== selectedGroupId.value) {
        return false
      }
    }
    if (selectedQuizId.value) {
      if (item.quizId !== selectedQuizId.value) {
        return false
      }
    }
    if (selectedDate.value) {
      if (!item.lastMatchAt) return false
      const matchDate = new Date(item.lastMatchAt)
      const filterDate = new Date(selectedDate.value)
      const isSameDay =
        matchDate.getFullYear() === filterDate.getFullYear() &&
        matchDate.getMonth() === filterDate.getMonth() &&
        matchDate.getDate() === filterDate.getDate()
      if (!isSameDay) return false
    }
    return true
  })
})

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
})

function onGroupChange() {
  store.fetchEvaluationsList(selectedGroupId.value || undefined)
}

function navigateToQuizEvaluation(quizId: string) {
  router.push(`/evaluations/quizzes/${quizId}`)
}

function formatDate(isoStr: string): string {
  if (!isoStr) return '-'
  const d = new Date(isoStr)
  return d.toLocaleString()
}
</script>
