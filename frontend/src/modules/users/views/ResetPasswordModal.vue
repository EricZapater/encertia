<script setup lang="ts">
import { ref, watch } from 'vue'
import { useUserStore } from '../store'
import type { User } from '../types'

import Dialog from 'primevue/dialog'
import Password from 'primevue/password'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { useToast } from 'primevue/usetoast'

const props = defineProps<{
  visible: boolean
  user: User | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'success', message: string): void
}>()

const userStore = useUserStore()
const toast = useToast()

const newPassword = ref('')
const confirmPassword = ref('')
const errorMessage = ref<string | null>(null)
const isSubmitting = ref(false)

watch(
  () => props.visible,
  (isOpen) => {
    if (isOpen) {
      newPassword.value = ''
      confirmPassword.value = ''
      errorMessage.value = null
    }
  }
)

function closeModal() {
  emit('update:visible', false)
  errorMessage.value = null
}

async function handleResetPassword() {
  errorMessage.value = null

  if (!newPassword.value) {
    const msg = 'Has d’introduir la nova contrasenya.'
    toast.add({ severity: 'error', summary: 'Error de Validació', detail: msg, life: 4000 })
    return
  }

  if (newPassword.value.length < 8) {
    const msg = 'La contrasenya ha de contenir com a mínim 8 caràcters.'
    toast.add({ severity: 'error', summary: 'Error de Validació', detail: msg, life: 4000 })
    return
  }

  if (newPassword.value !== confirmPassword.value) {
    const msg = 'Les contrasenyes no coincideixen.'
    toast.add({ severity: 'error', summary: 'Error de Validació', detail: msg, life: 4000 })
    return
  }

  if (!props.user?.id) {
    const msg = 'No s’ha seleccionat cap usuari.'
    toast.add({ severity: 'error', summary: 'Error', detail: msg, life: 4000 })
    return
  }

  isSubmitting.value = true

  try {
    const msg = await userStore.resetUserPassword(props.user.id, {
      newPassword: newPassword.value
    })
    const successMsg = msg || 'Contrasenya actualitzada correctament.'
    toast.add({ severity: 'success', summary: 'Contrasenya Restablerta', detail: successMsg, life: 4000 })
    emit('success', successMsg)
    closeModal()
  } catch (err: any) {
    const msg =
      err.response?.data?.error?.message ||
      err.message ||
      'Error en restablir la contrasenya.'
    toast.add({ severity: 'error', summary: 'Error de Contrasenya', detail: msg, life: 4000 })
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    @update:visible="(val) => emit('update:visible', val)"
    modal
    header="Restablir Contrasenya d'Usuari"
    :style="{ width: '90vw', maxWidth: '480px' }"
    :closable="!isSubmitting"
    data-testid="reset-password-dialog"
  >
    <div v-if="user" class="target-user-info">
      <i class="pi pi-user user-icon" />
      <div>
        <div class="user-name">{{ user.firstName }} {{ user.lastName }}</div>
        <div class="user-email">{{ user.email }}</div>
      </div>
    </div>

    <form @submit.prevent="handleResetPassword" class="reset-form">

      <Message severity="info" :closable="false">
        En restablir la contrasenya, totes les sessions actives de l'usuari seran revocades i haurà d'iniciar sessió novament.
      </Message>

      <div class="form-field">
        <label for="new-password">Nova Contrasenya <span class="required">*</span></label>
        <Password
          id="new-password"
          v-model="newPassword"
          toggleMask
          placeholder="Mínim 8 caràcters"
          required
          :disabled="isSubmitting"
          class="w-full"
          inputClass="w-full"
          data-testid="input-new-password"
        />
      </div>

      <div class="form-field">
        <label for="confirm-password">Confirmar Nova Contrasenya <span class="required">*</span></label>
        <Password
          id="confirm-password"
          v-model="confirmPassword"
          :feedback="false"
          toggleMask
          placeholder="Repeteix la contrasenya"
          required
          :disabled="isSubmitting"
          class="w-full"
          inputClass="w-full"
          data-testid="input-confirm-password"
        />
      </div>

      <div class="dialog-actions">
        <Button
          type="button"
          label="Cancel·lar"
          icon="pi pi-times"
          severity="secondary"
          text
          @click="closeModal"
          :disabled="isSubmitting"
        />
        <Button
          type="submit"
          label="Actualitzar Contrasenya"
          icon="pi pi-key"
          severity="warning"
          :loading="isSubmitting"
          data-testid="btn-submit-reset-password"
        />
      </div>
    </form>
  </Dialog>
</template>

<style scoped>
.target-user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background-color: #f1f5f9;
  border-radius: 0.5rem;
  margin-bottom: 1.25rem;
}

.user-icon {
  font-size: 1.5rem;
  color: #3b82f6;
}

.user-name {
  font-weight: 600;
  color: #1e293b;
}

.user-email {
  font-size: 0.85rem;
  color: #64748b;
}

.reset-form {
  display: flex;
  flex-direction: column;
  gap: 1.15rem;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-field label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #334155;
}

.required {
  color: #ef4444;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #f1f5f9;
}

.w-full {
  width: 100%;
}
</style>
