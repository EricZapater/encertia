<template>
  <div class="student-evaluation-container p-4">
    <div class="mb-4">
      <Button label="Tornar al Quiz" icon="pi pi-arrow-left" class="p-button-text mb-2" @click="router.push(`/evaluations/quizzes/${quizId}`)" />
      <h1 class="text-2xl font-bold m-0" v-if="studentData">Avaluació de {{ studentData.studentName }}</h1>
    </div>

    <div v-if="store.isLoading" class="text-center p-4">
      <i class="pi pi-spin pi-spinner text-2xl"></i>
    </div>

    <div v-else-if="studentData">
      <!-- Formulari de Qualificació -->
      <Card class="mb-4">
        <template #title>Qualificació Manual</template>
        <template #content>
          <div class="flex flex-wrap align-items-center gap-3">
            <div>
              <span class="text-sm text-gray-600 block mb-1">Nota Calculada (Última Partida)</span>
              <span class="text-xl font-bold text-gray-800">{{ studentData.calculatedGrade.toFixed(2) }}</span>
            </div>

            <div class="flex-1 min-w-12rem">
              <label class="text-sm font-semibold block mb-1">Nota Definitiva (0.00 - 10.00)</label>
              <InputNumber
                v-model="gradeInput"
                :min="0"
                :max="10"
                :minFractionDigits="2"
                :maxFractionDigits="2"
                placeholder="Ex. 7.50"
                class="w-full"
              />
            </div>

            <div class="align-self-end">
              <Button
                label="Desar Nota"
                icon="pi pi-save"
                class="p-button-success"
                :loading="store.isSavingGrade"
                @click="onSaveGrade"
              />
            </div>
          </div>
        </template>
      </Card>

      <!-- Historial de Partides -->
      <h2 class="text-xl font-bold mb-3">Historial de Partides</h2>
      <Card v-for="match in studentData.matches" :key="match.matchId" class="mb-3">
        <template #title>
          <div class="flex justify-content-between text-base">
            <span>Partida del {{ formatDate(match.matchDate) }}</span>
            <span>Puntuació: {{ match.score }} / {{ match.totalQuestions }}</span>
          </div>
        </template>
        <template #content>
          <DataTable :value="match.answers" class="p-datatable-sm">
            <Column field="questionIndex" header="#" style="width: 40px" />
            <Column field="questionText" header="Pregunta" />
            <Column header="Resultat" align="center">
              <template #body="slotProps">
                <i v-if="slotProps.data.isCorrect" class="pi pi-check text-green-500 font-bold"></i>
                <i v-else class="pi pi-times text-red-500 font-bold"></i>
              </template>
            </Column>
            <Column header="Temps de Resposta" align="center">
              <template #body="slotProps">
                {{ (slotProps.data.responseTimeMs / 1000).toFixed(1) }}s
              </template>
            </Column>
          </DataTable>
        </template>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useEvaluationStore } from '../store'
import Button from 'primevue/button'
import Card from 'primevue/card'
import InputNumber from 'primevue/inputnumber'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useToast } from 'primevue/usetoast'

const route = useRoute()
const router = useRouter()
const store = useEvaluationStore()
const toast = useToast()

const quizId = computed(() => route.params.quizId as string)
const studentId = computed(() => route.params.studentId as string)
const studentData = computed(() => store.activeStudentEvaluation)

const gradeInput = ref<number | null>(null)

watch(
  () => store.error,
  (err) => {
    if (err) {
      toast.add({ severity: 'error', summary: 'Error', detail: err, life: 4000 })
      store.clearError()
    }
  }
)

onMounted(async () => {
  if (quizId.value && studentId.value) {
    await store.fetchStudentEvaluation(quizId.value, studentId.value)
    initGradeInput()
  }
})

watch(studentData, () => {
  initGradeInput()
})

function initGradeInput() {
  if (studentData.value) {
    gradeInput.value = studentData.value.finalGrade ?? studentData.value.calculatedGrade
  }
}

async function onSaveGrade() {
  if (gradeInput.value === null || gradeInput.value === undefined) return
  try {
    const res = await store.saveStudentGrade(quizId.value, studentId.value, gradeInput.value)
    const msg = `Nota de ${res.finalGrade.toFixed(2)} desada correctament.`
    toast.add({
      severity: 'success',
      summary: 'Nota Desada',
      detail: msg,
      life: 4000
    })
  } catch (err) {
    // Error feedback handled via store.error watch
  }
}

function formatDate(isoStr: string): string {
  if (!isoStr) return '-'
  return new Date(isoStr).toLocaleString()
}
</script>
