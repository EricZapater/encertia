<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { useAuthStore } from '@/modules/auth/store'
import { useUserStore } from '../store'
import type { CreateUserRequest, UpdateUserRequest, User, UserRole } from '../types'

import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Select from 'primevue/select'
import ToggleSwitch from 'primevue/toggleswitch'
import Button from 'primevue/button'
import Message from 'primevue/message'

const props = defineProps<{
  visible: boolean
  user: User | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'saved', user: User): void
}>()

const authStore = useAuthStore()
const userStore = useUserStore()

const isEditMode = computed(() => Boolean(props.user?.id))
const isAdmin = computed(() => authStore.currentUser?.role === 'admin')
const isTeacher = computed(() => authStore.currentUser?.role === 'teacher')

// Opcions de rols segons qui estigui operant
const availableRoleOptions = computed(() => {
  if (isAdmin.value) {
    return [
      { label: 'Alumne (student)', value: 'student' as UserRole },
      { label: 'Professor (teacher)', value: 'teacher' as UserRole },
      { label: 'Administrador (admin)', value: 'admin' as UserRole }
    ]
  }
  // Si és professor només pot assignar 'student'
  return [{ label: 'Alumne (student)', value: 'student' as UserRole }]
})

// Estat del formulari
const form = reactive<{
  firstName: string
  lastName: string
  email: string
  password: string
  role: UserRole
  isActive: boolean
}>({
  firstName: '',
  lastName: '',
  email: '',
  password: '',
  role: 'student',
  isActive: true
})

const errorMessage = ref<string | null>(null)
const isSubmitting = ref(false)

// Sincronitza el formulari quan s'obre o canvia l'usuari seleccionat
watch(
  () => props.user,
  (newUser) => {
    errorMessage.value = null
    if (newUser) {
      form.firstName = newUser.firstName
      form.lastName = newUser.lastName
      form.email = newUser.email
      form.password = ''
      form.role = newUser.role
      form.isActive = newUser.isActive ?? true
    } else {
      // Mode Creació
      form.firstName = ''
      form.lastName = ''
      form.email = ''
      form.password = ''
      form.role = 'student'
      form.isActive = true
    }
  },
  { immediate: true }
)

function closeModal() {
  emit('update:visible', false)
  errorMessage.value = null
}

async function handleSubmit() {
  errorMessage.value = null

  if (!form.firstName.trim() || !form.lastName.trim() || !form.email.trim()) {
    errorMessage.value = 'El nom, els cognoms i el correu electrònic són obligatoris.'
    return
  }

  // En creació, la contrasenya és obligatòria i mínim 8 caràcters
  if (!isEditMode.value) {
    if (!form.password) {
      errorMessage.value = 'La contrasenya inicial és obligatòria en crear un usuari.'
      return
    }
    if (form.password.length < 8) {
      errorMessage.value = 'La contrasenya ha de tenir almenys 8 caràcters.'
      return
    }
  }

  // Validació de rol per a professor
  if (!isAdmin.value && form.role !== 'student') {
    errorMessage.value = 'Només pots crear o gestionar usuaris amb rol d’alumne (student).'
    return
  }

  isSubmitting.value = true

  try {
    let resultUser: User

    if (isEditMode.value && props.user) {
      const updatePayload: UpdateUserRequest = {
        firstName: form.firstName.trim(),
        lastName: form.lastName.trim(),
        email: form.email.trim()
      }

      // Només admin pot enviar canvis de role o isActive
      if (isAdmin.value) {
        updatePayload.role = form.role
        updatePayload.isActive = form.isActive
      }

      resultUser = await userStore.updateUser(props.user.id, updatePayload)
    } else {
      const createPayload: CreateUserRequest = {
        firstName: form.firstName.trim(),
        lastName: form.lastName.trim(),
        email: form.email.trim(),
        password: form.password,
        role: isAdmin.value ? form.role : 'student'
      }

      resultUser = await userStore.createUser(createPayload)
    }

    emit('saved', resultUser)
    closeModal()
  } catch (err: any) {
    errorMessage.value =
      err.response?.data?.error?.message ||
      err.message ||
      'Hi ha hagut un error en desar l’usuari.'
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
    :header="isEditMode ? 'Editar Usuari' : 'Crear Nou Usuari'"
    :style="{ width: '90vw', maxWidth: '560px' }"
    :closable="!isSubmitting"
    data-testid="user-form-dialog"
  >
    <form @submit.prevent="handleSubmit" class="user-form">
      <Message v-if="errorMessage" severity="error" :closable="false" class="form-error">
        {{ errorMessage }}
      </Message>

      <div class="form-grid">
        <div class="form-field">
          <label for="user-firstName">Nom <span class="required">*</span></label>
          <InputText
            id="user-firstName"
            v-model="form.firstName"
            placeholder="Ex. Laura"
            required
            :disabled="isSubmitting"
            class="w-full"
            data-testid="input-firstname"
          />
        </div>

        <div class="form-field">
          <label for="user-lastName">Cognoms <span class="required">*</span></label>
          <InputText
            id="user-lastName"
            v-model="form.lastName"
            placeholder="Ex. Soler Pons"
            required
            :disabled="isSubmitting"
            class="w-full"
            data-testid="input-lastname"
          />
        </div>
      </div>

      <div class="form-field">
        <label for="user-email">Correu Electrònic <span class="required">*</span></label>
        <InputText
          id="user-email"
          v-model="form.email"
          type="email"
          placeholder="nom@encertia.cat"
          required
          :disabled="isSubmitting"
          class="w-full"
          data-testid="input-email"
        />
      </div>

      <!-- Camp de contrasenya només en creació -->
      <div v-if="!isEditMode" class="form-field">
        <label for="user-password">Contrasenya Inicial <span class="required">*</span></label>
        <Password
          id="user-password"
          v-model="form.password"
          toggleMask
          placeholder="Mínim 8 caràcters"
          required
          :disabled="isSubmitting"
          class="w-full"
          inputClass="w-full"
          data-testid="input-password"
        />
        <small class="field-hint">Mínim 8 caràcters. L’usuari podrà canviar-la posteriorment.</small>
      </div>

      <!-- Selector de Rol -->
      <div class="form-field">
        <label for="user-role">Rol d'Usuari</label>
        <Select
          id="user-role"
          v-model="form.role"
          :options="availableRoleOptions"
          optionLabel="label"
          optionValue="value"
          placeholder="Selecciona un rol"
          :disabled="isSubmitting || (!isAdmin && isEditMode) || isTeacher"
          class="w-full"
          data-testid="select-role"
        />
        <small v-if="!isAdmin" class="field-hint">
          {{ isTeacher ? 'Com a professor només pots gestionar alumnes (student).' : 'El rol només pot ser modificat per un administrador.' }}
        </small>
      </div>

      <!-- Estat Actiu / Inactiu (només editable per admin en mode edició) -->
      <div v-if="isEditMode" class="form-field status-field">
        <div class="status-toggle-wrapper">
          <div>
            <label class="status-label">Estat del Compte</label>
            <p class="status-desc">
              {{ form.isActive ? 'Compte actiu amb accés a la plataforma.' : 'Compte inactiu (accés bloquejat).' }}
            </p>
          </div>
          <ToggleSwitch
            v-model="form.isActive"
            :disabled="!isAdmin || isSubmitting"
            data-testid="switch-isactive"
          />
        </div>
        <small v-if="!isAdmin" class="field-hint">
          Només un administrador pot activar o desactivar comptes d'usuaris.
        </small>
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
          :label="isEditMode ? 'Desar Canvis' : 'Crear Usuari'"
          :icon="isEditMode ? 'pi pi-check' : 'pi pi-user-plus'"
          :loading="isSubmitting"
          data-testid="btn-submit-user"
        />
      </div>
    </form>
  </Dialog>
</template>

<style scoped>
.user-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding-top: 0.5rem;
}

.form-error {
  margin-bottom: 0.5rem;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
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

.field-hint {
  font-size: 0.75rem;
  color: #64748b;
}

.status-field {
  padding: 0.75rem 1rem;
  background-color: #f8fafc;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
}

.status-toggle-wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.status-label {
  font-size: 0.9rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.status-desc {
  font-size: 0.8rem;
  color: #64748b;
  margin: 0.2rem 0 0 0;
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

@media (max-width: 520px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
