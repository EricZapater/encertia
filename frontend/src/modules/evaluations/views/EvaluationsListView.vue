<template>
  <div class="evaluations-list-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('evaluations.title') }}</h1>
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
          data-testid="btn-clear-filters"
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
            :placeholder="$t('evaluations.filters.allGames')"
            showClear
            class="filter-select"
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

    <div class="table-card">
      <DataTable
        :value="filteredEvaluations"
        :loading="store.isLoading"
        responsiveLayout="scroll"
        class="p-datatable-sm"
        paginator
        :rows="10"
      >
        <template #empty>
          <div class="empty-state">
            <i class="pi pi-chart-bar empty-icon"></i>
            <p>{{ $t('common.noResults') }}</p>
          </div>
        </template>
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

function toLocalDateString(d: Date | string | null | undefined): string {
  if (!d) return ''
  const date = typeof d === 'string' ? new Date(d) : d
  if (isNaN(date.getTime())) return ''
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

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
      return toLocalDateString(item.lastMatchAt) === toLocalDateString(selectedDate.value)
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

function clearFilters() {
  selectedGroupId.value = null
  selectedQuizId.value = null
  selectedDate.value = null
  store.fetchEvaluationsList()
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

<style scoped>
.evaluations-list-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1.5rem 1rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
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

.table-card {
  background-color: #ffffff;
  border-radius: 0.75rem;
  border: 1px solid #e2e8f0;
  overflow: hidden;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 3rem 1rem;
  color: #64748b;
  gap: 0.5rem;
}

.empty-icon {
  font-size: 2.5rem;
  color: #cbd5e1;
  margin-bottom: 0.25rem;
}

/* Suport Mode Fosc */
:global(.dark-mode) .page-title {
  color: #f8fafc;
}

:global(.dark-mode) .evaluation-filter-panel,
:global(.dark-mode) .filter-panel-card,
:global(.dark-mode) .table-card {
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

:global(.dark-mode) .empty-state {
  color: #94a3b8;
}

:global(.dark-mode) .empty-icon {
  color: #475569;
}

@media (prefers-color-scheme: dark) {
  :global(:not(.light-mode)) .page-title {
    color: #f8fafc;
  }

  :global(:not(.light-mode)) .evaluation-filter-panel,
  :global(:not(.light-mode)) .filter-panel-card,
  :global(:not(.light-mode)) .table-card {
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

  :global(:not(.light-mode)) .empty-state {
    color: #94a3b8;
  }

  :global(:not(.light-mode)) .empty-icon {
    color: #475569;
  }
}
</style>
