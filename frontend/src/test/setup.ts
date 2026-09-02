import { vi } from 'vitest'
import { config } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import { i18n } from '../i18n'

config.global.plugins.push(PrimeVue)
config.global.plugins.push(ToastService)
config.global.plugins.push(i18n)

// Mock matchMedia for PrimeVue UI components
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn()
  }))
})

// Mock localStorage
const storage: Record<string, string> = {}
Object.defineProperty(window, 'localStorage', {
  value: {
    getItem: vi.fn((key: string) => storage[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      storage[key] = value
    }),
    removeItem: vi.fn((key: string) => {
      delete storage[key]
    }),
    clear: vi.fn(() => {
      Object.keys(storage).forEach((key) => delete storage[key])
    })
  },
  writable: true
})
