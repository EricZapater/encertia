import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import QuizPreviewModal from '../views/QuizPreviewModal.vue'
import type { QuizDetail } from '../types'

const dialogStub = {
  Dialog: {
    template: '<div v-if="visible" class="p-dialog"><slot name="header" /><slot /><slot name="footer" /></div>',
    props: ['visible']
  }
}

describe('QuizPreviewModal Component', () => {
  const sampleQuiz: QuizDetail = {
    id: 'quiz-1',
    creatorId: 'user-1',
    title: 'Geografia Catalana',
    status: 'published',
    tags: ['geo'],
    questionCount: 1,
    createdAt: '',
    updatedAt: '',
    questions: [
      {
        id: 'q-1',
        text: 'Quina és la capital de Catalunya?',
        imageUrl: null,
        questionType: 'single_choice',
        timeLimitSeconds: 20,
        orderIndex: 0,
        answers: [
          { id: 'ans-1', text: 'Barcelona', isCorrect: true, orderIndex: 0 },
          { id: 'ans-2', text: 'Girona', isCorrect: false, orderIndex: 1 },
          { id: 'ans-3', text: 'Lleida', isCorrect: false, orderIndex: 2 },
          { id: 'ans-4', text: 'Tarragona', isCorrect: false, orderIndex: 3 }
        ]
      }
    ]
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders question text, timer and answers', () => {
    const wrapper = mount(QuizPreviewModal, {
      props: {
        visible: true,
        quiz: sampleQuiz
      },
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })

    expect(wrapper.find('[data-testid="preview-question-text"]').text()).toBe(
      'Quina és la capital de Catalunya?'
    )
    expect(wrapper.find('[data-testid="preview-timer-value"]').text()).toBe('20s')
    expect(wrapper.findAll('.kahoot-answer-card').length).toBe(4)
  })

  it('answers a single-choice question, reveals correct answer and finishes simulation', async () => {
    const wrapper = mount(QuizPreviewModal, {
      props: {
        visible: true,
        quiz: sampleQuiz
      },
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })

    // Click Barcelona (correct option 0)
    const btnBarcelona = wrapper.find('[data-testid="preview-answer-btn-0"]')
    await btnBarcelona.trigger('click')

    // Expect revealed correct class
    expect(btnBarcelona.classes()).toContain('is-revealed-correct')

    // Next / Finish button should appear
    const nextBtn = wrapper.find('[data-testid="btn-next-preview-question"]')
    expect(nextBtn.exists()).toBe(true)

    // Click finish
    await nextBtn.trigger('click')

    // Completed screen should be displayed
    expect(wrapper.find('[data-testid="preview-completed-screen"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Simulació Completada!')
    expect(wrapper.text()).toContain('100%')
  })

  it('suporta mode individual (singleQuestionMode) per previsualitzar només una pregunta', async () => {
    const multiQuiz: QuizDetail = {
      ...sampleQuiz,
      questions: [
        sampleQuiz.questions[0],
        {
          id: 'q-2',
          text: 'Segona pregunta específica',
          imageUrl: null,
          questionType: 'multiple_choice',
          timeLimitSeconds: 15,
          orderIndex: 1,
          answers: [
            { id: 'ans-5', text: 'Opcio A', isCorrect: true, orderIndex: 0 },
            { id: 'ans-6', text: 'Opcio B', isCorrect: false, orderIndex: 1 }
          ]
        }
      ]
    }

    const wrapper = mount(QuizPreviewModal, {
      props: {
        visible: true,
        quiz: multiQuiz,
        initialQuestionIndex: 1,
        singleQuestionMode: true
      },
      global: {
        plugins: [PrimeVue],
        stubs: dialogStub
      }
    })

    expect(wrapper.find('[data-testid="preview-question-text"]').text()).toBe('Segona pregunta específica')
    expect(wrapper.text()).toContain('(Mode Individual)')
    expect(wrapper.find('[data-testid="preview-multichoice-badge"]').exists()).toBe(true)

    // Respondre
    await wrapper.find('[data-testid="preview-answer-btn-0"]').trigger('click')
    await wrapper.find('[data-testid="btn-submit-preview-multiple"]').trigger('click')

    // El botó final és 'Tornar a l\'Editor'
    const finishBtn = wrapper.find('[data-testid="btn-next-preview-question"]')
    expect(finishBtn.text()).toContain("Tornar a l'Editor")

    await finishBtn.trigger('click')
    expect(wrapper.emitted('update:visible')?.[0]).toEqual([false])
  })
})
