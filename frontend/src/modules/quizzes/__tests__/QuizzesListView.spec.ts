import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PrimeVue from 'primevue/config'
import QuizzesListView from '../views/QuizzesListView.vue'
import { useQuizStore } from '../store'
import type { Quiz } from '../types'

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: mockPush
  })
}))

vi.mock('../api', () => ({
  listQuizzes: vi.fn().mockResolvedValue({
    items: [
      {
        id: 'quiz-1',
        creatorId: 'user-1',
        title: 'Geografia Catalana',
        description: 'Capitals i comarques',
        coverImageUrl: 'https://pub-r2.encertia.cat/cover.jpg',
        status: 'published',
        tags: ['geografia', 'catalunya'],
        questionCount: 8,
        createdAt: '2026-08-21T10:00:00Z',
        updatedAt: '2026-08-21T11:00:00Z'
      }
    ],
    pagination: { page: 1, pageSize: 12, totalCount: 1, totalPages: 1 }
  }),
  deleteQuiz: vi.fn().mockResolvedValue({ message: 'Qüestionari eliminat' })
}))

describe('QuizzesListView Component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders page header, create button and filters', async () => {
    const wrapper = mount(QuizzesListView, {
      global: {
        plugins: [PrimeVue]
      }
    })

    expect(wrapper.find('.page-title').text()).toContain('Els meus Jocs i Qüestionaris')
    expect(wrapper.find('[data-testid="btn-create-quiz"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="input-search-quizzes"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="filter-status-select"]').exists()).toBe(true)
  })

  it('navigates to /quizzes/new when clicking Nou Joc', async () => {
    const wrapper = mount(QuizzesListView, {
      global: {
        plugins: [PrimeVue]
      }
    })

    const createBtn = wrapper.find('[data-testid="btn-create-quiz"]')
    await createBtn.trigger('click')
    expect(mockPush).toHaveBeenCalledWith('/quizzes/new')
  })

  it('displays quiz card with title, tags and question count', async () => {
    const store = useQuizStore()
    await store.fetchQuizzes()

    const wrapper = mount(QuizzesListView, {
      global: {
        plugins: [PrimeVue]
      }
    })

    expect(wrapper.find('[data-testid="quiz-card-quiz-1"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Geografia Catalana')
    expect(wrapper.text()).toContain('8 preguntes')
    expect(wrapper.text()).toContain('#geografia')
  })

  it('navigates to edit when clicking Editar button', async () => {
    const store = useQuizStore()
    await store.fetchQuizzes()

    const wrapper = mount(QuizzesListView, {
      global: {
        plugins: [PrimeVue]
      }
    })

    const editBtn = wrapper.find('[data-testid="btn-edit-quiz"]')
    await editBtn.trigger('click')
    expect(mockPush).toHaveBeenCalledWith('/quizzes/quiz-1/edit')
  })
})
