<template>
  <div class="quiz-evaluation-container p-4">
    <div class="mb-4">
      <Button label="Tornar" icon="pi pi-arrow-left" class="p-button-text mb-2" @click="router.push('/evaluations')" />
      <h1 class="text-2xl font-bold m-0" v-if="evalData">{{ evalData.quizTitle }}</h1>
    </div>

    <Message v-if="store.error" severity="error" :closable="true" @close="store.clearError()">
      {{ store.error }}
    </Message>

    <div v-if="store.isLoading" class="text-center p-4">
      <i class="pi pi-spin pi-spinner text-2xl"></i>
    </div>

    <div v-else-if="evalData">
      <!-- Secció A: Estadístiques Globals -->
      <Card class="mb-4">
        <template #title>Estadístiques Globals per Pregunta</template>
        <template #content>
          <DataTable :value="evalData.stats" class="p-datatable-sm">
            <Column field="questionIndex" header="#" style="width: 50px" />
            <Column field="questionText" header="Pregunta" />
            <Column header="Taxa d'Encert" align="center">
              <template #body="slotProps">
                <Tag
                  :severity="slotProps.data.hitRate >= 0.7 ? 'success' : slotProps.data.hitRate >= 0.5 ? 'warning' : 'danger'"
                >
                  {{ (slotProps.data.hitRate * 100).toFixed(0) }}%
                </Tag>
              </template>
            </Column>
            <Column header="Temps Mitjà" align="center">
              <template #body="slotProps">
                {{ (slotProps.data.avgResponseTimeMs / 1000).toFixed(1) }}s
              </template>
            </Column>
            <Column header="Distribució de Respostes">
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
            <Column field="noAnswerCount" header="Sense Resposta" align="center" />
          </DataTable>
        </template>
      </Card>

      <!-- Secció B: Taula d'Alumnes -->
      <Card>
        <template #title>Resultats dels Alumnes</template>
        <template #content>
          <DataTable :value="evalData.students" class="p-datatable-sm" paginator :rows="10">
            <Column field="studentName" header="Nom de l'Alumne" sortable />
            <Column field="matchesCount" header="Partides" sortable align="center" />
            <Column header="Nota Calculada (Última Partida)" sortable field="calculatedGrade" align="center">
              <template #body="slotProps">
                <span class="text-gray-600 font-bold">{{ slotProps.data.calculatedGrade.toFixed(2) }}</span>
              </template>
            </Column>
            <Column header="Nota Definitiva" align="center">
              <template #body="slotProps">
                <Tag v-if="slotProps.data.isGraded" severity="success">
                  {{ slotProps.data.finalGrade?.toFixed(2) }}
                </Tag>
                <span v-else class="text-gray-400 font-italic">Pendent</span>
              </template>
            </Column>
            <Column header="Accions" align="center">
              <template #body="slotProps">
                <Button
                  :label="slotProps.data.isGraded ? 'Editar Nota' : 'Qualificar'"
                  :icon="slotProps.data.isGraded ? 'pi pi-pencil' : 'pi pi-check-circle'"
                  :class="slotProps.data.isGraded ? 'p-button-sm p-button-outlined' : 'p-button-sm p-button-success'"
                  @click="navigateToStudentEvaluation(slotProps.data.studentId)"
                />
              </template>
            </Column>
          </DataTable>
        </template>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useEvaluationStore } from '../store'
import Button from 'primevue/button'
import Card from 'primevue/card'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Message from 'primevue/message'

const route = useRoute()
const router = useRouter()
const store = useEvaluationStore()

const quizId = computed(() => route.params.quizId as string)
const evalData = computed(() => store.activeQuizEvaluation)

onMounted(() => {
  if (quizId.value) {
    store.fetchQuizEvaluation(quizId.value)
  }
})

function navigateToStudentEvaluation(studentId: string) {
  router.push(`/evaluations/quizzes/${quizId.value}/students/${studentId}`)
}
</script>
