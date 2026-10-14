import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PrimeVue from 'primevue/config'
import LoginView from '../views/LoginView.vue'
import { useAuthStore } from '../store'
import { i18n } from '@/i18n'

const mockToastAdd = vi.fn()
vi.mock('primevue/usetoast', () => ({
  useToast: () => ({
    add: mockToastAdd
  })
}))

const mockRouterPush = vi.fn()
let mockRouteQuery: Record<string, string> = {}

vi.mock('vue-router', () => ({
  useRoute: () => ({
    query: mockRouteQuery
  }),
  useRouter: () => ({
    push: mockRouterPush
  })
}))

describe('LoginView.vue Component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    mockRouteQuery = {}
  })

  it('renders standard login form without device verified badge by default', () => {
    mockRouteQuery = {}

    const wrapper = mount(LoginView, {
      global: {
        plugins: [PrimeVue, i18n],
        stubs: {
          'router-link': true
        }
      }
    })

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.find('[data-testid="device-verified-badge"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Encertia')
  })

  it('displays device verified badge when deviceToken query param is present', () => {
    mockRouteQuery = {
      deviceToken: 'dbt_test_token_999'
    }

    const wrapper = mount(LoginView, {
      global: {
        plugins: [PrimeVue, i18n],
        stubs: {
          'router-link': true
        }
      }
    })

    const badge = wrapper.find('[data-testid="device-verified-badge"]')
    expect(badge.exists()).toBe(true)
    expect(badge.text()).toContain('Dispositiu autoritzat i verificat')
    expect(badge.find('.pi-shield').exists()).toBe(true)

    const authStore = useAuthStore()
    expect(authStore.deviceToken).toBe('dbt_test_token_999')
  })

  it('pre-fills email field when email query param is present', () => {
    mockRouteQuery = {
      email: 'joan.docent@encertia.cat',
      deviceToken: 'dbt_test_token_888'
    }

    const wrapper = mount(LoginView, {
      global: {
        plugins: [PrimeVue, i18n],
        stubs: {
          'router-link': true
        }
      }
    })

    const emailInput = wrapper.find('input#email')
    expect(emailInput.exists()).toBe(true)
    expect((emailInput.element as HTMLInputElement).value).toBe('joan.docent@encertia.cat')
  })

  it('submits login with deviceToken when present in query params', async () => {
    mockRouteQuery = {
      email: 'joan.docent@encertia.cat',
      deviceToken: 'dbt_hardware_token_abc'
    }

    const authStore = useAuthStore()
    const loginSpy = vi.spyOn(authStore, 'login').mockResolvedValueOnce({
      user: {
        id: 'u-1',
        email: 'joan.docent@encertia.cat',
        firstName: 'Joan',
        lastName: 'Docent',
        role: 'teacher',
        createdAt: '2026-08-21T10:00:00Z'
      },
      tokens: {
        accessToken: 'acc_jwt',
        refreshToken: 'ref_jwt',
        tokenType: 'Bearer',
        expiresIn: 900
      }
    })

    const wrapper = mount(LoginView, {
      global: {
        plugins: [PrimeVue, i18n],
        stubs: {
          'router-link': true
        }
      }
    })

    const passwordInput = wrapper.find('#password input')
    await passwordInput.setValue('supersecret123')

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(loginSpy).toHaveBeenCalledWith({
      email: 'joan.docent@encertia.cat',
      password: 'supersecret123',
      deviceToken: 'dbt_hardware_token_abc'
    })
    expect(mockRouterPush).toHaveBeenCalledWith('/profile')
  })

  it('submits login without deviceToken when query param is absent', async () => {
    mockRouteQuery = {}

    const authStore = useAuthStore()
    const loginSpy = vi.spyOn(authStore, 'login').mockResolvedValueOnce({
      user: {
        id: 'u-2',
        email: 'student@encertia.cat',
        firstName: 'Alumne',
        lastName: 'U',
        role: 'student',
        createdAt: '2026-08-21T10:00:00Z'
      },
      tokens: {
        accessToken: 'acc_jwt',
        refreshToken: 'ref_jwt',
        tokenType: 'Bearer',
        expiresIn: 900
      }
    })

    const wrapper = mount(LoginView, {
      global: {
        plugins: [PrimeVue, i18n],
        stubs: {
          'router-link': true
        }
      }
    })

    await wrapper.find('input#email').setValue('student@encertia.cat')
    await wrapper.find('#password input').setValue('studentpass123')

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(loginSpy).toHaveBeenCalledWith({
      email: 'student@encertia.cat',
      password: 'studentpass123'
    })
    expect(mockRouterPush).toHaveBeenCalledWith('/profile')
  })
})
