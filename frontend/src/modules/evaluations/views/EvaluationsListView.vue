<template>
  <div class="evaluations-list-container p-4">
    <div class="flex justify-content-between align-items-center mb-4">
      <h1 class="text-2xl font-bold m-0">{{ $t('evaluations.title') }}</h1>
    </div>

    <DataTable
      :value="store.evaluationsList"
      :loading="store.isLoading"
      responsiveLayout="scroll"
      class="p-datatable-sm"
      paginator
      :rows="10"
    >
      <Column field="quizTitle" header="Qüestionari" sortable />
      <Column field="totalMatches" header="Partides Jugades" sortable align="center" />
      <Column field="totalStudents" header="Alumnes Participants" sortable align="center" />
      <Column header="Avaluats / Total" align="center">
        <template #body="slotProps">
          <span>{{ slotProps.data.gradedCount }} / {{ slotProps.data.totalStudents }}</span>
        </template>
      </Column>
      <Column header="Última Partida" sortable field="lastMatchAt">
        <template #body="slotProps">
          {{ formatDate(slotProps.data.lastMatchAt) }}
        </template>
      </Column>
      <Column header="Accions" align="center">
        <template #body="slotProps">
          <Button
            label="Veure Avaluació"
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
import { onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useEvaluationStore } from '../store'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const store = useEvaluationStore()
const toast = useToast()

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
  store.fetchEvaluationsList()
})

function navigateToQuizEvaluation(quizId: string) {
  router.push(`/evaluations/quizzes/${quizId}`)
}

function formatDate(isoStr: string): string {
  if (!isoStr) return '-'
  const d = new Date(isoStr)
  return d.toLocaleString()
}
</script>
