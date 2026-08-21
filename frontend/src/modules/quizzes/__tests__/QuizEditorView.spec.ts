import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PrimeVue from 'primevue/config'
import QuizEditorView from '../views/QuizEditorView.vue'

const mockRoute = {
  params: { id: 'new' }
}

const mockRouter = {
  push: vi.fn(),
  replace: vi.fn()
}

vi.mock('vue-router', () => ({
  useRoute: () => mockRoute,
  useRouter: () => mockRouter
}))

vi.mock('../api', () => ({
  createQuiz: vi.fn().mockResolvedValue({
    id: 'created-quiz-1',
    title: 'Geografia 101',
    status: 'draft',
    tags: [],
    questionCount: 1,
    createdAt: '',
    updatedAt: '',
    questions: []
  }),
  updateQuiz: vi.fn(),
  getQuizById: vi.fn(),
  listQuizzes: vi.fn().mockResolvedValue({ items: [], pagination: {} })
}))

const dialogStub = {
  Dialog: {
    template: '<div v-if="visible" class="p-dialog"><slot name="header" /><slot /><slot name="footer" /></div>',
    props: ['visible']
  }
}

describe('QuizEditorView Component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    mockRoute.params.id = 'new'
  })

  it('initializes with a default question and 4 Kahoot answer options for a new quiz', async () => {
    const wrapper = mount(QuizEditorView, {
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="questions-sidebar"]').exists()).toBe(true)
    expect(wrapper.findAll('.question-thumb-card').length).toBe(1)
    expect(wrapper.findAll('.kahoot-editor-card').length).toBe(4)
  })

  it('allows adding a new question to the sidebar', async () => {
    const wrapper = mount(QuizEditorView, {
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })
    await flushPromises()

    const addBtn = wrapper.find('[data-testid="btn-add-question"]')
    await addBtn.trigger('click')

    expect(wrapper.findAll('.question-thumb-card').length).toBe(2)
  })

  it('allows adding up to 6 answers and removing down to 2 answers', async () => {
    const wrapper = mount(QuizEditorView, {
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })
    await flushPromises()

    // Initial has 4 answers
    expect(wrapper.findAll('.kahoot-editor-card').length).toBe(4)

    // Add 5th and 6th answer
    const addAnsBtn = wrapper.find('[data-testid="btn-add-answer-option"]')
    await addAnsBtn.trigger('click')
    expect(wrapper.findAll('.kahoot-editor-card').length).toBe(5)

    await wrapper.find('[data-testid="btn-add-answer-option"]').trigger('click')
    expect(wrapper.findAll('.kahoot-editor-card').length).toBe(6)

    // With 6 answers, the add button should no longer exist (max 6)
    expect(wrapper.find('[data-testid="btn-add-answer-option"]').exists()).toBe(false)

    // Remove one answer (button testid is btn-remove-answer-0)
    const removeBtn = wrapper.find('[data-testid="btn-remove-answer-0"]')
    await removeBtn.trigger('click')
    expect(wrapper.findAll('.kahoot-editor-card').length).toBe(5)
  })

  it('shows error if trying to save with an empty question text', async () => {
    const wrapper = mount(QuizEditorView, {
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })
    await flushPromises()

    // Set title
    const settingsModal = wrapper.findComponent({ name: 'QuizSettingsModal' })
    settingsModal.vm.$emit('saved', { title: 'Test Quiz Valid Title' })

    const saveBtn = wrapper.find('[data-testid="btn-save-quiz"]')
    await saveBtn.trigger('click')

    expect(wrapper.find('[data-testid="editor-feedback-msg"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('La pregunta #1 no té enunciat')
  })

  it('toggles correct answer indicator properly', async () => {
    const wrapper = mount(QuizEditorView, {
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })
    await flushPromises()

    // Toggle option 1 to correct in single_choice mode
    const toggleBtn1 = wrapper.find('[data-testid="btn-toggle-correct-1"]')
    await toggleBtn1.trigger('click')

    expect(toggleBtn1.classes()).toContain('is-correct-active')
    const toggleBtn0 = wrapper.find('[data-testid="btn-toggle-correct-0"]')
    expect(toggleBtn0.classes()).not.toContain('is-correct-active')
  })
})
