<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../store'
import type { RegisterRequest } from '../types'
import { setAppLanguage, type SupportedLanguage } from '@/i18n'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const authStore = useAuthStore()
const toast = useToast()
const { t, locale } = useI18n()

const form = reactive<RegisterRequest>({
  firstName: '',
  lastName: '',
  email: '',
  password: '',
  role: 'student'
})

const errorMessage = ref<string | null>(null)
const isSubmitting = ref(false)

const supportedLangs: { code: SupportedLanguage; label: string }[] = [
  { code: 'ca', label: 'CA' },
  { code: 'es', label: 'ES' },
  { code: 'en', label: 'EN' }
]

function changeLang(langCode: SupportedLanguage) {
  setAppLanguage(langCode)
}

async function handleRegister() {
  errorMessage.value = null

  if (!form.firstName.trim() || !form.lastName.trim() || !form.email.trim() || !form.password) {
    const msg = t('common.error')
    toast.add({ severity: 'error', summary: t('common.error'), detail: msg, life: 4000 })
    return
  }

  if (form.password.length < 8) {
    const msg = t('common.error')
    toast.add({ severity: 'error', summary: t('common.error'), detail: msg, life: 4000 })
    return
  }

  isSubmitting.value = true
  try {
    await authStore.register({
      firstName: form.firstName.trim(),
      lastName: form.lastName.trim(),
      email: form.email.trim(),
      password: form.password,
      role: 'student',
      language: locale.value as SupportedLanguage
    })
    toast.add({ severity: 'success', summary: t('common.success'), detail: t('auth.register.title'), life: 3000 })
    router.push('/profile')
  } catch (err: any) {
    const msg =
      err.response?.data?.error?.message ||
      authStore.error ||
      t('common.error')
    toast.add({ severity: 'error', summary: t('common.error'), detail: msg, life: 4000 })
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <div class="auth-card-wrapper">
      <div class="auth-header-top">
        <div class="lang-selector-top">
          <span class="lang-globe">🌐</span>
          <button
            v-for="lang in supportedLangs"
            :key="lang.code"
            type="button"
            class="lang-btn"
            :class="{ active: locale === lang.code }"
            @click="changeLang(lang.code)"
          >
            {{ lang.label }}
          </button>
        </div>
      </div>

      <div class="auth-brand">
        <h1 class="brand-title">Encertia</h1>
        <p class="brand-subtitle">{{ $t('auth.register.subtitle') }}</p>
      </div>

      <Card class="auth-card">
        <template #title>
          <div class="card-title">{{ $t('auth.register.title') }}</div>
        </template>

        <template #content>
          <form @submit.prevent="handleRegister" class="auth-form">
            <div class="form-row">
              <div class="form-field flex-1">
                <label for="firstName">{{ $t('auth.register.firstName') }}</label>
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
                <label for="lastName">{{ $t('auth.register.lastName') }}</label>
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
              <label for="email">{{ $t('auth.register.email') }}</label>
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
              <label for="password">{{ $t('auth.register.password') }}</label>
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
              :label="$t('auth.register.submit')"
              icon="pi pi-user-plus"
              :loading="isSubmitting"
              class="w-full submit-btn"
            />
          </form>
        </template>

        <template #footer>
          <div class="auth-footer">
            <span>{{ $t('auth.register.hasAccount') }}</span>
            <router-link to="/login" class="auth-link">{{ $t('auth.register.loginLink') }}</router-link>
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

.auth-header-top {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 0.75rem;
}

.lang-selector-top {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  background-color: #ffffff;
  padding: 0.25rem 0.5rem;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.lang-globe {
  font-size: 0.85rem;
  margin-right: 0.15rem;
}

.lang-btn {
  background: none;
  border: none;
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
  padding: 0.2rem 0.4rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.lang-btn:hover {
  color: #0f172a;
}

.lang-btn.active {
  background-color: #4338ca;
  color: #ffffff;
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

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1.15rem;
  margin-top: 0.5rem;
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
