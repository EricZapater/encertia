<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/modules/auth/store'
import { useMatchStore } from '../store'

import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Card from 'primevue/card'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const matchStore = useMatchStore()

const pin = ref('')
const nickname = ref('')
const formError = ref<string | null>(null)
const isSubmitting = ref(false)

onMounted(async () => {
  // Inicialitza l'usuari si cal
  if (!authStore.isInitialized) {
    await authStore.initAuth()
  }

  // Si no està autenticat, redirigeix al login guardant el redirect
  if (!authStore.isAuthenticated) {
    const currentPin = (route.query.pin as string) || ''
    const redirectUrl = currentPin ? `/play?pin=${encodeURIComponent(currentPin)}` : '/play'
    router.push({
      name: 'login',
      query: { redirect: redirectUrl }
    })
    return
  }

  // Pre-omplir el PIN des del query param si existeix
  if (route.query.pin && typeof route.query.pin === 'string') {
    pin.value = route.query.pin.trim()
  }

  // Pre-omplir el Nickname amb el nom de l'usuari autenticat
  if (authStore.currentUser) {
    nickname.value =
      authStore.currentUser.firstName ||
      authStore.fullName ||
      authStore.currentUser.email.split('@')[0]
  }
})

const isValidPin = computed(() => /^\d{6}$/.test(pin.value.trim()))
const isValidNickname = computed(() => nickname.value.trim().length >= 2 && nickname.value.trim().length <= 30)
const canSubmit = computed(() => isValidPin.value && isValidNickname.value && !isSubmitting.value)

async function handleSubmit() {
  formError.value = null

  if (!isValidPin.value) {
    formError.value = 'El PIN ha de tenir exactament 6 dígits numèrics.'
    return
  }

  if (!isValidNickname.value) {
    formError.value = 'El Nickname ha de tenir entre 2 i 30 caràcters.'
    return
  }

  isSubmitting.value = true
  try {
    const cleanPin = pin.value.trim()
    const cleanNick = nickname.value.trim()

    await matchStore.joinAndConnectAsPlayer(cleanPin, cleanNick)
    router.push(`/play/${cleanPin}`)
  } catch (err: any) {
    formError.value =
      err.response?.data?.error?.message ||
      err.message ||
      'No s’ha pogut connectar a la partida. Revisa el PIN.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="player-join-container">
    <div class="join-card-wrapper">
      <!-- Logo Encertia -->
      <div class="brand-header">
        <div class="brand-icon">
          <i class="pi pi-bolt" />
        </div>
        <h1 class="brand-title">Encertia Live</h1>
        <p class="brand-subtitle">Entra a la partida en directe</p>
      </div>

      <Card class="join-card">
        <template #content>
          <!-- Missatge d'error -->
          <Message
            v-if="formError || matchStore.error"
            severity="error"
            class="mb-4"
            :closable="true"
            @close="formError = null; matchStore.clearError()"
            data-testid="msg-join-error"
          >
            {{ formError || matchStore.error }}
          </Message>

          <form @submit.prevent="handleSubmit" class="join-form" data-testid="form-join-match">
            <!-- Camp PIN -->
            <div class="field-group">
              <label for="input-pin" class="field-label">Codi PIN de la partida</label>
              <div class="pin-input-container">
                <InputText
                  id="input-pin"
                  v-model="pin"
                  placeholder="Ex. 482910"
                  maxlength="6"
                  inputmode="numeric"
                  pattern="[0-9]*"
                  class="pin-input"
                  data-testid="input-match-pin"
                  autofocus
                />
              </div>
              <small class="field-hint">Introdueix els 6 dígits que mostra el moderador</small>
            </div>

            <!-- Camp Nickname -->
            <div class="field-group">
              <label for="input-nickname" class="field-label">El teu Nickname</label>
              <InputText
                id="input-nickname"
                v-model="nickname"
                placeholder="Com et vols dir a la sala?"
                maxlength="30"
                class="nickname-input"
                data-testid="input-player-nickname"
              />
              <small class="field-hint">Aquest nom apareixerà al rànquing i a la pantalla</small>
            </div>

            <!-- Botó d'unió -->
            <Button
              type="submit"
              label="Entrar a la Partida"
              icon="pi pi-sign-in"
              size="large"
              class="join-btn"
              :disabled="!canSubmit"
              :loading="isSubmitting || matchStore.isLoading"
              data-testid="btn-submit-join"
            />
          </form>
        </template>
      </Card>

      <!-- Peu de pàgina informatiu -->
      <div class="join-footer">
        <p>
          Connectat com a <strong>{{ authStore.fullName || authStore.currentUser?.email }}</strong>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.player-join-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1e3a8a 0%, #3b82f6 50%, #60a5fa 100%);
  padding: 1.5rem;
}

.join-card-wrapper {
  width: 100%;
  max-width: 440px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.brand-header {
  text-align: center;
  color: #ffffff;
  margin-bottom: 1.5rem;
}

.brand-icon {
  width: 64px;
  height: 64px;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(8px);
  border-radius: 1rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  margin-bottom: 0.75rem;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.brand-title {
  font-size: 2.25rem;
  font-weight: 800;
  letter-spacing: -0.025em;
  margin: 0;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.brand-subtitle {
  font-size: 1rem;
  opacity: 0.9;
  margin-top: 0.25rem;
}

.join-card {
  width: 100%;
  border-radius: 1.25rem;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.25);
  border: none;
  background-color: #ffffff;
}

.join-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.field-label {
  font-size: 0.9rem;
  font-weight: 600;
  color: #334155;
}

.pin-input {
  width: 100%;
  text-align: center;
  font-size: 1.75rem;
  font-weight: 700;
  letter-spacing: 0.25em;
  padding: 0.75rem;
}

.nickname-input {
  width: 100%;
  font-size: 1.1rem;
  padding: 0.75rem;
}

.field-hint {
  font-size: 0.8rem;
  color: #64748b;
}

.join-btn {
  margin-top: 0.5rem;
  padding: 0.9rem;
  font-size: 1.15rem;
  font-weight: 700;
  border-radius: 0.75rem;
  background-color: #2563eb;
  border-color: #2563eb;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.join-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(37, 99, 235, 0.4);
}

.join-footer {
  margin-top: 1.5rem;
  text-align: center;
  color: rgba(255, 255, 255, 0.85);
  font-size: 0.85rem;
}

.mb-4 {
  margin-bottom: 1rem;
}
</style>
