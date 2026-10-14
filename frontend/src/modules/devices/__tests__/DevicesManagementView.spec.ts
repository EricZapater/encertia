import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import DevicesManagementView from '../views/DevicesManagementView.vue'
import * as devicesApi from '../api'

vi.mock('../api', () => ({
  listDevices: vi.fn(),
  revokeDevice: vi.fn(),
  downloadAgent: vi.fn()
}))

const dialogStub = {
  Dialog: {
    template:
      '<div v-if="visible" class="p-dialog" v-bind="$attrs"><slot name="header" /><slot /><slot name="footer" /></div>',
    props: ['visible']
  }
}

describe('DevicesManagementView.vue', () => {
  const mockDevices = [
    {
      id: 'dev-1',
      name: 'Portàtil Aula 3B',
      teacherId: 'teacher-1',
      teacherName: 'Joan Docent',
      teacherEmail: 'joan.docent@encertia.cat',
      isActive: true,
      lastUsedAt: '2026-10-14T09:30:00Z',
      createdAt: '2026-09-01T08:00:00Z'
    },
    {
      id: 'dev-2',
      name: 'Ordinador Seminari',
      teacherId: 'teacher-2',
      teacherName: 'Marta Professora',
      teacherEmail: 'marta@encertia.cat',
      isActive: false,
      lastUsedAt: null,
      createdAt: '2026-08-15T12:00:00Z'
    }
  ]

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.mocked(devicesApi.listDevices).mockResolvedValue([...mockDevices])
  })

  it('renders download cards for Windows, macOS, Linux and Certificate', async () => {
    const wrapper = mount(DevicesManagementView, {
      global: {
        plugins: [PrimeVue, ToastService],
        stubs: dialogStub
      }
    })

    await flushPromises()

    expect(wrapper.find('[data-testid="card-agent-windows"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="card-agent-darwin"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="card-agent-linux"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="card-agent-cert"]').exists()).toBe(true)

    expect(wrapper.text()).toContain('.\\encertia-agent-windows.exe --register')
    expect(wrapper.text()).toContain('chmod +x encertia-agent-darwin')
    expect(wrapper.text()).toContain('chmod +x encertia-agent-linux')
    expect(wrapper.text()).toContain('certutil -addstore -f "ROOT" encertia-codesign.cer')

    expect(wrapper.find('[data-testid="btn-download-windows"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="btn-download-darwin"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="btn-download-linux"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="btn-download-cert"]').exists()).toBe(true)
  })

  it('loads and renders authorized devices in the table', async () => {
    const wrapper = mount(DevicesManagementView, {
      global: {
        plugins: [PrimeVue, ToastService],
        stubs: dialogStub
      }
    })

    await flushPromises()

    expect(devicesApi.listDevices).toHaveBeenCalledTimes(1)
    expect(wrapper.find('[data-testid="devices-table"]').exists()).toBe(true)

    const tableText = wrapper.text()
    expect(tableText).toContain('Portàtil Aula 3B')
    expect(tableText).toContain('Joan Docent')
    expect(tableText).toContain('joan.docent@encertia.cat')
    expect(tableText).toContain('Ordinador Seminari')
    expect(tableText).toContain('Marta Professora')

    const revokeButtons = wrapper.findAll('[data-testid="btn-revoke-device"]')
    expect(revokeButtons.length).toBe(1) // Only active device has revoke button
  })

  it('triggers agent download correctly when clicking download button', async () => {
    const mockBlob = new Blob(['binary-content'], { type: 'application/octet-stream' })
    vi.mocked(devicesApi.downloadAgent).mockResolvedValue(mockBlob)

    const createObjectURLMock = vi.fn().mockReturnValue('blob:http://localhost/mock-uuid')
    const revokeObjectURLMock = vi.fn()
    window.URL.createObjectURL = createObjectURLMock
    window.URL.revokeObjectURL = revokeObjectURLMock

    const wrapper = mount(DevicesManagementView, {
      global: {
        plugins: [PrimeVue, ToastService],
        stubs: dialogStub
      }
    })

    await flushPromises()

    const downloadBtn = wrapper.find('[data-testid="btn-download-windows"]')
    await downloadBtn.trigger('click')
    await flushPromises()

    expect(devicesApi.downloadAgent).toHaveBeenCalledWith('windows')
    expect(createObjectURLMock).toHaveBeenCalledWith(mockBlob)
    expect(revokeObjectURLMock).toHaveBeenCalledWith('blob:http://localhost/mock-uuid')
  })

  it('opens confirmation modal and revokes device', async () => {
    vi.mocked(devicesApi.revokeDevice).mockResolvedValue()

    const wrapper = mount(DevicesManagementView, {
      global: {
        plugins: [PrimeVue, ToastService],
        stubs: dialogStub
      }
    })

    await flushPromises()

    const revokeBtn = wrapper.find('[data-testid="btn-revoke-device"]')
    await revokeBtn.trigger('click')

    expect(wrapper.find('[data-testid="revoke-confirm-dialog"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Portàtil Aula 3B')

    const confirmBtn = wrapper.find('[data-testid="btn-confirm-revoke"]')
    await confirmBtn.trigger('click')
    await flushPromises()

    expect(devicesApi.revokeDevice).toHaveBeenCalledWith('dev-1')
    expect(devicesApi.listDevices).toHaveBeenCalledTimes(2) // Initial + after revoke
  })
})
