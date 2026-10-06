<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import Button from 'primevue/button'
import { useGroupStore } from '@/modules/groups/store'
import { useMatchStore } from '@/modules/match/store'

const props = defineProps<{
  visible: boolean
  quiz: { id: string; title: string } | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const router = useRouter()
const groupStore = useGroupStore()
const matchStore = useMatchStore()

const selectedGroupId = ref<string | null>(null)
const isLoading = ref(false)
const errorMsg = ref<string | null>(null)

watch(
  () => props.visible,
  async (val) => {
    if (val) {
      selectedGroupId.value = null
      errorMsg.value = null
      if (!groupStore.hasGroups) {
        await groupStore.fetchGroups()
      }
    }
  }
)

function handleClose() {
  emit('update:visible', false)
}

async function handleConfirmLaunch() {
  if (!props.quiz?.id) return
  isLoading.value = true
  errorMsg.value = null
  try {
    const res = await matchStore.initHostMatch(props.quiz.id, selectedGroupId.value)
    emit('update:visible', false)
    router.push(`/matches/${res.id}/host`)
  } catch (err: any) {
    errorMsg.value =
      err.response?.data?.error?.message || err.message || 'Error en iniciar la partida.'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    header="Iniciar Partida en Directe"
    :style="{ width: '90vw', maxWidth: '480px' }"
    @update:visible="handleClose"
    data-testid="launch-match-dialog"
  >
    <div class="launch-modal-content">
      <p v-if="quiz?.title" class="quiz-info-text">
        Estàs a punt d'iniciar la partida per al qüestionari <strong>{{ quiz.title }}</strong>.
      </p>

      <div class="field-group mt-3">
        <label for="group-select" class="field-label font-semibold block mb-2">
          Selecciona el Grup d'Alumnes (Opcional):
        </label>
        <Select
          id="group-select"
          v-model="selectedGroupId"
          :options="[
            { label: 'Sense grup associat (Partida lliure)', value: null },
            ...groupStore.groupList.map((g) => ({
              label: `${g.name} (${g.academicYear})`,
              value: g.id
            }))
          ]"
          optionLabel="label"
          optionValue="value"
          placeholder="Selecciona un grup..."
          class="w-full"
          :loading="groupStore.isLoading"
          data-testid="select-group-match"
        />
        <small class="field-hint text-secondary block mt-1">
          Si vincules la partida a un grup, els resultats i estadístiques es desaran per a l'avaluació dels alumnes del grup.
        </small>
      </div>

      <p v-if="errorMsg" class="error-msg text-red-500 mt-2">{{ errorMsg }}</p>
    </div>

    <template #footer>
      <Button
        label="Cancel·lar"
        icon="pi pi-times"
        text
        severity="secondary"
        @click="handleClose"
      />
      <Button
        label="Iniciar Partida"
        icon="pi pi-play"
        severity="success"
        :loading="isLoading"
        @click="handleConfirmLaunch"
        data-testid="btn-confirm-launch-match"
      />
    </template>
  </Dialog>
</template>

<style scoped>
.launch-modal-content {
  padding: 0.5rem 0;
}

.quiz-info-text {
  font-size: 1rem;
  color: #334155;
  margin: 0;
}

.field-label {
  color: #0f172a;
}

.field-hint {
  font-size: 0.825rem;
  color: #64748b;
}

.error-msg {
  font-size: 0.875rem;
}
</style>
