<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store'
import type { RegisterRequest } from '../types'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const authStore = useAuthStore()
const toast = useToast()

const form = reactive<RegisterRequest>({
  firstName: '',
  lastName: '',
  email: '',
  password: '',
  role: 'student'
})

const errorMessage = ref<string | null>(null)
const isSubmitting = ref(false)

async function handleRegister() {
  errorMessage.value = null

  if (!form.firstName.trim() || !form.lastName.trim() || !form.email.trim() || !form.password) {
    const msg = 'Tots els camps són obligatoris.'
    toast.add({ severity: 'error', summary: 'Error de Validació', detail: msg, life: 4000 })
    return
  }

  if (form.password.length < 8) {
    const msg = 'La contrasenya ha de tenir com a mínim 8 caràcters.'
    toast.add({ severity: 'error', summary: 'Error de Validació', detail: msg, life: 4000 })
    return
  }

  isSubmitting.value = true
  try {
    await authStore.register({
      firstName: form.firstName.trim(),
      lastName: form.lastName.trim(),
      email: form.email.trim(),
      password: form.password,
      role: 'student'
    })
    toast.add({ severity: 'success', summary: 'Compte Creat', detail: 'Benvingut a Encertia!', life: 3000 })
    router.push('/profile')
  } catch (err: any) {
    const msg =
      err.response?.data?.error?.message ||
      authStore.error ||
      'Error en crear el compte. Si us plau, revisa les dades.'
    toast.add({ severity: 'error', summary: 'Error de Registre', detail: msg, life: 4000 })
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <div class="auth-card-wrapper">
      <div class="auth-brand">
        <h1 class="brand-title">Encertia</h1>
        <p class="brand-subtitle">Crea el teu compte per començar</p>
      </div>

      <Card class="auth-card">
        <template #title>
          <div class="card-title">Registre d'Usuari</div>
        </template>
        <template #subtitle>
          <div class="card-subtitle">Completa el formulari amb les teves dades d'estudiant</div>
        </template>

        <template #content>
          <form @submit.prevent="handleRegister" class="auth-form">

            <div class="form-row">
              <div class="form-field flex-1">
                <label for="firstName">Nom</label>
                <InputText
                  id="firstName"
                  v-model="form.firstName"
                  type="text"
                  placeholder="Maria"
                  required
                  class="w-full"
                  :disabled="isSubmitting"
                />
              </div>

              <div class="form-field flex-1">
                <label for="lastName">Cognoms</label>
                <InputText
                  id="lastName"
                  v-model="form.lastName"
                  type="text"
                  placeholder="Garcia"
                  required
                  class="w-full"
                  :disabled="isSubmitting"
                />
              </div>
            </div>

            <div class="form-field">
              <label for="email">Correu Electrònic</label>
              <InputText
                id="email"
                v-model="form.email"
                type="email"
                placeholder="maria.garcia@encertia.cat"
                autocomplete="email"
                required
                class="w-full"
                :disabled="isSubmitting"
              />
            </div>

            <div class="form-field">
              <label for="password">Contrasenya (mínim 8 caràcters)</label>
              <Password
                id="password"
                v-model="form.password"
                toggleMask
                placeholder="••••••••"
                autocomplete="new-password"
                required
                class="w-full"
                inputClass="w-full"
                :disabled="isSubmitting"
              />
            </div>

            <Button
              type="submit"
              label="Crear Compte"
              icon="pi pi-user-plus"
              :loading="isSubmitting"
              class="w-full submit-btn"
            />
          </form>
        </template>

        <template #footer>
          <div class="auth-footer">
            <span>Ja tens un compte?</span>
            <router-link to="/login" class="auth-link">Inicia sessió aquí</router-link>
          </div>
        </template>
      </Card>
    </div>
  </div>
</template>

<style scoped>
.auth-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 1.5rem;
  background: linear-gradient(135deg, #f0f4f8 0%, #e2e8f0 100%);
}

.auth-card-wrapper {
  width: 100%;
  max-width: 520px;
}

.auth-brand {
  text-align: center;
  margin-bottom: 1.5rem;
}

.brand-title {
  font-size: 2.25rem;
  font-weight: 800;
  color: #1e3a8a;
  margin: 0;
  letter-spacing: -0.025em;
}

.brand-subtitle {
  font-size: 0.925rem;
  color: #64748b;
  margin-top: 0.25rem;
}

.auth-card {
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
  border-radius: 1rem;
  border: 1px solid #e2e8f0;
}

.card-title {
  font-size: 1.35rem;
  font-weight: 700;
  color: #0f172a;
}

.card-subtitle {
  font-size: 0.875rem;
  color: #64748b;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1.15rem;
  margin-top: 0.5rem;
}

.auth-message {
  margin-bottom: 0.5rem;
}

.form-row {
  display: flex;
  gap: 0.75rem;
}

.flex-1 {
  flex: 1;
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

.role-selector {
  display: flex;
}

.role-selector :deep(.p-button) {
  flex: 1;
  font-size: 0.85rem;
}

.w-full {
  width: 100%;
}

.submit-btn {
  margin-top: 0.5rem;
}

.auth-footer {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: #64748b;
  padding-top: 0.5rem;
  border-top: 1px solid #f1f5f9;
}

.auth-link {
  color: #2563eb;
  font-weight: 600;
  text-decoration: none;
}

.auth-link:hover {
  text-decoration: underline;
}

@media (max-width: 480px) {
  .form-row {
    flex-direction: column;
  }
}
</style>
