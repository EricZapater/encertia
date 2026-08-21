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

  it('strips non-UUID temporary IDs (new-q-0, new-ans-0) when saving a new quiz', async () => {
    const { createQuiz } = await import('../api')
    const wrapper = mount(QuizEditorView, {
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })
    await flushPromises()

    // Set valid title
    const settingsModal = wrapper.findComponent({ name: 'QuizSettingsModal' })
    settingsModal.vm.$emit('saved', { title: 'Qüestionari de Prova' })

    // Fill question text
    const textInput = wrapper.find('[data-testid="input-question-text"]')
    await textInput.setValue('Com es diu la capital de Catalunya?')

    // Fill answer texts
    const ansInputs = wrapper.findAll('input.answer-text-input')
    for (let i = 0; i < ansInputs.length; i++) {
      await ansInputs[i].setValue(`Opció ${i + 1}`)
    }

    const saveBtn = wrapper.find('[data-testid="btn-save-quiz"]')
    await saveBtn.trigger('click')
    await flushPromises()

    expect(createQuiz).toHaveBeenCalledTimes(1)
    const callArg = (createQuiz as any).mock.calls[0][0]
    expect(callArg.title).toBe('Qüestionari de Prova')
    expect(callArg.questions).toHaveLength(1)
    expect(callArg.questions[0].id).toBeUndefined()
    expect(callArg.questions[0].answers).toHaveLength(4)
    callArg.questions[0].answers.forEach((ans: any) => {
      expect(ans.id).toBeUndefined()
    })
  })

  it('preserves valid UUIDs when updating an existing quiz', async () => {
    const validQuestionId = '123e4567-e89b-12d3-a456-426614174000'
    const validAnswerId = '987fcdeb-51a2-43f7-9abc-def012345678'
    const quizApi = await import('../api')
    ;(quizApi.getQuizById as any).mockResolvedValue({
      id: 'existing-quiz-id',
      creatorId: 'user-1',
      title: 'Quiz Existent',
      description: 'Descripció',
      coverImageUrl: null,
      status: 'draft',
      tags: [],
      questionCount: 1,
      createdAt: '',
      updatedAt: '',
      questions: [
        {
          id: validQuestionId,
          text: 'Pregunta existent?',
          imageUrl: null,
          questionType: 'single_choice',
          timeLimitSeconds: 20,
          orderIndex: 0,
          answers: [
            { id: validAnswerId, text: 'Resposta 1', isCorrect: true, orderIndex: 0 },
            { id: 'new-ans-1', text: 'Resposta 2', isCorrect: false, orderIndex: 1 }
          ]
        }
      ]
    })

    mockRoute.params.id = 'existing-quiz-id'
    const wrapper = mount(QuizEditorView, {
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })
    await flushPromises()

    const saveBtn = wrapper.find('[data-testid="btn-save-quiz"]')
    await saveBtn.trigger('click')
    await flushPromises()

    expect(quizApi.updateQuiz).toHaveBeenCalledTimes(1)
    const callArg = (quizApi.updateQuiz as any).mock.calls[0][1]
    expect(callArg.questions[0].id).toBe(validQuestionId)
    expect(callArg.questions[0].answers[0].id).toBe(validAnswerId)
    expect(callArg.questions[0].answers[1].id).toBeUndefined()
  })
})
