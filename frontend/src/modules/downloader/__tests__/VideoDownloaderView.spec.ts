import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import VideoDownloaderView from '../views/VideoDownloaderView.vue'
import * as downloaderApi from '../api'

vi.mock('../api', () => ({
  fetchVideoInfo: vi.fn(),
  downloadVideoFile: vi.fn()
}))

describe('VideoDownloaderView.vue', () => {
  const mockVideoInfo = {
    id: 'dQw4w9WgXcQ',
    title: 'Rick Astley - Never Gonna Give You Up',
    thumbnail: 'https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg',
    duration: 213,
    extractor: 'youtube',
    uploader: 'Rick Astley',
    description: 'Official Rick Astley video'
  }

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.mocked(downloaderApi.fetchVideoInfo).mockResolvedValue(mockVideoInfo)
    vi.mocked(downloaderApi.downloadVideoFile).mockResolvedValue({
      blob: new Blob(['fake video mp4'], { type: 'video/mp4' }),
      filename: 'Rick_Astley.mp4'
    })
  })

  it('renders correctly with input field and action buttons', () => {
    const wrapper = mount(VideoDownloaderView, {
      global: {
        plugins: [PrimeVue, ToastService]
      }
    })

    expect(wrapper.find('[data-testid="video-downloader-view"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="input-video-url"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="btn-paste-url"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="btn-inspect-video"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="btn-direct-download"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="btn-toggle-advanced"]').exists()).toBe(true)
  })

  it('disables inspect and direct download buttons when URL is empty', () => {
    const wrapper = mount(VideoDownloaderView, {
      global: {
        plugins: [PrimeVue, ToastService]
      }
    })

    const inspectBtn = wrapper.find('[data-testid="btn-inspect-video"]')
    const downloadBtn = wrapper.find('[data-testid="btn-direct-download"]')

    expect(inspectBtn.attributes('disabled')).toBeDefined()
    expect(downloadBtn.attributes('disabled')).toBeDefined()
  })

  it('enables buttons when valid URL is entered and inspects video', async () => {
    const wrapper = mount(VideoDownloaderView, {
      global: {
        plugins: [PrimeVue, ToastService]
      }
    })

    const input = wrapper.find('[data-testid="input-video-url"]')
    await input.setValue('https://www.youtube.com/watch?v=dQw4w9WgXcQ')

    const inspectBtn = wrapper.find('[data-testid="btn-inspect-video"]')
    expect(inspectBtn.attributes('disabled')).toBeUndefined()

    await inspectBtn.trigger('click')
    await flushPromises()

    expect(downloaderApi.fetchVideoInfo).toHaveBeenCalledWith('https://www.youtube.com/watch?v=dQw4w9WgXcQ', '')

    // Preview card appears
    expect(wrapper.find('[data-testid="card-video-preview"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="video-title"]').text()).toBe('Rick Astley - Never Gonna Give You Up')
    expect(wrapper.find('[data-testid="video-platform"]').text()).toContain('YouTube')
    expect(wrapper.find('[data-testid="btn-download-mp4"]').exists()).toBe(true)
  })

  it('handles video download correctly', async () => {
    // Mock URL.createObjectURL and URL.revokeObjectURL
    const createObjectURLSpy = vi.fn().mockReturnValue('blob:http://localhost/dummy')
    const revokeObjectURLSpy = vi.fn()
    window.URL.createObjectURL = createObjectURLSpy
    window.URL.revokeObjectURL = revokeObjectURLSpy

    const wrapper = mount(VideoDownloaderView, {
      global: {
        plugins: [PrimeVue, ToastService]
      }
    })

    const input = wrapper.find('[data-testid="input-video-url"]')
    await input.setValue('https://www.youtube.com/watch?v=dQw4w9WgXcQ')

    const downloadBtn = wrapper.find('[data-testid="btn-direct-download"]')
    await downloadBtn.trigger('click')
    await flushPromises()

    expect(downloaderApi.downloadVideoFile).toHaveBeenCalledWith('https://www.youtube.com/watch?v=dQw4w9WgXcQ', '')
    expect(createObjectURLSpy).toHaveBeenCalled()
    expect(revokeObjectURLSpy).toHaveBeenCalled()
  })

  it('toggles advanced cookies panel and passes cookies to api', async () => {
    const wrapper = mount(VideoDownloaderView, {
      global: {
        plugins: [PrimeVue, ToastService]
      }
    })

    expect(wrapper.find('[data-testid="advanced-panel"]').exists()).toBe(false)
    await wrapper.find('[data-testid="btn-toggle-advanced"]').trigger('click')
    expect(wrapper.find('[data-testid="advanced-panel"]').exists()).toBe(true)

    const cookiesInput = wrapper.find('[data-testid="input-cookies"]')
    await cookiesInput.setValue('test_cookie_data=123')

    const urlInput = wrapper.find('[data-testid="input-video-url"]')
    await urlInput.setValue('https://www.youtube.com/watch?v=dQw4w9WgXcQ')

    await wrapper.find('[data-testid="btn-inspect-video"]').trigger('click')
    await flushPromises()

    expect(downloaderApi.fetchVideoInfo).toHaveBeenCalledWith(
      'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
      'test_cookie_data=123'
    )
  })

  it('shows error banner when inspect fails', async () => {
    vi.mocked(downloaderApi.fetchVideoInfo).mockRejectedValueOnce({
      response: { data: { message: "No s'ha pogut obtenir el vídeo" } }
    })

    const wrapper = mount(VideoDownloaderView, {
      global: {
        plugins: [PrimeVue, ToastService]
      }
    })

    const input = wrapper.find('[data-testid="input-video-url"]')
    await input.setValue('https://invalid-url.com/video')

    await wrapper.find('[data-testid="btn-inspect-video"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="downloader-error-message"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="downloader-error-message"]').text()).toContain("No s'ha pogut obtenir el vídeo")
  })

  it('clears URL and state when clear button is clicked', async () => {
    const wrapper = mount(VideoDownloaderView, {
      global: {
        plugins: [PrimeVue, ToastService]
      }
    })

    const input = wrapper.find('[data-testid="input-video-url"]')
    await input.setValue('https://www.youtube.com/watch?v=dQw4w9WgXcQ')

    const clearBtn = wrapper.find('[data-testid="btn-clear-url"]')
    expect(clearBtn.exists()).toBe(true)

    await clearBtn.trigger('click')
    expect((input.element as HTMLInputElement).value).toBe('')
    expect(wrapper.find('[data-testid="card-video-preview"]').exists()).toBe(false)
  })
})
