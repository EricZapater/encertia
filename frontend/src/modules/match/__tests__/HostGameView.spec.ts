import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import HostGameView from '../views/HostGameView.vue'
import { useAuthStore } from '@/modules/auth/store'
import { useMatchStore } from '../store'

vi.mock('qrcode', () => ({
  default: {
    toDataURL: vi.fn().mockResolvedValue('data:image/png;base64,fake-qr-code')
  }
}))

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { id: 'm123' }, fullPath: '/matches/m123/host' }),
  useRouter: () => ({ push: mockPush })
}))

describe('HostGameView.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mockPush.mockClear()
    vi.clearAllMocks()
  })

  it('mostra el PIN, codi QR i llista de jugadors a la fase Lobby', () => {
    const authStore = useAuthStore()
    authStore.user = {
      id: 'u1',
      firstName: 'Teacher',
      lastName: 'Ensenya',
      email: 'prof@encertia.cat',
      role: 'teacher',
      isActive: true,
      createdAt: '2026-08-21T10:00:00Z',
      updatedAt: '2026-08-21T10:00:00Z'
    }
    authStore.accessToken = 'fake-token'
    authStore.isInitialized = true

    const matchStore = useMatchStore()
    matchStore.matchId = 'm123'
    matchStore.pin = '654321'
    matchStore.quizTitle = 'Geografia de Catalunya'
    matchStore.status = 'lobby'
    matchStore.role = 'host'
    matchStore.players = [
      { id: 'p1', nickname: 'Joan', score: 0, isConnected: true, isKicked: false },
      { id: 'p2', nickname: 'Mireia', score: 0, isConnected: true, isKicked: false }
    ]
    matchStore.isConnected = true

    const wrapper = mount(HostGameView, {
      global: {
        stubs: {
          Tag: { template: '<span><slot /></span>' },
          Button: { template: '<button><slot /></button>' },
          Dialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    expect(wrapper.find('[data-testid="host-phase-lobby"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="host-pin-display"]').text()).toContain('654321')
    expect(wrapper.find('[data-testid="host-players-grid"]').text()).toContain('Joan')
    expect(wrapper.find('[data-testid="host-players-grid"]').text()).toContain('Mireia')
  })

  it('fase question_preview permet clicar a Iniciar Temps', async () => {
    const authStore = useAuthStore()
    authStore.user = {
      id: 'u1',
      firstName: 'Teacher',
      lastName: 'Ensenya',
      email: 'prof@encertia.cat',
      role: 'teacher',
      isActive: true,
      createdAt: '2026-08-21T10:00:00Z',
      updatedAt: '2026-08-21T10:00:00Z'
    }
    authStore.accessToken = 'fake-token'
    authStore.isInitialized = true

    const matchStore = useMatchStore()
    matchStore.matchId = 'm123'
    matchStore.pin = '654321'
    matchStore.status = 'question_preview'
    matchStore.role = 'host'
    matchStore.currentQuestionIndex = 0
    matchStore.totalQuestions = 3
    matchStore.currentQuestion = {
      id: 'q1',
      orderIndex: 0,
      title: 'Quin és el cim més alt de Catalunya?',
      type: 'single_choice',
      timeLimitSeconds: 30,
      points: 1,
      options: []
    }
    matchStore.players = []
    matchStore.isConnected = true

    vi.spyOn(matchStore, 'startQuestionTimer')

    const wrapper = mount(HostGameView, {
      global: {
        stubs: {
          Tag: { template: '<span><slot /></span>' },
          Button: { template: '<button><slot /></button>' },
          Dialog: { template: '<div></div>' }
        }
      }
    })

    expect(wrapper.find('[data-testid="host-phase-preview"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="host-preview-title"]').text()).toContain('Quin és el cim més alt de Catalunya?')

    const startTimerBtn = wrapper.find('[data-testid="btn-start-timer"]')
    await startTimerBtn.trigger('click')

    expect(matchStore.startQuestionTimer).toHaveBeenCalled()
  })

  it('fase question_active mostra el temporitzador i botó per tancar resultats', async () => {
    const authStore = useAuthStore()
    authStore.user = {
      id: 'u1',
      firstName: 'Teacher',
      lastName: 'Ensenya',
      email: 'prof@encertia.cat',
      role: 'teacher',
      isActive: true,
      createdAt: '2026-08-21T10:00:00Z',
      updatedAt: '2026-08-21T10:00:00Z'
    }
    authStore.accessToken = 'fake-token'
    authStore.isInitialized = true

    const matchStore = useMatchStore()
    matchStore.matchId = 'm123'
    matchStore.pin = '654321'
    matchStore.status = 'question_active'
    matchStore.role = 'host'
    matchStore.timerSeconds = 18
    matchStore.answeredCount = 4
    matchStore.players = [
      { id: 'p1', nickname: 'A', score: 0, isConnected: true, isKicked: false },
      { id: 'p2', nickname: 'B', score: 0, isConnected: true, isKicked: false }
    ]
    matchStore.currentQuestion = {
      id: 'q1',
      orderIndex: 0,
      title: 'Pregunta activa',
      type: 'single_choice',
      timeLimitSeconds: 20,
      points: 1,
      options: [{ id: 'opt-1', text: 'Opció 1' }]
    }
    matchStore.isConnected = true

    vi.spyOn(matchStore, 'showResults')

    const wrapper = mount(HostGameView, {
      global: {
        stubs: {
          Tag: { template: '<span><slot /></span>' },
          Button: { template: '<button><slot /></button>' },
          Dialog: { template: '<div></div>' }
        }
      }
    })

    expect(wrapper.find('[data-testid="host-phase-active"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="host-active-timer"]').text()).toBe('18')

    const showResultsBtn = wrapper.find('[data-testid="btn-show-results"]')
    await showResultsBtn.trigger('click')

    expect(matchStore.showResults).toHaveBeenCalled()
  })
})
