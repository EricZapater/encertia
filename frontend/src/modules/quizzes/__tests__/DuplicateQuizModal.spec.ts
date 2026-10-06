import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PrimeVue from 'primevue/config'
import DuplicateQuizModal from '../views/DuplicateQuizModal.vue'
import { useQuizStore } from '../store'
import type { Quiz } from '../types'

const dialogStub = {
  Dialog: {
    template: '<div v-if="visible" class="p-dialog"><slot name="header" /><slot /><slot name="footer" /></div>',
    props: ['visible']
  }
}

describe('DuplicateQuizModal Component', () => {
  const sampleQuiz: Quiz = {
    id: 'quiz-1',
    creatorId: 'user-1',
    title: 'Geografia de Catalunya',
    description: 'Qüestionari interactiu',
    coverImageUrl: null,
    status: 'published',
    tags: ['geografia'],
    questionCount: 10,
    createdAt: '2026-08-21T10:00:00Z',
    updatedAt: '2026-08-21T10:00:00Z'
  }

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders modal with default values when opened', () => {
    const wrapper = mount(DuplicateQuizModal, {
      props: {
        visible: true,
        quiz: sampleQuiz
      },
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })

    const titleInput = wrapper.find('[data-testid="input-duplicate-title"]')
    expect(titleInput.exists()).toBe(true)
    expect((titleInput.element as HTMLInputElement).value).toBe('[Còpia] Geografia de Catalunya')

    const checkbox = wrapper.find('[data-testid="checkbox-include-answers"]')
    expect(checkbox.exists()).toBe(true)
  })

  it('calls store.duplicateQuiz with custom title and includeAnswers flag', async () => {
    const store = useQuizStore()
    const duplicateSpy = vi.spyOn(store, 'duplicateQuiz').mockResolvedValue({
      id: 'copy-1',
      creatorId: 'user-1',
      title: 'Geografia - Examen A',
      status: 'draft',
      tags: ['geografia'],
      questionCount: 10,
      createdAt: '',
      updatedAt: '',
      questions: []
    })

    const wrapper = mount(DuplicateQuizModal, {
      props: {
        visible: true,
        quiz: sampleQuiz
      },
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })

    const titleInput = wrapper.find('[data-testid="input-duplicate-title"]')
    await titleInput.setValue('Geografia - Examen A')

    // Click confirm button
    const confirmBtn = wrapper.find('[data-testid="btn-confirm-duplicate"]')
    await confirmBtn.trigger('click')

    expect(duplicateSpy).toHaveBeenCalledWith('quiz-1', {
      title: 'Geografia - Examen A',
      includeAnswers: false
    })
    expect(wrapper.emitted('duplicated')).toBeTruthy()
  })
})
