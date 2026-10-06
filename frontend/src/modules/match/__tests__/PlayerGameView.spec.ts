import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PlayerGameView from '../views/PlayerGameView.vue'
import { useAuthStore } from '@/modules/auth/store'
import { useMatchStore } from '../store'

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { pin: '123456' } }),
  useRouter: () => ({ push: mockPush })
}))

describe('PlayerGameView.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mockPush.mockClear()
    vi.clearAllMocks()
  })

  it('mostra la sala d’espera quan status és lobby', () => {
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
    matchStore.pin = '123456'
    matchStore.status = 'lobby'
    matchStore.myNickname = 'Pol'
    matchStore.myScore = 0
    matchStore.isConnected = true

    const wrapper = mount(PlayerGameView, {
      global: {
        stubs: {
          Tag: { template: '<span><slot /></span>' },
          Button: { template: '<button><slot /></button>' },
          ProgressBar: { template: '<div></div>' }
        }
      }
    })

    expect(wrapper.find('[data-testid="screen-player-lobby"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Estàs a dins!')
  })

  it('mostra les opcions de resposta quan status és question_active i permet seleccionar', async () => {
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
    matchStore.pin = '123456'
    matchStore.status = 'question_active'
    matchStore.myNickname = 'Pol'
    matchStore.timerSeconds = 15
    matchStore.totalQuestions = 5
    matchStore.currentQuestionIndex = 0
    matchStore.currentQuestion = {
      id: 'q1',
      orderIndex: 0,
      title: 'Quina és la capital de Catalunya?',
      type: 'single_choice',
      timeLimitSeconds: 20,
      points: 1,
      options: [
        { id: 'opt-1', text: 'Barcelona' },
        { id: 'opt-2', text: 'Girona' },
        { id: 'opt-3', text: 'Lleida' },
        { id: 'opt-4', text: 'Tarragona' }
      ]
    }
    matchStore.mySelectedAnswerIds = []
    matchStore.hasSubmittedAnswer = false
    matchStore.isConnected = true

    vi.spyOn(matchStore, 'selectAnswer')
    vi.spyOn(matchStore, 'submitAnswer')

    const wrapper = mount(PlayerGameView, {
      global: {
        stubs: {
          Tag: { template: '<span><slot /></span>' },
          Button: { template: '<button><slot /></button>' },
          ProgressBar: { template: '<div></div>' }
        }
      }
    })

    expect(wrapper.find('[data-testid="screen-player-active"]').exists()).toBe(true)

    // Clic a la primera opció (Barcelona)
    const opt0 = wrapper.find('[data-testid="btn-option-0"]')
    expect(opt0.exists()).toBe(true)
    expect(opt0.text()).toContain('Barcelona')

    await opt0.trigger('click')
    expect(matchStore.selectAnswer).toHaveBeenCalledWith('opt-1', false)
    expect(matchStore.submitAnswer).toHaveBeenCalled()
  })

  it('mostra retroacció de resultat correcte quan status és question_results', () => {
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
    matchStore.pin = '123456'
    matchStore.status = 'question_results'
    matchStore.myNickname = 'Pol'
    matchStore.myScore = 1
    matchStore.lastAnswerResult = {
      isCorrect: true,
      scoreAwarded: 1,
      totalScore: 1
    }
    matchStore.isConnected = true

    const wrapper = mount(PlayerGameView, {
      global: {
        stubs: {
          Tag: { template: '<span><slot /></span>' },
          Button: { template: '<button><slot /></button>' },
          ProgressBar: { template: '<div></div>' }
        }
      }
    })

    expect(wrapper.find('[data-testid="screen-player-results"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="feedback-correct"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Resposta Correcta!')
  })
})
