import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PrimeVue from 'primevue/config'
import BatchImportModal from '../views/BatchImportModal.vue'

describe('BatchImportModal Component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders correctly in initial step', () => {
    const wrapper = mount(BatchImportModal, {
      props: {
        visible: true
      },
      global: {
        plugins: [PrimeVue]
      }
    })

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('Selecciona o arrossega el fitxer CSV')
    expect(wrapper.text()).toContain('Descarregar Plantilla CSV')
  })
})
