import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, RouterLinkStub } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import AppNavbar from '../AppNavbar.vue'
import { useAuthStore } from '@/modules/auth/store'

vi.mock('vue-router', () => ({
  useRoute: () => ({ path: '/quizzes' }),
  useRouter: () => ({ push: vi.fn() })
}))

describe('AppNavbar.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders brand and navigation links for teacher', () => {
    const authStore = useAuthStore()
    authStore.setUser({
      id: 'u-1',
      email: 'teacher@encertia.cat',
      firstName: 'Joan',
      lastName: 'Docent',
      role: 'teacher',
      createdAt: '2026-01-01',
      updatedAt: '2026-01-01'
    })

    const wrapper = mount(AppNavbar, {
      global: {
        stubs: {
          'router-link': RouterLinkStub,
          Tag: true,
          Button: true
        }
      }
    })

    expect(wrapper.text()).toContain('Encertia')
    expect(wrapper.text()).toContain('Jocs & Quizzes')
    expect(wrapper.text()).toContain('Avaluacions')
    expect(wrapper.text()).toContain('Usuaris')
    expect(wrapper.text()).toContain('Joan Docent')
    expect(wrapper.find('[data-testid="nav-link-metrics"]').exists()).toBe(false)
  })

  it('shows metrics link for admin role', () => {
    const authStore = useAuthStore()
    authStore.setUser({
      id: 'a-1',
      email: 'admin@encertia.cat',
      firstName: 'Admin',
      lastName: 'User',
      role: 'admin',
      createdAt: '2026-01-01',
      updatedAt: '2026-01-01'
    })

    const wrapper = mount(AppNavbar, {
      global: {
        stubs: {
          'router-link': RouterLinkStub,
          Tag: true,
          Button: true
        }
      }
    })

    expect(wrapper.find('[data-testid="nav-link-metrics"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Mètriques & Auditoria')
  })

  it('hides teacher/admin links for student role', () => {
    const authStore = useAuthStore()
    authStore.setUser({
      id: 's-1',
      email: 'student@encertia.cat',
      firstName: 'Pau',
      lastName: 'Alumne',
      role: 'student',
      createdAt: '2026-01-01',
      updatedAt: '2026-01-01'
    })

    const wrapper = mount(AppNavbar, {
      global: {
        stubs: {
          'router-link': RouterLinkStub,
          Tag: true,
          Button: true
        }
      }
    })

    expect(wrapper.text()).toContain('Jocs & Quizzes')
    expect(wrapper.text()).not.toContain('Avaluacions')
    expect(wrapper.text()).not.toContain('Usuaris')
    expect(wrapper.find('[data-testid="nav-link-metrics"]').exists()).toBe(false)
  })
})
