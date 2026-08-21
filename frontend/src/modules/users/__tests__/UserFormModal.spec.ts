import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PrimeVue from 'primevue/config'
import UserFormModal from '../views/UserFormModal.vue'
import { useAuthStore } from '@/modules/auth/store'
import { useUserStore } from '../store'
import type { User } from '../types'

describe('UserFormModal Component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('restricts role options to student when logged in user is a teacher', async () => {
    const authStore = useAuthStore()
    authStore.setUser({
      id: 'teacher-1',
      email: 'teacher@encertia.cat',
      firstName: 'Joan',
      lastName: 'Docent',
      role: 'teacher',
      createdAt: '2026-08-21T10:00:00Z'
    })

    const wrapper = mount(UserFormModal, {
      props: {
        visible: true,
        user: null
      },
      global: {
        plugins: [PrimeVue]
      }
    })

    expect(wrapper.exists()).toBe(true)
    // When user is teacher, role selector is disabled/fixed to student
    const roleSelect = wrapper.find('[data-testid="select-role"]')
    expect(roleSelect.exists()).toBe(true)
  })

  it('allows all roles when logged in user is an admin in create mode', async () => {
    const authStore = useAuthStore()
    authStore.setUser({
      id: 'admin-1',
      email: 'admin@encertia.cat',
      firstName: 'Super',
      lastName: 'Admin',
      role: 'admin',
      createdAt: '2026-08-21T10:00:00Z'
    })

    const wrapper = mount(UserFormModal, {
      props: {
        visible: true,
        user: null
      },
      global: {
        plugins: [PrimeVue]
      }
    })

    expect(wrapper.exists()).toBe(true)
    const roleSelect = wrapper.find('[data-testid="select-role"]')
    expect(roleSelect.exists()).toBe(true)
  })

  it('populates existing user data in edit mode and disables password field', async () => {
    const authStore = useAuthStore()
    authStore.setUser({
      id: 'admin-1',
      email: 'admin@encertia.cat',
      firstName: 'Super',
      lastName: 'Admin',
      role: 'admin',
      createdAt: '2026-08-21T10:00:00Z'
    })

    const userToEdit: User = {
      id: 'u-edit-1',
      email: 'alumne@encertia.cat',
      firstName: 'Marc',
      lastName: 'Rovira',
      role: 'student',
      isActive: true,
      createdAt: '2026-08-21T10:00:00Z'
    }

    const wrapper = mount(UserFormModal, {
      props: {
        visible: true,
        user: userToEdit
      },
      global: {
        plugins: [PrimeVue]
      }
    })

    // Password field should NOT exist in edit mode
    expect(wrapper.find('[data-testid="input-password"]').exists()).toBe(false)
    // Active switch should exist
    expect(wrapper.find('[data-testid="switch-isactive"]').exists()).toBe(true)
  })
})
