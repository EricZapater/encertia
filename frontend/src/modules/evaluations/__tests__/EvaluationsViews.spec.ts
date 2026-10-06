import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import EvaluationsListView from '../views/EvaluationsListView.vue'
import QuizEvaluationView from '../views/QuizEvaluationView.vue'
import { useEvaluationStore } from '../store'
import { useGroupStore } from '@/modules/groups/store'
import { useQuizStore } from '@/modules/quizzes/store'

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn()
  }),
  useRoute: () => ({
    params: {
      quizId: 'q-test-1'
    }
  })
}))

describe('Evaluations Views', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders EvaluationsListView with filter bar and $t keys correctly', async () => {
    const evalStore = useEvaluationStore()
    evalStore.evaluationsList = [
      {
        quizId: 'q-test-1',
        quizTitle: 'Matemàtiques — Tema 1',
        totalMatches: 3,
        totalStudents: 10,
        gradedCount: 5,
        lastMatchAt: '2026-08-21T10:00:00Z',
        groupId: 'g-1',
        groupName: '1A'
      }
    ]

    const wrapper = mount(EvaluationsListView, {
      global: {
        plugins: [PrimeVue, ToastService],
        mocks: {
          $t: (key: string) => key
        }
      }
    })

    expect(wrapper.text()).toContain('evaluations.title')
    expect(wrapper.text()).toContain('evaluations.filters.group')
    expect(wrapper.text()).toContain('evaluations.filters.game')
    expect(wrapper.text()).toContain('evaluations.filters.date')
    expect(wrapper.text()).toContain('Matemàtiques — Tema 1')
  })

  it('filters EvaluationsListView by date correctly', async () => {
    const evalStore = useEvaluationStore()
    evalStore.evaluationsList = [
      {
        quizId: 'q-test-1',
        quizTitle: 'Quiz August 21',
        totalMatches: 1,
        totalStudents: 5,
        gradedCount: 2,
        lastMatchAt: '2026-08-21T10:00:00Z'
      },
      {
        quizId: 'q-test-2',
        quizTitle: 'Quiz September 15',
        totalMatches: 2,
        totalStudents: 8,
        gradedCount: 4,
        lastMatchAt: '2026-09-15T12:00:00Z'
      }
    ]

    const wrapper = mount(EvaluationsListView, {
      global: {
        plugins: [PrimeVue, ToastService],
        mocks: {
          $t: (key: string) => key
        }
      }
    })

    expect(wrapper.text()).toContain('Quiz August 21')
    expect(wrapper.text()).toContain('Quiz September 15')

    // Set date filter to 2026-08-21
    const datePicker = wrapper.findComponent({ name: 'DatePicker' })
    await datePicker.vm.$emit('update:modelValue', new Date(2026, 7, 21))

    expect(wrapper.text()).toContain('Quiz August 21')
    expect(wrapper.text()).not.toContain('Quiz September 15')
  })

  it('renders QuizEvaluationView with collapsible panels, filter bar and $t keys', async () => {
    const evalStore = useEvaluationStore()
    vi.spyOn(evalStore, 'fetchQuizEvaluation').mockImplementation(async () => {})
    vi.spyOn(evalStore, 'fetchEvaluationsList').mockImplementation(async () => {})

    evalStore.activeQuizEvaluation = {
      quizId: 'q-test-1',
      quizTitle: 'Matemàtiques — Tema 1',
      totalMatches: 3,
      stats: [
        {
          questionId: 'q1',
          questionText: 'Pregunta 1',
          questionIndex: 0,
          hitRate: 0.8,
          avgResponseTimeMs: 3000,
          answerDistribution: [],
          noAnswerCount: 0
        }
      ],
      students: [
        {
          studentId: 's1',
          studentName: 'Alumne 1',
          groupId: 'g1',
          groupName: '1A',
          matchesCount: 1,
          calculatedGrade: 8.5,
          finalGrade: null,
          isGraded: false,
          lastMatchAt: '2026-08-21T10:00:00Z'
        },
        {
          studentId: 's2',
          studentName: 'Alumne 2',
          groupId: 'g1',
          groupName: '1A',
          matchesCount: 1,
          calculatedGrade: 7.0,
          finalGrade: null,
          isGraded: false,
          lastMatchAt: '2026-09-15T10:00:00Z'
        }
      ]
    }

    const wrapper = mount(QuizEvaluationView, {
      global: {
        plugins: [PrimeVue, ToastService],
        mocks: {
          $t: (key: string) => key
        }
      }
    })

    expect(wrapper.text()).toContain('evaluations.panels.globalStats')
    expect(wrapper.text()).toContain('evaluations.panels.studentResults')
    expect(wrapper.text()).toContain('Alumne 1')
    expect(wrapper.text()).toContain('Alumne 2')

    // Filter by date 2026-08-21
    const datePicker = wrapper.findComponent({ name: 'DatePicker' })
    await datePicker.vm.$emit('update:modelValue', new Date(2026, 7, 21))

    expect(wrapper.text()).toContain('Alumne 1')
    expect(wrapper.text()).not.toContain('Alumne 2')
  })

  it('renders common.noResults empty slot in DataTable when list is empty', async () => {
    const evalStore = useEvaluationStore()
    evalStore.evaluationsList = []

    const wrapper = mount(EvaluationsListView, {
      global: {
        plugins: [PrimeVue, ToastService],
        mocks: {
          $t: (key: string) => key
        }
      }
    })

    expect(wrapper.text()).toContain('common.noResults')
  })

  it('renders evaluation-filter-panel and clears filters on clear button click in EvaluationsListView', async () => {
    const evalStore = useEvaluationStore()
    const fetchSpy = vi.spyOn(evalStore, 'fetchEvaluationsList').mockImplementation(async () => {})

    const wrapper = mount(EvaluationsListView, {
      global: {
        plugins: [PrimeVue, ToastService],
        mocks: {
          $t: (key: string) => key
        }
      }
    })

    expect(wrapper.find('.evaluation-filter-panel').exists()).toBe(true)
    expect(wrapper.find('.filter-panel-header').exists()).toBe(true)
    expect(wrapper.find('.filter-grid').exists()).toBe(true)

    // Trigger clear button
    const clearBtn = wrapper.find('[data-testid="btn-clear-filters"]')
    expect(clearBtn.exists()).toBe(true)
    await clearBtn.trigger('click')

    expect(fetchSpy).toHaveBeenCalled()
  })

  it('renders evaluation-filter-panel and clears filters in QuizEvaluationView', async () => {
    const evalStore = useEvaluationStore()
    const fetchQuizSpy = vi.spyOn(evalStore, 'fetchQuizEvaluation').mockImplementation(async () => {})
    vi.spyOn(evalStore, 'fetchEvaluationsList').mockImplementation(async () => {})

    const wrapper = mount(QuizEvaluationView, {
      global: {
        plugins: [PrimeVue, ToastService],
        mocks: {
          $t: (key: string) => key
        }
      }
    })

    expect(wrapper.find('.evaluation-filter-panel').exists()).toBe(true)
    expect(wrapper.find('.filter-panel-header').exists()).toBe(true)
    expect(wrapper.find('.filter-grid').exists()).toBe(true)

    const clearBtn = wrapper.find('[data-testid="btn-clear-quiz-filters"]')
    expect(clearBtn.exists()).toBe(true)
    await clearBtn.trigger('click')

    expect(fetchQuizSpy).toHaveBeenCalled()
  })
})

