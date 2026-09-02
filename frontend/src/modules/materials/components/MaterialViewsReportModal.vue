<script setup lang="ts">
import { watch } from 'vue'
import { useMaterialStore } from '../store'

import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'

const props = defineProps<{
  visible: boolean
  materialId: string
  materialTitle?: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const materialStore = useMaterialStore()

watch(
  () => [props.visible, props.materialId],
  async ([visible, id]) => {
    if (visible && id) {
      try {
        await materialStore.fetchViewsReport(id as string)
      } catch (_e) {
        // Handled in store
      }
    }
  }
)

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return d.toLocaleString('ca-ES', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch (_e) {
    return dateStr
  }
}

function handleClose() {
  emit('update:visible', false)
}
</script>

<template>
  <Dialog
    :visible="visible"
    :header="`Informe d'Accessos — ${materialTitle || 'Material'}`"
    :modal="true"
    :style="{ width: '700px' }"
    @update:visible="emit('update:visible', $event)"
  >
    <div v-if="materialStore.isLoading" class="loading-state">
      <i class="pi pi-spin pi-spinner spinner-icon"></i>
      <p>Carregant l'informe d'accessos dels alumnes...</p>
    </div>

    <div v-else-if="materialStore.currentViewsReport" class="report-container">
      <!-- Summary Cards -->
      <div class="metrics-grid">
        <div class="metric-card">
          <span class="metric-value">{{ materialStore.currentViewsReport.totalViews }}</span>
          <span class="metric-label">Total de Visualitzacions</span>
        </div>
        <div class="metric-card">
          <span class="metric-value">{{ materialStore.currentViewsReport.totalStudentsViewed }}</span>
          <span class="metric-label">Alumnes Únics</span>
        </div>
      </div>

      <!-- Students Table -->
      <div class="table-section">
        <h3 class="table-title">Detall d'Accés per Alumne</h3>

        <DataTable
          :value="materialStore.currentViewsReport.studentViews || []"
          responsiveLayout="scroll"
          class="p-datatable-sm"
          paginator
          :rows="5"
        >
          <Column field="studentName" header="Alumne" sortable>
            <template #body="slotProps">
              <div class="font-medium text-slate-800">
                {{ slotProps.data.studentName }}
              </div>
            </template>
          </Column>

          <Column field="studentEmail" header="Email">
            <template #body="slotProps">
              <span class="text-slate-500 text-sm">{{ slotProps.data.studentEmail }}</span>
            </template>
          </Column>

          <Column field="viewCount" header="Lectures" sortable align="center">
            <template #body="slotProps">
              <span class="font-bold text-indigo-600">{{ slotProps.data.viewCount }}</span>
            </template>
          </Column>

          <Column field="lastViewedAt" header="Últim accés" sortable>
            <template #body="slotProps">
              <span class="text-slate-600 text-sm">{{ formatDate(slotProps.data.lastViewedAt) }}</span>
            </template>
          </Column>

          <template #empty>
            <div class="empty-table-msg">
              Cap alumne ha accedit encara a aquest material.
            </div>
          </template>
        </DataTable>
      </div>
    </div>

    <template #footer>
      <Button label="Tancar" severity="secondary" @click="handleClose" />
    </template>
  </Dialog>
</template>

<style scoped>
.loading-state {
  text-align: center;
  padding: 3rem 1rem;
  color: #64748b;
}

.spinner-icon {
  font-size: 2rem;
  color: #6366f1;
  margin-bottom: 0.75rem;
}

.report-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.metrics-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.metric-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.25rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
}

.metric-value {
  font-size: 2rem;
  font-weight: 800;
  color: #4f46e5;
}

.metric-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #64748b;
  margin-top: 0.25rem;
}

.table-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.table-title {
  font-size: 1rem;
  font-weight: 700;
  color: #1e293b;
  margin: 0;
}

.empty-table-msg {
  text-align: center;
  padding: 2rem 1rem;
  color: #94a3b8;
}
</style>
