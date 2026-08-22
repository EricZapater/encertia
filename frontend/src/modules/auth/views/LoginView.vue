<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../store'
import type { LoginRequest } from '../types'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const toast = useToast()

const credentials = reactive<LoginRequest>({
  email: '',
  password: ''
})

const errorMessage = ref<string | null>(null)
const isSubmitting = ref(false)

async function handleLogin() {
  errorMessage.value = null

  if (!credentials.email || !credentials.password) {
    const msg = 'Si us plau, omple tots els camps.'
    toast.add({ severity: 'error', summary: 'Error de Validació', detail: msg, life: 4000 })
    return
  }

  isSubmitting.value = true
  try {
    await authStore.login(credentials)
    const redirectPath = (route.query.redirect as string) || '/profile'
    router.push(redirectPath)
  } catch (err: any) {
    const msg =
      err.response?.data?.error?.message ||
      authStore.error ||
      'Credencials incorrectes o error en la connexió.'
    toast.add({ severity: 'error', summary: 'Error d’Accés', detail: msg, life: 4000 })
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
        <p class="brand-subtitle">Plataforma Educativa de Qüestionaris i Aprenentatge</p>
      </div>

      <Card class="auth-card">
        <template #title>
          <div class="card-title">Inici de Sessió</div>
        </template>
        <template #subtitle>
          <div class="card-subtitle">Introdueix les teves credencials per accedir</div>
        </template>

        <template #content>
          <form @submit.prevent="handleLogin" class="auth-form">

            <div class="form-field">
              <label for="email">Correu Electrònic</label>
              <InputText
                id="email"
                v-model="credentials.email"
                type="email"
                placeholder="nom@exemple.cat"
                autocomplete="email"
                required
                class="w-full"
                :disabled="isSubmitting"
              />
            </div>

            <div class="form-field">
              <label for="password">Contrasenya</label>
              <Password
                id="password"
                v-model="credentials.password"
                :feedback="false"
                toggleMask
                placeholder="••••••••"
                autocomplete="current-password"
                required
                class="w-full"
                inputClass="w-full"
                :disabled="isSubmitting"
              />
            </div>

            <Button
              type="submit"
              label="Iniciar Sessió"
              icon="pi pi-sign-in"
              :loading="isSubmitting"
              class="w-full submit-btn"
            />
          </form>
        </template>

        <template #footer>
          <div class="auth-footer">
            <span>No tens un compte?</span>
            <router-link to="/register" class="auth-link">Registra't aquí</router-link>
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
  max-width: 440px;
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
  gap: 1.25rem;
  margin-top: 0.5rem;
}

.auth-message {
  margin-bottom: 0.5rem;
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
</style>
