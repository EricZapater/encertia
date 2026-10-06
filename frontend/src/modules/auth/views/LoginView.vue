<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../store'
import type { LoginRequest } from '../types'
import { setAppLanguage, type SupportedLanguage } from '@/i18n'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const toast = useToast()
const { t, locale } = useI18n()

const credentials = reactive<LoginRequest>({
  email: '',
  password: ''
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

async function handleLogin() {
  errorMessage.value = null

  if (!credentials.email || !credentials.password) {
    const msg = t('auth.login.error')
    toast.add({ severity: 'error', summary: t('common.error'), detail: msg, life: 4000 })
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
      t('auth.login.error')
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
        <p class="brand-subtitle">{{ $t('auth.login.subtitle') }}</p>
      </div>

      <Card class="auth-card">
        <template #title>
          <div class="card-title">{{ $t('auth.login.title') }}</div>
        </template>

        <template #content>
          <form @submit.prevent="handleLogin" class="auth-form">
            <div class="form-field">
              <label for="email">{{ $t('auth.login.email') }}</label>
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
              <label for="password">{{ $t('auth.login.password') }}</label>
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
              :label="$t('auth.login.submit')"
              icon="pi pi-sign-in"
              :loading="isSubmitting"
              class="w-full submit-btn"
            />
          </form>
        </template>

        <template #footer>
          <div class="auth-footer">
            <span>{{ $t('auth.login.noAccount') }}</span>
            <router-link to="/register" class="auth-link">{{ $t('auth.login.registerLink') }}</router-link>
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
  gap: 1.25rem;
  margin-top: 0.5rem;
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
