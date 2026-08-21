import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PlayerJoinView from '../views/PlayerJoinView.vue'
import { useAuthStore } from '@/modules/auth/store'
import { useMatchStore } from '../store'

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRoute: () => ({ query: { pin: '123456' } }),
  useRouter: () => ({ push: mockPush })
}))

describe('PlayerJoinView.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mockPush.mockClear()
    vi.clearAllMocks()
  })

  it('inicialitza el camp PIN des del query param i el Nickname de l’usuari', async () => {
    const authStore = useAuthStore()
    authStore.user = {
      id: 'u1',
      firstName: 'Pol',
      lastName: 'García',
      email: 'pol@encertia.cat',
      role: 'student',
      isActive: true,
      createdAt: '2026-08-21T10:00:00Z',
      updatedAt: '2026-08-21T10:00:00Z'
    }
    authStore.accessToken = 'fake-token'
    authStore.isInitialized = true

    const wrapper = mount(PlayerJoinView, {
      global: {
        stubs: {
          Card: { template: '<div class="p-card"><slot name="content" /></div>' },
          InputText: {
            props: ['modelValue', 'placeholder', 'id'],
            template: '<input :id="id" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />'
          },
          Button: {
            props: ['disabled', 'loading'],
            template: '<button :disabled="disabled"><slot /></button>'
          },
          Message: { template: '<div><slot /></div>' }
        }
      }
    })

    await flushPromises()

    const pinInput = wrapper.find('[data-testid="input-match-pin"]').element as HTMLInputElement
    const nickInput = wrapper.find('[data-testid="input-player-nickname"]').element as HTMLInputElement

    expect(pinInput.value).toBe('123456')
    expect(nickInput.value).toBe('Pol')
  })

  it('unió satisfactòria redirigeix a /play/:pin', async () => {
    const authStore = useAuthStore()
    authStore.user = {
      id: 'u1',
      firstName: 'Pol',
      lastName: 'García',
      email: 'pol@encertia.cat',
      role: 'student',
      isActive: true,
      createdAt: '2026-08-21T10:00:00Z',
      updatedAt: '2026-08-21T10:00:00Z'
    }
    authStore.accessToken = 'fake-token'
    authStore.isInitialized = true

    const matchStore = useMatchStore()
    vi.spyOn(matchStore, 'joinAndConnectAsPlayer').mockResolvedValue(undefined)

    const wrapper = mount(PlayerJoinView, {
      global: {
        stubs: {
          Card: { template: '<div class="p-card"><slot name="content" /></div>' },
          InputText: {
            props: ['modelValue', 'id'],
            template: '<input :id="id" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />'
          },
          Button: {
            props: ['disabled'],
            template: '<button type="submit" :disabled="disabled"><slot /></button>'
          },
          Message: { template: '<div><slot /></div>' }
        }
      }
    })

    await flushPromises()

    const form = wrapper.find('[data-testid="form-join-match"]')
    await form.trigger('submit.prevent')

    expect(matchStore.joinAndConnectAsPlayer).toHaveBeenCalledWith('123456', 'Pol')
    expect(mockPush).toHaveBeenCalledWith('/play/123456')
  })
})
