import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PrimeVue from 'primevue/config'
import UsersListView from '../views/UsersListView.vue'
import { useAuthStore } from '@/modules/auth/store'
import { useUserStore } from '../store'

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn()
  })
}))

vi.mock('../api', () => ({
  listUsers: vi.fn().mockResolvedValue({
    items: [],
    pagination: { page: 1, pageSize: 20, totalCount: 0, totalPages: 0 }
  })
}))

describe('UsersListView Component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders page header and action buttons', () => {
    const authStore = useAuthStore()
    authStore.setUser({
      id: 'admin-1',
      email: 'admin@encertia.cat',
      firstName: 'Admin',
      lastName: 'Principal',
      role: 'admin',
      createdAt: '2026-08-21T10:00:00Z'
    })

    const wrapper = mount(UsersListView, {
      global: {
        plugins: [PrimeVue]
      }
    })

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.find('[data-testid="btn-open-create-user"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="btn-open-batch-import"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="input-search-users"]').exists()).toBe(true)
  })

  it('shows role filter for admin user', () => {
    const authStore = useAuthStore()
    authStore.setUser({
      id: 'admin-1',
      email: 'admin@encertia.cat',
      firstName: 'Admin',
      lastName: 'Principal',
      role: 'admin',
      createdAt: '2026-08-21T10:00:00Z'
    })

    const wrapper = mount(UsersListView, {
      global: {
        plugins: [PrimeVue]
      }
    })

    expect(wrapper.find('[data-testid="filter-role-select"]').exists()).toBe(true)
  })

  it('hides multi-role filter for teacher user and restricts role to student', () => {
    const authStore = useAuthStore()
    authStore.setUser({
      id: 'teacher-1',
      email: 'teacher@encertia.cat',
      firstName: 'Docent',
      lastName: 'Profe',
      role: 'teacher',
      createdAt: '2026-08-21T10:00:00Z'
    })

    const wrapper = mount(UsersListView, {
      global: {
        plugins: [PrimeVue]
      }
    })

    // In teacher view, role filter select is hidden because role is fixed to student
    expect(wrapper.find('[data-testid="filter-role-select"]').exists()).toBe(false)
  })
})
